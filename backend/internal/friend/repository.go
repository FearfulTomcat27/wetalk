package friend

import (
	"database/sql"

	"wetalk/db"
)

// repository 好友数据仓库
type repository struct{}

// Repository 好友仓库实例
var Repository = &repository{}

// Create 创建好友请求
func (r *repository) Create(userID, friendID int64) (*Friend, error) {
	result, err := db.DB.Exec(
		"INSERT INTO friends (user_id, friend_id, status) VALUES (?, ?, ?)",
		userID, friendID, StatusPending,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &Friend{
		ID:       id,
		UserID:   userID,
		FriendID: friendID,
		Status:   StatusPending,
	}, nil
}

// FindByUserAndFriend 查找两个用户之间的好友记录
func (r *repository) FindByUserAndFriend(userID, friendID int64) (*Friend, error) {
	f := &Friend{}
	err := db.DB.QueryRow(
		"SELECT id, user_id, friend_id, status, created_at, updated_at FROM friends WHERE user_id = ? AND friend_id = ?",
		userID, friendID,
	).Scan(&f.ID, &f.UserID, &f.FriendID, &f.Status, &f.CreatedAt, &f.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return f, nil
}

// FindByID 根据 ID 查找好友记录
func (r *repository) FindByID(id int64) (*Friend, error) {
	f := &Friend{}
	err := db.DB.QueryRow(
		"SELECT id, user_id, friend_id, status, created_at, updated_at FROM friends WHERE id = ?",
		id,
	).Scan(&f.ID, &f.UserID, &f.FriendID, &f.Status, &f.CreatedAt, &f.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return f, nil
}

// ListByUserID 获取用户的好友列表（已接受的双向关系）
func (r *repository) ListByUserID(userID int64) ([]Friend, error) {
	rows, err := db.DB.Query(
		`SELECT id, user_id, friend_id, status, created_at, updated_at
		 FROM friends
		 WHERE user_id = ? AND status = ?
		 ORDER BY updated_at DESC`,
		userID, StatusAccepted,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friends []Friend
	for rows.Next() {
		var f Friend
		if err := rows.Scan(&f.ID, &f.UserID, &f.FriendID, &f.Status, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		friends = append(friends, f)
	}
	return friends, rows.Err()
}

// UpdateStatus 更新好友状态
func (r *repository) UpdateStatus(id int64, status string) error {
	_, err := db.DB.Exec("UPDATE friends SET status = ? WHERE id = ?", status, id)
	return err
}

// Delete 删除好友记录
func (r *repository) Delete(id int64) error {
	_, err := db.DB.Exec("DELETE FROM friends WHERE id = ?", id)
	return err
}
