package apiv1_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"

	apiv1 "github.com/jeb-maker/revues/internal/api/v1"
	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
	appweb "github.com/jeb-maker/revues/internal/web"
)

// fakeRunNotifier captures RunNotifier calls made by the API handlers.
type fakeRunNotifier struct {
	mu        sync.Mutex
	completed []int64
	assigned  [][2]int64 // runID, itemID
}

var _ apiv1.RunNotifier = (*fakeRunNotifier)(nil)

func (f *fakeRunNotifier) NotifyRunCompleted(_ context.Context, runID int64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.completed = append(f.completed, runID)
}

func (f *fakeRunNotifier) NotifyItemAssigned(_ context.Context, runID, itemID int64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.assigned = append(f.assigned, [2]int64{runID, itemID})
}

func (f *fakeRunNotifier) counts() (completed, assigned int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.completed), len(f.assigned)
}

func newTestRouterWithNotifier(t *testing.T, cfg config.Config, notifier apiv1.RunNotifier) (http.Handler, *store.Store) {
	t.Helper()

	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/test.db", 0)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err = store.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	st := store.New(db)
	_ = testutil.DefaultOrgContext(ctx, st)

	handler, _, _, err := appweb.NewRouter(appweb.Deps{Config: cfg, DB: db, Notifier: notifier})
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	return handler, st
}

// runItemFixture is one in-progress run with several optional items, owned by a lead who can
// update/assign, plus a contributor member usable as assignee.
type runItemFixture struct {
	handler     http.Handler
	session     *http.Cookie
	csrf        string
	runID       int64
	itemIDs     []int64
	assigneeID  int64
	notifier    *fakeRunNotifier
	itemPath    func(itemID int64) string
	currentItem func(t *testing.T, itemID int64) map[string]any
}

func newRunItemFixture(t *testing.T, itemCount int) runItemFixture {
	t.Helper()

	cfg := config.Config{
		SessionSecret:         "test-secret-at-least-thirty-two-bytes",
		LoginRequireWhitelist: false,
		BootstrapAdminEmail:   "admin@example.com",
		Env:                   "development",
	}
	notifier := &fakeRunNotifier{}
	handler, st := newTestRouterWithNotifier(t, cfg, notifier)

	lead, session, csrf := seedSessionUser(t, st, cfg, "notify-lead@example.com", "Lead", auth.RoleEditor, true)
	contrib, _, _ := seedSessionUser(t, st, cfg, "notify-contrib@example.com", "Contrib", auth.RoleEditor, true)

	items := make([]map[string]any, 0, itemCount)
	for i := 0; i < itemCount; i++ {
		items = append(items, map[string]any{"label": fmt.Sprintf("Point %d", i+1), "required": false})
	}
	tplRec := doJSON(t, handler, http.MethodPost, "/api/v1/templates", map[string]any{
		"name":  "Notif",
		"items": items,
	}, session, csrf)
	if tplRec.Code != http.StatusCreated {
		t.Fatalf("template: %d %s", tplRec.Code, tplRec.Body.String())
	}
	var tpl map[string]any
	_ = json.Unmarshal(tplRec.Body.Bytes(), &tpl)

	subRec := doJSON(t, handler, http.MethodPost, "/api/v1/subjects", map[string]any{"name": "Sujet notif"}, session, csrf)
	if subRec.Code != http.StatusCreated {
		t.Fatalf("subject: %d %s", subRec.Code, subRec.Body.String())
	}
	var sub map[string]any
	_ = json.Unmarshal(subRec.Body.Bytes(), &sub)
	subjectID := int64(sub["id"].(float64))

	for _, m := range []struct {
		email, role string
	}{{lead.Email, "lead"}, {contrib.Email, "contributor"}} {
		rec := doJSON(t, handler, http.MethodPost,
			fmt.Sprintf("/api/v1/subjects/%d/members", subjectID),
			map[string]any{"email": m.email, "role": m.role}, session, csrf)
		if rec.Code != http.StatusCreated {
			t.Fatalf("add member %s: %d %s", m.email, rec.Code, rec.Body.String())
		}
	}

	launch := doJSON(t, handler, http.MethodPost,
		fmt.Sprintf("/api/v1/subjects/%d/runs", subjectID),
		map[string]any{"template_id": int64(tpl["id"].(float64))}, session, csrf)
	if launch.Code != http.StatusCreated {
		t.Fatalf("launch: %d %s", launch.Code, launch.Body.String())
	}
	var run map[string]any
	_ = json.Unmarshal(launch.Body.Bytes(), &run)
	runID := int64(run["id"].(float64))
	var itemIDs []int64
	for _, raw := range run["items"].([]any) {
		itemIDs = append(itemIDs, int64(raw.(map[string]any)["id"].(float64)))
	}
	if len(itemIDs) != itemCount {
		t.Fatalf("items = %d, want %d", len(itemIDs), itemCount)
	}

	itemPath := func(itemID int64) string {
		return fmt.Sprintf("/api/v1/runs/%d/items/%d", runID, itemID)
	}
	currentItem := func(t *testing.T, itemID int64) map[string]any {
		t.Helper()
		rec := doJSON(t, handler, http.MethodGet, itemPath(itemID), nil, session, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("get item %d: %d %s", itemID, rec.Code, rec.Body.String())
		}
		var detail map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &detail)
		return detail["item"].(map[string]any)
	}

	return runItemFixture{
		handler:     handler,
		session:     session,
		csrf:        csrf,
		runID:       runID,
		itemIDs:     itemIDs,
		assigneeID:  contrib.ID,
		notifier:    notifier,
		itemPath:    itemPath,
		currentItem: currentItem,
	}
}

