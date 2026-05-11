package service

import (
	"wetalk/dto"
	"wetalk/model"
	"wetalk/type"
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
		return nil, types.ErrUnauthorized
	}

	// 如果设置了 quoted_message_id，验证引用的消息存在且属于同一聊天
	if req.QuoteMessageID != nil {
		quoted, err := dto.Message.GetByID(*req.QuoteMessageID)
		if err != nil {
			return nil, err
		}
		if quoted == nil || quoted.ChatID != req.ChatID {
			return nil, types.ErrInvalidParam
		}
	}

	contentType := req.ContentType
	if contentType == "" {
		contentType = model.ContentTypeText
	}

	// 创建消息
	msg, err := dto.Message.Create(senderID, req.ChatID, req.Content, contentType, req.QuoteMessageID)
	if err != nil {
		return nil, err
	}

	// 更新聊天的最后一条消息信息
	if updateErr := s.chatSvc.UpdateLastMessage(req.ChatID, msg.ID, msg.Content, msg.CreatedAt); updateErr != nil {
		// 非关键路径，仅记录错误但不中断
		_ = updateErr
	}

	// 如果有文件元数据，创建 file_metadata 记录
	var fileMeta *model.FileMetadata
	if req.FileMetadata != nil {
		fm, err := dto.FileMetadata.Create(
			msg.ID,
			req.FileMetadata.URL,
			req.FileMetadata.OriginalName,
			req.FileMetadata.FileSize,
			req.FileMetadata.MimeType,
			req.FileMetadata.Width,
			req.FileMetadata.Height,
		)
		if err != nil {
			return nil, err
		}
		fileMeta = fm
	}

	// 构建引用消息摘要
	var quotedContent *string
	if msg.QuoteMessageID != nil {
		quoted, _ := dto.Message.GetByID(*msg.QuoteMessageID)
		if quoted != nil {
			quotedContent = &quoted.Content
		}
	}

	return &model.MessageResponse{
		ID:             msg.ID,
		ChatID:         msg.ChatID,
		SenderID:       msg.SenderID,
		Content:        msg.Content,
		ContentType:    msg.ContentType,
		QuoteMessageID: msg.QuoteMessageID,
		QuotedContent:  quotedContent,
		FileMetadata:   fileMeta,
		Status:         msg.Status,
		CreatedAt:      msg.CreatedAt,
	}, nil
}

// MarkAsRead 标记消息已读
func (s *MessageService) MarkAsRead(userID, chatID int64) error {
	return dto.Message.MarkAsRead(chatID, userID)
}

// GetConversation 获取聊天记录
func (s *MessageService) GetConversation(chatID int64, offset, limit int) ([]model.MessageResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	messages, err := dto.Message.ListByChat(chatID, offset, limit)
	if err != nil {
		return nil, err
	}

	// 批量收集需要查询的 ID
	var fileMsgIDs []int64
	var quoteIDs []int64
	for _, msg := range messages {
		if msg.ContentType == model.ContentTypeImage || msg.ContentType == model.ContentTypeFile {
			fileMsgIDs = append(fileMsgIDs, msg.ID)
		}
		if msg.QuoteMessageID != nil {
			quoteIDs = append(quoteIDs, *msg.QuoteMessageID)
		}
	}

	// 批量查询文件元数据
	metadataMap := make(map[int64]*model.FileMetadata)
	if len(fileMsgIDs) > 0 {
		metadataList, err := dto.FileMetadata.ListByMessageIDs(fileMsgIDs)
		if err == nil {
			for i := range metadataList {
				metadataMap[metadataList[i].MessageID] = &metadataList[i]
			}
		}
	}

	// 批量查询引用消息内容
	quoteMap := make(map[int64]string)
	if len(quoteIDs) > 0 {
		quoteList, err := dto.Message.ListByIDs(quoteIDs)
		if err == nil {
			for i := range quoteList {
				quoteMap[quoteList[i].ID] = quoteList[i].Content
			}
		}
	}

	// 构建响应
	responses := make([]model.MessageResponse, len(messages))
	for i, msg := range messages {
		resp := model.MessageResponse{
			ID:             msg.ID,
			ChatID:         msg.ChatID,
			SenderID:       msg.SenderID,
			Content:        msg.Content,
			ContentType:    msg.ContentType,
			QuoteMessageID: msg.QuoteMessageID,
			Status:         msg.Status,
			CreatedAt:      msg.CreatedAt,
		}
		if fm, ok := metadataMap[msg.ID]; ok {
			resp.FileMetadata = fm
		}
		if msg.QuoteMessageID != nil {
			if content, ok := quoteMap[*msg.QuoteMessageID]; ok {
				resp.QuotedContent = &content
			}
		}
		responses[i] = resp
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
		return &ws.SentMessage{
			ID:             msgResp.ID,
			ChatID:         msgResp.ChatID,
			SenderID:       msgResp.SenderID,
			Content:        msgResp.Content,
			ContentType:    msgResp.ContentType,
			QuoteMessageID: msgResp.QuoteMessageID,
			QuotedContent:  msgResp.QuotedContent,
			FileMetadata:   meta,
			Status:         msgResp.Status,
			CreatedAt:      msgResp.CreatedAt,
		}, nil
	}
}
