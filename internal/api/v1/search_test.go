package apiv1_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
)

func TestSearchAPI_MultiKindRBACAndGuards(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		SessionSecret:         "test-secret-at-least-thirty-two-bytes",
		LoginRequireWhitelist: false,
		BootstrapAdminEmail:   "admin@example.com",
		Env:                   "development",
	}
	handler, st := newTestRouterWithStore(t, cfg)

	lead, leadSession, leadCSRF := seedSessionUser(t, st, cfg, "search-lead@example.com", "Lead", auth.RoleEditor, true)
	assignee, assigneeSession, _ := seedSessionUser(t, st, cfg, "search-assignee@example.com", "Assignee", auth.RoleEditor, true)
	_, readerSession, _ := seedSessionUser(t, st, cfg, "search-reader@example.com", "Reader", auth.RoleReader, true)
	_ = lead

	tplRec := doJSON(t, handler, http.MethodPost, "/api/v1/templates", map[string]any{
		"name":    "Modèle AlphaSearch",
		"domains": []string{"alpha-domain"},
		"items": []map[string]any{
			{"section": "Sec", "label": "Point AlphaSearch", "required": true},
		},
	}, leadSession, leadCSRF)
	if tplRec.Code != http.StatusCreated {
		t.Fatalf("create template status = %d body=%s", tplRec.Code, tplRec.Body.String())
	}
	var tpl map[string]any
	_ = json.Unmarshal(tplRec.Body.Bytes(), &tpl)
	templateID := int64(tpl["id"].(float64))

	subRec := doJSON(t, handler, http.MethodPost, "/api/v1/subjects", map[string]any{
		"name":        "Sujet AlphaSearch",
		"description": "desc unique",
		"domains":     []string{"alpha-domain"},
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
	itemID := int64(run["items"].([]any)[0].(map[string]any)["id"].(float64))

	assignRec := doJSON(t, handler, http.MethodPatch,
		fmt.Sprintf("/api/v1/runs/%d/items/%d", runID, itemID),
		map[string]any{"assigned_to": assignee.ID},
		leadSession, leadCSRF)
	if assignRec.Code != http.StatusOK {
		t.Fatalf("assign status = %d body=%s", assignRec.Code, assignRec.Body.String())
	}

	// 401 without session
	unauth := doJSON(t, handler, http.MethodGet, "/api/v1/search?q=AlphaSearch", nil, nil, "")
	if unauth.Code != http.StatusUnauthorized {
		t.Fatalf("unauth status = %d body=%s", unauth.Code, unauth.Body.String())
	}

	// 400 empty q
	emptyQ := doJSON(t, handler, http.MethodGet, "/api/v1/search?q="+url.QueryEscape("   "), nil, leadSession, "")
	if emptyQ.Code != http.StatusBadRequest {
		t.Fatalf("empty q status = %d body=%s", emptyQ.Code, emptyQ.Body.String())
	}

	// Editor: subjects + runs + templates (+ tasks if ≥2 members — yes)
	editorRec := doJSON(t, handler, http.MethodGet, "/api/v1/search?q=AlphaSearch", nil, leadSession, "")
	if editorRec.Code != http.StatusOK {
		t.Fatalf("editor search status = %d body=%s", editorRec.Code, editorRec.Body.String())
	}
	var editorBody map[string]any
	if err := json.Unmarshal(editorRec.Body.Bytes(), &editorBody); err != nil {
		t.Fatalf("json: %v", err)
	}
	kinds := resultKinds(t, editorBody)
	for _, want := range []string{"subject", "run", "template"} {
		if !kinds[want] {
			t.Fatalf("editor missing kind %q in %s", want, editorRec.Body.String())
		}
	}

	// Assignee: task hit + same org entities
	assigneeRec := doJSON(t, handler, http.MethodGet, "/api/v1/search?q=AlphaSearch", nil, assigneeSession, "")
	if assigneeRec.Code != http.StatusOK {
		t.Fatalf("assignee search status = %d body=%s", assigneeRec.Code, assigneeRec.Body.String())
	}
	assigneeKinds := resultKinds(t, mustJSONMap(t, assigneeRec.Body.Bytes()))
	if !assigneeKinds["task"] {
		t.Fatalf("assignee missing task in %s", assigneeRec.Body.String())
	}
	if !strings.Contains(assigneeRec.Body.String(), fmt.Sprintf(`/runs/%d/items/%d`, runID, itemID)) {
		t.Fatalf("assignee task href missing: %s", assigneeRec.Body.String())
	}

	// Org member without project grant: templates OK ; sujet/run absents
	readerRec := doJSON(t, handler, http.MethodGet, "/api/v1/search?q=AlphaSearch", nil, readerSession, "")
	if readerRec.Code != http.StatusOK {
		t.Fatalf("reader search status = %d body=%s", readerRec.Code, readerRec.Body.String())
	}
	readerBody := mustJSONMap(t, readerRec.Body.Bytes())
	readerKinds := resultKinds(t, readerBody)
	if !readerKinds["template"] {
		t.Fatalf("org member should see templates: %s", readerRec.Body.String())
	}
	if readerKinds["subject"] || readerKinds["run"] {
		t.Fatalf("member without grant must not see gated subject/run: %s", readerRec.Body.String())
	}

	// Subject with only creator lead: other members do not see it
	openRec := doJSON(t, handler, http.MethodPost, "/api/v1/subjects", map[string]any{
		"name":    "Ouvert AlphaSearchOpen",
		"domains": []string{"alpha-domain"},
	}, leadSession, leadCSRF)
	if openRec.Code != http.StatusCreated {
		t.Fatalf("create open subject status = %d body=%s", openRec.Code, openRec.Body.String())
	}
	readerOpen := doJSON(t, handler, http.MethodGet, "/api/v1/search?q=AlphaSearchOpen", nil, readerSession, "")
	if readerOpen.Code != http.StatusOK {
		t.Fatalf("reader open search status=%d body=%s", readerOpen.Code, readerOpen.Body.String())
	}
	if strings.Contains(readerOpen.Body.String(), "Ouvert AlphaSearchOpen") {
		t.Fatalf("member without grant must not see creator-only subject: %s", readerOpen.Body.String())
	}

	// IDOR: private subject in same org not visible to non-member
	orgID := mustDefaultOrgID(t, st)
	orgCtx := orgctx.WithOrganizationID(context.Background(), orgID)
	priv, err := st.CreateSubjectWithVisibility(orgCtx, "Secret AlphaSearchPrivate", "", lead.ID, nil, store.SubjectVisibilityPrivate)
	if err != nil {
		t.Fatalf("CreateSubjectWithVisibility: %v", err)
	}
	_ = priv
	privSearch := doJSON(t, handler, http.MethodGet, "/api/v1/search?q=AlphaSearchPrivate", nil, assigneeSession, "")
	if privSearch.Code != http.StatusOK {
		t.Fatalf("private search status = %d", privSearch.Code)
	}
	if strings.Contains(privSearch.Body.String(), "Secret AlphaSearchPrivate") {
		t.Fatalf("assignee must not see private subject: %s", privSearch.Body.String())
	}
	leadPriv := doJSON(t, handler, http.MethodGet, "/api/v1/search?q=AlphaSearchPrivate", nil, leadSession, "")
	if leadPriv.Code != http.StatusOK || !strings.Contains(leadPriv.Body.String(), "Secret AlphaSearchPrivate") {
		t.Fatalf("creator should see private subject: status=%d body=%s", leadPriv.Code, leadPriv.Body.String())
	}

	// Cross-org: subject in other org absent from active-org search
	ctx := context.Background()
	otherOrg, err := st.CreateOrganization(ctx, "Other Search Org", "other-search-org", lead.ID)
	if err != nil {
		t.Fatalf("CreateOrganization: %v", err)
	}
	if err = st.AddOrganizationMember(ctx, otherOrg.ID, lead.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember: %v", err)
	}
	otherCtx := orgctx.WithOrganizationID(ctx, otherOrg.ID)
	_, err = st.CreateSubject(otherCtx, "CrossOrg AlphaSearchX", "", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject other org: %v", err)
	}
	cross := doJSON(t, handler, http.MethodGet, "/api/v1/search?q=AlphaSearchX", nil, leadSession, "")
	if cross.Code != http.StatusOK {
		t.Fatalf("cross-org search status = %d", cross.Code)
	}
	if strings.Contains(cross.Body.String(), "CrossOrg") {
		t.Fatalf("must not leak other org subject: %s", cross.Body.String())
	}
}

func resultKinds(t *testing.T, body map[string]any) map[string]bool {
	t.Helper()
	out := map[string]bool{}
	results, _ := body["results"].([]any)
	for _, raw := range results {
		item, _ := raw.(map[string]any)
		kind, _ := item["kind"].(string)
		out[kind] = true
	}
	return out
}

func mustJSONMap(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("json: %v", err)
	}
	return m
}
