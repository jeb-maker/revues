package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jeb-maker/revues/internal/auth"
)

// ErrEmailNotAllowed is returned when strict login rejects an email
// (no org membership, pending invitation, or bootstrap match).
var ErrEmailNotAllowed = errors.New("email not allowed")

// ResolveLoginRole determines the global role for a verified email at login.
// Login is allowed when the user belongs to at least one org, matches
// REVUES_BOOTSTRAP_ADMIN_EMAIL, has a pending invitation, or (unless
// requireInvite) has no org yet (self-service org creation).
func (s *Store) ResolveLoginRole(ctx context.Context, email, bootstrapAdmin string) (string, error) {
	return s.ResolveLoginRoleStrict(ctx, email, bootstrapAdmin, false)
}

// ResolveLoginRoleStrict is ResolveLoginRole with an optional invite lock.
// When requireInvite is true, unknown emails without org membership, bootstrap
// match, or pending invitation are rejected with ErrEmailNotAllowed (same
// message path for anti-enumeration).
//
// Global admin is granted only via REVUES_BOOTSTRAP_ADMIN_EMAIL (legacy until #310).
func (s *Store) ResolveLoginRoleStrict(ctx context.Context, email, bootstrapAdmin string, requireInvite bool) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	bootstrapAdmin = strings.ToLower(strings.TrimSpace(bootstrapAdmin))

	if bootstrapAdmin != "" && email == bootstrapAdmin {
		return auth.RoleAdmin, nil
	}

	if user, err := s.UserByEmail(ctx, email); err == nil {
		count, countErr := s.CountUserOrganizations(ctx, user.ID)
		if countErr != nil {
			return "", fmt.Errorf("count user organizations: %w", countErr)
		}
		if count > 0 {
			return user.Role, nil
		}
	} else if !errors.Is(err, ErrUserNotFound) {
		return "", err
	}

	if ok, err := s.HasPendingInvitationByEmail(ctx, email); err != nil {
		return "", fmt.Errorf("pending invitation lookup: %w", err)
	} else if ok {
		return auth.RoleEditor, nil
	}

	if requireInvite {
		return "", ErrEmailNotAllowed
	}

	return auth.RoleEditor, nil
}

// EnsureBootstrapOrgOwner adds the bootstrap admin as owner of the default org.
func (s *Store) EnsureBootstrapOrgOwner(ctx context.Context, userID int64, email, bootstrapAdmin string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	bootstrapAdmin = strings.ToLower(strings.TrimSpace(bootstrapAdmin))
	if bootstrapAdmin == "" || email != bootstrapAdmin {
		return nil
	}

	defaultOrg, err := s.OrganizationBySlug(ctx, "default")
	if err != nil {
		return fmt.Errorf("default organization: %w", err)
	}

	if err := s.AddOrganizationMember(ctx, defaultOrg.ID, userID, OrgRoleOwner); err != nil {
		return fmt.Errorf("bootstrap org owner: %w", err)
	}

	return nil
}
