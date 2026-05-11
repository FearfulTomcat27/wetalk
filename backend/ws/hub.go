package ws

import (
	"sync"

	"wetalk/common/logger"
)

// GetMemberIDsFunc 获取聊天成员ID列表的回调类型（避免 ws → chat 循环依赖）
type GetMemberIDsFunc func(chatID int64) ([]int64, error)

// Hub 管理 WebSocket 连接
type Hub struct {
	clients      map[int64]*Client
	register     chan *Client
	unregister   chan *Client
	mu           sync.RWMutex
	done         chan struct{}
	getMemberIDs GetMemberIDsFunc
}

// NewHub 创建 Hub 实例
func NewHub(getMemberIDs GetMemberIDsFunc) *Hub {
	return &Hub{
		clients:      make(map[int64]*Client),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		done:         make(chan struct{}),
		getMemberIDs: getMemberIDs,
	}
}

// Run 启动 Hub 主循环
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// 同一用户可能重连，关闭旧连接
			if old, exists := h.clients[client.userID]; exists {
				close(old.send)
			}
			h.clients[client.userID] = client
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if existing, ok := h.clients[client.userID]; ok && existing == client {
				delete(h.clients, client.userID)
				close(client.send)
			}
			h.mu.Unlock()

		case <-h.done:
			h.mu.Lock()
			for _, client := range h.clients {
				close(client.send)
			}
			h.clients = make(map[int64]*Client)
			h.mu.Unlock()
			return
		}
	}
}

// SendTo 向指定用户发送消息
func (h *Hub) SendTo(userID int64, msg interface{}) {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok {
		return
	}
	// recover 防止向已关闭 channel 发送导致 panic
	defer func() {
		if r := recover(); r != nil {
			logger.Module("hub").Warn("send to disconnected client",
				"user_id", userID, "recover", r)
		}
	}()
	select {
	case client.send <- msg:
	default:
		logger.Module("hub").Warn("client send channel full, dropping message",
			"user_id", userID)
	}
}

// SendToChat 向聊天所有在线成员广播消息（排除发送者）
func (h *Hub) SendToChat(chatID, senderID int64, msg interface{}) {
	if h.getMemberIDs == nil {
		return
	}
	memberIDs, err := h.getMemberIDs(chatID)
	if err != nil {
		logger.Module("hub").Error("get member ids for chat",
			"chat_id", chatID, "err", err)
		return
	}
	for _, memberID := range memberIDs {
		// 不再发回给发送者（发送者已通过 message.sent 拿到确认）
		if memberID != senderID {
			h.SendTo(memberID, msg)
		}
	}
}

// SetGetMemberIDs 设置获取成员 ID 的回调（用于打破循环依赖）
func (h *Hub) SetGetMemberIDs(fn GetMemberIDsFunc) {
	h.getMemberIDs = fn
}

// IsOnline 检查用户是否在线
func (h *Hub) IsOnline(userID int64) bool {
	h.mu.RLock()
	_, ok := h.clients[userID]
	h.mu.RUnlock()
	return ok
}

// Register 注册客户端
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister 注销客户端
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Shutdown 关闭 Hub
func (h *Hub) Shutdown() {
	close(h.done)
}
