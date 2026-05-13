package service

import (
	"context"
	"io"
	"time"

	"github.com/stretchr/testify/mock"

	"wetalk/dto"
	"wetalk/model"
)

// ========== UserService 测试 Mock ==========

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) FindByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) FindByID(id int64) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) SearchByUsername(keyword string, limit int) ([]model.User, error) {
	args := m.Called(keyword, limit)
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *MockUserRepo) Create(username, passwordHash, nickname, avatar string) (*model.User, error) {
	args := m.Called(username, passwordHash, nickname, avatar)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) UpdateAvatar(userID int64, avatarURL string) error {
	args := m.Called(userID, avatarURL)
	return args.Error(0)
}

// ========== ChatService 测试 Mock ==========

type MockChatRepo struct {
	mock.Mock
}

func (m *MockChatRepo) CreateSingleChat(user1ID, user2ID int64) (*model.Chat, error) {
	args := m.Called(user1ID, user2ID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Chat), args.Error(1)
}

func (m *MockChatRepo) FindSingleChatByUserIDs(user1ID, user2ID int64) (*model.Chat, error) {
	args := m.Called(user1ID, user2ID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Chat), args.Error(1)
}

func (m *MockChatRepo) GetMemberIDs(chatID int64) ([]int64, error) {
	args := m.Called(chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockChatRepo) IsMember(chatID, userID int64) (bool, error) {
	args := m.Called(chatID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockChatRepo) UpdateLastMessage(chatID, messageID int64, content string, contentType string, msgTime time.Time) error {
	args := m.Called(chatID, messageID, content, contentType, msgTime)
	return args.Error(0)
}

// ========== Cache 测试 Mock ==========

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(key string, dest interface{}) (bool, error) {
	args := m.Called(key, dest)
	if args.Bool(0) && len(args) > 2 {
		switch d := dest.(type) {
		case *model.User:
			if u, ok := args.Get(2).(*model.User); ok {
				*d = *u
			}
		case *[]int64:
			if ids, ok := args.Get(2).([]int64); ok {
				*d = ids
			}
		case *[]model.FriendshipInfo:
			if fi, ok := args.Get(2).([]model.FriendshipInfo); ok {
				*d = fi
			}
		case *[]dto.UnreadCount:
			if uc, ok := args.Get(2).([]dto.UnreadCount); ok {
				*d = uc
			}
		}
	}
	return args.Bool(0), args.Error(1)
}

func (m *MockCache) Set(key string, value interface{}, ttl time.Duration) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) Del(keys ...string) error {
	args := m.Called(keys)
	return args.Error(0)
}

// ========== OSS 测试 Mock ==========

type MockOSS struct {
	mock.Mock
}

func (m *MockOSS) PutObject(ctx context.Context, key string, body io.Reader, contentType string, contentLength int64) (string, error) {
	args := m.Called(ctx, key, body, contentType, contentLength)
	return args.String(0), args.Error(1)
}

// ========== FriendService 测试 Mock ==========

type MockFriendRepo struct {
	mock.Mock
}

func (m *MockFriendRepo) Create(userID, friendID int64) (*model.FriendRequest, error) {
	args := m.Called(userID, friendID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.FriendRequest), args.Error(1)
}

func (m *MockFriendRepo) FindByID(id int64) (*model.FriendRequest, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.FriendRequest), args.Error(1)
}

func (m *MockFriendRepo) FindRequestBetween(userID, friendID int64) (*model.FriendRequest, error) {
	args := m.Called(userID, friendID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.FriendRequest), args.Error(1)
}

func (m *MockFriendRepo) FindFriendshipBetween(userID, friendID int64) (*model.Friendship, error) {
	args := m.Called(userID, friendID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Friendship), args.Error(1)
}

func (m *MockFriendRepo) FindFriendships(userID int64, chatOnly ...bool) ([]model.FriendshipInfo, error) {
	allArgs := []interface{}{userID}
	for _, v := range chatOnly {
		allArgs = append(allArgs, v)
	}
	args := m.Called(allArgs...)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.FriendshipInfo), args.Error(1)
}

func (m *MockFriendRepo) FindPendingByUserID(userID int64) ([]model.PendingRequest, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.PendingRequest), args.Error(1)
}

func (m *MockFriendRepo) UpdateStatus(id int64, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockFriendRepo) CreateFriendship(userID, friendID, chatID int64) (*model.Friendship, error) {
	args := m.Called(userID, friendID, chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Friendship), args.Error(1)
}

func (m *MockFriendRepo) DeleteFriendship(userID, friendID int64) error {
	args := m.Called(userID, friendID)
	return args.Error(0)
}

func (m *MockFriendRepo) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

// ========== MessageService 测试 Mock ==========

type MockMessageRepo struct {
	mock.Mock
}

func (m *MockMessageRepo) Create(senderID, chatID int64, content string, contentType string, quoteID *int64, fileMeta *model.FileMetadataPayload) (*model.MessageResponse, error) {
	args := m.Called(senderID, chatID, content, contentType, quoteID, fileMeta)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.MessageResponse), args.Error(1)
}

func (m *MockMessageRepo) GetByID(msgID int64) (*model.Message, error) {
	args := m.Called(msgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Message), args.Error(1)
}

func (m *MockMessageRepo) ListByChat(chatID int64, userID int64, offset, limit int) ([]model.MessageResponse, error) {
	args := m.Called(chatID, userID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.MessageResponse), args.Error(1)
}

func (m *MockMessageRepo) MarkAsRead(chatID, userID int64) error {
	args := m.Called(chatID, userID)
	return args.Error(0)
}

func (m *MockMessageRepo) GetUnreadCounts(userID int64) ([]dto.UnreadCount, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.UnreadCount), args.Error(1)
}
