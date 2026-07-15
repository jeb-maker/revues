package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrSubjectNotFound is returned when a subject id does not exist.
var ErrSubjectNotFound = errors.New("subject not found")

// Subject is a review subject container (site, asset, app, etc.).
type Subject struct {
	ID             int64
	OrganizationID int64
	Name           string
	Description    string
	ArchivedAt     sql.NullString
	CreatedAt      string
	UpdatedAt      string
}

// SubjectMember links a user to a subject's organization with a display role (v1 org-scoped access).
type SubjectMember struct {
	UserID      int64
	Login       string
	Email       string
	DisplayName string
	Role        string
	CreatedAt   string
}

// CreateSubject inserts a subject, matching domains and ensures the creator belongs to the org.
func (s *Store) CreateSubject(ctx context.Context, name, description string, creatorID int64, domains []string) (*Subject, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	domains = NormalizeTags(domains)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO subjects (organization_id, name, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, orgID, name, description, now, now)
	if err != nil {
		return nil, fmt.Errorf("insert subject: %w", err)
	}

	subjectID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("subject id: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO organization_members (organization_id, user_id, role, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(organization_id, user_id) DO NOTHING
	`, orgID, creatorID, OrgRoleMember, now)
	if err != nil {
		return nil, fmt.Errorf("ensure org member: %w", err)
	}

	if err := setSubjectDomainsTx(ctx, tx, subjectID, domains); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create subject: %w", err)
	}

	return s.SubjectByID(ctx, subjectID)
}

// SubjectByID loads a subject by primary key within the active organization.
func (s *Store) SubjectByID(ctx context.Context, id int64) (*Subject, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.subjectByID(ctx, id, orgID)
}

func (s *Store) subjectByID(ctx context.Context, id, orgID int64) (*Subject, error) {
	var sub Subject
	err := s.db.QueryRowContext(ctx, `
		SELECT id, organization_id, name, description, archived_at, created_at, updated_at
		FROM subjects WHERE id = ? AND organization_id = ?
	`, id, orgID).Scan(&sub.ID, &sub.OrganizationID, &sub.Name, &sub.Description, &sub.ArchivedAt, &sub.CreatedAt, &sub.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSubjectNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("subject by id: %w", err)
	}
	return &sub, nil
}

// SubjectByIDUnscoped loads a subject without organization filtering (system jobs).
func (s *Store) SubjectByIDUnscoped(ctx context.Context, id int64) (*Subject, error) {
	var sub Subject
	err := s.db.QueryRowContext(ctx, `
		SELECT id, organization_id, name, description, archived_at, created_at, updated_at
		FROM subjects WHERE id = ?
	`, id).Scan(&sub.ID, &sub.OrganizationID, &sub.Name, &sub.Description, &sub.ArchivedAt, &sub.CreatedAt, &sub.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSubjectNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("subject by id unscoped: %w", err)
	}
	return &sub, nil
}

// ListSubjects returns active subjects visible to the user in the active organization.
func (s *Store) ListSubjects(ctx context.Context, userID int64, admin bool, query string) ([]Subject, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	sqlQuery := `
		SELECT s.id, s.organization_id, s.name, s.description, s.archived_at, s.created_at, s.updated_at
		FROM subjects s`
	var args []any

	if admin {
		sqlQuery += `
		WHERE s.organization_id = ? AND s.archived_at IS NULL`
		args = append(args, orgID)
	} else {
		sqlQuery += `
		INNER JOIN organization_members om ON om.organization_id = s.organization_id
		WHERE s.organization_id = ? AND om.user_id = ? AND s.archived_at IS NULL`
		args = append(args, orgID, userID)
	}

	for _, term := range searchTerms(query) {
		pattern := likeContainsPattern(term)
		sqlQuery += ` AND (s.name LIKE ? ESCAPE '\' OR s.description LIKE ? ESCAPE '\')`
		args = append(args, pattern, pattern)
	}

	sqlQuery += ` ORDER BY s.name`

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("list subjects: %w", err)
	}
	defer rows.Close()

	var subjects []Subject
	for rows.Next() {
		var sub Subject
		if err := rows.Scan(&sub.ID, &sub.OrganizationID, &sub.Name, &sub.Description, &sub.ArchivedAt, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan subject: %w", err)
		}
		subjects = append(subjects, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subjects: %w", err)
	}

	return subjects, nil
}

// UpdateSubject changes name, description and matching domains.
func (s *Store) UpdateSubject(ctx context.Context, id int64, name, description string, domains []string) error {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	domains = NormalizeTags(domains)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	res, err := tx.ExecContext(ctx, `
		UPDATE subjects SET name = ?, description = ?, updated_at = ?
		WHERE id = ? AND organization_id = ? AND archived_at IS NULL
	`, name, description, now, id, orgID)
	if err != nil {
		return fmt.Errorf("update subject: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update subject rows: %w", err)
	}
	if n == 0 {
		return ErrSubjectNotFound
	}

	if err := setSubjectDomainsTx(ctx, tx, id, domains); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit update subject: %w", err)
	}
	return nil
}

