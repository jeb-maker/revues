package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strconv"
	"strings"

	"github.com/jeb-maker/revues/internal/crypto"
	"github.com/jeb-maker/revues/internal/store"
)

const (
	SettingKeySMTP     = "smtp"
	SettingKeyWebhooks = "webhooks"
)

// ErrEncryptionNotConfigured is returned when REVUES_ENCRYPTION_KEY is missing.
var ErrEncryptionNotConfigured = errors.New("encryption key not configured")

// SMTPConfig holds decrypted SMTP relay settings.
type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	TLS      bool   `json:"tls"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
}

// Enabled reports whether outbound email can be sent.
func (c SMTPConfig) Enabled() bool {
	return strings.TrimSpace(c.Host) != "" && c.Port > 0 && strings.TrimSpace(c.From) != ""
}

// WebhookConfig holds decrypted webhook notification settings.
type WebhookConfig struct {
	URLs            []string
	Secret          string
	ReviewCompleted bool
	ReviewItemNOK   bool
}

// Enabled reports whether webhook delivery can proceed.
func (c WebhookConfig) Enabled() bool {
	return len(c.URLs) > 0 && strings.TrimSpace(c.Secret) != ""
}

// EventEnabled reports whether a given event type should be delivered.
func (c WebhookConfig) EventEnabled(eventType string) bool {
	switch eventType {
	case "review.completed":
		return c.ReviewCompleted
	case "review.item.nok":
		return c.ReviewItemNOK
	case "webhook.test":
		return true
	default:
		return false
	}
}

type webhooksPayload struct {
	URLs            []string `json:"urls"`
	Secret          string   `json:"secret"`
	ReviewCompleted bool     `json:"review_completed"`
	ReviewItemNOK   bool     `json:"review_item_nok"`
}

// SettingsService loads and stores encrypted application settings.
type SettingsService struct {
	Store         *store.Store
	EncryptionKey []byte
}

// LoadSMTP returns stored SMTP config. The second value is false when unset.
func (s *SettingsService) LoadSMTP(ctx context.Context) (SMTPConfig, bool, error) {
	if len(s.EncryptionKey) != crypto.KeySize {
		return SMTPConfig{}, false, nil
	}

	encrypted, err := s.Store.GetSetting(ctx, SettingKeySMTP)
	if errors.Is(err, store.ErrSettingNotFound) {
		return SMTPConfig{}, false, nil
	}
	if err != nil {
		return SMTPConfig{}, false, fmt.Errorf("load smtp setting: %w", err)
	}

	plaintext, err := crypto.Decrypt(s.EncryptionKey, encrypted)
	if err != nil {
		return SMTPConfig{}, false, fmt.Errorf("decrypt smtp setting: %w", err)
	}

	var cfg SMTPConfig
	if err := json.Unmarshal(plaintext, &cfg); err != nil {
		return SMTPConfig{}, false, fmt.Errorf("parse smtp setting: %w", err)
	}

	return cfg, true, nil
}

// SaveSMTP encrypts and stores SMTP config.
func (s *SettingsService) SaveSMTP(ctx context.Context, cfg SMTPConfig) error {
	if len(s.EncryptionKey) != crypto.KeySize {
		return ErrEncryptionNotConfigured
	}
	if err := ValidateSMTP(cfg); err != nil {
		return err
	}

	stored := SMTPConfig{
		Host:     strings.TrimSpace(cfg.Host),
		Port:     cfg.Port,
		TLS:      cfg.TLS,
		Username: strings.TrimSpace(cfg.Username),
		Password: cfg.Password,
		From:     strings.TrimSpace(cfg.From),
	}
	payload, err := json.Marshal(stored)
	if err != nil {
		return fmt.Errorf("marshal smtp setting: %w", err)
	}

	encrypted, err := crypto.Encrypt(s.EncryptionKey, payload)
	if err != nil {
		return fmt.Errorf("encrypt smtp setting: %w", err)
	}

	if err := s.Store.UpsertSetting(ctx, SettingKeySMTP, encrypted); err != nil {
		return fmt.Errorf("store smtp setting: %w", err)
	}

	return nil
}

// ClearSMTP removes stored SMTP configuration.
func (s *SettingsService) ClearSMTP(ctx context.Context) error {
	if err := s.Store.DeleteSetting(ctx, SettingKeySMTP); err != nil {
		if errors.Is(err, store.ErrSettingNotFound) {
			return nil
		}
		return fmt.Errorf("clear smtp setting: %w", err)
	}
	return nil
}

// LoadWebhooks returns stored webhook config. The second value is false when unset.
func (s *SettingsService) LoadWebhooks(ctx context.Context) (WebhookConfig, bool, error) {
	if len(s.EncryptionKey) != crypto.KeySize {
		return WebhookConfig{}, false, nil
	}
	encrypted, err := s.Store.GetSetting(ctx, SettingKeyWebhooks)
	if errors.Is(err, store.ErrSettingNotFound) {
		return WebhookConfig{}, false, nil
	}
	if err != nil {
		return WebhookConfig{}, false, fmt.Errorf("load webhooks setting: %w", err)
	}
	plaintext, err := crypto.Decrypt(s.EncryptionKey, encrypted)
	if err != nil {
		return WebhookConfig{}, false, fmt.Errorf("decrypt webhooks setting: %w", err)
	}
	var payload webhooksPayload
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return WebhookConfig{}, false, fmt.Errorf("parse webhooks setting: %w", err)
	}
	return WebhookConfig{
		URLs: normalizeURLs(payload.URLs), Secret: payload.Secret,
		ReviewCompleted: payload.ReviewCompleted, ReviewItemNOK: payload.ReviewItemNOK,
	}, true, nil
}

// SaveWebhooks encrypts and stores webhook config.
func (s *SettingsService) SaveWebhooks(ctx context.Context, cfg WebhookConfig) error {
	if len(s.EncryptionKey) != crypto.KeySize {
		return ErrEncryptionNotConfigured
	}
	if err := ValidateWebhooks(cfg); err != nil {
		return err
	}
	payload, err := json.Marshal(webhooksPayload{
		URLs: normalizeURLs(cfg.URLs), Secret: cfg.Secret,
		ReviewCompleted: cfg.ReviewCompleted, ReviewItemNOK: cfg.ReviewItemNOK,
	})
	if err != nil {
		return fmt.Errorf("marshal webhooks setting: %w", err)
	}
	encrypted, err := crypto.Encrypt(s.EncryptionKey, payload)
	if err != nil {
		return fmt.Errorf("encrypt webhooks setting: %w", err)
	}
	if err := s.Store.UpsertSetting(ctx, SettingKeyWebhooks, encrypted); err != nil {
		return fmt.Errorf("store webhooks setting: %w", err)
	}
	return nil
}

// ValidateSMTP checks required SMTP fields.
func ValidateSMTP(cfg SMTPConfig) error {
	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		return errors.New("hôte SMTP requis")
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return errors.New("port SMTP invalide")
	}
	from := strings.TrimSpace(cfg.From)
	if from == "" {
		return errors.New("expéditeur requis")
	}
	if _, err := mail.ParseAddress(from); err != nil {
		return errors.New("expéditeur invalide")
	}
	return nil
}

// ParsePort converts a port string to int.
func ParsePort(raw string) (int, error) {
	port, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || port <= 0 || port > 65535 {
		return 0, errors.New("port SMTP invalide")
	}
	return port, nil
}

// MergePassword keeps the existing password when the form leaves it blank.
func MergePassword(current SMTPConfig, submitted string) string {
	if strings.TrimSpace(submitted) != "" {
		return submitted
	}
	return current.Password
}

// ValidateWebhooks checks required webhook fields.
func ValidateWebhooks(cfg WebhookConfig) error {
	urls := normalizeURLs(cfg.URLs)
	if len(urls) == 0 {
		return errors.New("au moins une URL webhook est requise")
	}
	for _, raw := range urls {
		if err := validateWebhookURL(raw); err != nil {
			return err
		}
	}
	if strings.TrimSpace(cfg.Secret) == "" {
		return errors.New("secret HMAC requis")
	}
	if !cfg.ReviewCompleted && !cfg.ReviewItemNOK {
		return errors.New("activez au moins un type d'événement")
	}
	return nil
}

// ParseWebhookURLs parses a multiline/comma-separated URL list.
func ParseWebhookURLs(raw string) []string {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	parts := strings.FieldsFunc(raw, func(r rune) bool { return r == '\n' || r == ',' })
	return normalizeURLs(parts)
}

// MergeWebhookSecret keeps the existing secret when the form leaves it blank.
func MergeWebhookSecret(current WebhookConfig, submitted string) string {
	if strings.TrimSpace(submitted) != "" {
		return submitted
	}
	return current.Secret
}

func normalizeURLs(urls []string) []string {
	var out []string
	seen := map[string]struct{}{}
	for _, raw := range urls {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if _, ok := seen[raw]; ok {
			continue
		}
		seen[raw] = struct{}{}
		out = append(out, raw)
	}
	return out
}

func validateWebhookURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("URL webhook invalide : %s", raw)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return fmt.Errorf("URL webhook invalide : %s", raw)
	}
	if u.Host == "" {
		return fmt.Errorf("URL webhook invalide : %s", raw)
	}
	return nil
}
