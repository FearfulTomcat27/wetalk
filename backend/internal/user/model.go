package user

import "time"

// User 用户模型
type User struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Username     string    `json:"username" gorm:"column:username;type:varchar(64);uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"column:password_hash;type:varchar(255);not null"`
	Nickname     string    `json:"nickname" gorm:"column:nickname;type:varchar(128);not null;default:''"`
	Avatar       string    `json:"avatar" gorm:"column:avatar;type:varchar(512);not null;default:''"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
