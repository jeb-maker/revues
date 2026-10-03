package apiv1

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/features/organizations"
	"github.com/jeb-maker/revues/internal/store"
	appmiddleware "github.com/jeb-maker/revues/internal/web/middleware"
)

// ListOrganizations serves GET /api/v1/orgs.
func (s *Server) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	user, ok := appmiddleware.UserFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthenticated", "Authentification requise.")
		return
	}
	if s.Orgs == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	token := appmiddleware.SessionTokenFromContext(r)
	result, err := s.Orgs.List(r.Context(), user, token, r)
	if err != nil {
		slog.Error("list organizations", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	resp := OrganizationListResponse{
		Organizations: mapMemberships(r.Context(), s.Store, user, result.Memberships, result.ActiveOrganizationID),
		Invitations:   mapInvitations(result.Invitations),
		CanCreate:     result.CanCreate,
		Redirect:      &result.Redirect,
	}
	if result.ActiveOrganizationID > 0 {
		id := result.ActiveOrganizationID
		resp.ActiveOrganizationId = &id
	}
	if result.DefaultOrganizationID > 0 {
		id := result.DefaultOrganizationID
		resp.DefaultOrganizationId = &id
	}
	writeJSON(w, http.StatusOK, resp)
}

// CreateOrganization serves POST /api/v1/orgs.
func (s *Server) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	user, ok := appmiddleware.UserFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthenticated", "Authentification requise.")
		return
	}

	var req CreateOrganizationRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}

	slug := ""
	if req.Slug != nil {
		slug = *req.Slug
	}

	token := appmiddleware.SessionTokenFromContext(r)
	result, err := s.Orgs.Create(r.Context(), user, token, req.Name, slug, w)
	if err != nil {
		writeOrgError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, OrganizationActionResponse{
		Organization: mapOrganization(r.Context(), s.Store, user, result.Organization, result.Role, true),
		Redirect:     result.Redirect,
	})
}

// SelectActiveOrganization serves POST /api/v1/orgs/active.
func (s *Server) SelectActiveOrganization(w http.ResponseWriter, r *http.Request) {
	user, ok := appmiddleware.UserFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthenticated", "Authentification requise.")
		return
	}

	var req SelectOrganizationRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}

	token := appmiddleware.SessionTokenFromContext(r)
	result, err := s.Orgs.Select(r.Context(), user, token, req.OrganizationId, w)
	if err != nil {
		writeOrgError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, OrganizationActionResponse{
		Organization: mapOrganization(r.Context(), s.Store, user, result.Organization, result.Role, true),
		Redirect:     result.Redirect,
	})
}

// AcceptOrganizationInvitation serves POST /api/v1/orgs/invitations/{invitationID}/accept.
func (s *Server) AcceptOrganizationInvitation(w http.ResponseWriter, r *http.Request, invitationID int64) {
	user, ok := appmiddleware.UserFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthenticated", "Authentification requise.")
		return
	}

	token := appmiddleware.SessionTokenFromContext(r)
	result, err := s.Orgs.AcceptInvitation(r.Context(), user, token, invitationID, w)
	if err != nil {
		writeOrgError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, OrganizationActionResponse{
		Organization: mapOrganization(r.Context(), s.Store, user, result.Organization, result.Role, true),
		Redirect:     result.Redirect,
	})
}

func writeOrgError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, organizations.ErrUnauthenticated):
		writeAPIError(w, http.StatusUnauthorized, "unauthenticated", "Authentification requise.")
	case errors.Is(err, organizations.ErrAlreadyHasOrganization):
		writeAPIError(w, http.StatusConflict, "already_has_organization",
			"Vous appartenez déjà à une organisation.")
	case errors.Is(err, organizations.ErrNotFound):
		writeAPIError(w, http.StatusNotFound, "not_found", "Organisation introuvable.")
	case errors.Is(err, organizations.ErrForbidden):
		writeAPIError(w, http.StatusForbidden, "forbidden", "Action non autorisée.")
	case errors.Is(err, organizations.ErrValidation):
		writeAPIError(w, http.StatusBadRequest, "validation_failed", orgValidationMessage(err))
	default:
		slog.Error("organization action", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
	}
}

func orgValidationMessage(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "name required"):
		return "Le nom est obligatoire."
	case strings.Contains(msg, "invalid slug"):
		return "Identifiant invalide (lettres minuscules, chiffres et tirets uniquement)."
	case strings.Contains(msg, "slug taken"):
		return "Cet identifiant est déjà utilisé."
	case strings.Contains(msg, "organization_id required"):
		return "Choisissez une organisation."
	default:
		return "Requête invalide."
	}
}

func mapMemberships(ctx context.Context, st *store.Store, user *store.User, ms []store.OrganizationMembership, activeOrgID int64) []Organization {
	out := make([]Organization, 0, len(ms))
	for _, m := range ms {
		out = append(out, mapOrganization(ctx, st, user, &m.Organization, m.Role, m.Organization.ID == activeOrgID))
	}
	return out
}

func mapInvitations(invites []store.OrganizationInvitation) []OrganizationInvitation {
	out := make([]OrganizationInvitation, 0, len(invites))
	for _, inv := range invites {
		out = append(out, OrganizationInvitation{
			Id:               inv.ID,
			OrganizationId:   inv.OrganizationID,
			OrganizationName: inv.OrganizationName,
			OrgRole:          inv.OrgRole,
			CreatedAt:        inv.CreatedAt,
		})
	}
	return out
}

func mapOrganization(ctx context.Context, st *store.Store, user *store.User, org *store.Organization, role string, includeSubjectCount bool) Organization {
	o := Organization{
		Id:          org.ID,
		Name:        org.Name,
		Slug:        org.Slug,
		Role:        role,
		MemberCount: 0,
	}
	if org.UISubjectLabel != "" {
		label := org.UISubjectLabel
		o.UiSubjectLabel = &label
	}
	if org.UIRunLabel != "" {
		label := org.UIRunLabel
		o.UiRunLabel = &label
	}
	if st != nil {
		if n, err := st.CountOrganizationMembers(ctx, org.ID); err == nil {
			o.MemberCount = n
		}
		if includeSubjectCount && user != nil {
			admin := user.Role == auth.RoleAdmin
			if subjects, err := st.ListSubjects(ctx, user.ID, admin, ""); err == nil {
				n := len(subjects)
				o.VisibleSubjectCount = &n
			}
		}
	}
	return o
}
