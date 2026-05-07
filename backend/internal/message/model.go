package message

import "time"

// Message 消息模型
type Message struct {
	ID         int64     `json:"id"`
	SenderID   int64     `json:"sender_id"`
	ReceiverID int64     `json:"receiver_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}
