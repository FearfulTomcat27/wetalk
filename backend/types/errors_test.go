package types

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppError_Error(t *testing.T) {
	err := &AppError{Code: CodeInvalidParam, Message: "test", HTTPStatus: http.StatusBadRequest}
	assert.Equal(t, "[invalid_param] test", err.Error())
}

func TestAppError_Unwrap(t *testing.T) {
	err := &AppError{Code: CodeInvalidParam, Message: "test", HTTPStatus: http.StatusBadRequest}
	assert.Nil(t, err.Unwrap())
}

func TestNewInvalidParamf(t *testing.T) {
	err := NewInvalidParamf("username=%s", "admin")
	assert.Equal(t, "参数校验失败: username=admin", err.Message)
	assert.Equal(t, CodeInvalidParam, err.Code)
	assert.Equal(t, http.StatusBadRequest, err.HTTPStatus)
}

func TestErrorsAs(t *testing.T) {
	err := ErrUserAlreadyExists
	var appErr *AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, CodeUserAlreadyExists, appErr.Code)
}

func TestSentinelErrors_Unique(t *testing.T) {
	tests := []struct {
		name       string
		err        *AppError
		wantCode   ErrorCode
		wantStatus int
		wantMsg    string
	}{
		{"InvalidParam", ErrInvalidParam, CodeInvalidParam, http.StatusBadRequest, "请求参数错误"},
		{"NotFound", ErrNotFound, CodeNotFound, http.StatusNotFound, "资源未找到"},
		{"Unauthorized", ErrUnauthorized, CodeUnauthorized, http.StatusUnauthorized, "认证失败"},
		{"UserAlreadyExists", ErrUserAlreadyExists, CodeUserAlreadyExists, http.StatusConflict, "用户名已存在"},
		{"InvalidCredentials", ErrInvalidCredentials, CodeInvalidCredentials, http.StatusUnauthorized, "用户名或密码错误"},
		{"FriendRequestNotFound", ErrFriendRequestNotFound, CodeFriendRequestNotFound, http.StatusNotFound, "好友请求不存在"},
		{"QuotedMessageNotFound", ErrQuotedMessageNotFound, CodeQuotedMessageNotFound, http.StatusBadRequest, "引用的消息不存在"},
		{"ImageTooLarge", ErrImageTooLarge, CodeImageTooLarge, http.StatusBadRequest, "图片大小不能超过 10MB"},
		{"FileTooLarge", ErrFileTooLarge, CodeFileTooLarge, http.StatusBadRequest, "文件大小不能超过 20MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantCode, tt.err.Code)
			assert.Equal(t, tt.wantStatus, tt.err.HTTPStatus)
			assert.Equal(t, tt.wantMsg, tt.err.Message)
		})
	}
}
