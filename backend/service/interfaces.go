package service

import (
	"context"
	"io"
	"time"

	"wetalk/dto"
	"wetalk/model"
)

// UserRepository 用户数据访问接口
type UserRepository interface {
	FindByUsername(username string) (*model.User, error)
	FindByID(id int64) (*model.User, error)
	SearchByUsername(keyword string, limit int) ([]model.User, error)
	Create(username, passwordHash, nickname, avatar string) (*model.User, error)
	UpdateAvatar(userID int64, avatarURL string) error
}

// ChatRepository 聊天数据访问接口
type ChatRepository interface {
	CreateSingleChat(user1ID, user2ID int64) (*model.Chat, error)
	FindSingleChatByUserIDs(user1ID, user2ID int64) (*model.Chat, error)
	GetMemberIDs(chatID int64) ([]int64, error)
	IsMember(chatID, userID int64) (bool, error)
	UpdateLastMessage(chatID, messageID int64, content string, contentType string, msgTime time.Time) error
}

// FriendRepository 好友数据访问接口
type FriendRepository interface {
	Create(userID, friendID int64) (*model.FriendRequest, error)
	FindByID(id int64) (*model.FriendRequest, error)
	FindRequestBetween(userID, friendID int64) (*model.FriendRequest, error)
	FindFriendshipBetween(userID, friendID int64) (*model.Friendship, error)
	FindFriendships(userID int64, chatOnly ...bool) ([]model.FriendshipInfo, error)
	FindPendingByUserID(userID int64) ([]model.PendingRequest, error)
	UpdateStatus(id int64, status string) error
	CreateFriendship(userID, friendID, chatID int64) (*model.Friendship, error)
	DeleteFriendship(userID, friendID int64) error
	Delete(id int64) error
}

// MessageRepository 消息数据访问接口
type MessageRepository interface {
	Create(senderID, chatID int64, content string, contentType string, quoteID *int64, fileMeta *model.FileMetadataPayload) (*model.MessageResponse, error)
	GetByID(msgID int64) (*model.Message, error)
	ListByChat(chatID int64, userID int64, offset, limit int) ([]model.MessageResponse, error)
	MarkAsRead(chatID, userID int64) error
	GetUnreadCounts(userID int64) ([]dto.UnreadCount, error)
}

// Cache 缓存接口
type Cache interface {
	Get(key string, dest interface{}) (bool, error)
	Set(key string, value interface{}, ttl time.Duration) error
	Del(keys ...string) error
}

// ObjectStorage OSS 对象存储接口
type ObjectStorage interface {
	PutObject(ctx context.Context, key string, body io.Reader, contentType string, contentLength int64) (string, error)
}
