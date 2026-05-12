package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"wetalk/model"
	"wetalk/types"
)

func newTestFriendService(friendRepo *MockFriendRepo, userRepo *MockUserRepo, chatSvc *ChatService, cache *MockCache) *FriendService {
	return NewFriendService(friendRepo, userRepo, chatSvc, cache)
}

func TestFriendService_AddFriend_SameUser(t *testing.T) {
	friendRepo := new(MockFriendRepo)
	userRepo := new(MockUserRepo)
	cache := new(MockCache)
	chatRepo := new(MockChatRepo)
	chatSvc := newTestChatService(chatRepo, cache)
	svc := newTestFriendService(friendRepo, userRepo, chatSvc, cache)

	_, err := svc.AddFriend(1, model.AddFriendRequest{FriendID: 1})
	assert.ErrorIs(t, err, types.ErrAddSelf)
}

func TestFriendService_AddFriend_AlreadyFriends(t *testing.T) {
	friendRepo := new(MockFriendRepo)
	userRepo := new(MockUserRepo)
	cache := new(MockCache)
	chatRepo := new(MockChatRepo)
	chatSvc := newTestChatService(chatRepo, cache)
	svc := newTestFriendService(friendRepo, userRepo, chatSvc, cache)

	friendRepo.On("FindFriendshipBetween", int64(1), int64(2)).
		Return(&model.Friendship{ID: 1}, nil)

	_, err := svc.AddFriend(1, model.AddFriendRequest{FriendID: 2})
	assert.ErrorIs(t, err, types.ErrFriendRequestExists)
	friendRepo.AssertExpectations(t)
}

func TestFriendService_AddFriend_ExistingRequest(t *testing.T) {
	friendRepo := new(MockFriendRepo)
	userRepo := new(MockUserRepo)
	cache := new(MockCache)
	chatRepo := new(MockChatRepo)
	chatSvc := newTestChatService(chatRepo, cache)
	svc := newTestFriendService(friendRepo, userRepo, chatSvc, cache)

	friendRepo.On("FindFriendshipBetween", int64(1), int64(2)).Return(nil, nil)
	friendRepo.On("FindRequestBetween", int64(1), int64(2)).
		Return(&model.FriendRequest{ID: 1, UserID: 1, FriendID: 2, Status: model.FriendStatusPending}, nil)

	_, err := svc.AddFriend(1, model.AddFriendRequest{FriendID: 2})
	assert.ErrorIs(t, err, types.ErrFriendRequestExists)
	friendRepo.AssertExpectations(t)
}

func TestFriendService_AddFriend_Success(t *testing.T) {
	friendRepo := new(MockFriendRepo)
	userRepo := new(MockUserRepo)
	cache := new(MockCache)
	chatRepo := new(MockChatRepo)
	chatSvc := newTestChatService(chatRepo, cache)
	svc := newTestFriendService(friendRepo, userRepo, chatSvc, cache)

	now := time.Now()
	friendRepo.On("FindFriendshipBetween", int64(1), int64(2)).Return(nil, nil)
	friendRepo.On("FindRequestBetween", int64(1), int64(2)).Return(nil, nil)
	friendRepo.On("Create", int64(1), int64(2)).Return(&model.FriendRequest{
		ID: 10, UserID: 1, FriendID: 2, Status: model.FriendStatusPending, CreatedAt: now,
	}, nil)
	userRepo.On("FindByID", int64(1)).Return(&model.User{
		ID: 1, Username: "alice", Nickname: "Alice", Avatar: "https://example.com/avatar.png",
	}, nil)

	result, err := svc.AddFriend(1, model.AddFriendRequest{FriendID: 2})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(10), result.Request.ID)
	assert.Equal(t, "Alice", result.Sender.Nickname)

	friendRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestFriendService_AcceptFriend_NotFound(t *testing.T) {
	friendRepo := new(MockFriendRepo)
	userRepo := new(MockUserRepo)
	cache := new(MockCache)
	chatRepo := new(MockChatRepo)
	chatSvc := newTestChatService(chatRepo, cache)
	svc := newTestFriendService(friendRepo, userRepo, chatSvc, cache)

	friendRepo.On("FindByID", int64(99)).Return(nil, nil)

	err := svc.AcceptFriend(2, 99)
	assert.ErrorIs(t, err, types.ErrFriendRequestNotFound)
	friendRepo.AssertExpectations(t)
}

