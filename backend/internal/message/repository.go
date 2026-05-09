package message

import (
	"wetalk/db"
)

// repository 消息数据仓库
type repository struct{}

// Repository 消息仓库实例
var Repository = &repository{}

// Create 创建消息
func (r *repository) Create(senderID, receiverID int64, content string, contentType string) (*Message, error) {
	msg := &Message{
		SenderID:    senderID,
		ReceiverID:  receiverID,
		Content:     content,
		ContentType: contentType,
		Status:      StatusSent,
	}
	if err := db.DB.Create(msg).Error; err != nil {
		return nil, err
	}
	return msg, nil
}

// MarkAsRead 将发送者发给接收者的未读消息标记为已读
func (r *repository) MarkAsRead(senderID, receiverID int64) error {
	return db.DB.Model(&Message{}).
		Where("sender_id = ? AND receiver_id = ? AND status != ?", senderID, receiverID, StatusRead).
		Update("status", StatusRead).Error
}

// ListByUsers 获取两个用户之间的消息记录（分页）
func (r *repository) ListByUsers(userID1, userID2 int64, offset, limit int) ([]Message, error) {
	var messages []Message
	err := db.DB.Where(
		"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		userID1, userID2, userID2, userID1,
	).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}
