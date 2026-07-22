package templates

import (
	"net/http"

	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/jeb-maker/revues/internal/web/middleware"
)

// ApplyHeaderFromContext copies organization switcher data into page view data.
func ApplyHeaderFromContext(r *http.Request, data *PageData) {
	data.RequestID = chimw.GetReqID(r.Context())
	hd, ok := middleware.HeaderDataFromContext(r.Context())
	if ok {
		data.ActiveOrganization = hd.ActiveOrg
		data.UserOrganizations = hd.UserOrganizations
		data.PendingInvitations = hd.PendingInvitations
		data.CanManageOrgUsers = hd.CanManageOrgUsers
		data.ShowOrganisationNav = hd.ShowOrganisationNav
		data.SimpleUI = hd.SimpleUI
		data.SimpleSubjectID = hd.SimpleSubjectID
		data.ShowAssign = hd.ShowAssign
		data.ShowMyTasks = hd.ShowMyTasks
		data.ShowSubjectColumn = hd.ShowSubjectColumn
		data.ShowCollab = hd.ShowCollab
		data.DevAuth = hd.DevAuth
		data.DevAuthUsers = hd.DevAuthUsers
		if hd.ActiveOrg != nil {
			data.Labels = LabelsFromOrganization(hd.ActiveOrg)
		}
	}
	EnsureLabels(data)
}
