//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"goflare.io/auth"
	"goflare.io/auth/config"
	"goflare.io/auth/handler"
	"goflare.io/auth/middleware"
	"goflare.io/auth/permission"
	"goflare.io/auth/role"
	"goflare.io/auth/server"
	"goflare.io/auth/user"
)

func InitializeAuthService() (*server.Server, error) {

	wire.Build(
		config.ProvideApplicationConfig,
		config.NewLogger,
		config.ProvidePasetoSecret,
		config.ProvidePostgresConn,
		config.ProvideEnforcer,
		user.NewRepository,
		user.NewService,
		role.NewRepository,
		role.NewService,
		permission.NewRepository,
		permission.NewService,
		auth.NewAuthentication,
		middleware.NewAuthenticationMiddleware,
		handler.NewUserHandler,
		server.NewServer,
	)

	return &server.Server{}, nil
}
