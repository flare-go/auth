package config

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.flare.io/auth/driver"
	"go.flare.io/auth/models"
	"go.uber.org/zap"
	"os"
)

const (
	Local           = "local"
	Cloud           = "cloud"
	ServerStartPort = ":8080"
	ENVConfigType   = "env"
	Environment     = "environment"
)

type AppConfig struct {
	API              string `json:"api" mapstructure:"api"`
	UI               string `json:"ui" mapstructure:"ui"`
	PostgresURI      string `json:"postgres_uri" mapstructure:"db_connection_string"`
	RedisAddr        string `json:"redis_addr"`
	RedisPassword    string `json:"redis_password"`
	PasetoPrivateKey string `json:"paseto_private_key" mapstructure:"paseto_public_key"`
	PasetoPublicKey  string `json:"paseto_public_key" mapstructure:"paseto_private_key"`
	DB               *driver.DB
	//Validate         *validator.Validate
	Logger *zap.Logger
}

func ProvideApplicationConfig() *AppConfig {

	var appConfig AppConfig

	logger, err := zap.NewProduction()
	if err != nil {
		return nil
	}
	appConfig.Logger = logger

	_ = godotenv.Load(fmt.Sprintf("%s.%s", Local, ENVConfigType))
	if cmp.Equal(os.Getenv(Environment), Local) {

		appConfig.PostgresURI = os.Getenv("postgres_uri")
		appConfig.PasetoPrivateKey = os.Getenv("paseto_private_key")
		appConfig.PasetoPublicKey = os.Getenv("paseto_public_key")
		appConfig.API = os.Getenv("api")
		appConfig.UI = os.Getenv("ui")
		appConfig.RedisAddr = os.Getenv("redis_addr")

		return &appConfig
	}
	return &appConfig
}

func ProvidePostgresConn(appConfig *AppConfig) driver.PostgresPool {

	conn, err := driver.ConnectSQL(appConfig.PostgresURI)
	if err != nil {
		return nil
	}

	return conn.Pool
}

func ProvideRedisConn(appConfig *AppConfig) *redis.Client {

	conn, err := driver.ConnectRedis(appConfig.RedisAddr, appConfig.RedisPassword, 0)
	if err != nil {
		return nil
	}

	return conn
}

func ProvidePasetoSecret(appConfig *AppConfig) models.PasetoSecret {

	return models.PasetoSecret{
		PasetoPublicKey:  appConfig.PasetoPublicKey,
		PasetoPrivateKey: appConfig.PasetoPrivateKey,
	}
}

func NewLogger() *zap.Logger {

	logger, _ := zap.NewProduction()
	return logger
}
