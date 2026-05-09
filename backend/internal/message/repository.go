package message

import (
	"wetalk/db"
)

// repository 消息数据仓库
type repository struct{}

// Repository 消息仓库实例
var Repository = &repository{}

// Create 创建消息
func (r *repository) Create(senderID, chatID int64, content string, contentType string) (*Message, error) {
	msg := &Message{
		SenderID:    senderID,
		ChatID:      chatID,
		Content:     content,
		ContentType: contentType,
		Status:      StatusSent,
	}
	if err := db.DB.Create(msg).Error; err != nil {
		return nil, err
	}
	return msg, nil
}

// MarkAsRead 将聊天中对方发来的未读消息标记为已读
func (r *repository) MarkAsRead(chatID, currentUserID int64) error {
	return db.DB.Model(&Message{}).
		Where("chat_id = ? AND sender_id != ? AND status != ?", chatID, currentUserID, StatusRead).
		Update("status", StatusRead).Error
}

// ListByChat 获取聊天消息记录（分页，按时间倒序）
func (r *repository) ListByChat(chatID int64, offset, limit int) ([]Message, error) {
	var messages []Message
	err := db.DB.Where("chat_id = ?", chatID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// CountUnread 统计聊天中当前用户未读的消息数
func (r *repository) CountUnread(chatID, currentUserID int64) (int64, error) {
	var count int64
	err := db.DB.Model(&Message{}).
		Where("chat_id = ? AND sender_id != ? AND status != ?", chatID, currentUserID, StatusRead).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
