package apiv1_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
)

func TestMyTasksAPI_ListFilterSearchAndGuards(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		SessionSecret:         "test-secret-at-least-thirty-two-bytes",
		LoginRequireWhitelist: false,
		BootstrapAdminEmail:   "admin@example.com",
		Env:                   "development",
	}
	handler, st := newTestRouterWithStore(t, cfg)

	lead, leadSession, leadCSRF := seedSessionUser(t, st, cfg, "tasks-lead@example.com", "Lead", auth.RoleEditor, true)
	assignee, assigneeSession, _ := seedSessionUser(t, st, cfg, "tasks-assignee@example.com", "Assignee", auth.RoleEditor, true)
	_ = lead

	tplRec := doJSON(t, handler, http.MethodPost, "/api/v1/templates", map[string]any{
		"name":    "Checklist ops",
		"domains": []string{"ops"},
		"items": []map[string]any{
			{"section": "Infra", "label": "Backup quotidien", "required": true},
			{"section": "Infra", "label": "Certificat TLS", "required": false},
		},
	}, leadSession, leadCSRF)
	if tplRec.Code != http.StatusCreated {
		t.Fatalf("create template status = %d body=%s", tplRec.Code, tplRec.Body.String())
	}
	var tpl map[string]any
	_ = json.Unmarshal(tplRec.Body.Bytes(), &tpl)
	templateID := int64(tpl["id"].(float64))

	subRec := doJSON(t, handler, http.MethodPost, "/api/v1/subjects", map[string]any{
		"name":    "Passerelle API",
		"domains": []string{"ops"},
	}, leadSession, leadCSRF)
	if subRec.Code != http.StatusCreated {
		t.Fatalf("create subject status = %d body=%s", subRec.Code, subRec.Body.String())
	}
	var sub map[string]any
	_ = json.Unmarshal(subRec.Body.Bytes(), &sub)
	subjectID := int64(sub["id"].(float64))

	addRec := doJSON(t, handler, http.MethodPost,
		fmt.Sprintf("/api/v1/subjects/%d/members", subjectID),
		map[string]any{"email": assignee.Email, "role": "contributor"},
		leadSession, leadCSRF)
	if addRec.Code != http.StatusCreated {
		t.Fatalf("add member status = %d body=%s", addRec.Code, addRec.Body.String())
	}

	launchRec := doJSON(t, handler, http.MethodPost,
		fmt.Sprintf("/api/v1/subjects/%d/runs", subjectID),
		map[string]any{"template_id": templateID},
		leadSession, leadCSRF)
	if launchRec.Code != http.StatusCreated {
		t.Fatalf("launch status = %d body=%s", launchRec.Code, launchRec.Body.String())
	}
	var run map[string]any
	_ = json.Unmarshal(launchRec.Body.Bytes(), &run)
	runID := int64(run["id"].(float64))
	items := run["items"].([]any)
	if len(items) < 2 {
		t.Fatalf("items = %d, want >= 2", len(items))
	}
	item0 := items[0].(map[string]any)
	item1 := items[1].(map[string]any)
	itemID0 := int64(item0["id"].(float64))
	itemID1 := int64(item1["id"].(float64))

	assign0 := doJSON(t, handler, http.MethodPatch,
		fmt.Sprintf("/api/v1/runs/%d/items/%d", runID, itemID0),
		map[string]any{"assigned_to": assignee.ID},
		leadSession, leadCSRF)
	if assign0.Code != http.StatusOK {
		t.Fatalf("assign0 status = %d body=%s", assign0.Code, assign0.Body.String())
	}
	assign1 := doJSON(t, handler, http.MethodPatch,
		fmt.Sprintf("/api/v1/runs/%d/items/%d", runID, itemID1),
		map[string]any{"assigned_to": assignee.ID},
		leadSession, leadCSRF)
	if assign1.Code != http.StatusOK {
		t.Fatalf("assign1 status = %d body=%s", assign1.Code, assign1.Body.String())
	}

	assigneeCSRF := csrfForSession(t, handler, assigneeSession)
	okRec := doJSON(t, handler, http.MethodPatch,
		fmt.Sprintf("/api/v1/runs/%d/items/%d", runID, itemID0),
		map[string]any{"status": "ok", "comment": ""},
		assigneeSession, assigneeCSRF)
	if okRec.Code != http.StatusOK {
		t.Fatalf("mark ok status = %d body=%s", okRec.Code, okRec.Body.String())
	}

	tests := []struct {
		name       string
		path       string
		session    *http.Cookie
		wantStatus int
		wantCount  int
		wantLabel  string
	}{
		{
			name:       "list all assigned",
			path:       "/api/v1/me/tasks",
			session:    assigneeSession,
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			name:       "filter pending",
			path:       "/api/v1/me/tasks?status=pending",
			session:    assigneeSession,
			wantStatus: http.StatusOK,
			wantCount:  1,
			wantLabel:  "Certificat TLS",
		},
		{
			name:       "filter ok",
			path:       "/api/v1/me/tasks?status=ok",
			session:    assigneeSession,
			wantStatus: http.StatusOK,
			wantCount:  1,
			wantLabel:  "Backup quotidien",
		},
		{
			name:       "search by label",
			path:       "/api/v1/me/tasks?q=TLS",
			session:    assigneeSession,
			wantStatus: http.StatusOK,
			wantCount:  1,
			wantLabel:  "Certificat TLS",
		},
		{
			name:       "search by subject",
			path:       "/api/v1/me/tasks?q=Passerelle",
			session:    assigneeSession,
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			name:       "lead has no assigned tasks",
			path:       "/api/v1/me/tasks",
			session:    leadSession,
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
		{
			name:       "unauthenticated",
			path:       "/api/v1/me/tasks",
			session:    nil,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doJSON(t, handler, http.MethodGet, tt.path, nil, tt.session, "")
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus != http.StatusOK {
				return
			}
			var body map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("json: %v", err)
			}
			tasks, _ := body["tasks"].([]any)
			if len(tasks) != tt.wantCount {
				t.Fatalf("tasks = %d, want %d body=%s", len(tasks), tt.wantCount, rec.Body.String())
			}
			if tt.wantLabel != "" {
				found := false
				for _, raw := range tasks {
					task := raw.(map[string]any)
					if task["label"] == tt.wantLabel {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected label %q in %s", tt.wantLabel, rec.Body.String())
				}
			}
			if tt.wantCount > 0 {
				task := tasks[0].(map[string]any)
				if task["run_id"] == nil || task["subject_id"] == nil || task["run_title"] == "" {
					t.Fatalf("missing run/subject fields: %v", task)
				}
				if !strings.Contains(fmt.Sprint(task["run_title"]), "Passerelle") &&
					!strings.Contains(fmt.Sprint(task["subject_name"]), "Passerelle") {
					t.Fatalf("expected subject context in task: %v", task)
				}
			}
		})
	}
}

