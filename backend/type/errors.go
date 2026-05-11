package types

import "errors"

var (
	// ErrNotFound 资源未找到
	ErrNotFound = errors.New("资源未找到")
	// ErrConflict 资源冲突（如用户名已存在）
	ErrConflict = errors.New("资源冲突")
	// ErrUnauthorized 认证失败
	ErrUnauthorized = errors.New("认证失败")
	// ErrInvalidParam 参数校验失败
	ErrInvalidParam = errors.New("参数校验失败")
	// ErrInternal 内部错误
	ErrInternal = errors.New("内部错误")
)
