package config

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"goflare.io/auth/driver"
	"gopkg.in/yaml.v3"
)

// Local 是用於本地開發環境的常量
const Local = "local"

// Cloud 是用於雲端服務的常量
const Cloud = "cloud"

// ServerStartPort 是服務啟動的端口
const ServerStartPort = ":8080"

// ENVConfigType 是環境變量配置類型
const ENVConfigType = "env"

// Environment 是環境變量
const Environment = "environment"

// Config is the application config.
type Config struct {
	Postgres driver.PostgresConfig
	Paseto   PasetoConfig
	Firebase FirebaseConfig `yaml:"firebase"`
}

// PasetoConfig is the Paseto config.
type PasetoConfig struct {
	PublicSecretKey     string `yaml:"public_secret_key"`
	PrivateSecretKey    string `yaml:"private_secret_key"`
	TokenExpirationTime time.Duration
}

// FirebaseConfig is the Firebase config.
type FirebaseConfig struct {
	ServiceAccountFilePath string `yaml:"service_account_file_path"`
	ProjectID              string
	ClientID               string
	ClientSecret           string
	RedirectURL            string
}

// ProvideApplicationConfig provides the application config.
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
	if sslMode := os.Getenv("POSTGRES_SSLMODE"); sslMode != "" {
		config.Postgres.SSLMode = sslMode
	}
	if sslRootCert := os.Getenv("POSTGRES_SSLROOTCERT"); sslRootCert != "" {
		config.Postgres.SSLRootCert = sslRootCert
	}
	if cluster := os.Getenv("POSTGRES_CLUSTER"); cluster != "" {
		config.Postgres.Cluster = cluster
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

// ProvidePostgresConn provides a new Postgres connection.
func ProvidePostgresConn(appConfig *Config) (driver.PostgresPool, error) {
	conn, err := driver.ConnectSQL(appConfig.Postgres)
	if err != nil {
		return nil, err
	}

	return conn.Pool, nil
}

// NewLogger creates a new logger.
func NewLogger() *zap.Logger {

	logger, _ := zap.NewProduction()
	return logger
}
