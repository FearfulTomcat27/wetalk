package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wetalk/common"
	"wetalk/config"
	"wetalk/controller"
	"wetalk/db"
	"wetalk/router"
	"wetalk/service"
	"wetalk/ws"
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
	hub := ws.NewHub(nil)
	go hub.Run()

	// 初始化 OSS 客户端
	ossClient := common.NewClient(cfg.OSS)

	// 创建服务
	chatSvc := service.NewChatService()
	msgSvc := service.NewMessageService(chatSvc)
	userSvc := service.NewUserService(cfg.JWT.Secret, cfg.JWT.ExpireHours, ossClient)
	friendSvc := service.NewFriendService(chatSvc)

	// 注入 getMemberIDs 回调到 Hub（打破 ws → chat 循环依赖）
	hub.SetGetMemberIDs(chatSvc.GetMemberIDs)

	// WS 发送消息回调（桥接 ws 包与 message 包的类型转换）
	sendMsgFunc := service.NewSendMessageFunc(msgSvc)

	// 创建 Handler
	userHandler := controller.NewUserHandler(userSvc)
	friendHandler := controller.NewFriendHandler(friendSvc)
	msgHandler := controller.NewMessageHandler(msgSvc, chatSvc, hub)
	uploadHandler := controller.NewUploadHandler(ossClient)

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

	runServer(r, hub)
}

// runServer 启动 HTTP 服务并等待优雅关闭
func runServer(handler http.Handler, hub *ws.Hub) {
	srv := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	go func() {
		log.Println("服务启动在 :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 等待关闭信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务...")

	// 逆序关闭：HTTP → Hub → DB/Redis（defer 在 main 中处理）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("服务强制关闭: %v", err)
	}

	hub.Shutdown()
	log.Println("服务已关闭")
}
