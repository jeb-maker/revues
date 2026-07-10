package projects

import (
	"context"

	"github.com/jeb-maker/revues/internal/store"
)

type ProjectStore interface {
	ProjectByID(ctx context.Context, id int64) (*Project, error)
	ListProjects(ctx context.Context, userID int64, admin bool) ([]Project, error)
	CreateProject(ctx context.Context, name, description string, creatorID int64) (*Project, error)
	UpdateProject(ctx context.Context, id int64, name, description string) error
	ArchiveProject(ctx context.Context, id int64) error
	AddProjectMember(ctx context.Context, projectID, userID int64, role string) error
	RemoveProjectMember(ctx context.Context, projectID, userID int64) error
	MemberRole(ctx context.Context, projectID, userID int64) (string, bool, error)
	ListProjectMembers(ctx context.Context, projectID int64) ([]ProjectMember, error)
	ListActiveRunSummaries(ctx context.Context, userID int64, admin bool) ([]ActiveRunSummary, error)
	ListRunsWithProgressByProject(ctx context.Context, projectID int64) ([]RunWithProgress, error)
	ListProjectNokItems(ctx context.Context, projectID int64) ([]ProjectNokItemSummary, error)
	UserByEmail(ctx context.Context, email string) (*User, error)
	CountProjectLeads(ctx context.Context, projectID int64) (int, error)
	OrganizationMemberRole(ctx context.Context, organizationID, userID int64) (string, bool, error)
	AddOrganizationMember(ctx context.Context, organizationID, userID int64, role string) error
}

type Project = store.Project
type ProjectMember = store.ProjectMember
type User = store.User
type ActiveRunSummary = store.ActiveRunSummary
type RunWithProgress = store.RunWithProgress
type ProjectNokItemSummary = store.ProjectNokItemSummary

var ErrProjectNotFound = store.ErrProjectNotFound
var ErrUserNotFound = store.ErrUserNotFound
