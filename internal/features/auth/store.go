package auth

import (
	"context"

	"github.com/jeb-maker/revues/internal/store"
)

type AuthStore interface {
	UpsertGitHubUser(ctx context.Context, githubID int64, login, email, displayName, avatarURL, role string) (*store.User, error)
	UserByID(ctx context.Context, id int64) (*store.User, error)
	ListUsers(ctx context.Context) ([]store.User, error)
	ResolveLoginRole(ctx context.Context, email, bootstrapAdmin string) (string, error)
	EnsureBootstrapOrgOwner(ctx context.Context, userID int64, email, bootstrapAdmin string) error
	CountUserOrganizations(ctx context.Context, userID int64) (int, error)
	ListUserOrganizations(ctx context.Context, userID int64) ([]store.OrganizationMembership, error)
}

type User = store.User
