package controller

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"wetalk/common"
	"wetalk/model"
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
		common.Error(c, http.StatusBadRequest, "type 参数必须是 image 或 file")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		common.Error(c, http.StatusBadRequest, "获取文件失败")
		return
	}
	defer file.Close()

	mimeType := header.Header.Get("Content-Type")

	if fileType == model.ContentTypeImage {
		if !strings.HasPrefix(mimeType, "image/") {
			common.Error(c, http.StatusBadRequest, "只允许上传图片文件")
			return
		}
		if header.Size > 10*1024*1024 {
			common.Error(c, http.StatusBadRequest, "图片大小不能超过 10MB")
			return
		}
	} else {
		if header.Size > 20*1024*1024 {
			common.Error(c, http.StatusBadRequest, "文件大小不能超过 20MB")
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
