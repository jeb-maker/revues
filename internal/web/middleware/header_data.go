package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

const headerDataContextKey contextKey = 3
const unlockCookieName = "revues_unlock_seen"

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
	HasJira             bool // P3 — Jira configured (org)
	HasNotion           bool // P3 — Notion configured (org)
	HasWebhooks         bool // P3 — webhooks configured (org)
	UnlockFlash         string
	DevAuth             bool
	DevAuthUsers        []store.User
}

// LoadHeaderData preloads organization switcher data for authenticated requests.
// encryptionKey enables P3 capability flags (Jira / Notion / webhooks); may be nil.
func LoadHeaderData(st *store.Store, encryptionKey []byte) func(http.Handler) http.Handler {
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
			caps := resolveUICaps(r.Context(), st, user, hd, encryptionKey)
			hd.SimpleUI = caps.SimpleUI
			hd.SimpleSubjectID = caps.SimpleSubjectID
			hd.ShowAssign = caps.ShowAssign
			hd.ShowMyTasks = caps.ShowMyTasks
			hd.ShowSubjectColumn = caps.ShowSubjectColumn
			hd.ShowCollab = caps.ShowCollab
			hd.HasJira = caps.HasJira
			hd.HasNotion = caps.HasNotion
			hd.HasWebhooks = caps.HasWebhooks
			hd.UnlockFlash = resolveUnlockFlash(w, r, unlockFlashInput{
				caps:               caps,
				whitelistOrgUnlock: isWhitelistOrgUnlock(r.Context(), st, user, hd),
			})

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

type unlockFlashInput struct {
	caps               UICaps
	whitelistOrgUnlock bool
}

// isWhitelistOrgUnlock is true when the Organisation tab appears because a second
// whitelist email was added while the org is still solo (1 member) — not admin/global.
func isWhitelistOrgUnlock(ctx context.Context, st *store.Store, user *store.User, hd HeaderData) bool {
	if user == nil || !hd.ShowOrganisationNav {
		return false
	}
	if auth.HasMinRole(user.Role, auth.RoleAdmin) {
		return false
	}
	org, ok := OrganizationFromContext(ctx)
	if !ok {
		return false
	}
	n, err := st.CountOrganizationMembers(ctx, org.ID)
	if err != nil || n != 1 {
		return false
	}
	allowed, err := st.CountAllowedEmails(ctx)
	return err == nil && allowed > 1
}

func resolveUnlockFlash(w http.ResponseWriter, r *http.Request, in unlockFlashInput) string {
	seen := ""
	if c, err := r.Cookie(unlockCookieName); err == nil {
		seen = c.Value
	}
	level := ""
	msg := ""
	switch {
	case in.caps.ShowSubjectColumn && seen != "p2":
		level = "p2"
		msg = "Plusieurs sujets sont disponibles : la colonne Sujet et le vocabulaire « Modèles » sont maintenant actifs."
	case in.caps.ShowAssign && seen != "p1" && seen != "p2":
		level = "p1"
		msg = "Un second membre a rejoint l'organisation : assignation, Mes tâches et la collaboration sur les sujets sont disponibles."
	case in.whitelistOrgUnlock && seen != "org" && seen != "p1" && seen != "p2":
		level = "org"
		msg = "Un second e-mail a été autorisé : l'onglet Organisation est maintenant disponible."
	}
	if msg == "" {
		return ""
	}
	http.SetCookie(w, &http.Cookie{
		Name:     unlockCookieName,
		Value:    level,
		Path:     "/",
		MaxAge:   int((365 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return msg
}
