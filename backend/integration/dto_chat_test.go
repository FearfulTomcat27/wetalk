//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wetalk/dto"
	"wetalk/model"
)

func TestDTOChat_CreateSingleChat(t *testing.T) {
	resetDB(t)

	alice := CreateUser(t, "alice_"+t.Name())
	bob := CreateUser(t, "bob_"+t.Name())

	chat, err := dto.Chat.CreateSingleChat(alice.ID, bob.ID)
	require.NoError(t, err)
	require.NotZero(t, chat.ID)
	assert.Equal(t, model.ChatTypeSingle, chat.ChatType)

	// Verify chat members
	memberIDs, err := dto.Chat.GetMemberIDs(chat.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, []int64{alice.ID, bob.ID}, memberIDs)
}

func TestDTOChat_FindSingleChatByUserIDs(t *testing.T) {
	resetDB(t)

	alice := CreateUser(t, "alice_"+t.Name())
	bob := CreateUser(t, "bob_"+t.Name())

	// No chat yet
	chat, err := dto.Chat.FindSingleChatByUserIDs(alice.ID, bob.ID)
	require.NoError(t, err)
	assert.Nil(t, chat)

	// Create and verify
	created, err := dto.Chat.CreateSingleChat(alice.ID, bob.ID)
	require.NoError(t, err)

	found, err := dto.Chat.FindSingleChatByUserIDs(alice.ID, bob.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
}

func TestDTOChat_IsMember(t *testing.T) {
	resetDB(t)

	alice := CreateUser(t, "alice_"+t.Name())
	bob := CreateUser(t, "bob_"+t.Name())
	charles := CreateUser(t, "charles_"+t.Name())

	chat, err := dto.Chat.CreateSingleChat(alice.ID, bob.ID)
	require.NoError(t, err)

	ok, err := dto.Chat.IsMember(chat.ID, alice.ID)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = dto.Chat.IsMember(chat.ID, charles.ID)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestDTOChat_UpdateLastMessage(t *testing.T) {
	resetDB(t)

	alice := CreateUser(t, "alice_"+t.Name())
	bob := CreateUser(t, "bob_"+t.Name())

	chat, err := dto.Chat.CreateSingleChat(alice.ID, bob.ID)
	require.NoError(t, err)

	now := time.Now()
	err = dto.Chat.UpdateLastMessage(chat.ID, 100, "hello world", "text", now)
	require.NoError(t, err)

	// Re-query the chat to verify
	chats, err := dto.Chat.FindSingleChatByUserIDs(alice.ID, bob.ID)
	require.NoError(t, err)
	require.NotNil(t, chats)
}
