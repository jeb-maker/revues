package store

import (
	"context"
	"errors"

	"github.com/jeb-maker/revues/internal/orgctx"
)

// ErrOrganizationRequired is returned when organization context is missing.
var ErrOrganizationRequired = errors.New("organization context required")

func organizationIDFromContext(ctx context.Context) (int64, error) {
	orgID, ok := orgctx.OrganizationID(ctx)
	if !ok {
		return 0, ErrOrganizationRequired
	}
	return orgID, nil
}

func optionalOrganizationIDFromContext(ctx context.Context) int64 {
	orgID, ok := orgctx.OrganizationID(ctx)
	if !ok {
		return 0
	}
	return orgID
}
