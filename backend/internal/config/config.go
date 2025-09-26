package config

import (
	"os"
	"strconv"
	"time"
)

// Config captures the runtime configuration for the VidFriends backend service.
type Config struct {
	AppPort          int
	DatabaseURL      string
	MigrationDir     string
	SeedDir          string
	LogLevel         string
	YTDLPPath        string
	YTDLPTimeout     time.Duration
	MetadataCacheTTL time.Duration
	ObjectStore      ObjectStoreConfig
}

// ObjectStoreConfig captures configuration for the S3/MinIO compatible storage
// that persists downloaded video assets.
type ObjectStoreConfig struct {
	Endpoint      string
	Bucket        string
	Region        string
	PublicBaseURL string
}

// Load reads configuration from environment variables, applying sensible defaults
// for local development while allowing overrides through environment variables.
func Load() (Config, error) {
	cfg := Config{
		AppPort:          getInt("VIDFRIENDS_PORT", 8080),
		DatabaseURL:      getString("VIDFRIENDS_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/vidfriends?sslmode=disable"),
		MigrationDir:     getString("VIDFRIENDS_MIGRATIONS", "migrations"),
		SeedDir:          getString("VIDFRIENDS_SEEDS", "seeds"),
		LogLevel:         getString("VIDFRIENDS_LOG_LEVEL", "info"),
		YTDLPPath:        getString("VIDFRIENDS_YTDLP_PATH", "yt-dlp"),
		YTDLPTimeout:     getDuration("VIDFRIENDS_YTDLP_TIMEOUT", 30*time.Second),
		MetadataCacheTTL: getDuration("VIDFRIENDS_METADATA_CACHE_TTL", 15*time.Minute),
		ObjectStore: ObjectStoreConfig{
			Endpoint:      getString("VIDFRIENDS_S3_ENDPOINT", "http://localhost:9000"),
			Bucket:        getString("VIDFRIENDS_S3_BUCKET", "vidfriends"),
			Region:        getString("VIDFRIENDS_S3_REGION", "us-east-1"),
			PublicBaseURL: getString("VIDFRIENDS_S3_PUBLIC_BASE_URL", "http://localhost:9000/vidfriends"),
		},
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

func getDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return d
}
