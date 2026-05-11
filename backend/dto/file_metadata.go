package dto

import (
	"errors"

	"gorm.io/gorm"

	"wetalk/db"
	"wetalk/model"
)

// fileMetadataRepo 文件元数据数据仓库
type fileMetadataRepo struct{}

// FileMetadata 文件元数据仓库实例
var FileMetadata = &fileMetadataRepo{}

// Create 创建文件元数据
func (r *fileMetadataRepo) Create(msgID int64, url string, originalName string, fileSize int64, mimeType string, width int, height int) (*model.FileMetadata, error) {
	fm := &model.FileMetadata{
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
func (r *fileMetadataRepo) GetByMessageID(msgID int64) (*model.FileMetadata, error) {
	var fm model.FileMetadata
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
func (r *fileMetadataRepo) ListByMessageIDs(msgIDs []int64) ([]model.FileMetadata, error) {
	if len(msgIDs) == 0 {
		return nil, nil
	}
	var list []model.FileMetadata
	err := db.DB.Where("message_id IN ?", msgIDs).Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// DeleteByMessageID 删除消息的文件元数据
func (r *fileMetadataRepo) DeleteByMessageID(msgID int64) error {
	return db.DB.Where("message_id = ?", msgID).Delete(&model.FileMetadata{}).Error
}
