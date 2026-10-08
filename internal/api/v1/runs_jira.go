package apiv1

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/integrations/atlassian"
	"github.com/jeb-maker/revues/internal/integrations/jira"
	"github.com/jeb-maker/revues/internal/store"
)

// GetRunItemJira serves GET /api/v1/runs/{runId}/items/{itemId}/jira.
func (s *Server) GetRunItemJira(w http.ResponseWriter, r *http.Request, runID RunId, itemID RunItemId) {
	run, user, access, ok := s.ensureRunAccess(w, r, runID)
	if !ok {
		return
	}
	item, err := s.Store.RunItemByID(r.Context(), run.ID, itemID)
	if err != nil {
		if errors.Is(err, store.ErrRunItemNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Point introuvable.")
			return
		}
		slog.Error("load run item for jira get", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	out, ok := s.buildRunItemJira(w, r, run, item, user, access)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// PutRunItemJiraLink serves PUT /api/v1/runs/{runId}/items/{itemId}/jira.
func (s *Server) PutRunItemJiraLink(w http.ResponseWriter, r *http.Request, runID RunId, itemID RunItemId) {
	run, user, access, ok := s.ensureRunAccess(w, r, runID)
	if !ok {
		return
	}
	if !runs.CanLinkJiraAccess(user, access) {
		writeAPIError(w, http.StatusNotFound, "not_found", "Revue introuvable.")
		return
	}
	if _, err := s.Store.RunItemByID(r.Context(), run.ID, itemID); err != nil {
		if errors.Is(err, store.ErrRunItemNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Point introuvable.")
			return
		}
		slog.Error("load run item for jira link", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	var req JiraLinkRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}

	link, err := s.jiraLinkService().LinkRunItem(r.Context(), user.ID, itemID, strings.TrimSpace(req.Issue))
	if err != nil {
		writeAPIError(w, jiraLinkStatus(err), jiraLinkCode(err), jiraLinkMessage(err))
		return
	}
	writeJSON(w, http.StatusOK, mapJiraLink(link))
}

// PostRunItemJiraCreate serves POST /api/v1/runs/{runId}/items/{itemId}/jira.
func (s *Server) PostRunItemJiraCreate(w http.ResponseWriter, r *http.Request, runID RunId, itemID RunItemId) {
	run, user, access, ok := s.ensureRunAccess(w, r, runID)
	if !ok {
		return
	}
	if !runs.CanLinkJiraAccess(user, access) {
		writeAPIError(w, http.StatusNotFound, "not_found", "Revue introuvable.")
		return
	}

	var req JiraCreateRequest
	if r.Body != nil && r.ContentLength != 0 {
		if err := decodeJSONBody(r, &req); err != nil && !errors.Is(err, io.EOF) {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
			return
		}
	}

	subject, err := s.Store.SubjectByID(r.Context(), run.SubjectID)
	if err != nil {
		slog.Error("load subject for jira create", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	title, _ := s.Store.RunDisplayLabelForRun(r.Context(), run)
	itemCtx := jira.RunItemContext{
		SubjectName: subject.Name,
		RunTitle:    title,
		ItemURL:     s.runItemURL(run.ID, itemID),
	}

	input := jira.CreateInput{}
	if req.Title != nil {
		input.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		input.Description = strings.TrimSpace(*req.Description)
	}

	link, err := s.jiraCreateService().CreateRunItem(r.Context(), user.ID, run.ID, itemID, input, itemCtx)
	if err != nil {
		status, code, msg := jiraCreateError(err)
		writeAPIError(w, status, code, msg)
		return
	}
	writeJSON(w, http.StatusCreated, mapJiraLink(link))
}

func (s *Server) buildRunItemJira(
	w http.ResponseWriter,
	r *http.Request,
	run *store.ChecklistRun,
	item *store.RunItem,
	user *store.User,
	access store.SubjectAccess,
) (RunItemJira, bool) {
	configured := false
	orgProjectKey := ""
	var jiraCfg jira.Config
	if svc := s.jiraService(); svc != nil {
		cfg, ok, err := svc.Load(r.Context())
		if err != nil {
			slog.Error("load jira for item status", "err", err)
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
			return RunItemJira{}, false
		}
		if ok {
			jiraCfg = cfg
			configured = cfg.Configured() || strings.TrimSpace(cfg.BaseURL) != ""
			orgProjectKey = cfg.ProjectKey
		}
	}

	oauthEnabled := s.Config.AtlassianOAuthConfigured()
	userOAuthConnected := false
	if oauthEnabled {
		if tokens := s.atlassianTokenService(); tokens != nil {
			connected, _, err := tokens.Connected(r.Context(), user.ID)
			if err != nil {
				slog.Error("atlassian oauth connected check", "err", err)
				writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
				return RunItemJira{}, false
			}
			userOAuthConnected = connected
		}
	}

	subjectKey := ""
	if subject, err := s.Store.SubjectByID(r.Context(), run.SubjectID); err == nil {
		subjectKey = subject.JiraProjectKey
	} else if !errors.Is(err, store.ErrSubjectNotFound) {
		slog.Error("subject for jira project key", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return RunItemJira{}, false
	}
	projectKey := jira.EffectiveProjectKey(orgProjectKey, subjectKey)

	var link *JiraLink
	stored, err := s.Store.IntegrationLinkByRunItemAndType(r.Context(), item.ID, store.IntegrationTypeJira)
	if err == nil && stored != nil {
		mapped := mapJiraLink(stored)
		fetchCfg := s.jiraStatusConfig(r, user.ID, jiraCfg, oauthEnabled, userOAuthConnected)
		if fetchCfg.Configured() {
			if info, fetchErr := s.jiraClient().GetIssueInfo(r.Context(), fetchCfg, stored.ExternalKey); fetchErr == nil {
				if st := strings.TrimSpace(info.Status); st != "" {
					mapped.Status = &st
				}
				if info.Key != "" {
					mapped.ExternalKey = info.Key
				}
			} else {
				slog.Debug("jira status fetch skipped", "key", stored.ExternalKey, "err", fetchErr)
			}
		}
		link = &mapped
	} else if err != nil && !errors.Is(err, store.ErrIntegrationLinkNotFound) {
		slog.Error("load jira link", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return RunItemJira{}, false
	}

	canLink := runs.CanLinkJiraAccess(user, access)
	oauthRequired := oauthEnabled && !userOAuthConnected
	out := RunItemJira{
		Configured:           configured,
		CanLink:              canLink && (!oauthEnabled || userOAuthConnected),
		Link:                 link,
		UserOauthConnected:   &userOAuthConnected,
		OauthRequired:        &oauthRequired,
		AtlassianOauthEnabled: &oauthEnabled,
	}

	canCreate := configured && canLink && link == nil &&
		item.Status == store.RunItemStatusNOK && strings.TrimSpace(projectKey) != "" &&
		(!oauthEnabled || userOAuthConnected)
	out.CanCreate = &canCreate

	if item.Status == store.RunItemStatusNOK {
		subject, err := s.Store.SubjectByID(r.Context(), run.SubjectID)
		if err != nil {
			slog.Error("subject for jira defaults", "err", err)
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
			return RunItemJira{}, false
		}
		title, _ := s.Store.RunDisplayLabelForRun(r.Context(), run)
		defTitle, defDesc := jira.DefaultIssueContent(item, jira.RunItemContext{
			SubjectName: subject.Name,
			RunTitle:    title,
			ItemURL:     s.runItemURL(run.ID, item.ID),
		})
		out.DefaultTitle = &defTitle
		out.DefaultDescription = &defDesc
	}

	return out, true
}

func (s *Server) jiraStatusConfig(r *http.Request, userID int64, orgCfg jira.Config, oauthEnabled, userConnected bool) jira.Config {
	if oauthEnabled && userConnected {
		if tokens := s.atlassianTokenService(); tokens != nil {
			creds, err := tokens.EnsureAccessToken(r.Context(), userID)
			if err == nil {
				cfg := orgCfg
				cfg.InstanceType = jira.InstanceCloud
				cfg.AccessToken = creds.AccessToken
				cfg.CloudID = creds.CloudID
				if site := strings.TrimSpace(creds.SiteURL); site != "" {
					cfg.SiteURL = site
					if cfg.BaseURL == "" {
						cfg.BaseURL = site
					}
				}
				return cfg
			}
			slog.Debug("jira status user oauth fallback", "err", err)
		}
	}
	return orgCfg
}

func (s *Server) atlassianTokenService() *atlassian.TokenService {
	key, err := s.Config.EncryptionKeyBytes()
	if err != nil || len(key) == 0 {
		return nil
	}
	svc := &atlassian.TokenService{
		Store:         s.Store,
		EncryptionKey: key,
	}
	if s.Auth != nil && s.Auth.Atlassian != nil {
		svc.OAuth = s.Auth.Atlassian
	}
	return svc
}

func (s *Server) jiraLinkService() *jira.LinkService {
	var key []byte
	if svc := s.jiraService(); svc != nil {
		key = svc.EncryptionKey
	}
	requireOAuth := s.Config.AtlassianOAuthConfigured()
	return &jira.LinkService{
		Store:            s.Store,
		EncryptionKey:    key,
		Client:           s.jiraClient(),
		Tokens:           s.atlassianTokenService(),
		RequireUserOAuth: requireOAuth,
	}
}

func (s *Server) jiraCreateService() *jira.CreateService {
	var key []byte
	if svc := s.jiraService(); svc != nil {
		key = svc.EncryptionKey
	}
	requireOAuth := s.Config.AtlassianOAuthConfigured()
	return &jira.CreateService{
		Store:            s.Store,
		EncryptionKey:    key,
		Client:           s.jiraClient(),
		Tokens:           s.atlassianTokenService(),
		RequireUserOAuth: requireOAuth,
	}
}

func (s *Server) runItemURL(runID, itemID int64) string {
	base := strings.TrimRight(s.Config.BaseURL, "/")
	if base == "" {
		base = "http://localhost:8080"
	}
	return fmt.Sprintf("%s/runs/%s/items/%s",
		base,
		strconv.FormatInt(runID, 10),
		strconv.FormatInt(itemID, 10),
	)
}

func mapJiraLink(link *store.IntegrationLink) JiraLink {
	if link == nil {
		return JiraLink{}
	}
	return JiraLink{
		ExternalKey: link.ExternalKey,
		ExternalUrl: link.ExternalURL,
	}
}

func jiraLinkStatus(err error) int {
	switch {
	case errors.Is(err, jira.ErrUserOAuthRequired):
		return http.StatusBadRequest
	case errors.Is(err, jira.ErrNotConfigured),
		errors.Is(err, jira.ErrInvalidIssueReference),
		errors.Is(err, jira.ErrIssueNotFound),
		errors.Is(err, jira.ErrConnectionFailed):
		return http.StatusBadRequest
	default:
		msg := err.Error()
		if strings.Contains(msg, "requise") || strings.Contains(msg, "invalide") {
			return http.StatusBadRequest
		}
		return http.StatusBadRequest
	}
}

func jiraLinkCode(err error) string {
	if errors.Is(err, jira.ErrUserOAuthRequired) {
		return "oauth_required"
	}
	return "validation_failed"
}

func jiraLinkMessage(err error) string {
	switch {
	case errors.Is(err, jira.ErrUserOAuthRequired):
		return "Connectez votre compte Atlassian pour lier une issue Jira."
	case errors.Is(err, jira.ErrNotConfigured):
		return "Jira n'est pas configuré. Contactez un administrateur."
	case errors.Is(err, jira.ErrInvalidIssueReference):
		return "Clé ou URL Jira invalide (ex. PROJ-123)."
	case errors.Is(err, jira.ErrIssueNotFound):
		return "Issue Jira introuvable."
	case errors.Is(err, jira.ErrConnectionFailed):
		return "Impossible de contacter Jira. Réessayez plus tard."
	default:
		return "Impossible de lier l'issue Jira."
	}
}

func jiraCreateError(err error) (status int, code, message string) {
	switch {
	case errors.Is(err, jira.ErrUserOAuthRequired):
		return http.StatusBadRequest, "oauth_required", "Connectez votre compte Atlassian pour créer un ticket Jira."
	case errors.Is(err, jira.ErrNotConfigured):
		return http.StatusBadRequest, "validation_failed", "Jira n'est pas configuré. Contactez un administrateur."
	case errors.Is(err, jira.ErrProjectKeyMissing):
		return http.StatusBadRequest, "validation_failed", "Clé projet Jira manquante dans la configuration admin."
	case errors.Is(err, jira.ErrNotNOK):
		return http.StatusBadRequest, "validation_failed", "Seuls les points non validés peuvent générer un ticket Jira."
	case errors.Is(err, jira.ErrAlreadyLinked):
		return http.StatusConflict, "conflict", "Une issue Jira est déjà liée à ce point."
	case errors.Is(err, jira.ErrConnectionFailed):
		return http.StatusBadRequest, "validation_failed", "Impossible de contacter Jira. Réessayez plus tard."
	case errors.Is(err, jira.ErrCreateFailed):
		return http.StatusBadRequest, "validation_failed", "Jira a refusé la création du ticket. Vérifiez la configuration (projet, type d'issue)."
	case errors.Is(err, store.ErrRunItemNotFound):
		return http.StatusNotFound, "not_found", "Point introuvable."
	default:
		return http.StatusBadRequest, "validation_failed", "Impossible de créer le ticket Jira."
	}
}
