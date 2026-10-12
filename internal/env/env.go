package env

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func GetString(key string, fallback string) string {
	_ = godotenv.Load()

	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
