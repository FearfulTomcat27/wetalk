package message

import "time"

const (
	// ContentType 常量
	ContentTypeText  = "text"
	ContentTypeImage = "image"
	ContentTypeFile  = "file"

	// Status 常量
	StatusSent      = "sent"
	StatusDelivered = "delivered"
	StatusRead      = "read"
)

// Message 消息模型
type Message struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ChatID      int64     `json:"chat_id" gorm:"column:chat_id;not null;index:idx_chat_created"`
	SenderID    int64     `json:"sender_id" gorm:"column:sender_id;not null;index:idx_chat_created"`
	Content     string    `json:"content" gorm:"column:content;type:text;not null"`
	ContentType string    `json:"content_type" gorm:"column:content_type;type:varchar(16);not null;default:text"`
	Status      string    `json:"status" gorm:"column:status;type:varchar(16);not null;default:sent"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime;index"`
}

// TableName 指定表名
func (Message) TableName() string {
	return "messages"
}