func TestMyTasksAPI_RequiresOrg(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		SessionSecret:         "test-secret-at-least-thirty-two-bytes",
		LoginRequireWhitelist: false,
		BootstrapAdminEmail:   "admin@example.com",
		Env:                   "development",
	}
	handler, st := newTestRouterWithStore(t, cfg)

	ctx := context.Background()
	user, err := st.CreateLocalUser(ctx, "tasks-noorg@example.com", "tasks-noorg", "No Org", auth.RoleEditor, mustHash(t, "password1234"))
	if err != nil {
		t.Fatalf("CreateLocalUser: %v", err)
	}
	sessions := &auth.SessionManager{Store: st, SessionSecret: cfg.SessionSecret}
	token, _, err := sessions.CreateLoginSession(ctx, user.ID, auth.SessionOrgPending)
	if err != nil {
		t.Fatalf("CreateLoginSession pending: %v", err)
	}
	session := &http.Cookie{Name: "revues_session", Value: token}

	rec := doJSON(t, handler, http.MethodGet, "/api/v1/me/tasks", nil, session, "")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "org_required") {
		t.Fatalf("body = %s, want org_required", rec.Body.String())
	}
}

func csrfForSession(t *testing.T, handler http.Handler, session *http.Cookie) string {
	t.Helper()
	rec := doJSON(t, handler, http.MethodGet, "/api/v1/bootstrap", nil, session, "")
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	csrf, _ := body["csrf_token"].(string)
	if csrf == "" {
		t.Fatal("expected csrf_token from bootstrap")
	}
	return csrf
}
