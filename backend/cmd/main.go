package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wetalk/common/logger"
	"wetalk/db"
	"wetalk/router"
	"wetalk/ws"
)

// @title WeTalk API
// @version 1.0.0
// @description WeTalk 在线聊天应用 API 文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.wetalk.app/support
// @contact.email support@wetalk.app

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api
// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入 "Bearer " + JWT token

func main() {
	// 加载配置（wire 构建依赖图）
	deps, err := BuildDependencies("config.yaml")
	if err != nil {
		slog.Error("加载配置失败", "err", err)
		os.Exit(1)
	}
	cfg := deps.Config

	// 初始化日志
	logger.Init(logger.Config{
		Level:     cfg.Logger.Level,
		JSON:      cfg.Logger.JSON,
		AddSource: cfg.Logger.AddSource,
	})

	// 初始化数据库
	if err := db.Init(cfg.Database); err != nil {
		slog.Error("初始化数据库失败", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// 初始化 Redis
	if err := db.InitRedis(cfg.Redis); err != nil {
		slog.Error("初始化 Redis 失败", "err", err)
		os.Exit(1)
	}
	defer db.CloseRedis()

	// 初始化 MongoDB
	if err := db.InitMongo(cfg.Mongo); err != nil {
		slog.Error("初始化 MongoDB 失败", "err", err)
		os.Exit(1)
	}
	defer db.CloseMongo()

	// 启动 WebSocket Hub
	hub := deps.Hub
	go hub.Run()

	// 注册路由（wire 已注入 Config、Hub、OSSClient）
	r := router.Setup(deps)

	runServer(r, hub)
}

// runServer 启动 HTTP 服务并等待优雅关闭
func runServer(handler http.Handler, hub *ws.Hub) {
	l := logger.Module("main")

	srv := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	go func() {
		l.Info("服务启动", "addr", ":8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Error("服务启动失败", "err", err)
			os.Exit(1)
		}
	}()

	// 等待关闭信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	l.Info("正在关闭服务...")

	// 逆序关闭：HTTP → Hub → DB/Redis（defer 在 main 中处理）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		l.Error("服务强制关闭", "err", err)
	}

	hub.Shutdown()
	l.Info("服务已关闭")
}
