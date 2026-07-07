package runs

import (
	"github.com/jeb-maker/revues/internal/store"
)

// Store wraps a *store.Store to expose run-related queries from the runs
// feature package. It embeds *store.Store so existing run SQL methods
// (CreateChecklistRun, RunByID, StartRun, ListRunItems, RunItemByID,
// UpdateRunItemStatus, AssignRunItem, ListRunItemEvents, ListRunExportRows,
// etc.) are promoted and accessible through the runs feature namespace.
//
// The underlying SQL stays in internal/store because it is shared by other
// features (notifications, jira, notion integrations, dashboard, my_tasks)
// and by router-level tests. A dedicated issue will migrate the SQL into this
// package once the depending features are themselves extracted.
type Store struct {
	*store.Store
}

// New returns a runs Store backed by the given store.Store.
func New(s *store.Store) *Store {
	return &Store{Store: s}
}

// ChecklistRun re-exports store.ChecklistRun under the runs namespace.
type ChecklistRun = store.ChecklistRun

// RunItem re-exports store.RunItem under the runs namespace.
type RunItem = store.RunItem

// RunItemEvent re-exports store.RunItemEvent under the runs namespace.
type RunItemEvent = store.RunItemEvent

// RunExportRow re-exports store.RunExportRow under the runs namespace.
type RunExportRow = store.RunExportRow

// AssignedRunItemSummary re-exports store.AssignedRunItemSummary.
type AssignedRunItemSummary = store.AssignedRunItemSummary

// ErrRunNotFound re-exports store.ErrRunNotFound under the runs namespace.
var ErrRunNotFound = store.ErrRunNotFound

// ErrInvalidRunStatus re-exports store.ErrInvalidRunStatus.
var ErrInvalidRunStatus = store.ErrInvalidRunStatus

// ErrRunItemNotFound re-exports store.ErrRunItemNotFound.
var ErrRunItemNotFound = store.ErrRunItemNotFound

// ErrRunNotEditable re-exports store.ErrRunNotEditable.
var ErrRunNotEditable = store.ErrRunNotEditable

// ErrInvalidAssignee re-exports store.ErrInvalidAssignee.
var ErrInvalidAssignee = store.ErrInvalidAssignee

// Item status constants re-exported under the runs namespace (formerly the
// items package). They mirror store.RunItemStatus* values.
const (
	StatusPending = store.RunItemStatusPending
	StatusOK      = store.RunItemStatusOK
	StatusNOK     = store.RunItemStatusNOK
	StatusNA      = store.RunItemStatusNA
)
