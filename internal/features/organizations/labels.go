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

// SubjectLabelsShow renders the org admin form to pick subject and run UI label presets.
func (h *Organizations) SubjectLabelsShow(w http.ResponseWriter, r *http.Request) {
	h.renderSubjectLabels(w, r, templates.AdminSubjectLabelsData{
		Message: r.URL.Query().Get("msg"),
	})
}

// SubjectLabelsSave persists the selected subject and run UI label presets for the active org.
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

	subjectLabel := strings.TrimSpace(r.FormValue("ui_subject_label"))
	runLabel := strings.TrimSpace(r.FormValue("ui_run_label"))

	if err := h.Store.UpdateOrganizationUISubjectLabel(r.Context(), org.ID, subjectLabel); err != nil {
		if errors.Is(err, store.ErrInvalidUISubjectLabel) {
			h.renderSubjectLabels(w, r, templates.AdminSubjectLabelsData{
				Error:      "Preset de libellé sujet inconnu.",
				Current:    subjectLabel,
				CurrentRun: runLabel,
			})
			return
		}
		slog.Error("update organization ui subject label", "err", err, "organization_id", org.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := h.Store.UpdateOrganizationUIRunLabel(r.Context(), org.ID, runLabel); err != nil {
		if errors.Is(err, store.ErrInvalidUIRunLabel) {
			h.renderSubjectLabels(w, r, templates.AdminSubjectLabelsData{
				Error:      "Preset de libellé d'instances inconnu.",
				Current:    subjectLabel,
				CurrentRun: runLabel,
			})
			return
		}
		slog.Error("update organization ui run label", "err", err, "organization_id", org.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/settings/labels?msg=Libell%C3%A9s+mis+%C3%A0+jour", http.StatusFound)
}

func (h *Organizations) renderSubjectLabels(w http.ResponseWriter, r *http.Request, data templates.AdminSubjectLabelsData) {
	pd := h.pageData(r)
	pd.ActiveTab = "org"
	pd.AdminSection = "labels"
	data.PageData = templates.ApplyPageMeta(pd, templates.BCAdminSubjectLabels(pd.Labels.Subject))
	data.Presets = templates.SubjectLabelPresets()
	data.RunPresets = templates.RunLabelPresets()

	if org, ok := middleware.OrganizationFromContext(r.Context()); ok {
		refreshed := org
		if got, err := h.Store.OrganizationByID(r.Context(), org.ID); err == nil {
			refreshed = got
			data.Labels = templates.LabelsFromOrganization(got)
		}
		if data.Current == "" {
			data.Current = refreshed.UISubjectLabel
		}
		if data.CurrentRun == "" {
			data.CurrentRun = refreshed.UIRunLabel
		}
	}
	if data.Current == "" {
		data.Current = store.UISubjectLabelSujet
	}
	if data.CurrentRun == "" {
		data.CurrentRun = store.UIRunLabelRevues
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "admin_subject_labels", data); err != nil {
		slog.Error("render admin subject labels", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
