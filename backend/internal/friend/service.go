package friend

import (
	pkgerrors "wetalk/pkg/errors"
)

// Service 好友业务逻辑
type Service struct{}

// NewService 创建好友服务
func NewService() *Service {
	return &Service{}
}

// AddFriendRequest 添加好友请求
type AddFriendRequest struct {
	FriendID int64 `json:"friend_id" binding:"required"`
}

// FriendInfo 好友信息（含对方用户资料）
type FriendInfo struct {
	ID         int64  `json:"id"`
	UserID     int64  `json:"user_id"`
	FriendID   int64  `json:"friend_id"`
	Status     string `json:"status"`
	FriendName string `json:"friend_name"`
	CreatedAt  string `json:"created_at"`
}

// AddFriend 发送好友请求
func (s *Service) AddFriend(userID int64, req AddFriendRequest) (*Friend, error) {
	if userID == req.FriendID {
		return nil, pkgerrors.ErrInvalidParam
	}

	// 检查是否已存在
	existing, err := Repository.FindByUserAndFriend(userID, req.FriendID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, pkgerrors.ErrConflict
	}

	return Repository.Create(userID, req.FriendID)
}

// AcceptFriend 接受好友请求
func (s *Service) AcceptFriend(userID int64, friendReqID int64) error {
	f, err := Repository.FindByID(friendReqID)
	if err != nil {
		return err
	}
	if f == nil {
		return pkgerrors.ErrNotFound
	}
	// 只有接收方可以接受请求
	if f.FriendID != userID {
		return pkgerrors.ErrUnauthorized
	}
	if f.Status != StatusPending {
		return pkgerrors.ErrConflict
	}

	return Repository.UpdateStatus(friendReqID, StatusAccepted)
}

// ListFriends 获取好友列表
func (s *Service) ListFriends(userID int64) ([]Friend, error) {
	return Repository.ListByUserID(userID)
}

// DeleteFriend 删除好友
func (s *Service) DeleteFriend(userID int64, friendReqID int64) error {
	f, err := Repository.FindByID(friendReqID)
	if err != nil {
		return err
	}
	if f == nil {
		return pkgerrors.ErrNotFound
	}
	// 只有请求发起方或接收方可以删除
	if f.UserID != userID && f.FriendID != userID {
		return pkgerrors.ErrUnauthorized
	}

	return Repository.Delete(friendReqID)
}
