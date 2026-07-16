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
	UpdateSubject(ctx context.Context, id int64, name, description string, domains []string) error
	ListSubjectDomains(ctx context.Context, subjectID int64) ([]string, error)
	ListSubjectTags(ctx context.Context, subjectID int64) ([]string, error)
	SetSubjectTags(ctx context.Context, subjectID int64, tags []string) error
	ArchiveSubject(ctx context.Context, id int64) error
	ListSubjectMembers(ctx context.Context, subjectID int64) ([]SubjectMember, error)
	ListRunsWithProgressBySubject(ctx context.Context, subjectID int64) ([]RunWithProgress, error)
	ListSubjectNokItems(ctx context.Context, subjectID int64) ([]SubjectNokItemSummary, error)
	OrganizationMemberRole(ctx context.Context, organizationID, userID int64) (string, bool, error)
}

type Subject = store.Subject
type SubjectMember = store.SubjectMember
type RunWithProgress = store.RunWithProgress
type SubjectNokItemSummary = store.SubjectNokItemSummary

var ErrSubjectNotFound = store.ErrSubjectNotFound
