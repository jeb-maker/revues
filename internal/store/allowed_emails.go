package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jeb-maker/revues/internal/auth"
)

// ErrEmailNotAllowed is returned when an email is not on the whitelist.
var ErrEmailNotAllowed = errors.New("email not allowed")

// ErrAllowedEmailNotFound is returned when a whitelist entry is missing.
var ErrAllowedEmailNotFound = errors.New("allowed email not found")

// AllowedEmail is a whitelisted login email scoped to an organization.
type AllowedEmail struct {
	Email     string
	Role      string
	CreatedAt string
}

// AllowedRole returns the role for email if whitelisted in the active organization.
func (s *Store) AllowedRole(ctx context.Context, email string) (string, bool, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return "", false, err
	}

	var role string
	err = s.db.QueryRowContext(ctx, `
		SELECT role FROM allowed_emails WHERE organization_id = ? AND email = ?
	`, orgID, strings.ToLower(strings.TrimSpace(email))).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("allowed role lookup: %w", err)
	}

	return role, true, nil
}

// InsertAllowedEmail adds an email to the whitelist for the active organization.
func (s *Store) InsertAllowedEmail(ctx context.Context, email, role string) error {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	email = strings.ToLower(strings.TrimSpace(email))
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO allowed_emails (organization_id, email, role, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(organization_id, email) DO UPDATE SET role = excluded.role
	`, orgID, email, role, now)
	if err != nil {
		return fmt.Errorf("insert allowed email: %w", err)
	}

	return nil
}

// CountAllowedEmails returns whitelist size for the active organization.
func (s *Store) CountAllowedEmails(ctx context.Context) (int, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return 0, err
	}

	var count int
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM allowed_emails WHERE organization_id = ?
	`, orgID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count allowed emails: %w", err)
	}

	return count, nil
}

// ListAllowedEmails returns whitelist entries for the active organization.
func (s *Store) ListAllowedEmails(ctx context.Context) ([]AllowedEmail, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT email, role, created_at
		FROM allowed_emails
		WHERE organization_id = ?
		ORDER BY email
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("list allowed emails: %w", err)
	}
	defer rows.Close()

	var emails []AllowedEmail
	for rows.Next() {
		var row AllowedEmail
		if err := rows.Scan(&row.Email, &row.Role, &row.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan allowed email: %w", err)
		}
		emails = append(emails, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate allowed emails: %w", err)
	}

	return emails, nil
}

// DeleteAllowedEmail removes an email from the whitelist in the active organization.
func (s *Store) DeleteAllowedEmail(ctx context.Context, email string) error {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	email = strings.ToLower(strings.TrimSpace(email))
	res, err := s.db.ExecContext(ctx, `
		DELETE FROM allowed_emails WHERE organization_id = ? AND email = ?
	`, orgID, email)
	if err != nil {
		return fmt.Errorf("delete allowed email: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete allowed email rows: %w", err)
	}
	if n == 0 {
		return ErrAllowedEmailNotFound
	}

	return nil
}

// ResolveLoginRole determines the global role for a verified GitHub email at login.
// Login is allowed when the email is whitelisted in an org, the user belongs to
// at least one org, matches REVUES_BOOTSTRAP_ADMIN_EMAIL, or has no org yet
// (self-service org creation).
func (s *Store) ResolveLoginRole(ctx context.Context, email, bootstrapAdmin string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	bootstrapAdmin = strings.ToLower(strings.TrimSpace(bootstrapAdmin))

	if role, ok, err := s.allowedRoleAnyOrg(ctx, email); err != nil {
		return "", err
	} else if ok {
		return role, nil
	}

	if user, err := s.UserByEmail(ctx, email); err == nil {
		count, err := s.CountUserOrganizations(ctx, user.ID)
		if err != nil {
			return "", fmt.Errorf("count user organizations: %w", err)
		}
		if count > 0 {
			return user.Role, nil
		}
	} else if !errors.Is(err, ErrUserNotFound) {
		return "", err
	}

	if bootstrapAdmin != "" && email == bootstrapAdmin {
		return auth.RoleAdmin, nil
	}

	if ok, err := s.HasPendingInvitationByEmail(ctx, email); err != nil {
		return "", fmt.Errorf("pending invitation lookup: %w", err)
	} else if ok {
		return auth.RoleEditor, nil
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

func (s *Store) allowedRoleAnyOrg(ctx context.Context, email string) (string, bool, error) {
	var role string
	err := s.db.QueryRowContext(ctx, `
		SELECT role FROM allowed_emails WHERE email = ? LIMIT 1
	`, email).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("allowed role any org: %w", err)
	}
	return role, true, nil
}