func TestRunsAPI_Notifications(t *testing.T) {
	t.Parallel()
	fx := newRunItemFixture(t, 4)
	const stale = "2000-01-01T00:00:00Z"

	// Sequential table: cases share items on purpose (assign → re-assign → unassign).
	tests := []struct {
		name         string
		item         int
		body         func(current map[string]any) map[string]any
		wantCode     int
		wantAssigned int // delta of NotifyItemAssigned calls
	}{
		{
			name:     "status only does not notify",
			item:     0,
			body:     func(map[string]any) map[string]any { return map[string]any{"status": "ok"} },
			wantCode: http.StatusOK,
		},
		{
			name:         "assign notifies once",
			item:         1,
			body:         func(map[string]any) map[string]any { return map[string]any{"assigned_to": fx.assigneeID} },
			wantCode:     http.StatusOK,
			wantAssigned: 1,
		},
		{
			name: "re-saving the same assignee does not notify again",
			item: 1,
			body: func(map[string]any) map[string]any {
				return map[string]any{"assigned_to": fx.assigneeID, "status": "ok"}
			},
			wantCode: http.StatusOK,
		},
		{
			name:     "unassign does not notify",
			item:     1,
			body:     func(map[string]any) map[string]any { return map[string]any{"unassign": true} },
			wantCode: http.StatusOK,
		},
		{
			name: "validation failure does not notify",
			item: 2,
			body: func(map[string]any) map[string]any {
				return map[string]any{"status": "nok", "comment": "", "assigned_to": fx.assigneeID}
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "stale updated_at conflicts and does not notify",
			item: 2,
			body: func(map[string]any) map[string]any {
				return map[string]any{"assigned_to": fx.assigneeID, "updated_at": stale}
			},
			wantCode: http.StatusConflict,
		},
		{
			name: "status + assign with fresh updated_at notifies once",
			item: 3,
			body: func(current map[string]any) map[string]any {
				return map[string]any{"status": "ok", "assigned_to": fx.assigneeID, "updated_at": current["updated_at"]}
			},
			wantCode:     http.StatusOK,
			wantAssigned: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itemID := fx.itemIDs[tt.item]
			_, assignedBefore := fx.notifier.counts()

			rec := doJSON(t, fx.handler, http.MethodPatch, fx.itemPath(itemID), tt.body(fx.currentItem(t, itemID)), fx.session, fx.csrf)
			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantCode, rec.Body.String())
			}

			completed, assignedAfter := fx.notifier.counts()
			if completed != 0 {
				t.Fatalf("NotifyRunCompleted calls = %d before completion", completed)
			}
			if got := assignedAfter - assignedBefore; got != tt.wantAssigned {
				t.Fatalf("NotifyItemAssigned calls = %d, want %d", got, tt.wantAssigned)
			}
			if tt.wantAssigned == 1 {
				fx.notifier.mu.Lock()
				last := fx.notifier.assigned[len(fx.notifier.assigned)-1]
				fx.notifier.mu.Unlock()
				if last != [2]int64{fx.runID, itemID} {
					t.Fatalf("NotifyItemAssigned args = %v, want [%d %d]", last, fx.runID, itemID)
				}
			}
		})
	}

	t.Run("complete notifies once", func(t *testing.T) {
		rec := doJSON(t, fx.handler, http.MethodPost,
			fmt.Sprintf("/api/v1/runs/%d/complete", fx.runID),
			map[string]any{"closing_note": "ok"}, fx.session, fx.csrf)
		if rec.Code != http.StatusOK {
			t.Fatalf("complete: %d %s", rec.Code, rec.Body.String())
		}
		completed, _ := fx.notifier.counts()
		if completed != 1 {
			t.Fatalf("NotifyRunCompleted calls = %d, want 1", completed)
		}
		fx.notifier.mu.Lock()
		got := fx.notifier.completed[0]
		fx.notifier.mu.Unlock()
		if got != fx.runID {
			t.Fatalf("NotifyRunCompleted run = %d, want %d", got, fx.runID)
		}
	})
}

