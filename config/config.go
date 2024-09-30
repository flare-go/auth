package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"goflare.io/auth/driver"
)

const (
	Local           = "local"
	Cloud           = "cloud"
	ServerStartPort = ":8080"
	ENVConfigType   = "env"
	Environment     = "environment"
)

type Config struct {
	Postgres PostgresConfig
}

type PostgresConfig struct {
	URL string `mapstructure:"url"`
}

func ProvideApplicationConfig() (*Config, error) {

	viper.SetConfigFile("./config.yaml")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func ProvidePostgresConn(appConfig *Config) (driver.PostgresPool, error) {

	conn, err := driver.ConnectSQL(appConfig.Postgres.URL)
	if err != nil {
		return nil, err
	}

	return conn.Pool, nil
}

func NewLogger() *zap.Logger {

	logger, _ := zap.NewProduction()
	return logger
}
