package apiv1

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jeb-maker/revues/internal/integrations/jira"
)

// GetAdminJiraSettings serves GET /api/v1/admin/integrations/jira.
func (s *Server) GetAdminJiraSettings(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	svc := s.jiraService()
	if svc == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	writeJSON(w, http.StatusOK, mapJiraSettings(s.loadJiraMasked(r, svc)))
}

// PutAdminJiraSettings serves PUT /api/v1/admin/integrations/jira.
func (s *Server) PutAdminJiraSettings(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	svc := s.jiraService()
	if svc == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	var req JiraSettingsUpdate
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}

	cfg := jira.Config{
		InstanceType: jira.InstanceCloud,
		BaseURL:      strings.TrimSpace(req.BaseUrl),
		Email:        strings.TrimSpace(string(req.Email)),
		ProjectKey:   "",
		IssueType:    jira.DefaultIssueType,
	}
	if req.ProjectKey != nil {
		cfg.ProjectKey = strings.TrimSpace(*req.ProjectKey)
	}
	if req.IssueType != nil && strings.TrimSpace(*req.IssueType) != "" {
		cfg.IssueType = strings.TrimSpace(*req.IssueType)
	}

	submittedToken := ""
	if req.ApiToken != nil {
		submittedToken = *req.ApiToken
	}

	current, hasCurrent, err := svc.Load(r.Context())
	if err != nil {
		slog.Error("load jira for merge", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if hasCurrent {
		cfg.APIToken = jira.MergeSecret(current.APIToken, submittedToken)
		// Preserve server PAT if somehow present — Cloud-only API does not expose it.
		cfg.PAT = current.PAT
	} else {
		cfg.APIToken = submittedToken
	}

	if err := svc.Save(r.Context(), cfg); err != nil {
		if errors.Is(err, jira.ErrEncryptionNotConfigured) {
			writeAPIError(w, http.StatusBadRequest, "validation_failed",
				"REVUES_ENCRYPTION_KEY est requis pour enregistrer la configuration Jira.")
			return
		}
		writeAPIError(w, http.StatusBadRequest, "validation_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, mapJiraSettings(s.loadJiraMasked(r, svc)))
}

// DeleteAdminJiraSettings serves DELETE /api/v1/admin/integrations/jira.
func (s *Server) DeleteAdminJiraSettings(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	svc := s.jiraService()
	if svc == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if err := svc.Clear(r.Context()); err != nil {
		slog.Error("clear jira", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PostAdminJiraTest serves POST /api/v1/admin/integrations/jira/test.
func (s *Server) PostAdminJiraTest(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	svc := s.jiraService()
	if svc == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	cfg, configured, err := svc.Load(r.Context())
	if err != nil {
		slog.Error("load jira for test", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if !configured || !cfg.Configured() {
		writeAPIError(w, http.StatusBadRequest, "validation_failed",
			"Configurez et enregistrez Jira avant de tester la connexion.")
		return
	}

	if err := s.jiraClient().TestConnection(r.Context(), cfg); err != nil {
		slog.Error("jira test connection", "err", err)
		writeAPIError(w, http.StatusBadRequest, "validation_failed",
			"Échec du test de connexion Jira. Vérifiez l'URL et les identifiants.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type jiraMasked struct {
	configured  bool
	baseURL     string
	email       string
	projectKey  string
	issueType   string
	hasAPIToken bool
}

func (s *Server) jiraService() *jira.Service {
	if s.Integrations != nil && s.Integrations.Jira != nil {
		return s.Integrations.Jira
	}
	return nil
}

func (s *Server) jiraClient() *jira.Client {
	if s.JiraClient != nil {
		return s.JiraClient
	}
	return &jira.Client{}
}

func (s *Server) loadJiraMasked(r *http.Request, svc *jira.Service) jiraMasked {
	cfg, ok, err := svc.Load(r.Context())
	if err != nil {
		slog.Error("load jira settings", "err", err)
		return jiraMasked{issueType: jira.DefaultIssueType}
	}
	if !ok {
		return jiraMasked{issueType: jira.DefaultIssueType}
	}
	return jiraMasked{
		configured:  cfg.Configured(),
		baseURL:     cfg.BaseURL,
		email:       cfg.Email,
		projectKey:  cfg.ProjectKey,
		issueType:   cfg.IssueType,
		hasAPIToken: jira.HasSecret(cfg.APIToken),
	}
}

func mapJiraSettings(m jiraMasked) JiraSettings {
	issueType := m.issueType
	if issueType == "" {
		issueType = jira.DefaultIssueType
	}
	return JiraSettings{
		Configured:  m.configured,
		BaseUrl:     m.baseURL,
		Email:       m.email,
		ProjectKey:  m.projectKey,
		IssueType:   issueType,
		HasApiToken: m.hasAPIToken,
	}
}
