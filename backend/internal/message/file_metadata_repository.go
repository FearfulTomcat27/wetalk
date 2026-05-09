package message

import (
	"errors"

	"gorm.io/gorm"

	"wetalk/db"
)

// fileMetadataRepository 文件元数据数据仓库
type fileMetadataRepository struct{}

// FileMetadataRepository 文件元数据仓库实例
var FileMetadataRepository = &fileMetadataRepository{}

// Create 创建文件元数据
func (r *fileMetadataRepository) Create(msgID int64, url string, originalName string, fileSize int64, mimeType string, width int, height int) (*FileMetadata, error) {
	fm := &FileMetadata{
		MessageID:    msgID,
		URL:          url,
		OriginalName: originalName,
		FileSize:     fileSize,
		MimeType:     mimeType,
		Width:        width,
		Height:       height,
	}
	if err := db.DB.Create(fm).Error; err != nil {
		return nil, err
	}
	return fm, nil
}

// GetByMessageID 根据消息ID获取文件元数据
func (r *fileMetadataRepository) GetByMessageID(msgID int64) (*FileMetadata, error) {
	var fm FileMetadata
	err := db.DB.Where("message_id = ?", msgID).First(&fm).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &fm, nil
}

// ListByMessageIDs 批量查询文件元数据
func (r *fileMetadataRepository) ListByMessageIDs(msgIDs []int64) ([]FileMetadata, error) {
	if len(msgIDs) == 0 {
		return nil, nil
	}
	var list []FileMetadata
	err := db.DB.Where("message_id IN ?", msgIDs).Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// DeleteByMessageID 删除消息的文件元数据
func (r *fileMetadataRepository) DeleteByMessageID(msgID int64) error {
	return db.DB.Where("message_id = ?", msgID).Delete(&FileMetadata{}).Error
}
