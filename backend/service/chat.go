package service

import (
	"time"

	"wetalk/dto"
	"wetalk/model"
	"wetalk/type"
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
	return dto.Chat.GetMemberIDs(chatID)
}

// IsMember 检查用户是否为聊天成员
func (s *ChatService) IsMember(chatID, userID int64) (bool, error) {
	return dto.Chat.IsMember(chatID, userID)
}

// UpdateLastMessage 更新聊天的最后一条消息
func (s *ChatService) UpdateLastMessage(chatID, messageID int64, content string, msgTime time.Time) error {
	return dto.Chat.UpdateLastMessage(chatID, messageID, content, msgTime)
}
