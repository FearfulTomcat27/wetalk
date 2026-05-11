package service

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"wetalk/common"
	"wetalk/dto"
	"wetalk/model"
	"wetalk/type"
)

// UserService 用户业务逻辑
type UserService struct {
	jwtSecret      string
	jwtExpireHours int
	ossClient      *common.Client
}

// NewUserService 创建用户业务逻辑服务
func NewUserService(jwtSecret string, jwtExpireHours int, ossClient *common.Client) *UserService {
	return &UserService{
		jwtSecret:      jwtSecret,
		jwtExpireHours: jwtExpireHours,
		ossClient:      ossClient,
	}
}

// Register 注册业务逻辑
func (s *UserService) Register(req model.RegisterRequest) (*model.AuthResponse, error) {
	// 检查用户名是否已存在
	existUser, err := dto.User.FindByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if existUser != nil {
		return nil, types.ErrUserAlreadyExists
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
	user, err := dto.User.Create(req.Username, string(hashedPassword), nickname, avatar)
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

	return &model.AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

// Login 登录业务逻辑
func (s *UserService) Login(req model.LoginRequest) (*model.AuthResponse, error) {
	// 查询用户
	user, err := dto.User.FindByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, types.ErrInvalidCredentials
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, types.ErrInvalidCredentials
	}

	// 生成 JWT
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

// GetUserByID 根据 ID 获取用户信息
func (s *UserService) GetUserByID(id int64) (*model.User, error) {
	return dto.User.FindByID(id)
}

// SearchUsers 搜索用户
func (s *UserService) SearchUsers(keyword string) ([]model.User, error) {
	if keyword == "" {
		return []model.User{}, nil
	}
	return dto.User.SearchByUsername(keyword, 20)
}

// generateToken 生成 JWT 令牌
func (s *UserService) generateToken(u *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  u.ID,
		"username": u.Username,
		"exp":      time.Now().Add(time.Duration(s.jwtExpireHours) * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// UploadAvatar 上传头像
func (s *UserService) UploadAvatar(ctx context.Context, userID int64, file io.Reader, filename string, size int64) (string, error) {
	// 校验文件大小 ≤ 2MB
	if size > 2*1024*1024 {
		return "", types.ErrAvatarTooLarge
	}

	// 校验文件扩展名（忽略大小写）
	ext := strings.ToLower(filepath.Ext(filename))
	contentType := ""
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	default:
		return "", types.ErrInvalidAvatarFormat
	}

	// 生成 OSS key: avatars/{uid}/{uuid}.{ext}
	objectKey := fmt.Sprintf("avatars/%d/%d%s", userID, time.Now().UnixNano(), ext)

	// 上传到 OSS (ACL=public-read)
	avatarURL, err := s.ossClient.PutObject(ctx, objectKey, file, contentType, size)
	if err != nil {
		return "", err
	}

	// 更新 DB
	if err := dto.User.UpdateAvatar(userID, avatarURL); err != nil {
		return "", err
	}

	return avatarURL, nil
}
