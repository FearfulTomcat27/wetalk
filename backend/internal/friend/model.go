package friend

import "time"

// Status 好友状态常量
const (
	StatusPending  = "pending"
	StatusAccepted = "accepted"
	StatusBlocked  = "blocked"
)

// Friend 好友关系模型
type Friend struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	FriendID  int64     `json:"friend_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
