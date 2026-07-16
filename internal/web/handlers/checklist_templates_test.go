package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
)

func TestChecklistTemplates_TaggedTemplateNotInProjectList(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 10, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	projectK8s, err := st.CreateProject(ctx, "K8s", "", lead.ID, []string{"k8s"})
	if err != nil {
		t.Fatalf("CreateProject(k8s): %v", err)
	}
	projectMobile, err := st.CreateProject(ctx, "Mobile", "", lead.ID, []string{"mobile"})
	if err != nil {
		t.Fatalf("CreateProject(mobile): %v", err)
	}

	if _, _, err = st.CreateChecklistTemplate(ctx, "Global", lead.ID, nil, []store.TemplateItemInput{{Label: "A"}}); err != nil {
		t.Fatalf("CreateChecklistTemplate(global): %v", err)
	}
	if _, _, err = st.CreateChecklistTemplate(ctx, "K8s only", lead.ID, []string{"k8s"}, []store.TemplateItemInput{{Label: "B"}}); err != nil {
		t.Fatalf("CreateChecklistTemplate(k8s): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, lead.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	assertList := func(projectID int64, wantNames ...string) {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, "/subjects/"+strconv.FormatInt(projectID, 10)+"/modeles", nil)
		req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("list status = %d, want %d", rec.Code, http.StatusOK)
		}
		body := rec.Body.String()
		for _, name := range wantNames {
			if !strings.Contains(body, name) {
				t.Fatalf("expected %q in project %d list, body=%s", name, projectID, body)
			}
		}
		if len(wantNames) == 1 && strings.Contains(body, "K8s only") && projectID == projectMobile.ID {
			t.Fatal("tagged k8s template must not appear on mobile project list")
		}
	}

	assertList(projectK8s.ID, "Global", "K8s only")
	assertList(projectMobile.ID, "Global")
}

func TestChecklistTemplates_TagMismatchRunCreate(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 11, "lead2", "lead2@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	project, err := st.CreateProject(ctx, "Mobile", "", lead.ID, []string{"mobile"})
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "K8s", lead.ID, []string{"k8s"}, []store.TemplateItemInput{{Label: "Point"}})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, lead.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}
	csrf := auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes")

	form := url.Values{}
	form.Set("csrf_token", csrf)
	form.Set("template_id", strconv.FormatInt(template.ID, 10))
	form.Set("title", "Revue")
	req := httptest.NewRequest(http.MethodPost, "/subjects/"+strconv.FormatInt(project.ID, 10)+"/revues", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d (tag mismatch must block run create)", rec.Code, http.StatusNotFound)
	}
}

func TestChecklistTemplates_CreateAndVersion(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 30, "lead3", "lead3@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "Alpha", "", lead.ID, []string{"k8s"})
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, lead.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}
	csrf := auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes")

	form := url.Values{}
	form.Set("csrf_token", csrf)
	form.Set("name", "Revue code")
	form.Set("tags", "k8s, secu")
	form.Add("section_idx", "0")
	form.Add("section_title", "Général")
	form.Add("item_section_idx", "0")
	form.Add("item_row_idx", "0")
	form.Add("item_label", "Tests OK")
	form.Add("item_help", "Vérifier CI")
	form.Add("item_required", "0")
	req := httptest.NewRequest(http.MethodPost, "/modeles", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("create status = %d, want %d", rec.Code, http.StatusSeeOther)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/subjects/"+strconv.FormatInt(project.ID, 10)+"/modeles", nil)
	listReq.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", listRec.Code, http.StatusOK)
	}
	if !strings.Contains(listRec.Body.String(), "Revue code") {
		t.Fatal("expected untagged-compatible template in list")
	}

	templates, err := st.ListChecklistTemplates(ctx, project.ID)
	if err != nil {
		t.Fatalf("ListChecklistTemplates(): %v", err)
	}
	var createdID int64
	for _, tpl := range templates {
		if tpl.Name == "Revue code" {
			createdID = tpl.ID
		}
	}
	if createdID == 0 {
		t.Fatal("created template not found in store")
	}
	tags, err := st.ListTemplateTags(ctx, createdID)
	if err != nil || len(tags) != 2 || tags[0] != "k8s" || tags[1] != "secu" {
		t.Fatalf("tags = %v, %v", tags, err)
	}

	editForm := url.Values{}
	editForm.Set("csrf_token", csrf)
	editForm.Set("name", "Revue code v2")
	editForm.Set("tags", "k8s")
	editForm.Add("section_idx", "0")
	editForm.Add("item_section_idx", "0")
	editForm.Add("item_row_idx", "0")
	editForm.Add("item_label", "Tests OK bis")
	editForm.Add("item_help", "")
	editForm.Add("item_required", "0")
	saveReq := httptest.NewRequest(http.MethodPost, "/modeles/"+strconv.FormatInt(createdID, 10), strings.NewReader(editForm.Encode()))
	saveReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	saveReq.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	saveRec := httptest.NewRecorder()
	handler.ServeHTTP(saveRec, saveReq)
	if saveRec.Code != http.StatusSeeOther {
		t.Fatalf("save status = %d, want %d", saveRec.Code, http.StatusSeeOther)
	}

	version, err := st.LatestTemplateVersion(ctx, createdID)
	if err != nil {
		t.Fatalf("LatestTemplateVersion(): %v", err)
	}
	if version.Version != 2 {
		t.Fatalf("version = %d, want 2", version.Version)
	}
}

