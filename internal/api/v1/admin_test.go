package apiv1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
)

func TestAdminAPI_RBACAndParity(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		SessionSecret:         "test-secret-at-least-thirty-two-bytes",
		LoginRequireWhitelist: false,
		BootstrapAdminEmail:   "admin@example.com",
		Env:                   "development",
	}

	tests := []struct {
		name string
		fn   func(t *testing.T, handler http.Handler, st *store.Store)
	}{
		{
			name: "list allowed emails requires session",
			fn: func(t *testing.T, handler http.Handler, _ *store.Store) {
				req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/allowed-emails", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				assertStatusBody(t, rec, http.StatusUnauthorized, "unauthenticated")
			},
		},
		{
			name: "member without org admin is forbidden",
			fn: func(t *testing.T, handler http.Handler, st *store.Store) {
				ctx := context.Background()
				owner := registerSession(t, handler, "owner-rbac@example.com", "Owner")
				createOrgAPI(t, handler, owner, "RBAC Org", "rbac-org")
				org, err := st.OrganizationBySlug(ctx, "rbac-org")
				if err != nil {
					t.Fatalf("OrganizationBySlug: %v", err)
				}

				member := registerSession(t, handler, "member-rbac@example.com", "Member")
				memberUser, err := st.UserByEmail(ctx, "member-rbac@example.com")
				if err != nil {
					t.Fatalf("UserByEmail: %v", err)
				}
				if err = st.AddOrganizationMember(ctx, org.ID, memberUser.ID, store.OrgRoleMember); err != nil {
					t.Fatalf("AddOrganizationMember: %v", err)
				}
				activateOrgSession(t, handler, st, &member, org.ID)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/allowed-emails", nil)
				req.AddCookie(member.cookie)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				assertStatusBody(t, rec, http.StatusForbidden, "forbidden")
			},
		},
		{
			name: "org admin can CRUD allowed emails",
			fn: func(t *testing.T, handler http.Handler, st *store.Store) {
				owner := registerSession(t, handler, "owner-emails@example.com", "OwnerEmails")
				createOrgAPI(t, handler, owner, "Emails Org", "emails-org")
				owner = refreshSessionCSRF(t, handler, owner.cookie)

				payload := map[string]string{"email": "colleague@example.com", "role": "editor"}
				raw, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/allowed-emails", bytes.NewReader(raw))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-CSRF-Token", owner.csrf)
				req.AddCookie(owner.cookie)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				if rec.Code != http.StatusCreated {
					t.Fatalf("create status=%d body=%s", rec.Code, rec.Body.String())
				}

				reqList := httptest.NewRequest(http.MethodGet, "/api/v1/admin/allowed-emails", nil)
				reqList.AddCookie(owner.cookie)
				recList := httptest.NewRecorder()
				handler.ServeHTTP(recList, reqList)
				if recList.Code != http.StatusOK {
					t.Fatalf("list status=%d body=%s", recList.Code, recList.Body.String())
				}
				var list map[string]any
				_ = json.Unmarshal(recList.Body.Bytes(), &list)
				emails, _ := list["emails"].([]any)
				if len(emails) != 1 {
					t.Fatalf("emails=%v", list["emails"])
				}

				owner = refreshSessionCSRF(t, handler, owner.cookie)
				delPath := "/api/v1/admin/allowed-emails/" + url.PathEscape("colleague@example.com")
				reqDel := httptest.NewRequest(http.MethodDelete, delPath, nil)
				reqDel.Header.Set("X-CSRF-Token", owner.csrf)
				reqDel.AddCookie(owner.cookie)
				recDel := httptest.NewRecorder()
				handler.ServeHTTP(recDel, reqDel)
				if recDel.Code != http.StatusNoContent {
					t.Fatalf("delete status=%d body=%s", recDel.Code, recDel.Body.String())
				}

				// Store scoped to active org should be empty.
				org, err := st.OrganizationBySlug(context.Background(), "emails-org")
				if err != nil {
					t.Fatalf("OrganizationBySlug: %v", err)
				}
				orgCtx := orgctx.WithOrganizationID(context.Background(), org.ID)
				left, err := st.ListAllowedEmails(orgCtx)
				if err != nil || len(left) != 0 {
					t.Fatalf("remaining emails=%v err=%v", left, err)
				}
			},
		},
		{
			name: "create allowed email requires CSRF",
			fn: func(t *testing.T, handler http.Handler, _ *store.Store) {
				owner := registerSession(t, handler, "owner-csrf@example.com", "OwnerCSRF")
				createOrgAPI(t, handler, owner, "CSRF Org", "csrf-org")
				body := `{"email":"x@example.com","role":"reader"}`
				req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/allowed-emails", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				req.AddCookie(owner.cookie)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				assertStatusBody(t, rec, http.StatusForbidden, "csrf")
			},
		},
		{
			name: "cannot delete own whitelist email",
			fn: func(t *testing.T, handler http.Handler, st *store.Store) {
				owner := registerSession(t, handler, "owner-self@example.com", "OwnerSelf")
				createOrgAPI(t, handler, owner, "Self Org", "self-org")
				org, err := st.OrganizationBySlug(context.Background(), "self-org")
				if err != nil {
					t.Fatalf("OrganizationBySlug: %v", err)
				}
				orgCtx := orgctx.WithOrganizationID(context.Background(), org.ID)
				if err = st.InsertAllowedEmail(orgCtx, "owner-self@example.com", auth.RoleEditor); err != nil {
					t.Fatalf("InsertAllowedEmail: %v", err)
				}
				// InsertAllowedEmail revokes sessions for that email — re-login.
				owner = loginSession(t, handler, "owner-self@example.com", "password123")
				activateOrgSession(t, handler, st, &owner, org.ID)

				delPath := "/api/v1/admin/allowed-emails/" + url.PathEscape("owner-self@example.com")
				req := httptest.NewRequest(http.MethodDelete, delPath, nil)
				req.Header.Set("X-CSRF-Token", owner.csrf)
				req.AddCookie(owner.cookie)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				assertStatusBody(t, rec, http.StatusBadRequest, "propre")
			},
		},
		{
			name: "update member role and refuse last owner demotion",
			fn: func(t *testing.T, handler http.Handler, st *store.Store) {
				ctx := context.Background()
				owner := registerSession(t, handler, "owner-roles@example.com", "OwnerRoles")
				createOrgAPI(t, handler, owner, "Roles Org", "roles-org")
				org, err := st.OrganizationBySlug(ctx, "roles-org")
				if err != nil {
					t.Fatalf("OrganizationBySlug: %v", err)
				}
				memberUser, err := st.UpsertGitHubUser(ctx, 9001, "memroles", "mem-roles@example.com", "Mem", "", auth.RoleEditor)
				if err != nil {
					t.Fatalf("UpsertGitHubUser: %v", err)
				}
				if err = st.AddOrganizationMember(ctx, org.ID, memberUser.ID, store.OrgRoleMember); err != nil {
					t.Fatalf("AddOrganizationMember: %v", err)
				}

				owner = refreshSessionCSRF(t, handler, owner.cookie)
				payload := map[string]string{"role": "admin"}
				raw, _ := json.Marshal(payload)
				path := fmt.Sprintf("/api/v1/admin/members/%d", memberUser.ID)
				req := httptest.NewRequest(http.MethodPatch, path, bytes.NewReader(raw))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-CSRF-Token", owner.csrf)
				req.AddCookie(owner.cookie)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				if rec.Code != http.StatusOK {
					t.Fatalf("promote status=%d body=%s", rec.Code, rec.Body.String())
				}

				ownerUserID := mustUserID(t, st, "owner-roles@example.com")
				owner = refreshSessionCSRF(t, handler, owner.cookie)
				demote := map[string]string{"role": "member"}
				rawDemote, _ := json.Marshal(demote)
				reqDemote := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/admin/members/%d", ownerUserID), bytes.NewReader(rawDemote))
				reqDemote.Header.Set("Content-Type", "application/json")
				reqDemote.Header.Set("X-CSRF-Token", owner.csrf)
				reqDemote.AddCookie(owner.cookie)
				recDemote := httptest.NewRecorder()
				handler.ServeHTTP(recDemote, reqDemote)
				assertStatusBody(t, recDemote, http.StatusConflict, "last_owner")
			},
		},
		{
			name: "teams create show add remove member",
			fn: func(t *testing.T, handler http.Handler, st *store.Store) {
				ctx := context.Background()
				owner := registerSession(t, handler, "owner-teams@example.com", "OwnerTeams")
				createOrgAPI(t, handler, owner, "Teams Org", "teams-org")
				org, err := st.OrganizationBySlug(ctx, "teams-org")
				if err != nil {
					t.Fatalf("OrganizationBySlug: %v", err)
				}
				memberUser, err := st.UpsertGitHubUser(ctx, 9002, "teammem", "team-mem@example.com", "TeamMem", "", auth.RoleEditor)
				if err != nil {
					t.Fatalf("UpsertGitHubUser: %v", err)
				}
				if err = st.AddOrganizationMember(ctx, org.ID, memberUser.ID, store.OrgRoleMember); err != nil {
					t.Fatalf("AddOrganizationMember: %v", err)
				}

				owner = refreshSessionCSRF(t, handler, owner.cookie)
				payload := map[string]string{"name": "Qualité", "slug": "qualite", "description": "QA"}
				raw, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/teams", bytes.NewReader(raw))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-CSRF-Token", owner.csrf)
				req.AddCookie(owner.cookie)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				if rec.Code != http.StatusCreated {
					t.Fatalf("create team status=%d body=%s", rec.Code, rec.Body.String())
				}
				var team map[string]any
				_ = json.Unmarshal(rec.Body.Bytes(), &team)
				teamID := int64(team["id"].(float64))

				owner = refreshSessionCSRF(t, handler, owner.cookie)
				addBody, _ := json.Marshal(map[string]any{"user_id": memberUser.ID})
				reqAdd := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/admin/teams/%d/members", teamID), bytes.NewReader(addBody))
				reqAdd.Header.Set("Content-Type", "application/json")
				reqAdd.Header.Set("X-CSRF-Token", owner.csrf)
				reqAdd.AddCookie(owner.cookie)
				recAdd := httptest.NewRecorder()
				handler.ServeHTTP(recAdd, reqAdd)
				if recAdd.Code != http.StatusNoContent {
					t.Fatalf("add member status=%d body=%s", recAdd.Code, recAdd.Body.String())
				}

				reqShow := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/admin/teams/%d", teamID), nil)
				reqShow.AddCookie(owner.cookie)
				recShow := httptest.NewRecorder()
				handler.ServeHTTP(recShow, reqShow)
				if recShow.Code != http.StatusOK {
					t.Fatalf("show status=%d body=%s", recShow.Code, recShow.Body.String())
				}
				var detail map[string]any
				_ = json.Unmarshal(recShow.Body.Bytes(), &detail)
				members, _ := detail["members"].([]any)
				if len(members) != 1 {
					t.Fatalf("members=%v", detail["members"])
				}

				owner = refreshSessionCSRF(t, handler, owner.cookie)
				reqRm := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/admin/teams/%d/members/%d", teamID, memberUser.ID), nil)
				reqRm.Header.Set("X-CSRF-Token", owner.csrf)
				reqRm.AddCookie(owner.cookie)
				recRm := httptest.NewRecorder()
				handler.ServeHTTP(recRm, reqRm)
				if recRm.Code != http.StatusNoContent {
					t.Fatalf("remove status=%d body=%s", recRm.Code, recRm.Body.String())
				}
			},
		},
		{
			name: "lead policies get and put",
			fn: func(t *testing.T, handler http.Handler, _ *store.Store) {
				owner := registerSession(t, handler, "owner-pol@example.com", "OwnerPol")
				createOrgAPI(t, handler, owner, "Pol Org", "pol-org")
				owner = refreshSessionCSRF(t, handler, owner.cookie)

				reqGet := httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings/policies", nil)
				reqGet.AddCookie(owner.cookie)
				recGet := httptest.NewRecorder()
				handler.ServeHTTP(recGet, reqGet)
				if recGet.Code != http.StatusOK {
					t.Fatalf("get policies status=%d body=%s", recGet.Code, recGet.Body.String())
				}

				payload := map[string]bool{
					"leads_may_assign_teams":     false,
					"leads_may_invite_members":   true,
					"leads_may_invite_externals": true,
				}
				raw, _ := json.Marshal(payload)
				reqPut := httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings/policies", bytes.NewReader(raw))
				reqPut.Header.Set("Content-Type", "application/json")
				reqPut.Header.Set("X-CSRF-Token", owner.csrf)
				reqPut.AddCookie(owner.cookie)
				recPut := httptest.NewRecorder()
				handler.ServeHTTP(recPut, reqPut)
				if recPut.Code != http.StatusOK {
					t.Fatalf("put policies status=%d body=%s", recPut.Code, recPut.Body.String())
				}
				var pol map[string]any
				_ = json.Unmarshal(recPut.Body.Bytes(), &pol)
				if pol["leads_may_assign_teams"] != false || pol["leads_may_invite_externals"] != true {
					t.Fatalf("policies=%v", pol)
				}
			},
		},
		{
			name: "team IDOR other org 404",
			fn: func(t *testing.T, handler http.Handler, st *store.Store) {
				ctx := context.Background()
				ownerA := registerSession(t, handler, "owner-a@example.com", "OwnerA")
				createOrgAPI(t, handler, ownerA, "Org A", "org-a-idor")
				orgA, err := st.OrganizationBySlug(ctx, "org-a-idor")
				if err != nil {
					t.Fatalf("OrganizationBySlug A: %v", err)
				}
				orgCtxA := orgctx.WithOrganizationID(ctx, orgA.ID)
				team, err := st.CreateTeam(orgCtxA, "Secret", "secret", "")
				if err != nil {
					t.Fatalf("CreateTeam: %v", err)
				}

				ownerB := registerSession(t, handler, "owner-b@example.com", "OwnerB")
				createOrgAPI(t, handler, ownerB, "Org B", "org-b-idor")

				req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/admin/teams/%d", team.ID), nil)
				req.AddCookie(ownerB.cookie)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				assertStatusBody(t, rec, http.StatusNotFound, "not_found")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handler, st := newTestRouterWithStore(t, cfg)
			tt.fn(t, handler, st)
		})
	}
}

