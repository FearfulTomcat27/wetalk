package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"wetalk/common"
	"wetalk/model"
	"wetalk/types"
)

// UploadService 文件上传业务逻辑
type UploadService struct {
	ossClient *common.Client
}

// NewUploadService 创建上传服务
func NewUploadService(ossClient *common.Client) *UploadService {
	return &UploadService{ossClient: ossClient}
}

// Upload 上传文件/图片到 OSS（含类型/大小校验、OSS key 生成、上传）
func (s *UploadService) Upload(ctx context.Context, userID int64, fileType string, file io.Reader, filename string, fileSize int64, mimeType string) (*model.UploadResponse, error) {
	if fileType != model.ContentTypeImage && fileType != model.ContentTypeFile {
		return nil, types.ErrInvalidUploadType
	}

	if fileType == model.ContentTypeImage {
		if !strings.HasPrefix(mimeType, "image/") {
			return nil, types.ErrNotImage
		}
		if fileSize > 10*1024*1024 {
			return nil, types.ErrImageTooLarge
		}
	} else {
		if fileSize > 20*1024*1024 {
			return nil, types.ErrFileTooLarge
		}
	}

	key := fmt.Sprintf("uploads/%s/%d/%d_%s", fileType, userID, time.Now().UnixMilli(), filename)

	url, err := s.ossClient.PutObject(ctx, key, file, mimeType, fileSize)
	if err != nil {
		return nil, err
	}

	return &model.UploadResponse{
		URL:         url,
		ContentType: fileType,
		FileName:    filename,
		FileSize:    fileSize,
		FileType:    mimeType,
	}, nil
}
