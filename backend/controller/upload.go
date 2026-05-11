package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"wetalk/common"
	"wetalk/model"
	"wetalk/service"
	"wetalk/types"
)

// UploadHandler 上传处理器
type UploadHandler struct {
	svc *service.UploadService
}

// NewUploadHandler 创建上传处理器
func NewUploadHandler(ossClient *common.Client) *UploadHandler {
	return &UploadHandler{
		svc: service.NewUploadService(ossClient),
	}
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

	resp, err := h.svc.Upload(c.Request.Context(), userID, fileType, file, header.Filename, header.Size, mimeType)
	if err != nil {
		var appErr *types.AppError
		if errors.As(err, &appErr) {
			common.AppError(c, appErr)
			return
		}
		common.Error(c, http.StatusInternalServerError, "上传失败")
		return
	}

	common.Success(c, http.StatusOK, "上传成功", resp)
}
