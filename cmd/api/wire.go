//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"go.flare.io/auth"
	"go.flare.io/auth/config"
	"go.flare.io/auth/permission"
	"go.flare.io/auth/role"
	"go.flare.io/auth/user"
)

func InitializeAuthService() (auth.Authentication, error) {

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
	)

	return &auth.AuthenticationImpl{}, nil
}
