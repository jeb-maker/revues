package notion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/jeb-maker/revues/internal/crypto"
	"github.com/jeb-maker/revues/internal/store"
)

var ErrEncryptionNotConfigured = errors.New("encryption key not configured")

var databaseIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{32}$|^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type Config struct {
	APIToken          string
	WorkspaceName     string
	DefaultDatabaseID string
}

func (c Config) Configured() bool { return strings.TrimSpace(c.APIToken) != "" }

type configPayload struct {
	APIToken          string `json:"api_token"`
	WorkspaceName     string `json:"workspace_name"`
	DefaultDatabaseID string `json:"default_database_id"`
}

type Service struct {
	Store         *store.Store
	EncryptionKey []byte
}

func (s *Service) Load(ctx context.Context) (Config, bool, error) {
	if len(s.EncryptionKey) != crypto.KeySize {
		return Config{}, false, nil
	}
	row, err := s.Store.GetIntegrationByType(ctx, store.IntegrationTypeNotion)
	if errors.Is(err, store.ErrIntegrationNotFound) {
		return Config{}, false, nil
	}
	if err != nil {
		return Config{}, false, fmt.Errorf("load notion integration: %w", err)
	}
	plaintext, err := crypto.Decrypt(s.EncryptionKey, row.ConfigEncrypted)
	if err != nil {
		return Config{}, false, fmt.Errorf("decrypt notion integration: %w", err)
	}
	var payload configPayload
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return Config{}, false, fmt.Errorf("parse notion integration: %w", err)
	}
	cfg := Config(payload)
	cfg.WorkspaceName = strings.TrimSpace(cfg.WorkspaceName)
	cfg.DefaultDatabaseID = NormalizeDatabaseID(cfg.DefaultDatabaseID)
	return cfg, true, nil
}

func (s *Service) Save(ctx context.Context, cfg Config) error {
	if len(s.EncryptionKey) != crypto.KeySize {
		return ErrEncryptionNotConfigured
	}
	if err := Validate(cfg); err != nil {
		return err
	}
	payload, err := json.Marshal(configPayload{
		APIToken:          cfg.APIToken,
		WorkspaceName:     strings.TrimSpace(cfg.WorkspaceName),
		DefaultDatabaseID: NormalizeDatabaseID(cfg.DefaultDatabaseID),
	})
	if err != nil {
		return fmt.Errorf("marshal notion integration: %w", err)
	}
	encrypted, err := crypto.Encrypt(s.EncryptionKey, payload)
	if err != nil {
		return fmt.Errorf("encrypt notion integration: %w", err)
	}
	if err := s.Store.UpsertIntegrationByType(ctx, store.IntegrationTypeNotion, true, encrypted); err != nil {
		return fmt.Errorf("store notion integration: %w", err)
	}
	return nil
}

func Validate(cfg Config) error {
	if strings.TrimSpace(cfg.APIToken) == "" {
		return errors.New("jeton d'intégration Notion requis")
	}
	if id := NormalizeDatabaseID(cfg.DefaultDatabaseID); id != "" && !databaseIDPattern.MatchString(id) {
		return errors.New("identifiant base Notion invalide")
	}
	return nil
}

func NormalizeDatabaseID(raw string) string {
	return strings.ReplaceAll(strings.TrimSpace(raw), "-", "")
}

func MergeSecret(current, submitted string) string {
	if strings.TrimSpace(submitted) != "" {
		return submitted
	}
	return current
}

func HasSecret(secret string) bool { return secret != "" }
