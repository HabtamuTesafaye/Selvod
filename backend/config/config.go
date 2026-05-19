package config

import (
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	Port         string
	DBPath       string
	StoragePath  string
	StreamSecret string
	BaseURL      string
	WebhookURL   string
	MaxWorkers   int
	APIKey       string
	PlaybackKey  string // Dedicated key for signing
}

func Load() *Config {
	cfg := &Config{
		Port:         getEnv("SV_PORT", "8080"),
		DBPath:       getEnv("SV_DB_PATH", "data/selvod.db"),
		StoragePath:  getEnv("SV_STORAGE_PATH", "data"),
		StreamSecret: getEnv("SV_STREAM_SECRET", ""),
		BaseURL:      getEnv("SV_BASE_URL", "http://localhost:18080"),
		WebhookURL:   getEnv("SV_WEBHOOK_URL", ""),
		MaxWorkers:   getEnvInt("SV_MAX_WORKERS", 2),
		APIKey:       getEnv("SV_API_KEY", ""),
		PlaybackKey:  getEnv("SV_PLAYBACK_KEY", ""), // INDEPENDENT KEY
	}

	// Hardening: Require all security credentials
	if cfg.StreamSecret == "" || cfg.APIKey == "" || cfg.PlaybackKey == "" {
		slog.Error("CRITICAL: SV_STREAM_SECRET, SV_API_KEY, and SV_PLAYBACK_KEY must be set.")
		os.Exit(1)
	}
	if cfg.MaxWorkers < 1 {
		slog.Warn("SV_MAX_WORKERS must be at least 1; using 1", "configured", cfg.MaxWorkers)
		cfg.MaxWorkers = 1
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.Atoi(value)
		if err == nil {
			return i
		}
	}
	return fallback
}
