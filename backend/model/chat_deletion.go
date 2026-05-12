package model

import "time"

// ChatDeletion MongoDB 单方删除聊天记录的时间戳记录
type ChatDeletion struct {
	UserID    int64     `bson:"user_id"`
	ChatID    int64     `bson:"chat_id"`
	DeletedAt time.Time `bson:"deleted_at"`
}
