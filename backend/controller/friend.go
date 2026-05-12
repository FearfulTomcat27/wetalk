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
func NewFriendHandler(svc *service.FriendService, hub *ws.Hub) *FriendHandler {
	return &FriendHandler{
		svc: svc,
		hub: hub,
	}
}

// Add 添加好友
// @Summary 发送好友请求
// @Description 向指定用户发送好友请求，并通过 WebSocket 实时推送给接收方
// @Tags 好友
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.AddFriendRequest true "好友请求"
// @Success 201 {object} types.Response{data=model.FriendRequest} "好友请求已发送"
// @Failure 400 {object} types.Response "参数错误或不能添加自己"
// @Failure 401 {object} types.Response "未认证"
// @Failure 409 {object} types.Response "请求已存在"
// @Router /api/friends [post]
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
// @Summary 获取好友列表
// @Description 获取当前用户的好友列表，可选仅显示有聊天的好友
// @Tags 好友
// @Produce json
// @Security BearerAuth
// @Param chatted query int false "是否仅显示有聊天的好友 (1=是)"
// @Success 200 {object} types.Response{data=[]model.FriendshipInfo} "成功"
// @Failure 401 {object} types.Response "未认证"
// @Router /api/friends [get]
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
// @Summary 接受好友请求
// @Description 接受一个待处理的好友请求，建立好友关系并创建聊天会话
// @Tags 好友
// @Produce json
// @Security BearerAuth
// @Param id path int true "好友请求 ID"
// @Success 200 {object} types.Response "已接受好友请求"
// @Failure 400 {object} types.Response "参数错误"
// @Failure 401 {object} types.Response "未认证"
// @Failure 403 {object} types.Response "无权操作"
// @Router /api/friends/{id}/accept [put]
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
// @Summary 获取待处理的好友请求
// @Description 获取当前用户收到的待处理好友请求列表
// @Tags 好友
// @Produce json
// @Security BearerAuth
// @Success 200 {object} types.Response{data=[]model.PendingRequest} "成功"
// @Failure 401 {object} types.Response "未认证"
// @Router /api/friends/pending [get]
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
// @Summary 删除好友
// @Description 删除与指定用户的好友关系
// @Tags 好友
// @Produce json
// @Security BearerAuth
// @Param id path int true "好友 ID"
// @Success 200 {object} types.Response "已删除好友"
// @Failure 400 {object} types.Response "参数错误"
// @Failure 401 {object} types.Response "未认证"
// @Failure 404 {object} types.Response "好友关系不存在"
// @Router /api/friends/{id} [delete]
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
