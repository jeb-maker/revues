package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/jeb-maker/revues/internal/crypto"
	"github.com/jeb-maker/revues/internal/store"
)

const SettingKeyWebhooks = "webhooks"

type WebhookConfig struct {
	URLs            []string
	Secret          string
	ReviewCompleted bool
	ReviewItemNOK   bool
}

func (c WebhookConfig) Enabled() bool {
	return len(c.URLs) > 0 && strings.TrimSpace(c.Secret) != ""
}

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

func ParseWebhookURLs(raw string) []string {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	parts := strings.FieldsFunc(raw, func(r rune) bool { return r == '\n' || r == ',' })
	return normalizeURLs(parts)
}

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
