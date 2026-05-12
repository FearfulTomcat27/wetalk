package dto

import (
	"errors"

	"gorm.io/gorm"

	"wetalk/db"
	"wetalk/model"
)

// userRepo 用户数据仓库
type userRepo struct{}

// User 用户仓库实例
var User = &userRepo{}

// FindByUsername 根据用户名查找用户（登录/注册共用，需 password_hash 做 bcrypt 校验）
func (r *userRepo) FindByUsername(username string) (*model.User, error) {
	var u model.User
	err := db.DB.Select("id, username, password_hash, nickname, avatar, created_at, updated_at").
		Where("username = ?", username).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByID 根据 ID 查找用户
func (r *userRepo) FindByID(id int64) (*model.User, error) {
	var u model.User
	err := db.DB.Select("id, username, nickname, avatar, created_at, updated_at").
		Where("id = ?", id).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// SearchByUsername 根据用户名模糊搜索
func (r *userRepo) SearchByUsername(keyword string, limit int) ([]model.User, error) {
	var users []model.User
	err := db.DB.Select("id, username, nickname, avatar").
		Where("username LIKE ?", "%"+keyword+"%").
		Limit(limit).
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// Create 创建用户
func (r *userRepo) Create(username, passwordHash, nickname, avatar string) (*model.User, error) {
	user := &model.User{
		Username:     username,
		PasswordHash: passwordHash,
		Nickname:     nickname,
		Avatar:       avatar,
	}
	if err := db.DB.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateAvatar 更新用户头像 URL
func (r *userRepo) UpdateAvatar(userID int64, avatarURL string) error {
	return db.DB.Model(&model.User{}).Where("id = ?", userID).Update("avatar", avatarURL).Error
}
