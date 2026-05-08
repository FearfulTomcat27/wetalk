package friend

import "time"

// Status 好友状态常量
const (
	StatusPending  = "pending"
	StatusAccepted = "accepted"
	StatusBlocked  = "blocked"
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
	User1ID   int64     `json:"user1_id" gorm:"column:user1_id;not null;uniqueIndex:uk_user1_user2;index"`
	User2ID   int64     `json:"user2_id" gorm:"column:user2_id;not null;uniqueIndex:uk_user1_user2;index"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

// TableName 指定表名
func (Friendship) TableName() string {
	return "friendships"
}
