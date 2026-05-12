//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wetalk/dto"
	"wetalk/model"
)

func TestDTOFriend_CreateAndFindRequest(t *testing.T) {
	resetDB(t)

	alice := CreateUser(t, "alice_"+t.Name())
	bob := CreateUser(t, "bob_"+t.Name())

	// Create friend request
	fr, err := dto.Friend.Create(alice.ID, bob.ID)
	require.NoError(t, err)
	require.NotZero(t, fr.ID)
	assert.Equal(t, alice.ID, fr.UserID)
	assert.Equal(t, bob.ID, fr.FriendID)
	assert.Equal(t, model.FriendStatusPending, fr.Status)

	// FindByUserAndFriend
	found, err := dto.Friend.FindByUserAndFriend(alice.ID, bob.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, fr.ID, found.ID)

	// FindRequestBetween (reverse direction)
	found2, err := dto.Friend.FindRequestBetween(bob.ID, alice.ID)
	require.NoError(t, err)
	require.NotNil(t, found2)
	assert.Equal(t, fr.ID, found2.ID)
}

func TestDTOFriend_FindRequestBetween_NotFound(t *testing.T) {
	resetDB(t)

	fr, err := dto.Friend.FindRequestBetween(1, 2)
	require.NoError(t, err)
	assert.Nil(t, fr)
}

func TestDTOFriend_FindByUserAndFriend_NotFound(t *testing.T) {
	resetDB(t)

	fr, err := dto.Friend.FindByUserAndFriend(1, 2)
	require.NoError(t, err)
	assert.Nil(t, fr)
}

func TestDTOFriend_UpdateStatus(t *testing.T) {
	resetDB(t)

	alice := CreateUser(t, "alice_"+t.Name())
	bob := CreateUser(t, "bob_"+t.Name())

	fr, err := dto.Friend.Create(alice.ID, bob.ID)
	require.NoError(t, err)

	err = dto.Friend.UpdateStatus(fr.ID, model.FriendStatusAccepted)
	require.NoError(t, err)

	found, err := dto.Friend.FindByID(fr.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, model.FriendStatusAccepted, found.Status)
}

func TestDTOFriend_CreateFriendship(t *testing.T) {
	resetDB(t)

	alice := CreateUser(t, "alice_"+t.Name())
	bob := CreateUser(t, "bob_"+t.Name())
	chat, err := dto.Chat.CreateSingleChat(alice.ID, bob.ID)
	require.NoError(t, err)

	// Create friendship
	fs, err := dto.Friend.CreateFriendship(alice.ID, bob.ID, chat.ID)
	require.NoError(t, err)
	require.NotZero(t, fs.ID)
	assert.Equal(t, chat.ID, fs.ChatID)

	// Verify friendship via FindFriendshipBetween
	found, err := dto.Friend.FindFriendshipBetween(alice.ID, bob.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, chat.ID, found.ChatID)

	// FindFriendships
	friends, err := dto.Friend.FindFriendships(alice.ID)
	require.NoError(t, err)
	assert.Len(t, friends, 1)
	assert.Equal(t, bob.ID, friends[0].FriendID)
}

func TestDTOFriend_FindFriendshipBetween_NotFound(t *testing.T) {
	resetDB(t)

	alice := CreateUser(t, "alice_"+t.Name())
	bob := CreateUser(t, "bob_"+t.Name())

	f, err := dto.Friend.FindFriendshipBetween(alice.ID, bob.ID)
	require.NoError(t, err)
	assert.Nil(t, f)
}

func TestDTOFriend_FindPendingByUserID(t *testing.T) {
	resetDB(t)

	alice := CreateUser(t, "alice_"+t.Name())
	bob := CreateUser(t, "bob_"+t.Name())

	// Alice sends request to Bob
	_, err := dto.Friend.Create(alice.ID, bob.ID)
	require.NoError(t, err)

	// Bob should see pending request
	pending, err := dto.Friend.FindPendingByUserID(bob.ID)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	assert.Equal(t, alice.ID, pending[0].User.ID)

	// Alice should not see any pending (she sent, not received)
	pending2, err := dto.Friend.FindPendingByUserID(alice.ID)
	require.NoError(t, err)
	assert.Len(t, pending2, 0)
}

func TestDTOFriend_FindFriendships_WithMultipleFriends(t *testing.T) {
	resetDB(t)

	alice := CreateUser(t, "alice_"+t.Name())
	bob := CreateUser(t, "bob_"+t.Name())
	charles := CreateUser(t, "charles_"+t.Name())

	// Alice is friends with Bob
	chat1, _ := CreateFriendship(t, alice, bob)
	// Alice is friends with Charles
	chat2, _ := CreateFriendship(t, alice, charles)

	friends, err := dto.Friend.FindFriendships(alice.ID)
	require.NoError(t, err)
	assert.Len(t, friends, 2)

	// Verify both friends are present
	friendIDs := []int64{friends[0].FriendID, friends[1].FriendID}
	assert.ElementsMatch(t, []int64{bob.ID, charles.ID}, friendIDs)
	_ = chat1
	_ = chat2
}
