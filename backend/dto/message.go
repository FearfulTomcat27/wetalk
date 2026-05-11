package dto

import (
	"errors"

	"gorm.io/gorm"

	"wetalk/db"
	"wetalk/model"
)

// messageRepo 消息数据仓库
type messageRepo struct{}

// Message 消息仓库实例
var Message = &messageRepo{}

// Create 创建消息
func (r *messageRepo) Create(senderID, chatID int64, content string, contentType string, quoteID *int64) (*model.Message, error) {
	msg := &model.Message{
		SenderID:       senderID,
		ChatID:         chatID,
		Content:        content,
		ContentType:    contentType,
		Status:         model.MessageStatusSent,
		QuoteMessageID: quoteID,
	}
	if err := db.DB.Create(msg).Error; err != nil {
		return nil, err
	}
	return msg, nil
}

// GetByID 根据 ID 查询单条消息
func (r *messageRepo) GetByID(id int64) (*model.Message, error) {
	var msg model.Message
	err := db.DB.Where("id = ?", id).First(&msg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// ListByIDs 批量查询消息（用于引用消息查找）
func (r *messageRepo) ListByIDs(ids []int64) ([]model.Message, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var messages []model.Message
	err := db.DB.Where("id IN ?", ids).Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// MarkAsRead 将聊天中对方发来的未读消息标记为已读
func (r *messageRepo) MarkAsRead(chatID, currentUserID int64) error {
	return db.DB.Model(&model.Message{}).
		Where("chat_id = ? AND sender_id != ? AND status != ?", chatID, currentUserID, model.MessageStatusRead).
		Update("status", model.MessageStatusRead).Error
}

// ListByChat 获取聊天消息记录（分页，按时间倒序）
func (r *messageRepo) ListByChat(chatID int64, offset, limit int) ([]model.Message, error) {
	var messages []model.Message
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
func (r *messageRepo) CountUnread(chatID, currentUserID int64) (int64, error) {
	var count int64
	err := db.DB.Model(&model.Message{}).
		Where("chat_id = ? AND sender_id != ? AND status != ?", chatID, currentUserID, model.MessageStatusRead).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// UnreadCount 单个聊天的未读数
type UnreadCount struct {
	ChatID int64 `json:"chat_id"`
	Count  int64 `json:"count"`
}

// GetUnreadCounts 查询当前用户在所有聊天中的未读消息数（轻量级，供侧边栏角标使用）
func (r *messageRepo) GetUnreadCounts(userID int64) ([]UnreadCount, error) {
	var counts []UnreadCount
	err := db.DB.Raw(`
		SELECT m.chat_id, COUNT(*) AS count
		FROM messages m
		JOIN friendships f ON f.chat_id = m.chat_id
		WHERE (f.user1_id = ? OR f.user2_id = ?)
		  AND m.sender_id != ?
		  AND m.status != ?
		GROUP BY m.chat_id
	`, userID, userID, userID, model.MessageStatusRead).Scan(&counts).Error
	if err != nil {
		return nil, err
	}
	return counts, nil
}
