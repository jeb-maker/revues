package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrSettingNotFound is returned when a settings key is missing.
var ErrSettingNotFound = errors.New("setting not found")

// GetSetting returns encrypted value for key in the active organization.
func (s *Store) GetSetting(ctx context.Context, key string) ([]byte, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var value []byte
	err = s.db.QueryRowContext(ctx, `
		SELECT value_encrypted FROM settings WHERE organization_id = ? AND key = ?
	`, orgID, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSettingNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get setting %q: %w", key, err)
	}

	return value, nil
}

// UpsertSetting stores encrypted value for key in the active organization.
func (s *Store) UpsertSetting(ctx context.Context, key string, value []byte) error {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO settings (organization_id, key, value_encrypted, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(organization_id, key) DO UPDATE SET
			value_encrypted = excluded.value_encrypted,
			updated_at = excluded.updated_at
	`, orgID, key, value, now)
	if err != nil {
		return fmt.Errorf("upsert setting %q: %w", key, err)
	}

	return nil
}

// DeleteSetting removes a settings key in the active organization.
func (s *Store) DeleteSetting(ctx context.Context, key string) error {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	res, err := s.db.ExecContext(ctx, `
		DELETE FROM settings WHERE organization_id = ? AND key = ?
	`, orgID, key)
	if err != nil {
		return fmt.Errorf("delete setting %q: %w", key, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete setting rows %q: %w", key, err)
	}
	if n == 0 {
		return ErrSettingNotFound
	}

	return nil
}
