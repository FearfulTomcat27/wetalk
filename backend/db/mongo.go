package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"wetalk/common/logger"
	"wetalk/config"
)

// MongoClient MongoDB 客户端全局实例
var MongoClient *mongo.Client

// MongoDB 数据库全局实例
var MongoDB *mongo.Database

// InitMongo 初始化 MongoDB 连接
func InitMongo(cfg config.MongoConfig) error {
	if cfg.URI == "" || cfg.Database == "" {
		logger.Module("mongo").Warn("MongoDB 未配置，跳过初始化")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return fmt.Errorf("连接 MongoDB 失败: %w", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("MongoDB 连通性测试失败: %w", err)
	}

	MongoClient = client
	MongoDB = client.Database(cfg.Database)

	// 创建索引
	if err := EnsureIndexes(); err != nil {
		// 索引创建失败不阻塞启动，仅记录警告
		logger.Module("mongo").Warn("创建索引失败", "err", err)
	}

	logger.Module("mongo").Info("MongoDB 连接成功", "database", cfg.Database)
	return nil
}

// CloseMongo 关闭 MongoDB 连接
func CloseMongo() {
	if MongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := MongoClient.Disconnect(ctx); err != nil {
			logger.Module("mongo").Error("关闭 MongoDB 连接失败", "err", err)
		}
	}
}

// MsgCollection 获取消息集合
func MsgCollection() *mongo.Collection {
	return MongoDB.Collection("messages")
}

// CountersCollection 获取计数器集合（用于 msg_id 自增）
func CountersCollection() *mongo.Collection {
	return MongoDB.Collection("counters")
}

// EnsureIndexes 创建必要的索引
func EnsureIndexes() error {
	ctx := context.Background()

	// 消息集合索引
	msgCol := MsgCollection()

	// chat_id + created_at 复合索引（聊天记录分页查询）
	if _, err := msgCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "chat_id", Value: 1},
			{Key: "created_at", Value: -1},
		},
	}); err != nil {
		return fmt.Errorf("创建 idx_chat_created 索引失败: %w", err)
	}

	// chat_id + sender_id + status 复合索引（未读统计和标记已读）
	if _, err := msgCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "chat_id", Value: 1},
			{Key: "sender_id", Value: 1},
			{Key: "status", Value: 1},
		},
	}); err != nil {
		return fmt.Errorf("创建 idx_chat_sender_status 索引失败: %w", err)
	}

	// msg_id 唯一索引（兼容现有 int64 ID 查询）
	if _, err := msgCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "msg_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return fmt.Errorf("创建 idx_msg_id 索引失败: %w", err)
	}

	return nil
}
