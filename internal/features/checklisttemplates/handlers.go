package checklisttemplates

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/integrations/notion"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/middleware"
	viewtemplates "github.com/jeb-maker/revues/internal/web/templates"
)

// Deps holds shared dependencies for the checklisttemplates HTTP handlers.
//
// This mirrors internal/web/handlerdeps.HandlerDeps but is local to the
// checklisttemplates feature package to avoid an import cycle.
type Deps struct {
	Templates     *template.Template
	Store         ChecklistTemplateStore
	SessionSecret string
}

// PageData builds shared view data with user and CSRF from the request context.
func (d *Deps) PageData(r *http.Request, title string) viewtemplates.PageData {
	data := viewtemplates.PageData{Title: title}
	if user, ok := middleware.UserFromContext(r.Context()); ok {
		data.User = user
		if token := middleware.SessionTokenFromContext(r); token != "" {
			data.CSRFToken = auth.CSRFToken(token, d.SessionSecret)
		}
	}
	viewtemplates.ApplyHeaderFromContext(r, &data)
	return data
}

// PageDataTab is PageData with ActiveTab set.
func (d *Deps) PageDataTab(r *http.Request, title, activeTab string) viewtemplates.PageData {
	data := d.PageData(r, title)
	data.ActiveTab = activeTab
	return data
}

const defaultTemplateEditorRows = 3

// ChecklistTemplates handles versioned checklist model CRUD.
type ChecklistTemplates struct {
	Deps
	EncryptionKey []byte
	NotionClient  *notion.Client
}

