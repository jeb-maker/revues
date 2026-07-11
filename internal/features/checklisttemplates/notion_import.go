package checklisttemplates

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jeb-maker/revues/internal/crypto"
	"github.com/jeb-maker/revues/internal/integrations/notion"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/middleware"
	viewtemplates "github.com/jeb-maker/revues/internal/web/templates"
)

const (
	notionImportStepSource  = "source"
	notionImportStepMapping = "mapping"
	notionImportStepPreview = "preview"
)

func (h *ChecklistTemplates) NotionImportForm(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if !CanManageGlobal(user) {
		http.NotFound(w, r)
		return
	}
	data := h.notionImportBaseData(r)
	data.Step = notionImportStepSource
	if cfg, configured, err := h.notionConfigured(r); err != nil {
		data.Error = "Impossible de charger la configuration Notion."
	} else {
		data.NotionConfigured = configured && cfg.Configured()
		if data.NotionConfigured && cfg.DefaultDatabaseID != "" {
			data.DatabaseRef = cfg.DefaultDatabaseID
		}
	}
	h.renderNotionImport(w, data)
}

func (h *ChecklistTemplates) NotionImport(w http.ResponseWriter, r *http.Request) {
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
	data := h.notionImportBaseData(r)
	data.DatabaseRef = strings.TrimSpace(r.FormValue("database_ref"))
	data.DatabaseID = strings.TrimSpace(r.FormValue("database_id"))
	data.TemplateName = strings.TrimSpace(r.FormValue("template_name"))
	data.Tags = strings.TrimSpace(r.FormValue("tags"))
	data.Mapping = notion.ColumnMapping{
		Label: strings.TrimSpace(r.FormValue("map_label")), Section: strings.TrimSpace(r.FormValue("map_section")),
		HelpText: strings.TrimSpace(r.FormValue("map_help")), Required: strings.TrimSpace(r.FormValue("map_required")),
	}
	cfg, configured, err := h.notionConfigured(r)
	if err != nil {
		data.Error = "Impossible de charger la configuration Notion."
		h.renderNotionImport(w, data)
		return
	}
	if !configured || !cfg.Configured() {
		data.Error = "Notion n'est pas configuré. Demandez à un administrateur de configurer l'intégration."
		data.Step = notionImportStepSource
		h.renderNotionImport(w, data)
		return
	}
	switch strings.TrimSpace(r.FormValue("action")) {
	case "fetch":
		h.notionImportFetch(w, r, cfg, data)
	case "preview":
		h.notionImportPreview(w, r, cfg, data)
	case "import":
		h.notionImportCreate(w, r, user, cfg, data)
	default:
		data.Error = "Action invalide."
		data.Step = notionImportStepSource
		h.renderNotionImport(w, data)
	}
}

func (h *ChecklistTemplates) notionImportFetch(w http.ResponseWriter, r *http.Request, cfg notion.Config, data viewtemplates.ChecklistTemplateNotionImportData) {
	ref := data.DatabaseRef
	if ref == "" {
		ref = data.DatabaseID
	}
	dbID, err := notion.ParseDatabaseRef(ref)
	if err != nil {
		data.Error, data.Step = err.Error(), notionImportStepSource
		h.renderNotionImport(w, data)
		return
	}
	db, err := h.notionClient().GetDatabase(r.Context(), cfg, dbID)
	if err != nil {
		if errors.Is(err, notion.ErrDatabaseNotFound) {
			data.Error = "Base Notion introuvable. Vérifiez l'URL ou l'identifiant."
		} else {
			data.Error = "Impossible de lire la base Notion. Vérifiez le jeton et les droits d'accès."
			slog.Error("notion get database", "err", err)
		}
		data.Step = notionImportStepSource
		h.renderNotionImport(w, data)
		return
	}
	data.DatabaseID, data.DatabaseTitle = db.ID, db.Title
	data.Properties = notionPropertiesToOptions(db.Properties)
	if data.TemplateName == "" {
		data.TemplateName = db.Title
	}
	if data.Mapping.Label == "" {
		data.Mapping = notion.DefaultMapping(db)
	}
	data.Step = notionImportStepMapping
	h.renderNotionImport(w, data)
}

func (h *ChecklistTemplates) notionImportPreview(w http.ResponseWriter, r *http.Request, cfg notion.Config, data viewtemplates.ChecklistTemplateNotionImportData) {
	preview, db, err := h.loadNotionPreview(r, cfg, data)
	if err != nil {
		data.Error, data.Step = err.Error(), notionImportStepMapping
		if db.ID != "" {
			data.DatabaseTitle, data.Properties, data.DatabaseID = db.Title, notionPropertiesToOptions(db.Properties), db.ID
		}
		h.renderNotionImport(w, data)
		return
	}
	data.Step, data.TemplateName = notionImportStepPreview, preview.TemplateName
	data.PreviewItems, data.PreviewCount = templateItemsToRows(preview.Items), len(preview.Items)
	h.renderNotionImport(w, data)
}

