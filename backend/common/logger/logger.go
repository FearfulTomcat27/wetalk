package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Config 日志配置
type Config struct {
	Level     string `yaml:"level"`      // debug | info | warn | error
	JSON      bool   `yaml:"json"`       // true=JSON 格式, false=文本格式
	AddSource bool   `yaml:"add_source"` // 是否添加源码位置
}

// Init 初始化全局 logger
func Init(cfg Config) {
	var level slog.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	var w io.Writer = os.Stdout
	var handler slog.Handler

	if cfg.JSON {
		handler = slog.NewJSONHandler(w, opts)
	} else {
		handler = slog.NewTextHandler(w, opts)
	}

	slog.SetDefault(slog.New(handler))
}

// Module 返回带 module 属性的 logger
func Module(module string) *slog.Logger {
	return slog.Default().With("module", module)
}

// 通用 context key 类型
type ctxKey string

const (
	ctxKeyReqID  ctxKey = "req_id"
	ctxKeyUserID ctxKey = "user_id"
)

// WithReqID 将 req_id 注入 context
func WithReqID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, ctxKeyReqID, reqID)
}

// Ctx 从 context 中提取常用字段（req_id, user_id）并返回带这些字段的 logger
// 使用方式: logger.Ctx(ctx, "module").Info("msg")
func Ctx(ctx context.Context, module string) *slog.Logger {
	l := slog.Default().With("module", module)
	if ctx != nil {
		if reqID, ok := ctx.Value(ctxKeyReqID).(string); ok && reqID != "" {
			l = l.With("req_id", reqID)
		}
		if userID, ok := ctx.Value(ctxKeyUserID).(int64); ok && userID != 0 {
			l = l.With("user_id", userID)
		}
	}
	return l
}
