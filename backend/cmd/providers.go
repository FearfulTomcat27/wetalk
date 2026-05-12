package main

import (
	"wetalk/common"
	"wetalk/config"
	"wetalk/ws"
)

// NewHub 创建 WebSocket Hub（用于 wire 注入）
func NewHub() *ws.Hub {
	return ws.NewHub(nil)
}

// NewOSSClient 创建 OSS 客户端（用于 wire 注入）
func NewOSSClient(cfg *config.Config) *common.Client {
	return common.NewClient(cfg.OSS)
}