func TestRunsAPI_UpdateItemOptimisticLock(t *testing.T) {
	t.Parallel()
	fx := newRunItemFixture(t, 3)
	const stale = "2000-01-01T00:00:00Z"

	tests := []struct {
		name       string
		item       int
		updatedAt  func(current string) any // nil = omit the field
		wantCode   int
		wantStatus string
	}{
		{"fresh updated_at applies", 0, func(c string) any { return c }, http.StatusOK, "ok"},
		{"stale updated_at → 409 conflict, nothing written", 1, func(string) any { return stale }, http.StatusConflict, "pending"},
		{"omitted updated_at keeps unconditional behaviour", 2, nil, http.StatusOK, "ok"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itemID := fx.itemIDs[tt.item]
			before := fx.currentItem(t, itemID)
			body := map[string]any{"status": "ok"}
			if tt.updatedAt != nil {
				body["updated_at"] = tt.updatedAt(before["updated_at"].(string))
			}

			rec := doJSON(t, fx.handler, http.MethodPatch, fx.itemPath(itemID), body, fx.session, fx.csrf)
			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantCode, rec.Body.String())
			}
			if tt.wantCode == http.StatusConflict {
				var errBody map[string]map[string]string
				_ = json.Unmarshal(rec.Body.Bytes(), &errBody)
				if errBody["error"]["code"] != "conflict" {
					t.Fatalf("error code = %q, want conflict", errBody["error"]["code"])
				}
				if errBody["error"]["message"] != "Ce point a été modifié entre-temps. Rechargez la page." {
					t.Fatalf("message = %q", errBody["error"]["message"])
				}
			}

			after := fx.currentItem(t, itemID)
			if after["status"] != tt.wantStatus {
				t.Fatalf("item status = %v, want %s", after["status"], tt.wantStatus)
			}
			if tt.wantCode == http.StatusConflict && after["updated_at"] != before["updated_at"] {
				t.Fatalf("updated_at changed on conflict: %v → %v", before["updated_at"], after["updated_at"])
			}
		})
	}
}
