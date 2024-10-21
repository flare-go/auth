//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"goflare.io/nexus"

	"goflare.io/auth/internal/authentication"
	"goflare.io/auth/internal/authorization"
	"goflare.io/auth/internal/firebase"
	"goflare.io/auth/internal/handler"
	"goflare.io/auth/internal/middleware"
	"goflare.io/auth/internal/role"
	"goflare.io/auth/internal/server"
	"goflare.io/auth/internal/user"
)

func InitializeAuthService() (*server.Server, error) {

	wire.Build(
		//config.ProvideApplicationConfig,
		//config.NewLogger,
		//config.ProvidePostgresConn,
		//config.ProvideEnforcer,
		nexus.NewCore,
		nexus.ProvideLogger,
		nexus.ProvidePostgresPool,
		nexus.ProvideConfig,
		nexus.ProvideEnforcer,
		user.NewRepository,
		role.NewRepository,
		firebase.NewService,
		authorization.NewService,
		authentication.NewService,
		middleware.NewAuthenticationMiddleware,
		handler.NewUserHandler,
		server.NewServer,
	)

	return &server.Server{}, nil
}
