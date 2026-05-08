package message

import pkgerrors "wetalk/pkg/errors"

// Service 消息业务逻辑
type Service struct{}

// NewService 创建消息服务
func NewService() *Service {
	return &Service{}
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	ReceiverID int64  `json:"receiver_id" binding:"required"`
	Content    string `json:"content" binding:"required"`
}

// SendMessage 发送消息
func (s *Service) SendMessage(senderID int64, req SendMessageRequest) (*Message, error) {
	if senderID == req.ReceiverID {
		return nil, pkgerrors.ErrInvalidParam
	}
	if req.Content == "" {
		return nil, pkgerrors.ErrInvalidParam
	}

	return Repository.Create(senderID, req.ReceiverID, req.Content)
}

// MarkAsRead 标记消息已读
func (s *Service) MarkAsRead(receiverID, senderID int64) error {
	return Repository.MarkAsRead(senderID, receiverID)
}

// GetConversation 获取聊天记录
func (s *Service) GetConversation(userID, friendID int64, offset, limit int) ([]Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return Repository.ListByUsers(userID, friendID, offset, limit)
}
