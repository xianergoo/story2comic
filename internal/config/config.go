package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	DBPath        string
	ImageDir      string
	SessionSecret string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		Port:          getEnv("PORT", "8080"),
		DBPath:        getEnv("DB_PATH", "data/novelforge.db"),
		ImageDir:      getEnv("IMAGE_DIR", "data/images"),
		SessionSecret: getEnv("SESSION_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
