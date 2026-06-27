package config

import "os"

// Config holds runtime settings loaded from the environment.
type Config struct {
	Addr                string
	BaseURL             string
	DatabasePath        string
	Env                 string
	SessionSecret       string
	GitHubClientID      string
	GitHubClientSecret  string
	BootstrapAdminEmail string
}

// Load reads configuration from REVUES_* environment variables.
func Load() Config {
	return Config{
		Addr:                envOr("REVUES_ADDR", ":8080"),
		BaseURL:             envOr("REVUES_BASE_URL", "http://localhost:8080"),
		DatabasePath:        envOr("REVUES_DATABASE_PATH", "data/revues.db"),
		Env:                 envOr("REVUES_ENV", "development"),
		SessionSecret:       envOr("REVUES_SESSION_SECRET", "change-me-32-random-bytes-minimum"),
		GitHubClientID:      os.Getenv("REVUES_GITHUB_CLIENT_ID"),
		GitHubClientSecret:  os.Getenv("REVUES_GITHUB_CLIENT_SECRET"),
		BootstrapAdminEmail: os.Getenv("REVUES_BOOTSTRAP_ADMIN_EMAIL"),
	}
}

// SecureCookies returns true when cookies must be Secure (production).
func (c Config) SecureCookies() bool {
	return c.Env == "production"
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
