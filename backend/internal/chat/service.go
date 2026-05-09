package chat

import (
	"time"

	pkgerrors "wetalk/pkg/errors"
)

// Service 聊天业务逻辑
type Service struct{}

// NewService 创建聊天服务
func NewService() *Service {
	return &Service{}
}

// EnsureSingleChat 获取或创建两个用户的单聊
// 先查现有聊天，不存在则创建
func (s *Service) EnsureSingleChat(user1ID, user2ID int64) (*Chat, error) {
	if user1ID == user2ID {
		return nil, pkgerrors.ErrInvalidParam
	}

	// 先查找已有聊天
	chat, err := Repository.FindSingleChatByUserIDs(user1ID, user2ID)
	if err != nil {
		return nil, err
	}
	if chat != nil {
		return chat, nil
	}

	// 不存在则创建
	return Repository.CreateSingleChat(user1ID, user2ID)
}

// GetMemberIDs 获取聊天成员 ID 列表
func (s *Service) GetMemberIDs(chatID int64) ([]int64, error) {
	return Repository.GetMemberIDs(chatID)
}

// IsMember 检查用户是否为聊天成员
func (s *Service) IsMember(chatID, userID int64) (bool, error) {
	return Repository.IsMember(chatID, userID)
}

// UpdateLastMessage 更新聊天的最后一条消息
func (s *Service) UpdateLastMessage(chatID, messageID int64, content string, msgTime time.Time) error {
	return Repository.UpdateLastMessage(chatID, messageID, content, msgTime)
}
