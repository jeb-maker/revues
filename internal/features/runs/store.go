package runs

import (
	"context"
	"database/sql"
	"github.com/jeb-maker/revues/internal/store"
)

type RunStore interface {
	ProjectByID(ctx context.Context, id int64) (*store.Project, error)
	ListProjects(ctx context.Context, userID int64, admin bool) ([]store.Project, error)
	ListActiveRunSummaries(ctx context.Context, userID int64, admin bool) ([]store.ActiveRunSummary, error)
	OrganizationMemberRole(ctx context.Context, organizationID, userID int64) (string, bool, error)
	MemberRole(ctx context.Context, projectID, userID int64) (string, bool, error)
	ListProjectMembers(ctx context.Context, projectID int64) ([]store.ProjectMember, error)
	ChecklistTemplateByID(ctx context.Context, id int64) (*store.ChecklistTemplate, error)
	ListChecklistTemplates(ctx context.Context, projectID int64) ([]store.ChecklistTemplateSummary, error)
	LatestTemplateVersion(ctx context.Context, templateID int64) (*store.TemplateVersion, error)
	ListTemplateItems(ctx context.Context, versionID int64) ([]store.TemplateItem, error)
	TemplateVersionInfo(ctx context.Context, versionID int64) (*store.TemplateVersionInfo, error)
	CreateChecklistRun(ctx context.Context, projectID, templateID int64, title string, createdBy int64, dueDate sql.NullString) (*store.ChecklistRun, error)
	RunByID(ctx context.Context, id int64) (*store.ChecklistRun, error)
	StartRun(ctx context.Context, id int64) error
	CompleteRun(ctx context.Context, id int64, closingNote string) error
	RunItemByID(ctx context.Context, runID, itemID int64) (*store.RunItem, error)
	ListRunItems(ctx context.Context, runID int64) ([]store.RunItem, error)
	ListNokRunItems(ctx context.Context, runID int64) ([]store.RunItem, error)
	UpdateRunItemStatus(ctx context.Context, runID, itemID, userID int64, status, comment string) error
	AssignRunItem(ctx context.Context, runID, itemID int64, assigneeID *int64) error
	ListRunExportRows(ctx context.Context, runID int64) ([]store.RunExportRow, error)
	ListRunItemEvents(ctx context.Context, runItemID int64) ([]store.RunItemEvent, error)
	AttachmentByRunItemID(ctx context.Context, runItemID int64) (*store.Attachment, error)
	ListAttachmentsByRunItemIDs(ctx context.Context, runItemIDs []int64) (map[int64]*store.Attachment, error)
	RunIDForAttachment(ctx context.Context, attachmentID int64) (int64, error)
	IntegrationLinkByRunItemAndType(ctx context.Context, runItemID int64, integrationType string) (*store.IntegrationLink, error)
	ListIntegrationLinksByRunItemIDs(ctx context.Context, runItemIDs []int64, integrationType string) (map[int64]store.IntegrationLink, error)
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
