package store_test

import (
	"context"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	runs "github.com/jeb-maker/revues/internal/features/runs"
)

func TestUpdateRunItemStatusCreatesAuditEvent(t *testing.T) {
	ctx := context.Background()
	st, run, itemID := setupInProgressRun(t)

	lead, err := st.UpsertGitHubUser(ctx, 1, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	if err = st.UpdateRunItemStatus(ctx, run.ID, itemID, lead.ID, runs.StatusOK, ""); err != nil {
		t.Fatalf("UpdateRunItemStatus(ok): %v", err)
	}

	events, err := st.ListRunItemEvents(ctx, itemID)
	if err != nil {
		t.Fatalf("ListRunItemEvents(): %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("events len = %d, want 1", len(events))
	}
	if events[0].OldStatus.String != runs.StatusPending || events[0].NewStatus != runs.StatusOK {
		t.Fatalf("event = %+v", events[0])
	}
	if events[0].UserLogin != "lead" {
		t.Fatalf("user login = %q", events[0].UserLogin)
	}

	if err = st.UpdateRunItemStatus(ctx, run.ID, itemID, lead.ID, runs.StatusOK, "note"); err != nil {
		t.Fatalf("UpdateRunItemStatus(same status): %v", err)
	}
	events, err = st.ListRunItemEvents(ctx, itemID)
	if err != nil {
		t.Fatalf("ListRunItemEvents(): %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("same status should not create event, len = %d", len(events))
	}

	if err = st.UpdateRunItemStatus(ctx, run.ID, itemID, lead.ID, runs.StatusNOK, "problème"); err != nil {
		t.Fatalf("UpdateRunItemStatus(nok): %v", err)
	}
	events, err = st.ListRunItemEvents(ctx, itemID)
	if err != nil {
		t.Fatalf("ListRunItemEvents(): %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("events len = %d, want 2", len(events))
	}
	if events[0].NewStatus != runs.StatusNOK || events[0].Comment != "problème" {
		t.Fatalf("latest event = %+v", events[0])
	}
}
