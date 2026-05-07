package user

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	pkgerrors "wetalk/pkg/errors"
	"wetalk/pkg/utils"
)

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
