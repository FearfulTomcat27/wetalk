//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"wetalk/dto"
)

func TestDTOUser_CreateAndFind(t *testing.T) {
	resetDB(t)

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	user, err := dto.User.Create("alice_"+t.Name(), string(hash), "Alice", "https://example.com/avatar.svg")
	require.NoError(t, err)
	require.NotZero(t, user.ID)
	assert.Equal(t, "Alice", user.Nickname)
	assert.Equal(t, "https://example.com/avatar.svg", user.Avatar)

	// FindByUsername
	found, err := dto.User.FindByUsername(user.Username)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "Alice", found.Nickname)
	assert.NotEmpty(t, found.PasswordHash)

	// FindByID
	found2, err := dto.User.FindByID(user.ID)
	require.NoError(t, err)
	require.NotNil(t, found2)
	assert.Equal(t, user.Username, found2.Username)
}

func TestDTOUser_FindByUsername_NotFound(t *testing.T) {
	resetDB(t)

	u, err := dto.User.FindByUsername("nonexistent")
	require.NoError(t, err)
	assert.Nil(t, u)
}

func TestDTOUser_FindByID_NotFound(t *testing.T) {
	resetDB(t)

	u, err := dto.User.FindByID(99999)
	require.NoError(t, err)
	assert.Nil(t, u)
}

func TestDTOUser_SearchByUsername(t *testing.T) {
	resetDB(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	_, err := dto.User.Create("alice", string(hash), "Alice", "")
	require.NoError(t, err)
	_, err = dto.User.Create("ali", string(hash), "Ali", "")
	require.NoError(t, err)
	_, err = dto.User.Create("bob", string(hash), "Bob", "")
	require.NoError(t, err)

	users, err := dto.User.SearchByUsername("ali", 20)
	require.NoError(t, err)
	assert.Len(t, users, 2)

	users, err = dto.User.SearchByUsername("bob", 20)
	require.NoError(t, err)
	assert.Len(t, users, 1)
}

func TestDTOUser_UpdateAvatar(t *testing.T) {
	resetDB(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user, err := dto.User.Create("update_avatar", string(hash), "AvatarTest", "")
	require.NoError(t, err)

	err = dto.User.UpdateAvatar(user.ID, "https://example.com/new-avatar.png")
	require.NoError(t, err)

	found, err := dto.User.FindByID(user.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "https://example.com/new-avatar.png", found.Avatar)
}
