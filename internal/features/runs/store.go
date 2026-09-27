package runs

import (
	"context"

	"github.com/jeb-maker/revues/internal/store"
)

// Small consumer-side interfaces (GO.md). RunStore composes them for Deps wiring.

// RunListStore backs the hub / filters.
type RunListStore interface {
	ListFilteredRunSummaries(ctx context.Context, userID int64, admin bool, status, query string, limit, offset int) ([]store.RunListSummary, int, error)
	ListActiveRunSummaries(ctx context.Context, userID int64, admin bool) ([]store.ActiveRunSummary, error)
	ListRecentCompletedRunSummaries(ctx context.Context, userID int64, admin bool) ([]store.CompletedRunSummary, error)
	ListSubjects(ctx context.Context, userID int64, admin bool, query string) ([]store.Subject, error)
}

// RunAccessStore resolves subject/run visibility.
type RunAccessStore interface {
	SubjectByID(ctx context.Context, id int64) (*store.Subject, error)
	ResolveSubjectAccess(ctx context.Context, userID, subjectID int64, globalRole string) (store.SubjectAccess, error)
	OrganizationMemberRole(ctx context.Context, organizationID, userID int64) (string, bool, error)
	UserByID(ctx context.Context, id int64) (*store.User, error)
}

// RunLaunchStore creates runs from templates.
type RunLaunchStore interface {
	ChecklistTemplateByID(ctx context.Context, id int64) (*store.ChecklistTemplate, error)
	ListChecklistTemplates(ctx context.Context, subjectID int64) ([]store.ChecklistTemplateSummary, error)
	LatestTemplateVersion(ctx context.Context, templateID int64) (*store.TemplateVersion, error)
	ListTemplateItems(ctx context.Context, versionID int64) ([]store.TemplateItem, error)
	TemplateMatchesSubject(ctx context.Context, subjectID, templateID int64) (bool, error)
	CreateChecklistRun(ctx context.Context, subjectID, templateID int64, createdBy int64) (*store.ChecklistRun, error)
}

// RunShowStore loads a run and related display data.
type RunShowStore interface {
	RunByID(ctx context.Context, id int64) (*store.ChecklistRun, error)
	ListRunItems(ctx context.Context, runID int64) ([]store.RunItem, error)
	ListNokRunItems(ctx context.Context, runID int64) ([]store.RunItem, error)
	TemplateVersionInfo(ctx context.Context, versionID int64) (*store.TemplateVersionInfo, error)
	ListSubjectMembers(ctx context.Context, subjectID int64) ([]store.SubjectMember, error)
	ListAttachmentsByRunItemIDs(ctx context.Context, runItemIDs []int64) (map[int64]*store.Attachment, error)
	ListIntegrationLinksByRunItemIDs(ctx context.Context, runItemIDs []int64, integrationType string) (map[int64]store.IntegrationLink, error)
}

// RunItemStore mutates checklist points.
type RunItemStore interface {
	RunItemByID(ctx context.Context, runID, itemID int64) (*store.RunItem, error)
	UpdateRunItemStatus(ctx context.Context, runID, itemID, userID int64, status, comment string) error
	AssignRunItem(ctx context.Context, runID, itemID int64, assigneeID *int64) error
	ListRunItemEvents(ctx context.Context, runItemID int64) ([]store.RunItemEvent, error)
	AttachmentByRunItemID(ctx context.Context, runItemID int64) (*store.Attachment, error)
	RunIDForAttachment(ctx context.Context, attachmentID int64) (int64, error)
	IntegrationLinkByRunItemAndType(ctx context.Context, runItemID int64, integrationType string) (*store.IntegrationLink, error)
}

// RunLifecycleStore starts and completes runs.
type RunLifecycleStore interface {
	StartRun(ctx context.Context, id int64) error
	CompleteRun(ctx context.Context, id int64, closingNote string) error
	CompleteRunWithEvidence(ctx context.Context, id int64, closingNote, csvSHA256 string) error
	SealRunEvidenceHash(ctx context.Context, id int64, csvSHA256 string) error
}

// RunExportStore backs CSV / evidence export.
type RunExportStore interface {
	ListRunExportRows(ctx context.Context, runID int64) ([]store.RunExportRow, error)
}

// RunStore is the full dependency surface for runs handlers (composition of the above).
type RunStore interface {
	RunListStore
	RunAccessStore
	RunLaunchStore
	RunShowStore
	RunItemStore
	RunLifecycleStore
	RunExportStore
}

type Store struct{ *store.Store }

func New(s *store.Store) *Store { return &Store{Store: s} }

type ChecklistRun = store.ChecklistRun
type RunItem = store.RunItem
type RunItemEvent = store.RunItemEvent
type RunExportRow = store.RunExportRow
type AssignedRunItemSummary = store.AssignedRunItemSummary

var ErrRunNotFound = store.ErrRunNotFound
var ErrInvalidRunStatus = store.ErrInvalidRunStatus
var ErrRunItemNotFound = store.ErrRunItemNotFound
var ErrRunNotEditable = store.ErrRunNotEditable
var ErrInvalidAssignee = store.ErrInvalidAssignee

const (
	StatusPending = store.RunItemStatusPending
	StatusOK      = store.RunItemStatusOK
	StatusNOK     = store.RunItemStatusNOK
	StatusNA      = store.RunItemStatusNA
)
