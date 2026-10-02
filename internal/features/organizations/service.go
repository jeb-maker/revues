package organizations

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

const spaHomePath = "/"

// Service holds organization onboarding / selection business logic.
type Service struct {
	Store         OrgStore
	Sessions      *auth.SessionManager
	SecureCookies bool
}

// ListResult is returned by List for the authenticated user.
type ListResult struct {
	Memberships           []store.OrganizationMembership
	Invitations           []store.OrganizationInvitation
	ActiveOrganizationID  int64 // 0 when pending / unset
	DefaultOrganizationID int64 // 0 when no remembered org among memberships
	CanCreate             bool
	Redirect              string
}

// ActionResult is returned after create / select / accept.
type ActionResult struct {
	Organization *store.Organization
	Role         string
	Redirect     string
}

var (
	// ErrUnauthenticated is returned when the caller has no session user.
	ErrUnauthenticated = errors.New("unauthenticated")
	// ErrValidation is returned for invalid create/select payloads.
	ErrValidation = errors.New("validation failed")
	// ErrAlreadyHasOrganization is returned when self-service create is blocked.
	ErrAlreadyHasOrganization = errors.New("already has organization")
	// ErrNotFound is returned for missing or non-visible org/invitation (IDOR).
	ErrNotFound = errors.New("not found")
	// ErrForbidden is returned when the action is refused for the caller.
	ErrForbidden = errors.New("forbidden")
)

// PostLoginRoute decides how to seed the session organization and where to send the SPA.
func PostLoginRoute(ctx context.Context, st interface {
	CountUserOrganizations(ctx context.Context, userID int64) (int, error)
	ListUserOrganizations(ctx context.Context, userID int64) ([]store.OrganizationMembership, error)
}, userID int64) (sessionOrgID int64, redirect string, err error) {
	count, err := st.CountUserOrganizations(ctx, userID)
	if err != nil {
		return 0, "", fmt.Errorf("count user organizations: %w", err)
	}

	switch count {
	case 0:
		return auth.SessionOrgPending, "/org/new", nil
	case 1:
		memberships, err := st.ListUserOrganizations(ctx, userID)
		if err != nil {
			return 0, "", fmt.Errorf("list user organizations: %w", err)
		}
		return memberships[0].Organization.ID, spaHomePath, nil
	default:
		return auth.SessionOrgPending, "/org/select", nil
	}
}

// List returns memberships, invitations, and onboarding hints for the user.
func (s *Service) List(ctx context.Context, user *store.User, sessionToken string, r *http.Request) (*ListResult, error) {
	if user == nil {
		return nil, ErrUnauthenticated
	}

	memberships, err := s.Store.ListUserOrganizations(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("list user organizations: %w", err)
	}

	invites, err := s.Store.ListPendingInvitationsByEmail(ctx, user.Email)
	if err != nil {
		return nil, fmt.Errorf("list pending invitations: %w", err)
	}

	var activeID int64
	if sessionToken != "" {
		_, orgID, sessErr := s.Store.SessionByTokenHash(ctx, auth.HashToken(sessionToken))
		if sessErr == nil && orgID > 0 {
			activeID = orgID
		}
	}

	defaultID := defaultOrgID(r, memberships)
	_, redirect, err := PostLoginRoute(ctx, s.Store, user.ID)
	if err != nil {
		return nil, err
	}
	// If session already has an active org among memberships, stay on home.
	if activeID > 0 {
		for _, m := range memberships {
			if m.Organization.ID == activeID {
				redirect = spaHomePath
				break
			}
		}
	}

	return &ListResult{
		Memberships:           memberships,
		Invitations:           invites,
		ActiveOrganizationID:  activeID,
		DefaultOrganizationID: defaultID,
		CanCreate:             len(memberships) == 0,
		Redirect:              redirect,
	}, nil
}

