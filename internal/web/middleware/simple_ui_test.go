package middleware

import (
	"context"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
)

func TestResolveSimpleUI_Particulier(t *testing.T) {
	ctx := context.Background()
	db := openOrgAdminTestDB(t)
	st := store.New(db)

	user, err := st.UpsertGitHubUser(ctx, 50, "particulier", "particulier@example.com", "Camille", "", auth.RoleEditor)
	if err != nil {
		t.Fatal(err)
	}
	org, err := st.CreateOrganization(ctx, "Perso", "perso-ui", user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = st.AddOrganizationMember(ctx, org.ID, user.ID, store.OrgRoleOwner); err != nil {
		t.Fatal(err)
	}
	octx := orgctx.WithOrganizationID(ctx, org.ID)
	if err = st.InsertAllowedEmail(octx, user.Email, auth.RoleEditor); err != nil {
		t.Fatal(err)
	}
	sub, err := st.CreateSubjectWithVisibility(octx, "Chez moi", "", user.ID, nil, store.SubjectVisibilityPrivate)
	if err != nil {
		t.Fatal(err)
	}

	reqCtx := testOrgNavCtx(t, st, org, user)
	hd := HeaderData{
		UserOrganizations: []store.OrganizationMembership{{Organization: *org, Role: store.OrgRoleOwner}},
	}
	simple, soleID := resolveSimpleUI(reqCtx, st, user, hd)
	if !simple {
		t.Fatal("expected simple UI for particulier")
	}
	if soleID != sub.ID {
		t.Fatalf("SimpleSubjectID = %d, want %d", soleID, sub.ID)
	}
}

func TestResolveSimpleUI_DefaultOrgMemberNotSimple(t *testing.T) {
	ctx := context.Background()
	db := openOrgAdminTestDB(t)
	st := store.New(db)

	owner, err := st.UpsertGitHubUser(ctx, 51, "owner", "owner@example.com", "Owner", "", auth.RoleEditor)
	if err != nil {
		t.Fatal(err)
	}
	member, err := st.UpsertGitHubUser(ctx, 52, "member", "member@example.com", "Member", "", auth.RoleEditor)
	if err != nil {
		t.Fatal(err)
	}
	org, err := st.CreateOrganization(ctx, "Team", "team-ui", owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = st.AddOrganizationMember(ctx, org.ID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatal(err)
	}
	if err = st.AddOrganizationMember(ctx, org.ID, member.ID, store.OrgRoleMember); err != nil {
		t.Fatal(err)
	}
	octx := orgctx.WithOrganizationID(ctx, org.ID)
	if _, err = st.CreateSubject(octx, "Shared", "", owner.ID, nil); err != nil {
		t.Fatal(err)
	}

	reqCtx := testOrgNavCtx(t, st, org, member)
	hd := HeaderData{
		UserOrganizations: []store.OrganizationMembership{{Organization: *org, Role: store.OrgRoleMember}},
	}
	if simple, _ := resolveSimpleUI(reqCtx, st, member, hd); simple {
		t.Fatal("multi-member org must not use simple UI")
	}
}

func TestResolveUICaps_DuoUnlocksAssign(t *testing.T) {
	ctx := context.Background()
	db := openOrgAdminTestDB(t)
	st := store.New(db)

	owner, err := st.UpsertGitHubUser(ctx, 53, "owner2", "owner2@example.com", "Owner", "", auth.RoleEditor)
	if err != nil {
		t.Fatal(err)
	}
	member, err := st.UpsertGitHubUser(ctx, 54, "member2", "member2@example.com", "Member", "", auth.RoleEditor)
	if err != nil {
		t.Fatal(err)
	}
	org, err := st.CreateOrganization(ctx, "Foyer", "foyer-ui", owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = st.AddOrganizationMember(ctx, org.ID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatal(err)
	}
	if err = st.AddOrganizationMember(ctx, org.ID, member.ID, store.OrgRoleMember); err != nil {
		t.Fatal(err)
	}
	octx := orgctx.WithOrganizationID(ctx, org.ID)
	if _, err = st.CreateSubject(octx, "Maison", "", owner.ID, nil); err != nil {
		t.Fatal(err)
	}

	reqCtx := testOrgNavCtx(t, st, org, owner)
	hd := HeaderData{
		UserOrganizations: []store.OrganizationMembership{{Organization: *org, Role: store.OrgRoleOwner}},
	}
	caps := resolveUICaps(reqCtx, st, owner, hd)
	if caps.SimpleUI {
		t.Fatal("duo must not be SimpleUI")
	}
	if !caps.ShowAssign || !caps.ShowMyTasks || !caps.ShowCollab {
		t.Fatalf("duo caps = %+v, want assign/tasks/collab", caps)
	}
	if caps.ShowSubjectColumn {
		t.Fatal("single subject must not show subject column")
	}
}

func TestResolveUICaps_SoloMultiSubject(t *testing.T) {
	ctx := context.Background()
	db := openOrgAdminTestDB(t)
	st := store.New(db)

	user, err := st.UpsertGitHubUser(ctx, 55, "solo2", "solo2@example.com", "Solo", "", auth.RoleEditor)
	if err != nil {
		t.Fatal(err)
	}
	org, err := st.CreateOrganization(ctx, "Perso2", "perso2-ui", user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = st.AddOrganizationMember(ctx, org.ID, user.ID, store.OrgRoleOwner); err != nil {
		t.Fatal(err)
	}
	octx := orgctx.WithOrganizationID(ctx, org.ID)
	if err = st.InsertAllowedEmail(octx, user.Email, auth.RoleEditor); err != nil {
		t.Fatal(err)
	}
	if _, err = st.CreateSubjectWithVisibility(octx, "Maison", "", user.ID, nil, store.SubjectVisibilityPrivate); err != nil {
		t.Fatal(err)
	}
	if _, err = st.CreateSubjectWithVisibility(octx, "Bureau", "", user.ID, nil, store.SubjectVisibilityPrivate); err != nil {
		t.Fatal(err)
	}

	reqCtx := testOrgNavCtx(t, st, org, user)
	hd := HeaderData{
		UserOrganizations: []store.OrganizationMembership{{Organization: *org, Role: store.OrgRoleOwner}},
	}
	caps := resolveUICaps(reqCtx, st, user, hd)
	if caps.SimpleUI {
		t.Fatal("multi-subject must not be SimpleUI")
	}
	if caps.ShowAssign || caps.ShowMyTasks || caps.ShowCollab {
		t.Fatalf("solo caps = %+v, want no collab", caps)
	}
	if !caps.ShowSubjectColumn {
		t.Fatal("expected subject column for ≥2 subjects")
	}
}

func TestResolveSimpleUI_GlobalAdminNeverSimple(t *testing.T) {
	ctx := context.Background()
	db := openOrgAdminTestDB(t)
	st := store.New(db)

	admin, err := st.UpsertGitHubUser(ctx, 91003, "gadmin", "gadmin@example.com", "GAdmin", "", auth.RoleAdmin)
	if err != nil {
		t.Fatal(err)
	}
	org, err := st.CreateOrganization(ctx, "Solo Admin Org", "solo-admin-org", admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	octx := orgctx.WithOrganizationID(ctx, org.ID)
	if _, err = st.CreateSubjectWithVisibility(octx, "Seul", "", admin.ID, nil, store.SubjectVisibilityPrivate); err != nil {
		t.Fatal(err)
	}

	reqCtx := testOrgNavCtx(t, st, org, admin)
	hd := HeaderData{
		UserOrganizations: []store.OrganizationMembership{{Organization: *org, Role: store.OrgRoleOwner}},
	}
	if simple, _ := resolveSimpleUI(reqCtx, st, admin, hd); simple {
		t.Fatal("global admin must never be SimpleUI")
	}
}
