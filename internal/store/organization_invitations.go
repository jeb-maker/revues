package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrOrganizationInvitationNotFound is returned when an invitation lookup fails.
var ErrOrganizationInvitationNotFound = errors.New("organization invitation not found")

// OrganizationInvitation is a pending invite for a user to join an organization.
type OrganizationInvitation struct {
	ID               int64
	Email            string
	OrganizationID   int64
	OrganizationName string
	OrgRole          string
	CreatedAt        string
}

// CreateOrganizationInvitation records a pending invite for an email address.
func (s *Store) CreateOrganizationInvitation(
	ctx context.Context,
	email string,
	organizationID int64,
) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return fmt.Errorf("invitation email: empty")
	}

	now := time.Now().UTC().Format(time.RFC3339)

	_, err := s.db.ExecContext(ctx, `
		DELETE FROM organization_invitations
		WHERE organization_id = ? AND email = ?
	`, organizationID, email)
	if err != nil {
		return fmt.Errorf("clear organization invitation: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO organization_invitations (email, organization_id, org_role, created_at)
		VALUES (?, ?, ?, ?)
	`, email, organizationID, OrgRoleMember, now)
	if err != nil {
		return fmt.Errorf("insert organization invitation: %w", err)
	}

	return nil
}

// ListPendingInvitationsByEmail returns open invitations for an email address.
func (s *Store) ListPendingInvitationsByEmail(ctx context.Context, email string) ([]OrganizationInvitation, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, nil
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT i.id, i.email, i.organization_id, o.name, i.org_role, i.created_at
		FROM organization_invitations i
		INNER JOIN organizations o ON o.id = i.organization_id
		WHERE i.email = ?
		ORDER BY i.created_at
	`, email)
	if err != nil {
		return nil, fmt.Errorf("list pending invitations: %w", err)
	}
	defer rows.Close()

	var invites []OrganizationInvitation
	for rows.Next() {
		var inv OrganizationInvitation
		if err := rows.Scan(
			&inv.ID,
			&inv.Email,
			&inv.OrganizationID,
			&inv.OrganizationName,
			&inv.OrgRole,
			&inv.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan organization invitation: %w", err)
		}
		invites = append(invites, inv)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate organization invitations: %w", err)
	}

	return invites, nil
}

// HasPendingInvitationByEmail reports whether an email has at least one pending invite.
func (s *Store) HasPendingInvitationByEmail(ctx context.Context, email string) (bool, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return false, nil
	}

	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM organization_invitations WHERE email = ?
	`, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("count pending invitations: %w", err)
	}

	return count > 0, nil
}

// OrganizationInvitationByID loads a pending invitation by primary key.
func (s *Store) OrganizationInvitationByID(ctx context.Context, id int64) (*OrganizationInvitation, error) {
	var inv OrganizationInvitation
	err := s.db.QueryRowContext(ctx, `
		SELECT i.id, i.email, i.organization_id, o.name, i.org_role, i.created_at
		FROM organization_invitations i
		INNER JOIN organizations o ON o.id = i.organization_id
		WHERE i.id = ?
	`, id).Scan(
		&inv.ID,
		&inv.Email,
		&inv.OrganizationID,
		&inv.OrganizationName,
		&inv.OrgRole,
		&inv.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrOrganizationInvitationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("organization invitation by id: %w", err)
	}

	return &inv, nil
}

// DeleteOrganizationInvitation removes a pending invitation.
func (s *Store) DeleteOrganizationInvitation(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM organization_invitations WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete organization invitation: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete organization invitation rows: %w", err)
	}
	if n == 0 {
		return ErrOrganizationInvitationNotFound
	}

	return nil
}
