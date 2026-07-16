package checklisttemplates

import (
	"context"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/features/subjects"
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

const defaultTemplateEditorRows = 3

const queryForRun = "for_run"
const queryTemplate = "template"

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
	filterQuery := strings.TrimSpace(r.URL.Query().Get("q"))
	rows, err := h.Store.ListTemplateIndex(r.Context(), user.ID, admin, filterQuery)
	if err != nil {
		slog.Error("list template index", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	pd := h.PageDataTab(r, "Modèles", "templates")
	pd.Breadcrumbs = viewtemplates.BCTemplatesIndex()
	data := viewtemplates.TemplatesIndexData{
		PageData:         pd,
		Templates:        rows,
		FilterQuery:      filterQuery,
		HasActiveFilters: filterQuery != "",
		CanManage:        CanManageGlobal(user),
		Message:          r.URL.Query().Get("msg"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "templates_index", data); err != nil {
		slog.Error("render templates index", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// List shows compatible checklist templates for a subject (read-only).
// With ?for_run=1, lists templates as step 2 of the run launch wizard.
func (h *ChecklistTemplates) List(w http.ResponseWriter, r *http.Request) {
	subject, user, memberRole, ok := h.loadSubject(w, r)
	if !ok {
		return
	}

	forRun := r.URL.Query().Get(queryForRun) == "1"
	if forRun {
		_, orgMember, err := h.Store.OrganizationMemberRole(r.Context(), subject.OrganizationID, user.ID)
		if err != nil {
			slog.Error("org member role", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if !subjects.CanLaunchRun(user, orgMember) {
			http.NotFound(w, r)
			return
		}
	}

	items, err := h.Store.ListChecklistTemplates(r.Context(), subject.ID)
	if err != nil {
		slog.Error("list checklist templates", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	filterQuery := strings.TrimSpace(r.URL.Query().Get("q"))
	if filterQuery != "" {
		items = filterChecklistTemplates(items, filterQuery)
	}

	selectedTemplateID := parseTemplateQuery(r)
	var selectedTemplateName string
	selectedTemplateCompatible := false
	if forRun && selectedTemplateID > 0 {
		items, selectedTemplateName, selectedTemplateCompatible = prioritizeSelectedTemplate(r.Context(), h.Store, items, selectedTemplateID)
	}

	var pd viewtemplates.PageData
	if forRun {
		pd = h.PageData(r, "Lancer")
		pd.Breadcrumbs = viewtemplates.BCRunWizardTemplates(subject.Name, subject.ID)
		pd.ActiveTab = "runs"
	} else {
		pd = h.PageDataTab(r, "Modèles — "+subject.Name, "templates")
		pd.Breadcrumbs = viewtemplates.BCSubjectTemplatesList(subject.Name, subject.ID, pd.Labels.Subject)
	}
	data := viewtemplates.ChecklistTemplatesListData{
		PageData:                   pd,
		Subject:                    subject,
		Templates:                  items,
		MemberRole:                 memberRole,
		CanManage:                  CanManageGlobal(user),
		ForRun:                     forRun,
		SelectedTemplateID:         selectedTemplateID,
		SelectedTemplateName:       selectedTemplateName,
		SelectedTemplateCompatible: selectedTemplateCompatible,
		FilterQuery:                filterQuery,
		HasActiveFilters:           filterQuery != "",
		Message:                    r.URL.Query().Get("msg"),
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
	pd := h.PageDataTab(r, "Nouveau modèle", "templates")
	pd.Breadcrumbs = viewtemplates.BCTemplatesNewWizard()
	data := viewtemplates.ChecklistTemplateFormData{
		PageData:        pd,
		Sections:        sections,
		SectionsEnabled: sectionsEnabled(sections),
		FormAction:      "/modeles",
	}
	applyTemplateFormLimits(&data)

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
	pd.Breadcrumbs = viewtemplates.BCTemplateGlobalShow(template.Name, template.ID)

	orgMember := h.callerOrgMembership(r, user)

	data := viewtemplates.ChecklistTemplateShowData{
		PageData:     pd,
		Template:     template,
		Version:      version,
		Tags:         tags,
		ItemSections: groupTemplateItems(items),
		ItemCount:    len(items),
		CanManage:    CanManageGlobal(user),
		CanLaunch:    subjects.CanLaunchRun(user, orgMember),
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

	pd := h.PageDataTab(r, "Modifier "+template.Name, "templates")
	pd.Breadcrumbs = viewtemplates.BCTemplateGlobalEdit(template.Name, template.ID)
	data := viewtemplates.ChecklistTemplateFormData{
		PageData:        pd,
		Template:        template,
		Version:         version,
		Tags:            store.FormatTagsCSV(tags),
		TagsList:        tags,
		Sections:        sections,
		SectionsEnabled: sectionsEnabled(sections),
		FormAction:      "/modeles/" + strconv.FormatInt(template.ID, 10),
	}
	applyTemplateFormLimits(&data)

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

func (h *ChecklistTemplates) loadSubject(w http.ResponseWriter, r *http.Request) (*store.Subject, *store.User, string, bool) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return nil, nil, "", false
	}

	subjectID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return nil, nil, "", false
	}

	subject, err := h.Store.SubjectByID(r.Context(), subjectID)
	if errors.Is(err, store.ErrSubjectNotFound) {
		http.NotFound(w, r)
		return nil, nil, "", false
	}
	if err != nil {
		slog.Error("load subject", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, "", false
	}

	_, orgMember, err := h.Store.OrganizationMemberRole(r.Context(), subject.OrganizationID, user.ID)
	if err != nil {
		slog.Error("org member role", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, "", false
	}

	if !CanView(user, orgMember) {
		http.NotFound(w, r)
		return nil, nil, "", false
	}

	memberRole := "lead"
	return subject, user, memberRole, true
}

func filterChecklistTemplates(items []store.ChecklistTemplateSummary, query string) []store.ChecklistTemplateSummary {
	terms := strings.Fields(strings.TrimSpace(query))
	if len(terms) == 0 {
		return items
	}
	var filtered []store.ChecklistTemplateSummary
	for _, item := range items {
		if checklistTemplateMatches(item, terms) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func checklistTemplateMatches(item store.ChecklistTemplateSummary, terms []string) bool {
	haystack := strings.ToLower(item.Name)
	for _, tag := range item.Tags {
		haystack += " " + strings.ToLower(tag)
	}
	for _, term := range terms {
		if !strings.Contains(haystack, strings.ToLower(term)) {
			return false
		}
	}
	return true
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
	pd := h.PageDataTab(r, title, "templates")
	if template != nil {
		pd.Breadcrumbs = viewtemplates.BCTemplateGlobalEdit(template.Name, template.ID)
	} else {
		pd.Breadcrumbs = viewtemplates.BCTemplatesNewWizard()
	}
	data := viewtemplates.ChecklistTemplateFormData{
		PageData:        pd,
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
	applyTemplateFormLimits(&data)
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

func (h *ChecklistTemplates) callerOrgMembership(r *http.Request, user *store.User) bool {
	orgMember := false
	if org, ok := middleware.OrganizationFromContext(r.Context()); ok {
		_, orgMember, _ = h.Store.OrganizationMemberRole(r.Context(), org.ID, user.ID)
	}
	return orgMember
}

func parseTemplateQuery(r *http.Request) int64 {
	raw := strings.TrimSpace(r.URL.Query().Get(queryTemplate))
	if raw == "" {
		return 0
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0
	}
	return id
}

func prioritizeSelectedTemplate(
	ctx context.Context,
	st ChecklistTemplateStore,
	items []store.ChecklistTemplateSummary,
	templateID int64,
) ([]store.ChecklistTemplateSummary, string, bool) {
	if templateID <= 0 {
		return items, "", false
	}

	var selected *store.ChecklistTemplateSummary
	rest := make([]store.ChecklistTemplateSummary, 0, len(items))
	for i := range items {
		if items[i].ID == templateID {
			selected = &items[i]
			continue
		}
		rest = append(rest, items[i])
	}
	if selected != nil {
		return append([]store.ChecklistTemplateSummary{*selected}, rest...), selected.Name, true
	}

	template, err := st.ChecklistTemplateByID(ctx, templateID)
	if err != nil || template.ArchivedAt.Valid {
		return items, "", false
	}
	return items, template.Name, false
}