// Create creates the user's first organization and activates it on the session.
func (s *Service) Create(ctx context.Context, user *store.User, sessionToken, name, slug string, w http.ResponseWriter) (*ActionResult, error) {
	if user == nil {
		return nil, ErrUnauthenticated
	}
	if sessionToken == "" {
		return nil, ErrUnauthenticated
	}

	count, err := s.Store.CountUserOrganizations(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("count user organizations: %w", err)
	}
	if count > 0 {
		return nil, ErrAlreadyHasOrganization
	}

	name = strings.TrimSpace(name)
	slug = strings.TrimSpace(slug)
	if name == "" {
		return nil, fmt.Errorf("%w: name required", ErrValidation)
	}
	if slug == "" {
		slug = name
	}

	org, err := s.Store.CreateOrganization(ctx, name, slug, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrInvalidOrganizationSlug):
			return nil, fmt.Errorf("%w: invalid slug", ErrValidation)
		case errors.Is(err, store.ErrOrganizationSlugTaken):
			return nil, fmt.Errorf("%w: slug taken", ErrValidation)
		default:
			return nil, fmt.Errorf("create organization: %w", err)
		}
	}

	if err := s.Store.AddOrganizationMember(ctx, org.ID, user.ID, store.OrgRoleOwner); err != nil {
		return nil, fmt.Errorf("add organization owner: %w", err)
	}

	if err := s.activate(ctx, w, sessionToken, org.ID); err != nil {
		return nil, err
	}

	return &ActionResult{Organization: org, Role: store.OrgRoleOwner, Redirect: spaHomePath}, nil
}

// Select activates an organization the user belongs to.
func (s *Service) Select(ctx context.Context, user *store.User, sessionToken string, organizationID int64, w http.ResponseWriter) (*ActionResult, error) {
	if user == nil || sessionToken == "" {
		return nil, ErrUnauthenticated
	}
	if organizationID <= 0 {
		return nil, fmt.Errorf("%w: organization_id required", ErrValidation)
	}

	role, member, err := s.Store.OrganizationMemberRole(ctx, organizationID, user.ID)
	if err != nil {
		return nil, fmt.Errorf("organization member role: %w", err)
	}
	if !member {
		return nil, ErrNotFound
	}

	org, err := s.Store.OrganizationByID(ctx, organizationID)
	if err != nil {
		if errors.Is(err, store.ErrOrganizationNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("organization by id: %w", err)
	}

	if err := s.activate(ctx, w, sessionToken, org.ID); err != nil {
		return nil, err
	}

	return &ActionResult{Organization: org, Role: role, Redirect: spaHomePath}, nil
}

// AcceptInvitation consumes a pending invite for the user's email and activates the org.
func (s *Service) AcceptInvitation(ctx context.Context, user *store.User, sessionToken string, invitationID int64, w http.ResponseWriter) (*ActionResult, error) {
	if user == nil || sessionToken == "" {
		return nil, ErrUnauthenticated
	}
	if invitationID <= 0 {
		return nil, ErrNotFound
	}

	invite, err := s.Store.OrganizationInvitationByID(ctx, invitationID)
	if errors.Is(err, store.ErrOrganizationInvitationNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("organization invitation by id: %w", err)
	}

	if strings.ToLower(strings.TrimSpace(user.Email)) != invite.Email {
		// Do not reveal invitation existence across emails.
		return nil, ErrNotFound
	}

	role, member, err := s.Store.OrganizationMemberRole(ctx, invite.OrganizationID, user.ID)
	if err != nil {
		return nil, fmt.Errorf("invitation org membership: %w", err)
	}
	if !member {
		if err := s.Store.AddOrganizationMember(ctx, invite.OrganizationID, user.ID, invite.OrgRole); err != nil {
			return nil, fmt.Errorf("accept invitation org member: %w", err)
		}
		role = invite.OrgRole
	}

	if err := s.Store.DeleteOrganizationInvitation(ctx, invite.ID); err != nil {
		return nil, fmt.Errorf("delete organization invitation: %w", err)
	}

	org, err := s.Store.OrganizationByID(ctx, invite.OrganizationID)
	if err != nil {
		if errors.Is(err, store.ErrOrganizationNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("organization by id: %w", err)
	}

	if err := s.activate(ctx, w, sessionToken, org.ID); err != nil {
		return nil, err
	}

	return &ActionResult{Organization: org, Role: role, Redirect: spaHomePath}, nil
}

func (s *Service) activate(ctx context.Context, w http.ResponseWriter, sessionToken string, orgID int64) error {
	if err := s.Sessions.SetActiveOrganization(ctx, sessionToken, orgID); err != nil {
		return fmt.Errorf("set active organization: %w", err)
	}
	auth.SetLastOrgCookie(w, orgID, s.SecureCookies)
	return nil
}

func defaultOrgID(r *http.Request, memberships []store.OrganizationMembership) int64 {
	if r == nil {
		return 0
	}
	lastID := auth.LastOrgIDFromRequest(r)
	if lastID <= 0 {
		return 0
	}
	for _, m := range memberships {
		if m.Organization.ID == lastID {
			return lastID
		}
	}
	return 0
}
