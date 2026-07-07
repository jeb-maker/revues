package runs_test

import (
	"testing"

	"github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/store"
)

func TestProgress(t *testing.T) {
	t.Parallel()

	runItems := []store.RunItem{
		{Status: runs.StatusOK},
		{Status: runs.StatusNA},
		{Status: runs.StatusNOK},
		{Status: runs.StatusPending},
	}

	done, total := runs.Progress(runItems)
	if total != 4 {
		t.Fatalf("total = %d, want 4", total)
	}
	if done != 2 {
		t.Fatalf("done = %d, want 2", done)
	}
}
