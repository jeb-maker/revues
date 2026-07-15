package organizations

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/middleware"
	"github.com/jeb-maker/revues/internal/web/templates"
)

// SubjectLabelsShow renders the org admin form to pick the subject UI label preset.
func (h *Organizations) SubjectLabelsShow(w http.ResponseWriter, r *http.Request) {
	h.renderSubjectLabels(w, r, templates.AdminSubjectLabelsData{
		Message: r.URL.Query().Get("msg"),
	})
}

// SubjectLabelsSave persists the selected subject UI label preset for the active org.
func (h *Organizations) SubjectLabelsSave(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	org, ok := middleware.OrganizationFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/org/select", http.StatusFound)
		return
	}

	label := strings.TrimSpace(r.FormValue("ui_subject_label"))
	if err := h.Store.UpdateOrganizationUISubjectLabel(r.Context(), org.ID, label); err != nil {
		if errors.Is(err, store.ErrInvalidUISubjectLabel) {
			h.renderSubjectLabels(w, r, templates.AdminSubjectLabelsData{
				Error:   "Preset de libellé inconnu.",
				Current: label,
			})
			return
		}
		slog.Error("update organization ui subject label", "err", err, "organization_id", org.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/settings/labels?msg=Libell%C3%A9+mis+%C3%A0+jour", http.StatusFound)
}

func (h *Organizations) renderSubjectLabels(w http.ResponseWriter, r *http.Request, data templates.AdminSubjectLabelsData) {
	pd := h.pageData(r)
	pd.ActiveTab = "org"
	pd.AdminSection = "labels"
	data.PageData = templates.ApplyPageMeta(pd, templates.BCAdminSubjectLabels(pd.Labels.Subject))
	data.Presets = templates.SubjectLabelPresets()

	if data.Current == "" {
		if org, ok := middleware.OrganizationFromContext(r.Context()); ok {
			data.Current = org.UISubjectLabel
			if refreshed, err := h.Store.OrganizationByID(r.Context(), org.ID); err == nil {
				data.Current = refreshed.UISubjectLabel
				data.Labels = templates.LabelsFromOrganization(refreshed)
			}
		}
	}
	if data.Current == "" {
		data.Current = store.UISubjectLabelSujet
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "admin_subject_labels", data); err != nil {
		slog.Error("render admin subject labels", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
