package service

import (
	"fmt"
	"time"

	"wetalk/model"
	"wetalk/types"
)

// FriendService 好友业务逻辑
type FriendService struct {
	friendRepo FriendRepository
	userRepo   UserRepository
	chatSvc    *ChatService
	cache      Cache
}

// NewFriendService 创建好友服务
func NewFriendService(friendRepo FriendRepository, userRepo UserRepository, chatSvc *ChatService, cache Cache) *FriendService {
	return &FriendService{
		friendRepo: friendRepo,
		userRepo:   userRepo,
		chatSvc:    chatSvc,
		cache:      cache,
	}
}

// AddFriendResult 添加好友结果（含请求记录和发送者信息）
type AddFriendResult struct {
	Request   *model.FriendRequest
	Sender    model.SenderInfo
	CreatedAt string
}

func friendshipCacheKeys(userID int64) []string {
	return []string{
		fmt.Sprintf("friendships:%d", userID),
		fmt.Sprintf("friendships:%d:chatted", userID),
	}
}

// invalidateFriendshipCache 使两个用户的好友列表缓存失效
func (s *FriendService) invalidateFriendshipCache(userID, friendID int64) {
	keys := append(friendshipCacheKeys(userID), friendshipCacheKeys(friendID)...)
	_ = s.cache.Del(keys...)
}

// AddFriend 发送好友请求
func (s *FriendService) AddFriend(userID int64, req model.AddFriendRequest) (*AddFriendResult, error) {
	if userID == req.FriendID {
		return nil, types.ErrAddSelf
	}

	// 检查是否已经是好友
	friendship, err := s.friendRepo.FindFriendshipBetween(userID, req.FriendID)
	if err != nil {
		return nil, err
	}
	if friendship != nil {
		return nil, types.ErrFriendRequestExists
	}

	// 检查是否已存在任意方向的好友请求
	existing, err := s.friendRepo.FindRequestBetween(userID, req.FriendID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, types.ErrFriendRequestExists
	}

	fr, err := s.friendRepo.Create(userID, req.FriendID)
	if err != nil {
		return nil, err
	}

	// 获取发送者信息
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	return &AddFriendResult{
		Request: fr,
		Sender: model.SenderInfo{
			ID:       user.ID,
			Username: user.Username,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
		},
		CreatedAt: fr.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// AcceptFriend 接受好友请求（同时写入 friendships 表 + 创建单聊）
func (s *FriendService) AcceptFriend(userID int64, friendReqID int64) error {
	f, err := s.friendRepo.FindByID(friendReqID)
	if err != nil {
		return err
	}
	if f == nil {
		return types.ErrFriendRequestNotFound
	}
	// 只有接收方可以接受请求
	if f.FriendID != userID {
		return types.ErrFriendOpForbidden
	}
	if f.Status != model.FriendStatusPending {
		return types.ErrRequestAlreadyHandled
	}

	if err := s.friendRepo.UpdateStatus(friendReqID, model.FriendStatusAccepted); err != nil {
		return err
	}

	// 先创建单聊
	chat, err := s.chatSvc.EnsureSingleChat(f.UserID, f.FriendID)
	if err != nil {
		return err
	}

	// 再写入 friendships（含 chat_id，FirstOrCreate 防重）
	_, err = s.friendRepo.CreateFriendship(f.UserID, f.FriendID, chat.ID)
	if err != nil {
		return err
	}

	// 异步使双方好友列表缓存失效
	go s.invalidateFriendshipCache(f.UserID, f.FriendID)
	return nil
}

// ListFriends 获取好友列表（从 friendships 表查询）
// chatOnly 为 true 时只返回有消息记录的好友
func (s *FriendService) ListFriends(userID int64, chatOnly ...bool) ([]model.FriendshipInfo, error) {
	key := fmt.Sprintf("friendships:%d", userID)
	if len(chatOnly) > 0 && chatOnly[0] {
		key += ":chatted"
	}

	var friends []model.FriendshipInfo
	if ok, _ := s.cache.Get(key, &friends); ok {
		return friends, nil
	}

	var err error
	friends, err = s.friendRepo.FindFriendships(userID, chatOnly...)
	if err != nil {
		return nil, err
	}
	if friends == nil {
		friends = []model.FriendshipInfo{}
	}

	if cacheErr := s.cache.Set(key, friends, 5*time.Minute); cacheErr != nil {
		_ = cacheErr
	}
	return friends, nil
}

// GetPendingRequests 获取待处理好友请求
func (s *FriendService) GetPendingRequests(userID int64) ([]model.PendingRequest, error) {
	requests, err := s.friendRepo.FindPendingByUserID(userID)
	if err != nil {
		return nil, err
	}
	if requests == nil {
		return []model.PendingRequest{}, nil
	}
	return requests, nil
}

// DeleteFriend 删除好友（同时删除 friend_requests 请求记录和 friendships 关系）
func (s *FriendService) DeleteFriend(userID int64, friendReqID int64) error {
	f, err := s.friendRepo.FindByID(friendReqID)
	if err != nil {
		return err
	}
	if f == nil {
		return types.ErrFriendshipNotFound
	}
	// 只有请求发起方或接收方可以删除
	if f.UserID != userID && f.FriendID != userID {
		return types.ErrFriendOpForbidden
	}

	// 删除 friendships 表中的关系
	if err := s.friendRepo.DeleteFriendship(f.UserID, f.FriendID); err != nil {
		return err
	}

	// 删除 friend_requests 表中的请求记录
	if err := s.friendRepo.Delete(friendReqID); err != nil {
		return err
	}

	// 异步使双方好友列表缓存失效
	go s.invalidateFriendshipCache(f.UserID, f.FriendID)
	return nil
}
