package config

import (
	"os"
	"strconv"
)

// Config captures the runtime configuration for the VidFriends backend service.
type Config struct {
	AppPort      int
	DatabaseURL  string
	MigrationDir string
	LogLevel     string
}

// Load reads configuration from environment variables, applying sensible defaults
// for local development while allowing overrides through environment variables.
func Load() (Config, error) {
	cfg := Config{
		AppPort:      getInt("VIDFRIENDS_PORT", 8080),
		DatabaseURL:  getString("VIDFRIENDS_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/vidfriends?sslmode=disable"),
		MigrationDir: getString("VIDFRIENDS_MIGRATIONS", "migrations"),
		LogLevel:     getString("VIDFRIENDS_LOG_LEVEL", "info"),
	}

	return cfg, nil
}

func getString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return i
}
