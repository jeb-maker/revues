package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strconv"
	"strings"

	"github.com/jeb-maker/revues/internal/crypto"
	"github.com/jeb-maker/revues/internal/store"
)

const SettingKeySMTP = "smtp"

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
