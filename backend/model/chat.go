package model

import "time"

const (
	// ChatType 常量
	ChatTypeSingle = "private"
	ChatTypeGroup  = "group"
)

// Chat 聊天会话模型
type Chat struct {
	ID              int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	ChatType        string     `json:"chat_type" gorm:"column:chat_type;type:varchar(16);not null;default:private;index"`
	ChatName        string     `json:"chat_name" gorm:"column:chat_name;type:varchar(128);default:null"`
	LastMessageID   *int64     `json:"last_message_id" gorm:"column:last_message_id;default:null"`
	LastMessageText *string    `json:"last_message_text" gorm:"column:last_message_text;type:text;default:null"`
	LastMessageType string     `json:"last_message_type" gorm:"column:last_message_type;type:varchar(16);default:text"`
	LastMessageTime *time.Time `json:"last_message_time" gorm:"column:last_message_time;default:null"`
	CreatedAt       time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名
func (Chat) TableName() string {
	return "chats"
}

// ChatMember 聊天成员模型
type ChatMember struct {
	ID       int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ChatID   int64     `json:"chat_id" gorm:"column:chat_id;not null;uniqueIndex:uk_chat_user;index"`
	UserID   int64     `json:"user_id" gorm:"column:user_id;not null;uniqueIndex:uk_chat_user;index"`
	JoinedAt time.Time `json:"joined_at" gorm:"column:joined_at;autoCreateTime"`
}

// TableName 指定表名
func (ChatMember) TableName() string {
	return "chat_members"
}
