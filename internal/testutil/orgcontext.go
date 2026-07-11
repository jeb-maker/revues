package testutil

import (
	"context"

	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
)

// DefaultOrgContext scopes ctx to the migrated default organization.
func DefaultOrgContext(ctx context.Context, st *store.Store) context.Context {
	org, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		panic("default organization: " + err.Error())
	}
	return orgctx.WithOrganizationID(ctx, org.ID)
}
