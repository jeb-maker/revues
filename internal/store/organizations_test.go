package store_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

func TestNormalizeOrganizationSlug(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{name: "lowercase trim", input: "  Acme Corp  ", want: "acme-corp"},
		{name: "underscores", input: "my_org", want: "my-org"},
		{name: "already valid", input: "acme-42", want: "acme-42"},
		{name: "french accents", input: "Squad Qualité", want: "squad-qualite"},
		{name: "cedilla and grave", input: "École Française", want: "ecole-francaise"},
		{name: "ligature oe", input: "Cœur", want: "coeur"},
		{name: "empty", input: "   ", wantErr: store.ErrInvalidOrganizationSlug},
		{name: "invalid only symbols", input: "!!!", wantErr: store.ErrInvalidOrganizationSlug},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.NormalizeOrganizationSlug(tt.input)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("NormalizeOrganizationSlug(%q) error = %v, want %v", tt.input, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeOrganizationSlug(%q) error = %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("NormalizeOrganizationSlug(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCreateOrganization(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)

	creator, err := st.UpsertGitHubUser(ctx, 1, "owner", "owner@example.com", "Owner", "", auth.RoleAdmin)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	tests := []struct {
		name    string
		orgName string
		slug    string
		wantErr error
	}{
		{name: "creates with normalized slug", orgName: "Acme Corp", slug: "Acme-Corp"},
		{name: "duplicate slug", orgName: "Other", slug: "acme-corp", wantErr: store.ErrOrganizationSlugTaken},
	}

	var firstID int64
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org, err := st.CreateOrganization(ctx, tt.orgName, tt.slug, creator.ID)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("CreateOrganization() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("CreateOrganization() error = %v", err)
			}
			if org.Slug != "acme-corp" {
				t.Fatalf("slug = %q, want acme-corp", org.Slug)
			}
			if !org.CreatedBy.Valid || org.CreatedBy.Int64 != creator.ID {
				t.Fatalf("created_by = %v, want %d", org.CreatedBy, creator.ID)
			}

			bySlug, err := st.OrganizationBySlug(ctx, "ACME-CORP")
			if err != nil || bySlug.ID != org.ID {
				t.Fatalf("OrganizationBySlug() = %v, %v", bySlug, err)
			}
			byID, err := st.OrganizationByID(ctx, org.ID)
			if err != nil || byID.Slug != org.Slug {
				t.Fatalf("OrganizationByID() = %v, %v", byID, err)
			}
			firstID = org.ID
		})
	}

	if firstID == 0 {
		t.Fatal("expected first organization id")
	}
}

func TestOrganizationMemberships(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)

	owner, err := st.UpsertGitHubUser(ctx, 1, "owner", "owner@example.com", "Owner", "", auth.RoleAdmin)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(owner): %v", err)
	}
	member, err := st.UpsertGitHubUser(ctx, 2, "member", "member@example.com", "Member", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(member): %v", err)
	}

	org, err := st.CreateOrganization(ctx, "Team", "team", owner.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(): %v", err)
	}

	if err = st.AddOrganizationMember(ctx, org.ID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(owner): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, org.ID, member.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(member): %v", err)
	}

	role, ok, err := st.OrganizationMemberRole(ctx, org.ID, member.ID)
	if err != nil || !ok || role != store.OrgRoleMember {
		t.Fatalf("OrganizationMemberRole() = %q, %v, %v", role, ok, err)
	}

	count, err := st.CountUserOrganizations(ctx, member.ID)
	if err != nil || count != 1 {
		t.Fatalf("CountUserOrganizations() = %d, %v", count, err)
	}

	memberships, err := st.ListUserOrganizations(ctx, member.ID)
	if err != nil {
		t.Fatalf("ListUserOrganizations(): %v", err)
	}
	if len(memberships) != 1 || memberships[0].Organization.Slug != "team" || memberships[0].Role != store.OrgRoleMember {
		t.Fatalf("ListUserOrganizations() = %+v", memberships)
	}

	if err = st.RemoveOrganizationMember(ctx, org.ID, member.ID); err != nil {
		t.Fatalf("RemoveOrganizationMember(): %v", err)
	}
	if err = st.RemoveOrganizationMember(ctx, org.ID, member.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("RemoveOrganizationMember() missing error = %v", err)
	}

	count, err = st.CountUserOrganizations(ctx, member.ID)
	if err != nil || count != 0 {
		t.Fatalf("CountUserOrganizations() after remove = %d, %v", count, err)
	}
}

