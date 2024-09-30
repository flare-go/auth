package config

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"goflare.io/auth/driver"
	"gopkg.in/yaml.v3"
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
	Paseto   PasetoConfig
}

type PostgresConfig struct {
	URL string `yaml:"url"`
}

type PasetoConfig struct {
	PublicSecretKey     string `yaml:"public_secret_key"`
	PrivateSecretKey    string `yaml:"private_secret_key"`
	TokenExpirationTime time.Duration
}

func ProvideApplicationConfig() (*Config, error) {
	data, err := os.ReadFile("./config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 環境變量覆蓋
	if url := os.Getenv("POSTGRES_URL"); url != "" {
		config.Postgres.URL = url
	}
	if pubKey := os.Getenv("PASETO_PUBLIC_SECRET_KEY"); pubKey != "" {
		config.Paseto.PublicSecretKey = pubKey
	}
	if privKey := os.Getenv("PASETO_PRIVATE_SECRET_KEY"); privKey != "" {
		config.Paseto.PrivateSecretKey = privKey
	}

	config.Paseto.TokenExpirationTime = 120 * time.Minute

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
