//go:build wireinject

package main

import (
	"github.com/google/wire"

	"wetalk/config"
	"wetalk/router"
)

// BuildDependencies 由 wire 自动构建依赖图
func BuildDependencies(cfgPath string) (*router.Dependencies, error) {
	panic(wire.Build(
		config.Load,
		NewHub,
		NewOSSClient,
		wire.Struct(new(router.Dependencies), "*"),
	))
}
