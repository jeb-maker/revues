package auth

import (
	"context"

	"github.com/jeb-maker/revues/internal/store"
)

type AuthStore interface {
	UpsertGitHubUser(ctx context.Context, githubID int64, login, email, displayName, avatarURL, role string) (*store.User, error)
	ResolveLoginRole(ctx context.Context, email, bootstrapAdmin string) (string, error)
	CountUserOrganizations(ctx context.Context, userID int64) (int, error)
	ListUserOrganizations(ctx context.Context, userID int64) ([]store.OrganizationMembership, error)
}

type User = store.User
