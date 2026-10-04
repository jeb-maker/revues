package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	runs "github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
)

func TestAssignRunItem(t *testing.T) {
	ctx, st, run, itemID := setupInProgressRun(t)

	contrib, err := st.UpsertGitHubUser(ctx, 2, "contrib", "contrib@example.com", "Contrib", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(contrib): %v", err)
	}
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, contrib.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}
	if err = st.UpsertDirectSubjectMember(ctx, run.SubjectID, contrib.ID, store.SubjectRoleContributor); err != nil {
		t.Fatalf("UpsertDirectSubjectMember(): %v", err)
	}

	if err = st.AssignRunItem(ctx, run.ID, itemID, &contrib.ID); err != nil {
		t.Fatalf("AssignRunItem(): %v", err)
	}

	item, err := st.RunItemByID(ctx, run.ID, itemID)
	if err != nil {
		t.Fatalf("RunItemByID(): %v", err)
	}
	if !item.AssignedTo.Valid || item.AssignedTo.Int64 != contrib.ID {
		t.Fatalf("assigned_to = %+v, want %d", item.AssignedTo, contrib.ID)
	}

	tasks, err := st.ListAssignedRunItems(ctx, contrib.ID, "", "")
	if err != nil || len(tasks) != 1 {
		t.Fatalf("ListAssignedRunItems() = %v, %v", tasks, err)
	}

	filtered, err := st.ListAssignedRunItems(ctx, contrib.ID, runs.StatusPending, "")
	if err != nil || len(filtered) != 1 {
		t.Fatalf("ListAssignedRunItems(filter) = %v, %v", filtered, err)
	}

	byLabel, err := st.ListAssignedRunItems(ctx, contrib.ID, "", "Point")
	if err != nil || len(byLabel) != 1 {
		t.Fatalf("ListAssignedRunItems(search) = %v, %v", byLabel, err)
	}

	empty, err := st.ListAssignedRunItems(ctx, contrib.ID, runs.StatusOK, "")
	if err != nil || len(empty) != 0 {
		t.Fatalf("ListAssignedRunItems(ok filter) = %v, %v", empty, err)
	}

	if err = st.AssignRunItem(ctx, run.ID, itemID, nil); err != nil {
		t.Fatalf("AssignRunItem(clear): %v", err)
	}
	item, err = st.RunItemByID(ctx, run.ID, itemID)
	if err != nil {
		t.Fatalf("RunItemByID(): %v", err)
	}
	if item.AssignedTo.Valid {
		t.Fatal("expected assignee cleared")
	}
}

func TestAssignRunItemChecked_OptimisticLock(t *testing.T) {
	const stale = "2000-01-01T00:00:00Z"
	tests := []struct {
		name         string
		expected     func(current string) string
		wantErr      error
		wantAssigned bool
	}{
		{"matching updated_at applies", func(c string) string { return c }, nil, true},
		{"stale updated_at conflicts", func(string) string { return stale }, store.ErrRunItemConflict, false},
		{"empty updated_at skips the check", func(string) string { return "" }, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, st, run, itemID := setupInProgressRun(t)
			contrib, err := st.UpsertGitHubUser(ctx, 2, "contrib", "contrib@example.com", "Contrib", "", auth.RoleEditor)
			if err != nil {
				t.Fatalf("UpsertGitHubUser(contrib): %v", err)
			}
			defaultOrg, orgErr := st.OrganizationBySlug(ctx, "default")
			if orgErr != nil {
				t.Fatalf("OrganizationBySlug(): %v", orgErr)
			}
			if err = st.AddOrganizationMember(ctx, defaultOrg.ID, contrib.ID, store.OrgRoleMember); err != nil {
				t.Fatalf("AddOrganizationMember(): %v", err)
			}
			if err = st.UpsertDirectSubjectMember(ctx, run.SubjectID, contrib.ID, store.SubjectRoleContributor); err != nil {
				t.Fatalf("UpsertDirectSubjectMember(): %v", err)
			}
			before, err := st.RunItemByID(ctx, run.ID, itemID)
			if err != nil {
				t.Fatalf("RunItemByID(): %v", err)
			}

			err = st.AssignRunItemChecked(ctx, run.ID, itemID, &contrib.ID, tt.expected(before.UpdatedAt))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("AssignRunItemChecked() error = %v, want %v", err, tt.wantErr)
			}

			after, err := st.RunItemByID(ctx, run.ID, itemID)
			if err != nil {
				t.Fatalf("RunItemByID(after): %v", err)
			}
			if after.AssignedTo.Valid != tt.wantAssigned {
				t.Fatalf("assigned_to = %+v, want assigned=%v", after.AssignedTo, tt.wantAssigned)
			}
		})
	}
}

