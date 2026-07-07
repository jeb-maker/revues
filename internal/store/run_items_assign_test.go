package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/features/projects"
	"github.com/jeb-maker/revues/internal/items"
	"github.com/jeb-maker/revues/internal/store"
)

func TestAssignRunItem(t *testing.T) {
	ctx := context.Background()
	st, run, itemID := setupInProgressRun(t)

	contrib, err := st.UpsertGitHubUser(ctx, 2, "contrib", "contrib@example.com", "Contrib", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(contrib): %v", err)
	}
	if err = st.AddProjectMember(ctx, run.ProjectID, contrib.ID, projects.LocalRoleContributor); err != nil {
		t.Fatalf("AddProjectMember(): %v", err)
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

	tasks, err := st.ListAssignedRunItems(ctx, contrib.ID, 0, "")
	if err != nil || len(tasks) != 1 {
		t.Fatalf("ListAssignedRunItems() = %v, %v", tasks, err)
	}

	filtered, err := st.ListAssignedRunItems(ctx, contrib.ID, run.ProjectID, items.StatusPending)
	if err != nil || len(filtered) != 1 {
		t.Fatalf("ListAssignedRunItems(filter) = %v, %v", filtered, err)
	}

	empty, err := st.ListAssignedRunItems(ctx, contrib.ID, run.ProjectID, items.StatusOK)
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

func TestAssignRunItemRejectsNonMember(t *testing.T) {
	ctx := context.Background()
	st, run, itemID := setupInProgressRun(t)

	outsider, err := st.UpsertGitHubUser(ctx, 99, "outsider", "out@example.com", "Out", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	err = st.AssignRunItem(ctx, run.ID, itemID, &outsider.ID)
	if !errors.Is(err, store.ErrInvalidAssignee) {
		t.Fatalf("AssignRunItem() error = %v, want ErrInvalidAssignee", err)
	}
}
