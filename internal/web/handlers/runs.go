package handlers

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/items"
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

	dueDateRaw, err := runs.ParseDueDate(r.FormValue("due_date"))
	if err != nil {
		h.renderLaunchError(w, r, project, nil, nil, templateID, "Échéance invalide.")
		return
	}
	var dueDate sql.NullString
	if dueDateRaw != "" {
		dueDate = sql.NullString{String: dueDateRaw, Valid: true}
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

	run, err := h.Store.CreateChecklistRun(r.Context(), project.ID, template.ID, title, user.ID, dueDate)
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

	h.renderRunShow(w, r, run, project, user, memberRole, viewtemplates.RunShowData{
		Message:   r.URL.Query().Get("msg"),
		ItemError: r.URL.Query().Get("item_error"),
	})
}

// UpdateItem changes status and comment on a run item.
func (h *Runs) UpdateItem(w http.ResponseWriter, r *http.Request) {
	run, project, user, memberRole, ok := h.loadRun(w, r)
	if !ok {
		return
	}
	if !items.CanUpdate(user, memberRole) {
		http.NotFound(w, r)
		return
	}
	if run.Status != store.RunStatusInProgress {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	itemID, err := strconv.ParseInt(chi.URLParam(r, "itemId"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	status := strings.TrimSpace(r.FormValue("status"))
	comment := strings.TrimSpace(r.FormValue("comment"))

	if err := items.ValidateUpdate(status, comment); err != nil {
		switch {
		case errors.Is(err, items.ErrCommentRequired):
			if h.isHTMX(r) {
				h.renderRunItemHTMXError(w, r, run, project, user, memberRole, itemID, "Un commentaire est obligatoire pour le statut nok.", "")
				return
			}
			h.renderRunShow(w, r, run, project, user, memberRole, viewtemplates.RunShowData{
				ItemError: "Un commentaire est obligatoire pour le statut nok.",
			})
		case errors.Is(err, items.ErrInvalidStatus):
			if h.isHTMX(r) {
				h.renderRunItemHTMXError(w, r, run, project, user, memberRole, itemID, "Statut invalide.", "")
				return
			}
			h.renderRunShow(w, r, run, project, user, memberRole, viewtemplates.RunShowData{
				ItemError: "Statut invalide.",
			})
		default:
			http.Error(w, "Bad Request", http.StatusBadRequest)
		}
		return
	}

	if err := h.Store.UpdateRunItemStatus(r.Context(), run.ID, itemID, user.ID, status, comment); err != nil {
		if errors.Is(err, store.ErrRunItemNotFound) {
			http.NotFound(w, r)
			return
		}
		slog.Error("update run item", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if h.isHTMX(r) {
		h.renderRunItemHTMXSuccess(w, r, run, project, user, memberRole, itemID, "", "")
		return
	}

	http.Redirect(w, r, "/runs/"+strconv.FormatInt(run.ID, 10)+"?msg=Point+mis+%C3%A0+jour", http.StatusSeeOther)
}

// ShowItem displays a run item and its status change history.
func (h *Runs) ShowItem(w http.ResponseWriter, r *http.Request) {
	run, project, user, memberRole, ok := h.loadRun(w, r)
	if !ok {
		return
	}

	itemID, err := strconv.ParseInt(chi.URLParam(r, "itemId"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	item, err := h.Store.RunItemByID(r.Context(), run.ID, itemID)
	if errors.Is(err, store.ErrRunItemNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		slog.Error("load run item", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	events, err := h.Store.ListRunItemEvents(r.Context(), item.ID)
	if err != nil {
		slog.Error("list run item events", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := viewtemplates.RunItemShowData{
		PageData:   h.pageData(r, item.Label),
		Project:    project,
		Run:        run,
		Item:       item,
		Events:     events,
		MemberRole: memberRole,
		CanCheck:   items.CanUpdate(user, memberRole),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "run_item_show", data); err != nil {
		slog.Error("render run item show", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// AssignItem sets or clears assignee on a run item.
func (h *Runs) AssignItem(w http.ResponseWriter, r *http.Request) {
	run, project, user, memberRole, ok := h.loadRun(w, r)
	if !ok {
		return
	}
	if !items.CanAssign(user, memberRole) {
		http.NotFound(w, r)
		return
	}
	if run.Status != store.RunStatusInProgress {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	itemID, err := strconv.ParseInt(chi.URLParam(r, "itemId"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var assigneeID *int64
	if raw := strings.TrimSpace(r.FormValue("assignee_id")); raw != "" {
		id, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr != nil {
			if h.isHTMX(r) {
				h.renderRunItemHTMXError(w, r, run, project, user, memberRole, itemID, "", "Assigné invalide.")
				return
			}
			h.renderRunShow(w, r, run, project, user, memberRole, viewtemplates.RunShowData{
				AssignError: "Assigné invalide.",
			})
			return
		}
		assigneeID = &id
	}

	if err := h.Store.AssignRunItem(r.Context(), run.ID, itemID, assigneeID); err != nil {
		if errors.Is(err, store.ErrInvalidAssignee) {
			if h.isHTMX(r) {
				h.renderRunItemHTMXError(w, r, run, project, user, memberRole, itemID, "", "Le membre doit appartenir au projet.")
				return
			}
			h.renderRunShow(w, r, run, project, user, memberRole, viewtemplates.RunShowData{
				AssignError: "Le membre doit appartenir au projet.",
			})
			return
		}
		if errors.Is(err, store.ErrRunItemNotFound) {
			http.NotFound(w, r)
			return
		}
		slog.Error("assign run item", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if h.isHTMX(r) {
		h.renderRunItemHTMXSuccess(w, r, run, project, user, memberRole, itemID, "", "")
		return
	}

	http.Redirect(w, r, "/runs/"+strconv.FormatInt(run.ID, 10)+"?msg=Assignation+enregistr%C3%A9e", http.StatusSeeOther)
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
	run, project, user, memberRole, ok := h.loadRun(w, r)
	if !ok {
		return
	}
	if !runs.CanComplete(user, memberRole) {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	closingNote := strings.TrimSpace(r.FormValue("closing_note"))
	if closingNote == "" {
		h.renderRunShow(w, r, run, project, user, memberRole, viewtemplates.RunShowData{
			ClosingNote:   r.FormValue("closing_note"),
			CompleteError: "La note de clôture est obligatoire.",
		})
		return
	}

	if err := h.Store.CompleteRun(r.Context(), run.ID, closingNote); err != nil {
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

// ExportCSV downloads a CSV export for a completed run.
func (h *Runs) ExportCSV(w http.ResponseWriter, r *http.Request) {
	run, _, _, _, ok := h.loadRun(w, r)
	if !ok {
		return
	}
	if run.Status != store.RunStatusDone {
		http.NotFound(w, r)
		return
	}

	rows, err := h.Store.ListRunExportRows(r.Context(), run.ID)
	if err != nil {
		slog.Error("list run export rows", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	csvData, err := runs.BuildRunCSV(rows)
	if err != nil {
		slog.Error("build run csv", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	filename := exportCSVFilename(run)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	if _, writeErr := w.Write(csvData); writeErr != nil {
		slog.Error("write run csv export", "err", writeErr)
	}
}

func exportCSVFilename(run *store.ChecklistRun) string {
	safe := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			return r
		case r == '-', r == '_':
			return r
		default:
			return '-'
		}
	}, strings.TrimSpace(run.Title))
	if safe == "" {
		return fmt.Sprintf("revue-%d.csv", run.ID)
	}
	return safe + ".csv"
}

func (h *Runs) renderRunShow(w http.ResponseWriter, r *http.Request, run *store.ChecklistRun, project *store.Project, user *store.User, memberRole string, extra viewtemplates.RunShowData) {
	runItems, err := h.Store.ListRunItems(r.Context(), run.ID)
	if err != nil {
		slog.Error("list run items", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	nokItems, err := h.Store.ListNokRunItems(r.Context(), run.ID)
	if err != nil {
		slog.Error("list nok run items", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	versionInfo, err := h.Store.TemplateVersionInfo(r.Context(), run.TemplateVersionID)
	if err != nil {
		slog.Error("template version info", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var members []store.ProjectMember
	if items.CanAssign(user, memberRole) {
		members, err = h.Store.ListProjectMembers(r.Context(), project.ID)
		if err != nil {
			slog.Error("list project members", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data := viewtemplates.RunShowData{
		PageData:      h.pageData(r, run.Title),
		Project:       project,
		Run:           run,
		Items:         runItems,
		NokItems:      nokItems,
		Members:       members,
		TemplateName:  versionInfo.Name,
		VersionNum:    versionInfo.Version,
		MemberRole:    memberRole,
		CanLaunch:     runs.CanLaunch(user, memberRole),
		CanCheck:      items.CanUpdate(user, memberRole),
		CanAssign:     items.CanAssign(user, memberRole),
		CanComplete:   runs.CanComplete(user, memberRole),
		Progress:      h.progressData(run.ID, runItems),
		Message:       extra.Message,
		ItemError:     extra.ItemError,
		AssignError:   extra.AssignError,
		CompleteError: extra.CompleteError,
		ClosingNote:   extra.ClosingNote,
		Error:         extra.Error,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	statusCode := http.StatusOK
	if extra.ItemError != "" || extra.CompleteError != "" || extra.AssignError != "" {
		statusCode = http.StatusBadRequest
	}
	w.WriteHeader(statusCode)
	if err := h.Templates.ExecuteTemplate(w, "run_show", data); err != nil {
		slog.Error("render run show", "err", err)
	}
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
		DueDate:    strings.TrimSpace(r.FormValue("due_date")),
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

func (h *Runs) isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") != ""
}

func (h *Runs) progressData(runID int64, runItems []store.RunItem) viewtemplates.RunProgressData {
	done, total := items.Progress(runItems)
	percent := 0
	if total > 0 {
		percent = done * 100 / total
	}
	return viewtemplates.RunProgressData{
		RunID:   runID,
		Done:    done,
		Total:   total,
		Percent: percent,
	}
}

func (h *Runs) renderRunItemHTMXSuccess(w http.ResponseWriter, r *http.Request, run *store.ChecklistRun, project *store.Project, user *store.User, memberRole string, itemID int64, itemErr, assignErr string) {
	runItems, err := h.Store.ListRunItems(r.Context(), run.ID)
	if err != nil {
		slog.Error("list run items for htmx", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	item, ok := findRunItem(runItems, itemID)
	if !ok {
		http.NotFound(w, r)
		return
	}

	h.renderRunItemHTMX(w, r, run, project, user, memberRole, item, runItems, itemErr, assignErr, http.StatusOK)
}

func (h *Runs) renderRunItemHTMXError(w http.ResponseWriter, r *http.Request, run *store.ChecklistRun, project *store.Project, user *store.User, memberRole string, itemID int64, itemErr, assignErr string) {
	item, err := h.Store.RunItemByID(r.Context(), run.ID, itemID)
	if errors.Is(err, store.ErrRunItemNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		slog.Error("load run item for htmx error", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	runItems, err := h.Store.ListRunItems(r.Context(), run.ID)
	if err != nil {
		slog.Error("list run items for htmx error", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.renderRunItemHTMX(w, r, run, project, user, memberRole, *item, runItems, itemErr, assignErr, http.StatusBadRequest)
}

func (h *Runs) renderRunItemHTMX(w http.ResponseWriter, r *http.Request, run *store.ChecklistRun, project *store.Project, user *store.User, memberRole string, item store.RunItem, runItems []store.RunItem, itemErr, assignErr string, statusCode int) {
	var members []store.ProjectMember
	if items.CanAssign(user, memberRole) {
		var err error
		members, err = h.Store.ListProjectMembers(r.Context(), project.ID)
		if err != nil {
			slog.Error("list project members for htmx", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	row := viewtemplates.RunItemRowData{
		RunID:       run.ID,
		RunStatus:   run.Status,
		Item:        item,
		Members:     members,
		CSRFToken:   h.pageData(r, "").CSRFToken,
		CanCheck:    items.CanUpdate(user, memberRole),
		CanAssign:   items.CanAssign(user, memberRole),
		ItemError:   itemErr,
		AssignError: assignErr,
	}
	progress := h.progressData(run.ID, runItems)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)

	var buf bytes.Buffer
	if err := h.Templates.ExecuteTemplate(&buf, "run_item_row_fragment", row); err != nil {
		slog.Error("render run item row fragment", "err", err)
		return
	}
	if _, err := w.Write(buf.Bytes()); err != nil {
		slog.Error("write run item row fragment", "err", err)
		return
	}
	if err := h.Templates.ExecuteTemplate(w, "run_progress_oob_fragment", progress); err != nil {
		slog.Error("render run progress oob fragment", "err", err)
	}
}

func findRunItem(runItems []store.RunItem, itemID int64) (store.RunItem, bool) {
	for _, item := range runItems {
		if item.ID == itemID {
			return item, true
		}
	}
	return store.RunItem{}, false
}
