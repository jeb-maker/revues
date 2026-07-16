package teams

import (
	"context"

	"github.com/jeb-maker/revues/internal/store"
)

// TeamStore is the persistence surface used by admin teams handlers.
type TeamStore interface {
	CreateTeam(ctx context.Context, name, slug, description string) (*store.OrganizationTeam, error)
	TeamByID(ctx context.Context, teamID int64) (*store.OrganizationTeam, error)
	ListOrganizationTeams(ctx context.Context) ([]store.OrganizationTeam, error)
	AddTeamMember(ctx context.Context, teamID, userID int64) error
	RemoveTeamMember(ctx context.Context, teamID, userID int64) error
	ListTeamMembers(ctx context.Context, teamID int64) ([]store.TeamMember, error)
	ListOrganizationMembers(ctx context.Context) ([]store.OrganizationMemberUser, error)
	OrganizationMemberRole(ctx context.Context, organizationID, userID int64) (string, bool, error)
}

var (
	ErrTeamNotFound       = store.ErrTeamNotFound
	ErrTeamMemberNotFound = store.ErrTeamMemberNotFound
	ErrTeamSlugTaken      = store.ErrTeamSlugTaken
)
