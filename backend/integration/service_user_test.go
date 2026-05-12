//go:build integration

package integration

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wetalk/model"
)

func TestServiceUser_RegisterAndLogin(t *testing.T) {
	resetDB(t)

	svc := newTestUserService()

	// Register
	resp, err := svc.Register(model.RegisterRequest{
		Username: "testuser_" + t.Name(),
		Password: "password123",
		Nickname: "Test User",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "testuser_"+t.Name(), resp.User.Username)
	assert.Equal(t, "Test User", resp.User.Nickname)
	assert.Contains(t, resp.User.Avatar, "dicebear.com")

	// Login
	loginResp, err := svc.Login(model.LoginRequest{
		Username: "testuser_" + t.Name(),
		Password: "password123",
	})
	require.NoError(t, err)
	require.NotNil(t, loginResp)
	assert.Equal(t, resp.User.ID, loginResp.User.ID)
}

func TestServiceUser_Register_Duplicate(t *testing.T) {
	resetDB(t)

	svc := newTestUserService()
	_, err := svc.Register(model.RegisterRequest{
		Username: "dupuser_" + t.Name(),
		Password: "password123",
	})
	require.NoError(t, err)

	_, err = svc.Register(model.RegisterRequest{
		Username: "dupuser_" + t.Name(),
		Password: "password123",
	})
	require.Error(t, err)
	assert.Equal(t, "[user_already_exists] 用户名已存在", err.Error())
}

func TestServiceUser_Login_InvalidCredentials(t *testing.T) {
	resetDB(t)

	svc := newTestUserService()

	// Wrong password
	_, err := svc.Login(model.LoginRequest{
		Username: "nobody",
		Password: "wrong",
	})
	require.Error(t, err)
	assert.Equal(t, "[invalid_credentials] 用户名或密码错误", err.Error())
}

func TestServiceUser_GetUserByID(t *testing.T) {
	resetDB(t)
	_ = t

	svc := newTestUserService()
	resp, err := svc.Register(model.RegisterRequest{
		Username: "getuser_" + t.Name(),
		Password: "password123",
	})
	require.NoError(t, err)

	// First call — cache miss
	user, err := svc.GetUserByID(resp.User.ID)
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, resp.User.Username, user.Username)

	// Second call — cache hit
	user2, err := svc.GetUserByID(resp.User.ID)
	require.NoError(t, err)
	require.NotNil(t, user2)
	assert.Equal(t, user.Username, user2.Username)
}

func TestServiceUser_SearchUsers(t *testing.T) {
	resetDB(t)

	svc := newTestUserService()

	_, err := svc.Register(model.RegisterRequest{
		Username: "searchme_" + t.Name(),
		Password: "password123",
	})
	require.NoError(t, err)

	users, err := svc.SearchUsers("searchme")
	require.NoError(t, err)
	assert.Len(t, users, 1)

	// Empty keyword returns empty slice
	empty, err := svc.SearchUsers("")
	require.NoError(t, err)
	assert.Empty(t, empty)
}

func TestServiceUser_UploadAvatar(t *testing.T) {
	resetDB(t)

	oss := &stubOSS{}
	svc := newTestUserService()
	resp, err := svc.Register(model.RegisterRequest{
		Username: "avatar_" + t.Name(),
		Password: "password123",
	})
	require.NoError(t, err)

	// Upload valid png
	imgData := bytes.NewReader([]byte("fake-png-data"))
	url, err := svc.UploadAvatar(context.Background(), resp.User.ID, imgData, "avatar.png", 100)
	require.NoError(t, err)
	assert.Contains(t, url, "test-bucket.oss-cn-test.aliyuncs.com")
	assert.Contains(t, url, "avatars/")

	// Avatar URL updated in DB
	user, err := svc.GetUserByID(resp.User.ID)
	require.NoError(t, err)
	assert.Equal(t, url, user.Avatar)
	_ = oss
}

func TestServiceUser_UploadAvatar_TooLarge(t *testing.T) {
	resetDB(t)

	svc := newTestUserService()
	resp, err := svc.Register(model.RegisterRequest{
		Username: "avatarsize_" + t.Name(),
		Password: "password123",
	})
	require.NoError(t, err)

	_, err = svc.UploadAvatar(context.Background(), resp.User.ID, nil, "avatar.png", 3*1024*1024)
	require.Error(t, err)
	assert.Equal(t, "[avatar_too_large] 头像文件不能超过 2MB", err.Error())
}

func TestServiceUser_UploadAvatar_InvalidFormat(t *testing.T) {
	resetDB(t)

	svc := newTestUserService()
	resp, err := svc.Register(model.RegisterRequest{
		Username: "avatarfmt_" + t.Name(),
		Password: "password123",
	})
	require.NoError(t, err)

	_, err = svc.UploadAvatar(context.Background(), resp.User.ID, nil, "document.pdf", 100)
	require.Error(t, err)
	assert.Equal(t, "[invalid_avatar_format] 仅支持 jpg/png/gif/webp 格式", err.Error())
}
