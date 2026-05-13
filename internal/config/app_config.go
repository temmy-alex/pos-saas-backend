package config

import (
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	AppName  string
	AppEnv   string
	AppPort  string
	AppDebug string
}

func LoadAppConfig() AppConfig {
	_ = godotenv.Load()

	return AppConfig{
		AppName:  getEnv("APP_NAME", "POS SaaS Backend"),
		AppEnv:   getEnv("APP_ENV", "local"),
		AppPort:  getEnv("APP_PORT", "8080"),
		AppDebug: getEnv("APP_DEBUG", "true"),
	}
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	return value
}
