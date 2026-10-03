package apiv1

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/jeb-maker/revues/internal/integrations/notion"
)

// GetAdminNotionSettings serves GET /api/v1/admin/integrations/notion.
func (s *Server) GetAdminNotionSettings(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	if s.Notion == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	writeJSON(w, http.StatusOK, s.mapNotionSettings(r))
}

// PutAdminNotionSettings serves PUT /api/v1/admin/integrations/notion.
func (s *Server) PutAdminNotionSettings(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	if s.Notion == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	var req NotionSettingsUpdate
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}

	cfg := notion.Config{
		WorkspaceName:     "",
		DefaultDatabaseID: "",
	}
	if req.WorkspaceName != nil {
		cfg.WorkspaceName = strings.TrimSpace(*req.WorkspaceName)
	}
	if req.DefaultDatabaseId != nil {
		cfg.DefaultDatabaseID = strings.TrimSpace(*req.DefaultDatabaseId)
	}
	submittedToken := ""
	if req.ApiToken != nil {
		submittedToken = *req.ApiToken
	}

	current, hasCurrent, err := s.Notion.Load(r.Context())
	if err != nil {
		slog.Error("load notion for merge", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if hasCurrent {
		cfg.APIToken = notion.MergeSecret(current.APIToken, submittedToken)
		if req.WorkspaceName == nil {
			cfg.WorkspaceName = current.WorkspaceName
		}
		if req.DefaultDatabaseId == nil {
			cfg.DefaultDatabaseID = current.DefaultDatabaseID
		}
	} else {
		cfg.APIToken = submittedToken
	}

	if err := notion.Validate(cfg); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", err.Error())
		return
	}
	if err := s.Notion.Save(r.Context(), cfg); err != nil {
		writeConfigSaveError(w, err, notion.ErrEncryptionNotConfigured,
			"REVUES_ENCRYPTION_KEY est requis pour enregistrer la configuration Notion.")
		return
	}

	writeJSON(w, http.StatusOK, s.mapNotionSettings(r))
}

// DeleteAdminNotionSettings serves DELETE /api/v1/admin/integrations/notion.
func (s *Server) DeleteAdminNotionSettings(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	if s.Notion == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if err := s.Notion.Clear(r.Context()); err != nil {
		slog.Error("clear notion", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PostAdminNotionTest serves POST /api/v1/admin/integrations/notion/test.
func (s *Server) PostAdminNotionTest(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	if s.Notion == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	cfg, ok, err := s.Notion.Load(r.Context())
	if err != nil {
		slog.Error("load notion for test", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if !ok || !cfg.Configured() {
		writeAPIError(w, http.StatusBadRequest, "validation_failed",
			"Configurez et enregistrez Notion avant de tester la connexion.")
		return
	}

	info, err := s.notionClient().TestConnection(r.Context(), cfg)
	if err != nil {
		slog.Error("notion test connection", "err", err)
		writeAPIError(w, http.StatusBadRequest, "validation_failed", notion.UserMessage(err))
		return
	}

	msg := "Connexion Notion réussie"
	if info.WorkspaceName != "" {
		msg += " (" + info.WorkspaceName + ")"
	}
	resp := NotionTestResponse{Ok: true, Message: msg}
	if info.UserName != "" {
		resp.UserName = &info.UserName
	}
	if info.WorkspaceName != "" {
		resp.WorkspaceName = &info.WorkspaceName
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) mapNotionSettings(r *http.Request) NotionSettings {
	cfg, ok, err := s.Notion.Load(r.Context())
	if err != nil {
		slog.Error("load notion settings", "err", err)
		return NotionSettings{}
	}
	if !ok {
		return NotionSettings{}
	}
	return NotionSettings{
		Configured:        cfg.Configured(),
		ExportReady:       notion.ExportReady(cfg),
		WorkspaceName:     cfg.WorkspaceName,
		DefaultDatabaseId: cfg.DefaultDatabaseID,
		HasApiToken:       notion.HasSecret(cfg.APIToken),
	}
}

func (s *Server) notionClient() *notion.Client {
	if s.NotionClient != nil {
		return s.NotionClient
	}
	return &notion.Client{}
}
