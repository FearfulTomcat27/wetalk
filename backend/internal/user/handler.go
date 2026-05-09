package user

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	pkgerrors "wetalk/pkg/errors"
	"wetalk/pkg/utils"
)

// maxAvatarSize 头像文件最大 2MB
const maxAvatarSize = 2 << 20

// Handler 用户 HTTP 处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建用户 HTTP 处理器
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register 注册
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数校验失败: "+err.Error())
		return
	}

	resp, err := h.svc.Register(req)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrConflict) {
			utils.Error(c, http.StatusConflict, "用户名已存在")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "注册失败")
		return
	}

	utils.Success(c, http.StatusCreated, "注册成功", resp)
}

// Login 登录
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数校验失败: "+err.Error())
		return
	}

	resp, err := h.svc.Login(req)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrUnauthorized) {
			utils.Error(c, http.StatusUnauthorized, "用户名或密码错误")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "登录失败")
		return
	}

	utils.Success(c, http.StatusOK, "登录成功", resp)
}

// Me 获取当前用户完整信息
func (h *Handler) Me(c *gin.Context) {
	userID := c.GetInt64("user_id")

	user, err := h.svc.GetUserByID(userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询用户失败")
		return
	}
	if user == nil {
		utils.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	utils.Success(c, http.StatusOK, "成功", gin.H{
		"user": user,
	})
}

// Search 搜索用户（用于添加好友）
func (h *Handler) Search(c *gin.Context) {
	keyword := c.Query("keyword")

	users, err := h.svc.SearchUsers(keyword)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "搜索用户失败")
		return
	}

	if users == nil {
		users = []User{}
	}
	utils.Success(c, http.StatusOK, "成功", users)
}

// UploadAvatar 上传头像
func (h *Handler) UploadAvatar(c *gin.Context) {
	userID := c.GetInt64("user_id")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "请选择头像文件")
		return
	}
	defer file.Close()

	if header.Size > maxAvatarSize {
		utils.Error(c, http.StatusBadRequest, "头像文件不能超过 2MB")
		return
	}

	avatarURL, err := h.svc.UploadAvatar(c.Request.Context(), userID, file, header.Filename, header.Size)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrInvalidParam) {
			utils.Error(c, http.StatusBadRequest, "仅支持 jpg/png/gif/webp 格式")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "上传头像失败")
		return
	}

	utils.Success(c, http.StatusOK, "上传成功", gin.H{"avatar": avatarURL})
}
