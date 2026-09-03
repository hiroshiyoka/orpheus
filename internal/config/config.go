package config

import (
	"os"
	"strconv"
)

const defaultCheckIntervalSeconds = 60

type Config struct {
	TelegramBotToken    string
	TelegramChatID      string
	CloudflareAPIToken  string
	DBPath              string
	CheckIntervalSeconds int
}

func Load() Config {
	return Config{
		TelegramBotToken:    os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:      os.Getenv("TELEGRAM_CHAT_ID"),
		CloudflareAPIToken:  os.Getenv("CLOUDFLARE_API_TOKEN"),
		DBPath:              getEnv("DB_PATH", "./data/orpheus.db"),
		CheckIntervalSeconds: getPositiveInt("CHECK_INTERVAL_SECONDS", defaultCheckIntervalSeconds),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getPositiveInt(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
