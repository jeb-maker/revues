package store_test

import (
	"errors"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	runs "github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/store"
)

func TestAssignRunItem(t *testing.T) {
	ctx, st, run, itemID := setupInProgressRun(t)

	contrib, err := st.UpsertGitHubUser(ctx, 2, "contrib", "contrib@example.com", "Contrib", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(contrib): %v", err)
	}
	if err = st.AddSubjectMember(ctx, run.SubjectID, contrib.ID, store.SubjectRoleContributor); err != nil {
		t.Fatalf("AddSubjectMember(): %v", err)
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
			if err = st.AddProjectMember(ctx, run.SubjectID, contrib.ID, subjects.LocalRoleContributor); err != nil {
				t.Fatalf("AddProjectMember(): %v", err)
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
