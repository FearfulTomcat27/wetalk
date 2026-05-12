//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	tcMongo "github.com/testcontainers/testcontainers-go/modules/mongodb"
	tcMySQL "github.com/testcontainers/testcontainers-go/modules/mysql"
	tcRedis "github.com/testcontainers/testcontainers-go/modules/redis"

	"wetalk/config"
	"wetalk/db"
)

var (
	mysqlHost string
	mysqlPort int
	redisHost string
	redisPort int
	mongoURI  string
)

// TestMain 启动 testcontainers 并初始化全局数据库连接
func TestMain(m *testing.M) {
	ctx := context.Background()

	// 启动 MySQL
	fmt.Println("=== 启动 MySQL 容器 ===")
	mysqlC, err := tcMySQL.Run(
		ctx,
		"mysql:8",
		tcMySQL.WithDatabase("wetalk_test"),
		tcMySQL.WithUsername("root"),
		tcMySQL.WithPassword("testpass"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "启动 MySQL 容器失败: %v\n", err)
		os.Exit(1)
	}

	host, _ := mysqlC.Host(ctx)
	port, _ := mysqlC.MappedPort(ctx, "3306/tcp")
	mysqlHost = host
	mysqlPort = port.Int()

	// 启动 Redis
	fmt.Println("=== 启动 Redis 容器 ===")
	redisC, err := tcRedis.Run(
		ctx,
		"redis:7-alpine",
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "启动 Redis 容器失败: %v\n", err)
		mysqlC.Terminate(ctx)
		os.Exit(1)
	}

	rHost, _ := redisC.Host(ctx)
	rPort, _ := redisC.MappedPort(ctx, "6379/tcp")
	redisHost = rHost
	redisPort = rPort.Int()

	// 启动 MongoDB
	fmt.Println("=== 启动 MongoDB 容器 ===")
	mongoC, err := tcMongo.Run(
		ctx,
		"mongo:7",
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "启动 MongoDB 容器失败: %v\n", err)
		mysqlC.Terminate(ctx)
		redisC.Terminate(ctx)
		os.Exit(1)
	}
	mongoURI, _ = mongoC.ConnectionString(ctx)

	// 初始化数据库连接
	fmt.Printf("MySQL: %s:%d, Redis: %s:%d, Mongo: %s\n", mysqlHost, mysqlPort, redisHost, redisPort, mongoURI)

		// MySQL
	if err := db.Init(config.DatabaseConfig{
		Host:         mysqlHost,
		Port:         mysqlPort,
		Username:     "root",
		Password:     "testpass",
		DBName:       "wetalk_test",
		Charset:      "utf8mb4",
		MaxOpenConns: 10,
		MaxIdleConns: 5,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "初始化 MySQL 失败: %v\n", err)
		os.Exit(1)
	}

	// 运行 MySQL DDL 迁移（禁用外键检查）
	fmt.Println("=== 执行 MySQL 迁移 ===")
	migrateMySQL()

	// Redis
	if err := db.InitRedis(config.RedisConfig{
		Host:     redisHost,
		Port:     redisPort,
		Password: "",
		DB:       0,
		PoolSize: 10,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "初始化 Redis 失败: %v\n", err)
		os.Exit(1)
	}

	// MongoDB
	if err := db.InitMongo(config.MongoConfig{
		URI:      mongoURI,
		Database: "wetalk_test",
	}); err != nil {
		fmt.Fprintf(os.Stderr, "初始化 MongoDB 失败: %v\n", err)
		os.Exit(1)
	}

	// 等待数据库就绪
	time.Sleep(time.Second)

	// 运行测试
	code := m.Run()

	// 清理
	fmt.Println("=== 清理容器 ===")
	mysqlC.Terminate(ctx)
	redisC.Terminate(ctx)
	mongoC.Terminate(ctx)

	os.Exit(code)
}

// migrateMySQL 执行 DDL 迁移（禁用外键检查以避免循环引用）
func migrateMySQL() {
	ddls := []string{
		// 按依赖顺序：先创建 messages（chats 的外键依赖它），
		// 但由于有循环依赖，全程禁用外键检查
		`CREATE TABLE IF NOT EXISTS users (
			id bigint NOT NULL AUTO_INCREMENT,
			username varchar(64) NOT NULL,
			password_hash varchar(255) NOT NULL,
			nickname varchar(128) NOT NULL DEFAULT '',
			avatar varchar(512) NOT NULL DEFAULT '',
			created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			UNIQUE KEY idx_username (username)
		) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		`CREATE TABLE IF NOT EXISTS messages (
			id bigint NOT NULL AUTO_INCREMENT,
			chat_id bigint NOT NULL,
			sender_id bigint NOT NULL,
			content text COLLATE utf8mb4_unicode_ci NOT NULL,
			content_type varchar(16) COLLATE utf8mb4_unicode_ci DEFAULT 'text',
			status varchar(16) COLLATE utf8mb4_unicode_ci DEFAULT 'sent',
			created_at datetime DEFAULT CURRENT_TIMESTAMP,
			quote_id bigint DEFAULT NULL,
			PRIMARY KEY (id),
			KEY idx_chat_id (chat_id),
			KEY idx_chat_created (chat_id, created_at)
		) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		`CREATE TABLE IF NOT EXISTS chats (
			id bigint NOT NULL AUTO_INCREMENT,
			chat_type varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'private',
			chat_name varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
			last_message_id bigint DEFAULT NULL,
			last_message_text text COLLATE utf8mb4_unicode_ci,
			last_message_type varchar(16) COLLATE utf8mb4_unicode_ci DEFAULT 'text',
			last_message_time datetime DEFAULT NULL,
			created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id)
		) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		`CREATE TABLE IF NOT EXISTS chat_members (
			id bigint NOT NULL AUTO_INCREMENT,
			chat_id bigint NOT NULL,
			user_id bigint NOT NULL,
			joined_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			UNIQUE KEY uk_chat_user (chat_id, user_id),
			KEY idx_chat_id (chat_id),
			KEY idx_user_id (user_id)
		) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		`CREATE TABLE IF NOT EXISTS friend_requests (
			id bigint NOT NULL AUTO_INCREMENT,
			user_id bigint NOT NULL,
			friend_id bigint NOT NULL,
			status varchar(16) NOT NULL DEFAULT 'pending',
			created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			KEY idx_user_id (user_id),
			KEY idx_friend_id (friend_id)
		) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		`CREATE TABLE IF NOT EXISTS friendships (
			id bigint NOT NULL AUTO_INCREMENT,
			chat_id bigint NOT NULL,
			user1_id bigint NOT NULL,
			user2_id bigint NOT NULL,
			created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			UNIQUE KEY uk_user1_user2 (user1_id, user2_id),
			KEY idx_user1_id (user1_id),
			KEY idx_user2_id (user2_id)
		) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}

	// 禁用外键检查
	db.DB.Exec("SET foreign_key_checks = 0")

	for _, ddl := range ddls {
		if err := db.DB.Exec(ddl).Error; err != nil {
			fmt.Fprintf(os.Stderr, "DDL 迁移失败: %v\nSQL: %s\n", err, ddl[:60])
		}
	}

	// 重新启用外键检查
	db.DB.Exec("SET foreign_key_checks = 1")

	fmt.Println("MySQL 迁移完成")
}
