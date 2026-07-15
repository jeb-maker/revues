package handlers_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/jeb-maker/revues/internal/testutil"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/store"
	appweb "github.com/jeb-maker/revues/internal/web"
)

func TestUpload_RejectsUnsupportedType(t *testing.T) {
	handler, db, _ := testRouterAttachments(t)
	st := store.New(db)
	ctx := testutil.DefaultOrgContext(context.Background(), st)
	run, item, token, csrf := seedRunItemForUpload(t, ctx, st)
	body, ct := multipartUpload(t, csrf, "bad.txt", []byte("x"))
	req := httptest.NewRequest(http.MethodPost, uploadURL(run.ID, item.ID), body)
	req.Header.Set("Content-Type", ct)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "non autorisé") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpload_RejectsOversize(t *testing.T) {
	handler, db, _ := testRouterAttachments(t)
	st := store.New(db)
	ctx := testutil.DefaultOrgContext(context.Background(), st)
	run, item, token, csrf := seedRunItemForUpload(t, ctx, st)
	data := make([]byte, 5*1024*1024+1)
	data[0], data[1], data[2] = 0xFF, 0xD8, 0xFF
	body, ct := multipartUpload(t, csrf, "big.jpg", data)
	req := httptest.NewRequest(http.MethodPost, uploadURL(run.ID, item.ID), body)
	req.Header.Set("Content-Type", ct)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "5 Mo") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpload_SuccessJPEG(t *testing.T) {
	handler, db, dir := testRouterAttachments(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	run, item, token, csrf := seedRunItemForUpload(t, ctx, st)
	src := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var imgBuf bytes.Buffer
	_ = jpeg.Encode(&imgBuf, src, nil)
	body, ct := multipartUpload(t, csrf, "proof.jpg", imgBuf.Bytes())
	req := httptest.NewRequest(http.MethodPost, uploadURL(run.ID, item.ID), body)
	req.Header.Set("Content-Type", ct)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	att, err := st.AttachmentByRunItemID(ctx, item.ID)
	if err != nil || att.Filename != "proof.jpg" {
		t.Fatalf("attachment=%+v err=%v", att, err)
	}
	if _, err := os.Stat(filepath.Join(dir, att.StoragePath)); err != nil {
		t.Fatalf("file missing: %v", err)
	}
}

func TestDownloadAttachment_InlineImage(t *testing.T) {
	handler, db, _ := testRouterAttachments(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	run, item, token, csrf := seedRunItemForUpload(t, ctx, st)
	var imgBuf bytes.Buffer
	_ = jpeg.Encode(&imgBuf, image.NewRGBA(image.Rect(0, 0, 4, 4)), nil)
	body, ct := multipartUpload(t, csrf, "shot.jpg", imgBuf.Bytes())
	uploadReq := httptest.NewRequest(http.MethodPost, uploadURL(run.ID, item.ID), body)
	uploadReq.Header.Set("Content-Type", ct)
	uploadReq.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	uploadRec := httptest.NewRecorder()
	handler.ServeHTTP(uploadRec, uploadReq)
	att, err := st.AttachmentByRunItemID(ctx, item.ID)
	if err != nil {
		t.Fatalf("AttachmentByRunItemID(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/attachments/%d?inline=1", att.ID), nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	if cd := rec.Header().Get("Content-Disposition"); cd != "inline" {
		t.Fatalf("Content-Disposition=%q want inline", cd)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/jpeg" {
		t.Fatalf("Content-Type=%q", ct)
	}
}

func TestDownloadAttachment_PDFDownload(t *testing.T) {
	handler, db, _ := testRouterAttachments(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	run, item, token, csrf := seedRunItemForUpload(t, ctx, st)
	body, ct := multipartUpload(t, csrf, "doc.pdf", []byte("%PDF-1.4 test"))
	uploadReq := httptest.NewRequest(http.MethodPost, uploadURL(run.ID, item.ID), body)
	uploadReq.Header.Set("Content-Type", ct)
	uploadReq.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	handler.ServeHTTP(httptest.NewRecorder(), uploadReq)
	att, err := st.AttachmentByRunItemID(ctx, item.ID)
	if err != nil {
		t.Fatalf("AttachmentByRunItemID(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/attachments/%d", att.ID), nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, `attachment; filename="doc.pdf"`) {
		t.Fatalf("Content-Disposition=%q", cd)
	}
}

func TestIDOR_CrossProjectAttachmentDownload(t *testing.T) {
	handler, db, dir := testRouterAttachments(t)
	ctx := context.Background()
	st := store.New(db)
	leadA, _ := st.UpsertGitHubUser(ctx, 701, "a", "a@ex.com", "A", "", auth.RoleEditor)
	leadB, _ := st.UpsertGitHubUser(ctx, 702, "b", "b@ex.com", "B", "", auth.RoleEditor)
	ctx = testutil.SetupIsolatedOrg(ctx, st, "Org B", "org-b-attach", leadB.ID)
	pB, _ := st.CreateProject(ctx, "B", "", leadB.ID, nil)
	tpl, _, _ := st.CreateChecklistTemplate(ctx, "T", leadB.ID, nil, []store.TemplateItemInput{{Label: "P"}})
	runB, _ := st.CreateChecklistRun(ctx, pB.ID, tpl.ID, leadB.ID)
	_ = st.StartRun(ctx, runB.ID)
	itemsB, _ := st.ListRunItems(ctx, runB.ID)
	att, err := st.ReplaceAttachment(ctx, itemsB[0].ID, "secret.pdf", "application/pdf", "secret.pdf", 10)
	if err != nil {
		t.Fatalf("ReplaceAttachment(): %v", err)
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("MkdirAll(): %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, att.StoragePath), []byte("%PDF-1"), 0o640); err != nil {
		t.Fatalf("WriteFile(): %v", err)
	}
	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	tokenA, _, _ := sessions.CreateLoginSession(ctx, leadA.ID, 0)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/attachments/%d", att.ID), nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: tokenA})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d want 404", rec.Code)
	}
}

func TestIDOR_CrossProjectAttachmentUpload(t *testing.T) {
	handler, db, _ := testRouterAttachments(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	leadA, _ := st.UpsertGitHubUser(ctx, 501, "a", "a@ex.com", "A", "", auth.RoleEditor)
	leadB, _ := st.UpsertGitHubUser(ctx, 502, "b", "b@ex.com", "B", "", auth.RoleEditor)
	pB, _ := st.CreateProject(ctx, "B", "", leadB.ID, nil)
	tpl, _, _ := st.CreateChecklistTemplate(ctx, "T", leadB.ID, nil, []store.TemplateItemInput{{Label: "P"}})
	runB, _ := st.CreateChecklistRun(ctx, pB.ID, tpl.ID, leadB.ID)
	_ = st.StartRun(ctx, runB.ID)
	itemsB, _ := st.ListRunItems(ctx, runB.ID)
	pA, _ := st.CreateProject(ctx, "A", "", leadA.ID, nil)
	tplA, _, _ := st.CreateChecklistTemplate(ctx, "T", leadA.ID, nil, []store.TemplateItemInput{{Label: "P"}})
	runA, _ := st.CreateChecklistRun(ctx, pA.ID, tplA.ID, leadA.ID)
	_ = st.StartRun(ctx, runA.ID)
	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, _ := sessions.CreateLoginSession(ctx, leadA.ID, 0)
	csrf := auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes")
	var imgBuf bytes.Buffer
	_ = jpeg.Encode(&imgBuf, image.NewRGBA(image.Rect(0, 0, 8, 8)), nil)
	body, ct := multipartUpload(t, csrf, "x.jpg", imgBuf.Bytes())
	req := httptest.NewRequest(http.MethodPost, uploadURL(runA.ID, itemsB[0].ID), body)
	req.Header.Set("Content-Type", ct)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d want 404", rec.Code)
	}
}

func testRouterAttachments(t *testing.T) (http.Handler, *sql.DB, string) {
	t.Helper()
	ctx := context.Background()
	dir := t.TempDir() + "/attachments"
	db, _ := store.Open(ctx, t.TempDir()+"/test.db", 0)
	t.Cleanup(func() { _ = db.Close() })
	_ = store.Migrate(ctx, db)
	h, _, _ := appweb.NewRouter(appweb.Deps{Config: config.Config{
		Addr: ":8080", BaseURL: "http://example.com", SessionSecret: "test-secret-at-least-thirty-two-bytes",
		Env: "development", AttachmentsDir: dir,
	}, DB: db})
	return h, db, dir
}

func seedRunItemForUpload(t *testing.T, ctx context.Context, st *store.Store) (*store.ChecklistRun, store.RunItem, string, string) {
	t.Helper()
	lead, _ := st.UpsertGitHubUser(ctx, 600, "lead", "up@ex.com", "L", "", auth.RoleEditor)
	p, _ := st.CreateProject(ctx, "P", "", lead.ID, nil)
	tpl, _, _ := st.CreateChecklistTemplate(ctx, "M", lead.ID, nil, []store.TemplateItemInput{{Label: "P"}})
	run, _ := st.CreateChecklistRun(ctx, p.ID, tpl.ID, lead.ID)
	_ = st.StartRun(ctx, run.ID)
	items, _ := st.ListRunItems(ctx, run.ID)
	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, _ := sessions.CreateLoginSession(ctx, lead.ID, 0)
	return run, items[0], token, auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes")
}

func uploadURL(runID, itemID int64) string {
	return fmt.Sprintf("/runs/%d/items/%d/attachment", runID, itemID)
}

func multipartUpload(t *testing.T, csrf, filename string, data []byte) (*bytes.Buffer, string) {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("csrf_token", csrf)
	part, _ := w.CreateFormFile("attachment", filename)
	_, _ = io.Copy(part, bytes.NewReader(data))
	_ = w.Close()
	return &body, w.FormDataContentType()
}
