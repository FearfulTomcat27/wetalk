package types

import (
	"fmt"
	"net/http"
)

// AppError 业务错误
type AppError struct {
	Code       int    `json:"-"`
	Message    string `json:"-"`
	HTTPStatus int    `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// NewInvalidParamf 创建带动态详细信息的参数校验错误
func NewInvalidParamf(format string, args ...interface{}) *AppError {
	return &AppError{
		Code:       90001,
		Message:    fmt.Sprintf("参数校验失败: "+format, args...),
		HTTPStatus: http.StatusBadRequest,
	}
}

// ==================== 通用错误 (9xxxx) ====================
var (
	ErrInvalidParam = &AppError{Code: 90001, Message: "请求参数错误", HTTPStatus: http.StatusBadRequest}
	ErrInternal     = &AppError{Code: 90002, Message: "服务器内部错误", HTTPStatus: http.StatusInternalServerError}
	ErrNotFound     = &AppError{Code: 90004, Message: "资源未找到", HTTPStatus: http.StatusNotFound}
	ErrConflict     = &AppError{Code: 90005, Message: "资源冲突", HTTPStatus: http.StatusConflict}
	ErrUnauthorized = &AppError{Code: 90006, Message: "认证失败", HTTPStatus: http.StatusUnauthorized}
)

// ==================== 用户模块 (10xxx) ====================
var (
	ErrUserAlreadyExists   = &AppError{Code: 10002, Message: "用户名已存在", HTTPStatus: http.StatusConflict}
	ErrInvalidCredentials  = &AppError{Code: 10003, Message: "用户名或密码错误", HTTPStatus: http.StatusUnauthorized}
	ErrInvalidAvatarFormat = &AppError{Code: 10004, Message: "仅支持 jpg/png/gif/webp 格式", HTTPStatus: http.StatusBadRequest}
	ErrAvatarTooLarge      = &AppError{Code: 10005, Message: "头像文件不能超过 2MB", HTTPStatus: http.StatusBadRequest}
)

// ==================== JWT 认证模块 (11xxx) ====================
var (
	ErrMissingToken       = &AppError{Code: 11001, Message: "缺少认证令牌", HTTPStatus: http.StatusUnauthorized}
	ErrInvalidTokenFormat = &AppError{Code: 11002, Message: "认证格式错误", HTTPStatus: http.StatusUnauthorized}
	ErrInvalidToken       = &AppError{Code: 11003, Message: "无效的认证令牌", HTTPStatus: http.StatusUnauthorized}
	ErrInvalidClaims      = &AppError{Code: 11004, Message: "无效的令牌声明", HTTPStatus: http.StatusUnauthorized}
)

// ==================== 好友模块 (20xxx) ====================
var (
	ErrFriendRequestNotFound = &AppError{Code: 20001, Message: "好友请求不存在", HTTPStatus: http.StatusNotFound}
	ErrFriendRequestExists   = &AppError{Code: 20002, Message: "好友请求已存在", HTTPStatus: http.StatusConflict}
	ErrAddSelf               = &AppError{Code: 20003, Message: "不能添加自己为好友", HTTPStatus: http.StatusBadRequest}
	ErrFriendOpForbidden     = &AppError{Code: 20004, Message: "无权操作", HTTPStatus: http.StatusForbidden}
	ErrRequestAlreadyHandled = &AppError{Code: 20005, Message: "请求已处理", HTTPStatus: http.StatusConflict}
	ErrFriendshipNotFound    = &AppError{Code: 20006, Message: "好友关系不存在", HTTPStatus: http.StatusNotFound}
)

// ==================== 消息模块 (30xxx) ====================
var (
	ErrNotChatMember         = &AppError{Code: 30001, Message: "不是聊天成员", HTTPStatus: http.StatusForbidden}
	ErrQuotedMessageNotFound = &AppError{Code: 30002, Message: "引用的消息不存在", HTTPStatus: http.StatusBadRequest}
)

// ==================== 上传模块 (50xxx) ====================
var (
	ErrInvalidUploadType = &AppError{Code: 50001, Message: "type 参数必须是 image 或 file", HTTPStatus: http.StatusBadRequest}
	ErrGetFileFailed     = &AppError{Code: 50002, Message: "获取文件失败", HTTPStatus: http.StatusBadRequest}
	ErrNotImage          = &AppError{Code: 50003, Message: "只允许上传图片文件", HTTPStatus: http.StatusBadRequest}
	ErrImageTooLarge     = &AppError{Code: 50004, Message: "图片大小不能超过 10MB", HTTPStatus: http.StatusBadRequest}
	ErrFileTooLarge      = &AppError{Code: 50005, Message: "文件大小不能超过 20MB", HTTPStatus: http.StatusBadRequest}
)
