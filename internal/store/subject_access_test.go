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

	tests := []struct {
		name       string
		ctx        context.Context
		userID     int64
		subjectID  int64
		globalRole string
		wantVis    bool
		wantRole   string
		wantSource string
	}{
		{
			name: "global admin", ctx: ctx, userID: admin.ID, subjectID: subject.ID, globalRole: auth.RoleAdmin,
			wantVis: true, wantSource: store.AccessSourceGlobalAdmin,
		},
		{
			name: "org admin", ctx: ctx, userID: orgAdmin.ID, subjectID: subject.ID, globalRole: auth.RoleEditor,
			wantVis: true, wantSource: store.AccessSourceOrgAdmin,
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
			name: "org member without grant", ctx: ctx, userID: outsider.ID, subjectID: subject.ID, globalRole: auth.RoleEditor,
			wantVis: false,
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
			got, err := st.ResolveSubjectAccess(tt.ctx, tt.userID, tt.subjectID, tt.globalRole)
			if err != nil {
				t.Fatalf("ResolveSubjectAccess(): %v", err)
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
		})
	}
}
