package users

import (
	"context"

	"github.com/jeb-maker/revues/internal/store"
)

type AllowedEmailStore interface {
	ListAllowedEmails(ctx context.Context) ([]store.AllowedEmail, error)
	InsertAllowedEmail(ctx context.Context, email, role string) error
	DeleteAllowedEmail(ctx context.Context, email string) error
}

// MemberStore is the persistence surface for org member role admin.
type MemberStore interface {
	ListOrganizationMembers(ctx context.Context) ([]store.OrganizationMemberUser, error)
	OrganizationMemberUserByID(ctx context.Context, userID int64) (*store.OrganizationMemberUser, error)
	AddOrganizationMember(ctx context.Context, organizationID, userID int64, role string) error
	CountOrganizationMembersWithRole(ctx context.Context, organizationID int64, role string) (int, error)
}

var (
	ErrAllowedEmailNotFound       = store.ErrAllowedEmailNotFound
	ErrOrganizationMemberNotFound = store.ErrOrganizationMemberNotFound
	ErrLastOrganizationOwner      = store.ErrLastOrganizationOwner
)
