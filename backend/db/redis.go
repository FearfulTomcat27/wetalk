package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"wetalk/common/logger"
	"wetalk/config"
)

// RedisClient 全局 Redis 客户端
var RedisClient *redis.Client

// InitRedis 初始化 Redis 连接
func InitRedis(cfg config.RedisConfig) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis 连接失败: %w", err)
	}

	logger.Module("redis").Info("Redis 连接成功")
	return nil
}

// GetRedis 获取 Redis 客户端
func GetRedis() *redis.Client {
	return RedisClient
}

// CloseRedis 关闭 Redis 连接
func CloseRedis() {
	if RedisClient != nil {
		if err := RedisClient.Close(); err != nil {
			logger.Module("redis").Error("关闭 Redis 连接失败", "err", err)
		}
	}
}
