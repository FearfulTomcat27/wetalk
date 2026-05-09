package chat

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"wetalk/db"
)

// repository 聊天数据仓库
type repository struct{}

// Repository 聊天仓库实例
var Repository = &repository{}

// CreateSingleChat 创建单聊（事务：chat + 两个成员）
func (r *repository) CreateSingleChat(user1ID, user2ID int64) (*Chat, error) {
	var chat *Chat

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		chat = &Chat{
			ChatType: ChatTypeSingle,
		}
		if err := tx.Create(chat).Error; err != nil {
			return err
		}

		members := []ChatMember{
			{ChatID: chat.ID, UserID: user1ID},
			{ChatID: chat.ID, UserID: user2ID},
		}
		if err := tx.Create(&members).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return chat, nil
}

// FindSingleChatByUserIDs 查找两个用户的单聊（两成员交集查询）
func (r *repository) FindSingleChatByUserIDs(user1ID, user2ID int64) (*Chat, error) {
	var chat Chat
	err := db.DB.
		Joins("JOIN chat_members cm1 ON cm1.chat_id = chats.id AND cm1.user_id = ?", user1ID).
		Joins("JOIN chat_members cm2 ON cm2.chat_id = chats.id AND cm2.user_id = ?", user2ID).
		Where("chats.chat_type = ?", ChatTypeSingle).
		First(&chat).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

// GetMemberIDs 获取聊天所有成员 ID
func (r *repository) GetMemberIDs(chatID int64) ([]int64, error) {
	var ids []int64
	err := db.DB.Model(&ChatMember{}).
		Where("chat_id = ?", chatID).
		Pluck("user_id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// IsMember 检查用户是否为聊天成员
func (r *repository) IsMember(chatID, userID int64) (bool, error) {
	var count int64
	err := db.DB.Model(&ChatMember{}).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateLastMessage 更新聊天的最后一条消息信息
func (r *repository) UpdateLastMessage(chatID, messageID int64, content string, msgTime time.Time) error {
	return db.DB.Model(&Chat{}).
		Where("id = ?", chatID).
		Updates(map[string]interface{}{
			"last_message_id":   messageID,
			"last_message_text": content,
			"last_message_time": msgTime,
		}).Error
}
