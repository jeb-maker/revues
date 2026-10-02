package apiv1

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jeb-maker/revues/internal/attachments"
	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	adminintegrations "github.com/jeb-maker/revues/internal/features/admin/integrations"
	adminsettings "github.com/jeb-maker/revues/internal/features/admin/settings"
	authfeature "github.com/jeb-maker/revues/internal/features/auth"
	"github.com/jeb-maker/revues/internal/features/checklisttemplates"
	"github.com/jeb-maker/revues/internal/features/organizations"
	"github.com/jeb-maker/revues/internal/features/subjects"
	"github.com/jeb-maker/revues/internal/integrations/webhooks"
	"github.com/jeb-maker/revues/internal/store"
	appmiddleware "github.com/jeb-maker/revues/internal/web/middleware"
)

// Server implements the OpenAPI ServerInterface for /api/v1.
type Server struct {
	Auth         *authfeature.Service
	Orgs         *organizations.Service
	Templates    *checklisttemplates.Service
	Store        *store.Store
	Config       config.Config
	Sessions     *auth.SessionManager
	Webhooks     *webhooks.Dispatcher
	Settings     *adminsettings.SettingsService
	Integrations *adminintegrations.IntegrationsService
	Attachments  *attachments.Service
}

// NewServer returns the API v1 server implementation.
func NewServer(
	authSvc *authfeature.Service,
	orgSvc *organizations.Service,
	templates *checklisttemplates.Service,
	st *store.Store,
	cfg config.Config,
	sessions *auth.SessionManager,
	hooks *webhooks.Dispatcher,
	settingsSvc *adminsettings.SettingsService,
	integrationsSvc *adminintegrations.IntegrationsService,
	attachmentsSvc *attachments.Service,
) *Server {
	return &Server{
		Auth:         authSvc,
		Orgs:         orgSvc,
		Templates:    templates,
		Store:        st,
		Config:       cfg,
		Sessions:     sessions,
		Webhooks:     hooks,
		Settings:     settingsSvc,
		Integrations: integrationsSvc,
		Attachments:  attachmentsSvc,
	}
}

// GetHealth serves GET /api/v1/health.
func (s *Server) GetHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{Status: HealthResponseStatusOk})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write api json response", "err", err)
	}
}

func writeAPIError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, ErrorResponse{
		Error: ErrorBody{Code: code, Message: message},
	})
}

// ensureSubjectAccess loads a subject and checks visibility. Returns false after writing the response.
func (s *Server) ensureSubjectAccess(w http.ResponseWriter, r *http.Request, subjectID int64) (*store.Subject, *store.User, store.SubjectAccess, bool) {
	user, ok := requireUser(w, r)
	if !ok {
		return nil, nil, store.SubjectAccess{}, false
	}
	if !requireOrg(w, r) {
		return nil, nil, store.SubjectAccess{}, false
	}

	subject, err := s.Store.SubjectByID(r.Context(), subjectID)
	if err != nil {
		if errors.Is(err, store.ErrSubjectNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Sujet introuvable.")
			return nil, nil, store.SubjectAccess{}, false
		}
		slog.Error("load subject", "err", err, "subject_id", subjectID)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return nil, nil, store.SubjectAccess{}, false
	}

	access, err := s.Store.ResolveSubjectAccess(r.Context(), user.ID, subject.ID, user.Role)
	if err != nil {
		slog.Error("resolve subject access", "err", err, "subject_id", subjectID)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return nil, nil, store.SubjectAccess{}, false
	}
	if !subjects.CanViewAccess(access) {
		writeAPIError(w, http.StatusNotFound, "not_found", "Sujet introuvable.")
		return nil, nil, store.SubjectAccess{}, false
	}
	return subject, user, access, true
}

func requireUser(w http.ResponseWriter, r *http.Request) (*store.User, bool) {
	user, ok := userFromRequest(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthenticated", "Authentification requise.")
		return nil, false
	}
	return user, true
}

func requireOrg(w http.ResponseWriter, r *http.Request) bool {
	if _, ok := orgFromRequest(r); ok {
		return true
	}
	writeAPIError(w, http.StatusForbidden, "org_required", "Organisation active requise.")
	return false
}

// requireOrgAdmin enforces org owner/admin (or global admin) + active org.
func (s *Server) requireOrgAdmin(w http.ResponseWriter, r *http.Request) (*store.User, *store.Organization, bool) {
	user, ok := requireUser(w, r)
	if !ok {
		return nil, nil, false
	}
	org, ok := orgFromRequest(r)
	if !ok {
		writeAPIError(w, http.StatusForbidden, "org_required", "Organisation active requise.")
		return nil, nil, false
	}
	if s.Store == nil || !appmiddleware.CanManageOrgUsers(r.Context(), s.Store, user) {
		writeAPIError(w, http.StatusForbidden, "forbidden", "Droits administrateur organisation requis.")
		return nil, nil, false
	}
	return user, org, true
}
