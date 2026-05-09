package config

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// Config keeps runtime options for the API service.
type Config struct {
	AppName string
	Port    string
	GinMode string
}

func Load() Config {
	cfg := Config{
		AppName: envOrDefault("APP_NAME", "jiezhang-backend"),
		Port:    envOrDefault("PORT", "10240"),
		GinMode: envOrDefault("GIN_MODE", gin.DebugMode),
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg
}

func (c Config) ListenAddr() string {
	if strings.HasPrefix(c.Port, ":") {
		return c.Port
	}
	return ":" + c.Port
}

func envOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
