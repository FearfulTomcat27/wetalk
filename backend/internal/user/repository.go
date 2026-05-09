package user

import (
	"errors"

	"gorm.io/gorm"

	"wetalk/db"
)

// repository 用户数据仓库
type repository struct{}

// Repository 用户仓库实例
var Repository = &repository{}

// FindByUsername 根据用户名查找用户
func (r *repository) FindByUsername(username string) (*User, error) {
	var u User
	err := db.DB.Where("username = ?", username).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByID 根据 ID 查找用户
func (r *repository) FindByID(id int64) (*User, error) {
	var u User
	err := db.DB.Where("id = ?", id).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// SearchByUsername 根据用户名模糊搜索
func (r *repository) SearchByUsername(keyword string, limit int) ([]User, error) {
	var users []User
	err := db.DB.Where("username LIKE ?", "%"+keyword+"%").
		Limit(limit).
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// Create 创建用户
func (r *repository) Create(username, passwordHash, nickname, avatar string) (*User, error) {
	user := &User{
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
func (r *repository) UpdateAvatar(userID int64, avatarURL string) error {
	return db.DB.Model(&User{}).Where("id = ?", userID).Update("avatar", avatarURL).Error
}
