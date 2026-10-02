package apiv1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/features/organizations"
	"github.com/jeb-maker/revues/internal/store"
)

func TestOrgsAPI_TableDriven(t *testing.T) {
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
			name: "list requires session",
			fn: func(t *testing.T, handler http.Handler, _ *store.Store) {
				req := httptest.NewRequest(http.MethodGet, "/api/v1/orgs", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				assertStatusBody(t, rec, http.StatusUnauthorized, "unauthenticated")
			},
		},
		{
			name: "create requires CSRF",
			fn: func(t *testing.T, handler http.Handler, _ *store.Store) {
				session := registerSession(t, handler, "nocsrf@example.com", "No CSRF")
				body := `{"name":"Acme","slug":"acme-nocsrf"}`
				req := httptest.NewRequest(http.MethodPost, "/api/v1/orgs", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				req.AddCookie(session.cookie)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				assertStatusBody(t, rec, http.StatusForbidden, "csrf")
			},
		},
		{
			name: "create first org success",
			fn: func(t *testing.T, handler http.Handler, st *store.Store) {
				session := registerSession(t, handler, "creator@example.com", "Creator")
				payload := map[string]string{"name": "Acme Corp", "slug": "acme-corp"}
				raw, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/orgs", bytes.NewReader(raw))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-CSRF-Token", session.csrf)
				req.AddCookie(session.cookie)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				if rec.Code != http.StatusCreated {
					t.Fatalf("create status=%d body=%s", rec.Code, rec.Body.String())
				}
				var resp map[string]any
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("json: %v", err)
				}
				org, _ := resp["organization"].(map[string]any)
				if org["slug"] != "acme-corp" || org["role"] != store.OrgRoleOwner {
					t.Fatalf("organization = %#v", org)
				}
				if resp["redirect"] != "/" {
					t.Fatalf("redirect = %v", resp["redirect"])
				}
				_, orgID, err := st.SessionByTokenHash(context.Background(), auth.HashToken(session.cookie.Value))
				if err != nil {
					t.Fatalf("session: %v", err)
				}
				dbOrg, err := st.OrganizationBySlug(context.Background(), "acme-corp")
				if err != nil {
					t.Fatalf("OrganizationBySlug: %v", err)
				}
				if orgID != dbOrg.ID {
					t.Fatalf("session org=%d want %d", orgID, dbOrg.ID)
				}
			},
		},
		{
			name: "create second org conflict",
			fn: func(t *testing.T, handler http.Handler, _ *store.Store) {
				session := registerSession(t, handler, "twice@example.com", "Twice")
				createOrgAPI(t, handler, session, "First", "first-org")
				session = refreshSessionCSRF(t, handler, session.cookie)
				payload := map[string]string{"name": "Second", "slug": "second-org"}
				raw, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/orgs", bytes.NewReader(raw))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-CSRF-Token", session.csrf)
				req.AddCookie(session.cookie)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				assertStatusBody(t, rec, http.StatusConflict, "already_has_organization")
			},
		},
		{
			name: "select IDOR non-member 404",
			fn: func(t *testing.T, handler http.Handler, st *store.Store) {
				ctx := context.Background()
				ownerSession := registerSession(t, handler, "owner-idor@example.com", "Owner")
				createOrgAPI(t, handler, ownerSession, "Private Org", "private-org")
				org, err := st.OrganizationBySlug(ctx, "private-org")
				if err != nil {
					t.Fatalf("OrganizationBySlug: %v", err)
				}

				stranger := registerSession(t, handler, "stranger@example.com", "Stranger")
				payload := map[string]any{"organization_id": org.ID}
				raw, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/orgs/active", bytes.NewReader(raw))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-CSRF-Token", stranger.csrf)
				req.AddCookie(stranger.cookie)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				assertStatusBody(t, rec, http.StatusNotFound, "not_found")
			},
		},
		{
			name: "select member success",
			fn: func(t *testing.T, handler http.Handler, st *store.Store) {
				ctx := context.Background()
				userSession := registerSession(t, handler, "picker@example.com", "Picker")
				user, err := st.UserByEmail(ctx, "picker@example.com")
				if err != nil {
					t.Fatalf("UserByEmail: %v", err)
				}
				orgA, err := st.CreateOrganization(ctx, "Alpha", "alpha-pick", user.ID)
				if err != nil {
					t.Fatalf("CreateOrganization A: %v", err)
				}
				orgB, err := st.CreateOrganization(ctx, "Beta", "beta-pick", user.ID)
				if err != nil {
					t.Fatalf("CreateOrganization B: %v", err)
				}
				for _, org := range []*store.Organization{orgA, orgB} {
					if err = st.AddOrganizationMember(ctx, org.ID, user.ID, store.OrgRoleMember); err != nil {
						t.Fatalf("AddOrganizationMember: %v", err)
					}
				}

				payload := map[string]any{"organization_id": orgB.ID}
				raw, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/orgs/active", bytes.NewReader(raw))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-CSRF-Token", userSession.csrf)
				req.AddCookie(userSession.cookie)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				if rec.Code != http.StatusOK {
					t.Fatalf("select status=%d body=%s", rec.Code, rec.Body.String())
				}
				_, orgID, err := st.SessionByTokenHash(ctx, auth.HashToken(userSession.cookie.Value))
				if err != nil {
					t.Fatalf("session: %v", err)
				}
				if orgID != orgB.ID {
					t.Fatalf("session org=%d want %d", orgID, orgB.ID)
				}
			},
		},
		{
			name: "accept invitation success",
			fn: func(t *testing.T, handler http.Handler, st *store.Store) {
				ctx := context.Background()
				ownerSession := registerSession(t, handler, "inv-owner@example.com", "InvOwner")
				createOrgAPI(t, handler, ownerSession, "Gamma", "gamma-inv")
				org, err := st.OrganizationBySlug(ctx, "gamma-inv")
				if err != nil {
					t.Fatalf("OrganizationBySlug: %v", err)
				}
				if err = st.CreateOrganizationInvitation(ctx, "invitee@example.com", org.ID); err != nil {
					t.Fatalf("CreateOrganizationInvitation: %v", err)
				}
				invites, err := st.ListPendingInvitationsByEmail(ctx, "invitee@example.com")
				if err != nil || len(invites) != 1 {
					t.Fatalf("invites=%v err=%v", invites, err)
				}

				invitee := registerSession(t, handler, "invitee@example.com", "Invitee")
				path := fmt.Sprintf("/api/v1/orgs/invitations/%d/accept", invites[0].ID)
				req := httptest.NewRequest(http.MethodPost, path, nil)
				req.Header.Set("X-CSRF-Token", invitee.csrf)
				req.AddCookie(invitee.cookie)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				if rec.Code != http.StatusOK {
					t.Fatalf("accept status=%d body=%s", rec.Code, rec.Body.String())
				}
				role, ok, err := st.OrganizationMemberRole(ctx, org.ID, mustUserID(t, st, "invitee@example.com"))
				if err != nil || !ok || role != store.OrgRoleMember {
					t.Fatalf("member role=%q ok=%v err=%v", role, ok, err)
				}
				_, orgID, err := st.SessionByTokenHash(ctx, auth.HashToken(invitee.cookie.Value))
				if err != nil {
					t.Fatalf("session: %v", err)
				}
				if orgID != org.ID {
					t.Fatalf("session org=%d want %d", orgID, org.ID)
				}
			},
		},
		{
			name: "accept invitation wrong email 404",
			fn: func(t *testing.T, handler http.Handler, st *store.Store) {
				ctx := context.Background()
				ownerSession := registerSession(t, handler, "inv-owner2@example.com", "InvOwner2")
				createOrgAPI(t, handler, ownerSession, "Delta", "delta-inv")
				org, err := st.OrganizationBySlug(ctx, "delta-inv")
				if err != nil {
					t.Fatalf("OrganizationBySlug: %v", err)
				}
				if err = st.CreateOrganizationInvitation(ctx, "real-invitee@example.com", org.ID); err != nil {
					t.Fatalf("CreateOrganizationInvitation: %v", err)
				}
				invites, err := st.ListPendingInvitationsByEmail(ctx, "real-invitee@example.com")
				if err != nil || len(invites) != 1 {
					t.Fatalf("invites=%v err=%v", invites, err)
				}

				thief := registerSession(t, handler, "thief@example.com", "Thief")
				path := fmt.Sprintf("/api/v1/orgs/invitations/%d/accept", invites[0].ID)
				req := httptest.NewRequest(http.MethodPost, path, nil)
				req.Header.Set("X-CSRF-Token", thief.csrf)
				req.AddCookie(thief.cookie)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				assertStatusBody(t, rec, http.StatusNotFound, "not_found")
			},
		},
		{
			name: "list returns memberships and can_create",
			fn: func(t *testing.T, handler http.Handler, _ *store.Store) {
				session := registerSession(t, handler, "lister@example.com", "Lister")
				req := httptest.NewRequest(http.MethodGet, "/api/v1/orgs", nil)
				req.AddCookie(session.cookie)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				if rec.Code != http.StatusOK {
					t.Fatalf("list status=%d body=%s", rec.Code, rec.Body.String())
				}
				var resp map[string]any
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("json: %v", err)
				}
				if resp["can_create"] != true {
					t.Fatalf("can_create = %v", resp["can_create"])
				}
				createOrgAPI(t, handler, session, "Listed", "listed-org")
				session = refreshSessionCSRF(t, handler, session.cookie)
				req2 := httptest.NewRequest(http.MethodGet, "/api/v1/orgs", nil)
				req2.AddCookie(session.cookie)
				rec2 := httptest.NewRecorder()
				handler.ServeHTTP(rec2, req2)
				if rec2.Code != http.StatusOK {
					t.Fatalf("list2 status=%d body=%s", rec2.Code, rec2.Body.String())
				}
				var resp2 map[string]any
				_ = json.Unmarshal(rec2.Body.Bytes(), &resp2)
				if resp2["can_create"] != false {
					t.Fatalf("can_create after create = %v", resp2["can_create"])
				}
				orgs, _ := resp2["organizations"].([]any)
				if len(orgs) != 1 {
					t.Fatalf("organizations len=%d", len(orgs))
				}
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

func TestOrganizations_PostLoginRoute(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler, st := newTestRouterWithStore(t, config.Config{
		SessionSecret:         "test-secret-at-least-thirty-two-bytes",
		LoginRequireWhitelist: false,
		BootstrapAdminEmail:   "admin@example.com",
		Env:                   "development",
	})
	_ = handler

	userNoOrg, err := st.UpsertGitHubUser(ctx, 910, "solo", "solo-route@example.com", "Solo", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser: %v", err)
	}
	userOne, err := st.UpsertGitHubUser(ctx, 911, "one", "one-route@example.com", "One", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser: %v", err)
	}
	orgOne, err := st.CreateOrganization(ctx, "Solo Org", "solo-route-org", userOne.ID)
	if err != nil {
		t.Fatalf("CreateOrganization: %v", err)
	}
	if err = st.AddOrganizationMember(ctx, orgOne.ID, userOne.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember: %v", err)
	}
	userMany, err := st.UpsertGitHubUser(ctx, 912, "many", "many-route@example.com", "Many", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser: %v", err)
	}
	for _, slug := range []string{"alpha-route", "beta-route"} {
		org, err := st.CreateOrganization(ctx, strings.ToUpper(slug), slug, userMany.ID)
		if err != nil {
			t.Fatalf("CreateOrganization(%s): %v", slug, err)
		}
		if err := st.AddOrganizationMember(ctx, org.ID, userMany.ID, store.OrgRoleMember); err != nil {
			t.Fatalf("AddOrganizationMember(%s): %v", slug, err)
		}
	}

	tests := []struct {
		name       string
		userID     int64
		wantOrgArg int64
		wantPath   string
	}{
		{name: "zero organizations", userID: userNoOrg.ID, wantOrgArg: auth.SessionOrgPending, wantPath: "/org/new"},
		{name: "one organization", userID: userOne.ID, wantOrgArg: orgOne.ID, wantPath: "/"},
		{name: "many organizations", userID: userMany.ID, wantOrgArg: auth.SessionOrgPending, wantPath: "/org/select"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOrg, gotPath, err := organizations.PostLoginRoute(ctx, st, tt.userID)
			if err != nil {
				t.Fatalf("PostLoginRoute error = %v", err)
			}
			if gotOrg != tt.wantOrgArg || gotPath != tt.wantPath {
				t.Fatalf("PostLoginRoute() = (%d, %q), want (%d, %q)", gotOrg, gotPath, tt.wantOrgArg, tt.wantPath)
			}
		})
	}
}

