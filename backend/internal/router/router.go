package router

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"wetalk/config"
	"wetalk/internal/friend"
	"wetalk/internal/message"
	"wetalk/internal/middleware"
	"wetalk/internal/user"
	"wetalk/internal/ws"
)

// Dependencies 路由注册所需的所有依赖
type Dependencies struct {
	Config        *config.Config
	Hub           *ws.Hub
	SendMsgFunc   ws.SendMessageFunc
	UserHandler   *user.Handler
	FriendHandler *friend.Handler
	MsgHandler    *message.Handler
}

// Setup 创建 Gin Engine 并注册所有路由
func Setup(deps *Dependencies) *gin.Engine {
	r := gin.Default()

	// 全局中间件
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 健康检查
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// WebSocket（认证通过 auth frame，不走 JWT middleware）
	r.GET("/ws", ws.WSHandler(deps.Hub, deps.Config.JWT.Secret, deps.SendMsgFunc))

	// 认证路由（无需 JWT）
	registerAuthRoutes(r, deps.UserHandler)

	// 需要认证的路由
	registerAPIRoutes(r, deps.Config.JWT.Secret, deps.UserHandler, deps.FriendHandler, deps.MsgHandler)

	return r
}

func registerAuthRoutes(r *gin.Engine, h *user.Handler) {
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
	}
}

func registerAPIRoutes(r *gin.Engine, jwtSecret string, uh *user.Handler, fh *friend.Handler, mh *message.Handler) {
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(jwtSecret))
	{
		// 当前用户
		api.GET("/me", uh.Me)

		// 用户搜索
		api.GET("/users", uh.Search)

		// 好友管理
		registerFriendRoutes(api, fh)

		// 消息管理
		registerMessageRoutes(api, mh)
	}
}

func registerFriendRoutes(api *gin.RouterGroup, h *friend.Handler) {
	friends := api.Group("/friends")
	{
		friends.POST("", h.Add)
		friends.GET("/pending", h.PendingRequests)
		friends.GET("", h.List)
		friends.PUT("/:id/accept", h.Accept)
		friends.DELETE("/:id", h.Delete)
	}
}

func registerMessageRoutes(api *gin.RouterGroup, h *message.Handler) {
	messages := api.Group("/messages")
	{
		messages.POST("", h.Send)
		messages.GET("", h.List)
		messages.PUT("/read", h.Read)
	}
}