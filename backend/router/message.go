package router

import (
	"github.com/gin-gonic/gin"

	"wetalk/controller"
)

func registerMessageRoutes(api *gin.RouterGroup, h *controller.MessageHandler) {
	messages := api.Group("/messages")
	{
		messages.POST("", h.Send)
		messages.GET("", h.List)
		messages.GET("/unread", h.Unread)
		messages.PUT("/read", h.Read)
	}
}

func registerUploadRoutes(api *gin.RouterGroup, h *controller.UploadHandler) {
	api.POST("/upload", h.Upload)
}
