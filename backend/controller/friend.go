package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"wetalk/common"
	"wetalk/model"
	"wetalk/service"
	"wetalk/types"
	"wetalk/ws"
)

// FriendHandler 好友 HTTP 处理器
type FriendHandler struct {
	svc *service.FriendService
	hub *ws.Hub
}

// NewFriendHandler 创建好友处理器
func NewFriendHandler(hub *ws.Hub) *FriendHandler {
	return &FriendHandler{
		svc: service.NewFriendService(service.NewChatService()),
		hub: hub,
	}
}

// Add 添加好友
func (h *FriendHandler) Add(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req model.AddFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AppError(c, types.NewInvalidParamf(err.Error()))
		return
	}

	result, err := h.svc.AddFriend(userID, req)
	if err != nil {
		var appErr *types.AppError
		if errors.As(err, &appErr) {
			common.AppError(c, appErr)
			return
		}
		common.Error(c, http.StatusInternalServerError, "添加好友失败")
		return
	}

	// 推送给接收方好友请求事件
	h.hub.SendTo(req.FriendID, &ws.FriendRequestNewEvent{
		Type:      ws.TypeFriendRequestNew,
		ID:        result.Request.ID,
		UserID:    result.Sender.ID,
		Username:  result.Sender.Username,
		Nickname:  result.Sender.Nickname,
		Avatar:    result.Sender.Avatar,
		CreatedAt: result.CreatedAt,
	})

	common.Success(c, http.StatusCreated, "好友请求已发送", result.Request)
}

// List 好友列表
func (h *FriendHandler) List(c *gin.Context) {
	userID := c.GetInt64("user_id")

	chatted := c.Query("chatted")
	var friends []model.FriendshipInfo
	var err error
	if chatted == "1" {
		friends, err = h.svc.ListFriends(userID, true)
	} else {
		friends, err = h.svc.ListFriends(userID)
	}
	if err != nil {
		common.Error(c, http.StatusInternalServerError, "获取好友列表失败")
		return
	}

	common.Success(c, http.StatusOK, "成功", friends)
}

// Accept 接受好友请求
func (h *FriendHandler) Accept(c *gin.Context) {
	userID := c.GetInt64("user_id")

	idStr := c.Param("id")
	friendID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		common.AppError(c, types.NewInvalidParamf("无效的请求 ID"))
		return
	}

	if err := h.svc.AcceptFriend(userID, friendID); err != nil {
		var appErr *types.AppError
		if errors.As(err, &appErr) {
			common.AppError(c, appErr)
			return
		}
		common.Error(c, http.StatusInternalServerError, "接受好友请求失败")
		return
	}

	common.Success(c, http.StatusOK, "已接受好友请求", nil)
}

// PendingRequests 获取待处理的好友请求
func (h *FriendHandler) PendingRequests(c *gin.Context) {
	userID := c.GetInt64("user_id")

	requests, err := h.svc.GetPendingRequests(userID)
	if err != nil {
		common.Error(c, http.StatusInternalServerError, "获取待处理请求失败")
		return
	}

	common.Success(c, http.StatusOK, "成功", requests)
}

// Delete 删除好友
func (h *FriendHandler) Delete(c *gin.Context) {
	userID := c.GetInt64("user_id")

	idStr := c.Param("id")
	friendID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		common.AppError(c, types.NewInvalidParamf("无效的 ID"))
		return
	}

	if err := h.svc.DeleteFriend(userID, friendID); err != nil {
		var appErr *types.AppError
		if errors.As(err, &appErr) {
			common.AppError(c, appErr)
			return
		}
		common.Error(c, http.StatusInternalServerError, "删除好友失败")
		return
	}

	common.Success(c, http.StatusOK, "已删除好友", nil)
}
