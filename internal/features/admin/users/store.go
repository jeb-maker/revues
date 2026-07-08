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

var ErrAllowedEmailNotFound = store.ErrAllowedEmailNotFound
