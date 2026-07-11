package organizations

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/middleware"
	"github.com/jeb-maker/revues/internal/web/templates"
)

const dashboardPath = "/projects"

// OrgStore is the persistence layer for organization onboarding handlers.
type OrgStore interface {
	CountUserOrganizations(ctx context.Context, userID int64) (int, error)
	ListUserOrganizations(ctx context.Context, userID int64) ([]store.OrganizationMembership, error)
	CreateOrganization(ctx context.Context, name, slug string, createdBy int64) (*store.Organization, error)
	AddOrganizationMember(ctx context.Context, organizationID, userID int64, role string) error
	OrganizationMemberRole(ctx context.Context, organizationID, userID int64) (string, bool, error)
}

// Deps holds shared dependencies for organization HTTP handlers.
type Deps struct {
	Templates     *template.Template
	Store         OrgStore
	Sessions      *auth.SessionManager
	SessionSecret string
	SecureCookies bool
}

// Organizations handles self-service org creation and selection.
type Organizations struct {
	Deps
}

// PostLoginRoute decides how to seed the session organization and where to redirect.
func PostLoginRoute(ctx context.Context, st interface {
	CountUserOrganizations(ctx context.Context, userID int64) (int, error)
	ListUserOrganizations(ctx context.Context, userID int64) ([]store.OrganizationMembership, error)
}, userID int64) (sessionOrgID int64, redirect string, err error) {
	count, err := st.CountUserOrganizations(ctx, userID)
	if err != nil {
		return 0, "", fmt.Errorf("count user organizations: %w", err)
	}

	switch count {
	case 0:
		return auth.SessionOrgPending, "/org/new", nil
	case 1:
		memberships, err := st.ListUserOrganizations(ctx, userID)
		if err != nil {
			return 0, "", fmt.Errorf("list user organizations: %w", err)
		}
		return memberships[0].Organization.ID, dashboardPath, nil
	default:
		return auth.SessionOrgPending, "/org/select", nil
	}
}

func (h *Organizations) pageData(r *http.Request, title string) templates.PageData {
	data := templates.PageData{Title: title}
	if user, ok := middleware.UserFromContext(r.Context()); ok {
		data.User = user
		if token := middleware.SessionTokenFromContext(r); token != "" {
			data.CSRFToken = auth.CSRFToken(token, h.SessionSecret)
		}
	}
	return data
}

