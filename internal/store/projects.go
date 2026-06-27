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
	ID          int64
	Name        string
	Description string
	ArchivedAt  sql.NullString
	CreatedAt   string
	UpdatedAt   string
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

// CreateProject inserts a project and adds creator as lead.
func (s *Store) CreateProject(ctx context.Context, name, description string, creatorID int64) (*Project, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO projects (name, description, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`, name, description, now, now)
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

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create project: %w", err)
	}

	return s.ProjectByID(ctx, projectID)
}

// ProjectByID loads a project by primary key.
func (s *Store) ProjectByID(ctx context.Context, id int64) (*Project, error) {
	var p Project
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, description, archived_at, created_at, updated_at
		FROM projects WHERE id = ?
	`, id).Scan(&p.ID, &p.Name, &p.Description, &p.ArchivedAt, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProjectNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("project by id: %w", err)
	}
	return &p, nil
}

// ListProjects returns active projects visible to the user.
func (s *Store) ListProjects(ctx context.Context, userID int64, admin bool) ([]Project, error) {
	var rows *sql.Rows
	var err error

	if admin {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, name, description, archived_at, created_at, updated_at
			FROM projects
			WHERE archived_at IS NULL
			ORDER BY name
		`)
	} else {
		rows, err = s.db.QueryContext(ctx, `
			SELECT p.id, p.name, p.description, p.archived_at, p.created_at, p.updated_at
			FROM projects p
			INNER JOIN project_members pm ON pm.project_id = p.id
			WHERE pm.user_id = ? AND p.archived_at IS NULL
			ORDER BY p.name
		`, userID)
	}
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.ArchivedAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects: %w", err)
	}

	return projects, nil
}

// UpdateProject changes name and description.
func (s *Store) UpdateProject(ctx context.Context, id int64, name, description string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, `
		UPDATE projects SET name = ?, description = ?, updated_at = ?
		WHERE id = ? AND archived_at IS NULL
	`, name, description, now, id)
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
	return nil
}

// ArchiveProject marks a project as archived.
func (s *Store) ArchiveProject(ctx context.Context, id int64) error {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, `
		UPDATE projects SET archived_at = ?, updated_at = ?
		WHERE id = ? AND archived_at IS NULL
	`, now, now, id)
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
	var role string
	err := s.db.QueryRowContext(ctx, `
		SELECT role FROM project_members WHERE project_id = ? AND user_id = ?
	`, projectID, userID).Scan(&role)
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
	rows, err := s.db.QueryContext(ctx, `
		SELECT u.id, u.login, u.email, u.display_name, pm.role, pm.created_at
		FROM project_members pm
		INNER JOIN users u ON u.id = pm.user_id
		WHERE pm.project_id = ?
		ORDER BY u.login
	`, projectID)
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
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `
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
	res, err := s.db.ExecContext(ctx, `
		DELETE FROM project_members WHERE project_id = ? AND user_id = ?
	`, projectID, userID)
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
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM project_members WHERE project_id = ? AND role = 'lead'
	`, projectID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count project leads: %w", err)
	}
	return count, nil
}
