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

// Config is the application config.
type Config struct {
	Postgres PostgresConfig
	Paseto   PasetoConfig
	Firebase FirebaseConfig
}

// PostgresConfig is the Postgres config.
type PostgresConfig struct {
	URL string `yaml:"url"`
}

// PasetoConfig is the Paseto config.
type PasetoConfig struct {
	PublicSecretKey     string `yaml:"public_secret_key"`
	PrivateSecretKey    string `yaml:"private_secret_key"`
	TokenExpirationTime time.Duration
}

// FirebaseConfig is the Firebase config.
type FirebaseConfig struct {
	ServiceAccountFilePath string
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
	if pubKey := os.Getenv("PASETO_PUBLIC_SECRET_KEY"); pubKey != "" {
		config.Paseto.PublicSecretKey = pubKey
	}
	if privKey := os.Getenv("PASETO_PRIVATE_SECRET_KEY"); privKey != "" {
		config.Paseto.PrivateSecretKey = privKey
	}

	config.Paseto.TokenExpirationTime = 120 * time.Minute

	// 讀取 Firebase 配置
	config.Firebase.ServiceAccountFilePath = os.Getenv("FIREBASE_SERVICE_ACCOUNT_FILE")
	config.Firebase.ProjectID = os.Getenv("FIREBASE_PROJECT_ID")
	config.Firebase.ClientID = os.Getenv("FIREBASE_CLIENT_ID")
	config.Firebase.ClientSecret = os.Getenv("FIREBASE_CLIENT_SECRET")
	config.Firebase.RedirectURL = os.Getenv("FIREBASE_REDIRECT_URL")

	return &config, nil
}

// ProvidePostgresConn provides a new Postgres connection.
func ProvidePostgresConn(appConfig *Config) (driver.PostgresPool, error) {
	conn, err := driver.ConnectSQL(appConfig.Postgres.URL)
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