// NewForm renders the organization creation form.
func (h *Organizations) NewForm(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	count, err := h.Store.CountUserOrganizations(r.Context(), user.ID)
	if err != nil {
		slog.Error("count user organizations", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Redirect(w, r, h.onboardingRedirect(count), http.StatusFound)
		return
	}

	data := templates.OrgNewData{
		PageData: h.pageData(r, "Nouvelle organisation"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "org_new", data); err != nil {
		slog.Error("render org new form", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Create stores a new organization and activates it on the session.
func (h *Organizations) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	count, err := h.Store.CountUserOrganizations(r.Context(), user.ID)
	if err != nil {
		slog.Error("count user organizations", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Redirect(w, r, h.onboardingRedirect(count), http.StatusFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	slugInput := strings.TrimSpace(r.FormValue("slug"))
	if slugInput == "" {
		slugInput = name
	}

	data := templates.OrgNewData{
		PageData: h.pageData(r, "Nouvelle organisation"),
		Name:     name,
		Slug:     slugInput,
	}

	if name == "" {
		data.Error = "Le nom est obligatoire."
		h.renderNewForm(w, data)
		return
	}

	org, err := h.Store.CreateOrganization(r.Context(), name, slugInput, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrInvalidOrganizationSlug):
			data.Error = "Identifiant invalide (lettres minuscules, chiffres et tirets uniquement)."
		case errors.Is(err, store.ErrOrganizationSlugTaken):
			data.Error = "Cet identifiant est déjà utilisé."
		default:
			slog.Error("create organization", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		h.renderNewForm(w, data)
		return
	}

	if err := h.Store.AddOrganizationMember(r.Context(), org.ID, user.ID, store.OrgRoleOwner); err != nil {
		slog.Error("add organization owner", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := h.activateOrganization(w, r, org.ID); err != nil {
		slog.Error("activate organization", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, dashboardPath, http.StatusSeeOther)
}

// SelectForm lists organizations the user can choose from.
func (h *Organizations) SelectForm(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	count, err := h.Store.CountUserOrganizations(r.Context(), user.ID)
	if err != nil {
		slog.Error("count user organizations", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	switch count {
	case 0:
		http.Redirect(w, r, "/org/new", http.StatusFound)
		return
	case 1:
		if err := h.autoSelectSingleOrg(w, r, user.ID); err != nil {
			slog.Error("auto select organization", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, dashboardPath, http.StatusFound)
		return
	}

	memberships, err := h.Store.ListUserOrganizations(r.Context(), user.ID)
	if err != nil {
		slog.Error("list user organizations", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	defaultOrgID := h.defaultOrgID(r, memberships)

	data := templates.OrgSelectData{
		PageData:      h.pageData(r, "Choisir une organisation"),
		Organizations: memberships,
		DefaultOrgID:  defaultOrgID,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "org_select", data); err != nil {
		slog.Error("render org select", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Select activates the chosen organization on the session.
func (h *Organizations) Select(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	orgID, err := strconv.ParseInt(r.FormValue("organization_id"), 10, 64)
	if err != nil || orgID <= 0 {
		h.renderSelectError(w, r, user.ID, "Choisissez une organisation.")
		return
	}

	if _, member, err := h.Store.OrganizationMemberRole(r.Context(), orgID, user.ID); err != nil || !member {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.activateOrganization(w, r, orgID); err != nil {
		slog.Error("activate organization", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, dashboardPath, http.StatusSeeOther)
}

func (h *Organizations) onboardingRedirect(count int) string {
	if count == 1 {
		return dashboardPath
	}
	return "/org/select"
}

func (h *Organizations) defaultOrgID(r *http.Request, memberships []store.OrganizationMembership) int64 {
	lastID := auth.LastOrgIDFromRequest(r)
	if lastID <= 0 {
		return 0
	}
	for _, m := range memberships {
		if m.Organization.ID == lastID {
			return lastID
		}
	}
	return 0
}

func (h *Organizations) autoSelectSingleOrg(w http.ResponseWriter, r *http.Request, userID int64) error {
	memberships, err := h.Store.ListUserOrganizations(r.Context(), userID)
	if err != nil {
		return err
	}
	if len(memberships) != 1 {
		return fmt.Errorf("expected one organization, got %d", len(memberships))
	}
	return h.activateOrganization(w, r, memberships[0].Organization.ID)
}

func (h *Organizations) activateOrganization(w http.ResponseWriter, r *http.Request, orgID int64) error {
	token := middleware.SessionTokenFromContext(r)
	if token == "" {
		return fmt.Errorf("session token missing")
	}
	if err := h.Sessions.SetActiveOrganization(r.Context(), token, orgID); err != nil {
		return err
	}
	auth.SetLastOrgCookie(w, orgID, h.SecureCookies)
	return nil
}

func (h *Organizations) renderNewForm(w http.ResponseWriter, data templates.OrgNewData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "org_new", data); err != nil {
		slog.Error("render org new form", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Organizations) renderSelectError(w http.ResponseWriter, r *http.Request, userID int64, msg string) {
	memberships, err := h.Store.ListUserOrganizations(r.Context(), userID)
	if err != nil {
		slog.Error("list user organizations", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := templates.OrgSelectData{
		PageData:      h.pageData(r, "Choisir une organisation"),
		Organizations: memberships,
		DefaultOrgID:  h.defaultOrgID(r, memberships),
		Error:         msg,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "org_select", data); err != nil {
		slog.Error("render org select", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
