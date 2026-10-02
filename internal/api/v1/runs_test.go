package apiv1_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
)

func TestRunsAPI_LaunchSnapshotUpdateCompleteAndGuards(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		SessionSecret:         "test-secret-at-least-thirty-two-bytes",
		LoginRequireWhitelist: false,
		BootstrapAdminEmail:   "admin@example.com",
		Env:                   "development",
	}
	handler, st := newTestRouterWithStore(t, cfg)

	editor, session, csrf := seedSessionUser(t, st, cfg, "runs-editor@example.com", "Runs Editor", auth.RoleEditor, true)
	_ = editor

	// Template via API
	tplRec := doJSON(t, handler, http.MethodPost, "/api/v1/templates", map[string]any{
		"name":    "Contrôle prod",
		"domains": []string{"ops"},
		"items": []map[string]any{
			{"section": "Sécurité", "label": "Backup", "required": true},
			{"section": "Sécurité", "label": "TLS", "required": false},
		},
	}, session, csrf)
	if tplRec.Code != http.StatusCreated {
		t.Fatalf("create template status = %d body=%s", tplRec.Code, tplRec.Body.String())
	}
	var tpl map[string]any
	_ = json.Unmarshal(tplRec.Body.Bytes(), &tpl)
	templateID := int64(tpl["id"].(float64))

	// Subject matching domain
	subRec := doJSON(t, handler, http.MethodPost, "/api/v1/subjects", map[string]any{
		"name":    "API Gateway",
		"domains": []string{"ops"},
	}, session, csrf)
	if subRec.Code != http.StatusCreated {
		t.Fatalf("create subject status = %d body=%s", subRec.Code, subRec.Body.String())
	}
	var sub map[string]any
	_ = json.Unmarshal(subRec.Body.Bytes(), &sub)
	subjectID := int64(sub["id"].(float64))
	caps := sub["capabilities"].(map[string]any)
	if caps["can_launch"] != true {
		t.Fatalf("can_launch = %v, want true", caps["can_launch"])
	}

	// Launch templates list
	tplList := doJSON(t, handler, http.MethodGet,
		fmt.Sprintf("/api/v1/subjects/%d/run-templates", subjectID), nil, session, "")
	if tplList.Code != http.StatusOK {
		t.Fatalf("run-templates status = %d body=%s", tplList.Code, tplList.Body.String())
	}
	if !strings.Contains(tplList.Body.String(), "Contrôle prod") {
		t.Fatalf("run-templates body = %s", tplList.Body.String())
	}

	// Launch run = transactional snapshot
	launchRec := doJSON(t, handler, http.MethodPost,
		fmt.Sprintf("/api/v1/subjects/%d/runs", subjectID),
		map[string]any{"template_id": templateID, "due_date": "2030-01-15"},
		session, csrf)
	if launchRec.Code != http.StatusCreated {
		t.Fatalf("launch status = %d body=%s", launchRec.Code, launchRec.Body.String())
	}
	var run map[string]any
	_ = json.Unmarshal(launchRec.Body.Bytes(), &run)
	runID := int64(run["id"].(float64))
	items := run["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("snapshot items = %d, want 2", len(items))
	}
	if run["status"] != "in_progress" {
		t.Fatalf("status = %v", run["status"])
	}
	progress := run["progress"].(map[string]any)
	if int(progress["total"].(float64)) != 2 || int(progress["done"].(float64)) != 0 {
		t.Fatalf("progress = %v", progress)
	}

	item0 := items[0].(map[string]any)
	item1 := items[1].(map[string]any)
	itemID0 := int64(item0["id"].(float64))
	itemID1 := int64(item1["id"].(float64))

	// nok without comment → 400
	nokNoComment := doJSON(t, handler, http.MethodPatch,
		fmt.Sprintf("/api/v1/runs/%d/items/%d", runID, itemID0),
		map[string]any{"status": "nok", "comment": ""},
		session, csrf)
	if nokNoComment.Code != http.StatusBadRequest {
		t.Fatalf("nok without comment status = %d body=%s", nokNoComment.Code, nokNoComment.Body.String())
	}

	// Complete blocked while required pending
	completeEarly := doJSON(t, handler, http.MethodPost,
		fmt.Sprintf("/api/v1/runs/%d/complete", runID),
		map[string]any{"closing_note": "trop tôt"},
		session, csrf)
	if completeEarly.Code != http.StatusBadRequest {
		t.Fatalf("complete early status = %d body=%s", completeEarly.Code, completeEarly.Body.String())
	}

	// Mark required ok
	okRec := doJSON(t, handler, http.MethodPatch,
		fmt.Sprintf("/api/v1/runs/%d/items/%d", runID, itemID0),
		map[string]any{"status": "ok", "comment": ""},
		session, csrf)
	if okRec.Code != http.StatusOK {
		t.Fatalf("patch ok status = %d body=%s", okRec.Code, okRec.Body.String())
	}
	var itemDetail map[string]any
	_ = json.Unmarshal(okRec.Body.Bytes(), &itemDetail)
	events := itemDetail["events"].([]any)
	if len(events) < 1 {
		t.Fatalf("expected audit event, got %v", itemDetail["events"])
	}

	// Optional nok with comment
	nokRec := doJSON(t, handler, http.MethodPatch,
		fmt.Sprintf("/api/v1/runs/%d/items/%d", runID, itemID1),
		map[string]any{"status": "nok", "comment": "Certificat expire bientôt"},
		session, csrf)
	if nokRec.Code != http.StatusOK {
		t.Fatalf("patch nok status = %d body=%s", nokRec.Code, nokRec.Body.String())
	}

	// Complete
	completeRec := doJSON(t, handler, http.MethodPost,
		fmt.Sprintf("/api/v1/runs/%d/complete", runID),
		map[string]any{"closing_note": "Revue OK avec alerte TLS"},
		session, csrf)
	if completeRec.Code != http.StatusOK {
		t.Fatalf("complete status = %d body=%s", completeRec.Code, completeRec.Body.String())
	}
	var doneRun map[string]any
	_ = json.Unmarshal(completeRec.Body.Bytes(), &doneRun)
	if doneRun["status"] != "done" {
		t.Fatalf("completed status = %v", doneRun["status"])
	}

	// Done run not editable
	patchDone := doJSON(t, handler, http.MethodPatch,
		fmt.Sprintf("/api/v1/runs/%d/items/%d", runID, itemID0),
		map[string]any{"status": "na"},
		session, csrf)
	if patchDone.Code != http.StatusConflict {
		t.Fatalf("patch done status = %d, want 409; body=%s", patchDone.Code, patchDone.Body.String())
	}

	// List runs includes completed
	listRec := doJSON(t, handler, http.MethodGet, "/api/v1/runs?status=done", nil, session, "")
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", listRec.Code, listRec.Body.String())
	}
	if !strings.Contains(listRec.Body.String(), strconv.FormatInt(runID, 10)) {
		t.Fatalf("list missing run: %s", listRec.Body.String())
	}
}

