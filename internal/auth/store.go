package auth

import "context"

type SessionStore interface {
	CreateSession(ctx context.Context, userID, organizationID int64, tokenHash string) error
	DeleteUserSessions(ctx context.Context, userID int64) error
	DeleteSession(ctx context.Context, tokenHash string) error
	ResolveSessionOrganizationID(ctx context.Context, userID, preferredOrganizationID int64) (int64, error)
}
