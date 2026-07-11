package handlers_test

import (
	"bufio"
	"context"
	"database/sql"
	"github.com/jeb-maker/revues/internal/testutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/crypto"
	adminsettings "github.com/jeb-maker/revues/internal/features/admin/settings"
	"github.com/jeb-maker/revues/internal/store"
	appweb "github.com/jeb-maker/revues/internal/web"
)

func TestAdminSMTP_ReaderForbidden(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	reader, err := st.UpsertGitHubUser(ctx, 1, "reader", "reader@example.com", "Reader", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, reader.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/settings/smtp", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestAdminSMTP_SaveAndTest(t *testing.T) {
	host, port := startTestSMTPServer(t)

	handler, db := testRouterWithEncryptionKey(t, config.TestEncryptionKey())
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	adminUser, err := st.UpsertGitHubUser(ctx, 99, "admin", "admin@example.com", "Admin", "", auth.RoleAdmin)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	if insertErr := st.InsertAllowedEmail(ctx, "admin@example.com", auth.RoleAdmin); insertErr != nil {
		t.Fatalf("InsertAllowedEmail(): %v", insertErr)
	}

	secret := "test-secret-at-least-thirty-two-bytes"
	sessions := &auth.SessionManager{Store: st, SessionSecret: secret}
	token, _, err := sessions.CreateLoginSession(ctx, adminUser.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}
	csrf := auth.CSRFToken(token, secret)

	saveForm := url.Values{}
	saveForm.Set("csrf_token", csrf)
	saveForm.Set("action", "save")
	saveForm.Set("host", host)
	saveForm.Set("port", strconv.Itoa(port))
	saveForm.Set("from", "revues@example.com")
	saveForm.Set("password", "secret")
	saveReq := httptest.NewRequest(http.MethodPost, "/admin/settings/smtp", strings.NewReader(saveForm.Encode()))
	saveReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	saveReq.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	saveRec := httptest.NewRecorder()
	handler.ServeHTTP(saveRec, saveReq)
	if saveRec.Code != http.StatusSeeOther {
		t.Fatalf("save status = %d, want %d, body=%q", saveRec.Code, http.StatusSeeOther, saveRec.Body.String())
	}

	key, err := crypto.DecodeKey(config.TestEncryptionKey())
	if err != nil {
		t.Fatalf("DecodeKey(): %v", err)
	}
	svc := &adminsettings.SettingsService{Store: st, EncryptionKey: key}
	cfg, ok, err := svc.LoadSMTP(ctx)
	if err != nil || !ok {
		t.Fatalf("LoadSMTP() = ok=%v err=%v", ok, err)
	}
	if cfg.Host != host || cfg.Port != port || cfg.From != "revues@example.com" {
		t.Fatalf("LoadSMTP() = %+v", cfg)
	}

	testForm := url.Values{}
	testForm.Set("csrf_token", csrf)
	testForm.Set("action", "test")
	testForm.Set("test_recipient", "admin@example.com")
	testReq := httptest.NewRequest(http.MethodPost, "/admin/settings/smtp", strings.NewReader(testForm.Encode()))
	testReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	testReq.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	testRec := httptest.NewRecorder()
	handler.ServeHTTP(testRec, testReq)
	if testRec.Code != http.StatusSeeOther {
		t.Fatalf("test status = %d, want %d, body=%q", testRec.Code, http.StatusSeeOther, testRec.Body.String())
	}
}

func testRouterWithEncryptionKey(t *testing.T, encryptionKey string) (http.Handler, *sql.DB) {
	t.Helper()

	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/test.db")
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
		Addr:          ":8080",
		BaseURL:       "http://example.com",
		SessionSecret: "test-secret-at-least-thirty-two-bytes",
		EncryptionKey: encryptionKey,
		Env:           "development",
	}

	handler, _, err := appweb.NewRouter(appweb.Deps{Config: cfg, DB: db})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	return handler, db
}

func startTestSMTPServer(t *testing.T) (host string, port int) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen(): %v", err)
	}
	t.Cleanup(func() {
		_ = ln.Close()
	})

	go func() {
		for {
			conn, acceptErr := ln.Accept()
			if acceptErr != nil {
				return
			}
			go handleTestSMTP(conn)
		}
	}()

	host, portStr, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatalf("SplitHostPort(): %v", err)
	}
	port, err = strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("Atoi(): %v", err)
	}

	return host, port
}

func handleTestSMTP(conn net.Conn) {
	defer conn.Close()

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	writeSMTPLine(rw, "220 localhost ESMTP")

	for {
		line, err := rw.ReadString('\n')
		if err != nil {
			return
		}
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) == 0 {
			continue
		}
		cmd := strings.ToUpper(fields[0])
		switch cmd {
		case "EHLO", "HELO":
			writeSMTPLine(rw, "250-localhost")
			writeSMTPLine(rw, "250 OK")
		case "MAIL", "RCPT":
			writeSMTPLine(rw, "250 OK")
		case "DATA":
			writeSMTPLine(rw, "354 End data with <CR><LF>.<CR><LF>")
			for {
				part, err := rw.ReadString('\n')
				if err != nil {
					return
				}
				if strings.TrimSpace(part) == "." {
					break
				}
			}
			writeSMTPLine(rw, "250 OK")
		case "QUIT":
			writeSMTPLine(rw, "221 Bye")
			return
		default:
			writeSMTPLine(rw, "250 OK")
		}
	}
}

func writeSMTPLine(rw *bufio.ReadWriter, line string) {
	_, _ = rw.WriteString(line + "\r\n")
	_ = rw.Flush()
}
