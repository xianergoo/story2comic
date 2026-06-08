package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	DBPath        string
	ImageDir      string
	SessionSecret string
	MaxChapters   int
	ImageEnabled  bool
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		Port:          getEnv("PORT", "8080"),
		DBPath:        getEnv("DB_PATH", "data/novelforge.db"),
		ImageDir:      getEnv("IMAGE_DIR", "data/images"),
		SessionSecret: getEnv("SESSION_SECRET", "dev-session-secret-change-me"),
		MaxChapters:   getEnvInt("MAX_CHAPTERS", 5),
		ImageEnabled:  getEnvBool("IMAGE_ENABLED", false),
	}
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v != "" {
		return strings.ToLower(v) == "true" || v == "1"
	}
	return fallback
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
