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

// GetSetting returns encrypted value for key.
func (s *Store) GetSetting(ctx context.Context, key string) ([]byte, error) {
	var value []byte
	err := s.db.QueryRowContext(ctx, `
		SELECT value_encrypted FROM settings WHERE key = ?
	`, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSettingNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get setting %q: %w", key, err)
	}

	return value, nil
}

// UpsertSetting stores encrypted value for key.
func (s *Store) UpsertSetting(ctx context.Context, key string, value []byte) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO settings (key, value_encrypted, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value_encrypted = excluded.value_encrypted,
			updated_at = excluded.updated_at
	`, key, value, now)
	if err != nil {
		return fmt.Errorf("upsert setting %q: %w", key, err)
	}

	return nil
}

// DeleteSetting removes a settings key.
func (s *Store) DeleteSetting(ctx context.Context, key string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM settings WHERE key = ?`, key)
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
