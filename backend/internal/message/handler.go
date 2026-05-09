package message

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"wetalk/internal/ws"
	pkgerrors "wetalk/pkg/errors"
	"wetalk/pkg/utils"
)

// Handler 消息 HTTP 处理器
type Handler struct {
	svc *Service
	hub *ws.Hub
}

// NewHandler 创建消息处理器
func NewHandler(svc *Service, hub *ws.Hub) *Handler {
	return &Handler{svc: svc, hub: hub}
}

// Send 发送消息
func (h *Handler) Send(c *gin.Context) {
	senderID := c.GetInt64("user_id")

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数校验失败: "+err.Error())
		return
	}

	msgResp, err := h.svc.SendMessage(senderID, req)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrInvalidParam) {
			utils.Error(c, http.StatusBadRequest, "无效的请求参数")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "发送消息失败")
		return
	}

	// 若接收者在线，通过 Hub 推送 message.new
	if h.hub.IsOnline(msgResp.ReceiverID) {
		var fileMeta *ws.WSFileMetadata
		if msgResp.FileMetadata != nil {
			fileMeta = &ws.WSFileMetadata{
				URL:          msgResp.FileMetadata.URL,
				OriginalName: msgResp.FileMetadata.OriginalName,
				FileSize:     msgResp.FileMetadata.FileSize,
				MimeType:     msgResp.FileMetadata.MimeType,
				Width:        msgResp.FileMetadata.Width,
				Height:       msgResp.FileMetadata.Height,
			}
		}
		h.hub.SendTo(msgResp.ReceiverID, &ws.MessageNewEvent{
			Type:         ws.TypeMessageNew,
			ID:           msgResp.ID,
			SenderID:     msgResp.SenderID,
			ReceiverID:   msgResp.ReceiverID,
			Content:      msgResp.Content,
			ContentType:  msgResp.ContentType,
			FileMetadata: fileMeta,
			Status:       msgResp.Status,
			CreatedAt:    msgResp.CreatedAt,
		})
	}

	utils.Success(c, http.StatusCreated, "消息已发送", msgResp)
}

// Read 标记消息已读
func (h *Handler) Read(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		SenderID int64 `json:"sender_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数校验失败: "+err.Error())
		return
	}

	if err := h.svc.MarkAsRead(userID, req.SenderID); err != nil {
		utils.Error(c, http.StatusInternalServerError, "标记已读失败")
		return
	}

	utils.Success(c, http.StatusOK, "已标记已读", nil)
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
		messages = []MessageResponse{}
	}

	utils.Success(c, http.StatusOK, "成功", messages)
}
