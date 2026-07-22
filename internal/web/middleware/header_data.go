package middleware

import (
	"context"
	"net/http"

	"github.com/jeb-maker/revues/internal/store"
)

const headerDataContextKey contextKey = 3

// HeaderData holds organization switcher and pending invitation view data.
type HeaderData struct {
	ActiveOrg           *store.Organization
	UserOrganizations   []store.OrganizationMembership
	PendingInvitations  []store.OrganizationInvitation
	CanManageOrgUsers   bool
	ShowOrganisationNav bool
	SimpleUI            bool
	SimpleSubjectID     int64
	ShowAssign          bool
	ShowMyTasks         bool
	ShowSubjectColumn   bool
	ShowCollab          bool
	DevAuth             bool
	DevAuthUsers        []store.User
}

// LoadHeaderData preloads organization switcher data for authenticated requests.
func LoadHeaderData(st *store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			var hd HeaderData
			if org, ok := OrganizationFromContext(r.Context()); ok {
				hd.ActiveOrg = org
			}

			orgs, err := st.ListUserOrganizations(r.Context(), user.ID)
			if err == nil {
				hd.UserOrganizations = orgs
			}

			if user.Email != "" {
				invites, err := st.ListPendingInvitationsByEmail(r.Context(), user.Email)
				if err == nil {
					hd.PendingInvitations = invites
				}
			}

			hd.CanManageOrgUsers = CanManageOrgUsers(r.Context(), st, user)
			hd.ShowOrganisationNav = showOrganisationNav(r.Context(), st, user, hd)
			caps := resolveUICaps(r.Context(), st, user, hd)
			hd.SimpleUI = caps.SimpleUI
			hd.SimpleSubjectID = caps.SimpleSubjectID
			hd.ShowAssign = caps.ShowAssign
			hd.ShowMyTasks = caps.ShowMyTasks
			hd.ShowSubjectColumn = caps.ShowSubjectColumn
			hd.ShowCollab = caps.ShowCollab

			if DevAuthUIActive(r.Context()) {
				hd.DevAuth = true
				if users, listErr := st.ListUsers(r.Context()); listErr == nil {
					hd.DevAuthUsers = users
				}
			}

			ctx := context.WithValue(r.Context(), headerDataContextKey, hd)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// HeaderDataFromContext returns preloaded header view data, if any.
func HeaderDataFromContext(ctx context.Context) (HeaderData, bool) {
	hd, ok := ctx.Value(headerDataContextKey).(HeaderData)
	return hd, ok
}