func TestFriendService_AcceptFriend_Forbidden(t *testing.T) {
	friendRepo := new(MockFriendRepo)
	userRepo := new(MockUserRepo)
	cache := new(MockCache)
	chatRepo := new(MockChatRepo)
	chatSvc := newTestChatService(chatRepo, cache)
	svc := newTestFriendService(friendRepo, userRepo, chatSvc, cache)

	friendRepo.On("FindByID", int64(10)).Return(&model.FriendRequest{
		ID: 10, UserID: 1, FriendID: 2, Status: model.FriendStatusPending,
	}, nil)

	// userID=3 既不是发起方也不是接收方
	err := svc.AcceptFriend(3, 10)
	assert.ErrorIs(t, err, types.ErrFriendOpForbidden)
	friendRepo.AssertExpectations(t)
}

func TestFriendService_AcceptFriend_Success(t *testing.T) {
	friendRepo := new(MockFriendRepo)
	userRepo := new(MockUserRepo)
	cache := new(MockCache)
	chatRepo := new(MockChatRepo)
	chatSvc := newTestChatService(chatRepo, cache)
	svc := newTestFriendService(friendRepo, userRepo, chatSvc, cache)

	friendRepo.On("FindByID", int64(10)).Return(&model.FriendRequest{
		ID: 10, UserID: 1, FriendID: 2, Status: model.FriendStatusPending,
	}, nil)
	friendRepo.On("UpdateStatus", int64(10), model.FriendStatusAccepted).Return(nil)
	friendRepo.On("CreateFriendship", int64(1), int64(2), int64(100)).Return(&model.Friendship{
		ID: 1, User1ID: 1, User2ID: 2, ChatID: 100,
	}, nil)

	chatRepo.On("FindSingleChatByUserIDs", int64(1), int64(2)).Return(nil, nil)
	chatRepo.On("CreateSingleChat", int64(1), int64(2)).Return(&model.Chat{ID: 100}, nil)

	// AcceptFriend starts a goroutine for cache invalidation
	cache.On("Del", mock.Anything).Return(nil)

	err := svc.AcceptFriend(2, 10)
	require.NoError(t, err)

	friendRepo.AssertExpectations(t)
	chatRepo.AssertExpectations(t)
	time.Sleep(5 * time.Millisecond) // wait for async cache invalidation goroutine
}

func TestFriendService_ListFriends_CacheHit(t *testing.T) {
	friendRepo := new(MockFriendRepo)
	userRepo := new(MockUserRepo)
	cache := new(MockCache)
	chatRepo := new(MockChatRepo)
	chatSvc := newTestChatService(chatRepo, cache)
	svc := newTestFriendService(friendRepo, userRepo, chatSvc, cache)

	cached := []model.FriendshipInfo{
		{ID: 1, FriendID: 2, FriendName: "Bob"},
	}
	cache.On("Get", "friendships:1", mock.Anything).Return(true, nil, cached)

	friends, err := svc.ListFriends(1)
	require.NoError(t, err)
	assert.Len(t, friends, 1)

	cache.AssertExpectations(t)
	friendRepo.AssertNotCalled(t, "FindFriendships")
}

