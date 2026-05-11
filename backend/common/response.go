package common

import (
	"github.com/gin-gonic/gin"

	"wetalk/types"
)

// Success 成功响应
func Success(c *gin.Context, httpStatus int, message string, data interface{}) {
	c.JSON(httpStatus, types.Response{
		Code:    httpStatus,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应（使用 HTTP 状态码作为 code，保留向后兼容）
func Error(c *gin.Context, httpStatus int, message string) {
	c.JSON(httpStatus, types.Response{
		Code:    httpStatus,
		Message: message,
	})
}

// AppError 使用业务错误码响应
func AppError(c *gin.Context, appErr *types.AppError) {
	c.JSON(appErr.HTTPStatus, types.Response{
		Code:    appErr.Code,
		Message: appErr.Message,
	})
}
