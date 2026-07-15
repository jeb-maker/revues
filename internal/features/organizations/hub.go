package organizations

import (
	"log/slog"
	"net/http"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/web/middleware"
	"github.com/jeb-maker/revues/internal/web/templates"
)

// AdminHub renders the minimal organisation admin landing page.
func (h *Organizations) AdminHub(w http.ResponseWriter, r *http.Request) {
	data := templates.AdminOrgHubData{
		PageData: templates.ApplyPageMeta(h.pageData(r), templates.BCAdminOrgHub()),
	}
	data.ActiveTab = "org"

	if user, ok := middleware.UserFromContext(r.Context()); ok {
		data.ShowIntegrations = auth.HasMinRole(user.Role, auth.RoleAdmin)
	}
	if org, ok := middleware.OrganizationFromContext(r.Context()); ok {
		data.OrganizationName = org.Name
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "admin_org_hub", data); err != nil {
		slog.Error("render admin org hub", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
