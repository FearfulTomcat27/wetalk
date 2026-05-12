package dto

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"wetalk/db"
	"wetalk/model"
)

// chatRepo 聊天数据仓库
type chatRepo struct{}

// Chat 聊天仓库实例
var Chat = &chatRepo{}

// CreateSingleChat 创建单聊（事务：chat + 两个成员）
func (r *chatRepo) CreateSingleChat(user1ID, user2ID int64) (*model.Chat, error) {
	var chat *model.Chat

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		chat = &model.Chat{
			ChatType: model.ChatTypeSingle,
		}
		if err := tx.Create(chat).Error; err != nil {
			return err
		}

		members := []model.ChatMember{
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
func (r *chatRepo) FindSingleChatByUserIDs(user1ID, user2ID int64) (*model.Chat, error) {
	var chat model.Chat
	err := db.DB.Select("chats.id, chats.chat_type, chats.created_at, chats.updated_at").
		Joins("JOIN chat_members cm1 ON cm1.chat_id = chats.id AND cm1.user_id = ?", user1ID).
		Joins("JOIN chat_members cm2 ON cm2.chat_id = chats.id AND cm2.user_id = ?", user2ID).
		Where("chats.chat_type = ?", model.ChatTypeSingle).
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
func (r *chatRepo) GetMemberIDs(chatID int64) ([]int64, error) {
	var ids []int64
	err := db.DB.Model(&model.ChatMember{}).
		Where("chat_id = ?", chatID).
		Pluck("user_id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// IsMember 检查用户是否为聊天成员
func (r *chatRepo) IsMember(chatID, userID int64) (bool, error) {
	var count int64
	err := db.DB.Model(&model.ChatMember{}).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateLastMessage 更新聊天的最后一条消息信息
func (r *chatRepo) UpdateLastMessage(chatID, messageID int64, content string, contentType string, msgTime time.Time) error {
	return db.DB.Model(&model.Chat{}).
		Where("id = ?", chatID).
		Updates(map[string]interface{}{
			"last_message_id":   messageID,
			"last_message_text": content,
			"last_message_type": contentType,
			"last_message_time": msgTime,
		}).Error
}
