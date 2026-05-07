package message

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	pkgerrors "wetalk/pkg/errors"
	"wetalk/pkg/utils"
)

// Handler 消息 HTTP 处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建消息处理器
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Send 发送消息
func (h *Handler) Send(c *gin.Context) {
	senderID := c.GetInt64("user_id")

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数校验失败: "+err.Error())
		return
	}

	msg, err := h.svc.SendMessage(senderID, req)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrInvalidParam) {
			utils.Error(c, http.StatusBadRequest, "无效的请求参数")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "发送消息失败")
		return
	}

	utils.Success(c, http.StatusCreated, "消息已发送", msg)
}

// List 获取聊天记录
func (h *Handler) List(c *gin.Context) {
	userID := c.GetInt64("user_id")

	friendIDStr := c.Query("friend_id")
	friendID, err := strconv.ParseInt(friendIDStr, 10, 64)
	if err != nil || friendID == 0 {
		utils.Error(c, http.StatusBadRequest, "缺少 friend_id 参数")
		return
	}

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	messages, err := h.svc.GetConversation(userID, friendID, offset, limit)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取消息失败")
		return
	}

	if messages == nil {
		messages = []Message{}
	}

	utils.Success(c, http.StatusOK, "成功", messages)
}
