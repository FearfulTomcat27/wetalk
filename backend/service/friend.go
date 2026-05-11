package service

import (
	"wetalk/dto"
	"wetalk/model"
	"wetalk/type"
)

// FriendService 好友业务逻辑
type FriendService struct {
	chatSvc *ChatService
}

// NewFriendService 创建好友服务
func NewFriendService(chatSvc *ChatService) *FriendService {
	return &FriendService{chatSvc: chatSvc}
}

// AddFriend 发送好友请求
func (s *FriendService) AddFriend(userID int64, req model.AddFriendRequest) (*model.FriendRequest, error) {
	if userID == req.FriendID {
		return nil, types.ErrInvalidParam
	}

	// 检查是否已经是好友
	friendship, err := dto.Friend.FindFriendshipBetween(userID, req.FriendID)
	if err != nil {
		return nil, err
	}
	if friendship != nil {
		return nil, types.ErrConflict
	}

	// 检查是否已存在任意方向的好友请求
	existing, err := dto.Friend.FindRequestBetween(userID, req.FriendID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, types.ErrConflict
	}

	return dto.Friend.Create(userID, req.FriendID)
}

// AcceptFriend 接受好友请求（同时写入 friendships 表 + 创建单聊）
func (s *FriendService) AcceptFriend(userID int64, friendReqID int64) error {
	f, err := dto.Friend.FindByID(friendReqID)
	if err != nil {
		return err
	}
	if f == nil {
		return types.ErrNotFound
	}
	// 只有接收方可以接受请求
	if f.FriendID != userID {
		return types.ErrUnauthorized
	}
	if f.Status != model.FriendStatusPending {
		return types.ErrConflict
	}

	if err := dto.Friend.UpdateStatus(friendReqID, model.FriendStatusAccepted); err != nil {
		return err
	}

	// 写入 friendships 表（小 ID 在前，FirstOrCreate 防重）
	friendship, err := dto.Friend.CreateFriendship(f.UserID, f.FriendID)
	if err != nil {
		return err
	}

	// 如果尚未关联聊天，创建单聊并关联
	if friendship.ChatID == 0 {
		chat, err := s.chatSvc.EnsureSingleChat(f.UserID, f.FriendID)
		if err != nil {
			return err
		}
		if err := dto.Friend.UpdateFriendshipChatID(friendship.ID, chat.ID); err != nil {
			return err
		}
	}

	return nil
}

// ListFriends 获取好友列表（从 friendships 表查询）
func (s *FriendService) ListFriends(userID int64) ([]dto.FriendshipInfo, error) {
	return dto.Friend.FindFriendships(userID)
}

// GetPendingRequests 获取待处理好友请求
func (s *FriendService) GetPendingRequests(userID int64) ([]dto.PendingRequest, error) {
	return dto.Friend.FindPendingByUserID(userID)
}

// DeleteFriend 删除好友（同时删除 friend_requests 请求记录和 friendships 关系）
func (s *FriendService) DeleteFriend(userID int64, friendReqID int64) error {
	f, err := dto.Friend.FindByID(friendReqID)
	if err != nil {
		return err
	}
	if f == nil {
		return types.ErrNotFound
	}
	// 只有请求发起方或接收方可以删除
	if f.UserID != userID && f.FriendID != userID {
		return types.ErrUnauthorized
	}

	// 删除 friendships 表中的关系
	if err := dto.Friend.DeleteFriendship(f.UserID, f.FriendID); err != nil {
		return err
	}

	// 删除 friend_requests 表中的请求记录
	return dto.Friend.Delete(friendReqID)
}
