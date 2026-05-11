package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"wetalk/config"
	"wetalk/controller"
	"wetalk/middleware"
	"wetalk/ws"
)

// Dependencies 路由注册所需的所有依赖
type Dependencies struct {
	Config        *config.Config
	Hub           *ws.Hub
	SendMsgFunc   ws.SendMessageFunc
	UserHandler   *controller.UserHandler
	FriendHandler *controller.FriendHandler
	MsgHandler    *controller.MessageHandler
	UploadHandler *controller.UploadHandler
}

// Setup 创建 Gin Engine 并注册所有路由
func Setup(deps *Dependencies) *gin.Engine {
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CORSMiddleware())

	// 健康检查
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// WebSocket（认证通过 auth frame，不走 JWT middleware）
	r.GET("/ws", ws.WSHandler(deps.Hub, deps.Config.JWT.Secret, deps.SendMsgFunc))

	// 认证路由（无需 JWT）
	registerAuthRoutes(r, deps.UserHandler)

	// 需要认证的路由
	registerAPIRoutes(r, deps.Config.JWT.Secret, deps.UserHandler, deps.FriendHandler, deps.MsgHandler, deps.UploadHandler)

	return r
}

func registerAuthRoutes(r *gin.Engine, h *controller.UserHandler) {
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
	}
}

func registerAPIRoutes(r *gin.Engine, jwtSecret string, uh *controller.UserHandler, fh *controller.FriendHandler, mh *controller.MessageHandler, uhUpload *controller.UploadHandler) {
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(jwtSecret))
	{
		// 当前用户
		registerUserRoutes(api, uh)

		// 好友管理
		registerFriendRoutes(api, fh)

		// 消息管理
		registerMessageRoutes(api, mh)

		// 上传管理
		registerUploadRoutes(api, uhUpload)
	}
}