func activateOrgSession(t *testing.T, handler http.Handler, st *store.Store, session *testSession, orgID int64) {
	t.Helper()
	*session = refreshSessionCSRF(t, handler, session.cookie)
	payload := map[string]any{"organization_id": orgID}
	raw, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orgs/active", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", session.csrf)
	req.AddCookie(session.cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("activate org: %d %s", rec.Code, rec.Body.String())
	}
	*session = refreshSessionCSRF(t, handler, session.cookie)
	_, active, err := st.SessionByTokenHash(context.Background(), auth.HashToken(session.cookie.Value))
	if err != nil || active != orgID {
		t.Fatalf("session org=%d err=%v want %d", active, err, orgID)
	}
}

func loginSession(t *testing.T, handler http.Handler, email, password string) testSession {
	t.Helper()
	guest, csrf := bootstrapGuest(t, handler)
	payload := map[string]string{"email": email, "password": password}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", csrf)
	req.AddCookie(guest)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login %s: %d %s", email, rec.Code, rec.Body.String())
	}
	var authOK map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &authOK)
	sessionCSRF, _ := authOK["csrf_token"].(string)
	var sessionCookie *http.Cookie
	for _, c := range resultCookies(t, rec) {
		if c.Name == "revues_session" {
			sessionCookie = c
		}
	}
	if sessionCookie == nil || sessionCSRF == "" {
		t.Fatal("missing session after login")
	}
	return testSession{cookie: sessionCookie, csrf: sessionCSRF}
}
