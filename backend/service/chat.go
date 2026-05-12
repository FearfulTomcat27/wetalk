package service

import (
	"fmt"
	"time"

	"wetalk/common"
	"wetalk/dto"
	"wetalk/model"
	"wetalk/types"
)

// ChatService 聊天业务逻辑
type ChatService struct{}

// NewChatService 创建聊天服务
func NewChatService() *ChatService {
	return &ChatService{}
}

// EnsureSingleChat 获取或创建两个用户的单聊
// 先查现有聊天，不存在则创建
func (s *ChatService) EnsureSingleChat(user1ID, user2ID int64) (*model.Chat, error) {
	if user1ID == user2ID {
		return nil, types.ErrInvalidParam
	}

	// 先查找已有聊天
	chat, err := dto.Chat.FindSingleChatByUserIDs(user1ID, user2ID)
	if err != nil {
		return nil, err
	}
	if chat != nil {
		return chat, nil
	}

	// 不存在则创建
	return dto.Chat.CreateSingleChat(user1ID, user2ID)
}

// GetMemberIDs 获取聊天成员 ID 列表
func (s *ChatService) GetMemberIDs(chatID int64) ([]int64, error) {
	key := fmt.Sprintf("chat_members:%d", chatID)
	var ids []int64
	if ok, _ := common.CacheGet(key, &ids); ok {
		return ids, nil
	}

	ids, err := dto.Chat.GetMemberIDs(chatID)
	if err != nil {
		return nil, err
	}

	if cacheErr := common.CacheSet(key, ids, 1*time.Hour); cacheErr != nil {
		_ = cacheErr
	}
	return ids, nil
}

// IsMember 检查用户是否为聊天成员
func (s *ChatService) IsMember(chatID, userID int64) (bool, error) {
	ids, err := s.GetMemberIDs(chatID)
	if err != nil {
		return false, err
	}
	for _, id := range ids {
		if id == userID {
			return true, nil
		}
	}
	return false, nil
}

// UpdateLastMessage 更新聊天的最后一条消息
func (s *ChatService) UpdateLastMessage(chatID, messageID int64, content string, contentType string, msgTime time.Time) error {
	return dto.Chat.UpdateLastMessage(chatID, messageID, content, contentType, msgTime)
}
