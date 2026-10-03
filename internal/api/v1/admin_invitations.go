package apiv1

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

// ListAdminInvitations serves GET /api/v1/admin/invitations.
func (s *Server) ListAdminInvitations(w http.ResponseWriter, r *http.Request) {
	_, org, ok := s.requireOrgAdmin(w, r)
	if !ok {
		return
	}
	invites, err := s.Store.ListPendingInvitationsByOrganization(r.Context(), org.ID)
	if err != nil {
		slog.Error("list admin invitations", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	out := make([]AdminInvitation, 0, len(invites))
	for _, inv := range invites {
		out = append(out, mapAdminInvitation(inv, false))
	}
	writeJSON(w, http.StatusOK, AdminInvitationListResponse{Invitations: out})
}

// CreateAdminInvitation serves POST /api/v1/admin/invitations.
func (s *Server) CreateAdminInvitation(w http.ResponseWriter, r *http.Request) {
	user, org, ok := s.requireOrgAdmin(w, r)
	if !ok {
		return
	}

	var req AdminInvitationCreateRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}
	email, err := normalizeEmail(string(req.Email))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Adresse email invalide.")
		return
	}

	orgRole := store.OrgRoleMember
	if req.OrgRole != nil && strings.TrimSpace(string(*req.OrgRole)) != "" {
		orgRole = string(*req.OrgRole)
	}
	if !validOrgRole(orgRole) {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Rôle organisation invalide.")
		return
	}
	if orgRole == store.OrgRoleOwner {
		callerRole, _, roleErr := s.Store.OrganizationMemberRole(r.Context(), org.ID, user.ID)
		if roleErr != nil {
			slog.Error("invitation caller role", "err", roleErr)
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
			return
		}
		if user.Role != auth.RoleAdmin && callerRole != store.OrgRoleOwner {
			writeAPIError(w, http.StatusForbidden, "forbidden",
				"Seul un propriétaire peut inviter avec le rôle propriétaire.")
			return
		}
	}

	if err := s.Store.CreateOrganizationInvitation(r.Context(), email, org.ID, orgRole); err != nil {
		if errors.Is(err, store.ErrAlreadyOrganizationMember) {
			writeAPIError(w, http.StatusConflict, "conflict", "Cette personne est déjà membre de l'organisation.")
			return
		}
		slog.Error("create organization invitation", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	invites, err := s.Store.ListPendingInvitationsByOrganization(r.Context(), org.ID)
	if err != nil {
		slog.Error("reload invitations", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	var created store.OrganizationInvitation
	for _, inv := range invites {
		if inv.Email == email {
			created = inv
			break
		}
	}
	if created.ID == 0 {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	emailQueued := s.enqueueOrgInvitationEmail(r, org, email)
	writeJSON(w, http.StatusCreated, mapAdminInvitation(created, emailQueued))
}

// DeleteAdminInvitation serves DELETE /api/v1/admin/invitations/{invitationId}.
func (s *Server) DeleteAdminInvitation(w http.ResponseWriter, r *http.Request, invitationId int64) {
	_, org, ok := s.requireOrgAdmin(w, r)
	if !ok {
		return
	}
	inv, err := s.Store.OrganizationInvitationByID(r.Context(), invitationId)
	if err != nil {
		if errors.Is(err, store.ErrOrganizationInvitationNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Invitation introuvable.")
			return
		}
		slog.Error("load invitation", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if inv.OrganizationID != org.ID {
		writeAPIError(w, http.StatusNotFound, "not_found", "Invitation introuvable.")
		return
	}
	if err := s.Store.DeleteOrganizationInvitation(r.Context(), invitationId); err != nil {
		if errors.Is(err, store.ErrOrganizationInvitationNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Invitation introuvable.")
			return
		}
		slog.Error("delete invitation", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) enqueueOrgInvitationEmail(r *http.Request, org *store.Organization, email string) bool {
	if s == nil || s.Store == nil || org == nil {
		return false
	}
	subject := fmt.Sprintf("Invitation à rejoindre %s — Revues", org.Name)
	body := fmt.Sprintf(
		"Vous êtes invité à rejoindre l'organisation « %s » sur Revues.\n\n"+
			"Connectez-vous avec cet email (%s), puis ouvrez /org/select pour accepter l'invitation.\n",
		org.Name, email,
	)
	now := time.Now().UTC()
	_, err := s.Store.EnqueueEmailDelivery(
		r.Context(), org.ID, email, subject, body, now, now.Add(24*time.Hour),
	)
	if err != nil {
		slog.Error("enqueue org invitation email", "err", err, "to", email)
		return false
	}
	return true
}

func mapAdminInvitation(inv store.OrganizationInvitation, emailQueued bool) AdminInvitation {
	return AdminInvitation{
		Id:          inv.ID,
		Email:       openapi_types.Email(inv.Email),
		OrgRole:     AdminInvitationOrgRole(inv.OrgRole),
		CreatedAt:   inv.CreatedAt,
		EmailQueued: emailQueued,
	}
}
