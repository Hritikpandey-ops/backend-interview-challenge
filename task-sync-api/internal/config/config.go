package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port          string
	DatabasePath  string
	SyncBatchSize int
	MaxRetries    int
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "3000"),
		DatabasePath:  getEnv("DATABASE_PATH", "./data/tasks.db"),
		SyncBatchSize: getEnvAsInt("SYNC_BATCH_SIZE", 50),
		MaxRetries:    getEnvAsInt("MAX_RETRIES", 3),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
