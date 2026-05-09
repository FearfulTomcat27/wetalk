package message

import (
	"time"

	pkgerrors "wetalk/pkg/errors"
)

// Service 消息业务逻辑
type Service struct{}

// NewService 创建消息服务
func NewService() *Service {
	return &Service{}
}

// FileMetadataPayload 文件元数据载荷（HTTP/WS 请求中传递）
type FileMetadataPayload struct {
	URL          string `json:"url"`
	OriginalName string `json:"original_name"`
	FileSize     int64  `json:"file_size"`
	MimeType     string `json:"mime_type"`
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	ReceiverID   int64                `json:"receiver_id" binding:"required"`
	Content      string               `json:"content" binding:"required"`
	ContentType  string               `json:"content_type"`
	FileMetadata *FileMetadataPayload `json:"file_metadata,omitempty"`
}

// MessageResponse 消息响应（含嵌套文件元数据）
type MessageResponse struct {
	ID           int64         `json:"id"`
	SenderID     int64         `json:"sender_id"`
	ReceiverID   int64         `json:"receiver_id"`
	Content      string        `json:"content"`
	ContentType  string        `json:"content_type"`
	FileMetadata *FileMetadata `json:"file_metadata,omitempty"`
	Status       string        `json:"status"`
	CreatedAt    time.Time     `json:"created_at"`
}

// SendMessage 发送消息
func (s *Service) SendMessage(senderID int64, req SendMessageRequest) (*MessageResponse, error) {
	if senderID == req.ReceiverID {
		return nil, pkgerrors.ErrInvalidParam
	}
	if req.Content == "" {
		return nil, pkgerrors.ErrInvalidParam
	}

	contentType := req.ContentType
	if contentType == "" {
		contentType = ContentTypeText
	}

	// 创建消息
	msg, err := Repository.Create(senderID, req.ReceiverID, req.Content, contentType)
	if err != nil {
		return nil, err
	}

	// 如果有文件元数据，创建 file_metadata 记录
	var fileMeta *FileMetadata
	if req.FileMetadata != nil {
		fm, err := FileMetadataRepository.Create(
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

	return &MessageResponse{
		ID:           msg.ID,
		SenderID:     msg.SenderID,
		ReceiverID:   msg.ReceiverID,
		Content:      msg.Content,
		ContentType:  msg.ContentType,
		FileMetadata: fileMeta,
		Status:       msg.Status,
		CreatedAt:    msg.CreatedAt,
	}, nil
}

// MarkAsRead 标记消息已读
func (s *Service) MarkAsRead(receiverID, senderID int64) error {
	return Repository.MarkAsRead(senderID, receiverID)
}

// GetConversation 获取聊天记录
func (s *Service) GetConversation(userID, friendID int64, offset, limit int) ([]MessageResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	messages, err := Repository.ListByUsers(userID, friendID, offset, limit)
	if err != nil {
		return nil, err
	}

	// 批量查询文件元数据
	var msgIDs []int64
	for _, msg := range messages {
		if msg.ContentType == ContentTypeImage || msg.ContentType == ContentTypeFile {
			msgIDs = append(msgIDs, msg.ID)
		}
	}

	metadataMap := make(map[int64]*FileMetadata)
	if len(msgIDs) > 0 {
		metadataList, err := FileMetadataRepository.ListByMessageIDs(msgIDs)
		if err == nil {
			for i := range metadataList {
				metadataMap[metadataList[i].MessageID] = &metadataList[i]
			}
		}
	}

	// 构建响应
	responses := make([]MessageResponse, len(messages))
	for i, msg := range messages {
		resp := MessageResponse{
			ID:          msg.ID,
			SenderID:    msg.SenderID,
			ReceiverID:  msg.ReceiverID,
			Content:     msg.Content,
			ContentType: msg.ContentType,
			Status:      msg.Status,
			CreatedAt:   msg.CreatedAt,
		}
		if fm, ok := metadataMap[msg.ID]; ok {
			resp.FileMetadata = fm
		}
		responses[i] = resp
	}

	return responses, nil
}
