package config

import (
	"os"
	"time"
)

type Config struct {
	ServerPort      string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	DBPath          string
}

// Load загружается конфигурация (из переменных окружения и констант)
func Load() *Config {
	cfg := &Config{
		ServerPort:      getEnv("SERVER_PORT", "8080"),
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		ShutdownTimeout: 10 * time.Second,
		DBPath:          getEnv("DB_PATH", "links.db"),
	}
	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
