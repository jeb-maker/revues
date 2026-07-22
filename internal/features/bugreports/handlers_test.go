package bugreports_test

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/features/bugreports"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
	appweb "github.com/jeb-maker/revues/internal/web"
)

func testRouter(t *testing.T, attachmentsDir string) (http.Handler, *sql.DB) {
	t.Helper()

	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/test.db", 0)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Errorf("Close() error = %v", closeErr)
		}
	})
	if migrateErr := store.Migrate(ctx, db); migrateErr != nil {
		t.Fatalf("Migrate() error = %v", migrateErr)
	}

	cfg := config.Config{
		Addr:           ":8080",
		BaseURL:        "http://example.com",
		SessionSecret:  "test-secret-at-least-thirty-two-bytes",
		Env:            "development",
		AttachmentsDir: attachmentsDir,
	}

	handler, _, err := appweb.NewRouter(appweb.Deps{Config: cfg, DB: db})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}
	return handler, db
}

func TestBugReport_RequiresAuth(t *testing.T) {
	dir := t.TempDir()
	handler, _ := testRouter(t, filepath.Join(dir, "attachments"))

	req := httptest.NewRequest(http.MethodGet, "/signaler", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("GET status = %d, want %d", rec.Code, http.StatusFound)
	}
	if loc := rec.Header().Get("Location"); !strings.HasPrefix(loc, "/login") {
		t.Fatalf("Location = %q, want /login…", loc)
	}
}

