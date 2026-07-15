package testutil

import (
	"context"

	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
)

// OrgContext scopes ctx to the given organization id.
func OrgContext(ctx context.Context, orgID int64) context.Context {
	return orgctx.WithOrganizationID(ctx, orgID)
}

// SetupIsolatedOrg creates an organization owned by ownerID and returns scoped context.
func SetupIsolatedOrg(ctx context.Context, st *store.Store, name, slug string, ownerID int64) context.Context {
	org, err := st.CreateOrganization(ctx, name, slug, ownerID)
	if err != nil {
		panic("create organization: " + err.Error())
	}
	if err := st.AddOrganizationMember(ctx, org.ID, ownerID, store.OrgRoleOwner); err != nil {
		panic("add organization owner: " + err.Error())
	}
	return OrgContext(ctx, org.ID)
}

func DefaultOrgContext(ctx context.Context, st *store.Store) context.Context {
	org, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		panic("default organization: " + err.Error())
	}
	return orgctx.WithOrganizationID(ctx, org.ID)
}
