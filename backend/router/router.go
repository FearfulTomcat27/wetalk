package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"wetalk/common"
	"wetalk/config"
	"wetalk/controller"
	"wetalk/middleware"
	"wetalk/service"
	"wetalk/ws"

	_ "wetalk/docs"
)

// Dependencies 路由注册所需的所有依赖（由 wire 组装）
type Dependencies struct {
	Config         *config.Config
	Hub            *ws.Hub
	OSSClient      *common.Client
	UserHandler    *controller.UserHandler
	FriendHandler  *controller.FriendHandler
	MessageHandler *controller.MessageHandler
	UploadHandler  *controller.UploadHandler
	ChatService    *service.ChatService
}

// Setup 创建 Gin Engine 并注册所有路由
func Setup(deps *Dependencies) *gin.Engine {
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CORSMiddleware())

	// WS 基础设施所需的共享依赖
	deps.Hub.SetGetMemberIDs(deps.ChatService.GetMemberIDs)
	sendMsgFunc := service.NewSendMessageFunc(deps.MessageHandler.Service())

	// 健康检查
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// Swagger 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// WebSocket（认证通过 auth frame，不走 JWT middleware）
	r.GET("/ws", ws.WSHandler(deps.Hub, deps.Config.JWT.Secret, sendMsgFunc))

	// 认证路由（无需 JWT）
	registerAuthRoutes(r, deps.UserHandler)

	// 需要认证的路由
	registerAPIRoutes(r, deps)

	return r
}

func registerAuthRoutes(r *gin.Engine, h *controller.UserHandler) {
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
	}
}

func registerAPIRoutes(r *gin.Engine, deps *Dependencies) {
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(deps.Config.JWT.Secret))
	{
		registerUserRoutes(api, deps.UserHandler)
		registerFriendRoutes(api, deps.FriendHandler)
		registerMessageRoutes(api, deps.MessageHandler)
		registerUploadRoutes(api, deps.UploadHandler)
	}
}
