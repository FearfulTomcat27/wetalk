package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FileMetadataEmbed 嵌入在消息文档中的文件元数据
type FileMetadataEmbed struct {
	URL          string `bson:"url" json:"url"`
	OriginalName string `bson:"original_name,omitempty" json:"original_name,omitempty"`
	FileSize     int64  `bson:"file_size,omitempty" json:"file_size,omitempty"`
	MimeType     string `bson:"mime_type,omitempty" json:"mime_type,omitempty"`
	Width        int    `bson:"width,omitempty" json:"width,omitempty"`
	Height       int    `bson:"height,omitempty" json:"height,omitempty"`
}

// MessageDoc MongoDB 消息文档结构
// 与 MySQL 的 Message 模型不同：file_metadata 直接嵌入文档中
type MessageDoc struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`
	MsgID             int64              `bson:"msg_id"`
	ChatID            int64              `bson:"chat_id"`
	SenderID          int64              `bson:"sender_id"`
	Content           string             `bson:"content"`
	ContentType       string             `bson:"content_type"`
	Status            string             `bson:"status"`
	QuoteMessageID    *int64             `bson:"quote_message_id,omitempty"`
	QuotedContent     *string            `bson:"quoted_content,omitempty"`
	QuotedSenderID    *int64             `bson:"quoted_sender_id,omitempty"`
	QuotedSenderName  *string            `bson:"quoted_sender_name,omitempty"`
	QuotedContentType *string            `bson:"quoted_content_type,omitempty"`
	QuotedFileMeta    *FileMetadataEmbed `bson:"quoted_file_metadata,omitempty"`
	FileMetadata      *FileMetadataEmbed `bson:"file_metadata,omitempty"`
	CreatedAt         time.Time          `bson:"created_at"`
}

// ToResponse 转换为 API 响应结构体
func (d *MessageDoc) ToResponse() MessageResponse {
	resp := MessageResponse{
		ID:                d.MsgID,
		ChatID:            d.ChatID,
		SenderID:          d.SenderID,
		Content:           d.Content,
		ContentType:       d.ContentType,
		Status:            d.Status,
		QuoteMessageID:    d.QuoteMessageID,
		QuotedContent:     d.QuotedContent,
		QuotedSenderID:    d.QuotedSenderID,
		QuotedSenderName:  d.QuotedSenderName,
		QuotedContentType: d.QuotedContentType,
		CreatedAt:         d.CreatedAt,
	}
	if d.FileMetadata != nil {
		resp.FileMetadata = &FileMetadata{
			URL:          d.FileMetadata.URL,
			OriginalName: d.FileMetadata.OriginalName,
			FileSize:     d.FileMetadata.FileSize,
			MimeType:     d.FileMetadata.MimeType,
			Width:        d.FileMetadata.Width,
			Height:       d.FileMetadata.Height,
		}
	}
	if d.QuotedFileMeta != nil {
		resp.QuotedFileMeta = &FileMetadata{
			URL:          d.QuotedFileMeta.URL,
			OriginalName: d.QuotedFileMeta.OriginalName,
			FileSize:     d.QuotedFileMeta.FileSize,
			MimeType:     d.QuotedFileMeta.MimeType,
			Width:        d.QuotedFileMeta.Width,
			Height:       d.QuotedFileMeta.Height,
		}
	}
	return resp
}
