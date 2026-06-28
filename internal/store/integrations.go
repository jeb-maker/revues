package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const (
	IntegrationTypeJira   = "jira"
	IntegrationTypeNotion = "notion"
)

// ErrIntegrationNotFound is returned when an integration type is missing.
var ErrIntegrationNotFound = errors.New("integration not found")

// Integration holds encrypted integration configuration.
type Integration struct {
	ID              int64
	Type            string
	Enabled         bool
	ConfigEncrypted []byte
	CreatedAt       string
	UpdatedAt       string
}

// GetIntegrationByType returns the integration row for type.
func (s *Store) GetIntegrationByType(ctx context.Context, integrationType string) (*Integration, error) {
	var row Integration
	var enabled int
	err := s.db.QueryRowContext(ctx, `
		SELECT id, type, enabled, config_encrypted, created_at, updated_at
		FROM integrations
		WHERE type = ?
		ORDER BY id
		LIMIT 1
	`, integrationType).Scan(
		&row.ID,
		&row.Type,
		&enabled,
		&row.ConfigEncrypted,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrIntegrationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get integration %q: %w", integrationType, err)
	}
	row.Enabled = enabled == 1

	return &row, nil
}

// UpsertIntegrationByType stores encrypted config for type.
func (s *Store) UpsertIntegrationByType(ctx context.Context, integrationType string, enabled bool, configEncrypted []byte) error {
	now := time.Now().UTC().Format(time.RFC3339)
	enabledInt := 0
	if enabled {
		enabledInt = 1
	}

	existing, err := s.GetIntegrationByType(ctx, integrationType)
	if errors.Is(err, ErrIntegrationNotFound) {
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO integrations (type, enabled, config_encrypted, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?)
		`, integrationType, enabledInt, configEncrypted, now, now)
		if err != nil {
			return fmt.Errorf("insert integration %q: %w", integrationType, err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("lookup integration %q: %w", integrationType, err)
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE integrations
		SET enabled = ?, config_encrypted = ?, updated_at = ?
		WHERE id = ?
	`, enabledInt, configEncrypted, now, existing.ID)
	if err != nil {
		return fmt.Errorf("update integration %q: %w", integrationType, err)
	}

	return nil
}
