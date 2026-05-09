package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Config keeps runtime options for the API service.
type Config struct {
	AppName            string
	Env                string
	Port               string
	GinMode            string
	MySQLDSN           string
	MiniProgramAppID   string
	MiniProgramSecret  string
	SessionTokenSecret string
}

func Load() Config {
	loadDotEnv()

	cfg := Config{
		AppName:            envOrDefault("APP_NAME", "jiezhang-backend"),
		Env:                envOrDefault("ENV", "dev"),
		Port:               envOrDefault("PORT", "10240"),
		GinMode:            envOrDefault("GIN_MODE", gin.DebugMode),
		MySQLDSN:           strings.TrimSpace(envOrDefault("MYSQL_DSN", "")),
		MiniProgramAppID:   strings.TrimSpace(envOrDefault("MINIPROGRAM_APPID", "")),
		MiniProgramSecret:  strings.TrimSpace(envOrDefault("MINIPROGRAM_SECRET", "")),
		SessionTokenSecret: strings.TrimSpace(envOrDefault("SESSION_TOKEN_SECRET", "")),
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

func loadDotEnv() {
	wd, err := os.Getwd()
	if err != nil {
		return
	}

	path := filepath.Join(wd, ".env")
	if _, statErr := os.Stat(path); statErr != nil {
		return
	}

	if err := godotenv.Load(path); err != nil {
		log.Printf("failed to load .env: %v", err)
	}
}
