package controller

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"wetalk/config"
	"wetalk/model"
	"wetalk/service"
	"wetalk/types"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ========== Controller 测试用 Mock ==========

type ctrlMockUserRepo struct {
	mock.Mock
}

func (m *ctrlMockUserRepo) FindByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *ctrlMockUserRepo) FindByID(id int64) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *ctrlMockUserRepo) SearchByUsername(keyword string, limit int) ([]model.User, error) {
	args := m.Called(keyword, limit)
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *ctrlMockUserRepo) Create(username, passwordHash, nickname, avatar string) (*model.User, error) {
	args := m.Called(username, passwordHash, nickname, avatar)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *ctrlMockUserRepo) UpdateAvatar(userID int64, avatarURL string) error {
	args := m.Called(userID, avatarURL)
	return args.Error(0)
}

type ctrlMockCache struct {
	mock.Mock
}

func (m *ctrlMockCache) Get(key string, dest interface{}) (bool, error) {
	args := m.Called(key, dest)
	return args.Bool(0), args.Error(1)
}

func (m *ctrlMockCache) Set(key string, value interface{}, ttl time.Duration) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}

func (m *ctrlMockCache) Del(keys ...string) error {
	args := m.Called(keys)
	return args.Error(0)
}

type ctrlMockOSS struct {
	mock.Mock
}

func (m *ctrlMockOSS) PutObject(ctx context.Context, key string, body io.Reader, contentType string, contentLength int64) (string, error) {
	args := m.Called(ctx, key, body, contentType, contentLength)
	return args.String(0), args.Error(1)
}

// ========== 测试辅助函数 ==========

func newTestUserHandler(repo *ctrlMockUserRepo, cache *ctrlMockCache, oss *ctrlMockOSS) *UserHandler {
	userSvc := service.NewUserService(repo, cache, config.JWTConfig{Secret: "test-secret", ExpireHours: 72}, oss)
	return NewUserHandler(userSvc)
}

func TestUserHandler_Register_Success(t *testing.T) {
	repo := new(ctrlMockUserRepo)
	cache := new(ctrlMockCache)
	oss := new(ctrlMockOSS)

	repo.On("FindByUsername", "newuser").Return(nil, nil)
	repo.On("Create", "newuser", mock.AnythingOfType("string"), "New", mock.AnythingOfType("string")).
		Return(&model.User{ID: 1, Username: "newuser", Nickname: "New", Avatar: "https://example.com/avatar.svg"}, nil)

	h := newTestUserHandler(repo, cache, oss)

	r := gin.New()
	r.POST("/api/auth/register", h.Register)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register",
		strings.NewReader(`{"username":"newuser","password":"pass123","nickname":"New"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var resp types.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, int(resp.Code.(float64)))
	assert.Equal(t, "注册成功", resp.Message)

	repo.AssertExpectations(t)
}

func TestUserHandler_Register_InvalidJSON(t *testing.T) {
	h := newTestUserHandler(new(ctrlMockUserRepo), new(ctrlMockCache), new(ctrlMockOSS))
	r := gin.New()
	r.POST("/api/auth/register", h.Register)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register",
		strings.NewReader(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Register_AlreadyExists(t *testing.T) {
	repo := new(ctrlMockUserRepo)

	repo.On("FindByUsername", "existing").Return(&model.User{ID: 1, Username: "existing"}, nil)

	h := newTestUserHandler(repo, new(ctrlMockCache), new(ctrlMockOSS))
	r := gin.New()
	r.POST("/api/auth/register", h.Register)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register",
		strings.NewReader(`{"username":"existing","password":"pass123"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestUserHandler_Login_Success(t *testing.T) {
	repo := new(ctrlMockUserRepo)

	u := &model.User{ID: 1, Username: "testuser", Nickname: "Test"}
	h, _ := bcrypt.GenerateFromPassword([]byte("pass123"), bcrypt.DefaultCost)
	u.PasswordHash = string(h)
	repo.On("FindByUsername", "testuser").Return(u, nil)

	handler := newTestUserHandler(repo, new(ctrlMockCache), new(ctrlMockOSS))
	r := gin.New()
	r.POST("/api/auth/login", handler.Login)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		strings.NewReader(`{"username":"testuser","password":"pass123"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_Login_InvalidCredentials(t *testing.T) {
	repo := new(ctrlMockUserRepo)
	repo.On("FindByUsername", "nobody").Return(nil, nil)

	h := newTestUserHandler(repo, new(ctrlMockCache), new(ctrlMockOSS))
	r := gin.New()
	r.POST("/api/auth/login", h.Login)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		strings.NewReader(`{"username":"nobody","password":"wrong"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUserHandler_Me(t *testing.T) {
	repo := new(ctrlMockUserRepo)
	cache := new(ctrlMockCache)

	repo.On("FindByID", int64(1)).Return(&model.User{ID: 1, Username: "testuser", Nickname: "Test"}, nil)
	cache.On("Get", "user:1", mock.Anything).Return(false, nil)
	cache.On("Set", "user:1", mock.Anything, mock.AnythingOfType("time.Duration")).Return(nil)

	h := newTestUserHandler(repo, cache, new(ctrlMockOSS))
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Next()
	})
	r.GET("/api/me", h.Me)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_Me_NotFound(t *testing.T) {
	repo := new(ctrlMockUserRepo)
	cache := new(ctrlMockCache)

	repo.On("FindByID", int64(999)).Return(nil, nil)
	cache.On("Get", "user:999", mock.Anything).Return(false, nil)

	h := newTestUserHandler(repo, cache, new(ctrlMockOSS))
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", int64(999))
		c.Next()
	})
	r.GET("/api/me", h.Me)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
