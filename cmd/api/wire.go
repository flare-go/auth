//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"goflare.io/auth/authentication"
	"goflare.io/auth/authorization"
	"goflare.io/auth/config"
	"goflare.io/auth/firebase"
	"goflare.io/auth/handler"
	"goflare.io/auth/middleware"
	"goflare.io/auth/role"
	"goflare.io/auth/server"
	"goflare.io/auth/user"
)

func InitializeAuthService() (*server.Server, error) {

	wire.Build(
		config.ProvideApplicationConfig,
		config.NewLogger,
		config.ProvidePostgresConn,
		config.ProvideEnforcer,
		firebase.NewFirebaseClient,
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
