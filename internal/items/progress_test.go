package items_test

import (
	"testing"

	"github.com/jeb-maker/revues/internal/items"
	"github.com/jeb-maker/revues/internal/store"
)

func TestProgress(t *testing.T) {
	t.Parallel()

	runItems := []store.RunItem{
		{Status: items.StatusOK},
		{Status: items.StatusNA},
		{Status: items.StatusNOK},
		{Status: items.StatusPending},
	}

	done, total := items.Progress(runItems)
	if total != 4 {
		t.Fatalf("total = %d, want 4", total)
	}
	if done != 2 {
		t.Fatalf("done = %d, want 2", done)
	}
}
