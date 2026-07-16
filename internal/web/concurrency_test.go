package web_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
	appweb "github.com/jeb-maker/revues/internal/web"
)

type httpLoadFixture struct {
	handler http.Handler
	token   string
	runID   int64
}

func TestConcurrentHTTPProjectListNoLock(t *testing.T) {
	f := newHTTPRunLoadFixture(t)
	runHTTPConcurrentGET(t, f, "/subjects", 24, 40)
}

func TestConcurrentHTTPRunShowNoLock(t *testing.T) {
	f := newHTTPRunLoadFixture(t)
	runHTTPConcurrentGET(t, f, fmt.Sprintf("/runs/%d", f.runID), 24, 40)
}

func runHTTPConcurrentGET(t *testing.T, f httpLoadFixture, path string, workers, iterations int) {
	t.Helper()

	var lockErrors atomic.Int64
	var serverErrors atomic.Int64
	var wg sync.WaitGroup
	wg.Add(workers)

	for range workers {
		go func() {
			defer wg.Done()
			for range iterations {
				req := httptest.NewRequest(http.MethodGet, path, nil)
				req.AddCookie(&http.Cookie{Name: "revues_session", Value: f.token})
				rec := httptest.NewRecorder()
				f.handler.ServeHTTP(rec, req)

				if rec.Code != http.StatusOK {
					serverErrors.Add(1)
					t.Errorf("GET %s status = %d", path, rec.Code)
					return
				}
				if isSQLiteLockBody(rec.Body.String()) {
					lockErrors.Add(1)
				}
			}
		}()
	}

	wg.Wait()

	if n := lockErrors.Load(); n > 0 {
		t.Fatalf("lock errors on GET %s: %d", path, n)
	}
	if n := serverErrors.Load(); n > 0 {
		t.Fatalf("server errors on GET %s: %d", path, n)
	}
}

func newHTTPRunLoadFixture(t *testing.T) httpLoadFixture {
	t.Helper()

	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/http_load.db", store.DefaultMaxOpenConns)
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Errorf("Close(): %v", closeErr)
		}
	})
	if migrateErr := store.Migrate(ctx, db); migrateErr != nil {
		t.Fatalf("Migrate(): %v", migrateErr)
	}

	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	cfg := config.Config{
		Addr:           ":8080",
		BaseURL:        "http://example.com",
		SessionSecret:  testSessionSecret,
		Env:            "development",
		DBMaxOpenConns: store.DefaultMaxOpenConns,
	}

	handler, _, err := appweb.NewRouter(appweb.Deps{Config: cfg, DB: db})
	if err != nil {
		t.Fatalf("NewRouter(): %v", err)
	}

	user, err := st.UpsertGitHubUser(ctx, 1, "load", "load@example.com", "Load", "", auth.RoleAdmin)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(default): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, user.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}
	project, err := st.CreateProject(ctx, "Load", "", user.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle", user.ID, nil, []store.TemplateItemInput{
		{Section: "S", Label: "Point 1", Required: true},
		{Section: "S", Label: "Point 2", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, user.ID)
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	if err = st.StartRun(ctx, run.ID); err != nil {
		t.Fatalf("StartRun(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: testSessionSecret}
	token, _, err := sessions.CreateLoginSession(ctx, user.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	return httpLoadFixture{handler: handler, token: token, runID: run.ID}
}

func isSQLiteLockBody(body string) bool {
	lower := strings.ToLower(body)
	return strings.Contains(lower, "database is locked") ||
		strings.Contains(lower, "sqlite_busy") ||
		strings.Contains(lower, "database table is locked")
}
