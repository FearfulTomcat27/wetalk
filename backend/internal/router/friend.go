package router

import (
	"github.com/gin-gonic/gin"

	"wetalk/internal/friend"
)

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
