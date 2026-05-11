package common

import (
	"github.com/gin-gonic/gin"

	"wetalk/type"
)

// Success 成功响应
func Success(c *gin.Context, httpStatus int, message string, data interface{}) {
	c.JSON(httpStatus, types.Response{
		Code:    httpStatus,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, httpStatus int, message string) {
	c.JSON(httpStatus, types.Response{
		Code:    httpStatus,
		Message: message,
	})
}
