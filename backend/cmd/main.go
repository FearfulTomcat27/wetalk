package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wetalk/config"
	"wetalk/db"
	"wetalk/internal/friend"
	"wetalk/internal/message"
	"wetalk/internal/router"
	"wetalk/internal/user"
	"wetalk/internal/ws"
	"wetalk/pkg/oss"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	if err := db.Init(cfg.Database); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer db.Close()

	// 初始化 Redis
	if err := db.InitRedis(cfg.Redis); err != nil {
		log.Fatalf("初始化 Redis 失败: %v", err)
	}
	defer db.CloseRedis()

	// 创建 Hub 并启动
	hub := ws.NewHub()
	go hub.Run()

	// 初始化 OSS 客户端
	ossClient := oss.NewClient(cfg.OSS)

	// 创建服务
	msgSvc := message.NewService()
	userSvc := user.NewService(cfg.JWT.Secret, cfg.JWT.ExpireHours, ossClient)
	friendSvc := friend.NewService()

	// WS 发送消息回调（桥接 ws 包与 message 包，避免循环依赖）
	sendMsgFunc := func(senderID int64, receiverID int64, content string, clientMsgID string) (*ws.SentMessage, error) {
		msg, err := msgSvc.SendMessage(senderID, message.SendMessageRequest{
			ReceiverID: receiverID,
			Content:    content,
		})
		if err != nil {
			return nil, err
		}
		return &ws.SentMessage{
			ID:          msg.ID,
			SenderID:    msg.SenderID,
			ReceiverID:  msg.ReceiverID,
			Content:     msg.Content,
			ContentType: msg.ContentType,
			Status:      msg.Status,
			CreatedAt:   msg.CreatedAt,
		}, nil
	}

	// 创建 Handler
	userHandler := user.NewHandler(userSvc)
	friendHandler := friend.NewHandler(friendSvc)
	msgHandler := message.NewHandler(msgSvc, hub)

	// 注册路由
	r := router.Setup(&router.Dependencies{
		Config:        cfg,
		Hub:           hub,
		SendMsgFunc:   sendMsgFunc,
		UserHandler:   userHandler,
		FriendHandler: friendHandler,
		MsgHandler:    msgHandler,
	})

	// 启动 HTTP 服务
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Println("服务启动在 :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务...")

	// 逆序关闭：HTTP → Hub → DB/Redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("服务强制关闭: %v", err)
	}

	hub.Shutdown()

	log.Println("服务已关闭")
}
