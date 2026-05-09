package ws

import "time"

const (
	// WS 消息类型
	TypeAuth        = "auth"
	TypeAuthOK      = "auth.ok"
	TypeAuthError   = "auth.error"
	TypeMessageSend = "message.send"
	TypeMessageSent = "message.sent"
	TypeMessageNew  = "message.new"
	TypePong        = "pong"
	TypePing        = "ping"
	TypeError       = "error"
)

// --- C2S 请求结构体（扁平格式）---

// AuthRequest 认证请求
type AuthRequest struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

// WSFileMetadata WS 消息中的文件元数据（与 message.FileMetadataPayload 字段一致，避免循环依赖）
type WSFileMetadata struct {
	URL          string `json:"url"`
	OriginalName string `json:"original_name"`
	FileSize     int64  `json:"file_size"`
	MimeType     string `json:"mime_type"`
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
}

// MessageSendRequest 发送消息请求
type MessageSendRequest struct {
	Type         string          `json:"type"`
	ReceiverID   int64           `json:"receiver_id"`
	Content      string          `json:"content"`
	ContentType  string          `json:"content_type,omitempty"`
	ClientMsgID  string          `json:"client_msg_id,omitempty"`
	FileMetadata *WSFileMetadata `json:"file_metadata,omitempty"`
}

// --- S2C 事件结构体（扁平格式）---

// AuthOKEvent 认证成功
type AuthOKEvent struct {
	Type string `json:"type"`
}

// AuthErrorEvent 认证失败
type AuthErrorEvent struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// ErrorEvent 通用错误
type ErrorEvent struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// MessageNewEvent 新消息通知（发给接收者）
type MessageNewEvent struct {
	Type         string          `json:"type"`
	ID           int64           `json:"id"`
	SenderID     int64           `json:"sender_id"`
	ReceiverID   int64           `json:"receiver_id"`
	Content      string          `json:"content"`
	ContentType  string          `json:"content_type"`
	FileMetadata *WSFileMetadata `json:"file_metadata,omitempty"`
	Status       string          `json:"status"`
	CreatedAt    time.Time       `json:"created_at"`
}

// MessageSentEvent 消息发送确认（发给发送者）
type MessageSentEvent struct {
	Type         string          `json:"type"`
	ID           int64           `json:"id"`
	SenderID     int64           `json:"sender_id"`
	ReceiverID   int64           `json:"receiver_id"`
	Content      string          `json:"content"`
	ContentType  string          `json:"content_type"`
	FileMetadata *WSFileMetadata `json:"file_metadata,omitempty"`
	Status       string          `json:"status"`
	CreatedAt    time.Time       `json:"created_at"`
	ClientMsgID  string          `json:"client_msg_id,omitempty"`
}

// PongEvent pong 响应
type PongEvent struct {
	Type string `json:"type"`
}
