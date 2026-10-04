package apiv1

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/jeb-maker/revues/internal/features/subjects"
	"github.com/jeb-maker/revues/internal/store"
	appmiddleware "github.com/jeb-maker/revues/internal/web/middleware"
)

// ListSubjects serves GET /api/v1/subjects.
func (s *Server) ListSubjects(w http.ResponseWriter, r *http.Request, params ListSubjectsParams) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}
	if !requireOrg(w, r) {
		return
	}

	q := ""
	if params.Q != nil {
		q = strings.TrimSpace(*params.Q)
	}
	orgRole, orgMember := s.orgMembership(r, user.ID)
	orgAdmin := orgMember && (orgRole == store.OrgRoleOwner || orgRole == store.OrgRoleAdmin)
	items, err := s.Store.ListSubjects(r.Context(), user.ID, orgAdmin, q)
	if err != nil {
		slog.Error("list subjects", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	out := make([]SubjectSummary, 0, len(items))
	for _, item := range items {
		out = append(out, mapSubjectSummary(item))
	}
	writeJSON(w, http.StatusOK, SubjectListResponse{
		Subjects:  out,
		CanCreate: subjects.CanCreateSubject(user, orgMember),
	})
}

// CreateSubject serves POST /api/v1/subjects.
func (s *Server) CreateSubject(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}
	if !requireOrg(w, r) {
		return
	}
	orgRole, orgMember := s.orgMembership(r, user.ID)
	if !subjects.CanCreateSubject(user, orgMember) {
		writeAPIError(w, http.StatusForbidden, "forbidden", "Droits insuffisants.")
		return
	}

	var req SubjectWriteRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Le nom est obligatoire.")
		return
	}

	canSetVisibility := subjects.CanSetSubjectVisibility(user, orgRole, orgMember, store.SubjectAccess{})
	visibility := store.SubjectVisibilityNormal
	if canSetVisibility && req.Visibility != nil {
		norm, err := store.NormalizeSubjectVisibility(string(*req.Visibility))
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", "Visibilité invalide.")
			return
		}
		visibility = norm
	}

	description := ""
	if req.Description != nil {
		description = strings.TrimSpace(*req.Description)
	}
	domains := normalizeOptionalTags(req.Domains)
	tags := normalizeOptionalTags(req.Tags)

	subject, err := s.Store.CreateSubjectWithVisibility(r.Context(), name, description, user.ID, domains, visibility)
	if err != nil {
		slog.Error("create subject", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if err := s.Store.SetSubjectTags(r.Context(), subject.ID, tags); err != nil {
		slog.Error("set subject tags", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	detail, ok := s.buildSubjectDetail(w, r, subject, user)
	if !ok {
		return
	}
	writeJSON(w, http.StatusCreated, detail)
}

// GetSubject serves GET /api/v1/subjects/{subjectID}.
func (s *Server) GetSubject(w http.ResponseWriter, r *http.Request, subjectID SubjectId) {
	subject, user, _, ok := s.ensureSubjectAccess(w, r, subjectID)
	if !ok {
		return
	}
	detail, ok := s.buildSubjectDetail(w, r, subject, user)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// UpdateSubject serves PATCH /api/v1/subjects/{subjectID}.
func (s *Server) UpdateSubject(w http.ResponseWriter, r *http.Request, subjectID SubjectId) {
	subject, user, access, ok := s.ensureSubjectAccess(w, r, subjectID)
	if !ok {
		return
	}
	if !subjects.CanManageAccess(user, access) {
		writeAPIError(w, http.StatusNotFound, "not_found", "Sujet introuvable.")
		return
	}

	var req SubjectWriteRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Le nom est obligatoire.")
		return
	}

	orgRole, orgMember := s.orgMembership(r, user.ID)
	canSetVisibility := subjects.CanSetSubjectVisibility(user, orgRole, orgMember, access)
	visibility := subject.Visibility
	if canSetVisibility && req.Visibility != nil {
		norm, err := store.NormalizeSubjectVisibility(string(*req.Visibility))
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", "Visibilité invalide.")
			return
		}
		visibility = norm
	}

	description := ""
	if req.Description != nil {
		description = strings.TrimSpace(*req.Description)
	}
	domains := normalizeOptionalTags(req.Domains)
	tags := normalizeOptionalTags(req.Tags)

	if err := s.Store.UpdateSubjectWithVisibility(r.Context(), subject.ID, name, description, domains, visibility); err != nil {
		if errors.Is(err, store.ErrSubjectNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Sujet introuvable.")
			return
		}
		slog.Error("update subject", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if visibility == store.SubjectVisibilityPrivate && !access.IsSupervisor() {
		if err := s.Store.UpsertDirectSubjectMember(r.Context(), subject.ID, user.ID, store.SubjectRoleLead); err != nil {
			slog.Error("ensure lead on private subject", "err", err)
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
			return
		}
	}
	if err := s.Store.SetSubjectTags(r.Context(), subject.ID, tags); err != nil {
		slog.Error("set subject tags", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	updated, err := s.Store.SubjectByID(r.Context(), subject.ID)
	if err != nil {
		slog.Error("reload subject", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	detail, ok := s.buildSubjectDetail(w, r, updated, user)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// ArchiveSubject serves POST /api/v1/subjects/{subjectID}/archive.
func (s *Server) ArchiveSubject(w http.ResponseWriter, r *http.Request, subjectID SubjectId) {
	_, user, access, ok := s.ensureSubjectAccess(w, r, subjectID)
	if !ok {
		return
	}
	if !subjects.CanManageAccess(user, access) {
		writeAPIError(w, http.StatusNotFound, "not_found", "Sujet introuvable.")
		return
	}
	if err := s.Store.ArchiveSubject(r.Context(), subjectID); err != nil {
		if errors.Is(err, store.ErrSubjectNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Sujet introuvable.")
			return
		}
		slog.Error("archive subject", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListSubjectMembers serves GET /api/v1/subjects/{subjectID}/members.
func (s *Server) ListSubjectMembers(w http.ResponseWriter, r *http.Request, subjectID SubjectId) {
	if _, _, _, ok := s.ensureSubjectAccess(w, r, subjectID); !ok {
		return
	}
	members, err := s.Store.ListDirectSubjectMembers(r.Context(), subjectID)
	if err != nil {
		slog.Error("list subject members", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	writeJSON(w, http.StatusOK, SubjectMemberListResponse{Members: mapMembers(members)})
}

// AddSubjectMember serves POST /api/v1/subjects/{subjectID}/members.
func (s *Server) AddSubjectMember(w http.ResponseWriter, r *http.Request, subjectID SubjectId) {
	subject, user, access, ok := s.ensureSubjectAccess(w, r, subjectID)
	if !ok {
		return
	}
	org, _ := appmiddleware.OrganizationFromContext(r.Context())
	policies := subjects.PoliciesFromOrganization(org)
	if !subjects.CanManageSubjectMembers(user, access, policies) {
		if subjects.CanLeadAccess(user, access) {
			writeAPIError(w, http.StatusForbidden, "forbidden",
				"La politique de l'organisation n'autorise pas les responsables à inviter des membres.")
			return
		}
		writeAPIError(w, http.StatusNotFound, "not_found", "Sujet introuvable.")
		return
	}

	var req AddSubjectMemberRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}
	email := strings.TrimSpace(string(req.Email))
	if email == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Email requis.")
		return
	}
	role := store.SubjectRoleViewer
	if req.Role != nil && string(*req.Role) != "" {
		role = string(*req.Role)
	}

	invitee, err := s.Store.UserByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			writeAPIError(w, http.StatusBadRequest, "validation_failed",
				"Aucun compte avec cet email. L'invité doit d'abord se connecter.")
			return
		}
		slog.Error("user by email for subject member", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	inviteeIsOrgMember := false
	if org != nil {
		_, inviteeIsOrgMember, _ = s.Store.OrganizationMemberRole(r.Context(), org.ID, invitee.ID)
	}
	if !subjects.CanInviteSubjectMember(user, access, policies, inviteeIsOrgMember) {
		if inviteeIsOrgMember {
			writeAPIError(w, http.StatusForbidden, "forbidden",
				"La politique de l'organisation n'autorise pas les responsables à inviter des membres.")
		} else {
			writeAPIError(w, http.StatusForbidden, "forbidden",
				"La politique de l'organisation n'autorise pas les responsables à inviter des externes.")
		}
		return
	}

	if err = s.Store.UpsertDirectSubjectMember(r.Context(), subject.ID, invitee.ID, role); err != nil {
		if errors.Is(err, store.ErrInvalidSubjectRole) {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", "Rôle invalide.")
			return
		}
		slog.Error("upsert direct subject member", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	// First grant on a previously ungated subject ends org_member_legacy access.
	// Keep the inviter able to manage the subject.
	afterAccess, err := s.Store.ResolveSubjectAccess(r.Context(), user.ID, subject.ID, user.Role)
	if err != nil {
		slog.Error("resolve access after member add", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if !subjects.CanManageSubjectMembers(user, afterAccess, policies) && invitee.ID != user.ID {
		if err := s.Store.UpsertDirectSubjectMember(r.Context(), subject.ID, user.ID, store.SubjectRoleLead); err != nil {
			slog.Error("ensure inviter lead after gating", "err", err)
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
			return
		}
	}

	writeJSON(w, http.StatusCreated, SubjectMember{
		UserId:      invitee.ID,
		Login:       invitee.Login,
		Email:       openapi_types.Email(invitee.Email),
		DisplayName: invitee.DisplayName,
		Role:        SubjectMemberRole(role),
	})
}

// RemoveSubjectMember serves DELETE /api/v1/subjects/{subjectID}/members/{userID}.
func (s *Server) RemoveSubjectMember(w http.ResponseWriter, r *http.Request, subjectID SubjectId, userID int64) {
	_, user, access, ok := s.ensureSubjectAccess(w, r, subjectID)
	if !ok {
		return
	}
	org, _ := appmiddleware.OrganizationFromContext(r.Context())
	policies := subjects.PoliciesFromOrganization(org)
	if !subjects.CanManageSubjectMembers(user, access, policies) {
		if subjects.CanLeadAccess(user, access) {
			writeAPIError(w, http.StatusForbidden, "forbidden",
				"La politique de l'organisation n'autorise pas les responsables à gérer les membres.")
			return
		}
		writeAPIError(w, http.StatusNotFound, "not_found", "Sujet introuvable.")
		return
	}
	if userID <= 0 {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Membre invalide.")
		return
	}
	if err := s.Store.RemoveDirectSubjectMember(r.Context(), subjectID, userID); err != nil {
		if errors.Is(err, store.ErrDirectSubjectMemberNotFound) || errors.Is(err, store.ErrSubjectNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Ce membre n'est pas affecté à ce sujet.")
			return
		}
		slog.Error("remove direct subject member", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) buildSubjectDetail(w http.ResponseWriter, r *http.Request, subject *store.Subject, user *store.User) (SubjectDetail, bool) {
	access, err := s.Store.ResolveSubjectAccess(r.Context(), user.ID, subject.ID, user.Role)
	if err != nil {
		slog.Error("resolve subject access", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return SubjectDetail{}, false
	}
	domains, err := s.Store.ListSubjectDomains(r.Context(), subject.ID)
	if err != nil {
		slog.Error("list subject domains", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return SubjectDetail{}, false
	}
	tags, err := s.Store.ListSubjectTags(r.Context(), subject.ID)
	if err != nil {
		slog.Error("list subject tags", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return SubjectDetail{}, false
	}
	members, err := s.Store.ListDirectSubjectMembers(r.Context(), subject.ID)
	if err != nil {
		slog.Error("list subject members", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return SubjectDetail{}, false
	}
	if domains == nil {
		domains = []string{}
	}
	if tags == nil {
		tags = []string{}
	}

	orgRole, orgMember := s.orgMembership(r, user.ID)
	org, _ := appmiddleware.OrganizationFromContext(r.Context())
	policies := subjects.PoliciesFromOrganization(org)

	sources := access.Sources
	if sources == nil {
		sources = []string{}
	}

	return SubjectDetail{
		Id:          subject.ID,
		Name:        subject.Name,
		Description: subject.Description,
		Visibility:  SubjectDetailVisibility(subject.Visibility),
		Domains:     domains,
		Tags:        tags,
		Members:     mapMembers(members),
		Access: SubjectAccessInfo{
			Role:    access.Role,
			Sources: sources,
		},
		Capabilities: SubjectCapabilities{
			CanManage:        subjects.CanManageAccess(user, access),
			CanManageMembers: subjects.CanManageSubjectMembers(user, access, policies),
			CanSetVisibility: subjects.CanSetSubjectVisibility(user, orgRole, orgMember, access),
			CanLaunch:        subjects.CanContributeAccess(user, access),
		},
	}, true
}

func (s *Server) orgMembership(r *http.Request, userID int64) (orgRole string, orgMember bool) {
	org, ok := appmiddleware.OrganizationFromContext(r.Context())
	if !ok {
		return "", false
	}
	orgRole, orgMember, _ = s.Store.OrganizationMemberRole(r.Context(), org.ID, userID)
	return orgRole, orgMember
}

func mapSubjectSummary(sub store.Subject) SubjectSummary {
	return SubjectSummary{
		Id:          sub.ID,
		Name:        sub.Name,
		Description: sub.Description,
		Visibility:  SubjectSummaryVisibility(sub.Visibility),
	}
}

func mapMembers(members []store.DirectSubjectMember) []SubjectMember {
	out := make([]SubjectMember, 0, len(members))
	for _, m := range members {
		out = append(out, SubjectMember{
			UserId:      m.UserID,
			Login:       m.Login,
			Email:       openapi_types.Email(m.Email),
			DisplayName: m.DisplayName,
			Role:        SubjectMemberRole(m.Role),
		})
	}
	return out
}

func normalizeOptionalTags(tags *[]string) []string {
	if tags == nil {
		return nil
	}
	return store.NormalizeTags(*tags)
}

func userFromRequest(r *http.Request) (*store.User, bool) {
	return appmiddleware.UserFromContext(r.Context())
}

func orgFromRequest(r *http.Request) (*store.Organization, bool) {
	return appmiddleware.OrganizationFromContext(r.Context())
}
