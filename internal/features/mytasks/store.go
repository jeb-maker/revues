package mytasks

import (
	"context"

	"github.com/jeb-maker/revues/internal/store"
)

type AssignedTaskStore interface {
	ListAssignedRunItems(ctx context.Context, userID int64, projectID int64, status string) ([]AssignedRunItemSummary, error)
	ListProjects(ctx context.Context, userID int64, admin bool) ([]Project, error)
}

type AssignedRunItemSummary = store.AssignedRunItemSummary
type Project = store.Project
