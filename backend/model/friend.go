package model

import "time"

// Status 好友状态常量
const (
	FriendStatusPending  = "pending"
	FriendStatusAccepted = "accepted"
	FriendStatusBlocked  = "blocked"
)

// FriendRequest 好友请求模型
type FriendRequest struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    int64     `json:"user_id" gorm:"column:user_id;not null;index"`
	FriendID  int64     `json:"friend_id" gorm:"column:friend_id;not null;index"`
	Status    string    `json:"status" gorm:"column:status;type:varchar(16);not null;default:pending;index"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名
func (FriendRequest) TableName() string {
	return "friend_requests"
}

// Friendship 已建立的好友关系模型
type Friendship struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ChatID    int64     `json:"chat_id" gorm:"column:chat_id;not null;index"`
	User1ID   int64     `json:"user1_id" gorm:"column:user1_id;not null;uniqueIndex:uk_user1_user2;index"`
	User2ID   int64     `json:"user2_id" gorm:"column:user2_id;not null;uniqueIndex:uk_user1_user2;index"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

// TableName 指定表名
func (Friendship) TableName() string {
	return "friendships"
}

// AddFriendRequest 添加好友请求
type AddFriendRequest struct {
	FriendID int64 `json:"friend_id" binding:"required"`
}

// ========== 好友列表/待处理请求 响应类型 ==========

// SenderInfo 发送者信息
type SenderInfo struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// PendingRequest 待处理好友请求（含发送者信息）
type PendingRequest struct {
	ID        int64      `json:"id"`
	Status    string     `json:"status"`
	User      SenderInfo `json:"user"`
	CreatedAt string     `json:"created_at"`
}

// FriendshipInfo 好友信息（从 friendships 表查询，JOIN users + LEFT JOIN chats 获取信息）
type FriendshipInfo struct {
	ID              int64  `json:"id"`
	ChatID          int64  `json:"chat_id"`
	FriendID        int64  `json:"friend_id"`
	FriendName      string `json:"friend_name"`
	FriendAvatar    string `json:"friend_avatar"`
	LastMessage     string `json:"last_message"`
	LastMessageType string `json:"last_message_type"`
	LastMessageTime string `json:"last_message_time"`
	UnreadCount     int    `json:"unread_count"`
	CreatedAt       string `json:"created_at"`
}
