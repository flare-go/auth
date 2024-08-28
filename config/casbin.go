package config

import (
	pgadapter "github.com/casbin/casbin-pg-adapter"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

func ProvideEnforcer(appConfig *AppConfig) (*casbin.Enforcer, error) {

	m, err := model.NewModelFromFile("./casbin.conf")
	if err != nil {
		appConfig.Logger.Error(err.Error())
		return nil, err
	}

	adapter, err := pgadapter.NewAdapter(appConfig.PostgresURI)
	if err != nil {
		appConfig.Logger.Error(err.Error())
		return nil, err
	}

	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		appConfig.Logger.Error(err.Error())
		return nil, err
	}

	// 加載策略
	//if err = enforcer.LoadPolicy(); err != nil {
	//	appConfig.Logger.Error(err.Error())
	//	return nil, err
	//}

	return enforcer, nil
}
