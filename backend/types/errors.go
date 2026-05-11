package types

import (
	"fmt"
	"net/http"
)

// ErrorCode 类型化业务错误码（字符串常量，自文档化）
type ErrorCode string

const (
	// 通用错误
	CodeInvalidParam ErrorCode = "invalid_param"
	CodeInternal     ErrorCode = "internal_error"
	CodeNotFound     ErrorCode = "not_found"
	CodeConflict     ErrorCode = "conflict"
	CodeUnauthorized ErrorCode = "unauthorized"

	// 用户模块
	CodeUserAlreadyExists   ErrorCode = "user_already_exists"
	CodeInvalidCredentials  ErrorCode = "invalid_credentials"
	CodeInvalidAvatarFormat ErrorCode = "invalid_avatar_format"
	CodeAvatarTooLarge      ErrorCode = "avatar_too_large"

	// JWT 认证
	CodeMissingToken       ErrorCode = "missing_token"
	CodeInvalidTokenFormat ErrorCode = "invalid_token_format"
	CodeInvalidToken       ErrorCode = "invalid_token"
	CodeInvalidClaims      ErrorCode = "invalid_claims"

	// 好友模块
	CodeFriendRequestNotFound ErrorCode = "friend_request_not_found"
	CodeFriendRequestExists   ErrorCode = "friend_request_exists"
	CodeAddSelf               ErrorCode = "add_self"
	CodeFriendOpForbidden     ErrorCode = "friend_op_forbidden"
	CodeRequestAlreadyHandled ErrorCode = "request_already_handled"
	CodeFriendshipNotFound    ErrorCode = "friendship_not_found"

	// 消息模块
	CodeNotChatMember         ErrorCode = "not_chat_member"
	CodeQuotedMessageNotFound ErrorCode = "quoted_message_not_found"

	// 上传模块
	CodeInvalidUploadType ErrorCode = "invalid_upload_type"
	CodeGetFileFailed     ErrorCode = "get_file_failed"
	CodeNotImage          ErrorCode = "not_image"
	CodeImageTooLarge     ErrorCode = "image_too_large"
	CodeFileTooLarge      ErrorCode = "file_too_large"
)

// AppError 业务错误
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	HTTPStatus int       `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 无底层 error，返回 nil；此处让 errors.As 通过指针匹配正常工作
// 保持接口一致性，便于未来扩展
func (e *AppError) Unwrap() error { return nil }

// NewInvalidParamf 创建带动态详细信息的参数校验错误
func NewInvalidParamf(format string, args ...interface{}) *AppError {
	return &AppError{
		Code:       CodeInvalidParam,
		Message:    fmt.Sprintf("参数校验失败: "+format, args...),
		HTTPStatus: http.StatusBadRequest,
	}
}

// ==================== 通用错误 ====================
var (
	ErrInvalidParam = &AppError{Code: CodeInvalidParam, Message: "请求参数错误", HTTPStatus: http.StatusBadRequest}
	ErrInternal     = &AppError{Code: CodeInternal, Message: "服务器内部错误", HTTPStatus: http.StatusInternalServerError}
	ErrNotFound     = &AppError{Code: CodeNotFound, Message: "资源未找到", HTTPStatus: http.StatusNotFound}
	ErrConflict     = &AppError{Code: CodeConflict, Message: "资源冲突", HTTPStatus: http.StatusConflict}
	ErrUnauthorized = &AppError{Code: CodeUnauthorized, Message: "认证失败", HTTPStatus: http.StatusUnauthorized}
)

// ==================== 用户模块 ====================
var (
	ErrUserAlreadyExists   = &AppError{Code: CodeUserAlreadyExists, Message: "用户名已存在", HTTPStatus: http.StatusConflict}
	ErrInvalidCredentials  = &AppError{Code: CodeInvalidCredentials, Message: "用户名或密码错误", HTTPStatus: http.StatusUnauthorized}
	ErrInvalidAvatarFormat = &AppError{Code: CodeInvalidAvatarFormat, Message: "仅支持 jpg/png/gif/webp 格式", HTTPStatus: http.StatusBadRequest}
	ErrAvatarTooLarge      = &AppError{Code: CodeAvatarTooLarge, Message: "头像文件不能超过 2MB", HTTPStatus: http.StatusBadRequest}
)

// ==================== JWT 认证模块 ====================
var (
	ErrMissingToken       = &AppError{Code: CodeMissingToken, Message: "缺少认证令牌", HTTPStatus: http.StatusUnauthorized}
	ErrInvalidTokenFormat = &AppError{Code: CodeInvalidTokenFormat, Message: "认证格式错误", HTTPStatus: http.StatusUnauthorized}
	ErrInvalidToken       = &AppError{Code: CodeInvalidToken, Message: "无效的认证令牌", HTTPStatus: http.StatusUnauthorized}
	ErrInvalidClaims      = &AppError{Code: CodeInvalidClaims, Message: "无效的令牌声明", HTTPStatus: http.StatusUnauthorized}
)

// ==================== 好友模块 ====================
var (
	ErrFriendRequestNotFound = &AppError{Code: CodeFriendRequestNotFound, Message: "好友请求不存在", HTTPStatus: http.StatusNotFound}
	ErrFriendRequestExists   = &AppError{Code: CodeFriendRequestExists, Message: "好友请求已存在", HTTPStatus: http.StatusConflict}
	ErrAddSelf               = &AppError{Code: CodeAddSelf, Message: "不能添加自己为好友", HTTPStatus: http.StatusBadRequest}
	ErrFriendOpForbidden     = &AppError{Code: CodeFriendOpForbidden, Message: "无权操作", HTTPStatus: http.StatusForbidden}
	ErrRequestAlreadyHandled = &AppError{Code: CodeRequestAlreadyHandled, Message: "请求已处理", HTTPStatus: http.StatusConflict}
	ErrFriendshipNotFound    = &AppError{Code: CodeFriendshipNotFound, Message: "好友关系不存在", HTTPStatus: http.StatusNotFound}
)

// ==================== 消息模块 ====================
var (
	ErrNotChatMember         = &AppError{Code: CodeNotChatMember, Message: "不是聊天成员", HTTPStatus: http.StatusForbidden}
	ErrQuotedMessageNotFound = &AppError{Code: CodeQuotedMessageNotFound, Message: "引用的消息不存在", HTTPStatus: http.StatusBadRequest}
)

// ==================== 上传模块 ====================
var (
	ErrInvalidUploadType = &AppError{Code: CodeInvalidUploadType, Message: "type 参数必须是 image 或 file", HTTPStatus: http.StatusBadRequest}
	ErrGetFileFailed     = &AppError{Code: CodeGetFileFailed, Message: "获取文件失败", HTTPStatus: http.StatusBadRequest}
	ErrNotImage          = &AppError{Code: CodeNotImage, Message: "只允许上传图片文件", HTTPStatus: http.StatusBadRequest}
	ErrImageTooLarge     = &AppError{Code: CodeImageTooLarge, Message: "图片大小不能超过 10MB", HTTPStatus: http.StatusBadRequest}
	ErrFileTooLarge      = &AppError{Code: CodeFileTooLarge, Message: "文件大小不能超过 20MB", HTTPStatus: http.StatusBadRequest}
)
