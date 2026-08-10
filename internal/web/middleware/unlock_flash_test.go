package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
)

func TestIsWhitelistOrgUnlock(t *testing.T) {
	ctx := context.Background()
	db := openOrgAdminTestDB(t)
	st := store.New(db)

	owner, err := st.UpsertGitHubUser(ctx, 20, "solo", "solo@example.com", "Solo", "", auth.RoleEditor)
	if err != nil {
		t.Fatal(err)
	}
	org, err := st.CreateOrganization(ctx, "Solo Org", "solo-org", owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.AddOrganizationMember(ctx, org.ID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatal(err)
	}
	octx := orgctx.WithOrganizationID(ctx, org.ID)
	if err := st.InsertAllowedEmail(octx, "solo@example.com", auth.RoleEditor); err != nil {
		t.Fatal(err)
	}

	reqCtx := testOrgNavCtx(t, st, org, owner)
	hd := HeaderData{
		ShowOrganisationNav: false,
		UserOrganizations:   []store.OrganizationMembership{{Organization: *org, Role: store.OrgRoleOwner}},
	}
	if isWhitelistOrgUnlock(reqCtx, st, owner, hd) {
		t.Fatal("solo with one whitelist email must not unlock org tab")
	}

	if err := st.InsertAllowedEmail(octx, "colleague@example.com", auth.RoleReader); err != nil {
		t.Fatal(err)
	}
	hd.ShowOrganisationNav = showOrganisationNav(reqCtx, st, owner, hd)
	if !hd.ShowOrganisationNav {
		t.Fatal("second whitelist email must show Organisation nav")
	}
	if !isWhitelistOrgUnlock(reqCtx, st, owner, hd) {
		t.Fatal("second whitelist email (still solo member) must be whitelist→org unlock")
	}
}

func TestResolveUnlockFlash_WhitelistOrg(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	msg := resolveUnlockFlash(rec, req, unlockFlashInput{whitelistOrgUnlock: true})
	want := "Un second e-mail a été autorisé : l'onglet Organisation est maintenant disponible."
	if msg != want {
		t.Fatalf("msg = %q, want %q", msg, want)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != unlockCookieName || cookies[0].Value != "org" {
		t.Fatalf("cookie = %+v, want %s=org", cookies, unlockCookieName)
	}

	// One-shot: already seen org level → no flash.
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.AddCookie(&http.Cookie{Name: unlockCookieName, Value: "org"})
	rec2 := httptest.NewRecorder()
	if got := resolveUnlockFlash(rec2, req2, unlockFlashInput{whitelistOrgUnlock: true}); got != "" {
		t.Fatalf("expected empty flash when org already seen, got %q", got)
	}
}

func TestResolveUnlockFlash_MemberUnlockTakesPriority(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	msg := resolveUnlockFlash(rec, req, unlockFlashInput{
		caps:               UICaps{ShowAssign: true},
		whitelistOrgUnlock: true,
	})
	if msg == "" || msg == "Un second e-mail a été autorisé : l'onglet Organisation est maintenant disponible." {
		t.Fatalf("P1 member unlock should win over whitelist→org, got %q", msg)
	}
}

func TestIsWhitelistOrgUnlock_GlobalAdminNever(t *testing.T) {
	ctx := context.Background()
	db := openOrgAdminTestDB(t)
	st := store.New(db)

	admin, err := st.UpsertGitHubUser(ctx, 21, "admin", "admin@example.com", "Admin", "", auth.RoleAdmin)
	if err != nil {
		t.Fatal(err)
	}
	org, err := st.CreateOrganization(ctx, "Admin Solo", "admin-solo", admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.AddOrganizationMember(ctx, org.ID, admin.ID, store.OrgRoleOwner); err != nil {
		t.Fatal(err)
	}
	octx := orgctx.WithOrganizationID(ctx, org.ID)
	if err := st.InsertAllowedEmail(octx, "admin@example.com", auth.RoleAdmin); err != nil {
		t.Fatal(err)
	}
	if err := st.InsertAllowedEmail(octx, "other@example.com", auth.RoleReader); err != nil {
		t.Fatal(err)
	}

	reqCtx := testOrgNavCtx(t, st, org, admin)
	hd := HeaderData{ShowOrganisationNav: true}
	if isWhitelistOrgUnlock(reqCtx, st, admin, hd) {
		t.Fatal("global admin must not get whitelist→org unlock flash")
	}
}
