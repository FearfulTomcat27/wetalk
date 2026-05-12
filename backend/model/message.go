package model

import "time"

const (
	// ContentType 常量
	ContentTypeText  = "text"
	ContentTypeImage = "image"
	ContentTypeFile  = "file"

	// Status 常量
	MessageStatusSent      = "sent"
	MessageStatusDelivered = "delivered"
	MessageStatusRead      = "read"
)

// Message 消息模型（MongoDB 迁移后仅作响应/转换结构体，不再映射 MySQL 表）
type Message struct {
	ID             int64         `json:"id"`
	ChatID         int64         `json:"chat_id"`
	SenderID       int64         `json:"sender_id"`
	Content        string        `json:"content"`
	ContentType    string        `json:"content_type"`
	Status         string        `json:"status"`
	QuoteMessageID *int64        `json:"quoted_message_id"`
	FileMetadata   *FileMetadata `json:"file_metadata,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
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

// UploadResponse 文件上传响应
type UploadResponse struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	FileType    string `json:"file_type"`
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	ChatID         int64                `json:"chat_id" binding:"required"`
	Content        string               `json:"content" binding:"required"`
	ContentType    string               `json:"content_type"`
	QuoteMessageID *int64               `json:"quoted_message_id"`
	FileMetadata   *FileMetadataPayload `json:"file_metadata,omitempty"`
}

// MessageResponse 消息响应（含嵌套文件元数据 + 引用消息）
type MessageResponse struct {
	ID                int64         `json:"id"`
	ChatID            int64         `json:"chat_id"`
	SenderID          int64         `json:"sender_id"`
	Content           string        `json:"content"`
	ContentType       string        `json:"content_type"`
	FileMetadata      *FileMetadata `json:"file_metadata,omitempty"`
	QuoteMessageID    *int64        `json:"quoted_message_id,omitempty"`
	QuotedContent     *string       `json:"quoted_content,omitempty"`
	QuotedSenderID    *int64        `json:"quoted_sender_id,omitempty"`
	QuotedSenderName  *string       `json:"quoted_sender_name,omitempty"`
	QuotedContentType *string       `json:"quoted_content_type,omitempty"`
	QuotedFileMeta    *FileMetadata `json:"quoted_file_metadata,omitempty"`
	Status            string        `json:"status"`
	CreatedAt         time.Time     `json:"created_at"`
}
