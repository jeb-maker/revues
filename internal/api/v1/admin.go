package apiv1

import (
	"errors"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/jeb-maker/revues/internal/store"
)

// ListAllowedEmails serves GET /api/v1/admin/allowed-emails.
func (s *Server) ListAllowedEmails(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	emails, err := s.Store.ListAllowedEmails(r.Context())
	if err != nil {
		slog.Error("list allowed emails", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	out := make([]AllowedEmail, 0, len(emails))
	for _, e := range emails {
		out = append(out, mapAllowedEmail(e))
	}
	writeJSON(w, http.StatusOK, AllowedEmailListResponse{Emails: out})
}

// CreateAllowedEmail serves POST /api/v1/admin/allowed-emails.
func (s *Server) CreateAllowedEmail(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}

	var req AllowedEmailWriteRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}

	email, err := normalizeEmail(string(req.Email))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Adresse email invalide.")
		return
	}
	role := string(req.Role)
	if !store.ValidWhitelistRole(role) {
		writeAPIError(w, http.StatusBadRequest, "validation_failed",
			"Rôle invalide. La whitelist n'accepte que lecteur ou éditeur ; l'admin global passe par REVUES_BOOTSTRAP_ADMIN_EMAIL.")
		return
	}

	if err := s.Store.InsertAllowedEmail(r.Context(), email, role); err != nil {
		if errors.Is(err, store.ErrInvalidAllowedRole) {
			writeAPIError(w, http.StatusBadRequest, "validation_failed",
				"Rôle invalide. La whitelist n'accepte que lecteur ou éditeur.")
			return
		}
		slog.Error("insert allowed email", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	created := AllowedEmail{
		Email:     openapi_types.Email(email),
		Role:      AllowedEmailRole(role),
		CreatedAt: "",
	}
	if emails, listErr := s.Store.ListAllowedEmails(r.Context()); listErr == nil {
		for _, e := range emails {
			if e.Email == email {
				created = mapAllowedEmail(e)
				break
			}
		}
	}
	writeJSON(w, http.StatusCreated, created)
}

// DeleteAllowedEmail serves DELETE /api/v1/admin/allowed-emails/{email}.
func (s *Server) DeleteAllowedEmail(w http.ResponseWriter, r *http.Request, emailParam openapi_types.Email) {
	user, _, ok := s.requireOrgAdmin(w, r)
	if !ok {
		return
	}

	email, err := normalizeEmail(string(emailParam))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Adresse email invalide.")
		return
	}
	if strings.EqualFold(user.Email, email) {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Vous ne pouvez pas retirer votre propre email.")
		return
	}

	if err := s.Store.DeleteAllowedEmail(r.Context(), email); err != nil {
		if errors.Is(err, store.ErrAllowedEmailNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Email introuvable.")
			return
		}
		slog.Error("delete allowed email", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListOrganizationMembers serves GET /api/v1/admin/members.
func (s *Server) ListOrganizationMembers(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	members, err := s.Store.ListOrganizationMembers(r.Context())
	if err != nil {
		slog.Error("list organization members", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	writeJSON(w, http.StatusOK, OrganizationMemberListResponse{
		Members: mapOrgMembers(members),
	})
}

// UpdateOrganizationMemberRole serves PATCH /api/v1/admin/members/{userId}.
//
//nolint:staticcheck // SA1003: userId matches OpenAPI path param / oapi-codegen ServerInterface
func (s *Server) UpdateOrganizationMemberRole(w http.ResponseWriter, r *http.Request, userId int64) {
	_, org, ok := s.requireOrgAdmin(w, r)
	if !ok {
		return
	}

	var req UpdateOrganizationMemberRoleRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}
	role := string(req.Role)
	if !validOrgRole(role) {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Rôle organisation invalide.")
		return
	}

	current, err := s.Store.OrganizationMemberUserByID(r.Context(), userId)
	if err != nil {
		if errors.Is(err, store.ErrOrganizationMemberNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Membre introuvable.")
			return
		}
		slog.Error("organization member by id", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	if current.Role == store.OrgRoleOwner && role != store.OrgRoleOwner {
		owners, countErr := s.Store.CountOrganizationMembersWithRole(r.Context(), org.ID, store.OrgRoleOwner)
		if countErr != nil {
			slog.Error("count organization owners", "err", countErr)
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
			return
		}
		if owners <= 1 {
			writeAPIError(w, http.StatusConflict, "last_owner",
				"Impossible de rétrograder le dernier propriétaire de l'organisation.")
			return
		}
	}

	if addErr := s.Store.AddOrganizationMember(r.Context(), org.ID, userId, role); addErr != nil {
		slog.Error("update organization member role", "err", addErr)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	updated, err := s.Store.OrganizationMemberUserByID(r.Context(), userId)
	if err != nil {
		slog.Error("reload organization member", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	writeJSON(w, http.StatusOK, mapOrgMember(*updated))
}

// ListAdminTeams serves GET /api/v1/admin/teams.
func (s *Server) ListAdminTeams(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	teams, err := s.Store.ListOrganizationTeams(r.Context())
	if err != nil {
		slog.Error("list organization teams", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	out := make([]AdminTeam, 0, len(teams))
	for _, t := range teams {
		out = append(out, mapAdminTeam(t))
	}
	writeJSON(w, http.StatusOK, AdminTeamListResponse{Teams: out})
}

// CreateAdminTeam serves POST /api/v1/admin/teams.
func (s *Server) CreateAdminTeam(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}

	var req CreateAdminTeamRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Le nom de l'équipe est requis.")
		return
	}
	slug := name
	if req.Slug != nil && strings.TrimSpace(*req.Slug) != "" {
		slug = strings.TrimSpace(*req.Slug)
	}
	description := ""
	if req.Description != nil {
		description = strings.TrimSpace(*req.Description)
	}

	team, err := s.Store.CreateTeam(r.Context(), name, slug, description)
	if err != nil {
		if errors.Is(err, store.ErrInvalidOrganizationSlug) {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", "Slug invalide (lettres, chiffres et tirets).")
			return
		}
		if errors.Is(err, store.ErrTeamSlugTaken) {
			writeAPIError(w, http.StatusConflict, "slug_taken", "Ce slug d'équipe est déjà utilisé.")
			return
		}
		slog.Error("create team", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	writeJSON(w, http.StatusCreated, mapAdminTeam(*team))
}

// GetAdminTeam serves GET /api/v1/admin/teams/{teamId}.
//
//nolint:staticcheck // SA1003: teamId matches OpenAPI path param / oapi-codegen ServerInterface
func (s *Server) GetAdminTeam(w http.ResponseWriter, r *http.Request, teamId int64) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}

	team, err := s.Store.TeamByID(r.Context(), teamId)
	if err != nil {
		if errors.Is(err, store.ErrTeamNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Équipe introuvable.")
			return
		}
		slog.Error("team by id", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	members, err := s.Store.ListTeamMembers(r.Context(), teamId)
	if err != nil {
		slog.Error("list team members", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	orgMembers, err := s.Store.ListOrganizationMembers(r.Context())
	if err != nil {
		slog.Error("list organization members", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	onTeam := make(map[int64]struct{}, len(members))
	for _, m := range members {
		onTeam[m.UserID] = struct{}{}
	}
	candidates := make([]store.OrganizationMemberUser, 0)
	for _, m := range orgMembers {
		if _, exists := onTeam[m.UserID]; !exists {
			candidates = append(candidates, m)
		}
	}

	mapped := mapAdminTeam(*team)
	mapped.MemberCount = len(members)
	writeJSON(w, http.StatusOK, AdminTeamDetailResponse{
		Team:       mapped,
		Members:    mapTeamMembers(members),
		Candidates: mapOrgMembers(candidates),
	})
}

// AddAdminTeamMember serves POST /api/v1/admin/teams/{teamId}/members.
//
//nolint:staticcheck // SA1003: teamId matches OpenAPI path param / oapi-codegen ServerInterface
func (s *Server) AddAdminTeamMember(w http.ResponseWriter, r *http.Request, teamId int64) {
	_, org, ok := s.requireOrgAdmin(w, r)
	if !ok {
		return
	}

	var req AdminTeamMemberRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}
	if req.UserId <= 0 {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Membre invalide.")
		return
	}

	_, isMember, err := s.Store.OrganizationMemberRole(r.Context(), org.ID, req.UserId)
	if err != nil {
		slog.Error("organization member role", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if !isMember {
		writeAPIError(w, http.StatusBadRequest, "validation_failed",
			"L'utilisateur n'est pas membre de l'organisation.")
		return
	}

	if err := s.Store.AddTeamMember(r.Context(), teamId, req.UserId); err != nil {
		if errors.Is(err, store.ErrTeamNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Équipe introuvable.")
			return
		}
		slog.Error("add team member", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RemoveAdminTeamMember serves DELETE /api/v1/admin/teams/{teamId}/members/{userId}.
//
//nolint:staticcheck // SA1003: teamId/userId match OpenAPI path params / oapi-codegen ServerInterface
func (s *Server) RemoveAdminTeamMember(w http.ResponseWriter, r *http.Request, teamId int64, userId int64) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}

	if err := s.Store.RemoveTeamMember(r.Context(), teamId, userId); err != nil {
		if errors.Is(err, store.ErrTeamNotFound) || errors.Is(err, store.ErrTeamMemberNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Membre ou équipe introuvable.")
			return
		}
		slog.Error("remove team member", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetLeadPolicies serves GET /api/v1/admin/settings/policies.
func (s *Server) GetLeadPolicies(w http.ResponseWriter, r *http.Request) {
	_, org, ok := s.requireOrgAdmin(w, r)
	if !ok {
		return
	}
	refreshed, err := s.Store.OrganizationByID(r.Context(), org.ID)
	if err != nil {
		slog.Error("organization by id for policies", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	writeJSON(w, http.StatusOK, mapLeadPolicies(refreshed.LeadPolicies()))
}

// UpdateLeadPolicies serves PUT /api/v1/admin/settings/policies.
func (s *Server) UpdateLeadPolicies(w http.ResponseWriter, r *http.Request) {
	_, org, ok := s.requireOrgAdmin(w, r)
	if !ok {
		return
	}

	var req LeadPolicies
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}
	policies := store.OrgLeadPolicies{
		LeadsMayAssignTeams:     req.LeadsMayAssignTeams,
		LeadsMayInviteMembers:   req.LeadsMayInviteMembers,
		LeadsMayInviteExternals: req.LeadsMayInviteExternals,
	}
	if err := s.Store.UpdateOrganizationLeadPolicies(r.Context(), org.ID, policies); err != nil {
		slog.Error("update organization lead policies", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	writeJSON(w, http.StatusOK, mapLeadPolicies(policies))
}

func mapAllowedEmail(e store.AllowedEmail) AllowedEmail {
	return AllowedEmail{
		Email:     openapi_types.Email(e.Email),
		Role:      AllowedEmailRole(e.Role),
		CreatedAt: e.CreatedAt,
	}
}

func mapOrgMembers(members []store.OrganizationMemberUser) []OrganizationMember {
	out := make([]OrganizationMember, 0, len(members))
	for _, m := range members {
		out = append(out, mapOrgMember(m))
	}
	return out
}

func mapOrgMember(m store.OrganizationMemberUser) OrganizationMember {
	return OrganizationMember{
		UserId:      m.UserID,
		Login:       m.Login,
		Email:       openapi_types.Email(m.Email),
		DisplayName: m.DisplayName,
		Role:        OrganizationMemberRole(m.Role),
		JoinedAt:    m.JoinedAt,
	}
}

func mapAdminTeam(t store.OrganizationTeam) AdminTeam {
	out := AdminTeam{
		Id:          t.ID,
		Name:        t.Name,
		Slug:        t.Slug,
		CreatedAt:   t.CreatedAt,
		MemberCount: t.MemberCount,
	}
	if t.Description != "" {
		desc := t.Description
		out.Description = &desc
	}
	return out
}

func mapTeamMembers(members []store.TeamMember) []AdminTeamMember {
	out := make([]AdminTeamMember, 0, len(members))
	for _, m := range members {
		out = append(out, AdminTeamMember{
			UserId:      m.UserID,
			Login:       m.Login,
			Email:       openapi_types.Email(m.Email),
			DisplayName: m.DisplayName,
			CreatedAt:   m.CreatedAt,
		})
	}
	return out
}

func mapLeadPolicies(p store.OrgLeadPolicies) LeadPolicies {
	return LeadPolicies{
		LeadsMayAssignTeams:     p.LeadsMayAssignTeams,
		LeadsMayInviteMembers:   p.LeadsMayInviteMembers,
		LeadsMayInviteExternals: p.LeadsMayInviteExternals,
	}
}

func validOrgRole(role string) bool {
	switch role {
	case store.OrgRoleOwner, store.OrgRoleAdmin, store.OrgRoleMember:
		return true
	default:
		return false
	}
}

func normalizeEmail(raw string) (string, error) {
	email := strings.TrimSpace(strings.ToLower(raw))
	if email == "" {
		return "", errors.New("empty email")
	}
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return "", err
	}
	return strings.ToLower(addr.Address), nil
}
