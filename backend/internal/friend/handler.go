package friend

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	pkgerrors "wetalk/pkg/errors"
	"wetalk/pkg/utils"
)

// Handler 好友 HTTP 处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建好友处理器
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Add 添加好友
func (h *Handler) Add(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req AddFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数校验失败: "+err.Error())
		return
	}

	friend, err := h.svc.AddFriend(userID, req)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrInvalidParam) {
			utils.Error(c, http.StatusBadRequest, "不能添加自己为好友")
			return
		}
		if errors.Is(err, pkgerrors.ErrConflict) {
			utils.Error(c, http.StatusConflict, "好友请求已存在")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "添加好友失败")
		return
	}

	utils.Success(c, http.StatusCreated, "好友请求已发送", friend)
}

// List 好友列表
func (h *Handler) List(c *gin.Context) {
	userID := c.GetInt64("user_id")

	friends, err := h.svc.ListFriends(userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取好友列表失败")
		return
	}

	if friends == nil {
		friends = []FriendshipInfo{}
	}

	utils.Success(c, http.StatusOK, "成功", friends)
}

// Accept 接受好友请求
func (h *Handler) Accept(c *gin.Context) {
	userID := c.GetInt64("user_id")

	idStr := c.Param("id")
	friendID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的请求 ID")
		return
	}

	if err := h.svc.AcceptFriend(userID, friendID); err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			utils.Error(c, http.StatusNotFound, "好友请求不存在")
			return
		}
		if errors.Is(err, pkgerrors.ErrUnauthorized) {
			utils.Error(c, http.StatusForbidden, "无权操作")
			return
		}
		if errors.Is(err, pkgerrors.ErrConflict) {
			utils.Error(c, http.StatusConflict, "请求已处理")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "接受好友请求失败")
		return
	}

	utils.Success(c, http.StatusOK, "已接受好友请求", nil)
}

// PendingRequests 获取待处理的好友请求
func (h *Handler) PendingRequests(c *gin.Context) {
	userID := c.GetInt64("user_id")

	requests, err := h.svc.GetPendingRequests(userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取待处理请求失败")
		return
	}

	if requests == nil {
		requests = []PendingRequest{}
	}

	utils.Success(c, http.StatusOK, "成功", requests)
}

// Delete 删除好友
func (h *Handler) Delete(c *gin.Context) {
	userID := c.GetInt64("user_id")

	idStr := c.Param("id")
	friendID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的 ID")
		return
	}

	if err := h.svc.DeleteFriend(userID, friendID); err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			utils.Error(c, http.StatusNotFound, "好友关系不存在")
			return
		}
		if errors.Is(err, pkgerrors.ErrUnauthorized) {
			utils.Error(c, http.StatusForbidden, "无权操作")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "删除好友失败")
		return
	}

	utils.Success(c, http.StatusOK, "已删除好友", nil)
}
