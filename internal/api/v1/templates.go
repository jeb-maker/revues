package apiv1

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/jeb-maker/revues/internal/features/checklisttemplates"
	"github.com/jeb-maker/revues/internal/store"
	appmiddleware "github.com/jeb-maker/revues/internal/web/middleware"
)

// ListTemplates serves GET /api/v1/templates.
func (s *Server) ListTemplates(w http.ResponseWriter, r *http.Request, params ListTemplatesParams) {
	user, ok := requireAPIUser(w, r)
	if !ok {
		return
	}
	q := ""
	if params.Q != nil {
		q = *params.Q
	}
	rows, err := s.Templates.ListIndex(r.Context(), user, q)
	if err != nil {
		slog.Error("list templates", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	out := make([]TemplateSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapTemplateSummary(row))
	}
	writeJSON(w, http.StatusOK, TemplateListResponse{Templates: out})
}

// CreateTemplate serves POST /api/v1/templates.
func (s *Server) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	user, ok := requireAPIUser(w, r)
	if !ok {
		return
	}
	var req TemplateWriteRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}
	detail, err := s.Templates.Create(r.Context(), user, writeRequestToInput(req))
	if err != nil {
		writeTemplateErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, mapTemplateDetail(detail, checklisttemplates.CanManageGlobal(user)))
}

// GetTemplate serves GET /api/v1/templates/{templateId}.
//
//nolint:staticcheck // OpenAPI operation param name (templateId)
func (s *Server) GetTemplate(w http.ResponseWriter, r *http.Request, templateId TemplateId) {
	user, ok := requireAPIUser(w, r)
	if !ok {
		return
	}
	detail, err := s.Templates.GetLatest(r.Context(), templateId)
	if err != nil {
		writeTemplateErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapTemplateDetail(detail, checklisttemplates.CanManageGlobal(user)))
}

// SaveTemplate serves PUT /api/v1/templates/{templateId}.
//
//nolint:staticcheck // OpenAPI operation param name (templateId)
func (s *Server) SaveTemplate(w http.ResponseWriter, r *http.Request, templateId TemplateId) {
	user, ok := requireAPIUser(w, r)
	if !ok {
		return
	}
	var req TemplateWriteRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}
	detail, err := s.Templates.Save(r.Context(), user, templateId, checklisttemplates.SaveInput{
		Name:    req.Name,
		Domains: domainsOrEmpty(req.Domains),
		Items:   mapItemInputs(req.Items),
	})
	if err != nil {
		writeTemplateErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapTemplateDetail(detail, true))
}

// ArchiveTemplate serves DELETE /api/v1/templates/{templateId}.
//
//nolint:staticcheck // OpenAPI operation param name (templateId)
func (s *Server) ArchiveTemplate(w http.ResponseWriter, r *http.Request, templateId TemplateId) {
	user, ok := requireAPIUser(w, r)
	if !ok {
		return
	}
	if err := s.Templates.Archive(r.Context(), user, templateId); err != nil {
		writeTemplateErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListTemplateVersions serves GET /api/v1/templates/{templateId}/versions.
//
//nolint:staticcheck // OpenAPI operation param name (templateId)
func (s *Server) ListTemplateVersions(w http.ResponseWriter, r *http.Request, templateId TemplateId) {
	if _, ok := requireAPIUser(w, r); !ok {
		return
	}
	versions, err := s.Templates.ListVersions(r.Context(), templateId)
	if err != nil {
		writeTemplateErr(w, err)
		return
	}
	out := make([]TemplateVersionSummary, 0, len(versions))
	for _, v := range versions {
		out = append(out, mapVersionSummary(v))
	}
	writeJSON(w, http.StatusOK, TemplateVersionListResponse{Versions: out})
}

// CreateTemplateVersion serves POST /api/v1/templates/{templateId}/versions.
//
//nolint:staticcheck // OpenAPI operation param name (templateId)
func (s *Server) CreateTemplateVersion(w http.ResponseWriter, r *http.Request, templateId TemplateId) {
	user, ok := requireAPIUser(w, r)
	if !ok {
		return
	}
	var req TemplateVersionCreateRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}
	detail, err := s.Templates.CreateVersion(r.Context(), user, templateId, mapItemInputs(req.Items))
	if err != nil {
		writeTemplateErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, mapTemplateDetail(detail, true))
}

// GetTemplateVersion serves GET /api/v1/templates/{templateId}/versions/{version}.
//
//nolint:staticcheck // OpenAPI operation param name (templateId)
func (s *Server) GetTemplateVersion(w http.ResponseWriter, r *http.Request, templateId TemplateId, version TemplateVersionNumber) {
	user, ok := requireAPIUser(w, r)
	if !ok {
		return
	}
	detail, err := s.Templates.GetVersion(r.Context(), templateId, version)
	if err != nil {
		writeTemplateErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapTemplateDetail(detail, checklisttemplates.CanManageGlobal(user)))
}

func requireAPIUser(w http.ResponseWriter, r *http.Request) (*store.User, bool) {
	user, ok := appmiddleware.UserFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthenticated", "Authentification requise.")
		return nil, false
	}
	return user, true
}

func writeTemplateErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, checklisttemplates.ErrNotFound):
		writeAPIError(w, http.StatusNotFound, "not_found", "Modèle introuvable.")
	case errors.Is(err, checklisttemplates.ErrForbidden):
		writeAPIError(w, http.StatusNotFound, "not_found", "Modèle introuvable.")
	case errors.Is(err, checklisttemplates.ErrImmutable):
		writeAPIError(w, http.StatusConflict, "version_immutable",
			"Les versions publiées sont immuables. Publiez une nouvelle version.")
	case errors.Is(err, checklisttemplates.ErrValidation):
		writeAPIError(w, http.StatusBadRequest, "validation_failed", validationMessage(err))
	case errors.Is(err, store.ErrOrganizationRequired):
		writeAPIError(w, http.StatusBadRequest, "org_required",
			"Organisation active requise. Sélectionnez une organisation.")
	default:
		slog.Error("templates api", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
	}
}

func writeRequestToInput(req TemplateWriteRequest) checklisttemplates.CreateInput {
	return checklisttemplates.CreateInput{
		Name:    req.Name,
		Domains: domainsOrEmpty(req.Domains),
		Items:   mapItemInputs(req.Items),
	}
}

func domainsOrEmpty(d *[]string) []string {
	if d == nil {
		return nil
	}
	return *d
}

func mapItemInputs(items []TemplateItemInput) []store.TemplateItemInput {
	out := make([]store.TemplateItemInput, 0, len(items))
	for _, it := range items {
		section := ""
		if it.Section != nil {
			section = *it.Section
		}
		help := ""
		if it.HelpText != nil {
			help = *it.HelpText
		}
		required := false
		if it.Required != nil {
			required = *it.Required
		}
		out = append(out, store.TemplateItemInput{
			Section:  section,
			Label:    it.Label,
			HelpText: help,
			Required: required,
		})
	}
	return out
}

func mapTemplateSummary(row store.TemplateIndexRow) TemplateSummary {
	domains := row.Tags
	if domains == nil {
		domains = []string{}
	}
	return TemplateSummary{
		Id:            row.ID,
		Name:          row.Name,
		LatestVersion: row.LatestVersion,
		ItemCount:     row.ItemCount,
		Domains:       domains,
		CreatedAt:     row.CreatedAt,
	}
}

func mapVersionSummary(v store.TemplateVersion) TemplateVersionSummary {
	sum := TemplateVersionSummary{
		Id:          v.ID,
		Version:     v.Version,
		PublishedAt: v.PublishedAt,
	}
	if v.CreatedBy.Valid {
		id := v.CreatedBy.Int64
		sum.CreatedBy = &id
	}
	return sum
}

func mapTemplateDetail(d *checklisttemplates.Detail, canManage bool) TemplateDetail {
	domains := d.Domains
	if domains == nil {
		domains = []string{}
	}
	items := make([]TemplateItem, 0, len(d.Items))
	for _, it := range d.Items {
		items = append(items, TemplateItem{
			Id:       it.ID,
			Position: it.Position,
			Section:  it.Section,
			Label:    it.Label,
			HelpText: it.HelpText,
			Required: it.Required,
		})
	}
	return TemplateDetail{
		Id:        d.Template.ID,
		Name:      d.Template.Name,
		Domains:   domains,
		Version:   mapVersionSummary(*d.Version),
		Items:     items,
		CreatedAt: d.Template.CreatedAt,
		CanManage: &canManage,
	}
}
