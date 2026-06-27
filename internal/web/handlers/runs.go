package handlers

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/runs"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/middleware"
	viewtemplates "github.com/jeb-maker/revues/internal/web/templates"
)

// Runs handles review launch wizard and run lifecycle.
type Runs struct {
	Templates     *template.Template
	Store         *store.Store
	SessionSecret string
}

// WizardProjects is step 1: choose a project.
func (h *Runs) WizardProjects(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	admin := auth.HasMinRole(user.Role, auth.RoleAdmin)
	allProjects, err := h.Store.ListProjects(r.Context(), user.ID, admin)
	if err != nil {
		slog.Error("list projects for run wizard", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var launchProjects []store.Project
	for _, project := range allProjects {
		memberRole, isMember, err := h.Store.MemberRole(r.Context(), project.ID, user.ID)
		if err != nil {
			slog.Error("member role", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if runs.CanLaunch(user, memberRoleForLaunch(isMember, memberRole)) {
			launchProjects = append(launchProjects, project)
		}
	}

	data := viewtemplates.RunWizardProjectsData{
		PageData: h.pageData(r, "Lancer une revue"),
		Projects: launchProjects,
		Step:     1,
		Message:  r.URL.Query().Get("msg"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "run_wizard_projects", data); err != nil {
		slog.Error("render run wizard step 1", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// WizardTemplates is step 2: choose a template for the project.
func (h *Runs) WizardTemplates(w http.ResponseWriter, r *http.Request) {
	project, user, memberRole, ok := h.loadProjectForLaunch(w, r)
	if !ok {
		return
	}

	templates, err := h.Store.ListChecklistTemplates(r.Context(), project.ID)
	if err != nil {
		slog.Error("list templates for run wizard", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := viewtemplates.RunWizardTemplatesData{
		PageData:   h.pageData(r, "Choisir un modèle"),
		Project:    project,
		Templates:  templates,
		Step:       2,
		MemberRole: memberRole,
		CanLaunch:  runs.CanLaunch(user, memberRole),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "run_wizard_templates", data); err != nil {
		slog.Error("render run wizard step 2", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// WizardLaunch is step 3: confirm title and launch.
func (h *Runs) WizardLaunch(w http.ResponseWriter, r *http.Request) {
	project, user, memberRole, ok := h.loadProjectForLaunch(w, r)
	if !ok {
		return
	}

	templateID, err := strconv.ParseInt(chi.URLParam(r, "tid"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	template, err := h.Store.ChecklistTemplateByID(r.Context(), templateID)
	if errors.Is(err, store.ErrChecklistTemplateNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		slog.Error("load template for run wizard", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if template.ProjectID != project.ID || template.ArchivedAt.Valid {
		http.NotFound(w, r)
		return
	}

	version, err := h.Store.LatestTemplateVersion(r.Context(), template.ID)
	if err != nil {
		slog.Error("load latest template version", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	items, err := h.Store.ListTemplateItems(r.Context(), version.ID)
	if err != nil {
		slog.Error("list template items for wizard", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := viewtemplates.RunWizardLaunchData{
		PageData:   h.pageData(r, "Lancer la revue"),
		Project:    project,
		Template:   template,
		Version:    version,
		ItemCount:  len(items),
		FormAction: "/projects/" + strconv.FormatInt(project.ID, 10) + "/runs",
		Step:       3,
		MemberRole: memberRole,
		CanLaunch:  runs.CanLaunch(user, memberRole),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "run_wizard_launch", data); err != nil {
		slog.Error("render run wizard step 3", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Create stores a new run with item snapshot.
func (h *Runs) Create(w http.ResponseWriter, r *http.Request) {
	project, user, memberRole, ok := h.loadProjectForLaunch(w, r)
	if !ok {
		return
	}
	if !runs.CanLaunch(user, memberRole) {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	templateID, err := strconv.ParseInt(r.FormValue("template_id"), 10, 64)
	if err != nil {
		h.renderLaunchError(w, r, project, nil, nil, 0, "Modèle invalide.")
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		h.renderLaunchError(w, r, project, nil, nil, templateID, "Le titre est obligatoire.")
		return
	}

	template, err := h.Store.ChecklistTemplateByID(r.Context(), templateID)
	if errors.Is(err, store.ErrChecklistTemplateNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		slog.Error("load template for run create", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if template.ProjectID != project.ID || template.ArchivedAt.Valid {
		http.NotFound(w, r)
		return
	}

	run, err := h.Store.CreateChecklistRun(r.Context(), project.ID, template.ID, title, user.ID)
	if err != nil {
		slog.Error("create checklist run", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/runs/"+strconv.FormatInt(run.ID, 10)+"?msg=Revue+cr%C3%A9%C3%A9e", http.StatusSeeOther)
}

// Show displays run detail and snapshot items.
func (h *Runs) Show(w http.ResponseWriter, r *http.Request) {
	run, project, user, memberRole, ok := h.loadRun(w, r)
	if !ok {
		return
	}

	items, err := h.Store.ListRunItems(r.Context(), run.ID)
	if err != nil {
		slog.Error("list run items", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	versionInfo, err := h.Store.TemplateVersionInfo(r.Context(), run.TemplateVersionID)
	if err != nil {
		slog.Error("template version info", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := viewtemplates.RunShowData{
		PageData:     h.pageData(r, run.Title),
		Project:      project,
		Run:          run,
		Items:        items,
		TemplateName: versionInfo.Name,
		VersionNum:   versionInfo.Version,
		MemberRole:   memberRole,
		CanLaunch:    runs.CanLaunch(user, memberRole),
		CanComplete:  runs.CanComplete(user, memberRole),
		Message:      r.URL.Query().Get("msg"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "run_show", data); err != nil {
		slog.Error("render run show", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Start moves a run from draft to in_progress.
func (h *Runs) Start(w http.ResponseWriter, r *http.Request) {
	run, _, user, memberRole, ok := h.loadRun(w, r)
	if !ok {
		return
	}
	if !runs.CanLaunch(user, memberRole) {
		http.NotFound(w, r)
		return
	}

	if err := h.Store.StartRun(r.Context(), run.ID); err != nil {
		if errors.Is(err, store.ErrInvalidRunStatus) {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		slog.Error("start run", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/runs/"+strconv.FormatInt(run.ID, 10)+"?msg=Revue+d%C3%A9marr%C3%A9e", http.StatusSeeOther)
}

// Complete moves a run from in_progress to done.
func (h *Runs) Complete(w http.ResponseWriter, r *http.Request) {
	run, _, user, memberRole, ok := h.loadRun(w, r)
	if !ok {
		return
	}
	if !runs.CanComplete(user, memberRole) {
		http.NotFound(w, r)
		return
	}

	if err := h.Store.CompleteRun(r.Context(), run.ID); err != nil {
		if errors.Is(err, store.ErrInvalidRunStatus) {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		slog.Error("complete run", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/runs/"+strconv.FormatInt(run.ID, 10)+"?msg=Revue+termin%C3%A9e", http.StatusSeeOther)
}

func (h *Runs) loadProjectForLaunch(w http.ResponseWriter, r *http.Request) (*store.Project, *store.User, string, bool) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return nil, nil, "", false
	}

	projectID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return nil, nil, "", false
	}

	project, err := h.Store.ProjectByID(r.Context(), projectID)
	if errors.Is(err, store.ErrProjectNotFound) {
		http.NotFound(w, r)
		return nil, nil, "", false
	}
	if err != nil {
		slog.Error("load project", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, "", false
	}

	memberRole, isMember, err := h.Store.MemberRole(r.Context(), projectID, user.ID)
	if err != nil {
		slog.Error("member role", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, "", false
	}

	if !runs.CanLaunch(user, memberRoleForLaunch(isMember, memberRole)) {
		http.NotFound(w, r)
		return nil, nil, "", false
	}

	return project, user, memberRole, true
}

func (h *Runs) loadRun(w http.ResponseWriter, r *http.Request) (*store.ChecklistRun, *store.Project, *store.User, string, bool) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return nil, nil, nil, "", false
	}

	runID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return nil, nil, nil, "", false
	}

	run, err := h.Store.RunByID(r.Context(), runID)
	if errors.Is(err, store.ErrRunNotFound) {
		http.NotFound(w, r)
		return nil, nil, nil, "", false
	}
	if err != nil {
		slog.Error("load run", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, nil, "", false
	}

	project, err := h.Store.ProjectByID(r.Context(), run.ProjectID)
	if errors.Is(err, store.ErrProjectNotFound) {
		http.NotFound(w, r)
		return nil, nil, nil, "", false
	}
	if err != nil {
		slog.Error("load run project", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, nil, "", false
	}

	memberRole, isMember, err := h.Store.MemberRole(r.Context(), project.ID, user.ID)
	if err != nil {
		slog.Error("member role", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, nil, "", false
	}

	if !runs.CanView(user, isMember) {
		http.NotFound(w, r)
		return nil, nil, nil, "", false
	}

	return run, project, user, memberRole, true
}

func (h *Runs) renderLaunchError(w http.ResponseWriter, r *http.Request, project *store.Project, template *store.ChecklistTemplate, version *store.TemplateVersion, templateID int64, message string) {
	if template == nil && templateID > 0 {
		var err error
		template, err = h.Store.ChecklistTemplateByID(r.Context(), templateID)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		version, err = h.Store.LatestTemplateVersion(r.Context(), template.ID)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
	}

	itemCount := 0
	if version != nil {
		items, err := h.Store.ListTemplateItems(r.Context(), version.ID)
		if err == nil {
			itemCount = len(items)
		}
	}

	data := viewtemplates.RunWizardLaunchData{
		PageData:   h.pageData(r, "Lancer la revue"),
		Project:    project,
		Template:   template,
		Version:    version,
		ItemCount:  itemCount,
		Title:      strings.TrimSpace(r.FormValue("title")),
		FormAction: "/projects/" + strconv.FormatInt(project.ID, 10) + "/runs",
		Step:       3,
		Error:      message,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	if err := h.Templates.ExecuteTemplate(w, "run_wizard_launch", data); err != nil {
		slog.Error("render run wizard launch error", "err", err)
	}
}

func (h *Runs) pageData(r *http.Request, title string) viewtemplates.PageData {
	data := viewtemplates.PageData{Title: title}
	if user, ok := middleware.UserFromContext(r.Context()); ok {
		data.User = user
		if token := middleware.SessionTokenFromContext(r); token != "" {
			data.CSRFToken = auth.CSRFToken(token, h.SessionSecret)
		}
	}
	return data
}

func memberRoleForLaunch(isMember bool, memberRole string) string {
	if !isMember {
		return ""
	}
	return memberRole
}
