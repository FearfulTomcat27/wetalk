package controller

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"wetalk/common"
	"wetalk/model"
	"wetalk/type"
)

// UploadHandler 上传处理器
type UploadHandler struct {
	ossClient *common.Client
}

// NewUploadHandler 创建上传处理器
func NewUploadHandler(ossClient *common.Client) *UploadHandler {
	return &UploadHandler{ossClient: ossClient}
}

// Upload 上传文件/图片到 OSS
func (h *UploadHandler) Upload(c *gin.Context) {
	userID := c.GetInt64("user_id")

	fileType := c.PostForm("type")
	if fileType != model.ContentTypeImage && fileType != model.ContentTypeFile {
		common.AppError(c, types.ErrInvalidUploadType)
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		common.AppError(c, types.ErrGetFileFailed)
		return
	}
	defer file.Close()

	mimeType := header.Header.Get("Content-Type")

	if fileType == model.ContentTypeImage {
		if !strings.HasPrefix(mimeType, "image/") {
			common.AppError(c, types.ErrNotImage)
			return
		}
		if header.Size > 10*1024*1024 {
			common.AppError(c, types.ErrImageTooLarge)
			return
		}
	} else {
		if header.Size > 20*1024*1024 {
			common.AppError(c, types.ErrFileTooLarge)
			return
		}
	}

	key := fmt.Sprintf("uploads/%s/%d/%d_%s", fileType, userID, time.Now().UnixMilli(), header.Filename)

	url, err := h.ossClient.PutObject(c.Request.Context(), key, file, mimeType, header.Size)
	if err != nil {
		common.Error(c, http.StatusInternalServerError, "上传失败")
		return
	}

	common.Success(c, http.StatusOK, "上传成功", gin.H{
		"url":          url,
		"content_type": fileType,
		"file_name":    header.Filename,
		"file_size":    header.Size,
		"file_type":    mimeType,
	})
}
