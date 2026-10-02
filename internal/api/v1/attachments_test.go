package apiv1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
)

func TestAttachments_UploadDownloadSecurity(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfg := config.Config{
		SessionSecret:  "test-secret-at-least-thirty-two-bytes",
		EncryptionKey:  config.TestEncryptionKey(),
		Env:            "development",
		AttachmentsDir: dir,
	}
	handler, st := newTestRouterWithStore(t, cfg)
	editor, session, csrf := seedSessionUser(t, st, cfg, "att-editor@example.com", "Att Editor", auth.RoleEditor, true)
	_ = editor

	runID, itemID := seedRunForAttachments(t, handler, session, csrf)

	// Reject unsupported magic bytes
	bad := doMultipart(t, handler, fmt.Sprintf("/api/v1/runs/%d/items/%d/attachments", runID, itemID),
		"file", "evil.exe", []byte{0x4D, 0x5A, 0x90, 0x00}, session, csrf)
	if bad.Code != http.StatusBadRequest || !strings.Contains(bad.Body.String(), "non autorisé") {
		t.Fatalf("unsupported status=%d body=%s", bad.Code, bad.Body.String())
	}

	// Reject oversize
	big := make([]byte, 5*1024*1024+1)
	big[0], big[1], big[2] = 0xFF, 0xD8, 0xFF
	over := doMultipart(t, handler, fmt.Sprintf("/api/v1/runs/%d/items/%d/attachments", runID, itemID),
		"file", "big.jpg", big, session, csrf)
	if over.Code != http.StatusBadRequest || !strings.Contains(over.Body.String(), "5 Mo") {
		t.Fatalf("oversize status=%d body=%s", over.Code, over.Body.String())
	}

	// CSRF required
	var imgBuf bytes.Buffer
	_ = jpeg.Encode(&imgBuf, image.NewRGBA(image.Rect(0, 0, 32, 32)), nil)
	noCSRF := doMultipart(t, handler, fmt.Sprintf("/api/v1/runs/%d/items/%d/attachments", runID, itemID),
		"file", "proof.jpg", imgBuf.Bytes(), session, "")
	if noCSRF.Code != http.StatusForbidden {
		t.Fatalf("no csrf status=%d", noCSRF.Code)
	}

	// Success
	ok := doMultipart(t, handler, fmt.Sprintf("/api/v1/runs/%d/items/%d/attachments", runID, itemID),
		"file", "proof.jpg", imgBuf.Bytes(), session, csrf)
	if ok.Code != http.StatusCreated {
		t.Fatalf("upload status=%d body=%s", ok.Code, ok.Body.String())
	}
	var att map[string]any
	if err := json.Unmarshal(ok.Body.Bytes(), &att); err != nil {
		t.Fatalf("json: %v", err)
	}
	attID := int64(att["id"].(float64))
	if att["filename"] != "proof.jpg" || att["is_image"] != true {
		t.Fatalf("att = %#v", att)
	}
	if _, err := os.Stat(filepath.Join(dir, mustStoragePath(t, st, itemID))); err != nil {
		t.Fatalf("stored file missing: %v", err)
	}

	// Included in item detail
	detail := doJSON(t, handler, http.MethodGet, fmt.Sprintf("/api/v1/runs/%d/items/%d", runID, itemID), nil, session, "")
	if detail.Code != http.StatusOK || !strings.Contains(detail.Body.String(), `"proof.jpg"`) {
		t.Fatalf("detail status=%d body=%s", detail.Code, detail.Body.String())
	}

	// Download with Content-Disposition attachment
	dl := doJSON(t, handler, http.MethodGet,
		fmt.Sprintf("/api/v1/runs/%d/items/%d/attachments/%d", runID, itemID, attID), nil, session, "")
	if dl.Code != http.StatusOK {
		t.Fatalf("download status=%d", dl.Code)
	}
	cd := dl.Header().Get("Content-Disposition")
	if !strings.HasPrefix(cd, "attachment;") || !strings.Contains(cd, "proof.jpg") {
		t.Fatalf("Content-Disposition = %q", cd)
	}
	if !strings.HasPrefix(dl.Header().Get("Content-Type"), "image/") {
		t.Fatalf("Content-Type = %q", dl.Header().Get("Content-Type"))
	}

	// IDOR: wrong item id for same attachment
	tplRec := doJSON(t, handler, http.MethodPost, "/api/v1/templates", map[string]any{
		"name": "Other", "domains": []string{"ops"},
		"items": []map[string]any{{"section": "A", "label": "X", "required": true}},
	}, session, csrf)
	var tpl map[string]any
	_ = json.Unmarshal(tplRec.Body.Bytes(), &tpl)
	subRec := doJSON(t, handler, http.MethodPost, "/api/v1/subjects", map[string]any{
		"name": "Other Sub", "domains": []string{"ops"},
	}, session, csrf)
	var sub map[string]any
	_ = json.Unmarshal(subRec.Body.Bytes(), &sub)
	launch := doJSON(t, handler, http.MethodPost,
		fmt.Sprintf("/api/v1/subjects/%d/runs", int64(sub["id"].(float64))),
		map[string]any{"template_id": int64(tpl["id"].(float64))}, session, csrf)
	var run2 map[string]any
	_ = json.Unmarshal(launch.Body.Bytes(), &run2)
	item2 := run2["items"].([]any)[0].(map[string]any)
	wrong := doJSON(t, handler, http.MethodGet,
		fmt.Sprintf("/api/v1/runs/%d/items/%d/attachments/%d",
			int64(run2["id"].(float64)), int64(item2["id"].(float64)), attID),
		nil, session, "")
	if wrong.Code != http.StatusNotFound {
		t.Fatalf("cross-item download status=%d want 404", wrong.Code)
	}
}

