package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"wetalk/common"
	"wetalk/model"
	"wetalk/service"
	"wetalk/type"
)

// maxAvatarSize 头像文件最大 2MB
const maxAvatarSize = 2 << 20

// UserHandler 用户 HTTP 处理器
type UserHandler struct {
	svc *service.UserService
}

// NewUserHandler 创建用户 HTTP 处理器
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Register 注册
func (h *UserHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AppError(c, types.NewInvalidParamf(err.Error()))
		return
	}

	resp, err := h.svc.Register(req)
	if err != nil {
		var appErr *types.AppError
		if errors.As(err, &appErr) {
			common.AppError(c, appErr)
			return
		}
		common.Error(c, http.StatusInternalServerError, "注册失败")
		return
	}

	common.Success(c, http.StatusCreated, "注册成功", resp)
}

// Login 登录
func (h *UserHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AppError(c, types.NewInvalidParamf(err.Error()))
		return
	}

	resp, err := h.svc.Login(req)
	if err != nil {
		var appErr *types.AppError
		if errors.As(err, &appErr) {
			common.AppError(c, appErr)
			return
		}
		common.Error(c, http.StatusInternalServerError, "登录失败")
		return
	}

	common.Success(c, http.StatusOK, "登录成功", resp)
}

// Me 获取当前用户完整信息
func (h *UserHandler) Me(c *gin.Context) {
	userID := c.GetInt64("user_id")

	user, err := h.svc.GetUserByID(userID)
	if err != nil {
		common.Error(c, http.StatusInternalServerError, "查询用户失败")
		return
	}
	if user == nil {
		common.AppError(c, types.ErrNotFound)
		return
	}

	common.Success(c, http.StatusOK, "成功", gin.H{
		"user": user,
	})
}

// Search 搜索用户（用于添加好友）
func (h *UserHandler) Search(c *gin.Context) {
	keyword := c.Query("keyword")

	users, err := h.svc.SearchUsers(keyword)
	if err != nil {
		common.Error(c, http.StatusInternalServerError, "搜索用户失败")
		return
	}

	if users == nil {
		users = []model.User{}
	}
	common.Success(c, http.StatusOK, "成功", users)
}

// UploadAvatar 上传头像
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID := c.GetInt64("user_id")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		common.AppError(c, types.ErrGetFileFailed)
		return
	}
	defer file.Close()

	if header.Size > maxAvatarSize {
		common.AppError(c, types.ErrAvatarTooLarge)
		return
	}

	avatarURL, err := h.svc.UploadAvatar(c.Request.Context(), userID, file, header.Filename, header.Size)
	if err != nil {
		var appErr *types.AppError
		if errors.As(err, &appErr) {
			common.AppError(c, appErr)
			return
		}
		common.Error(c, http.StatusInternalServerError, "上传头像失败")
		return
	}

	common.Success(c, http.StatusOK, "上传成功", gin.H{"avatar": avatarURL})
}