func TestDefaultOrganizationExistsAfterMigrate(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)

	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(default): %v", err)
	}
	if defaultOrg.Name != "Default" {
		t.Fatalf("default org name = %q, want Default", defaultOrg.Name)
	}
	if defaultOrg.UISubjectLabel != store.UISubjectLabelSujet {
		t.Fatalf("default ui_subject_label = %q, want %q", defaultOrg.UISubjectLabel, store.UISubjectLabelSujet)
	}
	if !defaultOrg.LeadsMayAssignTeams || !defaultOrg.LeadsMayInviteMembers || defaultOrg.LeadsMayInviteExternals {
		t.Fatalf("default lead policies = assign=%v invite=%v externals=%v, want true/true/false",
			defaultOrg.LeadsMayAssignTeams, defaultOrg.LeadsMayInviteMembers, defaultOrg.LeadsMayInviteExternals)
	}
}

func TestUpdateOrganizationLeadPolicies(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)

	creator, err := st.UpsertGitHubUser(ctx, 1, "owner", "owner@example.com", "Owner", "", auth.RoleAdmin)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	org, err := st.CreateOrganization(ctx, "Acme", "acme", creator.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(): %v", err)
	}
	if !org.LeadsMayAssignTeams || !org.LeadsMayInviteMembers || org.LeadsMayInviteExternals {
		t.Fatalf("create defaults = assign=%v invite=%v externals=%v",
			org.LeadsMayAssignTeams, org.LeadsMayInviteMembers, org.LeadsMayInviteExternals)
	}

	if err = st.UpdateOrganizationLeadPolicies(ctx, org.ID, store.OrgLeadPolicies{
		LeadsMayAssignTeams:     false,
		LeadsMayInviteMembers:   false,
		LeadsMayInviteExternals: true,
	}); err != nil {
		t.Fatalf("UpdateOrganizationLeadPolicies(): %v", err)
	}
	got, err := st.OrganizationByID(ctx, org.ID)
	if err != nil {
		t.Fatalf("OrganizationByID(): %v", err)
	}
	if got.LeadsMayAssignTeams || got.LeadsMayInviteMembers || !got.LeadsMayInviteExternals {
		t.Fatalf("updated policies = assign=%v invite=%v externals=%v, want false/false/true",
			got.LeadsMayAssignTeams, got.LeadsMayInviteMembers, got.LeadsMayInviteExternals)
	}

	if err = st.UpdateOrganizationLeadPolicies(ctx, 99999, store.DefaultOrgLeadPolicies()); !errors.Is(err, store.ErrOrganizationNotFound) {
		t.Fatalf("missing org error = %v, want %v", err, store.ErrOrganizationNotFound)
	}
}

func TestUpdateOrganizationUISubjectLabel(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)

	creator, err := st.UpsertGitHubUser(ctx, 1, "owner", "owner@example.com", "Owner", "", auth.RoleAdmin)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	org, err := st.CreateOrganization(ctx, "Acme", "acme", creator.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(): %v", err)
	}
	if org.UISubjectLabel != store.UISubjectLabelSujet {
		t.Fatalf("UISubjectLabel = %q, want %q", org.UISubjectLabel, store.UISubjectLabelSujet)
	}

	if err = st.UpdateOrganizationUISubjectLabel(ctx, org.ID, store.UISubjectLabelCible); err != nil {
		t.Fatalf("UpdateOrganizationUISubjectLabel(): %v", err)
	}
	got, err := st.OrganizationByID(ctx, org.ID)
	if err != nil {
		t.Fatalf("OrganizationByID(): %v", err)
	}
	if got.UISubjectLabel != store.UISubjectLabelCible {
		t.Fatalf("UISubjectLabel = %q, want %q", got.UISubjectLabel, store.UISubjectLabelCible)
	}

	if err = st.UpdateOrganizationUISubjectLabel(ctx, org.ID, "inconnu"); !errors.Is(err, store.ErrInvalidUISubjectLabel) {
		t.Fatalf("invalid label error = %v, want %v", err, store.ErrInvalidUISubjectLabel)
	}
}
