package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrIntegrationLinkNotFound is returned when no link exists for a run item.
var ErrIntegrationLinkNotFound = errors.New("integration link not found")

// IntegrationLink associates an external resource with a run item.
type IntegrationLink struct {
	ID            int64
	RunItemID     int64
	IntegrationID int64
	ExternalKey   string
	ExternalURL   string
	CreatedAt     string
}

// IntegrationLinkByRunItemAndType returns the link for a run item and integration type.
func (s *Store) IntegrationLinkByRunItemAndType(ctx context.Context, runItemID int64, integrationType string) (*IntegrationLink, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT il.id, il.run_item_id, il.integration_id, il.external_key, il.external_url, il.created_at
		FROM integration_links il
		INNER JOIN integrations i ON i.id = il.integration_id
		WHERE il.run_item_id = ? AND i.type = ?
		ORDER BY il.id DESC
		LIMIT 1
	`, runItemID, integrationType)

	link, err := scanIntegrationLink(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrIntegrationLinkNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("integration link by run item: %w", err)
	}
	return &link, nil
}

// ListIntegrationLinksByRunItemIDs returns links indexed by run item id for the given type.
func (s *Store) ListIntegrationLinksByRunItemIDs(ctx context.Context, runItemIDs []int64, integrationType string) (map[int64]IntegrationLink, error) {
	if len(runItemIDs) == 0 {
		return map[int64]IntegrationLink{}, nil
	}

	placeholders := make([]string, len(runItemIDs))
	args := make([]any, 0, len(runItemIDs)+1)
	args = append(args, integrationType)
	for i, id := range runItemIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := fmt.Sprintf(`
		SELECT il.id, il.run_item_id, il.integration_id, il.external_key, il.external_url, il.created_at
		FROM integration_links il
		INNER JOIN integrations i ON i.id = il.integration_id
		WHERE i.type = ? AND il.run_item_id IN (%s)
		ORDER BY il.id
	`, strings.Join(placeholders, ","))

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list integration links: %w", err)
	}
	defer rows.Close()

	links := make(map[int64]IntegrationLink, len(runItemIDs))
	for rows.Next() {
		link, err := scanIntegrationLink(rows)
		if err != nil {
			return nil, fmt.Errorf("scan integration link: %w", err)
		}
		links[link.RunItemID] = link
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate integration links: %w", err)
	}

	return links, nil
}

// UpsertIntegrationLink stores or replaces the link for a run item and integration.
func (s *Store) UpsertIntegrationLink(ctx context.Context, runItemID, integrationID int64, externalKey, externalURL string) (*IntegrationLink, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	externalKey = strings.TrimSpace(externalKey)
	externalURL = strings.TrimSpace(externalURL)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var existingID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id FROM integration_links
		WHERE run_item_id = ? AND integration_id = ?
	`, runItemID, integrationID).Scan(&existingID)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		res, insertErr := tx.ExecContext(ctx, `
			INSERT INTO integration_links (run_item_id, integration_id, external_key, external_url, created_at)
			VALUES (?, ?, ?, ?, ?)
		`, runItemID, integrationID, externalKey, externalURL, now)
		if insertErr != nil {
			return nil, fmt.Errorf("insert integration link: %w", insertErr)
		}
		id, insertErr := res.LastInsertId()
		if insertErr != nil {
			return nil, fmt.Errorf("integration link id: %w", insertErr)
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return nil, fmt.Errorf("commit integration link: %w", commitErr)
		}
		return &IntegrationLink{
			ID:            id,
			RunItemID:     runItemID,
			IntegrationID: integrationID,
			ExternalKey:   externalKey,
			ExternalURL:   externalURL,
			CreatedAt:     now,
		}, nil
	case err != nil:
		return nil, fmt.Errorf("lookup integration link: %w", err)
	default:
		if _, updateErr := tx.ExecContext(ctx, `
			UPDATE integration_links
			SET external_key = ?, external_url = ?, created_at = ?
			WHERE id = ?
		`, externalKey, externalURL, now, existingID); updateErr != nil {
			return nil, fmt.Errorf("update integration link: %w", updateErr)
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return nil, fmt.Errorf("commit integration link: %w", commitErr)
		}
		return &IntegrationLink{
			ID:            existingID,
			RunItemID:     runItemID,
			IntegrationID: integrationID,
			ExternalKey:   externalKey,
			ExternalURL:   externalURL,
			CreatedAt:     now,
		}, nil
	}
}

func scanIntegrationLink(row rowScanner) (IntegrationLink, error) {
	var link IntegrationLink
	err := row.Scan(
		&link.ID,
		&link.RunItemID,
		&link.IntegrationID,
		&link.ExternalKey,
		&link.ExternalURL,
		&link.CreatedAt,
	)
	if err != nil {
		return IntegrationLink{}, err
	}
	return link, nil
}
