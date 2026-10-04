package apiv1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
)

func TestSubjectsAPI_CRUDAndMembers(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		SessionSecret:         "test-secret-at-least-thirty-two-bytes",
		LoginRequireWhitelist: false,
		BootstrapAdminEmail:   "admin@example.com",
		Env:                   "development",
	}
	handler, st := newTestRouterWithStore(t, cfg)
	ctx := context.Background()

	editor, session, csrf := seedSessionUser(t, st, cfg, "editor@example.com", "Editor", auth.RoleEditor, true)
	_ = editor

	// Create
	createBody := map[string]any{
		"name":        "Portail",
		"description": "App web",
		"domains":     []string{"frontend", "k8s"},
		"tags":        []string{"prio"},
		"visibility":  "normal",
	}
	rec := doJSON(t, handler, http.MethodPost, "/api/v1/subjects", createBody, session, csrf)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", rec.Code, rec.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("json: %v", err)
	}
	id := int64(created["id"].(float64))
	if created["name"] != "Portail" {
		t.Fatalf("name = %v", created["name"])
	}
	domains, _ := created["domains"].([]any)
	if len(domains) != 2 {
		t.Fatalf("domains = %v", created["domains"])
	}

	// List
	listRec := doJSON(t, handler, http.MethodGet, "/api/v1/subjects", nil, session, "")
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", listRec.Code, listRec.Body.String())
	}
	if !strings.Contains(listRec.Body.String(), "Portail") {
		t.Fatalf("list body = %s", listRec.Body.String())
	}

	// Get
	getRec := doJSON(t, handler, http.MethodGet, "/api/v1/subjects/"+strconv.FormatInt(id, 10), nil, session, "")
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d body=%s", getRec.Code, getRec.Body.String())
	}

	// Update
	patchBody := map[string]any{
		"name":        "Portail v2",
		"description": "Updated",
		"domains":     []string{"api"},
		"tags":        []string{"beta"},
	}
	patchRec := doJSON(t, handler, http.MethodPatch, "/api/v1/subjects/"+strconv.FormatInt(id, 10), patchBody, session, csrf)
	if patchRec.Code != http.StatusOK {
		t.Fatalf("patch status = %d body=%s", patchRec.Code, patchRec.Body.String())
	}
	if !strings.Contains(patchRec.Body.String(), "Portail v2") {
		t.Fatalf("patch body = %s", patchRec.Body.String())
	}

	// Invitee for members
	invitee, err := st.CreateLocalUser(ctx, "member@example.com", "member", "Member", auth.RoleEditor, mustHash(t, "password1234"))
	if err != nil {
		t.Fatalf("CreateLocalUser(invitee): %v", err)
	}
	orgCtx := orgctx.WithOrganizationID(ctx, mustDefaultOrgID(t, st))
	if err = st.AddOrganizationMember(orgCtx, mustDefaultOrgID(t, st), invitee.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember: %v", err)
	}

	addRec := doJSON(t, handler, http.MethodPost, fmt.Sprintf("/api/v1/subjects/%d/members", id), map[string]any{
		"email": "member@example.com",
		"role":  "contributor",
	}, session, csrf)
	if addRec.Code != http.StatusCreated {
		t.Fatalf("add member status = %d body=%s", addRec.Code, addRec.Body.String())
	}

	memRec := doJSON(t, handler, http.MethodGet, fmt.Sprintf("/api/v1/subjects/%d/members", id), nil, session, "")
	if memRec.Code != http.StatusOK || !strings.Contains(memRec.Body.String(), "member@example.com") {
		t.Fatalf("list members status=%d body=%s", memRec.Code, memRec.Body.String())
	}

	delRec := doJSON(t, handler, http.MethodDelete, fmt.Sprintf("/api/v1/subjects/%d/members/%d", id, invitee.ID), nil, session, csrf)
	if delRec.Code != http.StatusNoContent {
		t.Fatalf("remove member status = %d body=%s", delRec.Code, delRec.Body.String())
	}

	archRec := doJSON(t, handler, http.MethodPost, fmt.Sprintf("/api/v1/subjects/%d/archive", id), nil, session, csrf)
	if archRec.Code != http.StatusNoContent {
		t.Fatalf("archive status = %d body=%s", archRec.Code, archRec.Body.String())
	}
	listAfter := doJSON(t, handler, http.MethodGet, "/api/v1/subjects", nil, session, "")
	if listAfter.Code != http.StatusOK {
		t.Fatalf("list after archive status = %d", listAfter.Code)
	}
	if strings.Contains(listAfter.Body.String(), "Portail v2") {
		t.Fatalf("archived subject still listed: %s", listAfter.Body.String())
	}
}

