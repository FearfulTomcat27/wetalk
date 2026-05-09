package router

import (
	"github.com/gin-gonic/gin"

	"wetalk/internal/message"
)

func registerMessageRoutes(api *gin.RouterGroup, h *message.Handler) {
	messages := api.Group("/messages")
	{
		messages.POST("", h.Send)
		messages.GET("", h.List)
		messages.PUT("/read", h.Read)
	}
}

func registerUploadRoutes(api *gin.RouterGroup, h *message.UploadHandler) {
	api.POST("/upload", h.Upload)
}
