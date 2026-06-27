package config

import "os"

// Config holds runtime settings loaded from the environment.
type Config struct {
	Addr         string
	BaseURL      string
	DatabasePath string
	Env          string
}

// Load reads configuration from REVUES_* environment variables.
func Load() Config {
	return Config{
		Addr:         envOr("REVUES_ADDR", ":8080"),
		BaseURL:      envOr("REVUES_BASE_URL", "http://localhost:8080"),
		DatabasePath: envOr("REVUES_DATABASE_PATH", "data/revues.db"),
		Env:          envOr("REVUES_ENV", "development"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
