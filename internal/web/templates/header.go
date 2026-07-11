package templates

import (
	"net/http"

	"github.com/jeb-maker/revues/internal/web/middleware"
)

// ApplyHeaderFromContext copies organization switcher data into page view data.
func ApplyHeaderFromContext(r *http.Request, data *PageData) {
	hd, ok := middleware.HeaderDataFromContext(r.Context())
	if !ok {
		return
	}
	data.ActiveOrganization = hd.ActiveOrg
	data.UserOrganizations = hd.UserOrganizations
	data.PendingInvitations = hd.PendingInvitations
}
