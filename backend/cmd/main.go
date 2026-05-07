package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"wetalk/config"
	"wetalk/db"
	"wetalk/internal/friend"
	"wetalk/internal/message"
	"wetalk/internal/middleware"
	"wetalk/internal/user"
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

	// 启动 HTTP 服务
	r := gin.Default()

	// CORS 中间件：允许前端跨域请求
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// ping
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// 认证路由（无需 JWT）
	userSvc := user.NewService(cfg.JWT.Secret, cfg.JWT.ExpireHours)
	userHandler := user.NewHandler(userSvc)
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", userHandler.Register)
		auth.POST("/login", userHandler.Login)
	}

	// 需要认证的路由
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
	{
		// 当前用户
		api.GET("/me", func(c *gin.Context) {
			userID := c.GetInt64("user_id")
			c.JSON(200, gin.H{"code": 200, "data": gin.H{"user_id": userID}})
		})

		// 用户搜索
		api.GET("/users", userHandler.Search)

		// 好友管理
		friendSvc := friend.NewService()
		friendHandler := friend.NewHandler(friendSvc)
		friends := api.Group("/friends")
		{
			friends.POST("", friendHandler.Add)
			friends.GET("", friendHandler.List)
			friends.PUT("/:id/accept", friendHandler.Accept)
			friends.DELETE("/:id", friendHandler.Delete)
		}

		// 消息管理
		msgSvc := message.NewService()
		msgHandler := message.NewHandler(msgSvc)
		messages := api.Group("/messages")
		{
			messages.POST("", msgHandler.Send)
			messages.GET("", msgHandler.List)
		}
	}

	log.Println("服务启动在 :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
