//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"

	"wetalk/common"
	"wetalk/config"
	"wetalk/controller"
	"wetalk/db"
	"wetalk/dto"
	"wetalk/middleware"
	"wetalk/model"
	"wetalk/service"
	"wetalk/types"
	"wetalk/ws"
)

// ========== 全局数据库清理 ==========

// resetDB 清空所有数据库表/集合，供每个测试独立使用
func resetDB(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	// MySQL：按外键依赖顺序清理
	db.DB.Exec("DELETE FROM friend_requests")
	db.DB.Exec("DELETE FROM friendships")
	db.DB.Exec("DELETE FROM chat_members")
	db.DB.Exec("DELETE FROM chats")
	db.DB.Exec("DELETE FROM users")

	// MongoDB：清空消息集合和计数器
	_, _ = db.MsgCollection().DeleteMany(ctx, bson.M{})
	_, _ = db.CountersCollection().DeleteMany(ctx, bson.M{})

	// Redis：清空数据库
	db.RedisClient.FlushDB(ctx)
}

// ========== OSS 存根 ==========

type stubOSS struct {
	mu    sync.Mutex
	calls []ossCall
}

type ossCall struct {
	Key           string
	ContentType   string
	ContentLength int64
}

func (s *stubOSS) PutObject(_ context.Context, key string, _ io.Reader, contentType string, contentLength int64) (string, error) {
	s.mu.Lock()
	s.calls = append(s.calls, ossCall{Key: key, ContentType: contentType, ContentLength: contentLength})
	s.mu.Unlock()
	return "https://test-bucket.oss-cn-test.aliyuncs.com/" + key, nil
}

func (s *stubOSS) ResetCalls() {
	s.mu.Lock()
	s.calls = nil
	s.mu.Unlock()
}

// cache 是共享的 RedisCache 实例，供 service 层 IT 使用
var cache = &common.RedisCache{}

// ========== 测试 Service 构建函数 ==========

func newTestUserService() *service.UserService {
	return service.NewUserService(dto.User, &common.RedisCache{},
		config.JWTConfig{Secret: "test-secret", ExpireHours: 72}, &stubOSS{})
}

func newTestChatService() *service.ChatService {
	return service.NewChatService(dto.Chat, &common.RedisCache{})
}

func newTestFriendService() *service.FriendService {
	return service.NewFriendService(dto.Friend, dto.User, newTestChatService(), &common.RedisCache{})
}

func newTestMessageService() *service.MessageService {
	return service.NewMessageService(dto.Message, newTestChatService(), &common.RedisCache{})
}

func newTestUploadService(oss service.ObjectStorage) *service.UploadService {
	return service.NewUploadService(oss)
}

// ========== JWT 令牌生成 ==========

func generateToken(userID int64, username string) string {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(72 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString([]byte("test-secret"))
	return s
}

// ========== API 测试辅助 ==========

// newTestEngine 创建测试用的 Gin 引擎，注册所有路由
func newTestEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	userSvc := newTestUserService()
	chatSvc := newTestChatService()

	hub := ws.NewHub(chatSvc.GetMemberIDs)
	go hub.Run()
	friendSvc := newTestFriendService()
	msgSvc := newTestMessageService()

	userH := controller.NewUserHandler(userSvc)
	friendH := controller.NewFriendHandler(friendSvc, hub)
	msgH := controller.NewMessageHandler(msgSvc, chatSvc, hub)

	// 公开路由
	r.POST("/api/auth/register", userH.Register)
	r.POST("/api/auth/login", userH.Login)
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// 需要 JWT 认证的路由
	auth := r.Group("/api")
	auth.Use(middleware.AuthMiddleware("test-secret"))
	{
		auth.GET("/me", userH.Me)
		auth.GET("/users", userH.Search)

		auth.POST("/friends", friendH.Add)
		auth.GET("/friends", friendH.List)
		auth.GET("/friends/pending", friendH.PendingRequests)
		auth.PUT("/friends/:id/accept", friendH.Accept)
		auth.DELETE("/friends/:id", friendH.Delete)

		auth.POST("/messages", msgH.Send)
		auth.GET("/messages", msgH.List)
		auth.GET("/messages/unread", msgH.Unread)
		auth.PUT("/messages/read", msgH.Read)
	}

	return r
}

// apiResult 解析 API 响应为 types.Response
func apiResult(t *testing.T, body []byte) types.Response {
	t.Helper()
	var resp types.Response
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("解析响应失败: %v, body: %s", err, string(body))
	}
	return resp
}

// registerUser 注册用户并返回用户信息
func registerUser(t *testing.T, engine *gin.Engine, username, password, nickname string) *model.AuthResponse {
	t.Helper()
	body := fmt.Sprintf(`{"username":"%s","password":"%s","nickname":"%s"}`, username, password, nickname)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("注册失败: code=%d body=%s", w.Code, w.Body.String())
	}
	resp := apiResult(t, w.Body.Bytes())
	data, _ := json.Marshal(resp.Data)
	var authResp model.AuthResponse
	json.Unmarshal(data, &authResp)
	return &authResp
}

// authHeaders 返回 JWT 认证头
func authHeaders(token string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + token,
	}
}

// doRequest 发起 HTTP 请求
func doRequest(t *testing.T, engine *gin.Engine, method, path string, body io.Reader, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	engine.ServeHTTP(w, req)
	return w
}
