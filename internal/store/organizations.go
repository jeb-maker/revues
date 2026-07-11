package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ErrOrganizationNotFound is returned when an organization lookup fails.
var ErrOrganizationNotFound = errors.New("organization not found")

// ErrOrganizationSlugTaken is returned when a slug is already in use.
var ErrOrganizationSlugTaken = errors.New("organization slug taken")

// ErrInvalidOrganizationSlug is returned when a slug cannot be normalized.
var ErrInvalidOrganizationSlug = errors.New("invalid organization slug")

const (
	OrgRoleOwner  = "owner"
	OrgRoleAdmin  = "admin"
	OrgRoleMember = "member"
)

var organizationSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// Organization is a multi-tenant container above projects.
type Organization struct {
	ID        int64
	Name      string
	Slug      string
	CreatedAt string
	CreatedBy sql.NullInt64
}

// OrganizationMembership links a user to an organization with a role.
type OrganizationMembership struct {
	Organization Organization
	Role         string
	JoinedAt     string
}

// NormalizeOrganizationSlug lowercases and restricts a slug to [a-z0-9-].
func NormalizeOrganizationSlug(slug string) (string, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if slug == "" {
		return "", ErrInvalidOrganizationSlug
	}

	var b strings.Builder
	prevHyphen := false
	for _, r := range slug {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevHyphen = false
		case r == '-' || r == ' ' || r == '_':
			if !prevHyphen && b.Len() > 0 {
				b.WriteByte('-')
				prevHyphen = true
			}
		default:
			if !prevHyphen && b.Len() > 0 {
				b.WriteByte('-')
				prevHyphen = true
			}
		}
	}

	normalized := strings.Trim(b.String(), "-")
	if normalized == "" || !organizationSlugPattern.MatchString(normalized) {
		return "", ErrInvalidOrganizationSlug
	}

	return normalized, nil
}

// CreateOrganization inserts an organization with a normalized unique slug.
func (s *Store) CreateOrganization(ctx context.Context, name, slug string, createdBy int64) (*Organization, error) {
	normalized, err := NormalizeOrganizationSlug(slug)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	var createdByVal sql.NullInt64
	if createdBy > 0 {
		createdByVal = sql.NullInt64{Int64: createdBy, Valid: true}
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO organizations (name, slug, created_at, created_by)
		VALUES (?, ?, ?, ?)
	`, name, normalized, now, createdByVal)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrOrganizationSlugTaken
		}
		return nil, fmt.Errorf("insert organization: %w", err)
	}

	return s.OrganizationBySlug(ctx, normalized)
}

// OrganizationBySlug loads an organization by slug.
func (s *Store) OrganizationBySlug(ctx context.Context, slug string) (*Organization, error) {
	slug, err := NormalizeOrganizationSlug(slug)
	if err != nil {
		return nil, err
	}

	var org Organization
	err = s.db.QueryRowContext(ctx, `
		SELECT id, name, slug, created_at, created_by
		FROM organizations WHERE slug = ?
	`, slug).Scan(&org.ID, &org.Name, &org.Slug, &org.CreatedAt, &org.CreatedBy)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrOrganizationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("organization by slug: %w", err)
	}

	return &org, nil
}

// OrganizationByID loads an organization by primary key.
func (s *Store) OrganizationByID(ctx context.Context, id int64) (*Organization, error) {
	var org Organization
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, slug, created_at, created_by
		FROM organizations WHERE id = ?
	`, id).Scan(&org.ID, &org.Name, &org.Slug, &org.CreatedAt, &org.CreatedBy)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrOrganizationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("organization by id: %w", err)
	}

	return &org, nil
}

// AddOrganizationMember assigns a user to an organization.
func (s *Store) AddOrganizationMember(ctx context.Context, organizationID, userID int64, role string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO organization_members (organization_id, user_id, role, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(organization_id, user_id) DO UPDATE SET role = excluded.role
	`, organizationID, userID, role, now)
	if err != nil {
		return fmt.Errorf("add organization member: %w", err)
	}

	return nil
}

// RemoveOrganizationMember removes a user from an organization.
func (s *Store) RemoveOrganizationMember(ctx context.Context, organizationID, userID int64) error {
	res, err := s.db.ExecContext(ctx, `
		DELETE FROM organization_members WHERE organization_id = ? AND user_id = ?
	`, organizationID, userID)
	if err != nil {
		return fmt.Errorf("remove organization member: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("remove organization member rows: %w", err)
	}
	if n == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// OrganizationMemberRole returns the role for a user in an organization.
func (s *Store) OrganizationMemberRole(ctx context.Context, organizationID, userID int64) (string, bool, error) {
	var role string
	err := s.db.QueryRowContext(ctx, `
		SELECT role FROM organization_members WHERE organization_id = ? AND user_id = ?
	`, organizationID, userID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("organization member role: %w", err)
	}

	return role, true, nil
}

// ListUserOrganizations returns organizations a user belongs to.
func (s *Store) ListUserOrganizations(ctx context.Context, userID int64) ([]OrganizationMembership, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT o.id, o.name, o.slug, o.created_at, o.created_by, om.role, om.created_at
		FROM organization_members om
		INNER JOIN organizations o ON o.id = om.organization_id
		WHERE om.user_id = ?
		ORDER BY o.name
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list user organizations: %w", err)
	}
	defer rows.Close()

	var memberships []OrganizationMembership
	for rows.Next() {
		var m OrganizationMembership
		if err := rows.Scan(
			&m.Organization.ID,
			&m.Organization.Name,
			&m.Organization.Slug,
			&m.Organization.CreatedAt,
			&m.Organization.CreatedBy,
			&m.Role,
			&m.JoinedAt,
		); err != nil {
			return nil, fmt.Errorf("scan organization membership: %w", err)
		}
		memberships = append(memberships, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user organizations: %w", err)
	}

	return memberships, nil
}

// CountUserOrganizations returns how many organizations a user belongs to.
func (s *Store) CountUserOrganizations(ctx context.Context, userID int64) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM organization_members WHERE user_id = ?
	`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count user organizations: %w", err)
	}

	return count, nil
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
