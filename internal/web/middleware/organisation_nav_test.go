package middleware

import (
	"context"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
)

func testOrgNavCtx(t *testing.T, st *store.Store, org *store.Organization, user *store.User) context.Context {
	t.Helper()
	reqCtx := orgctx.WithOrganizationID(context.Background(), org.ID)
	reqCtx = context.WithValue(reqCtx, userContextKey, user)
	reqCtx = context.WithValue(reqCtx, orgContextKey, org)
	return reqCtx
}

func TestShowOrganisationNav_SoloHidden(t *testing.T) {
	ctx := context.Background()
	db := openOrgAdminTestDB(t)
	st := store.New(db)

	owner, err := st.UpsertGitHubUser(ctx, 1, "solo", "solo@example.com", "Solo", "", auth.RoleEditor)
	if err != nil {
		t.Fatal(err)
	}
	org, err := st.CreateOrganization(ctx, "Solo", "solo", owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.AddOrganizationMember(ctx, org.ID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatal(err)
	}
	if err := st.InsertAllowedEmail(orgctx.WithOrganizationID(ctx, org.ID), "solo@example.com", auth.RoleEditor); err != nil {
		t.Fatal(err)
	}

	reqCtx := testOrgNavCtx(t, st, org, owner)
	hd := HeaderData{
		UserOrganizations: []store.OrganizationMembership{{Organization: *org, Role: store.OrgRoleOwner}},
	}

	if showOrganisationNav(reqCtx, st, owner, hd) {
		t.Fatal("solo owner should not see Organisation nav")
	}
}

func TestShowOrganisationNav_GlobalAdminAlways(t *testing.T) {
	ctx := context.Background()
	db := openOrgAdminTestDB(t)
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	admin, err := st.UpsertGitHubUser(ctx, 2, "admin", "admin@example.com", "Admin", "", auth.RoleAdmin)
	if err != nil {
		t.Fatal(err)
	}
	org, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatal(err)
	}

	reqCtx := testOrgNavCtx(t, st, org, admin)
	hd := HeaderData{UserOrganizations: []store.OrganizationMembership{{Organization: *org, Role: store.OrgRoleOwner}}}

	if !showOrganisationNav(reqCtx, st, admin, hd) {
		t.Fatal("global admin should see Organisation nav")
	}
}

func TestShowOrganisationNav_SecondEmailShows(t *testing.T) {
	ctx := context.Background()
	db := openOrgAdminTestDB(t)
	st := store.New(db)

	owner, err := st.UpsertGitHubUser(ctx, 3, "owner", "owner@example.com", "Owner", "", auth.RoleEditor)
	if err != nil {
		t.Fatal(err)
	}
	org, err := st.CreateOrganization(ctx, "Team", "team", owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.AddOrganizationMember(ctx, org.ID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatal(err)
	}
	octx := orgctx.WithOrganizationID(ctx, org.ID)
	if err := st.InsertAllowedEmail(octx, "owner@example.com", auth.RoleEditor); err != nil {
		t.Fatal(err)
	}
	if err := st.InsertAllowedEmail(octx, "colleague@example.com", auth.RoleReader); err != nil {
		t.Fatal(err)
	}

	reqCtx := testOrgNavCtx(t, st, org, owner)
	hd := HeaderData{
		UserOrganizations: []store.OrganizationMembership{{Organization: *org, Role: store.OrgRoleOwner}},
	}

	if !showOrganisationNav(reqCtx, st, owner, hd) {
		t.Fatal("should show Organisation nav when whitelist has multiple emails")
	}
}
