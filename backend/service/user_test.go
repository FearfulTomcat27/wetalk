package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"wetalk/config"
	"wetalk/model"
	"wetalk/types"
)

// ========== 测试辅助函数 ==========

func newTestUserService(repo *MockUserRepo, cache *MockCache, oss *MockOSS) *UserService {
	return NewUserService(repo, cache, config.JWTConfig{Secret: "test-secret", ExpireHours: 72}, oss)
}

func newTestUser() *model.User {
	return &model.User{
		ID:       1,
		Username: "testuser",
		Nickname: "Test",
		Avatar:   "https://api.dicebear.com/9.x/micah/svg?seed=testuser",
	}
}

func hashPassword(password string) string {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(h)
}

// ========== UserService 测试 ==========

func TestUserService_Register_Success(t *testing.T) {
	repo := new(MockUserRepo)
	cache := new(MockCache)
	oss := new(MockOSS)

	repo.On("FindByUsername", "newuser").Return(nil, nil)
	repo.On("Create", "newuser", mock.AnythingOfType("string"), "New User", mock.AnythingOfType("string")).
		Return(&model.User{ID: 1, Username: "newuser", Nickname: "New User", Avatar: "https://api.dicebear.com/9.x/micah/svg?seed=newuser"}, nil)

	svc := newTestUserService(repo, cache, oss)

	resp, err := svc.Register(model.RegisterRequest{
		Username: "newuser",
		Password: "password123",
		Nickname: "New User",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	require.NotNil(t, resp.User)
	assert.Equal(t, "newuser", resp.User.Username)
	assert.Contains(t, resp.User.Avatar, "https://api.dicebear.com")

	repo.AssertExpectations(t)
}

func TestUserService_Register_EmptyNickname(t *testing.T) {
	repo := new(MockUserRepo)
	cache := new(MockCache)
	oss := new(MockOSS)

	repo.On("FindByUsername", "nonickuser").Return(nil, nil)
	repo.On("Create", "nonickuser", mock.AnythingOfType("string"), "nonickuser", mock.AnythingOfType("string")).
		Return(&model.User{ID: 1, Username: "nonickuser", Nickname: "nonickuser"}, nil)

	svc := newTestUserService(repo, cache, oss)

	resp, err := svc.Register(model.RegisterRequest{
		Username: "nonickuser",
		Password: "password123",
	})
	require.NoError(t, err)
	assert.Equal(t, "nonickuser", resp.User.Nickname)

	repo.AssertExpectations(t)
}

func TestUserService_Register_AlreadyExists(t *testing.T) {
	repo := new(MockUserRepo)
	cache := new(MockCache)
	oss := new(MockOSS)

	repo.On("FindByUsername", "testuser").Return(newTestUser(), nil)

	svc := newTestUserService(repo, cache, oss)

	_, err := svc.Register(model.RegisterRequest{
		Username: "testuser",
		Password: "password123",
	})
	assert.ErrorIs(t, err, types.ErrUserAlreadyExists)

	repo.AssertExpectations(t)
}

func TestUserService_Register_DBFailure(t *testing.T) {
	repo := new(MockUserRepo)
	cache := new(MockCache)
	oss := new(MockOSS)

	repo.On("FindByUsername", "newuser").Return(nil, nil)
	repo.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, assert.AnError)

	svc := newTestUserService(repo, cache, oss)

	_, err := svc.Register(model.RegisterRequest{
		Username: "newuser",
		Password: "password123",
	})
	assert.Error(t, err)

	repo.AssertExpectations(t)
}

func TestUserService_Login_Success(t *testing.T) {
	repo := new(MockUserRepo)
	cache := new(MockCache)
	oss := new(MockOSS)

	u := newTestUser()
	u.PasswordHash = hashPassword("mypassword")
	repo.On("FindByUsername", "testuser").Return(u, nil)

	svc := newTestUserService(repo, cache, oss)

	resp, err := svc.Login(model.LoginRequest{
		Username: "testuser",
		Password: "mypassword",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	require.NotNil(t, resp.User)

	repo.AssertExpectations(t)
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	repo := new(MockUserRepo)
	cache := new(MockCache)
	oss := new(MockOSS)

	repo.On("FindByUsername", "nonexistent").Return(nil, nil)

	svc := newTestUserService(repo, cache, oss)

	_, err := svc.Login(model.LoginRequest{
		Username: "nonexistent",
		Password: "password123",
	})
	assert.ErrorIs(t, err, types.ErrInvalidCredentials)

	repo.AssertExpectations(t)
}

func TestUserService_Login_WrongPassword(t *testing.T) {
	repo := new(MockUserRepo)
	cache := new(MockCache)
	oss := new(MockOSS)

	u := newTestUser()
	u.PasswordHash = hashPassword("correctpassword")
	repo.On("FindByUsername", "testuser").Return(u, nil)

	svc := newTestUserService(repo, cache, oss)

	_, err := svc.Login(model.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	})
	assert.ErrorIs(t, err, types.ErrInvalidCredentials)

	repo.AssertExpectations(t)
}

func TestUserService_GetUserByID_CacheHit(t *testing.T) {
	repo := new(MockUserRepo)
	cache := new(MockCache)
	oss := new(MockOSS)

	u := newTestUser()
	cache.On("Get", "user:1", mock.Anything).Return(true, nil, u)

	svc := newTestUserService(repo, cache, oss)

	user, err := svc.GetUserByID(1)
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)

	repo.AssertNotCalled(t, "FindByID")
	cache.AssertExpectations(t)
}

func TestUserService_GetUserByID_CacheMiss(t *testing.T) {
	repo := new(MockUserRepo)
	cache := new(MockCache)
	oss := new(MockOSS)

	repo.On("FindByID", int64(1)).Return(newTestUser(), nil)
	cache.On("Get", "user:1", mock.Anything).Return(false, nil)
	cache.On("Set", "user:1", mock.Anything, mock.AnythingOfType("time.Duration")).Return(nil)

	svc := newTestUserService(repo, cache, oss)

	user, err := svc.GetUserByID(1)
	require.NoError(t, err)
	require.NotNil(t, user)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestUserService_GetUserByID_NotFound(t *testing.T) {
	repo := new(MockUserRepo)
	cache := new(MockCache)
	oss := new(MockOSS)

	repo.On("FindByID", int64(999)).Return(nil, nil)
	cache.On("Get", "user:999", mock.Anything).Return(false, nil)

	svc := newTestUserService(repo, cache, oss)

	user, err := svc.GetUserByID(999)
	require.NoError(t, err)
	assert.Nil(t, user)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestUserService_SearchUsers_EmptyKeyword(t *testing.T) {
	repo := new(MockUserRepo)
	cache := new(MockCache)
	oss := new(MockOSS)

	svc := newTestUserService(repo, cache, oss)

	users, err := svc.SearchUsers("")
	require.NoError(t, err)
	assert.Empty(t, users)

	repo.AssertNotCalled(t, "SearchByUsername")
}

func TestUserService_SearchUsers_WithKeyword(t *testing.T) {
	repo := new(MockUserRepo)
	cache := new(MockCache)
	oss := new(MockOSS)

	repo.On("SearchByUsername", "ali", 20).Return([]model.User{
		{ID: 2, Username: "alice", Nickname: "Alice"},
	}, nil)

	svc := newTestUserService(repo, cache, oss)

	users, err := svc.SearchUsers("ali")
	require.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "alice", users[0].Username)

	repo.AssertExpectations(t)
}

func TestUserService_GenerateToken(t *testing.T) {
	repo := new(MockUserRepo)
	cache := new(MockCache)
	oss := new(MockOSS)

	svc := newTestUserService(repo, cache, oss)
	u := newTestUser()

	token, err := svc.generateToken(u)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}
