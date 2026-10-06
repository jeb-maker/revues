package apiv1

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/integrations/confluence"
	"github.com/jeb-maker/revues/internal/store"
)

// GetRunConfluence serves GET /api/v1/runs/{runId}/confluence.
func (s *Server) GetRunConfluence(w http.ResponseWriter, r *http.Request, runID RunId) {
	run, user, access, ok := s.ensureRunAccess(w, r, runID)
	if !ok {
		return
	}

	configured := false
	if svc := s.confluenceService(); svc != nil {
		cfg, okCfg, err := svc.Load(r.Context())
		if err != nil {
			slog.Error("load confluence for run state", "err", err)
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
			return
		}
		configured = okCfg && cfg.Configured()
	}

	canPublish := configured &&
		run.Status == store.RunStatusDone &&
		runs.CanPublishConfluenceAccess(user, access)

	var link *ConfluenceLink
	if u := strings.TrimSpace(run.ConfluenceURL); u != "" {
		link = &ConfluenceLink{Url: u}
	}

	writeJSON(w, http.StatusOK, RunConfluence{
		Configured: configured,
		CanPublish: canPublish,
		Link:       link,
	})
}

// PostRunConfluencePublish serves POST /api/v1/runs/{runId}/confluence.
func (s *Server) PostRunConfluencePublish(w http.ResponseWriter, r *http.Request, runID RunId) {
	run, user, access, ok := s.ensureRunAccess(w, r, runID)
	if !ok {
		return
	}
	if !runs.CanPublishConfluenceAccess(user, access) {
		writeAPIError(w, http.StatusNotFound, "not_found", "Revue introuvable.")
		return
	}
	if run.Status != store.RunStatusDone {
		writeAPIError(w, http.StatusBadRequest, "validation_failed",
			"Seules les revues clôturées peuvent être publiées vers Confluence.")
		return
	}

	svc := s.confluenceService()
	if svc == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	publisher := &confluence.PublishService{
		Store:         s.Store,
		EncryptionKey: svc.EncryptionKey,
		Client:        s.confluenceClient(),
	}
	result, err := publisher.Publish(r.Context(), runID, confluence.PublishInput{})
	if err != nil {
		switch {
		case errors.Is(err, confluence.ErrNotConfigured):
			writeAPIError(w, http.StatusBadRequest, "validation_failed",
				"Confluence n'est pas configuré pour cette organisation.")
		case errors.Is(err, confluence.ErrRunNotDone):
			writeAPIError(w, http.StatusBadRequest, "validation_failed",
				"Seules les revues clôturées peuvent être publiées vers Confluence.")
		case errors.Is(err, confluence.ErrCreateFailed), errors.Is(err, confluence.ErrConnectionFailed):
			slog.Error("confluence publish", "err", err)
			writeAPIError(w, http.StatusBadRequest, "validation_failed",
				"Échec de la publication Confluence. Vérifiez la configuration et réessayez.")
		default:
			slog.Error("confluence publish", "err", err)
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		}
		return
	}

	writeJSON(w, http.StatusCreated, ConfluenceLink{Url: result.URL})
}
