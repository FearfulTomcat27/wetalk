//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wetalk/dto"
	"wetalk/service"
)

func TestServiceChat_EnsureSingleChat_Create(t *testing.T) {
	resetDB(t)

	alice := CreateUser(t, "alice_"+t.Name())
	bob := CreateUser(t, "bob_"+t.Name())

	chatSvc := service.NewChatService(dto.Chat, cache)
	chat, err := chatSvc.EnsureSingleChat(alice.ID, bob.ID)
	require.NoError(t, err)
	require.NotZero(t, chat.ID)
}

func TestServiceChat_EnsureSingleChat_Existing(t *testing.T) {
	resetDB(t)

	alice, bob, chat := CreateFriendPair(t)

	chatSvc := service.NewChatService(dto.Chat, cache)
	found, err := chatSvc.EnsureSingleChat(alice.ID, bob.ID)
	require.NoError(t, err)
	assert.Equal(t, chat.ID, found.ID)
}

func TestServiceChat_EnsureSingleChat_SameUser(t *testing.T) {
	resetDB(t)

	chatSvc := service.NewChatService(dto.Chat, cache)
	_, err := chatSvc.EnsureSingleChat(1, 1)
	require.Error(t, err)
}

func TestServiceChat_GetMemberIDs(t *testing.T) {
	resetDB(t)

	alice, bob, chat := CreateFriendPair(t)

	chatSvc := service.NewChatService(dto.Chat, cache)
	ids, err := chatSvc.GetMemberIDs(chat.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, []int64{alice.ID, bob.ID}, ids)
}

func TestServiceChat_IsMember(t *testing.T) {
	resetDB(t)

	alice, _, chat := CreateFriendPair(t)
	charles := CreateUser(t, "charles_"+t.Name())

	chatSvc := service.NewChatService(dto.Chat, cache)

	ok, err := chatSvc.IsMember(chat.ID, alice.ID)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = chatSvc.IsMember(chat.ID, charles.ID)
	require.NoError(t, err)
	assert.False(t, ok)
}
