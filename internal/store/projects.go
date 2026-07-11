package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrProjectNotFound is returned when a project id does not exist.
var ErrProjectNotFound = errors.New("project not found")

// Project is a review project container.
type Project struct {
	ID             int64
	OrganizationID int64
	Name           string
	Description    string
	ArchivedAt     sql.NullString
	CreatedAt      string
	UpdatedAt      string
}

// ProjectMember links a user to a project with a local role.
type ProjectMember struct {
	UserID      int64
	Login       string
	Email       string
	DisplayName string
	Role        string
	CreatedAt   string
}

// CreateProject inserts a project, tags and adds creator as lead.
func (s *Store) CreateProject(ctx context.Context, name, description string, creatorID int64, tags []string) (*Project, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	tags = NormalizeTags(tags)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO projects (organization_id, name, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, orgID, name, description, now, now)
	if err != nil {
		return nil, fmt.Errorf("insert project: %w", err)
	}

	projectID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("project id: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO project_members (project_id, user_id, role, created_at)
		VALUES (?, ?, ?, ?)
	`, projectID, creatorID, "lead", now)
	if err != nil {
		return nil, fmt.Errorf("insert project lead: %w", err)
	}

	if err := setProjectTagsTx(ctx, tx, projectID, tags); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create project: %w", err)
	}

	return s.ProjectByID(ctx, projectID)
}

// ProjectByID loads a project by primary key within the active organization.
func (s *Store) ProjectByID(ctx context.Context, id int64) (*Project, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.projectByID(ctx, id, orgID)
}

func (s *Store) projectByID(ctx context.Context, id, orgID int64) (*Project, error) {
	var p Project
	err := s.db.QueryRowContext(ctx, `
		SELECT id, organization_id, name, description, archived_at, created_at, updated_at
		FROM projects WHERE id = ? AND organization_id = ?
	`, id, orgID).Scan(&p.ID, &p.OrganizationID, &p.Name, &p.Description, &p.ArchivedAt, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProjectNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("project by id: %w", err)
	}
	return &p, nil
}

// ProjectByIDUnscoped loads a project without organization filtering (system jobs).
func (s *Store) ProjectByIDUnscoped(ctx context.Context, id int64) (*Project, error) {
	var p Project
	err := s.db.QueryRowContext(ctx, `
		SELECT id, organization_id, name, description, archived_at, created_at, updated_at
		FROM projects WHERE id = ?
	`, id).Scan(&p.ID, &p.OrganizationID, &p.Name, &p.Description, &p.ArchivedAt, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProjectNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("project by id unscoped: %w", err)
	}
	return &p, nil
}

// ListProjects returns active projects visible to the user in the active organization.
func (s *Store) ListProjects(ctx context.Context, userID int64, admin bool, query string) ([]Project, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	sqlQuery := `
		SELECT p.id, p.organization_id, p.name, p.description, p.archived_at, p.created_at, p.updated_at
		FROM projects p`
	var args []any

	if admin {
		sqlQuery += `
		WHERE p.organization_id = ? AND p.archived_at IS NULL`
		args = append(args, orgID)
	} else {
		sqlQuery += `
		INNER JOIN project_members pm ON pm.project_id = p.id
		WHERE p.organization_id = ? AND pm.user_id = ? AND p.archived_at IS NULL`
		args = append(args, orgID, userID)
	}

	for _, term := range searchTerms(query) {
		pattern := likeContainsPattern(term)
		sqlQuery += ` AND (p.name LIKE ? ESCAPE '\' OR p.description LIKE ? ESCAPE '\')`
		args = append(args, pattern, pattern)
	}

	sqlQuery += ` ORDER BY p.name`

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.OrganizationID, &p.Name, &p.Description, &p.ArchivedAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects: %w", err)
	}

	return projects, nil
}

