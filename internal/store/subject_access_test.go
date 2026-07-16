package store_test

import (
	"context"
	"slices"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
)

func TestResolveSubjectAccess(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	admin, err := st.UpsertGitHubUser(ctx, 1, "admin", "admin@example.com", "Admin", "", auth.RoleAdmin)
	if err != nil {
		t.Fatal(err)
	}
	orgAdmin, err := st.UpsertGitHubUser(ctx, 2, "orgadmin", "orgadmin@example.com", "OrgAdmin", "", auth.RoleEditor)
	if err != nil {
		t.Fatal(err)
	}
	lead, err := st.UpsertGitHubUser(ctx, 3, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatal(err)
	}
	viaTeam, err := st.UpsertGitHubUser(ctx, 4, "viaTeam", "viateam@example.com", "ViaTeam", "", auth.RoleEditor)
	if err != nil {
		t.Fatal(err)
	}
	outsider, err := st.UpsertGitHubUser(ctx, 5, "outsider", "outsider@example.com", "Outsider", "", auth.RoleEditor)
	if err != nil {
		t.Fatal(err)
	}
	both, err := st.UpsertGitHubUser(ctx, 6, "both", "both@example.com", "Both", "", auth.RoleEditor)
	if err != nil {
		t.Fatal(err)
	}

	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range []struct {
		id   int64
		role string
	}{
		{admin.ID, store.OrgRoleMember},
		{orgAdmin.ID, store.OrgRoleAdmin},
		{lead.ID, store.OrgRoleMember},
		{viaTeam.ID, store.OrgRoleMember},
		{both.ID, store.OrgRoleMember},
		{outsider.ID, store.OrgRoleMember},
	} {
		if err = st.AddOrganizationMember(ctx, defaultOrg.ID, m.id, m.role); err != nil {
			t.Fatal(err)
		}
	}

	subject, err := st.CreateSubject(ctx, "Alpha", "", lead.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	team, err := st.CreateTeam(ctx, "Squad", "squad", "")
	if err != nil {
		t.Fatal(err)
	}
	if err = st.AddTeamMember(ctx, team.ID, viaTeam.ID); err != nil {
		t.Fatal(err)
	}
	if err = st.AddTeamMember(ctx, team.ID, both.ID); err != nil {
		t.Fatal(err)
	}
	if err = st.GrantTeamSubjectRole(ctx, team.ID, subject.ID, store.SubjectRoleViewer, lead.ID); err != nil {
		t.Fatal(err)
	}
	if err = st.UpsertDirectSubjectMember(ctx, subject.ID, lead.ID, store.SubjectRoleLead); err != nil {
		t.Fatal(err)
	}
	if err = st.UpsertDirectSubjectMember(ctx, subject.ID, both.ID, store.SubjectRoleContributor); err != nil {
		t.Fatal(err)
	}
	if err = st.UpsertDirectSubjectMember(ctx, subject.ID, orgAdmin.ID, store.SubjectRoleLead); err != nil {
		t.Fatal(err)
	}

	otherOrg, err := st.CreateOrganization(ctx, "Other", "other-access", lead.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = st.AddOrganizationMember(ctx, otherOrg.ID, lead.ID, store.OrgRoleOwner); err != nil {
		t.Fatal(err)
	}
	otherCtx := orgctx.WithOrganizationID(ctx, otherOrg.ID)
	otherSubject, err := st.CreateSubject(otherCtx, "Secret", "", lead.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	ungated, err := st.CreateSubject(ctx, "LegacyOpen", "", lead.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name           string
		ctx            context.Context
		userID         int64
		subjectID      int64
		globalRole     string
		wantVis        bool
		wantRole       string
		wantSource     string
		wantAlsoDirect bool
	}{
		{
			name: "global admin", ctx: ctx, userID: admin.ID, subjectID: subject.ID, globalRole: auth.RoleAdmin,
			wantVis: true, wantSource: store.AccessSourceGlobalAdmin,
		},
		{
			name: "org admin with direct lead", ctx: ctx, userID: orgAdmin.ID, subjectID: subject.ID, globalRole: auth.RoleEditor,
			wantVis: true, wantRole: store.SubjectRoleLead, wantSource: store.AccessSourceOrgAdmin, wantAlsoDirect: true,
		},
		{
			name: "direct lead", ctx: ctx, userID: lead.ID, subjectID: subject.ID, globalRole: auth.RoleEditor,
			wantVis: true, wantRole: store.SubjectRoleLead, wantSource: store.AccessSourceDirect,
		},
		{
			name: "team viewer", ctx: ctx, userID: viaTeam.ID, subjectID: subject.ID, globalRole: auth.RoleEditor,
			wantVis: true, wantRole: store.SubjectRoleViewer,
		},
		{
			name: "direct+team max contributor", ctx: ctx, userID: both.ID, subjectID: subject.ID, globalRole: auth.RoleEditor,
			wantVis: true, wantRole: store.SubjectRoleContributor, wantSource: store.AccessSourceDirect,
		},
		{
			name: "org member without grant on gated subject", ctx: ctx, userID: outsider.ID, subjectID: subject.ID, globalRole: auth.RoleEditor,
			wantVis: false,
		},
		{
			name: "org member legacy ungated subject", ctx: ctx, userID: outsider.ID, subjectID: ungated.ID, globalRole: auth.RoleEditor,
			wantVis: true, wantRole: store.SubjectRoleContributor, wantSource: store.AccessSourceOrgMemberLegacy,
		},
		{
			name: "cross org subject", ctx: ctx, userID: lead.ID, subjectID: otherSubject.ID, globalRole: auth.RoleEditor,
			wantVis: false,
		},
		{
			name: "missing subject", ctx: ctx, userID: lead.ID, subjectID: 99999, globalRole: auth.RoleEditor,
			wantVis: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, resolveErr := st.ResolveSubjectAccess(tt.ctx, tt.userID, tt.subjectID, tt.globalRole)
			if resolveErr != nil {
				t.Fatalf("ResolveSubjectAccess(): %v", resolveErr)
			}
			if got.Visible != tt.wantVis {
				t.Fatalf("Visible = %v, want %v (got=%+v)", got.Visible, tt.wantVis, got)
			}
			if tt.wantRole != "" && got.Role != tt.wantRole {
				t.Fatalf("Role = %q, want %q", got.Role, tt.wantRole)
			}
			if !tt.wantVis && got.Role != "" {
				t.Fatalf("hidden access must have empty role, got %q", got.Role)
			}
			if tt.wantSource != "" && !slices.Contains(got.Sources, tt.wantSource) {
				t.Fatalf("Sources = %v, want contain %q", got.Sources, tt.wantSource)
			}
			if tt.wantAlsoDirect && !slices.Contains(got.Sources, store.AccessSourceDirect) {
				t.Fatalf("Sources = %v, want contain %q", got.Sources, store.AccessSourceDirect)
			}
		})
	}

	// Org admin without subject grant: visible, empty role (no implicit lead).
	gatedOnly, err := st.CreateSubject(ctx, "GatedOnly", "", lead.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = st.UpsertDirectSubjectMember(ctx, gatedOnly.ID, lead.ID, store.SubjectRoleLead); err != nil {
		t.Fatal(err)
	}
	got, err := st.ResolveSubjectAccess(ctx, orgAdmin.ID, gatedOnly.ID, auth.RoleEditor)
	if err != nil {
		t.Fatalf("ResolveSubjectAccess(org admin gated): %v", err)
	}
	if !got.Visible || got.Role != "" || !got.HasSource(store.AccessSourceOrgAdmin) {
		t.Fatalf("org admin gated access = %+v, want Visible Role=\"\" org_admin", got)
	}
}

func TestResolveSubjectAccess_PrivateNoLegacy(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	orgAdmin, err := st.UpsertGitHubUser(ctx, 1, "orgadmin", "orgadmin@example.com", "OrgAdmin", "", auth.RoleEditor)
	if err != nil {
		t.Fatal(err)
	}
	member, err := st.UpsertGitHubUser(ctx, 2, "member", "member@example.com", "Member", "", auth.RoleEditor)
	if err != nil {
		t.Fatal(err)
	}
	lead, err := st.UpsertGitHubUser(ctx, 3, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatal(err)
	}
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range []struct {
		id   int64
		role string
	}{
		{orgAdmin.ID, store.OrgRoleAdmin},
		{member.ID, store.OrgRoleMember},
		{lead.ID, store.OrgRoleMember},
	} {
		if err = st.AddOrganizationMember(ctx, defaultOrg.ID, m.id, m.role); err != nil {
			t.Fatal(err)
		}
	}

	private, err := st.CreateSubjectWithVisibility(ctx, "Private", "", lead.ID, nil, store.SubjectVisibilityPrivate)
	if err != nil {
		t.Fatal(err)
	}
	if err = st.RemoveDirectSubjectMember(ctx, private.ID, lead.ID); err != nil {
		t.Fatal(err)
	}

	got, err := st.ResolveSubjectAccess(ctx, member.ID, private.ID, auth.RoleEditor)
	if err != nil {
		t.Fatalf("ResolveSubjectAccess(member): %v", err)
	}
	if got.Visible {
		t.Fatalf("private without grants must not use legacy; got %+v", got)
	}

	got, err = st.ResolveSubjectAccess(ctx, orgAdmin.ID, private.ID, auth.RoleEditor)
	if err != nil {
		t.Fatalf("ResolveSubjectAccess(orgAdmin): %v", err)
	}
	if !got.Visible || !got.HasSource(store.AccessSourceOrgAdmin) {
		t.Fatalf("org admin must see private; got %+v", got)
	}

	if err = st.UpsertDirectSubjectMember(ctx, private.ID, member.ID, store.SubjectRoleViewer); err != nil {
		t.Fatal(err)
	}
	got, err = st.ResolveSubjectAccess(ctx, member.ID, private.ID, auth.RoleEditor)
	if err != nil {
		t.Fatalf("ResolveSubjectAccess(member granted): %v", err)
	}
	if !got.Visible || got.Role != store.SubjectRoleViewer {
		t.Fatalf("granted member must see private; got %+v", got)
	}
}

func TestListSubjects_PrivateFilter(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	owner, err := st.UpsertGitHubUser(ctx, 1, "owner", "owner@example.com", "Owner", "", auth.RoleEditor)
	if err != nil {
		t.Fatal(err)
	}
	member, err := st.UpsertGitHubUser(ctx, 2, "member", "member@example.com", "Member", "", auth.RoleEditor)
	if err != nil {
		t.Fatal(err)
	}
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatal(err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatal(err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, member.ID, store.OrgRoleMember); err != nil {
		t.Fatal(err)
	}

	normal, err := st.CreateSubject(ctx, "NormalOpen", "", owner.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	private, err := st.CreateSubjectWithVisibility(ctx, "PrivateClosed", "", owner.ID, nil, store.SubjectVisibilityPrivate)
	if err != nil {
		t.Fatal(err)
	}
	if private.Visibility != store.SubjectVisibilityPrivate {
		t.Fatalf("Visibility = %q, want private", private.Visibility)
	}

	listed, err := st.ListSubjects(ctx, member.ID, false, "")
	if err != nil {
		t.Fatalf("ListSubjects(member): %v", err)
	}
	var sawNormal, sawPrivate bool
	for _, s := range listed {
		if s.ID == normal.ID {
			sawNormal = true
		}
		if s.ID == private.ID {
			sawPrivate = true
		}
	}
	if !sawNormal {
		t.Fatal("member must see normal ungated subject")
	}
	if sawPrivate {
		t.Fatal("member must not see private subject without grant")
	}

	ownerListed, err := st.ListSubjects(ctx, owner.ID, false, "")
	if err != nil {
		t.Fatalf("ListSubjects(owner): %v", err)
	}
	sawPrivate = false
	for _, s := range ownerListed {
		if s.ID == private.ID {
			sawPrivate = true
			break
		}
	}
	if !sawPrivate {
		t.Fatal("org owner must see private subject")
	}
}
