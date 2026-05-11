package router

import (
	"github.com/gin-gonic/gin"

	"wetalk/controller"
)

func registerUserRoutes(api *gin.RouterGroup, h *controller.UserHandler) {
	me := api.Group("/me")
	{
		me.GET("", h.Me)
		me.POST("/avatar", h.UploadAvatar)
	}

	api.GET("/users", h.Search)
}
