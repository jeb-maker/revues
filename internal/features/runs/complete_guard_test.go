package runs_test

import (
	"errors"
	"testing"

	"github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/store"
)

func TestPendingRequiredItems(t *testing.T) {
	t.Parallel()

	items := []store.RunItem{
		{Label: "req-pending", Required: true, Status: runs.StatusPending},
		{Label: "req-ok", Required: true, Status: runs.StatusOK},
		{Label: "opt-pending", Required: false, Status: runs.StatusPending},
		{Label: "req-nok", Required: true, Status: runs.StatusNOK},
	}

	pending := runs.PendingRequiredItems(items)
	if len(pending) != 1 {
		t.Fatalf("len(pending) = %d, want 1", len(pending))
	}
	if pending[0].Label != "req-pending" {
		t.Fatalf("pending[0].Label = %q, want req-pending", pending[0].Label)
	}
}

func TestValidateComplete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		items   []store.RunItem
		wantErr error
	}{
		{
			name: "all ok",
			items: []store.RunItem{
				{Required: true, Status: runs.StatusOK},
			},
		},
		{
			name: "nok allowed",
			items: []store.RunItem{
				{Required: true, Status: runs.StatusNOK},
				{Required: true, Status: runs.StatusOK},
			},
		},
		{
			name: "optional pending allowed",
			items: []store.RunItem{
				{Required: true, Status: runs.StatusOK},
				{Required: false, Status: runs.StatusPending},
			},
		},
		{
			name: "required pending blocked",
			items: []store.RunItem{
				{Required: true, Status: runs.StatusPending},
			},
			wantErr: runs.ErrPendingRequired,
		},
		{
			name: "na counts as treated",
			items: []store.RunItem{
				{Required: true, Status: runs.StatusNA},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := runs.ValidateComplete(tt.items)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("ValidateComplete() = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateComplete() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
