package config

import (
	"fmt"

	pgadapter "github.com/casbin/casbin-pg-adapter"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"go.uber.org/zap"
)

// ProvideEnforcer provides a new enforcer.
func ProvideEnforcer(appConfig *Config, logger *zap.Logger) (*casbin.Enforcer, error) {

	m, err := model.NewModelFromFile("./casbin.conf")
	if err != nil {
		logger.Error(err.Error())
		return nil, fmt.Errorf("failed to create new model from file: %w", err)
	}

	adapter, err := pgadapter.NewAdapter(appConfig.Postgres.URL)
	if err != nil {
		logger.Error(err.Error())
		return nil, fmt.Errorf("failed to create new adapter: %w", err)
	}

	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		logger.Error(err.Error())
		return nil, fmt.Errorf("failed to create new enforcer: %w", err)
	}

	// 加載策略
	//if err = enforcer.LoadPolicy(); err != nil {
	//	appConfig.Logger.Error(err.Error())
	//	return nil, err
	//}

	return enforcer, nil
}
