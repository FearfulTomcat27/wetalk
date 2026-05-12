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

// maxAvatarSize 头像文件最大 2MB
const maxAvatarSize = 2 << 20

// UserHandler 用户 HTTP 处理器
type UserHandler struct {
	svc *service.UserService
}

// NewUserHandler 创建用户 HTTP 处理器
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		svc: svc,
	}
}

// Register 注册
// @Summary 用户注册
// @Description 使用用户名和密码注册新用户，自动生成默认头像
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body model.RegisterRequest true "注册信息"
// @Success 201 {object} types.Response{data=model.AuthResponse} "注册成功"
// @Failure 400 {object} types.Response "参数错误"
// @Failure 409 {object} types.Response "用户名已存在"
// @Router /api/auth/register [post]
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
// @Summary 用户登录
// @Description 使用用户名和密码登录，返回 JWT token 和用户信息
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "登录信息"
// @Success 200 {object} types.Response{data=model.AuthResponse} "登录成功"
// @Failure 400 {object} types.Response "参数错误"
// @Failure 401 {object} types.Response "用户名或密码错误"
// @Router /api/auth/login [post]
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
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的完整信息
// @Tags 用户
// @Produce json
// @Security BearerAuth
// @Success 200 {object} types.Response "成功"
// @Failure 401 {object} types.Response "未认证"
// @Router /api/me [get]
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

	common.Success(c, http.StatusOK, "成功", user)
}

// Search 搜索用户（用于添加好友）
// @Summary 搜索用户
// @Description 根据关键词搜索用户（用户名/昵称模糊匹配）
// @Tags 用户
// @Produce json
// @Security BearerAuth
// @Param keyword query string true "搜索关键词"
// @Success 200 {object} types.Response{data=[]model.User} "成功"
// @Failure 401 {object} types.Response "未认证"
// @Router /api/users [get]
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
// @Summary 上传头像
// @Description 上传用户头像图片（仅支持 jpg/png/gif/webp，最大 2MB）
// @Tags 用户
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "头像文件"
// @Success 200 {object} types.Response "上传成功"
// @Failure 400 {object} types.Response "参数错误或格式不支持"
// @Failure 401 {object} types.Response "未认证"
// @Router /api/me/avatar [post]
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
