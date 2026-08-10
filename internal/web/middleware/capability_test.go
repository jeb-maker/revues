package middleware

import (
	"context"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/crypto"
	"github.com/jeb-maker/revues/internal/features/admin/settings"
	"github.com/jeb-maker/revues/internal/integrations/jira"
	"github.com/jeb-maker/revues/internal/integrations/notion"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
)

func TestResolveUICaps_P3CapabilitiesIndependentOfSimpleUI(t *testing.T) {
	ctx := context.Background()
	db := openOrgAdminTestDB(t)
	st := store.New(db)
	key := make([]byte, crypto.KeySize)

	user, err := st.UpsertGitHubUser(ctx, 21001, "cap-solo", "cap-solo@example.com", "Cap", "", auth.RoleEditor)
	if err != nil {
		t.Fatal(err)
	}
	org, err := st.CreateOrganization(ctx, "Cap Org", "cap-org-ui", user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = st.AddOrganizationMember(ctx, org.ID, user.ID, store.OrgRoleOwner); err != nil {
		t.Fatal(err)
	}
	octx := orgctx.WithOrganizationID(ctx, org.ID)
	if err = st.InsertAllowedEmail(octx, user.Email, auth.RoleEditor); err != nil {
		t.Fatal(err)
	}
	if _, err = st.CreateSubjectWithVisibility(octx, "Seul", "", user.ID, nil, store.SubjectVisibilityPrivate); err != nil {
		t.Fatal(err)
	}

	reqCtx := testOrgNavCtx(t, st, org, user)
	hd := HeaderData{
		UserOrganizations: []store.OrganizationMembership{{Organization: *org, Role: store.OrgRoleOwner}},
	}

	before := resolveUICaps(reqCtx, st, user, hd, key)
	if !before.SimpleUI {
		t.Fatal("expected SimpleUI before integrations")
	}
	if before.HasJira || before.HasNotion || before.HasWebhooks {
		t.Fatalf("P3 unset = %+v", before)
	}

	if err = (&jira.Service{Store: st, EncryptionKey: key}).Save(octx, jira.Config{
		InstanceType: jira.InstanceCloud,
		BaseURL:      "https://example.atlassian.net",
		Email:        "user@example.com",
		APIToken:     "token",
	}); err != nil {
		t.Fatalf("Save jira: %v", err)
	}
	if err = (&notion.Service{Store: st, EncryptionKey: key}).Save(octx, notion.Config{APIToken: "notion-token"}); err != nil {
		t.Fatalf("Save notion: %v", err)
	}
	if err = (&settings.SettingsService{Store: st, EncryptionKey: key}).SaveWebhooks(octx, settings.WebhookConfig{
		URLs:            []string{"https://hooks.example.com/revues"},
		Secret:          "secret",
		ReviewCompleted: true,
	}); err != nil {
		t.Fatalf("Save webhooks: %v", err)
	}

	after := resolveUICaps(reqCtx, st, user, hd, key)
	if !after.SimpleUI {
		t.Fatal("P3 must not disable SimpleUI")
	}
	if !after.HasJira || !after.HasNotion || !after.HasWebhooks {
		t.Fatalf("P3 configured caps = %+v", after)
	}
}

func TestResolveCapabilityCaps_RequiresOrgAndKey(t *testing.T) {
	ctx := context.Background()
	db := openOrgAdminTestDB(t)
	st := store.New(db)
	key := make([]byte, crypto.KeySize)

	j, n, w := resolveCapabilityCaps(ctx, st, key)
	if j || n || w {
		t.Fatalf("no org context → false, got %v %v %v", j, n, w)
	}
	j, n, w = resolveCapabilityCaps(ctx, st, nil)
	if j || n || w {
		t.Fatalf("nil key → false, got %v %v %v", j, n, w)
	}
}
