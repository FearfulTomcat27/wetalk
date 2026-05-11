package dto

import (
	"errors"

	"gorm.io/gorm"

	"wetalk/db"
	"wetalk/model"
)

// friendRepo 好友数据仓库
type friendRepo struct{}

// Friend 好友仓库实例
var Friend = &friendRepo{}

// sortIDs 排序两个 ID（小在前）
func sortIDs(a, b int64) (int64, int64) {
	if a < b {
		return a, b
	}
	return b, a
}

// ========== friend_requests 表（好友请求） ==========

// Create 创建好友请求
func (r *friendRepo) Create(userID, friendID int64) (*model.FriendRequest, error) {
	fr := &model.FriendRequest{
		UserID:   userID,
		FriendID: friendID,
		Status:   model.FriendStatusPending,
	}
	if err := db.DB.Create(fr).Error; err != nil {
		return nil, err
	}
	return fr, nil
}

// FindByUserAndFriend 查找两个用户之间的好友记录（单向：user_id → friend_id）
func (r *friendRepo) FindByUserAndFriend(userID, friendID int64) (*model.FriendRequest, error) {
	var fr model.FriendRequest
	err := db.DB.Where("user_id = ? AND friend_id = ?", userID, friendID).First(&fr).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &fr, nil
}

// FindRequestBetween 查找两个用户之间任意方向的好友请求
func (r *friendRepo) FindRequestBetween(userID, friendID int64) (*model.FriendRequest, error) {
	var fr model.FriendRequest
	err := db.DB.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		userID, friendID, friendID, userID).First(&fr).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &fr, nil
}