// NewWizard shows a project picker before creating a new template.
func (h *ChecklistTemplates) NewWizard(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	admin := auth.HasMinRole(user.Role, auth.RoleAdmin)
	allProjects, err := h.Store.ListProjects(r.Context(), user.ID, admin)
	if err != nil {
		slog.Error("list projects for wizard", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var manageProjects []store.Project
	for _, project := range allProjects {
		memberRole, _, err := h.Store.MemberRole(r.Context(), project.ID, user.ID)
		if err != nil {
			slog.Error("member role", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if CanManage(user, memberRole) {
			manageProjects = append(manageProjects, project)
		}
	}

	pd := h.PageDataTab(r, "Nouveau modèle", "templates")
	pd.Breadcrumbs = []viewtemplates.Breadcrumb{
		{URL: "/modeles", Label: "Modèles"},
		{Label: "Nouveau modèle"},
	}
	data := viewtemplates.TemplateNewWizardData{
		PageData: pd,
		Projects: manageProjects,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "templates_new_wizard", data); err != nil {
		slog.Error("render template new wizard", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// AddRow returns an empty template editor row fragment for HTMX insertion.
func (h *ChecklistTemplates) AddRow(w http.ResponseWriter, r *http.Request) {
	project, user, memberRole, ok := h.loadProject(w, r)
	if !ok {
		return
	}
	if !CanManage(user, memberRole) {
		http.NotFound(w, r)
		return
	}

	idx := 0
	if raw := strings.TrimSpace(r.FormValue("idx")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			idx = n + 1
		}
	}

	csrf := ""
	if token := middleware.SessionTokenFromContext(r); token != "" {
		csrf = auth.CSRFToken(token, h.SessionSecret)
	}

	data := viewtemplates.TemplateRowFragmentData{
		ProjectID: project.ID,
		Index:     idx,
		CSRFToken: csrf,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "template_row_fragment", data); err != nil {
		slog.Error("render template row fragment", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// DeleteRow removes a template editor row via HTMX (returns empty).
func (h *ChecklistTemplates) DeleteRow(w http.ResponseWriter, r *http.Request) {
	_, user, memberRole, ok := h.loadProject(w, r)
	if !ok {
		return
	}
	if !CanManage(user, memberRole) {
		http.NotFound(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// IndexAll lists checklist templates across visible projects.
func (h *ChecklistTemplates) IndexAll(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	admin := auth.HasMinRole(user.Role, auth.RoleAdmin)
	rows, err := h.Store.ListTemplateIndex(r.Context(), user.ID, admin)
	if err != nil {
		slog.Error("list template index", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	pd := h.PageDataTab(r, "Modèles", "templates")
	pd.Breadcrumbs = []viewtemplates.Breadcrumb{
		{Label: "Modèles"},
	}
	data := viewtemplates.TemplatesIndexData{
		PageData:  pd,
		Templates: rows,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "templates_index", data); err != nil {
		slog.Error("render templates index", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// List shows checklist templates for a project.
func (h *ChecklistTemplates) List(w http.ResponseWriter, r *http.Request) {
	project, user, memberRole, ok := h.loadProject(w, r)
	if !ok {
		return
	}

	items, err := h.Store.ListChecklistTemplates(r.Context(), project.ID)
	if err != nil {
		slog.Error("list checklist templates", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	pd := h.PageDataTab(r, "Modèles — "+project.Name, "templates")
	pd.Breadcrumbs = []viewtemplates.Breadcrumb{
		{URL: "/projects", Label: "Projets"},
		{URL: "/projects/" + strconv.FormatInt(project.ID, 10), Label: project.Name},
		{Label: "Modèles"},
	}
	data := viewtemplates.ChecklistTemplatesListData{
		PageData:   pd,
		Project:    project,
		Templates:  items,
		MemberRole: memberRole,
		CanManage:  CanManage(user, memberRole),
		Message:    r.URL.Query().Get("msg"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "checklist_templates_list", data); err != nil {
		slog.Error("render checklist templates list", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// NewForm renders the create template editor.
func (h *ChecklistTemplates) NewForm(w http.ResponseWriter, r *http.Request) {
	project, user, memberRole, ok := h.loadProject(w, r)
	if !ok {
		return
	}
	if !CanManage(user, memberRole) {
		http.NotFound(w, r)
		return
	}

	pd2 := h.PageDataTab(r, "Nouveau modèle", "")
	pd2.Breadcrumbs = []viewtemplates.Breadcrumb{
		{URL: "/projects", Label: "Projets"},
		{URL: "/projects/" + strconv.FormatInt(project.ID, 10), Label: project.Name},
		{URL: "/projects/" + strconv.FormatInt(project.ID, 10) + "/templates", Label: "Modèles"},
		{Label: "Nouveau modèle"},
	}
	data := viewtemplates.ChecklistTemplateFormData{
		PageData:   pd2,
		Project:    project,
		Rows:       emptyEditorRows(extraRows(r, defaultTemplateEditorRows)),
		FormAction: "/projects/" + strconv.FormatInt(project.ID, 10) + "/templates",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "checklist_template_form", data); err != nil {
		slog.Error("render checklist template form", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Create stores a new checklist template with version 1.
func (h *ChecklistTemplates) Create(w http.ResponseWriter, r *http.Request) {
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
	items, itemErr := parseTemplateItems(r)
	if name == "" || itemErr != "" || len(items) == 0 {
		h.renderFormError(w, r, project, nil, nil, "/projects/"+strconv.FormatInt(project.ID, 10)+"/templates", formValidationMessage(name, itemErr, len(items)))
		return
	}

	template, _, err := h.Store.CreateChecklistTemplate(r.Context(), project.ID, name, user.ID, items)
	if err != nil {
		slog.Error("create checklist template", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, templateShowURL(project.ID, template.ID)+"?msg=Mod%C3%A8le+cr%C3%A9%C3%A9", http.StatusSeeOther)
}

// Show displays the latest template version.
func (h *ChecklistTemplates) Show(w http.ResponseWriter, r *http.Request) {
	project, template, version, items, user, memberRole, ok := h.loadTemplate(w, r)
	if !ok {
		return
	}

	pd := h.PageDataTab(r, template.Name, "")
	pd.Breadcrumbs = []viewtemplates.Breadcrumb{
		{URL: "/projects", Label: "Projets"},
		{URL: "/projects/" + strconv.FormatInt(project.ID, 10), Label: project.Name},
		{URL: "/projects/" + strconv.FormatInt(project.ID, 10) + "/templates", Label: "Modèles"},
		{Label: template.Name},
	}
	data := viewtemplates.ChecklistTemplateShowData{
		PageData:   pd,
		Project:    project,
		Template:   template,
		Version:    version,
		Items:      items,
		MemberRole: memberRole,
		CanManage:  CanManage(user, memberRole),
		Message:    r.URL.Query().Get("msg"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "checklist_template_show", data); err != nil {
		slog.Error("render checklist template show", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// EditForm renders the template editor prefilled with the latest version.
func (h *ChecklistTemplates) EditForm(w http.ResponseWriter, r *http.Request) {
	project, template, version, items, user, memberRole, ok := h.loadTemplate(w, r)
	if !ok {
		return
	}
	if !CanManage(user, memberRole) {
		http.NotFound(w, r)
		return
	}

	rows := itemsToEditorRows(items)
	rows = append(rows, emptyEditorRows(extraRows(r, 2))...)

	pd := h.PageDataTab(r, "Modifier "+template.Name, "")
	pd.Breadcrumbs = []viewtemplates.Breadcrumb{
		{URL: "/projects", Label: "Projets"},
		{URL: "/projects/" + strconv.FormatInt(project.ID, 10), Label: project.Name},
		{URL: "/projects/" + strconv.FormatInt(project.ID, 10) + "/templates", Label: "Modèles"},
		{URL: "/projects/" + strconv.FormatInt(project.ID, 10) + "/templates/" + strconv.FormatInt(template.ID, 10), Label: template.Name},
		{Label: "Modifier"},
	}
	data := viewtemplates.ChecklistTemplateFormData{
		PageData:   pd,
		Project:    project,
		Template:   template,
		Version:    version,
		Rows:       rows,
		FormAction: "/projects/" + strconv.FormatInt(project.ID, 10) + "/templates/" + strconv.FormatInt(template.ID, 10),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "checklist_template_form", data); err != nil {
		slog.Error("render checklist template edit", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Save creates a new template version from submitted items.
func (h *ChecklistTemplates) Save(w http.ResponseWriter, r *http.Request) {
	project, template, version, _, user, memberRole, ok := h.loadTemplate(w, r)
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
	items, itemErr := parseTemplateItems(r)
	if name == "" || itemErr != "" || len(items) == 0 {
		h.renderFormError(w, r, project, template, version, "/projects/"+strconv.FormatInt(project.ID, 10)+"/templates/"+strconv.FormatInt(template.ID, 10), formValidationMessage(name, itemErr, len(items)))
		return
	}

	if err := h.Store.UpdateChecklistTemplateName(r.Context(), template.ID, name); err != nil {
		slog.Error("update checklist template name", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	newVersion, err := h.Store.CreateTemplateVersion(r.Context(), template.ID, user.ID, items)
	if err != nil {
		slog.Error("create template version", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_ = version
	http.Redirect(w, r, templateShowURL(project.ID, template.ID)+"?msg=Version+"+strconv.Itoa(newVersion.Version)+"+enregistr%C3%A9e", http.StatusSeeOther)
}

// Archive marks a checklist template archived.
func (h *ChecklistTemplates) Archive(w http.ResponseWriter, r *http.Request) {
	project, template, _, _, user, memberRole, ok := h.loadTemplate(w, r)
	if !ok {
		return
	}
	if !CanManage(user, memberRole) {
		http.NotFound(w, r)
		return
	}

	if err := h.Store.ArchiveChecklistTemplate(r.Context(), template.ID); err != nil {
		slog.Error("archive checklist template", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/projects/"+strconv.FormatInt(project.ID, 10)+"/templates?msg=Mod%C3%A8le+archiv%C3%A9", http.StatusSeeOther)
}

func (h *ChecklistTemplates) loadProject(w http.ResponseWriter, r *http.Request) (*store.Project, *store.User, string, bool) {
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

	if !CanView(user, isMember) {
		http.NotFound(w, r)
		return nil, nil, "", false
	}

	return project, user, memberRole, true
}

func (h *ChecklistTemplates) loadTemplate(w http.ResponseWriter, r *http.Request) (*store.Project, *store.ChecklistTemplate, *store.TemplateVersion, []store.TemplateItem, *store.User, string, bool) {
	project, user, memberRole, ok := h.loadProject(w, r)
	if !ok {
		return nil, nil, nil, nil, nil, "", false
	}

	templateID, err := strconv.ParseInt(chi.URLParam(r, "tid"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return nil, nil, nil, nil, nil, "", false
	}

	template, err := h.Store.ChecklistTemplateByID(r.Context(), templateID)
	if errors.Is(err, store.ErrChecklistTemplateNotFound) {
		http.NotFound(w, r)
		return nil, nil, nil, nil, nil, "", false
	}
	if err != nil {
		slog.Error("load checklist template", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, nil, nil, nil, "", false
	}

	if template.ProjectID != project.ID {
		http.NotFound(w, r)
		return nil, nil, nil, nil, nil, "", false
	}
	if template.ArchivedAt.Valid {
		http.NotFound(w, r)
		return nil, nil, nil, nil, nil, "", false
	}

	version, err := h.Store.LatestTemplateVersion(r.Context(), template.ID)
	if err != nil {
		slog.Error("load latest template version", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, nil, nil, nil, "", false
	}

	items, err := h.Store.ListTemplateItems(r.Context(), version.ID)
	if err != nil {
		slog.Error("list template items", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, nil, nil, nil, "", false
	}

	return project, template, version, items, user, memberRole, true
}

func (h *ChecklistTemplates) renderFormError(w http.ResponseWriter, r *http.Request, project *store.Project, template *store.ChecklistTemplate, version *store.TemplateVersion, action, message string) {
	rows := parseTemplateItemsToRows(r)
	if len(rows) == 0 {
		rows = emptyEditorRows(defaultTemplateEditorRows)
	}

	pd := h.PageDataTab(r, "Modèle", "")
	pd.Breadcrumbs = []viewtemplates.Breadcrumb{
		{URL: "/projects", Label: "Projets"},
		{URL: "/projects/" + strconv.FormatInt(project.ID, 10), Label: project.Name},
		{Label: "Nouveau modèle"},
	}
	data := viewtemplates.ChecklistTemplateFormData{
		PageData:   pd,
		Project:    project,
		Template:   template,
		Version:    version,
		Name:       strings.TrimSpace(r.FormValue("name")),
		Rows:       rows,
		FormAction: action,
		Error:      message,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	if err := h.Templates.ExecuteTemplate(w, "checklist_template_form", data); err != nil {
		slog.Error("render checklist template form error", "err", err)
	}
}

func templateShowURL(projectID, templateID int64) string {
	return "/projects/" + strconv.FormatInt(projectID, 10) + "/templates/" + strconv.FormatInt(templateID, 10)
}

func extraRows(r *http.Request, fallback int) int {
	raw := strings.TrimSpace(r.URL.Query().Get("rows"))
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return fallback
	}
	if n > 50 {
		return 50
	}
	return n
}

func emptyEditorRows(count int) []viewtemplates.TemplateEditorRow {
	rows := make([]viewtemplates.TemplateEditorRow, count)
	return rows
}

func itemsToEditorRows(items []store.TemplateItem) []viewtemplates.TemplateEditorRow {
	rows := make([]viewtemplates.TemplateEditorRow, len(items))
	for i, item := range items {
		rows[i] = viewtemplates.TemplateEditorRow{
			Section:  item.Section,
			Label:    item.Label,
			HelpText: item.HelpText,
			Required: item.Required,
		}
	}
	return rows
}

func parseTemplateItems(r *http.Request) ([]store.TemplateItemInput, string) {
	sections := r.Form["item_section"]
	labels := r.Form["item_label"]
	helps := r.Form["item_help"]
	rowIndices := r.Form["item_row_idx"]
	if len(sections) != len(labels) || len(labels) != len(helps) {
		return nil, "Les lignes du modèle sont incohérentes."
	}
	if len(rowIndices) != 0 && len(rowIndices) != len(labels) {
		return nil, "Les lignes du modèle sont incohérentes."
	}

	required := map[int]bool{}
	for _, raw := range r.Form["item_required"] {
		index, err := strconv.Atoi(raw)
		if err != nil || index < 0 {
			return nil, "Point requis invalide."
		}
		required[index] = true
	}

	useRowIdx := len(rowIndices) == len(labels)

	var items []store.TemplateItemInput
	for i := range labels {
		label := strings.TrimSpace(labels[i])
		if label == "" {
			continue
		}
		isRequired := false
		if useRowIdx {
			if idx, err := strconv.Atoi(rowIndices[i]); err == nil {
				isRequired = required[idx]
			}
		} else {
			isRequired = required[i]
		}
		items = append(items, store.TemplateItemInput{
			Section:  strings.TrimSpace(sections[i]),
			Label:    label,
			HelpText: strings.TrimSpace(helps[i]),
			Required: isRequired,
		})
	}

	return items, ""
}

func parseTemplateItemsToRows(r *http.Request) []viewtemplates.TemplateEditorRow {
	sections := r.Form["item_section"]
	labels := r.Form["item_label"]
	helps := r.Form["item_help"]
	rowIndices := r.Form["item_row_idx"]
	maxLen := len(labels)
	if len(sections) > maxLen {
		maxLen = len(sections)
	}
	if len(helps) > maxLen {
		maxLen = len(helps)
	}

	required := map[int]bool{}
	for _, raw := range r.Form["item_required"] {
		index, err := strconv.Atoi(raw)
		if err == nil {
			required[index] = true
		}
	}

	useRowIdx := len(rowIndices) == maxLen

	rows := make([]viewtemplates.TemplateEditorRow, maxLen)
	for i := 0; i < maxLen; i++ {
		if i < len(sections) {
			rows[i].Section = sections[i]
		}
		if i < len(labels) {
			rows[i].Label = labels[i]
		}
		if i < len(helps) {
			rows[i].HelpText = helps[i]
		}
		if useRowIdx {
			if idx, err := strconv.Atoi(rowIndices[i]); err == nil {
				rows[i].Required = required[idx]
			}
		} else {
			rows[i].Required = required[i]
		}
	}
	return rows
}

func formValidationMessage(name, itemErr string, itemCount int) string {
	switch {
	case name == "":
		return "Le nom est obligatoire."
	case itemErr != "":
		return itemErr
	case itemCount == 0:
		return "Ajoutez au moins un point au modèle."
	default:
		return "Formulaire invalide."
	}
}
