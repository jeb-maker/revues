package templates

import (
	"net/http"

	"github.com/jeb-maker/revues/internal/web/middleware"
)

// ApplyHeaderFromContext copies organization switcher data into page view data.
func ApplyHeaderFromContext(r *http.Request, data *PageData) {
	hd, ok := middleware.HeaderDataFromContext(r.Context())
	if ok {
		data.ActiveOrganization = hd.ActiveOrg
		data.UserOrganizations = hd.UserOrganizations
		data.PendingInvitations = hd.PendingInvitations
		data.CanManageOrgUsers = hd.CanManageOrgUsers
		data.ShowOrganisationNav = hd.ShowOrganisationNav
		if hd.ActiveOrg != nil {
			data.Labels = LabelsFromOrganization(hd.ActiveOrg)
		}
	}
	EnsureLabels(data)
}
