package organizations

import (
	"log/slog"
	"net/http"

	"github.com/jeb-maker/revues/internal/web/middleware"
	"github.com/jeb-maker/revues/internal/web/templates"
)

// AdminHub renders the organisation admin landing page.
func (h *Organizations) AdminHub(w http.ResponseWriter, r *http.Request) {
	pd := h.pageData(r)
	pd.ActiveTab = "org"
	pd.AdminSection = "hub"
	data := templates.AdminOrgHubData{
		PageData: templates.ApplyPageMeta(pd, templates.BCAdminOrgHub()),
	}

	if org, ok := middleware.OrganizationFromContext(r.Context()); ok {
		data.OrganizationName = org.Name
		data.OrganizationSlug = org.Slug
		data.CreatedAt = org.CreatedAt
		if count, err := h.Store.CountOrganizationMembers(r.Context(), org.ID); err != nil {
			slog.Error("count organization members", "err", err)
		} else {
			data.MemberCount = count
		}
	}
	if count, err := h.Store.CountOrganizationSubjects(r.Context()); err != nil {
		slog.Error("count organization subjects", "err", err)
	} else {
		data.SubjectCount = count
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "admin_org_hub", data); err != nil {
		slog.Error("render admin org hub", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
