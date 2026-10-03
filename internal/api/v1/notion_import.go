package apiv1

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jeb-maker/revues/internal/features/checklisttemplates"
	"github.com/jeb-maker/revues/internal/integrations/notion"
	"github.com/jeb-maker/revues/internal/store"
)

// PostTemplatesNotionImport serves POST /api/v1/templates/notion-import.
func (s *Server) PostTemplatesNotionImport(w http.ResponseWriter, r *http.Request) {
	user, ok := requireAPIUser(w, r)
	if !ok {
		return
	}
	if !requireOrg(w, r) {
		return
	}
	if !checklisttemplates.CanManageGlobal(user) {
		writeAPIError(w, http.StatusForbidden, "forbidden", "Droits éditeur requis pour importer un modèle.")
		return
	}
	if s.Notion == nil || s.Templates == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	var req NotionImportRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}

	cfg, configured, err := s.Notion.Load(r.Context())
	if err != nil {
		slog.Error("load notion for import", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if !configured || !cfg.Configured() {
		writeAPIError(w, http.StatusBadRequest, "validation_failed",
			"Notion n'est pas configuré. Demandez à un administrateur de configurer l'intégration.")
		return
	}

	mapping := columnMappingFromRequest(req.Mapping)
	templateName := ""
	if req.TemplateName != nil {
		templateName = strings.TrimSpace(*req.TemplateName)
	}
	databaseRef := ""
	if req.DatabaseRef != nil {
		databaseRef = strings.TrimSpace(*req.DatabaseRef)
	}
	databaseID := ""
	if req.DatabaseId != nil {
		databaseID = strings.TrimSpace(*req.DatabaseId)
	}

	switch req.Action {
	case NotionImportRequestActionFetch:
		s.notionImportFetch(w, r, cfg, databaseRef, databaseID, templateName, mapping)
	case NotionImportRequestActionPreview:
		s.notionImportPreview(w, r, cfg, databaseRef, databaseID, templateName, mapping)
	case NotionImportRequestActionImport:
		s.notionImportCreate(w, r, user, cfg, databaseRef, databaseID, templateName, mapping, req.Domains)
	default:
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Action invalide.")
	}
}

func (s *Server) notionImportFetch(
	w http.ResponseWriter,
	r *http.Request,
	cfg notion.Config,
	databaseRef, databaseID, templateName string,
	mapping notion.ColumnMapping,
) {
	ref := databaseRef
	if ref == "" {
		ref = databaseID
	}
	dbID, err := notion.ParseDatabaseRef(ref)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", notion.UserMessage(err))
		return
	}
	db, err := s.notionClient().GetDatabase(r.Context(), cfg, dbID)
	if err != nil {
		if errors.Is(err, notion.ErrDatabaseNotFound) {
			writeAPIError(w, http.StatusBadRequest, "validation_failed",
				"Base Notion introuvable. Vérifiez l'URL ou l'identifiant.")
			return
		}
		slog.Error("notion get database", "err", err)
		writeAPIError(w, http.StatusBadRequest, "validation_failed", notion.UserMessage(err))
		return
	}
	if mapping.Label == "" {
		mapping = notion.DefaultMapping(db)
	}
	name := templateName
	if name == "" {
		name = db.Title
	}
	writeJSON(w, http.StatusOK, notionImportMappingResponse(db, mapping, name))
}

func (s *Server) notionImportPreview(
	w http.ResponseWriter,
	r *http.Request,
	cfg notion.Config,
	databaseRef, databaseID, templateName string,
	mapping notion.ColumnMapping,
) {
	preview, db, err := s.loadNotionPreview(r, cfg, databaseRef, databaseID, templateName, mapping)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", notion.UserMessage(err))
		return
	}
	writeJSON(w, http.StatusOK, notionImportPreviewResponse(db, mapping, preview))
}

