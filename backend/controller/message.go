package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"wetalk/common"
	"wetalk/model"
	"wetalk/service"
	"wetalk/type"
	"wetalk/ws"
)

// MessageHandler 消息 HTTP 处理器
type MessageHandler struct {
	svc     *service.MessageService
	chatSvc *service.ChatService
	hub     *ws.Hub
}

// NewMessageHandler 创建消息处理器
func NewMessageHandler(svc *service.MessageService, chatSvc *service.ChatService, hub *ws.Hub) *MessageHandler {
	return &MessageHandler{svc: svc, chatSvc: chatSvc, hub: hub}
}

// Send 发送消息
func (h *MessageHandler) Send(c *gin.Context) {
	senderID := c.GetInt64("user_id")

	var req model.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, http.StatusBadRequest, "参数校验失败: "+err.Error())
		return
	}

	msgResp, err := h.svc.SendMessage(senderID, req)
	if err != nil {
		if errors.Is(err, types.ErrInvalidParam) {
			common.Error(c, http.StatusBadRequest, "无效的请求参数")
			return
		}
		if errors.Is(err, types.ErrUnauthorized) {
			common.Error(c, http.StatusForbidden, "不是聊天成员")
			return
		}
		common.Error(c, http.StatusInternalServerError, "发送消息失败")
		return
	}

	// 推送 message.new 给聊天所有在线成员
	memberIDs, err := h.chatSvc.GetMemberIDs(msgResp.ChatID)
	if err == nil {
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
		event := &ws.MessageNewEvent{
			Type:           ws.TypeMessageNew,
			ID:             msgResp.ID,
			ChatID:         msgResp.ChatID,
			SenderID:       msgResp.SenderID,
			Content:        msgResp.Content,
			ContentType:    msgResp.ContentType,
			QuoteMessageID: msgResp.QuoteMessageID,
			QuotedContent:  msgResp.QuotedContent,
			FileMetadata:   fileMeta,
			Status:         msgResp.Status,
			CreatedAt:      msgResp.CreatedAt,
		}
		for _, memberID := range memberIDs {
			if h.hub.IsOnline(memberID) {
				h.hub.SendTo(memberID, event)
			}
		}
	}

	common.Success(c, http.StatusCreated, "消息已发送", msgResp)
}

// Read 标记消息已读
func (h *MessageHandler) Read(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		ChatID int64 `json:"chat_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, http.StatusBadRequest, "参数校验失败: "+err.Error())
		return
	}

	if err := h.svc.MarkAsRead(userID, req.ChatID); err != nil {
		common.Error(c, http.StatusInternalServerError, "标记已读失败")
		return
	}

	common.Success(c, http.StatusOK, "已标记已读", nil)
}

// List 获取聊天记录
func (h *MessageHandler) List(c *gin.Context) {
	userID := c.GetInt64("user_id")

	chatIDStr := c.Query("chat_id")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil || chatID == 0 {
		common.Error(c, http.StatusBadRequest, "缺少 chat_id 参数")
		return
	}

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	messages, err := h.svc.GetConversation(chatID, offset, limit)
	if err != nil {
		common.Error(c, http.StatusInternalServerError, "获取消息失败")
		return
	}

	if messages == nil {
		messages = []model.MessageResponse{}
	}

	common.Success(c, http.StatusOK, "成功", messages)

	// userID is unused in this handler but kept for potential auth verification
	_ = userID
}
