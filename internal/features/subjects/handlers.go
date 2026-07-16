package subjects

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/middleware"
	"github.com/jeb-maker/revues/internal/web/templates"
)

// Deps holds shared dependencies for the subjects HTTP handlers.
type Deps struct {
	Templates     *template.Template
	Store         SubjectStore
	SessionSecret string
}

// PageData builds shared view data with user and CSRF from the request context.
func (d *Deps) PageData(r *http.Request, title string) templates.PageData {
	data := templates.PageData{Title: title}
	if user, ok := middleware.UserFromContext(r.Context()); ok {
		data.User = user
		if token := middleware.SessionTokenFromContext(r); token != "" {
			data.CSRFToken = auth.CSRFToken(token, d.SessionSecret)
		}
	}
	templates.ApplyHeaderFromContext(r, &data)
	return data
}

// PageDataTab is PageData with ActiveTab set.
func (d *Deps) PageDataTab(r *http.Request, title, activeTab string) templates.PageData {
	data := d.PageData(r, title)
	data.ActiveTab = activeTab
	return data
}

func adminSubjectsRoute(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/admin/subjects")
}

func subjectsListPath(r *http.Request) string {
	if adminSubjectsRoute(r) {
		return templates.PathAdminSubjects
	}
	return templates.PathSubjects
}

func (d *Deps) subjectsPageData(r *http.Request, title string) templates.PageData {
	if adminSubjectsRoute(r) {
		data := d.PageDataTab(r, title, "org")
		data.AdminSection = "subjects"
		return data
	}
	return d.PageDataTab(r, title, "subjects")
}

// Subjects handles subject CRUD and the run wizard subject step.
type Subjects struct {
	Deps
}

