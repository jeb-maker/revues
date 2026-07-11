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

const defaultTemplateEditorRows = 1

// ChecklistTemplates handles versioned checklist model CRUD.
type ChecklistTemplates struct {
	Deps
	EncryptionKey []byte
	NotionClient  *notion.Client
}

// IndexAll lists all active global checklist templates.
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
		CanManage: CanManageGlobal(user),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "templates_index", data); err != nil {
		slog.Error("render templates index", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// List shows compatible checklist templates for a project (read-only).
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
		CanManage:  CanManageGlobal(user),
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
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if !CanManageGlobal(user) {
		http.NotFound(w, r)
		return
	}

	sections := emptyEditorSections(extraRows(r, defaultTemplateEditorRows))
	data := viewtemplates.ChecklistTemplateFormData{
		PageData:        h.PageDataTab(r, "Nouveau modèle", "templates"),
		Sections:        sections,
		SectionsEnabled: sectionsEnabled(sections),
		FormAction:      "/modeles",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "checklist_template_form", data); err != nil {
		slog.Error("render checklist template form", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Create stores a new global checklist template with version 1.
func (h *ChecklistTemplates) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if !CanManageGlobal(user) {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	tags := store.ParseTagsCSV(r.FormValue("tags"))
	items, itemErr := parseTemplateItems(r)
	nameErr, itemsErr := formFieldErrors(name, itemErr, len(items))
	if nameErr != "" || itemsErr != "" {
		h.renderFormError(w, r, nil, nil, "/modeles", name, store.FormatTagsCSV(tags), tags, nameErr, itemsErr)
		return
	}

	template, _, err := h.Store.CreateChecklistTemplate(r.Context(), name, user.ID, tags, items)
	if err != nil {
		slog.Error("create checklist template", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, templateShowURL(template.ID)+"?msg=Mod%C3%A8le+cr%C3%A9%C3%A9", http.StatusSeeOther)
}

// Show displays the latest template version.
func (h *ChecklistTemplates) Show(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	template, version, items, tags, ok := h.loadGlobalTemplate(w, r)
	if !ok {
		return
	}

	pd := h.PageDataTab(r, template.Name, "templates")
	pd.Breadcrumbs = []viewtemplates.Breadcrumb{
		{URL: "/modeles", Label: "Modèles"},
		{Label: template.Name},
	}
	data := viewtemplates.ChecklistTemplateShowData{
		PageData:     pd,
		Template:     template,
		Version:      version,
		Tags:         tags,
		ItemSections: groupTemplateItems(items),
		ItemCount:    len(items),
		CanManage:    CanManageGlobal(user),
		Message:      r.URL.Query().Get("msg"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "checklist_template_show", data); err != nil {
		slog.Error("render checklist template show", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// EditForm renders the template editor prefilled with the latest version.
func (h *ChecklistTemplates) EditForm(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if !CanManageGlobal(user) {
		http.NotFound(w, r)
		return
	}

	template, version, items, tags, ok := h.loadGlobalTemplate(w, r)
	if !ok {
		return
	}

	sections := itemsToEditorSections(items)

	data := viewtemplates.ChecklistTemplateFormData{
		PageData:        h.PageDataTab(r, "Modifier "+template.Name, "templates"),
		Template:        template,
		Version:         version,
		Tags:            store.FormatTagsCSV(tags),
		TagsList:        tags,
		Sections:        sections,
		SectionsEnabled: sectionsEnabled(sections),
		FormAction:      "/modeles/" + strconv.FormatInt(template.ID, 10),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "checklist_template_form", data); err != nil {
		slog.Error("render checklist template edit", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Save creates a new template version from submitted items.
func (h *ChecklistTemplates) Save(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if !CanManageGlobal(user) {
		http.NotFound(w, r)
		return
	}

	template, version, _, _, ok := h.loadGlobalTemplate(w, r)
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	tags := store.ParseTagsCSV(r.FormValue("tags"))
	items, itemErr := parseTemplateItems(r)
	action := "/modeles/" + strconv.FormatInt(template.ID, 10)
	nameErr, itemsErr := formFieldErrors(name, itemErr, len(items))
	if nameErr != "" || itemsErr != "" {
		h.renderFormError(w, r, template, version, action, name, store.FormatTagsCSV(tags), tags, nameErr, itemsErr)
		return
	}

	if err := h.Store.UpdateChecklistTemplateName(r.Context(), template.ID, name); err != nil {
		slog.Error("update checklist template name", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := h.Store.SetTemplateTags(r.Context(), template.ID, tags); err != nil {
		slog.Error("update template tags", "err", err)
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
	http.Redirect(w, r, templateShowURL(template.ID)+"?msg=Version+"+strconv.Itoa(newVersion.Version)+"+enregistr%C3%A9e", http.StatusSeeOther)
}

// Archive marks a checklist template archived.
func (h *ChecklistTemplates) Archive(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if !CanManageGlobal(user) {
		http.NotFound(w, r)
		return
	}

	template, _, _, _, ok := h.loadGlobalTemplate(w, r)
	if !ok {
		return
	}

	if err := h.Store.ArchiveChecklistTemplate(r.Context(), template.ID); err != nil {
		slog.Error("archive checklist template", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/modeles?msg=Mod%C3%A8le+archiv%C3%A9", http.StatusSeeOther)
}

// AddRow returns an empty template editor row fragment for HTMX insertion.
func (h *ChecklistTemplates) AddRow(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if !CanManageGlobal(user) {
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

	templateID := int64(0)
	if tid := chi.URLParam(r, "tid"); tid != "" {
		if n, err := strconv.ParseInt(tid, 10, 64); err == nil {
			templateID = n
		}
	}

	data := viewtemplates.TemplateRowFragmentData{
		TemplateID: templateID,
		Index:      idx,
		CSRFToken:  csrf,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "template_row_fragment", data); err != nil {
		slog.Error("render template row fragment", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// DeleteRow removes a template editor row via HTMX (returns empty).
func (h *ChecklistTemplates) DeleteRow(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if !CanManageGlobal(user) {
		http.NotFound(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
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

func (h *ChecklistTemplates) loadGlobalTemplate(w http.ResponseWriter, r *http.Request) (*store.ChecklistTemplate, *store.TemplateVersion, []store.TemplateItem, []string, bool) {
	templateID, err := strconv.ParseInt(chi.URLParam(r, "tid"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return nil, nil, nil, nil, false
	}

	template, err := h.Store.ChecklistTemplateByID(r.Context(), templateID)
	if errors.Is(err, store.ErrChecklistTemplateNotFound) {
		http.NotFound(w, r)
		return nil, nil, nil, nil, false
	}
	if err != nil {
		slog.Error("load checklist template", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, nil, nil, false
	}

	if template.ArchivedAt.Valid {
		http.NotFound(w, r)
		return nil, nil, nil, nil, false
	}

	version, err := h.Store.LatestTemplateVersion(r.Context(), template.ID)
	if err != nil {
		slog.Error("load latest template version", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, nil, nil, false
	}

	items, err := h.Store.ListTemplateItems(r.Context(), version.ID)
	if err != nil {
		slog.Error("list template items", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, nil, nil, false
	}

	tags, err := h.Store.ListTemplateTags(r.Context(), template.ID)
	if err != nil {
		slog.Error("list template tags", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, nil, nil, false
	}

	return template, version, items, tags, true
}

func (h *ChecklistTemplates) renderFormError(w http.ResponseWriter, r *http.Request, template *store.ChecklistTemplate, version *store.TemplateVersion, action, name, tagsCSV string, tags []string, nameErr, itemsErr string) {
	sections := parseTemplateItemsToSections(r)
	if len(sections) == 0 {
		sections = emptyEditorSections(defaultTemplateEditorRows)
	}

	title := "Nouveau modèle"
	if template != nil {
		title = "Modifier " + template.Name
	}
	data := viewtemplates.ChecklistTemplateFormData{
		PageData:        h.PageDataTab(r, title, "templates"),
		Template:        template,
		Version:         version,
		Name:            name,
		Tags:            tagsCSV,
		TagsList:        tags,
		Sections:        sections,
		SectionsEnabled: sectionsEnabled(sections),
		FormAction:      action,
		NameError:       nameErr,
		ItemsError:      itemsErr,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	if err := h.Templates.ExecuteTemplate(w, "checklist_template_form", data); err != nil {
		slog.Error("render checklist template form error", "err", err)
	}
}

func templateShowURL(templateID int64) string {
	return "/modeles/" + strconv.FormatInt(templateID, 10)
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

func formFieldErrors(name, itemErr string, itemCount int) (nameErr, itemsErr string) {
	if strings.TrimSpace(name) == "" {
		nameErr = "Le nom est obligatoire."
	}
	if itemErr != "" {
		itemsErr = itemErr
	} else if itemCount == 0 {
		itemsErr = "Ajoutez au moins un point au modèle."
	}
	return nameErr, itemsErr
}
