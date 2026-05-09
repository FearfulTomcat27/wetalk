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
	sendMsgFunc := func(senderID int64, receiverID int64, content string, contentType string, clientMsgID string, fileMetadata *ws.WSFileMetadata) (*ws.SentMessage, error) {
		var filePayload *message.FileMetadataPayload
		if fileMetadata != nil {
			filePayload = &message.FileMetadataPayload{
				URL:          fileMetadata.URL,
				OriginalName: fileMetadata.OriginalName,
				FileSize:     fileMetadata.FileSize,
				MimeType:     fileMetadata.MimeType,
				Width:        fileMetadata.Width,
				Height:       fileMetadata.Height,
			}
		}
		msgResp, err := msgSvc.SendMessage(senderID, message.SendMessageRequest{
			ReceiverID:   receiverID,
			Content:      content,
			ContentType:  contentType,
			FileMetadata: filePayload,
		})
		if err != nil {
			return nil, err
		}
		var meta *ws.WSFileMetadata
		if msgResp.FileMetadata != nil {
			meta = &ws.WSFileMetadata{
				URL:          msgResp.FileMetadata.URL,
				OriginalName: msgResp.FileMetadata.OriginalName,
				FileSize:     msgResp.FileMetadata.FileSize,
				MimeType:     msgResp.FileMetadata.MimeType,
				Width:        msgResp.FileMetadata.Width,
				Height:       msgResp.FileMetadata.Height,
			}
		}
		return &ws.SentMessage{
			ID:           msgResp.ID,
			SenderID:     msgResp.SenderID,
			ReceiverID:   msgResp.ReceiverID,
			Content:      msgResp.Content,
			ContentType:  msgResp.ContentType,
			FileMetadata: meta,
			Status:       msgResp.Status,
			CreatedAt:    msgResp.CreatedAt,
		}, nil
	}

	// 创建 Handler
	userHandler := user.NewHandler(userSvc)
	friendHandler := friend.NewHandler(friendSvc)
	msgHandler := message.NewHandler(msgSvc, hub)
	uploadHandler := message.NewUploadHandler(ossClient)

	// 注册路由
	r := router.Setup(&router.Dependencies{
		Config:        cfg,
		Hub:           hub,
		SendMsgFunc:   sendMsgFunc,
		UserHandler:   userHandler,
		FriendHandler: friendHandler,
		MsgHandler:    msgHandler,
		UploadHandler: uploadHandler,
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
