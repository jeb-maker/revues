package organizations

import (
	"context"

	"github.com/jeb-maker/revues/internal/store"
)

// OrgStore is the persistence layer for organization onboarding and selection.
type OrgStore interface {
	CountUserOrganizations(ctx context.Context, userID int64) (int, error)
	ListUserOrganizations(ctx context.Context, userID int64) ([]store.OrganizationMembership, error)
	CreateOrganization(ctx context.Context, name, slug string, createdBy int64) (*store.Organization, error)
	AddOrganizationMember(ctx context.Context, organizationID, userID int64, role string) error
	OrganizationMemberRole(ctx context.Context, organizationID, userID int64) (string, bool, error)
	OrganizationByID(ctx context.Context, id int64) (*store.Organization, error)
	OrganizationInvitationByID(ctx context.Context, id int64) (*store.OrganizationInvitation, error)
	DeleteOrganizationInvitation(ctx context.Context, id int64) error
	ListPendingInvitationsByEmail(ctx context.Context, email string) ([]store.OrganizationInvitation, error)
	SessionByTokenHash(ctx context.Context, tokenHash string) (userID, organizationID int64, err error)
}
