package types

// Response 统一响应结构
type Response struct {
	Code    interface{} `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
