package main

import (
	"wetalk/common"
	"wetalk/config"
	"wetalk/dto"
	"wetalk/service"
	"wetalk/ws"
)

// NewHub 创建 WebSocket Hub（用于 wire 注入）
// GetMemberIDs 回调在 router.Setup 中通过 SetGetMemberIDs 设置，避免循环依赖
func NewHub() *ws.Hub {
	return ws.NewHub(nil)
}

// ProvideJWTConfig 从全局配置中提取 JWT 配置
func ProvideJWTConfig(cfg *config.Config) config.JWTConfig {
	return cfg.JWT
}

// ProvideOSSConfig 从全局配置中提取 OSS 配置
func ProvideOSSConfig(cfg *config.Config) config.OSSConfig {
	return cfg.OSS
}

// ProvideUserRepository 提供 UserRepository 接口实现
func ProvideUserRepository() service.UserRepository {
	return dto.User
}

// ProvideChatRepository 提供 ChatRepository 接口实现
func ProvideChatRepository() service.ChatRepository {
	return dto.Chat
}

// ProvideCache 提供 Cache 接口实现
func ProvideCache() service.Cache {
	return &common.RedisCache{}
}

// ProvideObjectStorage 将 *common.Client 适配为 ObjectStorage 接口
func ProvideObjectStorage(client *common.Client) service.ObjectStorage {
	return client
}

// ProvideFriendRepository 提供 FriendRepository 接口实现
func ProvideFriendRepository() service.FriendRepository {
	return dto.Friend
}

// ProvideMessageRepository 提供 MessageRepository 接口实现
func ProvideMessageRepository() service.MessageRepository {
	return dto.Message
}