func TestBugReport_CSRFRequired(t *testing.T) {
	dir := t.TempDir()
	handler, db := testRouter(t, filepath.Join(dir, "attachments"))
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	user, err := st.UpsertGitHubUser(ctx, 901, "reader", "reader@example.com", "Reader", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, user.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, user.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	form := url.Values{}
	form.Set("title", "Bug CSRF")
	form.Set("description", "Should be rejected without CSRF")
	req := httptest.NewRequest(http.MethodPost, "/signaler", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("POST without CSRF status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestBugReport_ReaderCanSubmitAndPersist(t *testing.T) {
	dir := t.TempDir()
	attachmentsDir := filepath.Join(dir, "attachments")
	handler, db := testRouter(t, attachmentsDir)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	user, err := st.UpsertGitHubUser(ctx, 902, "reader2", "reader2@example.com", "Reader Two", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, user.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, user.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	form := url.Values{}
	form.Set("csrf_token", auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes"))
	form.Set("title", "Bouton inactif")
	form.Set("description", "Le bouton Enregistrer ne répond pas sur la revue.")
	form.Set("steps", "1. Ouvrir une revue\n2. Cliquer Enregistrer")
	form.Set("severity", "high")
	form.Set("page_url", "/runs/42?tab=items")

	req := httptest.NewRequest(http.MethodPost, "/signaler", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "RevuesTest/1.0")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("POST status = %d, want %d body=%s", rec.Code, http.StatusSeeOther, rec.Body.String())
	}
	loc := rec.Header().Get("Location")
	if !strings.HasPrefix(loc, "/signaler?") || !strings.Contains(loc, "msg=") {
		t.Fatalf("Location = %q, want /signaler?msg=…", loc)
	}
	if !strings.Contains(loc, "from=") {
		t.Fatalf("Location = %q, want from= return path", loc)
	}

	reportsDir := bugreports.ReportsDirFromAttachments(attachmentsDir)
	dayFile := filepath.Join(reportsDir, "bug-reports-"+time.Now().UTC().Format("2006-01-02")+".jsonl")
	f, err := os.Open(dayFile)
	if err != nil {
		t.Fatalf("open report file: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		t.Fatal("expected one JSONL line")
	}
	var report bugreports.Report
	if err := json.Unmarshal(scanner.Bytes(), &report); err != nil {
		t.Fatalf("unmarshal report: %v", err)
	}
	if report.Title != "Bouton inactif" {
		t.Fatalf("title = %q", report.Title)
	}
	if report.UserID != user.ID || report.UserRole != auth.RoleReader {
		t.Fatalf("user = %d/%s, want %d/reader", report.UserID, report.UserRole, user.ID)
	}
	if report.OrgID != defaultOrg.ID || report.OrgName == "" {
		t.Fatalf("org = %d/%q", report.OrgID, report.OrgName)
	}
	if report.PageURL != "/runs/42?tab=items" {
		t.Fatalf("page_url = %q", report.PageURL)
	}
	if report.Severity != "high" {
		t.Fatalf("severity = %q", report.Severity)
	}
	if report.UserAgent != "RevuesTest/1.0" {
		t.Fatalf("user_agent = %q", report.UserAgent)
	}
	if report.ID == "" || report.CreatedAt == "" {
		t.Fatal("missing id or created_at")
	}
}

func TestBugReport_FormShowsWidgetFallback(t *testing.T) {
	dir := t.TempDir()
	handler, db := testRouter(t, filepath.Join(dir, "attachments"))
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	user, err := st.UpsertGitHubUser(ctx, 903, "editor", "editor@example.com", "Editor", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, user.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, user.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/signaler?from=/revues", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	for _, want := range []string{
		"Signaler un problème",
		"data-reports-auto-open",
		"revues-reports-meta",
		"/static/vendor/jeb-maker-reports/reports.min.js",
		"data-reports-open",
		"csrf_token",
		"editor",
		`"app":"revues"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q", want)
		}
	}
}

func TestBugReportAPI_RequiresAuth(t *testing.T) {
	dir := t.TempDir()
	handler, _ := testRouter(t, filepath.Join(dir, "attachments"))

	req := httptest.NewRequest(http.MethodPost, "/signaler/api", strings.NewReader(`{"title":"x","message":"y"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	// No session → CSRF middleware rejects before RequireAuth redirect.
	if rec.Code != http.StatusForbidden {
		t.Fatalf("POST status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestBugReportAPI_CSRFRequired(t *testing.T) {
	dir := t.TempDir()
	handler, db := testRouter(t, filepath.Join(dir, "attachments"))
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	user, err := st.UpsertGitHubUser(ctx, 904, "apiuser", "apiuser@example.com", "API", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, user.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, user.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	body := `{"schemaVersion":1,"type":"bug","title":"T","message":"M"}`
	req := httptest.NewRequest(http.MethodPost, "/signaler/api", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("POST without CSRF status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestBugReportAPI_PersistsAndOverridesIdentity(t *testing.T) {
	dir := t.TempDir()
	attachmentsDir := filepath.Join(dir, "attachments")
	handler, db := testRouter(t, attachmentsDir)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	user, err := st.UpsertGitHubUser(ctx, 905, "widgetuser", "widget@example.com", "Widget", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, user.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, user.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}
	csrf := auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes")

	payload := map[string]any{
		"schemaVersion": 1,
		"id":            "rp_client_1",
		"type":          "bug",
		"title":         "Widget bug",
		"message":       "Bouton cassé sur la revue",
		"page":          map[string]any{"url": "http://example.com/runs/7"},
		"screenshot": map[string]any{
			"status":  "captured",
			"mime":    "image/jpeg",
			"dataUrl": "data:image/jpeg;base64," + strings.Repeat("A", 200),
			"bytes":   200,
		},
		"metadata": map[string]any{
			"app":        "revues",
			"user_id":    99999,
			"user_login": "attacker",
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/signaler/api", strings.NewReader(string(raw)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", csrf)
	req.Header.Set("User-Agent", "ReportsWidget/1.0")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var resp map[string]any
	if unmarshalErr := json.Unmarshal(rec.Body.Bytes(), &resp); unmarshalErr != nil {
		t.Fatalf("unmarshal response: %v", unmarshalErr)
	}
	if resp["ok"] != true {
		t.Fatalf("ok = %v", resp["ok"])
	}
	if resp["id"] == nil || resp["id"] == "" {
		t.Fatal("missing id in response")
	}

	reportsDir := bugreports.ReportsDirFromAttachments(attachmentsDir)
	dayFile := filepath.Join(reportsDir, "bug-reports-"+time.Now().UTC().Format("2006-01-02")+".jsonl")
	f, err := os.Open(dayFile)
	if err != nil {
		t.Fatalf("open report file: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		t.Fatal("expected one JSONL line")
	}
	var report bugreports.Report
	if err := json.Unmarshal(scanner.Bytes(), &report); err != nil {
		t.Fatalf("unmarshal report: %v", err)
	}
	if report.Title != "Widget bug" || report.Description != "Bouton cassé sur la revue" {
		t.Fatalf("title/desc = %q / %q", report.Title, report.Description)
	}
	if report.Source != "widget" || report.ReportType != "bug" {
		t.Fatalf("source/type = %q / %q", report.Source, report.ReportType)
	}
	if report.ClientID != "rp_client_1" {
		t.Fatalf("client_id = %q", report.ClientID)
	}
	if report.UserID != user.ID || report.UserLogin != "widgetuser" {
		t.Fatalf("trusted user = %d/%s, want %d/widgetuser (client metadata must not win)", report.UserID, report.UserLogin, user.ID)
	}
	if report.OrgID != defaultOrg.ID {
		t.Fatalf("org_id = %d, want %d", report.OrgID, defaultOrg.ID)
	}
	if report.PageURL != "/runs/7" {
		t.Fatalf("page_url = %q", report.PageURL)
	}
	if len(report.Payload) == 0 {
		t.Fatal("expected compact payload")
	}
	if strings.Contains(string(report.Payload), "data:image/jpeg") {
		t.Fatal("screenshot dataUrl must be omitted from stored payload")
	}
	if !strings.Contains(string(report.Payload), `"omitted":true`) {
		t.Fatalf("payload missing omitted screenshot flag: %s", report.Payload)
	}
}
