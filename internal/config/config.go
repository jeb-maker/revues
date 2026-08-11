package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jeb-maker/revues/internal/crypto"
)

const (
	defaultSessionSecret = "change-me-32-random-bytes-minimum"
	minSessionSecretLen  = 32
)

// Config holds runtime settings loaded from the environment.
type Config struct {
	Addr                  string
	BaseURL               string
	DatabasePath          string
	DBMaxOpenConns        int
	AttachmentsDir        string
	Env                   string
	SessionSecret         string
	EncryptionKey         string
	GitHubClientID        string
	GitHubClientSecret    string
	BootstrapAdminEmail   string
	LoginRequireWhitelist bool
	DevAuth               bool
	DevAuthEmail          string
}

// Load reads configuration from REVUES_* environment variables.
func Load() Config {
	return Config{
		Addr:                  envOr("REVUES_ADDR", ":8080"),
		BaseURL:               envOr("REVUES_BASE_URL", "http://localhost:8080"),
		DatabasePath:          envOr("REVUES_DATABASE_PATH", "data/revues.db"),
		DBMaxOpenConns:        envIntOr("REVUES_DB_MAX_OPEN_CONNS", 10),
		AttachmentsDir:        envOr("REVUES_ATTACHMENTS_DIR", "data/attachments"),
		Env:                   envOr("REVUES_ENV", "development"),
		SessionSecret:         envOr("REVUES_SESSION_SECRET", defaultSessionSecret),
		EncryptionKey:         os.Getenv("REVUES_ENCRYPTION_KEY"),
		GitHubClientID:        os.Getenv("REVUES_GITHUB_CLIENT_ID"),
		GitHubClientSecret:    os.Getenv("REVUES_GITHUB_CLIENT_SECRET"),
		BootstrapAdminEmail:   os.Getenv("REVUES_BOOTSTRAP_ADMIN_EMAIL"),
		LoginRequireWhitelist: envBool("REVUES_LOGIN_REQUIRE_WHITELIST"),
		DevAuth:               envBool("REVUES_DEV_AUTH"),
		DevAuthEmail:          envOr("REVUES_DEV_AUTH_EMAIL", "admin@example.com"),
	}
}

// Validate rejects unsafe production configuration (fail closed).
func (c Config) Validate() error {
	if c.Env != "production" {
		return nil
	}
	if err := c.validateSessionSecret(); err != nil {
		return err
	}
	if c.EncryptionKey != "" {
		if _, err := c.EncryptionKeyBytes(); err != nil {
			return fmt.Errorf("REVUES_ENCRYPTION_KEY: %w", err)
		}
	}
	return nil
}

func (c Config) validateSessionSecret() error {
	secret := c.SessionSecret
	if secret == "" || secret == defaultSessionSecret {
		return fmt.Errorf("REVUES_SESSION_SECRET must be set to a unique value in production (not the documented default)")
	}
	if len(secret) < minSessionSecretLen {
		return fmt.Errorf("REVUES_SESSION_SECRET must be at least %d characters in production", minSessionSecretLen)
	}
	return nil
}

// SecureCookies returns true when cookies must be Secure (HTTPS).
func (c Config) SecureCookies() bool {
	if c.Env == "production" {
		return true
	}
	return strings.HasPrefix(strings.ToLower(c.BaseURL), "https://")
}

// DevAuthEnabled is true only for local demo login bypass (never in production).
func (c Config) DevAuthEnabled() bool {
	return c.DevAuth && c.Env != "production"
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

func envBool(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// TestEncryptionKey returns a valid base64 key for tests.
func TestEncryptionKey() string {
	key := make([]byte, crypto.KeySize)
	for i := range key {
		key[i] = byte(i + 1)
	}
	return base64.StdEncoding.EncodeToString(key)
}
