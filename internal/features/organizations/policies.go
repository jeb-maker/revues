package organizations

import (
	"log/slog"
	"net/http"

	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/middleware"
	"github.com/jeb-maker/revues/internal/web/templates"
)

// LeadPoliciesShow renders the org admin form for lead-delegation policies.
func (h *Organizations) LeadPoliciesShow(w http.ResponseWriter, r *http.Request) {
	h.renderLeadPolicies(w, r, templates.AdminLeadPoliciesData{
		Message: r.URL.Query().Get("msg"),
	})
}

// LeadPoliciesSave persists lead-delegation policies for the active org.
func (h *Organizations) LeadPoliciesSave(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	org, ok := middleware.OrganizationFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/org/select", http.StatusFound)
		return
	}

	policies := store.OrgLeadPolicies{
		LeadsMayAssignTeams:     formCheckbox(r, "leads_may_assign_teams"),
		LeadsMayInviteMembers:   formCheckbox(r, "leads_may_invite_members"),
		LeadsMayInviteExternals: formCheckbox(r, "leads_may_invite_externals"),
	}
	if err := h.Store.UpdateOrganizationLeadPolicies(r.Context(), org.ID, policies); err != nil {
		slog.Error("update organization lead policies", "err", err, "organization_id", org.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/settings/policies?msg=Politiques+mises+%C3%A0+jour", http.StatusFound)
}

func formCheckbox(r *http.Request, name string) bool {
	v := r.FormValue(name)
	return v == "on" || v == "1" || v == "true"
}

func (h *Organizations) renderLeadPolicies(w http.ResponseWriter, r *http.Request, data templates.AdminLeadPoliciesData) {
	pd := h.pageData(r)
	pd.ActiveTab = "org"
	pd.AdminSection = "policies"
	data.PageData = templates.ApplyPageMeta(pd, templates.BCAdminLeadPolicies())

	if org, ok := middleware.OrganizationFromContext(r.Context()); ok {
		data.Policies = org.LeadPolicies()
		if refreshed, err := h.Store.OrganizationByID(r.Context(), org.ID); err == nil {
			data.Policies = refreshed.LeadPolicies()
		}
	} else {
		data.Policies = store.DefaultOrgLeadPolicies()
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "admin_lead_policies", data); err != nil {
		slog.Error("render admin lead policies", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
