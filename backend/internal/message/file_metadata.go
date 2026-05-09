package message

import "time"

// FileMetadata 文件元数据模型
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

// TableName 指定表名
func (FileMetadata) TableName() string {
	return "file_metadata"
}
