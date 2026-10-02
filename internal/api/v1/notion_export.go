package apiv1

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/integrations/notion"
	"github.com/jeb-maker/revues/internal/store"
)

// PostRunNotionExport serves POST /api/v1/runs/{runId}/notion-export.
//
//nolint:staticcheck // OpenAPI operation param name (runId)
func (s *Server) PostRunNotionExport(w http.ResponseWriter, r *http.Request, runId RunId) {
	run, user, access, ok := s.ensureRunAccess(w, r, runId)
	if !ok {
		return
	}
	if !runs.CanCompleteAccess(user, access) {
		writeAPIError(w, http.StatusNotFound, "not_found", "Revue introuvable.")
		return
	}
	if run.Status != store.RunStatusDone {
		writeAPIError(w, http.StatusBadRequest, "validation_failed",
			"Seules les revues terminées peuvent être exportées vers Notion.")
		return
	}

	url, err := s.notionExportService().ExportRun(r.Context(), run.ID)
	if err != nil {
		writeNotionExportErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, NotionExportResponse{NotionUrl: url})
}

func (s *Server) notionExportService() *notion.ExportService {
	return &notion.ExportService{
		Store:         s.Store,
		EncryptionKey: s.notionEncryptionKey(),
		BaseURL:       s.Config.BaseURL,
		Client:        s.notionClient(),
	}
}

func (s *Server) notionEncryptionKey() []byte {
	if s.Notion != nil {
		return s.Notion.EncryptionKey
	}
	key, err := s.Config.EncryptionKeyBytes()
	if err != nil {
		return nil
	}
	return key
}

func (s *Server) notionExportReady(r *http.Request) bool {
	if s.Notion == nil {
		return false
	}
	cfg, ok, err := s.Notion.Load(r.Context())
	if err != nil || !ok {
		return false
	}
	return notion.ExportReady(cfg)
}

func writeNotionExportErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, notion.ErrAlreadyExported):
		writeAPIError(w, http.StatusConflict, "conflict", notion.UserMessage(err))
	case errors.Is(err, notion.ErrNotConfigured),
		errors.Is(err, notion.ErrDatabaseMissing),
		errors.Is(err, notion.ErrRunNotDone):
		writeAPIError(w, http.StatusBadRequest, "validation_failed", notion.UserMessage(err))
	case errors.Is(err, store.ErrRunNotFound):
		writeAPIError(w, http.StatusNotFound, "not_found", "Revue introuvable.")
	default:
		slog.Error("notion export", "err", err)
		writeAPIError(w, http.StatusBadRequest, "validation_failed", notion.UserMessage(err))
	}
}
