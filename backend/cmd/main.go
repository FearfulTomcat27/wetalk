package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wetalk/common"
	"wetalk/common/logger"
	"wetalk/config"
	"wetalk/db"
	"wetalk/router"
	"wetalk/ws"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		slog.Error("加载配置失败", "err", err)
		os.Exit(1)
	}

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

	// 创建 Hub 并启动
	hub := ws.NewHub(nil)
	go hub.Run()

	// 初始化 OSS 客户端
	ossClient := common.NewClient(cfg.OSS)

	// 注册路由（内部构造 service 和 Handler）
	r := router.Setup(&router.Dependencies{
		Config:    cfg,
		Hub:       hub,
		OSSClient: ossClient,
	})

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
