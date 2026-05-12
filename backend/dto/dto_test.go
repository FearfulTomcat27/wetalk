package dto

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"wetalk/db"
	"wetalk/model"
)

// newMockDB 创建 go-sqlmock 并替换全局 db.DB，返回恢复函数
func newMockDB(t *testing.T) (sqlmock.Sqlmock, func()) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	oldDB := db.DB
	db.DB = gormDB

	cleanup := func() {
		db.DB = oldDB
		sqlDB.Close()
	}
	return mock, cleanup
}

func TestUser_FindByUsername_Found(t *testing.T) {
	mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, username, password_hash, nickname, avatar, created_at, updated_at FROM `users` WHERE username = ? ORDER BY `users`.`id` LIMIT ?",
	)).
		WithArgs("testuser", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname"}).
			AddRow(1, "testuser", "Test"))

	u, err := User.FindByUsername("testuser")
	require.NoError(t, err)
	require.NotNil(t, u)
	assert.Equal(t, "testuser", u.Username)
	assert.Equal(t, "Test", u.Nickname)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUser_FindByUsername_NotFound(t *testing.T) {
	mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, username, password_hash, nickname, avatar, created_at, updated_at FROM `users` WHERE username = ? ORDER BY `users`.`id` LIMIT ?",
	)).
		WithArgs("nobody", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	u, err := User.FindByUsername("nobody")
	require.NoError(t, err)
	assert.Nil(t, u)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUser_FindByID_Found(t *testing.T) {
	mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, username, nickname, avatar, created_at, updated_at FROM `users` WHERE id = ? ORDER BY `users`.`id` LIMIT ?",
	)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname"}).
			AddRow(1, "testuser", "Test"))

	u, err := User.FindByID(1)
	require.NoError(t, err)
	require.NotNil(t, u)
	assert.Equal(t, "testuser", u.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUser_SearchByUsername(t *testing.T) {
	mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, username, nickname, avatar FROM `users` WHERE username LIKE ? LIMIT ?",
	)).
		WithArgs("%ali%", 20).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname"}).
			AddRow(2, "alice", "Alice").
			AddRow(3, "ali", "Ali"))

	users, err := User.SearchByUsername("ali", 20)
	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, "alice", users[0].Username)
	assert.Equal(t, "ali", users[1].Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUser_Create(t *testing.T) {
	mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		"INSERT INTO `users` (`username`,`password_hash`,`nickname`,`avatar`,`created_at`,`updated_at`) VALUES (?,?,?,?,?,?)",
	)).
		WithArgs("newuser", "hashed_pw", "New User", "https://example.com/avatar.svg", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	u, err := User.Create("newuser", "hashed_pw", "New User", "https://example.com/avatar.svg")
	require.NoError(t, err)
	require.NotNil(t, u)
	assert.Equal(t, int64(1), u.ID)
	assert.Equal(t, "newuser", u.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUser_UpdateAvatar(t *testing.T) {
	mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `users` SET `avatar`=?,`updated_at`=? WHERE id = ?",
	)).
		WithArgs("https://example.com/new-avatar.png", sqlmock.AnyArg(), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := User.UpdateAvatar(1, "https://example.com/new-avatar.png")
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChat_GetMemberIDs(t *testing.T) {
	mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT `user_id` FROM `chat_members` WHERE chat_id = ?",
	)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).
			AddRow(1).AddRow(2).AddRow(3))

	ids, err := Chat.GetMemberIDs(1)
	require.NoError(t, err)
	assert.Equal(t, []int64{1, 2, 3}, ids)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChat_IsMember_True(t *testing.T) {
	mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT count(*) FROM `chat_members` WHERE chat_id = ? AND user_id = ?",
	)).
		WithArgs(1, 2).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	ok, err := Chat.IsMember(1, 2)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChat_IsMember_False(t *testing.T) {
	mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT count(*) FROM `chat_members` WHERE chat_id = ? AND user_id = ?",
	)).
		WithArgs(1, 99).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	ok, err := Chat.IsMember(1, 99)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFriend_FindByID_Found(t *testing.T) {
	mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, user_id, friend_id, status, created_at FROM `friend_requests` WHERE id = ? ORDER BY `friend_requests`.`id` LIMIT ?",
	)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "friend_id", "status"}).
			AddRow(1, 1, 2, model.FriendStatusPending))

	fr, err := Friend.FindByID(1)
	require.NoError(t, err)
	require.NotNil(t, fr)
	assert.Equal(t, int64(1), fr.UserID)
	assert.Equal(t, int64(2), fr.FriendID)
	assert.Equal(t, model.FriendStatusPending, fr.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFriend_FindRequestBetween_Found(t *testing.T) {
	mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, user_id, friend_id, status, created_at FROM `friend_requests` WHERE (user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?) ORDER BY `friend_requests`.`id` LIMIT ?",
	)).
		WithArgs(1, 2, 2, 1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "friend_id", "status"}).
			AddRow(1, 1, 2, model.FriendStatusPending))

	fr, err := Friend.FindRequestBetween(1, 2)
	require.NoError(t, err)
	require.NotNil(t, fr)
	assert.Equal(t, int64(1), fr.UserID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFriend_FindFriendshipBetween_Found(t *testing.T) {
	mock, cleanup := newMockDB(t)
	defer cleanup()

	smaller, larger := sortIDs(1, 2)
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, chat_id, user1_id, user2_id, created_at FROM `friendships` WHERE user1_id = ? AND user2_id = ? ORDER BY `friendships`.`id` LIMIT ?",
	)).
		WithArgs(smaller, larger, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "chat_id", "user1_id", "user2_id"}).
			AddRow(1, 100, 1, 2))

	f, err := Friend.FindFriendshipBetween(1, 2)
	require.NoError(t, err)
	require.NotNil(t, f)
	assert.Equal(t, int64(100), f.ChatID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFriend_FindFriendshipBetween_NotFound(t *testing.T) {
	mock, cleanup := newMockDB(t)
	defer cleanup()

	smaller, larger := sortIDs(1, 2)
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, chat_id, user1_id, user2_id, created_at FROM `friendships` WHERE user1_id = ? AND user2_id = ? ORDER BY `friendships`.`id` LIMIT ?",
	)).
		WithArgs(smaller, larger, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	f, err := Friend.FindFriendshipBetween(1, 2)
	require.NoError(t, err)
	assert.Nil(t, f)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFriend_Create(t *testing.T) {
	mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		"INSERT INTO `friend_requests` (`user_id`,`friend_id`,`status`,`created_at`,`updated_at`) VALUES (?,?,?,?,?)",
	)).
		WithArgs(1, 2, model.FriendStatusPending, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(10, 1))
	mock.ExpectCommit()

	fr, err := Friend.Create(1, 2)
	require.NoError(t, err)
	require.NotNil(t, fr)
	assert.Equal(t, int64(10), fr.ID)
	assert.Equal(t, model.FriendStatusPending, fr.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFriend_UpdateStatus(t *testing.T) {
	mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `friend_requests` SET `status`=?,`updated_at`=? WHERE id = ?",
	)).
		WithArgs(model.FriendStatusAccepted, sqlmock.AnyArg(), int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := Friend.UpdateStatus(10, model.FriendStatusAccepted)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChat_UpdateLastMessage(t *testing.T) {
	mock, cleanup := newMockDB(t)
	defer cleanup()

	now := time.Now()

	// Updates(map) produces non-deterministic SET column order, use regex
	mock.ExpectBegin()
	mock.ExpectExec(
		`^UPDATE `+regexp.QuoteMeta("`chats`")+` SET .+ WHERE id = \?$`,
	).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := Chat.UpdateLastMessage(1, 100, "hello", "text", now)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