func TestChecklistTemplates_MatchingTagRunCreate(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, _ := st.UpsertGitHubUser(ctx, 31, "lead4", "lead4@example.com", "Lead", "", auth.RoleEditor)
	project, _ := st.CreateProject(ctx, "K8s", "", lead.ID, []string{"k8s"})
	template, _, _ := st.CreateChecklistTemplate(ctx, "Cluster", lead.ID, []string{"k8s"}, []store.TemplateItemInput{{Label: "Point", Required: true}})

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, _ := sessions.CreateLoginSession(ctx, lead.ID, 0)
	csrf := auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes")

	form := url.Values{}
	form.Set("csrf_token", csrf)
	form.Set("template_id", strconv.FormatInt(template.ID, 10))
	form.Set("title", "Revue cluster")
	req := httptest.NewRequest(http.MethodPost, "/subjects/"+strconv.FormatInt(project.ID, 10)+"/revues", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
}

func TestChecklistTemplates_ForRunUsesLaunchMode(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 32, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "Team", "", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	_, _, err = st.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, []store.TemplateItemInput{
		{Label: "Point", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, lead.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/subjects/"+strconv.FormatInt(project.ID, 10)+"/modeles?for_run=1", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `action="/subjects/`+strconv.FormatInt(project.ID, 10)+`/revues"`) {
		t.Fatal("expected POST create form on template list")
	}
	if strings.Contains(body, `href="/modeles/`) {
		t.Fatal("expected launch mode links, not template detail links")
	}
}

func TestChecklistTemplates_ForRunFiltersByQuery(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 33, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "Team", "", lead.ID, []string{"k8s"})
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	if _, _, err = st.CreateChecklistTemplate(ctx, "Revue infra", lead.ID, []string{"k8s"}, []store.TemplateItemInput{{Label: "A"}}); err != nil {
		t.Fatalf("CreateChecklistTemplate(infra): %v", err)
	}
	if _, _, err = st.CreateChecklistTemplate(ctx, "Revue mobile", lead.ID, []string{"mobile"}, []store.TemplateItemInput{{Label: "B"}}); err != nil {
		t.Fatalf("CreateChecklistTemplate(mobile): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, lead.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/subjects/"+strconv.FormatInt(project.ID, 10)+"/modeles?for_run=1&q=infra", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Revue infra") {
		t.Fatal("expected infra template in filtered list")
	}
	if strings.Contains(body, "Revue mobile") {
		t.Fatal("mobile template must not appear when filtered by infra")
	}
	if !strings.Contains(body, `id="filter-query"`) {
		t.Fatal("expected search input on launch template list")
	}
}

func TestChecklistTemplateShow_LaunchCTA(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}

	editor, err := st.UpsertGitHubUser(ctx, 34, "tpl-editor", "tpl-editor@example.com", "Editor", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(editor): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, editor.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(editor): %v", err)
	}

	reader, err := st.UpsertGitHubUser(ctx, 35, "tpl-reader", "tpl-reader@example.com", "Reader", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(reader): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, reader.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(reader): %v", err)
	}

	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle CTA", editor.ID, nil, []store.TemplateItemInput{{Label: "Point"}})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}

	assertShow := func(userID int64, wantCTA bool) {
		t.Helper()
		token, _, err := sessions.CreateLoginSession(ctx, userID, defaultOrg.ID)
		if err != nil {
			t.Fatalf("CreateLoginSession(): %v", err)
		}
		req := httptest.NewRequest(http.MethodGet, "/modeles/"+strconv.FormatInt(template.ID, 10), nil)
		req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		body := rec.Body.String()
		hasCTA := strings.Contains(body, "Lancer avec ce modèle") &&
			strings.Contains(body, "/revues/nouvelle?template="+strconv.FormatInt(template.ID, 10))
		if hasCTA != wantCTA {
			t.Fatalf("launch CTA present = %v, want %v", hasCTA, wantCTA)
		}
	}

	assertShow(editor.ID, true)
	assertShow(reader.ID, false)
}

func TestChecklistTemplates_ForRunPreselectsTemplate(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 36, "preselect", "preselect@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "Team", "", lead.ID, []string{"k8s"})
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	selected, _, err := st.CreateChecklistTemplate(ctx, "Revue infra", lead.ID, []string{"k8s"}, []store.TemplateItemInput{{Label: "A"}})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(infra): %v", err)
	}
	if _, _, err = st.CreateChecklistTemplate(ctx, "Revue mobile", lead.ID, []string{"mobile"}, []store.TemplateItemInput{{Label: "B"}}); err != nil {
		t.Fatalf("CreateChecklistTemplate(mobile): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, lead.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/subjects/"+strconv.FormatInt(project.ID, 10)+"/modeles?for_run=1&template="+strconv.FormatInt(selected.ID, 10), nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Revue infra") {
		t.Fatal("expected compatible preselected template in list")
	}
	if strings.Contains(body, "Revue mobile") {
		t.Fatal("incompatible template must not appear when domains do not match")
	}
	if !strings.Contains(body, "présélectionné") {
		t.Fatal("expected preselection hint")
	}
	if !strings.Contains(body, `aria-current="true"`) {
		t.Fatal("expected highlighted row for preselected template")
	}
}
