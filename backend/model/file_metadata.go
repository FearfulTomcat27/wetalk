package model

import "time"

// FileMetadata 文件元数据（MongoDB 迁移后仅作响应结构体，不再映射 MySQL 表）
type FileMetadata struct {
	ID           int64     `json:"id"`
	MessageID    int64     `json:"message_id"`
	URL          string    `json:"url"`
	OriginalName string    `json:"original_name,omitempty"`
	FileSize     int64     `json:"file_size,omitempty"`
	MimeType     string    `json:"mime_type,omitempty"`
	Width        int       `json:"width,omitempty"`
	Height       int       `json:"height,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
