package friend

import (
	"errors"

	"gorm.io/gorm"

	"wetalk/db"
)

// repository 好友数据仓库
type repository struct{}

// Repository 好友仓库实例
var Repository = &repository{}

// sortIDs 排序两个 ID（小在前）
func sortIDs(a, b int64) (int64, int64) {
	if a < b {
		return a, b
	}
	return b, a
}

// ========== friend_requests 表（好友请求） ==========

// Create 创建好友请求
func (r *repository) Create(userID, friendID int64) (*FriendRequest, error) {
	fr := &FriendRequest{
		UserID:   userID,
		FriendID: friendID,
		Status:   StatusPending,
	}
	if err := db.DB.Create(fr).Error; err != nil {
		return nil, err
	}
	return fr, nil
}

// FindByUserAndFriend 查找两个用户之间的好友记录（单向：user_id → friend_id）
func (r *repository) FindByUserAndFriend(userID, friendID int64) (*FriendRequest, error) {
	var fr FriendRequest
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
func (r *repository) FindRequestBetween(userID, friendID int64) (*FriendRequest, error) {
	var fr FriendRequest
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
func (r *repository) FindFriendshipBetween(userID, friendID int64) (*Friendship, error) {
	smaller, larger := sortIDs(userID, friendID)
	var f Friendship
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
func (r *repository) FindByID(id int64) (*FriendRequest, error) {
	var fr FriendRequest
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
func (r *repository) UpdateStatus(id int64, status string) error {
	return db.DB.Model(&FriendRequest{}).Where("id = ?", id).Update("status", status).Error
}

// Delete 删除好友请求记录
func (r *repository) Delete(id int64) error {
	return db.DB.Delete(&FriendRequest{}, id).Error
}

// ========== PendingRequest (待处理请求) ==========

// SenderInfo 发送者信息
type SenderInfo struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// PendingRequest 待处理好友请求（含发送者信息）
type PendingRequest struct {
	ID        int64      `json:"id"`
	Status    string     `json:"status"`
	User      SenderInfo `json:"user"`
	CreatedAt string     `json:"created_at"`
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
func (r *repository) FindPendingByUserID(userID int64) ([]PendingRequest, error) {
	var rows []pendingRequestRow
	err := db.DB.Table("friend_requests").
		Select("friend_requests.id, friend_requests.status, users.id AS user_id, users.username, users.nickname, users.avatar, friend_requests.created_at").
		Joins("JOIN users ON users.id = friend_requests.user_id").
		Where("friend_requests.friend_id = ? AND friend_requests.status = ?", userID, StatusPending).
		Order("friend_requests.created_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	// 组装为嵌套结构
	requests := make([]PendingRequest, 0, len(rows))
	for _, row := range rows {
		requests = append(requests, PendingRequest{
			ID:     row.ID,
			Status: row.Status,
			User: SenderInfo{
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

// FriendshipInfo 好友信息（从 friendships 表查询，JOIN users 获取对方信息）
type FriendshipInfo struct {
	ID              int64  `json:"id"`
	FriendID        int64  `json:"friend_id"`
	FriendName      string `json:"friend_name"`
	FriendAvatar    string `json:"friend_avatar"`
	LastMessage     string `json:"last_message"`
	LastMessageType string `json:"last_message_type"`
	LastMessageTime string `json:"last_message_time"`
	UnreadCount     int    `json:"unread_count"`
	CreatedAt       string `json:"created_at"`
}

// CreateFriendship 创建好友关系（小 ID 在前，防重）
func (r *repository) CreateFriendship(user1ID, user2ID int64) error {
	smaller, larger := sortIDs(user1ID, user2ID)
	friendship := &Friendship{
		User1ID: smaller,
		User2ID: larger,
	}
	// FirstOrCreate: 若已存在则直接返回，防止重复插入
	return db.DB.Where("user1_id = ? AND user2_id = ?", smaller, larger).
		FirstOrCreate(friendship).Error
}

// FindFriendships 查询用户的所有好友（JOIN users，含 last_message 和 unread_count 子查询）
func (r *repository) FindFriendships(userID int64) ([]FriendshipInfo, error) {
	var friendships []FriendshipInfo
	err := db.DB.Table("friendships").
		Select(`friendships.id,
		        CASE WHEN friendships.user1_id = ? THEN friendships.user2_id ELSE friendships.user1_id END AS friend_id,
		        users.nickname AS friend_name,
		        users.avatar AS friend_avatar,
		        COALESCE((SELECT content FROM messages
		         WHERE ((sender_id = friendships.user1_id AND receiver_id = friendships.user2_id)
		             OR (sender_id = friendships.user2_id AND receiver_id = friendships.user1_id))
		         ORDER BY created_at DESC LIMIT 1), '') AS last_message,
		        COALESCE((SELECT content_type FROM messages
		         WHERE ((sender_id = friendships.user1_id AND receiver_id = friendships.user2_id)
		             OR (sender_id = friendships.user2_id AND receiver_id = friendships.user1_id))
		         ORDER BY created_at DESC LIMIT 1), 'text') AS last_message_type,
		        COALESCE((SELECT created_at FROM messages
		         WHERE ((sender_id = friendships.user1_id AND receiver_id = friendships.user2_id)
		             OR (sender_id = friendships.user2_id AND receiver_id = friendships.user1_id))
		         ORDER BY created_at DESC LIMIT 1), '') AS last_message_time,
		        (SELECT COUNT(*) FROM messages
		         WHERE receiver_id = ?
		           AND sender_id = CASE WHEN friendships.user1_id = ? THEN friendships.user2_id ELSE friendships.user1_id END
		           AND status != 'read') AS unread_count,
		        friendships.created_at`, userID, userID, userID).
		Joins("JOIN users ON users.id = CASE WHEN friendships.user1_id = ? THEN friendships.user2_id ELSE friendships.user1_id END", userID).
		Where("friendships.user1_id = ? OR friendships.user2_id = ?", userID, userID).
		Order("friendships.created_at DESC").
		Scan(&friendships).Error
	if err != nil {
		return nil, err
	}
	return friendships, nil
}

// DeleteFriendship 删除好友关系
func (r *repository) DeleteFriendship(user1ID, user2ID int64) error {
	smaller, larger := sortIDs(user1ID, user2ID)
	return db.DB.Where("user1_id = ? AND user2_id = ?", smaller, larger).
		Delete(&Friendship{}).Error
}
