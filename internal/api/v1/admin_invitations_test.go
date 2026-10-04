package apiv1_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
)

func TestAdminInvitationsAPI(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		SessionSecret:         "test-secret-at-least-thirty-two-bytes",
		LoginRequireWhitelist: true,
		BootstrapAdminEmail:   "admin@example.com",
		Env:                   "development",
	}
	handler, st := newTestRouterWithStore(t, cfg)

	owner, _, _ := seedSessionUser(t, st, cfg, "inv-owner@example.com", "Owner", auth.RoleEditor, true)
	orgID := mustDefaultOrgID(t, st)
	ctx := context.Background()
	if err := st.AddOrganizationMember(ctx, orgID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("promote owner: %v", err)
	}
	sessions := &auth.SessionManager{Store: st, SessionSecret: cfg.SessionSecret}
	token, ownerCSRF, err := sessions.CreateLoginSession(ctx, owner.ID, orgID)
	if err != nil {
		t.Fatalf("relogin: %v", err)
	}
	ownerSession := &http.Cookie{Name: "revues_session", Value: token}

	member, memberSession, memberCSRF := seedSessionUser(t, st, cfg, "inv-member@example.com", "Member", auth.RoleEditor, true)
	_ = member
	_ = memberCSRF

	// Member non-admin → 403
	forbid := doJSON(t, handler, http.MethodPost, "/api/v1/admin/invitations", map[string]any{
		"email": "someone@example.com",
	}, memberSession, memberCSRF)
	if forbid.Code != http.StatusForbidden {
		t.Fatalf("member create status=%d body=%s", forbid.Code, forbid.Body.String())
	}

	// Create invitation (normalization trim+lower côté serveur)
	create := doJSON(t, handler, http.MethodPost, "/api/v1/admin/invitations", map[string]any{
		"email":    "New.Invitee@Example.COM",
		"org_role": "admin",
	}, ownerSession, ownerCSRF)
	if create.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", create.Code, create.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(create.Body.Bytes(), &created)
	if created["email"] != "new.invitee@example.com" {
		t.Fatalf("email normalized = %v", created["email"])
	}
	if created["org_role"] != "admin" {
		t.Fatalf("org_role = %v", created["org_role"])
	}
	invID := int64(created["id"].(float64))

	// Idempotent re-invite
	again := doJSON(t, handler, http.MethodPost, "/api/v1/admin/invitations", map[string]any{
		"email":    "new.invitee@example.com",
		"org_role": "member",
	}, ownerSession, ownerCSRF)
	if again.Code != http.StatusCreated {
		t.Fatalf("reinvite status=%d body=%s", again.Code, again.Body.String())
	}

	list := doJSON(t, handler, http.MethodGet, "/api/v1/admin/invitations", nil, ownerSession, "")
	if list.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
	if !jsonContainsEmail(list.Body.Bytes(), "new.invitee@example.com") {
		t.Fatalf("list missing invitee: %s", list.Body.String())
	}

	// Already member → 409
	conflict := doJSON(t, handler, http.MethodPost, "/api/v1/admin/invitations", map[string]any{
		"email": member.Email,
	}, ownerSession, ownerCSRF)
	if conflict.Code != http.StatusConflict {
		t.Fatalf("already member status=%d body=%s", conflict.Code, conflict.Body.String())
	}

	// Invite-only gate still allows register via pending invitation
	ok, err := st.HasPendingInvitationByEmail(ctx, "new.invitee@example.com")
	if err != nil || !ok {
		t.Fatalf("pending invite missing: %v %v", ok, err)
	}

	// IDOR: invitation of other org → 404 on delete
	otherOrg, err := st.CreateOrganization(ctx, "Other Inv Org", "other-inv-org", owner.ID)
	if err != nil {
		t.Fatalf("CreateOrganization: %v", err)
	}
	if err = st.AddOrganizationMember(ctx, otherOrg.ID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember other: %v", err)
	}
	otherCtx := orgctx.WithOrganizationID(ctx, otherOrg.ID)
	if err = st.CreateOrganizationInvitation(otherCtx, "cross@example.com", otherOrg.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("CreateOrganizationInvitation other: %v", err)
	}
	otherInvites, err := st.ListPendingInvitationsByOrganization(otherCtx, otherOrg.ID)
	if err != nil || len(otherInvites) == 0 {
		t.Fatalf("other invites: %v %#v", err, otherInvites)
	}
	idor := doJSON(t, handler, http.MethodDelete,
		fmt.Sprintf("/api/v1/admin/invitations/%d", otherInvites[0].ID),
		nil, ownerSession, ownerCSRF)
	if idor.Code != http.StatusNotFound {
		t.Fatalf("cross-org delete status=%d body=%s", idor.Code, idor.Body.String())
	}

	// Revoke own invitation
	del := doJSON(t, handler, http.MethodDelete,
		fmt.Sprintf("/api/v1/admin/invitations/%d", invID),
		nil, ownerSession, ownerCSRF)
	if del.Code != http.StatusNoContent {
		// re-invite may have changed id — delete from list
		list2 := doJSON(t, handler, http.MethodGet, "/api/v1/admin/invitations", nil, ownerSession, "")
		var body map[string]any
		_ = json.Unmarshal(list2.Body.Bytes(), &body)
		invs, _ := body["invitations"].([]any)
		if len(invs) == 0 {
			t.Fatalf("no invitations to revoke")
		}
		id := int64(invs[0].(map[string]any)["id"].(float64))
		del = doJSON(t, handler, http.MethodDelete,
			fmt.Sprintf("/api/v1/admin/invitations/%d", id),
			nil, ownerSession, ownerCSRF)
		if del.Code != http.StatusNoContent {
			t.Fatalf("delete status=%d body=%s", del.Code, del.Body.String())
		}
	}
}

func jsonContainsEmail(raw []byte, email string) bool {
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		return false
	}
	invs, _ := body["invitations"].([]any)
	for _, rawInv := range invs {
		inv, _ := rawInv.(map[string]any)
		if inv["email"] == email {
			return true
		}
	}
	return false
}
