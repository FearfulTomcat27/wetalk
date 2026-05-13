package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"wetalk/common"
	"wetalk/dto"
	"wetalk/model"
	"wetalk/service"

	"wetalk/types"
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
	return &MessageHandler{
		svc:     svc,
		chatSvc: chatSvc,
		hub:     hub,
	}
}

// Service 暴露内部 MessageService 供 router 构造 WS 发送回调
func (h *MessageHandler) Service() *service.MessageService {
	return h.svc
}

// Send 发送消息
// @Summary 发送消息
// @Description 发送文本/图片/文件消息，支持引用回复；通过 WebSocket 实时推送给聊天中的在线成员
// @Tags 消息
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.SendMessageRequest true "消息内容"
// @Success 201 {object} types.Response{data=model.MessageResponse} "消息已发送"
// @Failure 400 {object} types.Response "参数错误"
// @Failure 401 {object} types.Response "未认证"
// @Failure 403 {object} types.Response "不是聊天成员"
// @Router /api/messages [post]
func (h *MessageHandler) Send(c *gin.Context) {
	senderID := c.GetInt64("user_id")

	var req model.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AppError(c, types.NewInvalidParamf(err.Error()))
		return
	}

	msgResp, err := h.svc.SendMessage(senderID, req)
	if err != nil {
		var appErr *types.AppError
		if errors.As(err, &appErr) {
			common.AppError(c, appErr)
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
// @Summary 标记消息已读
// @Description 将指定聊天中来自对方的消息全部标记为已读
// @Tags 消息
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{chat_id=int} true "聊天 ID"
// @Success 200 {object} types.Response "已标记已读"
// @Failure 400 {object} types.Response "参数错误"
// @Failure 401 {object} types.Response "未认证"
// @Router /api/messages/read [put]
func (h *MessageHandler) Read(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		ChatID int64 `json:"chat_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AppError(c, types.NewInvalidParamf(err.Error()))
		return
	}

	if err := h.svc.MarkAsRead(userID, req.ChatID); err != nil {
		common.Error(c, http.StatusInternalServerError, "标记已读失败")
		return
	}

	common.Success(c, http.StatusOK, "已标记已读", nil)
}

// Unread 获取当前用户所有聊天的未读消息数
// @Summary 获取未读消息数
// @Description 获取当前用户在所有聊天中的未读消息数量（供侧边栏角标使用）
// @Tags 消息
// @Produce json
// @Security BearerAuth
// @Success 200 {object} types.Response{data=[]dto.UnreadCount} "成功"
// @Failure 401 {object} types.Response "未认证"
// @Router /api/messages/unread [get]
func (h *MessageHandler) Unread(c *gin.Context) {
	userID := c.GetInt64("user_id")

	counts, err := h.svc.GetUnreadCounts(userID)
	if err != nil {
		common.Error(c, http.StatusInternalServerError, "获取未读消息数失败")
		return
	}
	if counts == nil {
		counts = []dto.UnreadCount{}
	}

	common.Success(c, http.StatusOK, "成功", counts)
}

// DeleteHistory 当前用户删除某个聊天的聊天记录
func (h *MessageHandler) DeleteHistory(c *gin.Context) {
	userID := c.GetInt64("user_id")

	chatIDStr := c.Param("chat_id")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil || chatID == 0 {
		common.AppError(c, types.NewInvalidParamf("无效的 chat_id"))
		return
	}

	if err := h.svc.DeleteChatHistory(userID, chatID); err != nil {
		var appErr *types.AppError
		if errors.As(err, &appErr) {
			common.AppError(c, appErr)
			return
		}
		common.Error(c, http.StatusInternalServerError, "删除聊天记录失败")
		return
	}

	common.Success(c, http.StatusOK, "聊天记录已删除", nil)
}

// List 获取聊天记录
// @Summary 获取聊天记录
// @Description 获取指定聊天的历史消息记录（按时间倒序，支持分页）
// @Tags 消息
// @Produce json
// @Security BearerAuth
// @Param chat_id query int true "聊天 ID"
// @Param offset query int false "偏移量（默认 0）"
// @Param limit query int false "每页数量（默认 50）"
// @Success 200 {object} types.Response{data=[]model.MessageResponse} "成功"
// @Failure 400 {object} types.Response "参数错误"
// @Failure 401 {object} types.Response "未认证"
// @Router /api/messages [get]
func (h *MessageHandler) List(c *gin.Context) {
	userID := c.GetInt64("user_id")
	chatIDStr := c.Query("chat_id")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil || chatID == 0 {
		common.AppError(c, types.NewInvalidParamf("缺少 chat_id 参数"))
		return
	}

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	messages, err := h.svc.GetConversation(chatID, userID, offset, limit)
	if err != nil {
		common.Error(c, http.StatusInternalServerError, "获取消息失败")
		return
	}

	if messages == nil {
		messages = []model.MessageResponse{}
	}

	common.Success(c, http.StatusOK, "成功", messages)
}
