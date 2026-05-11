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

// Message 消息模型
type Message struct {
	ID             int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ChatID         int64     `json:"chat_id" gorm:"column:chat_id;not null;index:idx_chat_created"`
	SenderID       int64     `json:"sender_id" gorm:"column:sender_id;not null;index:idx_chat_created"`
	Content        string    `json:"content" gorm:"column:content;type:text;not null"`
	ContentType    string    `json:"content_type" gorm:"column:content_type;type:varchar(16);not null;default:text"`
	Status         string    `json:"status" gorm:"column:status;type:varchar(16);not null;default:sent"`
	QuoteMessageID *int64    `json:"quoted_message_id" gorm:"column:quote_id;default:null"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime;index"`
}

// TableName 指定表名
func (Message) TableName() string {
	return "messages"
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
	ID             int64         `json:"id"`
	ChatID         int64         `json:"chat_id"`
	SenderID       int64         `json:"sender_id"`
	Content        string        `json:"content"`
	ContentType    string        `json:"content_type"`
	FileMetadata   *FileMetadata `json:"file_metadata,omitempty"`
	QuoteMessageID *int64        `json:"quoted_message_id,omitempty"`
	QuotedContent  *string       `json:"quoted_content,omitempty"`
	Status         string        `json:"status"`
	CreatedAt      time.Time     `json:"created_at"`
}
