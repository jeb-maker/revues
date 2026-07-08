package auth

import (
	"context"

	"github.com/jeb-maker/revues/internal/store"
)

type AuthStore interface {
	UpsertGitHubUser(ctx context.Context, githubID int64, login, email, displayName, avatarURL, role string) (*store.User, error)
	ResolveLoginRole(ctx context.Context, email, bootstrapAdmin string) (string, error)
}

type User = store.User
