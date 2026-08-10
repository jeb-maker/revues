package middleware

import (
	"context"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

// UICaps are progressive-disclosure flags derived from org structure (P0–P2)
// and capability gates (P3). Not a user preference.
type UICaps struct {
	SimpleUI          bool
	SimpleSubjectID   int64
	ShowAssign        bool // ≥2 org members — P1
	ShowMyTasks       bool // ≥2 org members — P1
	ShowSubjectColumn bool // ≥2 visible subjects — P2
	ShowCollab        bool // teams / membres sur fiche sujet — P1+
	// P3 — conformité (org config). Independent of SimpleUI.
	HasJira     bool
	HasNotion   bool
	HasWebhooks bool
}

// resolveUICaps detects particulier/solo (SimpleUI), finer unlocks for duo / multi-sujet,
// and P3 capability flags when encryptionKey is set.
func resolveUICaps(ctx context.Context, st *store.Store, user *store.User, hd HeaderData, encryptionKey []byte) UICaps {
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

	caps.HasJira, caps.HasNotion, caps.HasWebhooks = resolveCapabilityCaps(ctx, st, encryptionKey)

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
	caps := resolveUICaps(ctx, st, user, hd, nil)
	return caps.SimpleUI, caps.SimpleSubjectID
}
