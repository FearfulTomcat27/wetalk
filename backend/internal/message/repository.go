package message

import (
	"wetalk/db"
)

// repository 消息数据仓库
type repository struct{}

// Repository 消息仓库实例
var Repository = &repository{}

// Create 创建消息
func (r *repository) Create(senderID, receiverID int64, content string) (*Message, error) {
	result, err := db.DB.Exec(
		"INSERT INTO messages (sender_id, receiver_id, content) VALUES (?, ?, ?)",
		senderID, receiverID, content,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &Message{
		ID:         id,
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
	}, nil
}

// ListByUsers 获取两个用户之间的消息记录（分页）
func (r *repository) ListByUsers(userID1, userID2 int64, offset, limit int) ([]Message, error) {
	rows, err := db.DB.Query(
		`SELECT id, sender_id, receiver_id, content, created_at
		 FROM messages
		 WHERE (sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		userID1, userID2, userID2, userID1, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}