// ArchiveSubject marks a subject as archived.
func (s *Store) ArchiveSubject(ctx context.Context, id int64) error {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, `
		UPDATE subjects SET archived_at = ?, updated_at = ?
		WHERE id = ? AND organization_id = ? AND archived_at IS NULL
	`, now, now, id, orgID)
	if err != nil {
		return fmt.Errorf("archive subject: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive subject rows: %w", err)
	}
	if n == 0 {
		return ErrSubjectNotFound
	}
	return nil
}

// MemberRole returns whether the user belongs to the subject's organization (v1 org-scoped access).
func (s *Store) MemberRole(ctx context.Context, subjectID, userID int64) (string, bool, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return "", false, err
	}

	var role string
	err = s.db.QueryRowContext(ctx, `
		SELECT om.role
		FROM subjects s
		INNER JOIN organization_members om ON om.organization_id = s.organization_id
		WHERE s.id = ? AND om.user_id = ? AND s.organization_id = ?
	`, subjectID, userID, orgID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("member role: %w", err)
	}
	switch role {
	case OrgRoleOwner, OrgRoleAdmin:
		return "lead", true, nil
	default:
		return "lead", true, nil
	}
}

// ListSubjectMembers returns organization members for the subject's organization (v1).
func (s *Store) ListSubjectMembers(ctx context.Context, subjectID int64) ([]SubjectMember, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT u.id, u.login, u.email, u.display_name,
		       CASE om.role WHEN 'owner' THEN 'lead' WHEN 'admin' THEN 'lead' ELSE 'contributor' END,
		       om.created_at
		FROM organization_members om
		INNER JOIN users u ON u.id = om.user_id
		INNER JOIN subjects s ON s.id = ? AND s.organization_id = om.organization_id
		WHERE s.organization_id = ?
		ORDER BY u.login
	`, subjectID, orgID)
	if err != nil {
		return nil, fmt.Errorf("list subject members: %w", err)
	}
	defer rows.Close()

	var members []SubjectMember
	for rows.Next() {
		var m SubjectMember
		if err := rows.Scan(&m.UserID, &m.Login, &m.Email, &m.DisplayName, &m.Role, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan subject member: %w", err)
		}
		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subject members: %w", err)
	}

	return members, nil
}

// AddSubjectMember ensures the user belongs to the subject's organization (v1).
func (s *Store) AddSubjectMember(ctx context.Context, subjectID, userID int64, role string) error {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return err
	}
	if _, err = s.subjectByID(ctx, subjectID, orgID); err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO organization_members (organization_id, user_id, role, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(organization_id, user_id) DO NOTHING
	`, orgID, userID, OrgRoleMember, now)
	if err != nil {
		return fmt.Errorf("add org member: %w", err)
	}
	return nil
}

// RemoveSubjectMember is a no-op in v1 (access via organization membership).
func (s *Store) RemoveSubjectMember(ctx context.Context, subjectID, userID int64) error {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return err
	}
	if _, err = s.subjectByID(ctx, subjectID, orgID); err != nil {
		return err
	}
	return nil
}

// CountSubjectLeads returns org owner/admin count for the subject's organization (v1).
func (s *Store) CountSubjectLeads(ctx context.Context, subjectID int64) (int, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return 0, err
	}

	var count int
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM organization_members om
		INNER JOIN subjects s ON s.id = ? AND s.organization_id = om.organization_id
		WHERE om.organization_id = ? AND om.role IN ('owner', 'admin')
	`, subjectID, orgID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count subject leads: %w", err)
	}
	return count, nil
}

// Deprecated project aliases — kept for store tests; use Subject* APIs in new code.

type Project = Subject

type ProjectMember = SubjectMember

var ErrProjectNotFound = ErrSubjectNotFound

func (s *Store) CreateProject(ctx context.Context, name, description string, creatorID int64, domains []string) (*Project, error) {
	return s.CreateSubject(ctx, name, description, creatorID, domains)
}

func (s *Store) ProjectByID(ctx context.Context, id int64) (*Project, error) {
	return s.SubjectByID(ctx, id)
}

func (s *Store) ProjectByIDUnscoped(ctx context.Context, id int64) (*Project, error) {
	return s.SubjectByIDUnscoped(ctx, id)
}

func (s *Store) ListProjects(ctx context.Context, userID int64, admin bool, query string) ([]Project, error) {
	return s.ListSubjects(ctx, userID, admin, query)
}

func (s *Store) UpdateProject(ctx context.Context, id int64, name, description string, domains []string) error {
	return s.UpdateSubject(ctx, id, name, description, domains)
}

func (s *Store) ArchiveProject(ctx context.Context, id int64) error {
	return s.ArchiveSubject(ctx, id)
}

func (s *Store) ListProjectMembers(ctx context.Context, subjectID int64) ([]ProjectMember, error) {
	return s.ListSubjectMembers(ctx, subjectID)
}

func (s *Store) AddProjectMember(ctx context.Context, subjectID, userID int64, role string) error {
	return s.AddSubjectMember(ctx, subjectID, userID, role)
}

func (s *Store) RemoveProjectMember(ctx context.Context, subjectID, userID int64) error {
	return s.RemoveSubjectMember(ctx, subjectID, userID)
}

func (s *Store) CountProjectLeads(ctx context.Context, subjectID int64) (int, error) {
	return s.CountSubjectLeads(ctx, subjectID)
}
