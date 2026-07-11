package config

import (
	"encoding/base64"
	"os"
	"strconv"

	"github.com/jeb-maker/revues/internal/crypto"
)

// Config holds runtime settings loaded from the environment.
type Config struct {
	Addr                string
	BaseURL             string
	DatabasePath        string
	DBMaxOpenConns      int
	AttachmentsDir      string
	Env                 string
	SessionSecret       string
	EncryptionKey       string
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
		DBMaxOpenConns:      envIntOr("REVUES_DB_MAX_OPEN_CONNS", 10),
		AttachmentsDir:      envOr("REVUES_ATTACHMENTS_DIR", "data/attachments"),
		Env:                 envOr("REVUES_ENV", "development"),
		SessionSecret:       envOr("REVUES_SESSION_SECRET", "change-me-32-random-bytes-minimum"),
		EncryptionKey:       os.Getenv("REVUES_ENCRYPTION_KEY"),
		GitHubClientID:      os.Getenv("REVUES_GITHUB_CLIENT_ID"),
		GitHubClientSecret:  os.Getenv("REVUES_GITHUB_CLIENT_SECRET"),
		BootstrapAdminEmail: os.Getenv("REVUES_BOOTSTRAP_ADMIN_EMAIL"),
	}
}

// SecureCookies returns true when cookies must be Secure (production).
func (c Config) SecureCookies() bool {
	return c.Env == "production"
}

// EncryptionKeyBytes decodes REVUES_ENCRYPTION_KEY when configured.
func (c Config) EncryptionKeyBytes() ([]byte, error) {
	if c.EncryptionKey == "" {
		return nil, nil
	}
	return crypto.DecodeKey(c.EncryptionKey)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOr(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}

// TestEncryptionKey returns a valid base64 key for tests.
func TestEncryptionKey() string {
	key := make([]byte, crypto.KeySize)
	for i := range key {
		key[i] = byte(i + 1)
	}
	return base64.StdEncoding.EncodeToString(key)
}
