package ws

import (
	"log"
	"sync"
)

// Hub 管理 WebSocket 连接
type Hub struct {
	clients    map[int64]*Client
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	done       chan struct{}
}

// NewHub 创建 Hub 实例
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int64]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		done:       make(chan struct{}),
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
			log.Printf("send to disconnected client %d: %v", userID, r)
		}
	}()
	select {
	case client.send <- msg:
	default:
		log.Printf("client %d send channel full, dropping message", userID)
	}
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