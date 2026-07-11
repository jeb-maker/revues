package checklisttemplates

import (
	"context"

	"github.com/jeb-maker/revues/internal/store"
)

type ChecklistTemplateStore interface {
	ChecklistTemplateByID(ctx context.Context, id int64) (*ChecklistTemplate, error)
	ListChecklistTemplates(ctx context.Context, projectID int64) ([]ChecklistTemplateSummary, error)
	CreateChecklistTemplate(ctx context.Context, name string, createdBy int64, tags []string, items []TemplateItemInput) (*ChecklistTemplate, *TemplateVersion, error)
	ArchiveChecklistTemplate(ctx context.Context, id int64) error
	LatestTemplateVersion(ctx context.Context, templateID int64) (*TemplateVersion, error)
	CreateTemplateVersion(ctx context.Context, templateID, createdBy int64, items []TemplateItemInput) (*TemplateVersion, error)
	ListTemplateItems(ctx context.Context, versionID int64) ([]TemplateItem, error)
	ListTemplateIndex(ctx context.Context, userID int64, admin bool, query string) ([]TemplateIndexRow, error)
	UpdateChecklistTemplateName(ctx context.Context, id int64, name string) error
	SetTemplateTags(ctx context.Context, templateID int64, tags []string) error
	ListTemplateTags(ctx context.Context, templateID int64) ([]string, error)
	TemplateMatchesProject(ctx context.Context, projectID, templateID int64) (bool, error)
	ListProjectTags(ctx context.Context, projectID int64) ([]string, error)
	ProjectByID(ctx context.Context, id int64) (*Project, error)
	MemberRole(ctx context.Context, projectID, userID int64) (string, bool, error)
}

type ChecklistTemplate = store.ChecklistTemplate
type ChecklistTemplateSummary = store.ChecklistTemplateSummary
type TemplateVersion = store.TemplateVersion
type TemplateItem = store.TemplateItem
type TemplateItemInput = store.TemplateItemInput
type Project = store.Project
type TemplateIndexRow = store.TemplateIndexRow

var ErrChecklistTemplateNotFound = store.ErrChecklistTemplateNotFound