func TestSubjectsAPI_IDOR_CrossOrg(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		SessionSecret: "test-secret-at-least-thirty-two-bytes",
		Env:           "development",
	}
	handler, st := newTestRouterWithStore(t, cfg)
	ctx := context.Background()

	alice, _, _ := seedSessionUser(t, st, cfg, "alice-idor@example.com", "Alice", auth.RoleEditor, true)
	bob, _, _ := seedSessionUser(t, st, cfg, "bob-idor@example.com", "Bob", auth.RoleEditor, false)

	orgB, err := st.CreateOrganization(ctx, "Org B", "org-b-subjects", bob.ID)
	if err != nil {
		t.Fatalf("CreateOrganization: %v", err)
	}
	if err = st.AddOrganizationMember(ctx, orgB.ID, bob.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember: %v", err)
	}
	// Put bob's session on org B
	sessions := &auth.SessionManager{Store: st, SessionSecret: cfg.SessionSecret}
	bobToken, _, err := sessions.CreateLoginSession(ctx, bob.ID, orgB.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(bob): %v", err)
	}
	bobSession := &http.Cookie{Name: "revues_session", Value: bobToken}

	orgA := mustDefaultOrgID(t, st)
	ctxA := orgctx.WithOrganizationID(ctx, orgA)
	subject, err := st.CreateSubject(ctxA, "Secret A", "", alice.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject: %v", err)
	}

	// Bob (org B) must not see Alice's subject
	rec := doJSON(t, handler, http.MethodGet, fmt.Sprintf("/api/v1/subjects/%d", subject.ID), nil, bobSession, "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("cross-org get status = %d body=%s, want 404", rec.Code, rec.Body.String())
	}

	listRec := doJSON(t, handler, http.MethodGet, "/api/v1/subjects", nil, bobSession, "")
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d", listRec.Code)
	}
	if strings.Contains(listRec.Body.String(), "Secret A") {
		t.Fatalf("bob must not list alice subject: %s", listRec.Body.String())
	}
}

