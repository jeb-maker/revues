package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
)

func TestNormalizeTeamSlug(t *testing.T) {
	got, err := store.NormalizeTeamSlug("  QA Squad  ")
	if err != nil {
		t.Fatalf("NormalizeTeamSlug(): %v", err)
	}
	if got != "qa-squad" {
		t.Fatalf("slug = %q, want qa-squad", got)
	}
}

func TestTeamsStore(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	owner, err := st.UpsertGitHubUser(ctx, 1, "owner", "owner@example.com", "Owner", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(owner): %v", err)
	}
	alice, err := st.UpsertGitHubUser(ctx, 2, "alice", "alice@example.com", "Alice", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(alice): %v", err)
	}
	bob, err := st.UpsertGitHubUser(ctx, 3, "bob", "bob@example.com", "Bob", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(bob): %v", err)
	}

	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}
	for _, uid := range []int64{owner.ID, alice.ID, bob.ID} {
		if err = st.AddOrganizationMember(ctx, defaultOrg.ID, uid, store.OrgRoleMember); err != nil {
			t.Fatalf("AddOrganizationMember(%d): %v", uid, err)
		}
	}

	subject, err := st.CreateSubject(ctx, "Portail", "", owner.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
	}

	t.Run("create and list teams", func(t *testing.T) {
		created, createErr := st.CreateTeam(ctx, "Qualité", "Qualite", "revue QA")
		if createErr != nil {
			t.Fatalf("CreateTeam(): %v", createErr)
		}
		if created.Slug != "qualite" {
			t.Fatalf("slug = %q, want qualite", created.Slug)
		}
		if _, dupErr := st.CreateTeam(ctx, "Dup", "qualite", ""); !errors.Is(dupErr, store.ErrTeamSlugTaken) {
			t.Fatalf("duplicate slug error = %v, want ErrTeamSlugTaken", dupErr)
		}
		teams, listErr := st.ListOrganizationTeams(ctx)
		if listErr != nil {
			t.Fatalf("ListOrganizationTeams(): %v", listErr)
		}
		if len(teams) != 1 || teams[0].ID != created.ID {
			t.Fatalf("ListOrganizationTeams() = %+v", teams)
		}
	})

	team, err := st.CreateTeam(ctx, "Ops", "ops", "")
	if err != nil {
		t.Fatalf("CreateTeam(ops): %v", err)
	}

	t.Run("team members", func(t *testing.T) {
		if addErr := st.AddTeamMember(ctx, team.ID, alice.ID); addErr != nil {
			t.Fatalf("AddTeamMember(alice): %v", addErr)
		}
		if addErr := st.AddTeamMember(ctx, team.ID, bob.ID); addErr != nil {
			t.Fatalf("AddTeamMember(bob): %v", addErr)
		}
		members, listErr := st.ListTeamMembers(ctx, team.ID)
		if listErr != nil {
			t.Fatalf("ListTeamMembers(): %v", listErr)
		}
		if len(members) != 2 {
			t.Fatalf("len(members) = %d, want 2", len(members))
		}
		if members[0].Login == "" || members[0].Email == "" {
			t.Fatalf("ListTeamMembers missing login/email: %+v", members[0])
		}
		orgMembers, listOrgErr := st.ListOrganizationMembers(ctx)
		if listOrgErr != nil {
			t.Fatalf("ListOrganizationMembers(): %v", listOrgErr)
		}
		if len(orgMembers) < 3 {
			t.Fatalf("ListOrganizationMembers len = %d, want >= 3", len(orgMembers))
		}
		userTeams, listUserErr := st.ListUserTeams(ctx, alice.ID)
		if listUserErr != nil {
			t.Fatalf("ListUserTeams(): %v", listUserErr)
		}
		if len(userTeams) != 1 || userTeams[0].ID != team.ID {
			t.Fatalf("ListUserTeams() = %+v", userTeams)
		}
		if removeErr := st.RemoveTeamMember(ctx, team.ID, bob.ID); removeErr != nil {
			t.Fatalf("RemoveTeamMember(bob): %v", removeErr)
		}
		members, listErr = st.ListTeamMembers(ctx, team.ID)
		if listErr != nil {
			t.Fatalf("ListTeamMembers after remove: %v", listErr)
		}
		if len(members) != 1 || members[0].UserID != alice.ID {
			t.Fatalf("members after remove = %+v", members)
		}
	})

	t.Run("direct subject members", func(t *testing.T) {
		if err := st.UpsertDirectSubjectMember(ctx, subject.ID, bob.ID, store.SubjectRoleContributor); err != nil {
			t.Fatalf("UpsertDirectSubjectMember(): %v", err)
		}
		if err := st.UpsertDirectSubjectMember(ctx, subject.ID, bob.ID, store.SubjectRoleLead); err != nil {
			t.Fatalf("UpsertDirectSubjectMember(update): %v", err)
		}
		direct, err := st.ListDirectSubjectMembers(ctx, subject.ID)
		if err != nil {
			t.Fatalf("ListDirectSubjectMembers(): %v", err)
		}
		if len(direct) != 1 || direct[0].Role != store.SubjectRoleLead {
			t.Fatalf("direct = %+v", direct)
		}
		if err := st.RemoveDirectSubjectMember(ctx, subject.ID, bob.ID); err != nil {
			t.Fatalf("RemoveDirectSubjectMember(): %v", err)
		}
	})

	t.Run("team subject roles", func(t *testing.T) {
		if err := st.GrantTeamSubjectRole(ctx, team.ID, subject.ID, store.SubjectRoleViewer, owner.ID); err != nil {
			t.Fatalf("GrantTeamSubjectRole(): %v", err)
		}
		if err := st.GrantTeamSubjectRole(ctx, team.ID, subject.ID, store.SubjectRoleContributor, owner.ID); err != nil {
			t.Fatalf("GrantTeamSubjectRole(update): %v", err)
		}
		subjectTeams, listErr := st.ListSubjectTeams(ctx, subject.ID)
		if listErr != nil {
			t.Fatalf("ListSubjectTeams(): %v", listErr)
		}
		if len(subjectTeams) != 1 || subjectTeams[0].Role != store.SubjectRoleContributor {
			t.Fatalf("ListSubjectTeams() = %+v", subjectTeams)
		}
		teamSubjects, listTeamErr := st.ListTeamSubjects(ctx, team.ID)
		if listTeamErr != nil {
			t.Fatalf("ListTeamSubjects(): %v", listTeamErr)
		}
		if len(teamSubjects) != 1 || teamSubjects[0].SubjectID != subject.ID {
			t.Fatalf("ListTeamSubjects() = %+v", teamSubjects)
		}
		if revokeErr := st.RevokeTeamSubjectRole(ctx, team.ID, subject.ID); revokeErr != nil {
			t.Fatalf("RevokeTeamSubjectRole(): %v", revokeErr)
		}
		subjectTeams, listErr = st.ListSubjectTeams(ctx, subject.ID)
		if listErr != nil {
			t.Fatalf("ListSubjectTeams after revoke: %v", listErr)
		}
		if len(subjectTeams) != 0 {
			t.Fatalf("expected empty after revoke, got %+v", subjectTeams)
		}
	})

	t.Run("org isolation", func(t *testing.T) {
		other, err := st.CreateOrganization(ctx, "Other", "other-teams", owner.ID)
		if err != nil {
			t.Fatalf("CreateOrganization(other): %v", err)
		}
		if err := st.AddOrganizationMember(ctx, other.ID, owner.ID, store.OrgRoleOwner); err != nil {
			t.Fatalf("AddOrganizationMember(other): %v", err)
		}
		otherCtx := orgctx.WithOrganizationID(ctx, other.ID)
		if _, err := st.TeamByID(otherCtx, team.ID); !errors.Is(err, store.ErrTeamNotFound) {
			t.Fatalf("TeamByID cross-org = %v, want ErrTeamNotFound", err)
		}
		if _, err := st.CreateTeam(otherCtx, "Ops", "ops", ""); err != nil {
			t.Fatalf("CreateTeam in other org: %v", err)
		}
	})
}
