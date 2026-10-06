package confluence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"strings"

	"github.com/jeb-maker/revues/internal/crypto"
	"github.com/jeb-maker/revues/internal/safehttp"
	"github.com/jeb-maker/revues/internal/store"
)

// ErrEncryptionNotConfigured is returned when REVUES_ENCRYPTION_KEY is missing.
var ErrEncryptionNotConfigured = crypto.ErrEncryptionNotConfigured

// Config holds decrypted Confluence Cloud connection settings.
type Config struct {
	BaseURL      string
	Email        string
	APIToken     string
	SpaceKey     string
	ParentPageID string
}

// Configured reports whether Confluence credentials and space are complete.
func (c Config) Configured() bool {
	return strings.TrimSpace(c.BaseURL) != "" &&
		strings.TrimSpace(c.Email) != "" &&
		c.APIToken != "" &&
		strings.TrimSpace(c.SpaceKey) != ""
}

type configPayload struct {
	BaseURL      string `json:"base_url"`
	Email        string `json:"email"`
	APIToken     string `json:"api_token"`
	SpaceKey     string `json:"space_key"`
	ParentPageID string `json:"parent_page_id"`
}

// Service loads and stores encrypted Confluence integration settings.
type Service struct {
	Store         ConfigStore
	EncryptionKey []byte
}

// Load returns stored Confluence config. The second value is false when unset.
func (s *Service) Load(ctx context.Context) (Config, bool, error) {
	if len(s.EncryptionKey) != crypto.KeySize {
		return Config{}, false, nil
	}

	row, err := s.Store.GetIntegrationByType(ctx, store.IntegrationTypeConfluence)
	if errors.Is(err, ErrIntegrationNotFound) {
		return Config{}, false, nil
	}
	if err != nil {
		return Config{}, false, fmt.Errorf("load confluence integration: %w", err)
	}

	plaintext, err := crypto.Decrypt(s.EncryptionKey, row.ConfigEncrypted)
	if err != nil {
		return Config{}, false, fmt.Errorf("decrypt confluence integration: %w", err)
	}

	var payload configPayload
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return Config{}, false, fmt.Errorf("parse confluence integration: %w", err)
	}

	cfg := Config(payload)
	cfg.SpaceKey = strings.ToUpper(strings.TrimSpace(cfg.SpaceKey))
	cfg.ParentPageID = strings.TrimSpace(cfg.ParentPageID)
	return cfg, true, nil
}

// Clear removes stored Confluence configuration for the active organization.
func (s *Service) Clear(ctx context.Context) error {
	if err := s.Store.DeleteIntegrationByType(ctx, store.IntegrationTypeConfluence); err != nil {
		if errors.Is(err, ErrIntegrationNotFound) {
			return nil
		}
		return fmt.Errorf("clear confluence integration: %w", err)
	}
	return nil
}

// Save encrypts and stores Confluence config.
func (s *Service) Save(ctx context.Context, cfg Config) error {
	if len(s.EncryptionKey) != crypto.KeySize {
		return ErrEncryptionNotConfigured
	}
	if err := Validate(cfg); err != nil {
		return err
	}

	payload, err := json.Marshal(configPayload{
		BaseURL:      NormalizeBaseURL(cfg.BaseURL),
		Email:        strings.TrimSpace(cfg.Email),
		APIToken:     cfg.APIToken,
		SpaceKey:     strings.ToUpper(strings.TrimSpace(cfg.SpaceKey)),
		ParentPageID: strings.TrimSpace(cfg.ParentPageID),
	})
	if err != nil {
		return fmt.Errorf("marshal confluence integration: %w", err)
	}

	encrypted, err := crypto.Encrypt(s.EncryptionKey, payload)
	if err != nil {
		return fmt.Errorf("encrypt confluence integration: %w", err)
	}

	if err := s.Store.UpsertIntegrationByType(ctx, store.IntegrationTypeConfluence, true, encrypted); err != nil {
		return fmt.Errorf("store confluence integration: %w", err)
	}

	return nil
}

// Validate checks required Confluence Cloud fields.
func Validate(cfg Config) error {
	if err := ValidateBaseURL(cfg.BaseURL); err != nil {
		return err
	}

	email := strings.TrimSpace(cfg.Email)
	if email == "" {
		return errors.New("email Confluence requis")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New("email Confluence invalide")
	}
	if strings.TrimSpace(cfg.APIToken) == "" {
		return errors.New("jeton API Confluence requis")
	}
	if strings.TrimSpace(cfg.SpaceKey) == "" {
		return errors.New("clé d'espace Confluence requise")
	}
	return nil
}

// ValidateBaseURL checks that the Confluence URL uses HTTPS (localhost HTTP allowed)
// and rejects literal private/metadata IP targets (SSRF).
func ValidateBaseURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return errors.New("URL Confluence requise")
	}

	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return errors.New("URL Confluence invalide")
	}

	switch u.Scheme {
	case "https":
		// ok
	case "http":
		host := u.Hostname()
		if host == "localhost" || host == "127.0.0.1" {
			break
		}
		return errors.New("URL Confluence doit utiliser HTTPS")
	default:
		return errors.New("URL Confluence doit utiliser HTTPS")
	}

	host := u.Hostname()
	if safehttp.IsLocalhostHost(host) {
		return nil
	}
	if ip := net.ParseIP(host); ip != nil && safehttp.BlockedIP(ip) {
		return errors.New("URL Confluence pointe vers une adresse non autorisée")
	}
	return nil
}

// NormalizeBaseURL trims trailing slashes from the Confluence base URL.
func NormalizeBaseURL(raw string) string {
	return strings.TrimRight(strings.TrimSpace(raw), "/")
}

// MergeSecret keeps the existing secret when the form leaves it blank.
func MergeSecret(current, submitted string) string {
	return crypto.MergeSecret(current, submitted)
}

// HasSecret reports whether a credential is stored.
func HasSecret(secret string) bool {
	return crypto.HasSecret(secret)
}
