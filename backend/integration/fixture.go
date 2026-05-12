//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"wetalk/dto"
	"wetalk/model"
)

// ========== 测试数据生成器 ==========

// CreateUser 通过 DTO 层直接创建用户（不走 HTTP）
func CreateUser(t *testing.T, username string) *model.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)
	avatar := "https://api.dicebear.com/9.x/micah/svg?seed=" + username
	user, err := dto.User.Create(username, string(hash), username, avatar)
	require.NoError(t, err)
	require.NotZero(t, user.ID)
	return user
}

// CreateChat 创建单聊并返回聊天
func CreateChat(t *testing.T, userIDs ...int64) *model.Chat {
	t.Helper()
	require.Len(t, userIDs, 2, "CreateChat 需要恰好两个 userID")
	chat, err := dto.Chat.CreateSingleChat(userIDs[0], userIDs[1])
	require.NoError(t, err)
	require.NotZero(t, chat.ID)
	return chat
}

// CreateFriendship 创建好友关系并返回聊天和好友关系
func CreateFriendship(t *testing.T, user1, user2 *model.User) (*model.Chat, *model.Friendship) {
	t.Helper()
	chat := CreateChat(t, user1.ID, user2.ID)
	fs, err := dto.Friend.CreateFriendship(user1.ID, user2.ID, chat.ID)
	require.NoError(t, err)
	return chat, fs
}

// CreateMessage 创建消息并返回消息响应
func CreateMessage(t *testing.T, senderID, chatID int64, content string) *model.MessageResponse {
	t.Helper()
	msg, err := dto.Message.Create(senderID, chatID, content, model.ContentTypeText, nil, nil)
	require.NoError(t, err)
	require.NotZero(t, msg.ID)
	return msg
}

// CreateFriendPair 创建两个用户并建立好友关系
func CreateFriendPair(t *testing.T) (alice, bob *model.User, chat *model.Chat) {
	t.Helper()
	alice = CreateUser(t, "alice_"+t.Name())
	bob = CreateUser(t, "bob_"+t.Name())
	chat, _ = CreateFriendship(t, alice, bob)
	return
}
