package user

import (
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	pkgerrors "wetalk/pkg/errors"
)

// Service 用户业务逻辑
type Service struct {
	jwtSecret      string
	jwtExpireHours int
}

// NewService 创建用户业务逻辑服务
func NewService(jwtSecret string, jwtExpireHours int) *Service {
	return &Service{
		jwtSecret:      jwtSecret,
		jwtExpireHours: jwtExpireHours,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=128"`
	Nickname string `json:"nickname" binding:"max=128"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// Register 注册业务逻辑
func (s *Service) Register(req RegisterRequest) (*AuthResponse, error) {
	// 检查用户名是否已存在
	existUser, err := Repository.FindByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if existUser != nil {
		return nil, pkgerrors.ErrConflict
	}

	// bcrypt 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 默认昵称
	nickname := req.Nickname
	if nickname == "" {
		nickname = req.Username
	}

	// 生成 DiceBear 默认头像
	avatar := "https://api.dicebear.com/9.x/micah/svg?seed=" + url.QueryEscape(req.Username)

	// 创建用户
	user, err := Repository.Create(req.Username, string(hashedPassword), nickname, avatar)
	if err != nil {
		return nil, err
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// 生成 JWT
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

// Login 登录业务逻辑
func (s *Service) Login(req LoginRequest) (*AuthResponse, error) {
	// 查询用户
	user, err := Repository.FindByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, pkgerrors.ErrUnauthorized
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, pkgerrors.ErrUnauthorized
	}

	// 生成 JWT
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

// GetUserByID 根据 ID 获取用户信息
func (s *Service) GetUserByID(id int64) (*User, error) {
	return Repository.FindByID(id)
}

// SearchUsers 搜索用户
func (s *Service) SearchUsers(keyword string) ([]User, error) {
	if keyword == "" {
		return []User{}, nil
	}
	return Repository.SearchByUsername(keyword, 20)
}

// generateToken 生成 JWT 令牌
func (s *Service) generateToken(u *User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  u.ID,
		"username": u.Username,
		"exp":      time.Now().Add(time.Duration(s.jwtExpireHours) * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
