//go:build wireinject

package main

import (
	"github.com/google/wire"

	"wetalk/common"
	"wetalk/config"
	"wetalk/controller"
	"wetalk/router"
	"wetalk/service"
)

// BuildDependencies 由 wire 自动构建依赖图
func BuildDependencies(cfgPath string) (*router.Dependencies, error) {
	panic(wire.Build(
		config.Load,
		NewHub,
		ProvideJWTConfig,
		ProvideOSSConfig,
		ProvideUserRepository,
		ProvideChatRepository,
		ProvideFriendRepository,
		ProvideMessageRepository,
		ProvideCache,
		ProvideObjectStorage,
		common.NewClient,
		service.NewChatService,
		service.NewUserService,
		service.NewFriendService,
		service.NewMessageService,
		service.NewUploadService,
		controller.NewUserHandler,
		controller.NewFriendHandler,
		controller.NewMessageHandler,
		controller.NewUploadHandler,
		wire.Struct(new(router.Dependencies), "*"),
	))
}
