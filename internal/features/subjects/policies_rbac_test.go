package subjects_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
)

func TestRBAC_OrgLeadPolicies(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}

	owner, err := st.UpsertGitHubUser(ctx, 900, "pol-owner", "pol-owner@example.com", "Owner", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(owner): %v", err)
	}
	lead, err := st.UpsertGitHubUser(ctx, 901, "pol-lead", "pol-lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(lead): %v", err)
	}
	member, err := st.UpsertGitHubUser(ctx, 902, "pol-member", "pol-member@example.com", "Member", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(member): %v", err)
	}
	external, err := st.UpsertGitHubUser(ctx, 903, "pol-ext", "pol-ext@example.com", "External", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(external): %v", err)
	}

	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(owner): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, lead.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(lead): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, member.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(member): %v", err)
	}

	subject, err := st.CreateSubject(ctx, "Sujet politiques", "", owner.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
	}
	if err = st.UpsertDirectSubjectMember(ctx, subject.ID, lead.ID, store.SubjectRoleLead); err != nil {
		t.Fatalf("UpsertDirectSubjectMember(lead): %v", err)
	}

	team, err := st.CreateTeam(ctx, "Policy Team", "policy-team", "")
	if err != nil {
		t.Fatalf("CreateTeam(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	ownerToken, _, err := sessions.CreateLoginSession(ctx, owner.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(owner): %v", err)
	}
	leadToken, _, err := sessions.CreateLoginSession(ctx, lead.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(lead): %v", err)
	}
	ownerCSRF := auth.CSRFToken(ownerToken, "test-secret-at-least-thirty-two-bytes")
	leadCSRF := auth.CSRFToken(leadToken, "test-secret-at-least-thirty-two-bytes")
	subjectPath := "/subjects/" + strconv.FormatInt(subject.ID, 10)

	// Deny leads_may_assign_teams: lead POST denied with UI message; owner still OK.
	if err = st.UpdateOrganizationLeadPolicies(ctx, defaultOrg.ID, store.OrgLeadPolicies{
		LeadsMayAssignTeams:     false,
		LeadsMayInviteMembers:   true,
		LeadsMayInviteExternals: false,
	}); err != nil {
		t.Fatalf("UpdateOrganizationLeadPolicies(assign=false): %v", err)
	}

	showReq := httptest.NewRequest(http.MethodGet, subjectPath, nil)
	showReq.AddCookie(&http.Cookie{Name: "revues_session", Value: leadToken})
	showRec := httptest.NewRecorder()
	handler.ServeHTTP(showRec, showReq)
	if showRec.Code != http.StatusOK {
		t.Fatalf("show status = %d", showRec.Code)
	}
	if !strings.Contains(showRec.Body.String(), "n'autorise pas les leads à affecter des équipes") {
		t.Fatalf("show missing teams policy denial message")
	}

	leadTeamForm := url.Values{}
	leadTeamForm.Set("csrf_token", leadCSRF)
	leadTeamForm.Set("team_id", strconv.FormatInt(team.ID, 10))
	leadTeamForm.Set("role", "viewer")
	leadTeamReq := httptest.NewRequest(http.MethodPost, subjectPath+"/teams", strings.NewReader(leadTeamForm.Encode()))
	leadTeamReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	leadTeamReq.AddCookie(&http.Cookie{Name: "revues_session", Value: leadToken})
	leadTeamRec := httptest.NewRecorder()
	handler.ServeHTTP(leadTeamRec, leadTeamReq)
	if leadTeamRec.Code != http.StatusBadRequest {
		t.Fatalf("lead denied assign status = %d, want 400 with error message", leadTeamRec.Code)
	}
	if !strings.Contains(leadTeamRec.Body.String(), "autorise pas les leads") ||
		!strings.Contains(leadTeamRec.Body.String(), "affecter des équipes") {
		t.Fatalf("lead denied assign body missing policy message")
	}
	teams, err := st.ListSubjectTeams(ctx, subject.ID)
	if err != nil || len(teams) != 0 {
		t.Fatalf("teams after denied lead assign = %+v, %v", teams, err)
	}

	ownerTeamForm := url.Values{}
	ownerTeamForm.Set("csrf_token", ownerCSRF)
	ownerTeamForm.Set("team_id", strconv.FormatInt(team.ID, 10))
	ownerTeamForm.Set("role", "viewer")
	ownerTeamReq := httptest.NewRequest(http.MethodPost, subjectPath+"/teams", strings.NewReader(ownerTeamForm.Encode()))
	ownerTeamReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ownerTeamReq.AddCookie(&http.Cookie{Name: "revues_session", Value: ownerToken})
	ownerTeamRec := httptest.NewRecorder()
	handler.ServeHTTP(ownerTeamRec, ownerTeamReq)
	if ownerTeamRec.Code != http.StatusSeeOther {
		t.Fatalf("owner assign status = %d, want %d", ownerTeamRec.Code, http.StatusSeeOther)
	}

	// Deny leads_may_invite_members (externals still on so the invite form is available):
	// lead cannot invite an org member.
	if err = st.UpdateOrganizationLeadPolicies(ctx, defaultOrg.ID, store.OrgLeadPolicies{
		LeadsMayAssignTeams:     true,
		LeadsMayInviteMembers:   false,
		LeadsMayInviteExternals: true,
	}); err != nil {
		t.Fatalf("UpdateOrganizationLeadPolicies(invite=false): %v", err)
	}

	leadInviteForm := url.Values{}
	leadInviteForm.Set("csrf_token", leadCSRF)
	leadInviteForm.Set("email", member.Email)
	leadInviteForm.Set("role", "contributor")
	leadInviteReq := httptest.NewRequest(http.MethodPost, subjectPath+"/members", strings.NewReader(leadInviteForm.Encode()))
	leadInviteReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	leadInviteReq.AddCookie(&http.Cookie{Name: "revues_session", Value: leadToken})
	leadInviteRec := httptest.NewRecorder()
	handler.ServeHTTP(leadInviteRec, leadInviteReq)
	if leadInviteRec.Code != http.StatusBadRequest {
		t.Fatalf("lead denied invite member status = %d, want 400", leadInviteRec.Code)
	}
	if !strings.Contains(leadInviteRec.Body.String(), "autorise pas les leads") ||
		!strings.Contains(leadInviteRec.Body.String(), "inviter des membres") {
		t.Fatalf("lead denied invite body missing policy message: %s", leadInviteRec.Body.String())
	}

	ownerInviteForm := url.Values{}
	ownerInviteForm.Set("csrf_token", ownerCSRF)
	ownerInviteForm.Set("email", member.Email)
	ownerInviteForm.Set("role", "contributor")
	ownerInviteReq := httptest.NewRequest(http.MethodPost, subjectPath+"/members", strings.NewReader(ownerInviteForm.Encode()))
	ownerInviteReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ownerInviteReq.AddCookie(&http.Cookie{Name: "revues_session", Value: ownerToken})
	ownerInviteRec := httptest.NewRecorder()
	handler.ServeHTTP(ownerInviteRec, ownerInviteReq)
	if ownerInviteRec.Code != http.StatusSeeOther {
		t.Fatalf("owner invite member status = %d, want %d", ownerInviteRec.Code, http.StatusSeeOther)
	}

	// Deny leads_may_invite_externals (default false): lead cannot invite non-org user.
	if err = st.UpdateOrganizationLeadPolicies(ctx, defaultOrg.ID, store.OrgLeadPolicies{
		LeadsMayAssignTeams:     true,
		LeadsMayInviteMembers:   true,
		LeadsMayInviteExternals: false,
	}); err != nil {
		t.Fatalf("UpdateOrganizationLeadPolicies(externals=false): %v", err)
	}

	leadExtForm := url.Values{}
	leadExtForm.Set("csrf_token", leadCSRF)
	leadExtForm.Set("email", external.Email)
	leadExtForm.Set("role", "viewer")
	leadExtReq := httptest.NewRequest(http.MethodPost, subjectPath+"/members", strings.NewReader(leadExtForm.Encode()))
	leadExtReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	leadExtReq.AddCookie(&http.Cookie{Name: "revues_session", Value: leadToken})
	leadExtRec := httptest.NewRecorder()
	handler.ServeHTTP(leadExtRec, leadExtReq)
	if leadExtRec.Code != http.StatusBadRequest {
		t.Fatalf("lead denied invite external status = %d, want 400", leadExtRec.Code)
	}
	if !strings.Contains(leadExtRec.Body.String(), "autorise pas les leads") ||
		!strings.Contains(leadExtRec.Body.String(), "inviter des externes") {
		t.Fatalf("lead denied external body missing policy message: %s", leadExtRec.Body.String())
	}

	// Allow externals: lead can invite non-org user.
	if err = st.UpdateOrganizationLeadPolicies(ctx, defaultOrg.ID, store.OrgLeadPolicies{
		LeadsMayAssignTeams:     true,
		LeadsMayInviteMembers:   true,
		LeadsMayInviteExternals: true,
	}); err != nil {
		t.Fatalf("UpdateOrganizationLeadPolicies(externals=true): %v", err)
	}
	leadExtOK := httptest.NewRequest(http.MethodPost, subjectPath+"/members", strings.NewReader(leadExtForm.Encode()))
	leadExtOK.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	leadExtOK.AddCookie(&http.Cookie{Name: "revues_session", Value: leadToken})
	leadExtOKRec := httptest.NewRecorder()
	handler.ServeHTTP(leadExtOKRec, leadExtOK)
	if leadExtOKRec.Code != http.StatusSeeOther {
		t.Fatalf("lead invite external when allowed status = %d, want %d; body=%s",
			leadExtOKRec.Code, http.StatusSeeOther, leadExtOKRec.Body.String())
	}
}