func TestRunsAPI_IDORAndCSRF(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		SessionSecret:         "test-secret-at-least-thirty-two-bytes",
		LoginRequireWhitelist: false,
		BootstrapAdminEmail:   "admin@example.com",
		Env:                   "development",
	}
	handler, st := newTestRouterWithStore(t, cfg)

	alice, aliceSession, aliceCSRF := seedSessionUser(t, st, cfg, "runs-alice@example.com", "Alice", auth.RoleEditor, true)
	bob, bobSession, _ := seedSessionUser(t, st, cfg, "runs-bob@example.com", "Bob", auth.RoleEditor, true)
	_ = bob

	tplRec := doJSON(t, handler, http.MethodPost, "/api/v1/templates", map[string]any{
		"name": "T",
		"items": []map[string]any{
			{"label": "A", "required": true},
		},
	}, aliceSession, aliceCSRF)
	if tplRec.Code != http.StatusCreated {
		t.Fatalf("template: %d %s", tplRec.Code, tplRec.Body.String())
	}
	var tpl map[string]any
	_ = json.Unmarshal(tplRec.Body.Bytes(), &tpl)

	subRec := doJSON(t, handler, http.MethodPost, "/api/v1/subjects", map[string]any{
		"name": "Sujet Alice",
	}, aliceSession, aliceCSRF)
	if subRec.Code != http.StatusCreated {
		t.Fatalf("subject: %d %s", subRec.Code, subRec.Body.String())
	}
	var sub map[string]any
	_ = json.Unmarshal(subRec.Body.Bytes(), &sub)
	subjectID := int64(sub["id"].(float64))

	// Gate the subject: direct membership ends org_member_legacy for others.
	addRec := doJSON(t, handler, http.MethodPost,
		fmt.Sprintf("/api/v1/subjects/%d/members", subjectID),
		map[string]any{"email": alice.Email, "role": "lead"},
		aliceSession, aliceCSRF)
	if addRec.Code != http.StatusCreated {
		t.Fatalf("add member: %d %s", addRec.Code, addRec.Body.String())
	}

	launch := doJSON(t, handler, http.MethodPost,
		fmt.Sprintf("/api/v1/subjects/%d/runs", subjectID),
		map[string]any{"template_id": int64(tpl["id"].(float64))},
		aliceSession, aliceCSRF)
	if launch.Code != http.StatusCreated {
		t.Fatalf("launch: %d %s", launch.Code, launch.Body.String())
	}
	var run map[string]any
	_ = json.Unmarshal(launch.Body.Bytes(), &run)
	runID := int64(run["id"].(float64))

	// Bob (org member, no subject grant) cannot see
	denied := doJSON(t, handler, http.MethodGet, fmt.Sprintf("/api/v1/runs/%d", runID), nil, bobSession, "")
	if denied.Code != http.StatusNotFound {
		t.Fatalf("idor status = %d, want 404; body=%s", denied.Code, denied.Body.String())
	}

	// CSRF required on launch
	noCSRF := doJSON(t, handler, http.MethodPost,
		fmt.Sprintf("/api/v1/subjects/%d/runs", subjectID),
		map[string]any{"template_id": int64(tpl["id"].(float64))},
		aliceSession, "")
	if noCSRF.Code != http.StatusForbidden {
		t.Fatalf("csrf status = %d, want 403", noCSRF.Code)
	}
}

func TestRunsAPI_RequiresAuth(t *testing.T) {
	t.Parallel()
	handler := newTestRouter(t, config.Config{
		SessionSecret: "test-secret-at-least-thirty-two-bytes",
		Env:           "development",
	})
	rec := doJSON(t, handler, http.MethodGet, "/api/v1/runs", nil, nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}
