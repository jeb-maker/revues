package attachments

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/jeb-maker/revues/internal/store"
)

type Service struct {
	Store *store.Store
	Dir   string
}

func (s *Service) Save(ctx context.Context, runItemID int64, originalName string, data []byte) (*store.Attachment, error) {
	pf, err := ProcessUpload(originalName, data)
	if err != nil {
		return nil, err
	}
	oldPath, hadOld, err := s.Store.PreviousStoragePath(ctx, runItemID)
	if err != nil {
		return nil, fmt.Errorf("lookup previous attachment: %w", err)
	}
	storagePath, err := WriteFile(s.Dir, pf)
	if err != nil {
		return nil, err
	}
	att, err := s.Store.ReplaceAttachment(ctx, runItemID, pf.Filename, pf.MimeType, storagePath, pf.SizeBytes)
	if err != nil {
		_ = RemoveFile(s.Dir, storagePath)
		return nil, fmt.Errorf("replace attachment: %w", err)
	}
	if hadOld {
		if removeErr := RemoveFile(s.Dir, oldPath); removeErr != nil {
			return att, fmt.Errorf("remove old attachment file: %w", removeErr)
		}
	}
	return att, nil
}

func (s *Service) Open(ctx context.Context, attachmentID int64) (*store.Attachment, string, error) {
	att, err := s.Store.AttachmentByID(ctx, attachmentID)
	if err != nil {
		return nil, "", err
	}
	return att, filepath.Join(s.Dir, att.StoragePath), nil
}
