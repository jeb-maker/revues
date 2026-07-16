package subjects

import (
	"context"

	"github.com/jeb-maker/revues/internal/store"
)

// SubjectStore is the persistence layer for subject HTTP handlers.
type SubjectStore interface {
	SubjectByID(ctx context.Context, id int64) (*Subject, error)
	ListSubjects(ctx context.Context, userID int64, admin bool, query string) ([]Subject, error)
	CreateSubject(ctx context.Context, name, description string, creatorID int64, domains []string) (*Subject, error)
	CreateSubjectWithVisibility(ctx context.Context, name, description string, creatorID int64, domains []string, visibility string) (*Subject, error)
	UpdateSubject(ctx context.Context, id int64, name, description string, domains []string) error
	UpdateSubjectWithVisibility(ctx context.Context, id int64, name, description string, domains []string, visibility string) error
	ListSubjectDomains(ctx context.Context, subjectID int64) ([]string, error)
	ListSubjectTags(ctx context.Context, subjectID int64) ([]string, error)
	SetSubjectTags(ctx context.Context, subjectID int64, tags []string) error
	ArchiveSubject(ctx context.Context, id int64) error
	ListSubjectMembers(ctx context.Context, subjectID int64) ([]SubjectMember, error)
	ListDirectSubjectMembers(ctx context.Context, subjectID int64) ([]DirectSubjectMember, error)
	UpsertDirectSubjectMember(ctx context.Context, subjectID, userID int64, role string) error
	ListRunsWithProgressBySubject(ctx context.Context, subjectID int64) ([]RunWithProgress, error)
	ListSubjectNokItems(ctx context.Context, subjectID int64) ([]SubjectNokItemSummary, error)
	OrganizationMemberRole(ctx context.Context, organizationID, userID int64) (string, bool, error)
	ResolveSubjectAccess(ctx context.Context, userID, subjectID int64, globalRole string) (store.SubjectAccess, error)
	ListOrganizationTeams(ctx context.Context) ([]OrganizationTeam, error)
	ListSubjectTeams(ctx context.Context, subjectID int64) ([]TeamSubjectRole, error)
	ListTeamMembers(ctx context.Context, teamID int64) ([]TeamMember, error)
	TeamByID(ctx context.Context, teamID int64) (*OrganizationTeam, error)
	GrantTeamSubjectRole(ctx context.Context, teamID, subjectID int64, role string, grantedBy int64) error
	RevokeTeamSubjectRole(ctx context.Context, teamID, subjectID int64) error
}

type Subject = store.Subject
type SubjectMember = store.SubjectMember
type DirectSubjectMember = store.DirectSubjectMember
type OrganizationTeam = store.OrganizationTeam
type TeamMember = store.TeamMember
type TeamSubjectRole = store.TeamSubjectRole
type RunWithProgress = store.RunWithProgress
type SubjectNokItemSummary = store.SubjectNokItemSummary

var ErrSubjectNotFound = store.ErrSubjectNotFound
var ErrTeamNotFound = store.ErrTeamNotFound
var ErrInvalidSubjectRole = store.ErrInvalidSubjectRole
var ErrTeamSubjectRoleNotFound = store.ErrTeamSubjectRoleNotFound