func (s *Server) notionImportCreate(
	w http.ResponseWriter,
	r *http.Request,
	user *store.User,
	cfg notion.Config,
	databaseRef, databaseID, templateName string,
	mapping notion.ColumnMapping,
	domains *[]string,
) {
	preview, db, err := s.loadNotionPreview(r, cfg, databaseRef, databaseID, templateName, mapping)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", notion.UserMessage(err))
		return
	}
	domainList := []string{}
	if domains != nil {
		domainList = *domains
	}
	detail, err := s.Templates.Create(r.Context(), user, checklisttemplates.CreateInput{
		Name:    preview.TemplateName,
		Domains: domainList,
		Items:   preview.Items,
	})
	if err != nil {
		writeTemplateErr(w, err)
		return
	}
	mapped := mapTemplateDetail(detail, true)
	count := len(preview.Items)
	resp := NotionImportResponse{
		Step:         NotionImportResponseStepDone,
		DatabaseId:   ptrString(db.ID),
		TemplateName: ptrString(preview.TemplateName),
		PreviewCount: &count,
		Template:     &mapped,
	}
	if db.Title != "" {
		resp.DatabaseTitle = &db.Title
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (s *Server) loadNotionPreview(
	r *http.Request,
	cfg notion.Config,
	databaseRef, databaseID, templateName string,
	mapping notion.ColumnMapping,
) (notion.ImportPreview, notion.DatabaseInfo, error) {
	dbID := databaseID
	if dbID == "" {
		var err error
		ref := databaseRef
		if ref == "" {
			return notion.ImportPreview{}, notion.DatabaseInfo{}, errors.New("identifiant base Notion requis")
		}
		dbID, err = notion.ParseDatabaseRef(ref)
		if err != nil {
			return notion.ImportPreview{}, notion.DatabaseInfo{}, err
		}
	}
	db, err := s.notionClient().GetDatabase(r.Context(), cfg, dbID)
	if err != nil {
		if errors.Is(err, notion.ErrDatabaseNotFound) {
			return notion.ImportPreview{}, notion.DatabaseInfo{}, errors.New("base Notion introuvable")
		}
		slog.Error("notion get database for preview", "err", err)
		return notion.ImportPreview{}, notion.DatabaseInfo{}, errors.New(notion.UserMessage(err))
	}
	pages, err := s.notionClient().QueryDatabase(r.Context(), cfg, db.ID)
	if err != nil {
		slog.Error("notion query database", "err", err)
		return notion.ImportPreview{}, db, errors.New(notion.UserMessage(err))
	}
	preview, err := notion.BuildImportPreview(db, pages, mapping, templateName)
	if err != nil {
		return notion.ImportPreview{}, db, err
	}
	return preview, db, nil
}

func columnMappingFromRequest(m *NotionColumnMapping) notion.ColumnMapping {
	if m == nil {
		return notion.ColumnMapping{}
	}
	out := notion.ColumnMapping{}
	if m.Label != nil {
		out.Label = strings.TrimSpace(*m.Label)
	}
	if m.Section != nil {
		out.Section = strings.TrimSpace(*m.Section)
	}
	if m.HelpText != nil {
		out.HelpText = strings.TrimSpace(*m.HelpText)
	}
	if m.Required != nil {
		out.Required = strings.TrimSpace(*m.Required)
	}
	return out
}

func notionImportMappingResponse(db notion.DatabaseInfo, mapping notion.ColumnMapping, templateName string) NotionImportResponse {
	props := make([]NotionPropertyOption, 0, len(db.Properties))
	for _, p := range db.Properties {
		props = append(props, NotionPropertyOption{Name: p.Name, Type: p.Type})
	}
	mapped := mapColumnMapping(mapping)
	resp := NotionImportResponse{
		Step:         NotionImportResponseStepMapping,
		DatabaseId:   ptrString(db.ID),
		Properties:   &props,
		Mapping:      &mapped,
		TemplateName: ptrString(templateName),
	}
	if db.Title != "" {
		resp.DatabaseTitle = &db.Title
	}
	return resp
}

func notionImportPreviewResponse(db notion.DatabaseInfo, mapping notion.ColumnMapping, preview notion.ImportPreview) NotionImportResponse {
	items := make([]NotionImportPreviewItem, 0, len(preview.Items))
	for _, item := range preview.Items {
		items = append(items, NotionImportPreviewItem{
			Label:    item.Label,
			Section:  item.Section,
			HelpText: item.HelpText,
			Required: item.Required,
		})
	}
	count := len(items)
	mapped := mapColumnMapping(mapping)
	resp := NotionImportResponse{
		Step:         NotionImportResponseStepPreview,
		DatabaseId:   ptrString(db.ID),
		Mapping:      &mapped,
		TemplateName: ptrString(preview.TemplateName),
		PreviewItems: &items,
		PreviewCount: &count,
	}
	if db.Title != "" {
		resp.DatabaseTitle = &db.Title
	}
	return resp
}

func mapColumnMapping(m notion.ColumnMapping) NotionColumnMapping {
	out := NotionColumnMapping{}
	if m.Label != "" {
		out.Label = &m.Label
	}
	if m.Section != "" {
		out.Section = &m.Section
	}
	if m.HelpText != "" {
		out.HelpText = &m.HelpText
	}
	if m.Required != "" {
		out.Required = &m.Required
	}
	return out
}
