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

// AddFriend 发送好友请求
func (s *Service) AddFriend(userID int64, req AddFriendRequest) (*FriendRequest, error) {
	if userID == req.FriendID {
		return nil, pkgerrors.ErrInvalidParam
	}

	// 检查是否已经是好友
	friendship, err := Repository.FindFriendshipBetween(userID, req.FriendID)
	if err != nil {
		return nil, err
	}
	if friendship != nil {
		return nil, pkgerrors.ErrConflict
	}

	// 检查是否已存在任意方向的好友请求
	existing, err := Repository.FindRequestBetween(userID, req.FriendID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, pkgerrors.ErrConflict
	}

	return Repository.Create(userID, req.FriendID)
}

// AcceptFriend 接受好友请求（同时写入 friendships 表）
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

	if err := Repository.UpdateStatus(friendReqID, StatusAccepted); err != nil {
		return err
	}

	// 写入 friendships 表（小 ID 在前，FirstOrCreate 防重）
	return Repository.CreateFriendship(f.UserID, f.FriendID)
}

// ListFriends 获取好友列表（从 friendships 表查询）
func (s *Service) ListFriends(userID int64) ([]FriendshipInfo, error) {
	return Repository.FindFriendships(userID)
}

// GetPendingRequests 获取待处理好友请求
func (s *Service) GetPendingRequests(userID int64) ([]PendingRequest, error) {
	return Repository.FindPendingByUserID(userID)
}

// DeleteFriend 删除好友（同时删除 friend_requests 请求记录和 friendships 关系）
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

	// 删除 friendships 表中的关系
	if err := Repository.DeleteFriendship(f.UserID, f.FriendID); err != nil {
		return err
	}

	// 删除 friend_requests 表中的请求记录
	return Repository.Delete(friendReqID)
}
