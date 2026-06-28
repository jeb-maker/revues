package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrAttachmentNotFound = errors.New("attachment not found")

type Attachment struct {
	ID          int64
	RunItemID   int64
	Filename    string
	MimeType    string
	SizeBytes   int64
	StoragePath string
	CreatedAt   string
}

func (s *Store) AttachmentByRunItemID(ctx context.Context, runItemID int64) (*Attachment, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, run_item_id, filename, mime_type, size_bytes, storage_path, created_at
		FROM attachments WHERE run_item_id = ?
	`, runItemID)
	att, err := scanAttachment(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAttachmentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("attachment by run item: %w", err)
	}
	return &att, nil
}

func (s *Store) AttachmentByID(ctx context.Context, id int64) (*Attachment, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, run_item_id, filename, mime_type, size_bytes, storage_path, created_at
		FROM attachments WHERE id = ?
	`, id)
	att, err := scanAttachment(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAttachmentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("attachment by id: %w", err)
	}
	return &att, nil
}

func (s *Store) RunIDForAttachment(ctx context.Context, attachmentID int64) (int64, error) {
	var runID int64
	err := s.db.QueryRowContext(ctx, `
		SELECT ri.run_id FROM attachments a
		INNER JOIN run_items ri ON ri.id = a.run_item_id
		WHERE a.id = ?
	`, attachmentID).Scan(&runID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrAttachmentNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("run id for attachment: %w", err)
	}
	return runID, nil
}

func (s *Store) PreviousStoragePath(ctx context.Context, runItemID int64) (string, bool, error) {
	row := s.db.QueryRowContext(ctx, `SELECT storage_path FROM attachments WHERE run_item_id = ?`, runItemID)
	var path string
	if err := row.Scan(&path); errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	} else if err != nil {
		return "", false, fmt.Errorf("previous storage path: %w", err)
	}
	return path, true, nil
}

func (s *Store) ReplaceAttachment(ctx context.Context, runItemID int64, filename, mimeType, storagePath string, sizeBytes int64) (*Attachment, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err = tx.ExecContext(ctx, `DELETE FROM attachments WHERE run_item_id = ?`, runItemID); err != nil {
		return nil, fmt.Errorf("delete existing attachment: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := tx.ExecContext(ctx, `
		INSERT INTO attachments (run_item_id, filename, mime_type, size_bytes, storage_path, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, runItemID, filename, mimeType, sizeBytes, storagePath, now)
	if err != nil {
		return nil, fmt.Errorf("insert attachment: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("attachment last insert id: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit attachment: %w", err)
	}
	return &Attachment{
		ID: id, RunItemID: runItemID, Filename: filename, MimeType: mimeType,
		SizeBytes: sizeBytes, StoragePath: storagePath, CreatedAt: now,
	}, nil
}

func scanAttachment(row rowScanner) (Attachment, error) {
	var att Attachment
	err := row.Scan(&att.ID, &att.RunItemID, &att.Filename, &att.MimeType, &att.SizeBytes, &att.StoragePath, &att.CreatedAt)
	return att, err
}
