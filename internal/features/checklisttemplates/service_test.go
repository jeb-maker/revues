package checklisttemplates_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/features/checklisttemplates"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
)

func newTemplateService(t *testing.T) (context.Context, *checklisttemplates.Service, *store.Store, *store.User) {
	t.Helper()
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	user, err := st.UpsertGitHubUser(ctx, 42, "editor", "editor@example.com", "Editor", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser: %v", err)
	}
	return ctx, &checklisttemplates.Service{Store: st}, st, user
}

func openMemoryDB(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/tpl.db", 0)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err = store.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return db
}

func TestService_CreateAndSavePublishesNewImmutableVersions(t *testing.T) {
	ctx, svc, st, user := newTemplateService(t)

	created, err := svc.Create(ctx, user, checklisttemplates.CreateInput{
		Name:    "Modèle A",
		Domains: []string{"infra"},
		Items: []store.TemplateItemInput{
			{Section: "Général", Label: "Point A", Required: true},
			{Section: "Général", Label: "Point B"},
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Version.Version != 1 {
		t.Fatalf("version = %d, want 1", created.Version.Version)
	}
	v1ID := created.Version.ID
	v1Items, err := st.ListTemplateItems(ctx, v1ID)
	if err != nil {
		t.Fatalf("ListTemplateItems v1: %v", err)
	}
	if len(v1Items) != 2 || v1Items[0].Label != "Point A" {
		t.Fatalf("v1 items = %+v", v1Items)
	}

	saved, err := svc.Save(ctx, user, created.Template.ID, checklisttemplates.SaveInput{
		Name:    "Modèle A v2",
		Domains: []string{"infra", "ops"},
		Items: []store.TemplateItemInput{
			{Section: "Sécurité", Label: "Point C", Required: true},
		},
	})
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if saved.Version.Version != 2 {
		t.Fatalf("version = %d, want 2", saved.Version.Version)
	}
	if saved.Template.Name != "Modèle A v2" {
		t.Fatalf("name = %q", saved.Template.Name)
	}

	// Prior published version must be unchanged.
	stillV1, err := st.ListTemplateItems(ctx, v1ID)
	if err != nil {
		t.Fatalf("ListTemplateItems v1 after save: %v", err)
	}
	if len(stillV1) != 2 || stillV1[0].Label != "Point A" || stillV1[1].Label != "Point B" {
		t.Fatalf("published v1 mutated: %+v", stillV1)
	}

	gotV1, err := svc.GetVersion(ctx, created.Template.ID, 1)
	if err != nil {
		t.Fatalf("GetVersion(1): %v", err)
	}
	if gotV1.Version.ID != v1ID || len(gotV1.Items) != 2 {
		t.Fatalf("GetVersion(1) = %+v", gotV1)
	}

	versions, err := svc.ListVersions(ctx, created.Template.ID)
	if err != nil {
		t.Fatalf("ListVersions: %v", err)
	}
	if len(versions) != 2 || versions[0].Version != 2 || versions[1].Version != 1 {
		t.Fatalf("versions = %+v", versions)
	}
}

func TestService_SaveEditorialDoesNotBumpVersion(t *testing.T) {
	ctx, svc, st, user := newTemplateService(t)

	created, err := svc.Create(ctx, user, checklisttemplates.CreateInput{
		Name: "Édito",
		Items: []store.TemplateItemInput{
			{Section: "A", Label: "Point", HelpText: "typo icci", Required: true},
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	v1 := created.Version.Version
	v1ID := created.Version.ID

	saved, err := svc.Save(ctx, user, created.Template.ID, checklisttemplates.SaveInput{
		Name: "Édito renommé",
		Items: []store.TemplateItemInput{
			{Section: "A", Label: "Point", HelpText: "typo ici", Required: true},
		},
	})
	if err != nil {
		t.Fatalf("Save editorial: %v", err)
	}
	if saved.Version.Version != v1 || saved.Version.ID != v1ID {
		t.Fatalf("version bumped: got v%d id=%d, want v%d id=%d", saved.Version.Version, saved.Version.ID, v1, v1ID)
	}
	if saved.Template.Name != "Édito renommé" {
		t.Fatalf("name = %q", saved.Template.Name)
	}
	if len(saved.Items) != 1 || saved.Items[0].HelpText != "typo ici" {
		t.Fatalf("items = %+v", saved.Items)
	}

	versions, err := svc.ListVersions(ctx, created.Template.ID)
	if err != nil {
		t.Fatalf("ListVersions: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("versions = %+v, want 1", versions)
	}

	// Metadata-only (same items) also stays on v1.
	again, err := svc.Save(ctx, user, created.Template.ID, checklisttemplates.SaveInput{
		Name:    "Édito renommé",
		Domains: []string{"ops"},
		Items: []store.TemplateItemInput{
			{Section: "A", Label: "Point", HelpText: "typo ici", Required: true},
		},
	})
	if err != nil {
		t.Fatalf("Save metadata: %v", err)
	}
	if again.Version.Version != v1 {
		t.Fatalf("metadata save bumped version to %d", again.Version.Version)
	}
	tags, err := st.ListTemplateTags(ctx, created.Template.ID)
	if err != nil || len(tags) != 1 || tags[0] != "ops" {
		t.Fatalf("domains = %v (%v)", tags, err)
	}
}

func TestService_SaveStructuralStillBumpsAfterEditorial(t *testing.T) {
	ctx, svc, _, user := newTemplateService(t)

	created, err := svc.Create(ctx, user, checklisttemplates.CreateInput{
		Name:  "Mix",
		Items: []store.TemplateItemInput{{Label: "A", HelpText: "h"}},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = svc.Save(ctx, user, created.Template.ID, checklisttemplates.SaveInput{
		Name:  "Mix",
		Items: []store.TemplateItemInput{{Label: "A", HelpText: "h2"}},
	})
	if err != nil {
		t.Fatalf("editorial: %v", err)
	}

	saved, err := svc.Save(ctx, user, created.Template.ID, checklisttemplates.SaveInput{
		Name:  "Mix",
		Items: []store.TemplateItemInput{{Label: "A renommé", HelpText: "h2"}},
	})
	if err != nil {
		t.Fatalf("structural: %v", err)
	}
	if saved.Version.Version != 2 {
		t.Fatalf("version = %d, want 2", saved.Version.Version)
	}
}

func TestIsStructuralItemChange(t *testing.T) {
	base := []store.TemplateItem{
		{Section: "S", Label: "L", HelpText: "h", Required: true},
	}
	tests := []struct {
		name string
		next []store.TemplateItemInput
		want bool
	}{
		{name: "identical", next: []store.TemplateItemInput{{Section: "S", Label: "L", HelpText: "h", Required: true}}, want: false},
		{name: "help only", next: []store.TemplateItemInput{{Section: "S", Label: "L", HelpText: "h2", Required: true}}, want: false},
		{name: "label", next: []store.TemplateItemInput{{Section: "S", Label: "X", HelpText: "h", Required: true}}, want: true},
		{name: "section", next: []store.TemplateItemInput{{Section: "T", Label: "L", HelpText: "h", Required: true}}, want: true},
		{name: "required", next: []store.TemplateItemInput{{Section: "S", Label: "L", HelpText: "h", Required: false}}, want: true},
		{name: "count", next: []store.TemplateItemInput{{Section: "S", Label: "L", Required: true}, {Label: "Z"}}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checklisttemplates.IsStructuralItemChange(base, tt.next); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_PublishedVersionImmutable(t *testing.T) {
	ctx, svc, _, user := newTemplateService(t)

	created, err := svc.Create(ctx, user, checklisttemplates.CreateInput{
		Name: "Immut",
		Items: []store.TemplateItemInput{
			{Label: "A", Required: true},
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = svc.Store.ReplaceTemplateItems(ctx, created.Version.ID, []store.TemplateItemInput{{Label: "X"}})
	if !errors.Is(err, store.ErrPublishedVersionImmutable) {
		t.Fatalf("ReplaceTemplateItems err = %v", err)
	}
	if !errors.Is(err, checklisttemplates.ErrImmutable) {
		t.Fatalf("ReplaceTemplateItems err = %v, want ErrImmutable alias", err)
	}
}

func TestService_CreateVersionIncrementsWithoutTouchingName(t *testing.T) {
	ctx, svc, _, user := newTemplateService(t)

	created, err := svc.Create(ctx, user, checklisttemplates.CreateInput{
		Name: "Nom fixe",
		Items: []store.TemplateItemInput{
			{Label: "A"},
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	v2, err := svc.CreateVersion(ctx, user, created.Template.ID, []store.TemplateItemInput{
		{Label: "B"},
		{Label: "C"},
	})
	if err != nil {
		t.Fatalf("CreateVersion: %v", err)
	}
	if v2.Version.Version != 2 {
		t.Fatalf("version = %d, want 2", v2.Version.Version)
	}
	if v2.Template.Name != "Nom fixe" {
		t.Fatalf("name changed to %q", v2.Template.Name)
	}
	if len(v2.Items) != 2 {
		t.Fatalf("items = %+v", v2.Items)
	}
}

func TestService_ValidationRejectsEmptyItems(t *testing.T) {
	ctx, svc, _, user := newTemplateService(t)

	_, err := svc.Create(ctx, user, checklisttemplates.CreateInput{
		Name:  "Empty",
		Items: nil,
	})
	if !errors.Is(err, checklisttemplates.ErrValidation) {
		t.Fatalf("err = %v, want validation", err)
	}
}

func TestService_ReaderCannotManage(t *testing.T) {
	ctx, svc, st, _ := newTemplateService(t)
	reader, err := st.UpsertGitHubUser(ctx, 99, "reader", "reader@example.com", "Reader", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser: %v", err)
	}

	_, err = svc.Create(ctx, reader, checklisttemplates.CreateInput{
		Name:  "Nope",
		Items: []store.TemplateItemInput{{Label: "A"}},
	})
	if !errors.Is(err, checklisttemplates.ErrForbidden) {
		t.Fatalf("err = %v, want forbidden", err)
	}
}