// List shows subjects visible to the current user.
func (h *Subjects) List(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	admin := auth.HasMinRole(user.Role, auth.RoleAdmin)
	filterQuery := strings.TrimSpace(r.URL.Query().Get("q"))
	items, err := h.Store.ListSubjects(r.Context(), user.ID, admin, filterQuery)
	if err != nil {
		slog.Error("list subjects", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	orgRole, orgMember, _ := h.Store.OrganizationMemberRole(r.Context(), 0, user.ID)
	if org, ok := middleware.OrganizationFromContext(r.Context()); ok {
		orgRole, orgMember, _ = h.Store.OrganizationMemberRole(r.Context(), org.ID, user.ID)
	}

	pd := h.subjectsPageData(r, templates.DefaultUILabels().Subject.Plural)
	data := templates.SubjectsListData{
		PageData:          pd,
		Subjects:          items,
		FilterQuery:       filterQuery,
		HasActiveFilters:  filterQuery != "",
		CanCreate:         CanCreateSubject(user),
		CanManageOrgUsers: CanManageOrgUsers(user, orgRole, orgMember),
		Message:           r.URL.Query().Get("msg"),
	}
	if adminSubjectsRoute(r) {
		data.Breadcrumbs = templates.BCAdminSubjects(data.Labels.Subject)
	} else {
		data.Breadcrumbs = templates.BCSubjects(data.Labels.Subject)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "subjects_list", data); err != nil {
		slog.Error("render subjects list", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Subjects) orgMembership(r *http.Request, userID int64) (orgRole string, orgMember bool) {
	orgRole, orgMember, _ = h.Store.OrganizationMemberRole(r.Context(), 0, userID)
	if org, ok := middleware.OrganizationFromContext(r.Context()); ok {
		orgRole, orgMember, _ = h.Store.OrganizationMemberRole(r.Context(), org.ID, userID)
	}
	return orgRole, orgMember
}

// NewForm renders the create subject form.
func (h *Subjects) NewForm(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if !CanCreateSubject(user) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	orgRole, orgMember := h.orgMembership(r, user.ID)
	pd := h.subjectsPageData(r, "Nouveau "+strings.ToLower(templates.DefaultUILabels().Subject.Singular))
	if adminSubjectsRoute(r) {
		pd.Breadcrumbs = templates.BCAdminSubjectNew(pd.Labels.Subject)
	} else {
		pd.Breadcrumbs = templates.BCSubjectNew(pd.Labels.Subject)
	}
	data := templates.SubjectFormData{
		PageData:         pd,
		FormAction:       subjectsListPath(r),
		CanSetVisibility: CanSetSubjectVisibility(user, orgRole, orgMember, store.SubjectAccess{}),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "subject_form", data); err != nil {
		slog.Error("render subject form", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Create stores a new subject.
func (h *Subjects) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if !CanCreateSubject(user) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	orgRole, orgMember := h.orgMembership(r, user.ID)
	canSetVisibility := CanSetSubjectVisibility(user, orgRole, orgMember, store.SubjectAccess{})
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		h.renderFormError(w, r, nil, store.SubjectAccess{}, subjectsListPath(r), "Le nom est obligatoire.")
		return
	}
	description := strings.TrimSpace(r.FormValue("description"))
	domains := store.ParseTagsCSV(r.FormValue("domains"))
	tags := store.ParseTagsCSV(r.FormValue("tags"))
	visibility := store.SubjectVisibilityNormal
	if canSetVisibility {
		var normErr error
		visibility, normErr = store.NormalizeSubjectVisibility(strings.TrimSpace(r.FormValue("visibility")))
		if normErr != nil {
			h.renderFormError(w, r, nil, store.SubjectAccess{}, subjectsListPath(r), "Visibilité invalide.")
			return
		}
	}

	subject, err := h.Store.CreateSubjectWithVisibility(r.Context(), name, description, user.ID, domains, visibility)
	if err != nil {
		slog.Error("create subject", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := h.Store.SetSubjectTags(r.Context(), subject.ID, tags); err != nil {
		slog.Error("set subject tags", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	msg := templates.DefaultUILabels().Subject.Singular + "+cr%C3%A9%C3%A9"
	http.Redirect(w, r, "/subjects/"+strconv.FormatInt(subject.ID, 10)+"?msg="+msg, http.StatusSeeOther)
}

// Show displays subject details.
func (h *Subjects) Show(w http.ResponseWriter, r *http.Request) {
	subject, user, access, ok := h.loadSubject(w, r)
	if !ok {
		return
	}

	directMembers, err := h.Store.ListDirectSubjectMembers(r.Context(), subject.ID)
	if err != nil {
		slog.Error("list direct subject members", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	subjectTeams, err := h.Store.ListSubjectTeams(r.Context(), subject.ID)
	if err != nil {
		slog.Error("list subject teams", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	orgTeams, err := h.Store.ListOrganizationTeams(r.Context())
	if err != nil {
		slog.Error("list organization teams", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	subjectRuns, err := h.Store.ListRunsWithProgressBySubject(r.Context(), subject.ID)
	if err != nil {
		slog.Error("list subject runs", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	nokItems, err := h.Store.ListSubjectNokItems(r.Context(), subject.ID)
	if err != nil {
		slog.Error("list subject nok items", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	domains, err := h.Store.ListSubjectDomains(r.Context(), subject.ID)
	if err != nil {
		slog.Error("list subject domains", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tags, err := h.Store.ListSubjectTags(r.Context(), subject.ID)
	if err != nil {
		slog.Error("list subject tags", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	callerOrgRole, callerOrgMember, _ := h.Store.OrganizationMemberRole(r.Context(), subject.OrganizationID, user.ID)
	data, err := h.buildSubjectShowData(r, subject, user, access, callerOrgRole, callerOrgMember, domains, tags, nil, directMembers, subjectTeams, orgTeams, subjectRuns, nokItems, subjectShowExtras{
		message: r.URL.Query().Get("msg"),
	})
	if err != nil {
		slog.Error("build subject show data", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "subject_show", data); err != nil {
		slog.Error("render subject show", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func requestOrgPolicies(r *http.Request) store.OrgLeadPolicies {
	org, _ := middleware.OrganizationFromContext(r.Context())
	return PoliciesFromOrganization(org)
}

func (h *Subjects) denyAssignTeams(w http.ResponseWriter, r *http.Request, subject *Subject, user *User, access store.SubjectAccess, policies store.OrgLeadPolicies) bool {
	if CanAssignSubjectTeams(user, access, policies) {
		return false
	}
	if leadBlockedByAssignTeamsPolicy(user, access, policies) {
		h.renderShowError(w, r, subject, user, access, "La politique de l'organisation n'autorise pas les leads à affecter des équipes.")
		return true
	}
	http.NotFound(w, r)
	return true
}

// AddTeam grants an organization team a role on the subject.
func (h *Subjects) AddTeam(w http.ResponseWriter, r *http.Request) {
	subject, user, access, ok := h.loadSubject(w, r)
	if !ok {
		return
	}
	policies := requestOrgPolicies(r)
	if h.denyAssignTeams(w, r, subject, user, access, policies) {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	teamID, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("team_id")), 10, 64)
	if err != nil || teamID <= 0 {
		h.renderShowError(w, r, subject, user, access, "Équipe invalide.")
		return
	}
	role := strings.TrimSpace(r.FormValue("role"))
	if role == "" {
		role = store.SubjectRoleViewer
	}

	if err = h.Store.GrantTeamSubjectRole(r.Context(), teamID, subject.ID, role, user.ID); err != nil {
		if errors.Is(err, ErrTeamNotFound) || errors.Is(err, ErrSubjectNotFound) {
			http.NotFound(w, r)
			return
		}
		if errors.Is(err, ErrInvalidSubjectRole) {
			h.renderShowError(w, r, subject, user, access, "Rôle invalide.")
			return
		}
		slog.Error("grant team subject role", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/subjects/"+strconv.FormatInt(subject.ID, 10)+"?msg=%C3%89quipe+ajout%C3%A9e", http.StatusSeeOther)
}

// RemoveTeam revokes a team's role on the subject.
func (h *Subjects) RemoveTeam(w http.ResponseWriter, r *http.Request) {
	subject, user, access, ok := h.loadSubject(w, r)
	if !ok {
		return
	}
	policies := requestOrgPolicies(r)
	if h.denyAssignTeams(w, r, subject, user, access, policies) {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	teamID, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("team_id")), 10, 64)
	if err != nil || teamID <= 0 {
		h.renderShowError(w, r, subject, user, access, "Équipe invalide.")
		return
	}

	if err = h.Store.RevokeTeamSubjectRole(r.Context(), teamID, subject.ID); err != nil {
		if errors.Is(err, ErrTeamNotFound) || errors.Is(err, ErrSubjectNotFound) {
			http.NotFound(w, r)
			return
		}
		if errors.Is(err, ErrTeamSubjectRoleNotFound) {
			h.renderShowError(w, r, subject, user, access, "Cette équipe n'est pas affectée à ce sujet.")
			return
		}
		slog.Error("revoke team subject role", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/subjects/"+strconv.FormatInt(subject.ID, 10)+"?msg=%C3%89quipe+retir%C3%A9e", http.StatusSeeOther)
}

// PreviewTeam returns an HTMX fragment describing the team assignment impact.
func (h *Subjects) PreviewTeam(w http.ResponseWriter, r *http.Request) {
	_, user, access, ok := h.loadSubject(w, r)
	if !ok {
		return
	}
	policies := requestOrgPolicies(r)
	if !CanAssignSubjectTeams(user, access, policies) {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	teamID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("team_id")), 10, 64)
	if err != nil || teamID <= 0 {
		if execErr := h.Templates.ExecuteTemplate(w, "subject_team_preview_fragment", templates.SubjectTeamPreviewData{Empty: true}); execErr != nil {
			slog.Error("render team preview empty", "err", execErr)
		}
		return
	}
	role := strings.TrimSpace(r.URL.Query().Get("role"))
	if role == "" {
		role = store.SubjectRoleViewer
	}
	switch role {
	case store.SubjectRoleLead, store.SubjectRoleContributor, store.SubjectRoleViewer:
	default:
		if execErr := h.Templates.ExecuteTemplate(w, "subject_team_preview_fragment", templates.SubjectTeamPreviewData{Empty: true}); execErr != nil {
			slog.Error("render team preview invalid role", "err", execErr)
		}
		return
	}

	team, err := h.Store.TeamByID(r.Context(), teamID)
	if err != nil {
		if errors.Is(err, ErrTeamNotFound) {
			http.NotFound(w, r)
			return
		}
		slog.Error("team by id for preview", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	members, err := h.Store.ListTeamMembers(r.Context(), teamID)
	if err != nil {
		slog.Error("list team members for preview", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := templates.SubjectTeamPreviewData{
		TeamName:    team.Name,
		MemberCount: len(members),
		RoleLabel:   templates.FormatRole(role),
	}
	if execErr := h.Templates.ExecuteTemplate(w, "subject_team_preview_fragment", data); execErr != nil {
		slog.Error("render team preview", "err", execErr)
	}
}

// AddMember adds a direct subject member by email (org member or external).
func (h *Subjects) AddMember(w http.ResponseWriter, r *http.Request) {
	subject, user, access, ok := h.loadSubject(w, r)
	if !ok {
		return
	}
	policies := requestOrgPolicies(r)
	if !CanManageSubjectMembers(user, access, policies) {
		if CanLeadAccess(user, access) {
			h.renderShowError(w, r, subject, user, access, "La politique de l'organisation n'autorise pas les leads à inviter des membres.")
			return
		}
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	role := strings.TrimSpace(r.FormValue("role"))
	if role == "" {
		role = store.SubjectRoleViewer
	}
	if email == "" {
		h.renderShowError(w, r, subject, user, access, "Email requis.")
		return
	}

	invitee, err := h.Store.UserByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			h.renderShowError(w, r, subject, user, access, "Aucun compte avec cet email. L'invité doit d'abord se connecter via GitHub.")
			return
		}
		slog.Error("user by email for subject member", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	org, orgOK := middleware.OrganizationFromContext(r.Context())
	inviteeIsOrgMember := false
	if orgOK {
		_, inviteeIsOrgMember, _ = h.Store.OrganizationMemberRole(r.Context(), org.ID, invitee.ID)
	}
	if !CanInviteSubjectMember(user, access, policies, inviteeIsOrgMember) {
		if inviteeIsOrgMember {
			h.renderShowError(w, r, subject, user, access, "La politique de l'organisation n'autorise pas les leads à inviter des membres.")
		} else {
			h.renderShowError(w, r, subject, user, access, "La politique de l'organisation n'autorise pas les leads à inviter des externes.")
		}
		return
	}

	if err = h.Store.UpsertDirectSubjectMember(r.Context(), subject.ID, invitee.ID, role); err != nil {
		if errors.Is(err, ErrInvalidSubjectRole) {
			h.renderShowError(w, r, subject, user, access, "Rôle invalide.")
			return
		}
		slog.Error("upsert direct subject member", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/subjects/"+strconv.FormatInt(subject.ID, 10)+"?msg=Membre+ajout%C3%A9", http.StatusSeeOther)
}

// RemoveMember removes a direct subject member.
func (h *Subjects) RemoveMember(w http.ResponseWriter, r *http.Request) {
	subject, user, access, ok := h.loadSubject(w, r)
	if !ok {
		return
	}
	policies := requestOrgPolicies(r)
	if !CanManageSubjectMembers(user, access, policies) {
		if CanLeadAccess(user, access) {
			h.renderShowError(w, r, subject, user, access, "La politique de l'organisation n'autorise pas les leads à gérer les membres.")
			return
		}
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("user_id")), 10, 64)
	if err != nil || userID <= 0 {
		h.renderShowError(w, r, subject, user, access, "Membre invalide.")
		return
	}

	if err = h.Store.RemoveDirectSubjectMember(r.Context(), subject.ID, userID); err != nil {
		if errors.Is(err, ErrDirectSubjectMemberNotFound) || errors.Is(err, ErrSubjectNotFound) {
			h.renderShowError(w, r, subject, user, access, "Ce membre n'est pas affecté à ce sujet.")
			return
		}
		slog.Error("remove direct subject member", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/subjects/"+strconv.FormatInt(subject.ID, 10)+"?msg=Membre+retir%C3%A9", http.StatusSeeOther)
}

// EditForm renders the edit subject form.
func (h *Subjects) EditForm(w http.ResponseWriter, r *http.Request) {
	subject, user, access, ok := h.loadSubject(w, r)
	if !ok {
		return
	}
	if !CanManageAccess(user, access) {
		http.NotFound(w, r)
		return
	}

	pd := h.subjectsPageData(r, "Modifier "+subject.Name)
	if adminSubjectsRoute(r) {
		pd.Breadcrumbs = templates.BCAdminSubjectEdit(subject.Name, subject.ID, pd.Labels.Subject)
	} else {
		pd.Breadcrumbs = templates.BCSubjectEdit(subject.Name, subject.ID, pd.Labels.Subject)
	}
	orgRole, orgMember := h.orgMembership(r, user.ID)
	data := templates.SubjectFormData{
		PageData:         pd,
		Subject:          subject,
		FormAction:       subjectFormAction(r, subject.ID),
		CanSetVisibility: CanSetSubjectVisibility(user, orgRole, orgMember, access),
	}
	if domains, err := h.Store.ListSubjectDomains(r.Context(), subject.ID); err != nil {
		slog.Error("list subject domains", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	} else {
		data.Domains = store.FormatTagsCSV(domains)
	}
	if tags, err := h.Store.ListSubjectTags(r.Context(), subject.ID); err != nil {
		slog.Error("list subject tags", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	} else {
		data.Tags = store.FormatTagsCSV(tags)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "subject_form", data); err != nil {
		slog.Error("render subject edit", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Update saves subject fields.
func (h *Subjects) Update(w http.ResponseWriter, r *http.Request) {
	subject, user, access, ok := h.loadSubject(w, r)
	if !ok {
		return
	}
	if !CanManageAccess(user, access) {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	orgRole, orgMember := h.orgMembership(r, user.ID)
	canSetVisibility := CanSetSubjectVisibility(user, orgRole, orgMember, access)
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		h.renderFormError(w, r, subject, access, subjectFormAction(r, subject.ID), "Le nom est obligatoire.")
		return
	}
	description := strings.TrimSpace(r.FormValue("description"))
	domains := store.ParseTagsCSV(r.FormValue("domains"))
	tags := store.ParseTagsCSV(r.FormValue("tags"))
	visibility := subject.Visibility
	if canSetVisibility {
		var normErr error
		visibility, normErr = store.NormalizeSubjectVisibility(strings.TrimSpace(r.FormValue("visibility")))
		if normErr != nil {
			h.renderFormError(w, r, subject, access, subjectFormAction(r, subject.ID), "Visibilité invalide.")
			return
		}
	}

	if err := h.Store.UpdateSubjectWithVisibility(r.Context(), subject.ID, name, description, domains, visibility); err != nil {
		slog.Error("update subject", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if visibility == store.SubjectVisibilityPrivate && !access.IsSupervisor() {
		if err := h.Store.UpsertDirectSubjectMember(r.Context(), subject.ID, user.ID, store.SubjectRoleLead); err != nil {
			slog.Error("ensure lead on private subject", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	if err := h.Store.SetSubjectTags(r.Context(), subject.ID, tags); err != nil {
		slog.Error("set subject tags", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	msg := templates.DefaultUILabels().Subject.Singular + "+mis+%C3%A0+jour"
	redirectPath := "/subjects/" + strconv.FormatInt(subject.ID, 10)
	if adminSubjectsRoute(r) {
		redirectPath = subjectsListPath(r)
	}
	http.Redirect(w, r, redirectPath+"?msg="+msg, http.StatusSeeOther)
}

// Archive marks the subject archived.
func (h *Subjects) Archive(w http.ResponseWriter, r *http.Request) {
	subject, user, access, ok := h.loadSubject(w, r)
	if !ok {
		return
	}
	if !CanManageAccess(user, access) {
		http.NotFound(w, r)
		return
	}

	if err := h.Store.ArchiveSubject(r.Context(), subject.ID); err != nil {
		slog.Error("archive subject", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	msg := templates.DefaultUILabels().Subject.Singular + "+archiv%C3%A9"
	http.Redirect(w, r, subjectsListPath(r)+"?msg="+msg, http.StatusSeeOther)
}

// WizardNouvelle is run wizard step 1: choose or search subjects.
func (h *Subjects) WizardNouvelle(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	templateID := parseWizardTemplateID(r)
	filterQuery := strings.TrimSpace(r.URL.Query().Get("q"))
	allSubjects, err := h.Store.ListSubjects(r.Context(), user.ID, auth.HasMinRole(user.Role, auth.RoleAdmin), filterQuery)
	if err != nil {
		slog.Error("list subjects for run wizard", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var launchSubjects []Subject
	for _, subject := range allSubjects {
		access, err := h.Store.ResolveSubjectAccess(r.Context(), user.ID, subject.ID, user.Role)
		if err != nil {
			slog.Error("resolve subject access", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if CanContributeAccess(user, access) {
			launchSubjects = append(launchSubjects, subject)
		}
	}

	if len(launchSubjects) == 1 && filterQuery == "" {
		http.Redirect(w, r, templates.SubjectTemplatesForRunPath(launchSubjects[0].ID, templateID), http.StatusSeeOther)
		return
	}

	pd := h.PageData(r, "Lancer une revue")
	pd.Breadcrumbs = templates.BCRunWizardSubjects()
	pd.ActiveTab = "runs"
	data := templates.RunWizardSubjectsData{
		PageData:           pd,
		Subjects:           launchSubjects,
		SelectedTemplateID: templateID,
		Message:            r.URL.Query().Get("msg"),
		FilterQuery:        filterQuery,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "run_wizard_subjects", data); err != nil {
		slog.Error("render run wizard step 1", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// WizardNouvelleCreate selects or inline-creates a subject for the run wizard.
func (h *Subjects) WizardNouvelleCreate(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	subjectIDRaw := strings.TrimSpace(r.FormValue("subject_id"))
	templateID := parseWizardTemplateID(r)
	if templateID <= 0 {
		if raw := strings.TrimSpace(r.FormValue("template")); raw != "" {
			if id, err := strconv.ParseInt(raw, 10, 64); err == nil && id > 0 {
				templateID = id
			}
		}
	}
	if subjectIDRaw != "" {
		subjectID, err := strconv.ParseInt(subjectIDRaw, 10, 64)
		if err != nil {
			h.renderWizardError(w, r, "Sujet invalide.")
			return
		}
		_, _, access, ok := h.loadSubjectByID(w, r, subjectID)
		if !ok {
			return
		}
		if !CanContributeAccess(user, access) {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, templates.SubjectTemplatesForRunPath(subjectID, templateID), http.StatusSeeOther)
		return
	}

	if !CanCreateSubject(user) {
		http.NotFound(w, r)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		h.renderWizardError(w, r, "Le nom est obligatoire.")
		return
	}

	subject, err := h.Store.CreateSubject(r.Context(), name, "", user.ID, store.ParseTagsCSV(r.FormValue("domains")))
	if err != nil {
		slog.Error("inline create subject", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, templates.SubjectTemplatesForRunPath(subject.ID, templateID), http.StatusSeeOther)
}

func parseWizardTemplateID(r *http.Request) int64 {
	raw := strings.TrimSpace(r.URL.Query().Get("template"))
	if raw == "" {
		return 0
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0
	}
	return id
}

func (h *Subjects) loadSubject(w http.ResponseWriter, r *http.Request) (*Subject, *store.User, store.SubjectAccess, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return nil, nil, store.SubjectAccess{}, false
	}
	return h.loadSubjectByID(w, r, id)
}

func (h *Subjects) loadSubjectByID(w http.ResponseWriter, r *http.Request, id int64) (*Subject, *store.User, store.SubjectAccess, bool) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return nil, nil, store.SubjectAccess{}, false
	}

	subject, err := h.Store.SubjectByID(r.Context(), id)
	if errors.Is(err, ErrSubjectNotFound) {
		http.NotFound(w, r)
		return nil, nil, store.SubjectAccess{}, false
	}
	if err != nil {
		slog.Error("load subject", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, store.SubjectAccess{}, false
	}

	access, err := h.Store.ResolveSubjectAccess(r.Context(), user.ID, subject.ID, user.Role)
	if err != nil {
		slog.Error("resolve subject access", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, store.SubjectAccess{}, false
	}

	if !CanViewAccess(access) {
		http.NotFound(w, r)
		return nil, nil, store.SubjectAccess{}, false
	}

	return subject, user, access, true
}

func subjectFormAction(r *http.Request, id int64) string {
	if adminSubjectsRoute(r) {
		return templates.PathAdminSubjects + "/" + strconv.FormatInt(id, 10)
	}
	return templates.PathSubjects + "/" + strconv.FormatInt(id, 10)
}

func (h *Subjects) renderFormError(w http.ResponseWriter, r *http.Request, subject *Subject, access store.SubjectAccess, action, message string) {
	title := "Nouveau " + strings.ToLower(templates.DefaultUILabels().Subject.Singular)
	if subject != nil {
		title = "Modifier " + subject.Name
	}
	pd := h.subjectsPageData(r, title)
	if subject != nil {
		if adminSubjectsRoute(r) {
			pd.Breadcrumbs = templates.BCAdminSubjectEdit(subject.Name, subject.ID, pd.Labels.Subject)
		} else {
			pd.Breadcrumbs = templates.BCSubjectEdit(subject.Name, subject.ID, pd.Labels.Subject)
		}
	} else if adminSubjectsRoute(r) {
		pd.Breadcrumbs = templates.BCAdminSubjectNew(pd.Labels.Subject)
	} else {
		pd.Breadcrumbs = templates.BCSubjectNew(pd.Labels.Subject)
	}
	user, _ := middleware.UserFromContext(r.Context())
	orgRole, orgMember := "", false
	if user != nil {
		orgRole, orgMember = h.orgMembership(r, user.ID)
	}
	data := templates.SubjectFormData{
		PageData:         pd,
		Subject:          subject,
		FormAction:       action,
		CanSetVisibility: user != nil && CanSetSubjectVisibility(user, orgRole, orgMember, access),
		Error:            message,
	}
	if subject != nil {
		if domains, err := h.Store.ListSubjectDomains(r.Context(), subject.ID); err == nil {
			data.Domains = store.FormatTagsCSV(domains)
		}
		if tags, err := h.Store.ListSubjectTags(r.Context(), subject.ID); err == nil {
			data.Tags = store.FormatTagsCSV(tags)
		}
	} else {
		data.Domains = store.FormatTagsCSV(store.ParseTagsCSV(r.FormValue("domains")))
		data.Tags = store.FormatTagsCSV(store.ParseTagsCSV(r.FormValue("tags")))
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	if err := h.Templates.ExecuteTemplate(w, "subject_form", data); err != nil {
		slog.Error("render subject form error", "err", err)
	}
}

func (h *Subjects) renderWizardError(w http.ResponseWriter, r *http.Request, message string) {
	templateID := parseWizardTemplateID(r)
	pd := h.PageData(r, "Lancer une revue")
	pd.Breadcrumbs = templates.BCRunWizardSubjects()
	pd.ActiveTab = "runs"
	data := templates.RunWizardSubjectsData{
		PageData:           pd,
		SelectedTemplateID: templateID,
		Error:              message,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	if err := h.Templates.ExecuteTemplate(w, "run_wizard_subjects", data); err != nil {
		slog.Error("render run wizard error", "err", err)
	}
}

type subjectShowExtras struct {
	message        string
	errMsg         string
	addTeam        int64
	addRole        string
	addMemberEmail string
	addMemberRole  string
}

func (h *Subjects) buildSubjectShowData(
	r *http.Request,
	subject *Subject,
	user *store.User,
	access store.SubjectAccess,
	callerOrgRole string,
	callerOrgMember bool,
	domains, tags []string,
	members []SubjectMember,
	directMembers []DirectSubjectMember,
	subjectTeams []TeamSubjectRole,
	orgTeams []OrganizationTeam,
	subjectRuns []RunWithProgress,
	nokItems []SubjectNokItemSummary,
	extras subjectShowExtras,
) (templates.SubjectShowData, error) {
	policies := requestOrgPolicies(r)
	canManage := CanManageAccess(user, access)
	canLaunch := CanContributeAccess(user, access)
	canAssignTeams := CanAssignSubjectTeams(user, access, policies)
	canManageMembers := CanManageSubjectMembers(user, access, policies)
	editPath := templates.PathSubjects + "/" + strconv.FormatInt(subject.ID, 10) + "/edit"
	if CanManageOrgUsers(user, callerOrgRole, callerOrgMember) {
		editPath = templates.PathAdminSubjects + "/" + strconv.FormatInt(subject.ID, 10) + "/edit"
	}

	assigned := make(map[int64]struct{}, len(subjectTeams))
	teamNames := make(map[int64]string, len(orgTeams)+len(subjectTeams))
	for _, t := range subjectTeams {
		assigned[t.TeamID] = struct{}{}
		teamNames[t.TeamID] = t.TeamName
	}
	var available []OrganizationTeam
	for _, t := range orgTeams {
		teamNames[t.ID] = t.Name
		if _, ok := assigned[t.ID]; !ok {
			available = append(available, t)
		}
	}

	accessSources := make([]string, 0, len(access.Sources))
	for _, src := range access.Sources {
		accessSources = append(accessSources, templates.FormatAccessSource(src, teamNames))
	}

	addRole := extras.addRole
	if addRole == "" {
		addRole = store.SubjectRoleViewer
	}
	addMemberRole := extras.addMemberRole
	if addMemberRole == "" {
		addMemberRole = store.SubjectRoleViewer
	}

	pd := h.PageDataTab(r, subject.Name, "")
	pd.Breadcrumbs = templates.BCSubjectShow(subject.Name, pd.Labels.Subject)

	return templates.SubjectShowData{
		PageData:            pd,
		Subject:             subject,
		Domains:             domains,
		Tags:                tags,
		Members:             members,
		DirectMembers:       directMembers,
		Teams:               subjectTeams,
		AvailableTeams:      available,
		AccessSources:       accessSources,
		Runs:                subjectRuns,
		NokItems:            nokItems,
		MemberRole:          DisplayRole(access),
		CanManage:           canManage,
		CanManageMembers:    canManageMembers,
		CanAssignTeams:      canAssignTeams,
		TeamsPolicyDenied:   leadBlockedByAssignTeamsPolicy(user, access, policies),
		MembersPolicyDenied: CanLeadAccess(user, access) && !canManageMembers && !auth.HasMinRole(user.Role, auth.RoleAdmin) && !access.HasSource(store.AccessSourceOrgAdmin),
		CanLaunch:           canLaunch,
		EditPath:            editPath,
		AddMemberEmail:      extras.addMemberEmail,
		AddMemberRole:       addMemberRole,
		AddTeamID:           extras.addTeam,
		AddTeamRole:         addRole,
		Message:             extras.message,
		Error:               extras.errMsg,
	}, nil
}

func (h *Subjects) renderShowError(w http.ResponseWriter, r *http.Request, subject *Subject, user *store.User, access store.SubjectAccess, message string) {
	directMembers, err := h.Store.ListDirectSubjectMembers(r.Context(), subject.ID)
	if err != nil {
		slog.Error("list direct subject members", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	subjectTeams, err := h.Store.ListSubjectTeams(r.Context(), subject.ID)
	if err != nil {
		slog.Error("list subject teams", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	orgTeams, err := h.Store.ListOrganizationTeams(r.Context())
	if err != nil {
		slog.Error("list organization teams", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	subjectRuns, err := h.Store.ListRunsWithProgressBySubject(r.Context(), subject.ID)
	if err != nil {
		slog.Error("list subject runs", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	nokItems, err := h.Store.ListSubjectNokItems(r.Context(), subject.ID)
	if err != nil {
		slog.Error("list subject nok items", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	domains, err := h.Store.ListSubjectDomains(r.Context(), subject.ID)
	if err != nil {
		slog.Error("list subject domains", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	tags, err := h.Store.ListSubjectTags(r.Context(), subject.ID)
	if err != nil {
		slog.Error("list subject tags", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	teamID, _ := strconv.ParseInt(strings.TrimSpace(r.FormValue("team_id")), 10, 64)
	callerOrgRole, callerOrgMember, _ := h.Store.OrganizationMemberRole(r.Context(), subject.OrganizationID, user.ID)
	memberRole := strings.TrimSpace(r.FormValue("role"))
	data, err := h.buildSubjectShowData(r, subject, user, access, callerOrgRole, callerOrgMember, domains, tags, nil, directMembers, subjectTeams, orgTeams, subjectRuns, nokItems, subjectShowExtras{
		errMsg:         message,
		addTeam:        teamID,
		addRole:        memberRole,
		addMemberEmail: strings.TrimSpace(r.FormValue("email")),
		addMemberRole:  memberRole,
	})
	if err != nil {
		slog.Error("build subject show data", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	if err := h.Templates.ExecuteTemplate(w, "subject_show", data); err != nil {
		slog.Error("render subject show error", "err", err)
	}
}