// FindFriendshipBetween 检查两个用户是否已经是好友
func (r *friendRepo) FindFriendshipBetween(userID, friendID int64) (*model.Friendship, error) {
	smaller, larger := sortIDs(userID, friendID)
	var f model.Friendship
	err := db.DB.Where("user1_id = ? AND user2_id = ?", smaller, larger).First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// FindByID 根据 ID 查找好友记录
func (r *friendRepo) FindByID(id int64) (*model.FriendRequest, error) {
	var fr model.FriendRequest
	err := db.DB.Where("id = ?", id).First(&fr).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &fr, nil
}

// UpdateStatus 更新好友状态
func (r *friendRepo) UpdateStatus(id int64, status string) error {
	return db.DB.Model(&model.FriendRequest{}).Where("id = ?", id).Update("status", status).Error
}

// Delete 删除好友请求记录
func (r *friendRepo) Delete(id int64) error {
	return db.DB.Delete(&model.FriendRequest{}, id).Error
}

// pendingRequestRow 扁平查询结果（GORM 扫描用）
type pendingRequestRow struct {
	ID        int64  `gorm:"column:id"`
	Status    string `gorm:"column:status"`
	UserID    int64  `gorm:"column:user_id"`
	Username  string `gorm:"column:username"`
	Nickname  string `gorm:"column:nickname"`
	Avatar    string `gorm:"column:avatar"`
	CreatedAt string `gorm:"column:created_at"`
}

// FindPendingByUserID 获取发给当前用户的待处理好友请求（JOIN users 查发送者信息）
func (r *friendRepo) FindPendingByUserID(userID int64) ([]model.PendingRequest, error) {
	var rows []pendingRequestRow
	err := db.DB.Table("friend_requests").
		Select("friend_requests.id, friend_requests.status, users.id AS user_id, users.username, users.nickname, users.avatar, friend_requests.created_at").
		Joins("JOIN users ON users.id = friend_requests.user_id").
		Where("friend_requests.friend_id = ? AND friend_requests.status = ?", userID, model.FriendStatusPending).
		Order("friend_requests.created_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	// 组装为嵌套结构
	requests := make([]model.PendingRequest, 0, len(rows))
	for _, row := range rows {
		requests = append(requests, model.PendingRequest{
			ID:     row.ID,
			Status: row.Status,
			User: model.SenderInfo{
				ID:       row.UserID,
				Username: row.Username,
				Nickname: row.Nickname,
				Avatar:   row.Avatar,
			},
			CreatedAt: row.CreatedAt,
		})
	}
	return requests, nil
}

// ========== friendships 表（已建立的好友关系） ==========

// CreateFriendship 创建好友关系（小 ID 在前，带 chat_id，防重）
func (r *friendRepo) CreateFriendship(user1ID, user2ID, chatID int64) (*model.Friendship, error) {
	smaller, larger := sortIDs(user1ID, user2ID)
	friendship := &model.Friendship{
		User1ID: smaller,
		User2ID: larger,
		ChatID:  chatID,
	}
	// FirstOrCreate: 若已存在则直接返回，防止重复插入
	if err := db.DB.Where("user1_id = ? AND user2_id = ?", smaller, larger).
		FirstOrCreate(friendship).Error; err != nil {
		return nil, err
	}
	return friendship, nil
}

// FindFriendships 查询用户的所有好友（CTE + UNION ALL 替代 OR 以利用索引）
// chatOnly 为 true 时只返回有消息记录的好友（供聊天页使用）
func (r *friendRepo) FindFriendships(userID int64, chatOnly ...bool) ([]model.FriendshipInfo, error) {
	var friendships []model.FriendshipInfo

	filterChat := len(chatOnly) > 0 && chatOnly[0]
	chatFilter := ""
	if filterChat {
		chatFilter = " AND c.last_message_id IS NOT NULL"
	}

	query := `
		WITH unread AS (
			SELECT m.chat_id, COUNT(*) AS cnt
			FROM messages m
			WHERE m.sender_id != ? AND m.status != 'read'
			  AND m.chat_id IN (SELECT chat_id FROM friendships WHERE user1_id = ? OR user2_id = ?)
			GROUP BY m.chat_id
		)
		SELECT f.id, f.chat_id,
		       f.user2_id AS friend_id,
		       u.nickname AS friend_name,
		       u.avatar AS friend_avatar,
		       COALESCE(c.last_message_text, '') AS last_message,
		       COALESCE(m.content_type, 'text') AS last_message_type,
		       COALESCE(c.last_message_time, '') AS last_message_time,
		       COALESCE(un.cnt, 0) AS unread_count,
		       f.created_at
		FROM friendships f
		JOIN users u ON u.id = f.user2_id
		LEFT JOIN chats c ON c.id = f.chat_id
		LEFT JOIN messages m ON m.id = c.last_message_id
		LEFT JOIN unread un ON un.chat_id = f.chat_id
		WHERE f.user1_id = ?` + chatFilter + `

		UNION ALL

		SELECT f.id, f.chat_id,
		       f.user1_id AS friend_id,
		       u.nickname AS friend_name,
		       u.avatar AS friend_avatar,
		       COALESCE(c.last_message_text, '') AS last_message,
		       COALESCE(m.content_type, 'text') AS last_message_type,
		       COALESCE(c.last_message_time, '') AS last_message_time,
		       COALESCE(un.cnt, 0) AS unread_count,
		       f.created_at
		FROM friendships f
		JOIN users u ON u.id = f.user1_id
		LEFT JOIN chats c ON c.id = f.chat_id
		LEFT JOIN messages m ON m.id = c.last_message_id
		LEFT JOIN unread un ON un.chat_id = f.chat_id
		WHERE f.user2_id = ?` + chatFilter + `

		ORDER BY COALESCE(last_message_time, created_at) DESC`

	if err := db.DB.Raw(query, userID, userID, userID, userID, userID).Scan(&friendships).Error; err != nil {
		return nil, err
	}
	return friendships, nil
}

// DeleteFriendship 删除好友关系
func (r *friendRepo) DeleteFriendship(user1ID, user2ID int64) error {
	smaller, larger := sortIDs(user1ID, user2ID)
	return db.DB.Where("user1_id = ? AND user2_id = ?", smaller, larger).
		Delete(&model.Friendship{}).Error
}