func TestAttachments_IDOR_CrossUser(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		SessionSecret:  "test-secret-at-least-thirty-two-bytes",
		EncryptionKey:  config.TestEncryptionKey(),
		Env:            "development",
		AttachmentsDir: t.TempDir(),
	}
	handler, st := newTestRouterWithStore(t, cfg)
	alice, aliceSession, aliceCSRF := seedSessionUser(t, st, cfg, "att-alice@example.com", "Alice", auth.RoleEditor, true)
	_, bobSession, bobCSRF := seedSessionUser(t, st, cfg, "att-bob@example.com", "Bob", auth.RoleEditor, true)

	runID, itemID := seedRunForAttachments(t, handler, aliceSession, aliceCSRF)

	// Gate subject: direct membership ends org_member_legacy for others.
	detail := doJSON(t, handler, http.MethodGet, fmt.Sprintf("/api/v1/runs/%d", runID), nil, aliceSession, "")
	if detail.Code != http.StatusOK {
		t.Fatalf("run detail: %d %s", detail.Code, detail.Body.String())
	}
	var run map[string]any
	_ = json.Unmarshal(detail.Body.Bytes(), &run)
	subjectID := int64(run["subject_id"].(float64))
	addRec := doJSON(t, handler, http.MethodPost,
		fmt.Sprintf("/api/v1/subjects/%d/members", subjectID),
		map[string]any{"email": alice.Email, "role": "lead"},
		aliceSession, aliceCSRF)
	if addRec.Code != http.StatusCreated && addRec.Code != http.StatusOK && addRec.Code != http.StatusConflict {
		// Creator may already be lead; accept conflict/created.
		if addRec.Code != http.StatusBadRequest {
			t.Fatalf("add member: %d %s", addRec.Code, addRec.Body.String())
		}
	}
	var imgBuf bytes.Buffer
	_ = jpeg.Encode(&imgBuf, image.NewRGBA(image.Rect(0, 0, 8, 8)), nil)
	up := doMultipart(t, handler, fmt.Sprintf("/api/v1/runs/%d/items/%d/attachments", runID, itemID),
		"file", "a.jpg", imgBuf.Bytes(), aliceSession, aliceCSRF)
	if up.Code != http.StatusCreated {
		t.Fatalf("upload status=%d body=%s", up.Code, up.Body.String())
	}
	var att map[string]any
	_ = json.Unmarshal(up.Body.Bytes(), &att)
	attID := int64(att["id"].(float64))

	deniedGet := doJSON(t, handler, http.MethodGet,
		fmt.Sprintf("/api/v1/runs/%d/items/%d/attachments/%d", runID, itemID, attID),
		nil, bobSession, "")
	if deniedGet.Code != http.StatusNotFound {
		t.Fatalf("bob download status=%d want 404", deniedGet.Code)
	}

	deniedUp := doMultipart(t, handler, fmt.Sprintf("/api/v1/runs/%d/items/%d/attachments", runID, itemID),
		"file", "b.jpg", imgBuf.Bytes(), bobSession, bobCSRF)
	if deniedUp.Code != http.StatusNotFound {
		t.Fatalf("bob upload status=%d want 404", deniedUp.Code)
	}
}

func seedRunForAttachments(t *testing.T, handler http.Handler, session *http.Cookie, csrf string) (runID, itemID int64) {
	t.Helper()
	tplRec := doJSON(t, handler, http.MethodPost, "/api/v1/templates", map[string]any{
		"name": "AttTpl", "domains": []string{"ops"},
		"items": []map[string]any{{"section": "S", "label": "Point", "required": true}},
	}, session, csrf)
	if tplRec.Code != http.StatusCreated {
		t.Fatalf("template status=%d body=%s", tplRec.Code, tplRec.Body.String())
	}
	var tpl map[string]any
	_ = json.Unmarshal(tplRec.Body.Bytes(), &tpl)

	subRec := doJSON(t, handler, http.MethodPost, "/api/v1/subjects", map[string]any{
		"name": "AttSub", "domains": []string{"ops"},
	}, session, csrf)
	if subRec.Code != http.StatusCreated {
		t.Fatalf("subject status=%d body=%s", subRec.Code, subRec.Body.String())
	}
	var sub map[string]any
	_ = json.Unmarshal(subRec.Body.Bytes(), &sub)

	launch := doJSON(t, handler, http.MethodPost,
		fmt.Sprintf("/api/v1/subjects/%d/runs", int64(sub["id"].(float64))),
		map[string]any{"template_id": int64(tpl["id"].(float64))}, session, csrf)
	if launch.Code != http.StatusCreated {
		t.Fatalf("launch status=%d body=%s", launch.Code, launch.Body.String())
	}
	var run map[string]any
	_ = json.Unmarshal(launch.Body.Bytes(), &run)
	item := run["items"].([]any)[0].(map[string]any)
	return int64(run["id"].(float64)), int64(item["id"].(float64))
}

func doMultipart(t *testing.T, handler http.Handler, path, field, filename string, data []byte, session *http.Cookie, csrf string) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile(field, filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err = io.Copy(part, bytes.NewReader(data)); err != nil {
		t.Fatalf("copy: %v", err)
	}
	if err = w.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
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

func mustStoragePath(t *testing.T, st *store.Store, itemID int64) string {
	t.Helper()
	att, err := st.AttachmentByRunItemID(context.Background(), itemID)
	if err != nil {
		t.Fatalf("AttachmentByRunItemID: %v", err)
	}
	return att.StoragePath
}