func TestAssignRunItemRejectsNonMember(t *testing.T) {
	ctx, st, run, itemID := setupInProgressRun(t)

	outsider, err := st.UpsertGitHubUser(ctx, 99, "outsider", "out@example.com", "Out", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	err = st.AssignRunItem(ctx, run.ID, itemID, &outsider.ID)
	if !errors.Is(err, store.ErrInvalidAssignee) {
		t.Fatalf("AssignRunItem() error = %v, want ErrInvalidAssignee", err)
	}
}

func TestAssignRunItemRejectsPrivateSubjectWithoutGrant(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 1, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(lead): %v", err)
	}
	member, err := st.UpsertGitHubUser(ctx, 2, "member", "member@example.com", "Member", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(member): %v", err)
	}
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, member.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}

	subject, err := st.CreateSubjectWithVisibility(ctx, "Private", "", lead.ID, nil, store.SubjectVisibilityPrivate)
	if err != nil {
		t.Fatalf("CreateSubjectWithVisibility(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, []store.TemplateItemInput{
		{Section: "S", Label: "Point", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, subject.ID, template.ID, lead.ID)
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

	if err = st.AssignRunItem(ctx, run.ID, items[0].ID, &member.ID); !errors.Is(err, store.ErrInvalidAssignee) {
		t.Fatalf("AssignRunItem(no grant) = %v, want ErrInvalidAssignee", err)
	}

	if err = st.UpsertDirectSubjectMember(ctx, subject.ID, member.ID, store.SubjectRoleContributor); err != nil {
		t.Fatalf("UpsertDirectSubjectMember(): %v", err)
	}
	if err = st.AssignRunItem(ctx, run.ID, items[0].ID, &member.ID); err != nil {
		t.Fatalf("AssignRunItem(with grant): %v", err)
	}
}

func TestListAssignedRunItemsHidesPrivateWithoutAccess(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 1, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(lead): %v", err)
	}
	member, err := st.UpsertGitHubUser(ctx, 2, "member", "member@example.com", "Member", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(member): %v", err)
	}
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, member.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}

	subject, err := st.CreateSubjectWithVisibility(ctx, "Private", "", lead.ID, nil, store.SubjectVisibilityPrivate)
	if err != nil {
		t.Fatalf("CreateSubjectWithVisibility(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, []store.TemplateItemInput{
		{Section: "S", Label: "Secret point", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, subject.ID, template.ID, lead.ID)
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

	// Bypass AssignRunItem to simulate a legacy/wrong assignment on a private subject.
	if _, err = db.ExecContext(ctx, `UPDATE run_items SET assigned_to = ? WHERE id = ?`, member.ID, items[0].ID); err != nil {
		t.Fatalf("force assign: %v", err)
	}

	tasks, err := st.ListAssignedRunItems(ctx, member.ID, "", "")
	if err != nil {
		t.Fatalf("ListAssignedRunItems(): %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("ListAssignedRunItems() leaked %d private tasks", len(tasks))
	}

	if err = st.UpsertDirectSubjectMember(ctx, subject.ID, member.ID, store.SubjectRoleContributor); err != nil {
		t.Fatalf("UpsertDirectSubjectMember(): %v", err)
	}
	tasks, err = st.ListAssignedRunItems(ctx, member.ID, "", "")
	if err != nil || len(tasks) != 1 {
		t.Fatalf("ListAssignedRunItems(after grant) = %v, %v", tasks, err)
	}
}