func assertStatusBody(t *testing.T, rec *httptest.ResponseRecorder, status int, substr string) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, status, rec.Body.String())
	}
	if substr != "" && !strings.Contains(rec.Body.String(), substr) {
		t.Fatalf("body = %q, want substring %q", rec.Body.String(), substr)
	}
}

type testSession struct {
	cookie *http.Cookie
	csrf   string
}

func registerSession(t *testing.T, handler http.Handler, email, displayName string) testSession {
	t.Helper()
	guest, csrf := bootstrapGuest(t, handler)
	payload := map[string]string{
		"email":            email,
		"display_name":     displayName,
		"password":         "password123",
		"password_confirm": "password123",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", csrf)
	req.AddCookie(guest)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register %s: %d %s", email, rec.Code, rec.Body.String())
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
		t.Fatal("missing session after register")
	}
	return testSession{cookie: sessionCookie, csrf: sessionCSRF}
}

func createOrgAPI(t *testing.T, handler http.Handler, session testSession, name, slug string) {
	t.Helper()
	payload := map[string]string{"name": name, "slug": slug}
	raw, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orgs", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", session.csrf)
	req.AddCookie(session.cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create org %s: %d %s", slug, rec.Code, rec.Body.String())
	}
}

func refreshSessionCSRF(t *testing.T, handler http.Handler, cookie *http.Cookie) testSession {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/bootstrap", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("bootstrap: %d", rec.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	csrf, _ := body["csrf_token"].(string)
	if csrf == "" {
		t.Fatal("missing csrf")
	}
	return testSession{cookie: cookie, csrf: csrf}
}

func mustUserID(t *testing.T, st *store.Store, email string) int64 {
	t.Helper()
	u, err := st.UserByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("UserByEmail(%s): %v", email, err)
	}
	return u.ID
}