// UpdateProject changes name, description and tags.
func (s *Store) UpdateProject(ctx context.Context, id int64, name, description string, tags []string) error {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	tags = NormalizeTags(tags)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	res, err := tx.ExecContext(ctx, `
		UPDATE projects SET name = ?, description = ?, updated_at = ?
		WHERE id = ? AND organization_id = ? AND archived_at IS NULL
	`, name, description, now, id, orgID)
	if err != nil {
		return fmt.Errorf("update project: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update project rows: %w", err)
	}
	if n == 0 {
		return ErrProjectNotFound
	}

	if err := setProjectTagsTx(ctx, tx, id, tags); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit update project: %w", err)
	}
	return nil
}

// ArchiveProject marks a project as archived.
func (s *Store) ArchiveProject(ctx context.Context, id int64) error {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, `
		UPDATE projects SET archived_at = ?, updated_at = ?
		WHERE id = ? AND organization_id = ? AND archived_at IS NULL
	`, now, now, id, orgID)
	if err != nil {
		return fmt.Errorf("archive project: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive project rows: %w", err)
	}
	if n == 0 {
		return ErrProjectNotFound
	}
	return nil
}

// MemberRole returns the local role for a user on a project.
func (s *Store) MemberRole(ctx context.Context, projectID, userID int64) (string, bool, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return "", false, err
	}

	var role string
	err = s.db.QueryRowContext(ctx, `
		SELECT pm.role
		FROM project_members pm
		INNER JOIN projects p ON p.id = pm.project_id
		WHERE pm.project_id = ? AND pm.user_id = ? AND p.organization_id = ?
	`, projectID, userID, orgID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("member role: %w", err)
	}
	return role, true, nil
}

// ListProjectMembers returns members with user profile fields.
func (s *Store) ListProjectMembers(ctx context.Context, projectID int64) ([]ProjectMember, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT u.id, u.login, u.email, u.display_name, pm.role, pm.created_at
		FROM project_members pm
		INNER JOIN users u ON u.id = pm.user_id
		INNER JOIN projects p ON p.id = pm.project_id
		WHERE pm.project_id = ? AND p.organization_id = ?
		ORDER BY u.login
	`, projectID, orgID)
	if err != nil {
		return nil, fmt.Errorf("list project members: %w", err)
	}
	defer rows.Close()

	var members []ProjectMember
	for rows.Next() {
		var m ProjectMember
		if err := rows.Scan(&m.UserID, &m.Login, &m.Email, &m.DisplayName, &m.Role, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan project member: %w", err)
		}
		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project members: %w", err)
	}

	return members, nil
}

// AddProjectMember assigns a user to a project.
func (s *Store) AddProjectMember(ctx context.Context, projectID, userID int64, role string) error {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	if _, err = s.projectByID(ctx, projectID, orgID); err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO project_members (project_id, user_id, role, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(project_id, user_id) DO UPDATE SET role = excluded.role
	`, projectID, userID, role, now)
	if err != nil {
		return fmt.Errorf("add project member: %w", err)
	}
	return nil
}

// RemoveProjectMember removes a user from a project.
func (s *Store) RemoveProjectMember(ctx context.Context, projectID, userID int64) error {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	res, err := s.db.ExecContext(ctx, `
		DELETE FROM project_members
		WHERE project_id = ? AND user_id = ?
		  AND project_id IN (SELECT id FROM projects WHERE id = ? AND organization_id = ?)
	`, projectID, userID, projectID, orgID)
	if err != nil {
		return fmt.Errorf("remove project member: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("remove project member rows: %w", err)
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// CountProjectLeads returns lead members count.
func (s *Store) CountProjectLeads(ctx context.Context, projectID int64) (int, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return 0, err
	}

	var count int
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM project_members pm
		INNER JOIN projects p ON p.id = pm.project_id
		WHERE pm.project_id = ? AND pm.role = 'lead' AND p.organization_id = ?
	`, projectID, orgID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count project leads: %w", err)
	}
	return count, nil
}
