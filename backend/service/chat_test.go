package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"wetalk/model"
	"wetalk/types"
)

// ========== ChatService 测试 ==========

func newTestChatService(chatRepo *MockChatRepo, cache *MockCache) *ChatService {
	return NewChatService(chatRepo, cache)
}

func TestChatService_EnsureSingleChat_SameUser(t *testing.T) {
	repo := new(MockChatRepo)
	cache := new(MockCache)
	svc := newTestChatService(repo, cache)

	_, err := svc.EnsureSingleChat(1, 1)
	assert.ErrorIs(t, err, types.ErrInvalidParam)
}

func TestChatService_EnsureSingleChat_Existing(t *testing.T) {
	repo := new(MockChatRepo)
	cache := new(MockCache)

	existingChat := &model.Chat{ID: 5, ChatType: model.ChatTypeSingle}
	repo.On("FindSingleChatByUserIDs", int64(1), int64(2)).Return(existingChat, nil)

	svc := newTestChatService(repo, cache)

	chat, err := svc.EnsureSingleChat(1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(5), chat.ID)

	repo.AssertExpectations(t)
}

func TestChatService_EnsureSingleChat_CreateNew(t *testing.T) {
	repo := new(MockChatRepo)
	cache := new(MockCache)

	repo.On("FindSingleChatByUserIDs", int64(1), int64(2)).Return(nil, nil)
	repo.On("CreateSingleChat", int64(1), int64(2)).Return(&model.Chat{ID: 10, ChatType: model.ChatTypeSingle}, nil)

	svc := newTestChatService(repo, cache)

	chat, err := svc.EnsureSingleChat(1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(10), chat.ID)

	repo.AssertExpectations(t)
}

func TestChatService_GetMemberIDs_CacheHit(t *testing.T) {
	repo := new(MockChatRepo)
	cache := new(MockCache)

	cache.On("Get", "chat_members:1", mock.Anything).
		Return(true, nil, []int64{1, 2, 3})

	svc := newTestChatService(repo, cache)

	ids, err := svc.GetMemberIDs(1)
	require.NoError(t, err)
	assert.Equal(t, []int64{1, 2, 3}, ids)

	repo.AssertNotCalled(t, "GetMemberIDs")
	cache.AssertExpectations(t)
}

func TestChatService_GetMemberIDs_CacheMiss(t *testing.T) {
	repo := new(MockChatRepo)
	cache := new(MockCache)

	repo.On("GetMemberIDs", int64(1)).Return([]int64{10, 20}, nil)
	cache.On("Get", "chat_members:1", mock.Anything).Return(false, nil)
	cache.On("Set", "chat_members:1", []int64{10, 20}, mock.AnythingOfType("time.Duration")).Return(nil)

	svc := newTestChatService(repo, cache)

	ids, err := svc.GetMemberIDs(1)
	require.NoError(t, err)
	assert.Equal(t, []int64{10, 20}, ids)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestChatService_IsMember_True(t *testing.T) {
	repo := new(MockChatRepo)
	cache := new(MockCache)

	repo.On("GetMemberIDs", int64(1)).Return([]int64{1, 2, 3}, nil)
	cache.On("Get", "chat_members:1", mock.Anything).Return(false, nil)
	cache.On("Set", "chat_members:1", []int64{1, 2, 3}, mock.AnythingOfType("time.Duration")).Return(nil)

	svc := newTestChatService(repo, cache)

	ok, err := svc.IsMember(1, 2)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestChatService_IsMember_False(t *testing.T) {
	repo := new(MockChatRepo)
	cache := new(MockCache)

	repo.On("GetMemberIDs", int64(1)).Return([]int64{1, 3}, nil)
	cache.On("Get", "chat_members:1", mock.Anything).Return(false, nil)
	cache.On("Set", "chat_members:1", []int64{1, 3}, mock.AnythingOfType("time.Duration")).Return(nil)

	svc := newTestChatService(repo, cache)

	ok, err := svc.IsMember(1, 2)
	require.NoError(t, err)
	assert.False(t, ok)
}
