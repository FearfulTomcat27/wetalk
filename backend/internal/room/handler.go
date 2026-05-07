package room

import "github.com/gin-gonic/gin"

// Handler 聊天室 HTTP 处理器（预留）
type Handler struct{}

// NewHandler 创建聊天室处理器
func NewHandler() *Handler {
	return &Handler{}
}

// List 获取聊天室列表（预留）
func (h *Handler) List(c *gin.Context) {
	c.JSON(200, gin.H{"code": 200, "message": "聊天室模块待实现"})
}