func TestFriendService_ListFriends_CacheMiss(t *testing.T) {
	friendRepo := new(MockFriendRepo)
	userRepo := new(MockUserRepo)
	cache := new(MockCache)
	chatRepo := new(MockChatRepo)
	chatSvc := newTestChatService(chatRepo, cache)
	svc := newTestFriendService(friendRepo, userRepo, chatSvc, cache)

	friendRepo.On("FindFriendships", int64(1)).Return([]model.FriendshipInfo{
		{ID: 1, FriendID: 2, FriendName: "Bob"},
	}, nil)
	cache.On("Get", "friendships:1", mock.Anything).Return(false, nil)
	cache.On("Set", "friendships:1", mock.Anything, mock.AnythingOfType("time.Duration")).Return(nil)

	friends, err := svc.ListFriends(1)
	require.NoError(t, err)
	assert.Len(t, friends, 1)

	friendRepo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestFriendService_GetPendingRequests(t *testing.T) {
	friendRepo := new(MockFriendRepo)
	userRepo := new(MockUserRepo)
	cache := new(MockCache)
	chatRepo := new(MockChatRepo)
	chatSvc := newTestChatService(chatRepo, cache)
	svc := newTestFriendService(friendRepo, userRepo, chatSvc, cache)

	friendRepo.On("FindPendingByUserID", int64(1)).Return([]model.PendingRequest{
		{ID: 1, User: model.SenderInfo{ID: 2, Username: "alice", Nickname: "Alice"}},
	}, nil)

	requests, err := svc.GetPendingRequests(1)
	require.NoError(t, err)
	assert.Len(t, requests, 1)

	friendRepo.AssertExpectations(t)
}

func TestFriendService_DeleteFriend_NotFound(t *testing.T) {
	friendRepo := new(MockFriendRepo)
	userRepo := new(MockUserRepo)
	cache := new(MockCache)
	chatRepo := new(MockChatRepo)
	chatSvc := newTestChatService(chatRepo, cache)
	svc := newTestFriendService(friendRepo, userRepo, chatSvc, cache)

	friendRepo.On("FindByID", int64(99)).Return(nil, nil)

	err := svc.DeleteFriend(1, 99)
	assert.ErrorIs(t, err, types.ErrFriendshipNotFound)
	friendRepo.AssertExpectations(t)
}

func TestFriendService_DeleteFriend_Forbidden(t *testing.T) {
	friendRepo := new(MockFriendRepo)
	userRepo := new(MockUserRepo)
	cache := new(MockCache)
	chatRepo := new(MockChatRepo)
	chatSvc := newTestChatService(chatRepo, cache)
	svc := newTestFriendService(friendRepo, userRepo, chatSvc, cache)

	friendRepo.On("FindByID", int64(10)).Return(&model.FriendRequest{
		ID: 10, UserID: 1, FriendID: 2, Status: model.FriendStatusAccepted,
	}, nil)

	// userID=3 既不是发起方也不是接收方
	err := svc.DeleteFriend(3, 10)
	assert.ErrorIs(t, err, types.ErrFriendOpForbidden)
	friendRepo.AssertExpectations(t)
}

func TestFriendService_DeleteFriend_Success(t *testing.T) {
	friendRepo := new(MockFriendRepo)
	userRepo := new(MockUserRepo)
	cache := new(MockCache)
	chatRepo := new(MockChatRepo)
	chatSvc := newTestChatService(chatRepo, cache)
	svc := newTestFriendService(friendRepo, userRepo, chatSvc, cache)

	friendRepo.On("FindByID", int64(10)).Return(&model.FriendRequest{
		ID: 10, UserID: 1, FriendID: 2, Status: model.FriendStatusAccepted,
	}, nil)
	friendRepo.On("DeleteFriendship", int64(1), int64(2)).Return(nil)
	friendRepo.On("Delete", int64(10)).Return(nil)

	// DeleteFriend starts a goroutine for cache invalidation
	cache.On("Del", mock.Anything).Return(nil)

	err := svc.DeleteFriend(1, 10)
	require.NoError(t, err)

	friendRepo.AssertExpectations(t)
	time.Sleep(5 * time.Millisecond) // wait for async cache invalidation goroutine
}
