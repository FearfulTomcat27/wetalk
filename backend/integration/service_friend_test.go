//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wetalk/dto"
	"wetalk/model"
	"wetalk/service"
)

func TestServiceFriend_AddAndAccept(t *testing.T) {
	resetDB(t)

	alice := CreateUser(t, "alice_"+t.Name())
	bob := CreateUser(t, "bob_"+t.Name())

	chatSvc := service.NewChatService(dto.Chat, cache)
	friendSvc := service.NewFriendService(dto.Friend, dto.User, chatSvc, cache)

	// Alice sends friend request to Bob
	result, err := friendSvc.AddFriend(alice.ID, model.AddFriendRequest{FriendID: bob.ID})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, alice.ID, result.Sender.ID)
	assert.Equal(t, model.FriendStatusPending, result.Request.Status)

	// Bob accepts
	err = friendSvc.AcceptFriend(bob.ID, result.Request.ID)
	require.NoError(t, err)

	// Verify friendship exists
	friends, err := friendSvc.ListFriends(bob.ID)
	require.NoError(t, err)
	assert.Len(t, friends, 1)
	assert.Equal(t, alice.ID, friends[0].FriendID)
}

func TestServiceFriend_AddFriend_Duplicate(t *testing.T) {
	resetDB(t)

	alice, bob, _ := CreateFriendPair(t)

	chatSvc := service.NewChatService(dto.Chat, cache)
	friendSvc := service.NewFriendService(dto.Friend, dto.User, chatSvc, cache)

	_, err := friendSvc.AddFriend(alice.ID, model.AddFriendRequest{FriendID: bob.ID})
	require.Error(t, err)
	assert.Equal(t, "[friend_request_exists] 好友请求已存在", err.Error())
}

func TestServiceFriend_AddFriend_Self(t *testing.T) {
	resetDB(t)
	alice := CreateUser(t, "alice_"+t.Name())

	chatSvc := service.NewChatService(dto.Chat, cache)
	friendSvc := service.NewFriendService(dto.Friend, dto.User, chatSvc, cache)

	_, err := friendSvc.AddFriend(alice.ID, model.AddFriendRequest{FriendID: alice.ID})
	require.Error(t, err)
	assert.Equal(t, "[add_self] 不能添加自己为好友", err.Error())
}

func TestServiceFriend_AcceptFriend_NotFound(t *testing.T) {
	resetDB(t)

	chatSvc := service.NewChatService(dto.Chat, cache)
	friendSvc := service.NewFriendService(dto.Friend, dto.User, chatSvc, cache)

	err := friendSvc.AcceptFriend(1, 99999)
	require.Error(t, err)
	assert.Equal(t, "[friend_request_not_found] 好友请求不存在", err.Error())
}

func TestServiceFriend_ListFriends_WithCache(t *testing.T) {
	resetDB(t)

	alice := CreateUser(t, "alice_"+t.Name())
	bob := CreateUser(t, "bob_"+t.Name())
	charles := CreateUser(t, "charles_"+t.Name())

	chatSvc := service.NewChatService(dto.Chat, cache)
	friendSvc := service.NewFriendService(dto.Friend, dto.User, chatSvc, cache)

	// Alice adds Bob
	r1, _ := friendSvc.AddFriend(alice.ID, model.AddFriendRequest{FriendID: bob.ID})
	friendSvc.AcceptFriend(bob.ID, r1.Request.ID)

	// Alice adds Charles
	r2, _ := friendSvc.AddFriend(alice.ID, model.AddFriendRequest{FriendID: charles.ID})
	friendSvc.AcceptFriend(charles.ID, r2.Request.ID)

	// First call — DB query + cache set
	friends, err := friendSvc.ListFriends(alice.ID)
	require.NoError(t, err)
	assert.Len(t, friends, 2)
}

func TestServiceFriend_GetPendingRequests(t *testing.T) {
	resetDB(t)

	alice := CreateUser(t, "alice_"+t.Name())
	bob := CreateUser(t, "bob_"+t.Name())

	chatSvc := service.NewChatService(dto.Chat, cache)
	friendSvc := service.NewFriendService(dto.Friend, dto.User, chatSvc, cache)

	// Alice sends to Bob
	_, err := friendSvc.AddFriend(alice.ID, model.AddFriendRequest{FriendID: bob.ID})
	require.NoError(t, err)

	// Bob sees 1 pending request
	pending, err := friendSvc.GetPendingRequests(bob.ID)
	require.NoError(t, err)
	assert.Len(t, pending, 1)
	assert.Equal(t, alice.ID, pending[0].User.ID)
}

func TestServiceFriend_DeleteFriend(t *testing.T) {
	resetDB(t)

	alice := CreateUser(t, "alice_"+t.Name())
	bob := CreateUser(t, "bob_"+t.Name())

	chatSvc := service.NewChatService(dto.Chat, cache)
	friendSvc := service.NewFriendService(dto.Friend, dto.User, chatSvc, cache)

	// Alice sends request, Bob accepts
	result, err := friendSvc.AddFriend(alice.ID, model.AddFriendRequest{FriendID: bob.ID})
	require.NoError(t, err)

	err = friendSvc.AcceptFriend(bob.ID, result.Request.ID)
	require.NoError(t, err)

	// Delete by Bob (the receiver)
	err = friendSvc.DeleteFriend(bob.ID, result.Request.ID)
	require.NoError(t, err)

	// Verify friendship removed
	friends, err := friendSvc.ListFriends(alice.ID)
	require.NoError(t, err)
	assert.Len(t, friends, 0)
}

func TestServiceFriend_DeleteFriend_Forbidden(t *testing.T) {
	resetDB(t)

	alice := CreateUser(t, "alice_"+t.Name())
	bob := CreateUser(t, "bob_"+t.Name())
	charles := CreateUser(t, "charles_"+t.Name())

	chatSvc := service.NewChatService(dto.Chat, cache)
	friendSvc := service.NewFriendService(dto.Friend, dto.User, chatSvc, cache)

	result, err := friendSvc.AddFriend(alice.ID, model.AddFriendRequest{FriendID: bob.ID})
	require.NoError(t, err)

	// Charles tries to delete the request (not involved)
	err = friendSvc.DeleteFriend(charles.ID, result.Request.ID)
	require.Error(t, err)
	assert.Equal(t, "[friend_op_forbidden] 无权操作", err.Error())
}
