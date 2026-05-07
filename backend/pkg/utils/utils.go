package utils

import "github.com/gin-gonic/gin"

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, httpStatus int, message string, data interface{}) {
	c.JSON(httpStatus, Response{
		Code:    httpStatus,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, httpStatus int, message string) {
	c.JSON(httpStatus, Response{
		Code:    httpStatus,
		Message: message,
	})
}