func TestSubjectsAPI_IDOR_PrivateSubject(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		SessionSecret: "test-secret-at-least-thirty-two-bytes",
		Env:           "development",
	}
	handler, st := newTestRouterWithStore(t, cfg)
	ctx := context.Background()
	orgID := mustDefaultOrgID(t, st)
	orgCtx := orgctx.WithOrganizationID(ctx, orgID)

	lead, _, _ := seedSessionUser(t, st, cfg, "lead-priv@example.com", "Lead", auth.RoleEditor, true)
	// Org owner may set visibility=private on create.
	if err := st.AddOrganizationMember(ctx, orgID, lead.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("promote lead to org owner: %v", err)
	}
	// Refresh session so membership is consistent (role already in DB).
	sessions := &auth.SessionManager{Store: st, SessionSecret: cfg.SessionSecret}
	token, csrf, err := sessions.CreateLoginSession(ctx, lead.ID, orgID)
	if err != nil {
		t.Fatalf("relogin lead: %v", err)
	}
	leadSession := &http.Cookie{Name: "revues_session", Value: token}
	leadCSRF := csrf

	member, memberSession, _ := seedSessionUser(t, st, cfg, "member-priv@example.com", "Member", auth.RoleEditor, true)

	createBody := map[string]any{
		"name":       "Private Sujet",
		"visibility": "private",
	}
	rec := doJSON(t, handler, http.MethodPost, "/api/v1/subjects", createBody, leadSession, leadCSRF)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create private status = %d body=%s", rec.Code, rec.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	id := int64(created["id"].(float64))

	// Member without grant → 404
	denied := doJSON(t, handler, http.MethodGet, fmt.Sprintf("/api/v1/subjects/%d", id), nil, memberSession, "")
	if denied.Code != http.StatusNotFound {
		t.Fatalf("private get without grant status = %d body=%s", denied.Code, denied.Body.String())
	}

	// Grant viewer → OK
	if err := st.UpsertDirectSubjectMember(orgCtx, id, member.ID, store.SubjectRoleViewer); err != nil {
		t.Fatalf("UpsertDirectSubjectMember: %v", err)
	}
	allowed := doJSON(t, handler, http.MethodGet, fmt.Sprintf("/api/v1/subjects/%d", id), nil, memberSession, "")
	if allowed.Code != http.StatusOK {
		t.Fatalf("private get with grant status = %d body=%s", allowed.Code, allowed.Body.String())
	}
	_ = lead
}

func TestSubjectsAPI_ReaderCannotCreate(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		SessionSecret: "test-secret-at-least-thirty-two-bytes",
		Env:           "development",
	}
	handler, st := newTestRouterWithStore(t, cfg)
	_, session, csrf := seedSessionUser(t, st, cfg, "reader@example.com", "Reader", auth.RoleReader, true)

	rec := doJSON(t, handler, http.MethodPost, "/api/v1/subjects", map[string]any{"name": "Nope"}, session, csrf)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("reader create status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSubjectsAPI_RequiresAuth(t *testing.T) {
	t.Parallel()

	handler := newTestRouter(t, config.Config{
		SessionSecret: "test-secret-at-least-thirty-two-bytes",
		Env:           "development",
	})
	rec := doJSON(t, handler, http.MethodGet, "/api/v1/subjects", nil, nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func seedSessionUser(t *testing.T, st *store.Store, cfg config.Config, email, name, role string, joinDefault bool) (*store.User, *http.Cookie, string) {
	t.Helper()
	ctx := context.Background()
	user, err := st.CreateLocalUser(ctx, email, strings.Split(email, "@")[0], name, role, mustHash(t, "password1234"))
	if err != nil {
		t.Fatalf("CreateLocalUser(%s): %v", email, err)
	}
	orgID := int64(0)
	if joinDefault {
		orgID = mustDefaultOrgID(t, st)
		if err = st.AddOrganizationMember(ctx, orgID, user.ID, store.OrgRoleMember); err != nil {
			t.Fatalf("AddOrganizationMember: %v", err)
		}
	}
	sessions := &auth.SessionManager{Store: st, SessionSecret: cfg.SessionSecret}
	token, csrf, err := sessions.CreateLoginSession(ctx, user.ID, orgID)
	if err != nil {
		t.Fatalf("CreateLoginSession: %v", err)
	}
	return user, &http.Cookie{Name: "revues_session", Value: token}, csrf
}

func doJSON(t *testing.T, handler http.Handler, method, path string, body any, session *http.Cookie, csrf string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if csrf != "" {
		req.Header.Set("X-CSRF-Token", csrf)
	}
	if session != nil {
		req.AddCookie(session)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func mustHash(t *testing.T, password string) string {
	t.Helper()
	h, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	return h
}

func mustDefaultOrgID(t *testing.T, st *store.Store) int64 {
	t.Helper()
	org, err := st.OrganizationBySlug(context.Background(), "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug: %v", err)
	}
	return org.ID
}
