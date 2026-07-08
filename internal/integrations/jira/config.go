package jira

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strings"

	"github.com/jeb-maker/revues/internal/crypto"
	"github.com/jeb-maker/revues/internal/store"
)

const (
	InstanceCloud    = "cloud"
	InstanceServer   = "server"
	DefaultIssueType = "Task"
)

// ErrEncryptionNotConfigured is returned when REVUES_ENCRYPTION_KEY is missing.
var ErrEncryptionNotConfigured = errors.New("encryption key not configured")

// Config holds decrypted Jira connection settings.
type Config struct {
	InstanceType string
	BaseURL      string
	Email        string
	APIToken     string
	PAT          string
	ProjectKey   string
	IssueType    string
}

// Configured reports whether Jira credentials are complete for the instance type.
func (c Config) Configured() bool {
	if strings.TrimSpace(c.BaseURL) == "" {
		return false
	}
	switch c.InstanceType {
	case InstanceCloud:
		return strings.TrimSpace(c.Email) != "" && c.APIToken != ""
	case InstanceServer:
		return c.PAT != ""
	default:
		return false
	}
}

type configPayload struct {
	InstanceType string `json:"instance_type"`
	BaseURL      string `json:"base_url"`
	Email        string `json:"email"`
	APIToken     string `json:"api_token"`
	PAT          string `json:"pat"`
	ProjectKey   string `json:"project_key"`
	IssueType    string `json:"issue_type"`
}

// Service loads and stores encrypted Jira integration settings.
type Service struct {
	Store         ConfigStore
	EncryptionKey []byte
}

// Load returns stored Jira config. The second value is false when unset.
func (s *Service) Load(ctx context.Context) (Config, bool, error) {
	if len(s.EncryptionKey) != crypto.KeySize {
		return Config{}, false, nil
	}

	row, err := s.Store.GetIntegrationByType(ctx, store.IntegrationTypeJira)
	if errors.Is(err, ErrIntegrationNotFound) {
		return Config{}, false, nil
	}
	if err != nil {
		return Config{}, false, fmt.Errorf("load jira integration: %w", err)
	}

	plaintext, err := crypto.Decrypt(s.EncryptionKey, row.ConfigEncrypted)
	if err != nil {
		return Config{}, false, fmt.Errorf("decrypt jira integration: %w", err)
	}

	var payload configPayload
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return Config{}, false, fmt.Errorf("parse jira integration: %w", err)
	}

	cfg := Config(payload)
	cfg.ProjectKey = strings.ToUpper(strings.TrimSpace(cfg.ProjectKey))
	cfg.IssueType = strings.TrimSpace(cfg.IssueType)
	if cfg.IssueType == "" {
		cfg.IssueType = DefaultIssueType
	}
	return cfg, true, nil
}

// Save encrypts and stores Jira config.
func (s *Service) Save(ctx context.Context, cfg Config) error {
	if len(s.EncryptionKey) != crypto.KeySize {
		return ErrEncryptionNotConfigured
	}
	if err := Validate(cfg); err != nil {
		return err
	}

	issueType := strings.TrimSpace(cfg.IssueType)
	if issueType == "" {
		issueType = DefaultIssueType
	}

	payload, err := json.Marshal(configPayload{
		InstanceType: cfg.InstanceType,
		BaseURL:      NormalizeBaseURL(cfg.BaseURL),
		Email:        strings.TrimSpace(cfg.Email),
		APIToken:     cfg.APIToken,
		PAT:          cfg.PAT,
		ProjectKey:   strings.ToUpper(strings.TrimSpace(cfg.ProjectKey)),
		IssueType:    issueType,
	})
	if err != nil {
		return fmt.Errorf("marshal jira integration: %w", err)
	}

	encrypted, err := crypto.Encrypt(s.EncryptionKey, payload)
	if err != nil {
		return fmt.Errorf("encrypt jira integration: %w", err)
	}

	if err := s.Store.UpsertIntegrationByType(ctx, store.IntegrationTypeJira, true, encrypted); err != nil {
		return fmt.Errorf("store jira integration: %w", err)
	}

	return nil
}

// Validate checks required Jira fields for the selected instance type.
func Validate(cfg Config) error {
	instanceType := strings.TrimSpace(cfg.InstanceType)
	switch instanceType {
	case InstanceCloud, InstanceServer:
	default:
		return errors.New("type d'instance Jira invalide")
	}

	if err := ValidateBaseURL(cfg.BaseURL); err != nil {
		return err
	}

	switch instanceType {
	case InstanceCloud:
		email := strings.TrimSpace(cfg.Email)
		if email == "" {
			return errors.New("email Jira requis")
		}
		if _, err := mail.ParseAddress(email); err != nil {
			return errors.New("email Jira invalide")
		}
		if strings.TrimSpace(cfg.APIToken) == "" {
			return errors.New("jeton API Jira requis")
		}
	case InstanceServer:
		if strings.TrimSpace(cfg.PAT) == "" {
			return errors.New("PAT Jira requis")
		}
	}

	return nil
}

// ValidateBaseURL checks that the Jira URL uses HTTPS (localhost HTTP allowed).
func ValidateBaseURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return errors.New("URL Jira requise")
	}

	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return errors.New("URL Jira invalide")
	}

	switch u.Scheme {
	case "https":
		return nil
	case "http":
		host := u.Hostname()
		if host == "localhost" || host == "127.0.0.1" {
			return nil
		}
		return errors.New("URL Jira doit utiliser HTTPS")
	default:
		return errors.New("URL Jira doit utiliser HTTPS")
	}
}

// NormalizeBaseURL trims trailing slashes from the Jira base URL.
func NormalizeBaseURL(raw string) string {
	return strings.TrimRight(strings.TrimSpace(raw), "/")
}

// MergeSecret keeps the existing secret when the form leaves it blank.
func MergeSecret(current, submitted string) string {
	if strings.TrimSpace(submitted) != "" {
		return submitted
	}
	return current
}

// HasSecret reports whether a credential is stored.
func HasSecret(secret string) bool {
	return secret != ""
}