func (h *ChecklistTemplates) notionImportCreate(w http.ResponseWriter, r *http.Request, user *store.User, cfg notion.Config, data viewtemplates.ChecklistTemplateNotionImportData) {
	preview, db, err := h.loadNotionPreview(r, cfg, data)
	if err != nil {
		data.Error, data.Step = err.Error(), notionImportStepMapping
		if db.ID != "" {
			data.DatabaseTitle, data.Properties, data.DatabaseID = db.Title, notionPropertiesToOptions(db.Properties), db.ID
		}
		h.renderNotionImport(w, data)
		return
	}
	tags := store.ParseTagsCSV(data.Tags)
	template, _, err := h.Store.CreateChecklistTemplate(r.Context(), preview.TemplateName, user.ID, tags, preview.Items)
	if err != nil {
		slog.Error("create checklist template from notion", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, templateShowURL(template.ID)+"?msg=Mod%C3%A8le+import%C3%A9+depuis+Notion", http.StatusSeeOther)
}

func (h *ChecklistTemplates) loadNotionPreview(r *http.Request, cfg notion.Config, data viewtemplates.ChecklistTemplateNotionImportData) (notion.ImportPreview, notion.DatabaseInfo, error) {
	dbID := data.DatabaseID
	if dbID == "" {
		var err error
		dbID, err = notion.ParseDatabaseRef(data.DatabaseRef)
		if err != nil {
			return notion.ImportPreview{}, notion.DatabaseInfo{}, err
		}
	}
	db, err := h.notionClient().GetDatabase(r.Context(), cfg, dbID)
	if err != nil {
		return notion.ImportPreview{}, notion.DatabaseInfo{}, errors.New("impossible de relire la base Notion")
	}
	pages, err := h.notionClient().QueryDatabase(r.Context(), cfg, db.ID)
	if err != nil {
		slog.Error("notion query database", "err", err)
		return notion.ImportPreview{}, db, errors.New("impossible de lire les lignes Notion")
	}
	preview, err := notion.BuildImportPreview(db, pages, data.Mapping, data.TemplateName)
	if err != nil {
		return notion.ImportPreview{}, db, err
	}
	return preview, db, nil
}

func (h *ChecklistTemplates) notionImportBaseData(r *http.Request) viewtemplates.ChecklistTemplateNotionImportData {
	pd := h.PageDataTab(r, "Importer depuis Notion", "templates")
	pd.Breadcrumbs = viewtemplates.BCTemplateNotionImportGlobal()
	return viewtemplates.ChecklistTemplateNotionImportData{
		PageData:   pd,
		CanManage:  true,
		FormAction: "/modeles/notion-import",
	}
}

func (h *ChecklistTemplates) renderNotionImport(w http.ResponseWriter, data viewtemplates.ChecklistTemplateNotionImportData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if data.Error != "" && data.Step == "" {
		data.Step = notionImportStepSource
	}
	if err := h.Templates.ExecuteTemplate(w, "checklist_template_notion_import", data); err != nil {
		slog.Error("render notion import", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *ChecklistTemplates) notionConfigured(r *http.Request) (notion.Config, bool, error) {
	if len(h.EncryptionKey) != crypto.KeySize {
		return notion.Config{}, false, nil
	}
	st, ok := h.Store.(*store.Store)
	if !ok {
		return notion.Config{}, false, nil
	}
	return (&notion.Service{Store: st, EncryptionKey: h.EncryptionKey}).Load(r.Context())
}

func (h *ChecklistTemplates) notionClient() *notion.Client {
	if h.NotionClient != nil {
		return h.NotionClient
	}
	return &notion.Client{}
}

func templateItemsToRows(items []store.TemplateItemInput) []viewtemplates.TemplateEditorRow {
	rows := make([]viewtemplates.TemplateEditorRow, len(items))
	for i, item := range items {
		rows[i] = viewtemplates.TemplateEditorRow{Label: item.Label, HelpText: item.HelpText, Required: item.Required}
	}
	return rows
}

func notionPropertiesToOptions(props []notion.PropertyInfo) []viewtemplates.NotionPropertyOption {
	out := make([]viewtemplates.NotionPropertyOption, len(props))
	for i, prop := range props {
		out[i] = viewtemplates.NotionPropertyOption{Name: prop.Name, Type: prop.Type}
	}
	return out
}
