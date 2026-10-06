package apiv1

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/jeb-maker/revues/internal/integrations/confluence"
)

// GetAdminConfluenceSettings serves GET /api/v1/admin/integrations/confluence.
func (s *Server) GetAdminConfluenceSettings(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	svc := s.confluenceService()
	if svc == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	writeJSON(w, http.StatusOK, mapConfluenceSettings(s.loadConfluenceMasked(r, svc)))
}

// PutAdminConfluenceSettings serves PUT /api/v1/admin/integrations/confluence.
func (s *Server) PutAdminConfluenceSettings(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	svc := s.confluenceService()
	if svc == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	var req ConfluenceSettingsUpdate
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}

	cfg := confluence.Config{
		BaseURL:  strings.TrimSpace(req.BaseUrl),
		Email:    strings.TrimSpace(string(req.Email)),
		SpaceKey: strings.TrimSpace(req.SpaceKey),
	}
	if req.ParentPageId != nil {
		cfg.ParentPageID = strings.TrimSpace(*req.ParentPageId)
	}

	submittedToken := ""
	if req.ApiToken != nil {
		submittedToken = *req.ApiToken
	}

	current, hasCurrent, err := svc.Load(r.Context())
	if err != nil {
		slog.Error("load confluence for merge", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if hasCurrent {
		cfg.APIToken = confluence.MergeSecret(current.APIToken, submittedToken)
	} else {
		cfg.APIToken = submittedToken
	}

	if err := confluence.Validate(cfg); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", err.Error())
		return
	}
	if err := svc.Save(r.Context(), cfg); err != nil {
		writeConfigSaveError(w, err, confluence.ErrEncryptionNotConfigured,
			"REVUES_ENCRYPTION_KEY est requis pour enregistrer la configuration Confluence.")
		return
	}

	writeJSON(w, http.StatusOK, mapConfluenceSettings(s.loadConfluenceMasked(r, svc)))
}

// DeleteAdminConfluenceSettings serves DELETE /api/v1/admin/integrations/confluence.
func (s *Server) DeleteAdminConfluenceSettings(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	svc := s.confluenceService()
	if svc == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if err := svc.Clear(r.Context()); err != nil {
		slog.Error("clear confluence", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PostAdminConfluenceTest serves POST /api/v1/admin/integrations/confluence/test.
func (s *Server) PostAdminConfluenceTest(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	svc := s.confluenceService()
	if svc == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	cfg, configured, err := svc.Load(r.Context())
	if err != nil {
		slog.Error("load confluence for test", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if !configured || !cfg.Configured() {
		writeAPIError(w, http.StatusBadRequest, "validation_failed",
			"Configurez et enregistrez Confluence avant de tester la connexion.")
		return
	}

	if err := s.confluenceClient().TestConnection(r.Context(), cfg); err != nil {
		slog.Error("confluence test connection", "err", err)
		writeAPIError(w, http.StatusBadRequest, "validation_failed",
			"Échec du test de connexion Confluence. Vérifiez l'URL et les identifiants.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type confluenceMasked struct {
	configured   bool
	baseURL      string
	email        string
	spaceKey     string
	parentPageID string
	hasAPIToken  bool
}

func (s *Server) confluenceService() *confluence.Service {
	if s.Integrations != nil && s.Integrations.Confluence != nil {
		return s.Integrations.Confluence
	}
	return nil
}

func (s *Server) confluenceClient() *confluence.Client {
	if s.ConfluenceClient != nil {
		return s.ConfluenceClient
	}
	return &confluence.Client{}
}

func (s *Server) loadConfluenceMasked(r *http.Request, svc *confluence.Service) confluenceMasked {
	cfg, ok, err := svc.Load(r.Context())
	if err != nil {
		slog.Error("load confluence settings", "err", err)
		return confluenceMasked{}
	}
	if !ok {
		return confluenceMasked{}
	}
	return confluenceMasked{
		configured:   cfg.Configured(),
		baseURL:      cfg.BaseURL,
		email:        cfg.Email,
		spaceKey:     cfg.SpaceKey,
		parentPageID: cfg.ParentPageID,
		hasAPIToken:  confluence.HasSecret(cfg.APIToken),
	}
}

func mapConfluenceSettings(m confluenceMasked) ConfluenceSettings {
	return ConfluenceSettings{
		Configured:   m.configured,
		BaseUrl:      m.baseURL,
		Email:        m.email,
		SpaceKey:     m.spaceKey,
		ParentPageId: m.parentPageID,
		HasApiToken:  m.hasAPIToken,
	}
}
