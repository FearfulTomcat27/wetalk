package user

import (
	"database/sql"

	"wetalk/db"
)

// repository 用户数据仓库
type repository struct{}

// Repository 用户仓库实例
var Repository = &repository{}

// FindByUsername 根据用户名查找用户
func (r *repository) FindByUsername(username string) (*User, error) {
	u := &User{}
	err := db.DB.QueryRow(
		"SELECT id, username, password_hash, nickname, avatar, created_at, updated_at FROM users WHERE username = ?",
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Nickname, &u.Avatar, &u.CreatedAt, &u.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// SearchByUsername 根据用户名前缀搜索用户（排除密码字段）
func (r *repository) SearchByUsername(keyword string, limit int) ([]User, error) {
	rows, err := db.DB.Query(
		"SELECT id, username, '', nickname, avatar, created_at, updated_at FROM users WHERE username LIKE ? LIMIT ?",
		"%"+keyword+"%", limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Nickname, &u.Avatar, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// Create 创建用户
func (r *repository) Create(username, passwordHash, nickname string) (*User, error) {
	result, err := db.DB.Exec(
		"INSERT INTO users (username, password_hash, nickname) VALUES (?, ?, ?)",
		username, passwordHash, nickname,
	)
	if err != nil {
		return nil, err
	}

	userID, _ := result.LastInsertId()
	return &User{
		ID:       userID,
		Username: username,
		Nickname: nickname,
	}, nil
}
