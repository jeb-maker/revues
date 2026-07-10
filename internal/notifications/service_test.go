package notifications_test

import (
	"context"
	"database/sql"
	"github.com/jeb-maker/revues/internal/testutil"
	"sync"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/crypto"
	adminsettings "github.com/jeb-maker/revues/internal/features/admin/settings"
	"github.com/jeb-maker/revues/internal/features/projects"
	"github.com/jeb-maker/revues/internal/notifications"
	"github.com/jeb-maker/revues/internal/store"
)

func TestServiceSkipsWhenSMTPNotConfigured(t *testing.T) {
	st, settingsSvc, ctx := testNotificationDeps(t)

	lead, err := st.UpsertGitHubUser(ctx, 1, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "P", "", lead.ID)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, project.ID, "Modèle", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, "Revue", lead.ID, sql.NullString{})
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}

	svc := &notifications.Service{
		Store:    st,
		Settings: settingsSvc,
		BaseURL:  "http://example.com",
	}
	svc.NotifyRunCompleted(ctx, run.ID)
	time.Sleep(50 * time.Millisecond)
}

func TestServiceNotifyRunCompleted(t *testing.T) {
	st, settingsSvc, ctx := testNotificationDeps(t)
	host, port := startCapturingSMTPServer(t)
	saveTestSMTP(t, ctx, settingsSvc, host, port)

	lead, err := st.UpsertGitHubUser(ctx, 1, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(lead): %v", err)
	}
	member, err := st.UpsertGitHubUser(ctx, 2, "member", "member@example.com", "Member", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(member): %v", err)
	}
	project, err := st.CreateProject(ctx, "P", "", lead.ID)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	if err = st.AddProjectMember(ctx, project.ID, member.ID, projects.LocalRoleContributor); err != nil {
		t.Fatalf("AddProjectMember(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, project.ID, "Modèle", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, "Revue Q1", lead.ID, sql.NullString{})
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}

	svc := &notifications.Service{
		Store:    st,
		Settings: settingsSvc,
		BaseURL:  "http://example.com",
	}
	svc.NotifyRunCompleted(ctx, run.ID)

	waitForSMTPMessages(t, 2)
}

func TestServiceNotifyItemAssigned(t *testing.T) {
	st, settingsSvc, ctx := testNotificationDeps(t)
	host, port := startCapturingSMTPServer(t)
	saveTestSMTP(t, ctx, settingsSvc, host, port)

	lead, err := st.UpsertGitHubUser(ctx, 1, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	assignee, err := st.UpsertGitHubUser(ctx, 2, "assignee", "assignee@example.com", "Assignee", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(assignee): %v", err)
	}
	project, err := st.CreateProject(ctx, "P", "", lead.ID)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	if err = st.AddProjectMember(ctx, project.ID, assignee.ID, projects.LocalRoleContributor); err != nil {
		t.Fatalf("AddProjectMember(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, project.ID, "Modèle", lead.ID, []store.TemplateItemInput{
		{Label: "Point A", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, "Revue", lead.ID, sql.NullString{})
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	if err = st.StartRun(ctx, run.ID); err != nil {
		t.Fatalf("StartRun(): %v", err)
	}
	items, err := st.ListRunItems(ctx, run.ID)
	if err != nil || len(items) == 0 {
		t.Fatalf("ListRunItems() = %v, %v", items, err)
	}
	if err = st.AssignRunItem(ctx, run.ID, items[0].ID, &assignee.ID); err != nil {
		t.Fatalf("AssignRunItem(): %v", err)
	}

	svc := &notifications.Service{
		Store:    st,
		Settings: settingsSvc,
		BaseURL:  "http://example.com",
	}
	svc.NotifyItemAssigned(ctx, run.ID, items[0].ID)

	waitForSMTPMessages(t, 1)
}

func TestServiceSendDueReminders(t *testing.T) {
	st, settingsSvc, ctx := testNotificationDeps(t)
	host, port := startCapturingSMTPServer(t)
	saveTestSMTP(t, ctx, settingsSvc, host, port)

	lead, err := st.UpsertGitHubUser(ctx, 1, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "P", "", lead.ID)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, project.ID, "Modèle", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, "Revue due", lead.ID, sql.NullString{})
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	tomorrow := time.Now().UTC().Add(24 * time.Hour).Format("2006-01-02")
	if err = st.SetRunDueDate(ctx, run.ID, sql.NullString{String: tomorrow + "T00:00:00Z", Valid: true}); err != nil {
		t.Fatalf("SetRunDueDate(): %v", err)
	}
	if err = st.StartRun(ctx, run.ID); err != nil {
		t.Fatalf("StartRun(): %v", err)
	}

	svc := &notifications.Service{
		Store:    st,
		Settings: settingsSvc,
		BaseURL:  "http://example.com",
	}
	if err = svc.SendDueReminders(ctx); err != nil {
		t.Fatalf("SendDueReminders(): %v", err)
	}

	waitForSMTPMessages(t, 1)
}

func testNotificationDeps(t *testing.T) (*store.Store, *adminsettings.SettingsService, context.Context) {
	t.Helper()

	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("Close(): %v", err)
		}
	})
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("foreign_keys: %v", err)
	}
	if err := store.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate(): %v", err)
	}

	key := make([]byte, crypto.KeySize)
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	return st, &adminsettings.SettingsService{Store: st, EncryptionKey: key}, ctx
}

func saveTestSMTP(t *testing.T, ctx context.Context, settingsSvc *adminsettings.SettingsService, host string, port int) {
	t.Helper()
	if err := settingsSvc.SaveSMTP(ctx, adminsettings.SMTPConfig{
		Host: host,
		Port: port,
		From: "revues@example.com",
	}); err != nil {
		t.Fatalf("SaveSMTP(): %v", err)
	}
}

func startCapturingSMTPServer(t *testing.T) (host string, port int) {
	t.Helper()

	host, port = startTestSMTPServer(t)
	resetSMTPCapture()
	return host, port
}

func waitForSMTPMessages(t *testing.T, count int) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if smtpMessageCount() >= count {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("expected %d smtp messages, got %d", count, smtpMessageCount())
}

var (
	smtpCaptureMu sync.Mutex
	smtpCount     int
)

func resetSMTPCapture() {
	smtpCaptureMu.Lock()
	defer smtpCaptureMu.Unlock()
	smtpCount = 0
}

func smtpMessageCount() int {
	smtpCaptureMu.Lock()
	defer smtpCaptureMu.Unlock()
	return smtpCount
}

func notifySMTPCapture() {
	smtpCaptureMu.Lock()
	smtpCount++
	smtpCaptureMu.Unlock()
}
