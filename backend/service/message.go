package service

import (
	"fmt"
	"time"

	"wetalk/common"
	"wetalk/dto"
	"wetalk/model"
	"wetalk/types"
	"wetalk/ws"
)

// MessageService 消息业务逻辑
type MessageService struct {
	chatSvc *ChatService
}

// NewMessageService 创建消息服务
func NewMessageService(chatSvc *ChatService) *MessageService {
	return &MessageService{chatSvc: chatSvc}
}

// SendMessage 发送消息
func (s *MessageService) SendMessage(senderID int64, req model.SendMessageRequest) (*model.MessageResponse, error) {
	if req.Content == "" {
		return nil, types.ErrInvalidParam
	}

	// 验证发送者是聊天成员
	isMember, err := s.chatSvc.IsMember(req.ChatID, senderID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, types.ErrNotChatMember
	}

	// 如果设置了 quoted_message_id，验证引用的消息存在且属于同一聊天
	if req.QuoteMessageID != nil {
		quoted, err := dto.Message.GetByID(*req.QuoteMessageID)
		if err != nil {
			return nil, err
		}
		if quoted == nil || quoted.ChatID != req.ChatID {
			return nil, types.ErrQuotedMessageNotFound
		}
	}

	contentType := req.ContentType
	if contentType == "" {
		contentType = model.ContentTypeText
	}

	// 创建消息（MongoDB，含嵌入式 file_metadata 和引用消息内容预填充）
	msgResp, err := dto.Message.Create(senderID, req.ChatID, req.Content, contentType, req.QuoteMessageID, req.FileMetadata)
	if err != nil {
		return nil, err
	}

	// 更新聊天的最后一条消息信息
	if updateErr := s.chatSvc.UpdateLastMessage(req.ChatID, msgResp.ID, msgResp.Content, msgResp.ContentType, msgResp.CreatedAt); updateErr != nil {
		_ = updateErr
	}

	return msgResp, nil
}

// MarkAsRead 标记消息已读
func (s *MessageService) MarkAsRead(userID, chatID int64) error {
	err := dto.Message.MarkAsRead(chatID, userID)
	if err != nil {
		return err
	}
	go func() { _ = common.CacheDel(fmt.Sprintf("unread:%d", userID)) }()
	return nil
}

// GetUnreadCounts 获取当前用户所有聊天的未读消息数
func (s *MessageService) GetUnreadCounts(userID int64) ([]dto.UnreadCount, error) {
	key := fmt.Sprintf("unread:%d", userID)
	var counts []dto.UnreadCount
	if ok, _ := common.CacheGet(key, &counts); ok {
		return counts, nil
	}

	var err error
	counts, err = dto.Message.GetUnreadCounts(userID)
	if err != nil {
		return nil, err
	}
	if counts == nil {
		counts = []dto.UnreadCount{}
	}

	if cacheErr := common.CacheSet(key, counts, 20*time.Second); cacheErr != nil {
		_ = cacheErr
	}
	return counts, nil
}

// GetConversation 获取聊天记录
// MongoDB 文档自带嵌入式 file_metadata，无需额外查询
func (s *MessageService) GetConversation(chatID int64, offset, limit int) ([]model.MessageResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	responses, err := dto.Message.ListByChat(chatID, offset, limit)
	if err != nil {
		return nil, err
	}

	if responses == nil {
		return []model.MessageResponse{}, nil
	}
	return responses, nil
}

// NewSendMessageFunc 创建 WS 发送消息回调（桥接 ws 包与 message 包的类型转换）
func NewSendMessageFunc(msgSvc *MessageService) ws.SendMessageFunc {
	return func(senderID int64, chatID int64, content string, contentType string, clientMsgID string, quoteID *int64, fileMetadata *ws.WSFileMetadata) (*ws.SentMessage, error) {
		var filePayload *model.FileMetadataPayload
		if fileMetadata != nil {
			filePayload = &model.FileMetadataPayload{
				URL:          fileMetadata.URL,
				OriginalName: fileMetadata.OriginalName,
				FileSize:     fileMetadata.FileSize,
				MimeType:     fileMetadata.MimeType,
				Width:        fileMetadata.Width,
				Height:       fileMetadata.Height,
			}
		}
		msgResp, err := msgSvc.SendMessage(senderID, model.SendMessageRequest{
			ChatID:         chatID,
			Content:        content,
			ContentType:    contentType,
			QuoteMessageID: quoteID,
			FileMetadata:   filePayload,
		})
		if err != nil {
			return nil, err
		}
		var meta *ws.WSFileMetadata
		if msgResp.FileMetadata != nil {
			meta = &ws.WSFileMetadata{
				URL:          msgResp.FileMetadata.URL,
				OriginalName: msgResp.FileMetadata.OriginalName,
				FileSize:     msgResp.FileMetadata.FileSize,
				MimeType:     msgResp.FileMetadata.MimeType,
				Width:        msgResp.FileMetadata.Width,
				Height:       msgResp.FileMetadata.Height,
			}
		}
		var quoteMeta *ws.WSFileMetadata
		if msgResp.QuotedFileMeta != nil {
			quoteMeta = &ws.WSFileMetadata{
				URL:          msgResp.QuotedFileMeta.URL,
				OriginalName: msgResp.QuotedFileMeta.OriginalName,
				FileSize:     msgResp.QuotedFileMeta.FileSize,
				MimeType:     msgResp.QuotedFileMeta.MimeType,
				Width:        msgResp.QuotedFileMeta.Width,
				Height:       msgResp.QuotedFileMeta.Height,
			}
		}
		return &ws.SentMessage{
			ID:                msgResp.ID,
			ChatID:            msgResp.ChatID,
			SenderID:          msgResp.SenderID,
			Content:           msgResp.Content,
			ContentType:       msgResp.ContentType,
			QuoteMessageID:    msgResp.QuoteMessageID,
			QuotedContent:     msgResp.QuotedContent,
			QuotedSenderID:    msgResp.QuotedSenderID,
			QuotedSenderName:  msgResp.QuotedSenderName,
			QuotedContentType: msgResp.QuotedContentType,
			QuotedFileMeta:    quoteMeta,
			FileMetadata:      meta,
			Status:            msgResp.Status,
			CreatedAt:         msgResp.CreatedAt,
		}, nil
	}
}
