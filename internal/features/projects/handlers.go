package projects

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/middleware"
	"github.com/jeb-maker/revues/internal/web/templates"
)

// Deps holds shared dependencies for the projects HTTP handlers.
//
// This mirrors internal/web/handlerdeps.HandlerDeps but is local to the
// projects feature package to avoid an import cycle.
type Deps struct {
	Templates     *template.Template
	Store         ProjectStore
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

// Projects handles project CRUD and membership.
type Projects struct {
	Deps
}

// List shows projects visible to the current user.
func (h *Projects) List(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	admin := auth.HasMinRole(user.Role, auth.RoleAdmin)
	items, err := h.Store.ListProjects(r.Context(), user.ID, admin)
	if err != nil {
		slog.Error("list projects", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	orgRole, orgMember, _ := h.Store.OrganizationMemberRole(r.Context(), 0, user.ID)
	if org, ok := middleware.OrganizationFromContext(r.Context()); ok {
		orgRole, orgMember, _ = h.Store.OrganizationMemberRole(r.Context(), org.ID, user.ID)
	}

	data := templates.ProjectsListData{
		PageData:          h.PageDataTab(r, "Projets", "projects"),
		Projects:          items,
		CanCreate:         CanCreate(user),
		CanManageOrgUsers: CanManageOrgUsers(user, orgRole, orgMember),
		Message:           r.URL.Query().Get("msg"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "projects_list", data); err != nil {
		slog.Error("render projects list", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// NewForm renders the create project form.
func (h *Projects) NewForm(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if !CanCreate(user) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	pd := h.PageDataTab(r, "Nouveau projet", "projects")
	pd.Breadcrumbs = []templates.Breadcrumb{
		{URL: "/projects", Label: "Projets"},
		{Label: "Nouveau projet"},
	}
	data := templates.ProjectFormData{
		PageData:   pd,
		FormAction: "/projects",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "project_form", data); err != nil {
		slog.Error("render project form", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Create stores a new project.
func (h *Projects) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if !CanCreate(user) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		h.renderFormError(w, r, nil, "/projects", "Le nom est obligatoire.")
		return
	}
	description := strings.TrimSpace(r.FormValue("description"))

	project, err := h.Store.CreateProject(r.Context(), name, description, user.ID, store.ParseTagsCSV(r.FormValue("tags")))
	if err != nil {
		slog.Error("create project", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/projects/"+strconv.FormatInt(project.ID, 10)+"?msg=Projet+cr%C3%A9%C3%A9", http.StatusSeeOther)
}

// Show displays project details and members.
func (h *Projects) Show(w http.ResponseWriter, r *http.Request) {
	project, user, memberRole, ok := h.loadProject(w, r)
	if !ok {
		return
	}

	members, err := h.Store.ListProjectMembers(r.Context(), project.ID)
	if err != nil {
		slog.Error("list project members", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	projectRuns, err := h.Store.ListRunsWithProgressByProject(r.Context(), project.ID)
	if err != nil {
		slog.Error("list project runs", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	nokItems, err := h.Store.ListProjectNokItems(r.Context(), project.ID)
	if err != nil {
		slog.Error("list project nok items", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tags, err := h.Store.ListProjectTags(r.Context(), project.ID)
	if err != nil {
		slog.Error("list project tags", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	pd := h.PageDataTab(r, project.Name, "")
	pd.Breadcrumbs = []templates.Breadcrumb{
		{URL: "/projects", Label: "Projets"},
		{Label: project.Name},
	}
	callerOrgRole, _, _ := h.Store.OrganizationMemberRole(r.Context(), project.OrganizationID, user.ID)

	data := templates.ProjectShowData{
		PageData:         pd,
		Project:          project,
		Tags:             tags,
		Members:          members,
		Runs:             projectRuns,
		NokItems:         nokItems,
		MemberRole:       memberRole,
		CanManage:        CanManage(user, memberRole),
		CanManageMembers: CanAddProjectMember(user, memberRole, callerOrgRole),
		CanLaunch:        CanLaunch(user, memberRole),
		Message:          r.URL.Query().Get("msg"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "project_show", data); err != nil {
		slog.Error("render project show", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// EditForm renders the edit project form.
func (h *Projects) EditForm(w http.ResponseWriter, r *http.Request) {
	project, user, memberRole, ok := h.loadProject(w, r)
	if !ok {
		return
	}
	if !CanManage(user, memberRole) {
		http.NotFound(w, r)
		return
	}

	pd2 := h.PageDataTab(r, "Modifier "+project.Name, "")
	pd2.Breadcrumbs = []templates.Breadcrumb{
		{URL: "/projects", Label: "Projets"},
		{URL: "/projects/" + strconv.FormatInt(project.ID, 10), Label: project.Name},
		{Label: "Modifier"},
	}
	data := templates.ProjectFormData{
		PageData:   pd2,
		Project:    project,
		FormAction: "/projects/" + strconv.FormatInt(project.ID, 10),
	}
	if tags, err := h.Store.ListProjectTags(r.Context(), project.ID); err != nil {
		slog.Error("list project tags", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	} else {
		data.Tags = store.FormatTagsCSV(tags)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "project_form", data); err != nil {
		slog.Error("render project edit", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Update saves project fields.
func (h *Projects) Update(w http.ResponseWriter, r *http.Request) {
	project, user, memberRole, ok := h.loadProject(w, r)
	if !ok {
		return
	}
	if !CanManage(user, memberRole) {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		h.renderFormError(w, r, project, "/projects/"+strconv.FormatInt(project.ID, 10), "Le nom est obligatoire.")
		return
	}
	description := strings.TrimSpace(r.FormValue("description"))

	if err := h.Store.UpdateProject(r.Context(), project.ID, name, description, store.ParseTagsCSV(r.FormValue("tags"))); err != nil {
		slog.Error("update project", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/projects/"+strconv.FormatInt(project.ID, 10)+"?msg=Projet+mis+%C3%A0+jour", http.StatusSeeOther)
}

// Archive marks the project archived.
func (h *Projects) Archive(w http.ResponseWriter, r *http.Request) {
	project, user, memberRole, ok := h.loadProject(w, r)
	if !ok {
		return
	}
	if !CanManage(user, memberRole) {
		http.NotFound(w, r)
		return
	}

	if err := h.Store.ArchiveProject(r.Context(), project.ID); err != nil {
		slog.Error("archive project", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/projects?msg=Projet+archiv%C3%A9", http.StatusSeeOther)
}

// AddMember adds or updates a project member by email.
func (h *Projects) AddMember(w http.ResponseWriter, r *http.Request) {
	project, user, memberRole, ok := h.loadProject(w, r)
	if !ok {
		return
	}

	callerOrgRole, _, err := h.Store.OrganizationMemberRole(r.Context(), project.OrganizationID, user.ID)
	if err != nil {
		slog.Error("caller org role", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !CanAddProjectMember(user, memberRole, callerOrgRole) {
		http.NotFound(w, r)
		return
	}
	if err = r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	email, err := normalizeMemberEmail(r.FormValue("email"))
	if err != nil {
		h.renderShowError(w, r, project, user, memberRole, callerOrgRole, "Email invalide.")
		return
	}

	role := strings.TrimSpace(r.FormValue("role"))
	if !ValidLocalRole(role) {
		h.renderShowError(w, r, project, user, memberRole, callerOrgRole, "Rôle local invalide.")
		return
	}

	member, err := h.Store.UserByEmail(r.Context(), email)
	if errors.Is(err, ErrUserNotFound) {
		if !CanInviteExternalToOrg(user, memberRole, callerOrgRole) {
			http.NotFound(w, r)
			return
		}
		if err = h.Store.CreateOrganizationInvitation(r.Context(), email, project.OrganizationID, project.ID, role); err != nil {
			slog.Error("create organization invitation", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/projects/"+strconv.FormatInt(project.ID, 10)+"?msg=Invitation+envoy%C3%A9e", http.StatusSeeOther)
		return
	}
	if err != nil {
		slog.Error("lookup member user", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	inviteeOrgRole, inviteeInOrg, err := h.Store.OrganizationMemberRole(r.Context(), project.OrganizationID, member.ID)
	if err != nil {
		slog.Error("invitee org role", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !inviteeInOrg {
		if !CanInviteExternalToOrg(user, memberRole, callerOrgRole) {
			http.NotFound(w, r)
			return
		}
		if err = h.Store.AddOrganizationMember(r.Context(), project.OrganizationID, member.ID, store.OrgRoleMember); err != nil {
			slog.Error("add organization member", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		_ = inviteeOrgRole
	}

	if err = h.Store.AddProjectMember(r.Context(), project.ID, member.ID, role); err != nil {
		slog.Error("add project member", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	msg := "Membre+ajout%C3%A9"
	if !inviteeInOrg {
		msg = "Membre+ajout%C3%A9+%28adh%C3%A9sion+%C3%A0+l%27organisation%29"
	}
	http.Redirect(w, r, "/projects/"+strconv.FormatInt(project.ID, 10)+"?msg="+msg, http.StatusSeeOther)
}

// RemoveMember removes a member from the project.
func (h *Projects) RemoveMember(w http.ResponseWriter, r *http.Request) {
	project, user, memberRole, ok := h.loadProject(w, r)
	if !ok {
		return
	}
	if !CanManageMembers(user, memberRole) {
		http.NotFound(w, r)
		return
	}
	callerOrgRole, _, err := h.Store.OrganizationMemberRole(r.Context(), project.OrganizationID, user.ID)
	if err != nil {
		slog.Error("caller org role", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err = r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
	if err != nil {
		h.renderShowError(w, r, project, user, memberRole, callerOrgRole, "Membre invalide.")
		return
	}

	if userID == user.ID {
		h.renderShowError(w, r, project, user, memberRole, callerOrgRole, "Vous ne pouvez pas vous retirer vous-même.")
		return
	}

	targetRole, isMember, err := h.Store.MemberRole(r.Context(), project.ID, userID)
	if err != nil {
		slog.Error("member role lookup", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		h.renderShowError(w, r, project, user, memberRole, callerOrgRole, "Membre introuvable.")
		return
	}

	if targetRole == LocalRoleLead {
		leads, err := h.Store.CountProjectLeads(r.Context(), project.ID)
		if err != nil {
			slog.Error("count project leads", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if leads <= 1 {
			h.renderShowError(w, r, project, user, memberRole, callerOrgRole, "Impossible de retirer le dernier lead.")
			return
		}
	}

	if err := h.Store.RemoveProjectMember(r.Context(), project.ID, userID); err != nil {
		slog.Error("remove project member", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/projects/"+strconv.FormatInt(project.ID, 10)+"?msg=Membre+retir%C3%A9", http.StatusSeeOther)
}

func (h *Projects) loadProject(w http.ResponseWriter, r *http.Request) (*Project, *User, string, bool) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return nil, nil, "", false
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return nil, nil, "", false
	}

	project, err := h.Store.ProjectByID(r.Context(), id)
	if errors.Is(err, ErrProjectNotFound) {
		http.NotFound(w, r)
		return nil, nil, "", false
	}
	if err != nil {
		slog.Error("load project", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, "", false
	}

	memberRole, isMember, err := h.Store.MemberRole(r.Context(), id, user.ID)
	if err != nil {
		slog.Error("member role", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, "", false
	}

	if !CanView(user, isMember) {
		http.NotFound(w, r)
		return nil, nil, "", false
	}

	return project, user, memberRole, true
}

func (h *Projects) renderFormError(w http.ResponseWriter, r *http.Request, project *Project, action, message string) {
	pd := h.PageDataTab(r, "Projet", "projects")
	pd.Breadcrumbs = []templates.Breadcrumb{
		{URL: "/projects", Label: "Projets"},
		{Label: "Nouveau projet"},
	}
	data := templates.ProjectFormData{
		PageData:   pd,
		Project:    project,
		FormAction: action,
		Error:      message,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	if err := h.Templates.ExecuteTemplate(w, "project_form", data); err != nil {
		slog.Error("render project form error", "err", err)
	}
}

func (h *Projects) renderShowError(w http.ResponseWriter, r *http.Request, project *Project, user *User, memberRole, callerOrgRole, message string) {
	members, err := h.Store.ListProjectMembers(r.Context(), project.ID)
	if err != nil {
		slog.Error("list project members", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	projectRuns, err := h.Store.ListRunsWithProgressByProject(r.Context(), project.ID)
	if err != nil {
		slog.Error("list project runs", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	nokItems, err := h.Store.ListProjectNokItems(r.Context(), project.ID)
	if err != nil {
		slog.Error("list project nok items", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	pd := h.PageDataTab(r, project.Name, "")
	pd.Breadcrumbs = []templates.Breadcrumb{
		{URL: "/projects", Label: "Projets"},
		{Label: project.Name},
	}
	data := templates.ProjectShowData{
		PageData:         pd,
		Project:          project,
		Members:          members,
		Runs:             projectRuns,
		NokItems:         nokItems,
		MemberRole:       memberRole,
		CanManage:        CanManage(user, memberRole),
		CanManageMembers: CanAddProjectMember(user, memberRole, callerOrgRole),
		Error:            message,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	if err := h.Templates.ExecuteTemplate(w, "project_show", data); err != nil {
		slog.Error("render project show error", "err", err)
	}
}

func normalizeMemberEmail(raw string) (string, error) {
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
