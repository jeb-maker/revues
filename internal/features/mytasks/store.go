package mytasks

import (
	"context"

	"github.com/jeb-maker/revues/internal/store"
)

type AssignedTaskStore interface {
	ListAssignedRunItems(ctx context.Context, userID int64, status, query string) ([]AssignedRunItemSummary, error)
}

type AssignedRunItemSummary = store.AssignedRunItemSummary
