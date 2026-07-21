package middleware

import (
	"context"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

// UICaps are progressive-disclosure flags derived from org structure (not a user preference).
type UICaps struct {
	SimpleUI          bool
	SimpleSubjectID   int64
	ShowAssign        bool // ≥2 org members — P1
	ShowMyTasks       bool // ≥2 org members — P1
	ShowSubjectColumn bool // ≥2 visible subjects — P2
	ShowCollab        bool // teams / membres sur fiche sujet — P1+
}

// resolveUICaps detects particulier/solo (SimpleUI) and finer unlocks for duo / multi-sujet.
func resolveUICaps(ctx context.Context, st *store.Store, user *store.User, hd HeaderData) UICaps {
	var caps UICaps
	if user == nil {
		return caps
	}

	org, orgOK := OrganizationFromContext(ctx)
	members := 0
	if orgOK {
		n, err := st.CountOrganizationMembers(ctx, org.ID)
		if err == nil {
			members = n
		}
	}
	caps.ShowAssign = members >= 2
	caps.ShowMyTasks = members >= 2
	caps.ShowCollab = members >= 2

	admin := auth.HasMinRole(user.Role, auth.RoleAdmin)
	ids, subErr := st.ListVisibleSubjectIDs(ctx, user.ID, admin, 2)
	if subErr == nil {
		caps.ShowSubjectColumn = len(ids) >= 2
		if len(ids) == 1 {
			caps.SimpleSubjectID = ids[0]
		}
	}

	// SimpleUI (P0): one org, one member, ≤1 subject, whitelist ≤1, not global admin.
	if admin {
		return caps
	}
	if len(hd.UserOrganizations) != 1 || !orgOK {
		return caps
	}
	if members != 1 {
		return caps
	}
	allowed, err := st.CountAllowedEmails(ctx)
	if err != nil || allowed > 1 {
		return caps
	}
	if subErr != nil || len(ids) > 1 {
		return caps
	}
	caps.SimpleUI = true
	return caps
}

// resolveSimpleUI keeps the previous signature for focused tests.
func resolveSimpleUI(ctx context.Context, st *store.Store, user *store.User, hd HeaderData) (bool, int64) {
	caps := resolveUICaps(ctx, st, user, hd)
	return caps.SimpleUI, caps.SimpleSubjectID
}
