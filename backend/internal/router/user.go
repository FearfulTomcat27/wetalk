package router

import (
	"github.com/gin-gonic/gin"

	"wetalk/internal/user"
)

func registerUserRoutes(api *gin.RouterGroup, h *user.Handler) {
	me := api.Group("/me")
	{
		me.GET("", h.Me)
		me.POST("/avatar", h.UploadAvatar)
	}

	api.GET("/users", h.Search)
}
