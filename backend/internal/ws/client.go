package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 54 * time.Second
	maxMessageSize = 4096
)

// SendMessageFunc 发送消息回调函数类型（避免 ws → message 循环依赖）
type SendMessageFunc func(senderID int64, receiverID int64, content string, clientMsgID string) (*SentMessage, error)

// SentMessage 回调返回的消息数据（与 message.Message 对应）
type SentMessage struct {
	ID          int64
	SenderID    int64
	ReceiverID  int64
	Content     string
	ContentType string
	Status      string
	CreatedAt   time.Time
}

// Client WebSocket 客户端
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan interface{}
	userID    int64
	authed    bool
	jwtSecret string
	sendMsg   SendMessageFunc
}

// NewClient 创建客户端（未认证状态）
func NewClient(hub *Hub, conn *websocket.Conn, jwtSecret string, sendMsg SendMessageFunc) *Client {
	return &Client{
		hub:       hub,
		conn:      conn,
		send:      make(chan interface{}, 256),
		jwtSecret: jwtSecret,
		sendMsg:   sendMsg,
	}
}

// ReadPump 读取客户端消息
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("read error: %v", err)
			}
			break
		}

		// 解析扁平格式：先提取 type 字段
		var payload map[string]interface{}
		if err := json.Unmarshal(raw, &payload); err != nil {
			c.send <- &ErrorEvent{Type: TypeError, Message: "无效的消息格式"}
			continue
		}

		typeStr, _ := payload["type"].(string)
		switch typeStr {
		case TypeAuth:
			if c.authed {
				c.send <- &ErrorEvent{Type: TypeError, Message: "已认证"}
				continue
			}
			var req AuthRequest
			if err := json.Unmarshal(raw, &req); err != nil {
				c.sendAuthError("无效的认证数据")
				continue
			}
			c.handleAuth(&req)

		case TypeMessageSend:
			if !c.authed {
				c.send <- &ErrorEvent{Type: TypeError, Message: "请先认证"}
				continue
			}
			var req MessageSendRequest
			if err := json.Unmarshal(raw, &req); err != nil {
				c.send <- &ErrorEvent{Type: TypeError, Message: "无效的消息数据"}
				continue
			}
			c.handleMessageSend(&req)

		case TypePong:
			c.conn.SetReadDeadline(time.Now().Add(pongWait))

		case TypePing:
			c.conn.SetReadDeadline(time.Now().Add(pongWait))
			c.send <- &PongEvent{Type: TypePong}

		default:
			c.send <- &ErrorEvent{Type: TypeError, Message: "未知的消息类型"}
		}
	}
}

// WritePump 写入消息到客户端
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("marshal error: %v", err)
				continue
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleAuth 处理认证消息
func (c *Client) handleAuth(req *AuthRequest) {
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(c.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		c.sendAuthError("无效的认证令牌")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.sendAuthError("无效的令牌声明")
		return
	}

	userID, _ := claims["user_id"].(float64)
	c.userID = int64(userID)
	c.authed = true

	c.hub.Register(c)
	c.send <- &AuthOKEvent{Type: TypeAuthOK}
}

func (c *Client) sendAuthError(msg string) {
	c.send <- &AuthErrorEvent{Type: TypeAuthError, Message: msg}
}

// handleMessageSend 处理发送消息
func (c *Client) handleMessageSend(req *MessageSendRequest) {
	msg, err := c.sendMsg(c.userID, req.ReceiverID, req.Content, req.ClientMsgID)
	if err != nil {
		c.send <- &ErrorEvent{Type: TypeError, Message: "发送消息失败"}
		return
	}

	// 发送 message.sent 给发送者（扁平格式）
	c.send <- &MessageSentEvent{
		Type:        TypeMessageSent,
		ID:          msg.ID,
		SenderID:    msg.SenderID,
		ReceiverID:  msg.ReceiverID,
		Content:     msg.Content,
		ContentType: msg.ContentType,
		Status:      msg.Status,
		CreatedAt:   msg.CreatedAt.Format(time.RFC3339),
		ClientMsgID: req.ClientMsgID,
	}

	// 发送 message.new 给接收者（扁平格式）
	c.hub.SendTo(req.ReceiverID, &MessageNewEvent{
		Type:        TypeMessageNew,
		ID:          msg.ID,
		SenderID:    msg.SenderID,
		ReceiverID:  msg.ReceiverID,
		Content:     msg.Content,
		ContentType: msg.ContentType,
		Status:      msg.Status,
		CreatedAt:   msg.CreatedAt.Format(time.RFC3339),
	})
}
