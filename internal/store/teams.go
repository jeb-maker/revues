package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const (
	SubjectRoleLead        = "lead"
	SubjectRoleContributor = "contributor"
	SubjectRoleViewer      = "viewer"
)

// ErrTeamNotFound is returned when a team lookup fails in the active org.
var ErrTeamNotFound = errors.New("team not found")

// ErrTeamMemberNotFound is returned when a user is not a member of the team.
var ErrTeamMemberNotFound = errors.New("team member not found")

// ErrTeamSlugTaken is returned when a team slug is already used in the org.
var ErrTeamSlugTaken = errors.New("team slug taken")

// ErrInvalidSubjectRole is returned when a subject access role is unknown.
var ErrInvalidSubjectRole = errors.New("invalid subject role")

// ErrDirectSubjectMemberNotFound is returned when a direct subject_members row is missing.
var ErrDirectSubjectMemberNotFound = errors.New("direct subject member not found")

// ErrTeamSubjectRoleNotFound is returned when a team has no role on a subject.
var ErrTeamSubjectRoleNotFound = errors.New("team subject role not found")

// OrganizationTeam is a named team within an organization.
type OrganizationTeam struct {
	ID             int64
	OrganizationID int64
	Name           string
	Slug           string
	Description    string
	CreatedAt      string
	MemberCount    int // filled by ListOrganizationTeams
}

// TeamMember links a user to a team (with display fields from users).
type TeamMember struct {
	TeamID      int64
	UserID      int64
	Login       string
	Email       string
	DisplayName string
	CreatedAt   string
}

// DirectSubjectMember is a row in subject_members (exception path — not org listing).
type DirectSubjectMember struct {
	SubjectID   int64
	UserID      int64
	Role        string
	Login       string
	Email       string
	DisplayName string
	CreatedAt   string
}

// TeamSubjectRole assigns a team a role on a subject.
type TeamSubjectRole struct {
	TeamID      int64
	TeamName    string
	SubjectID   int64
	Role        string
	MemberCount int // filled by ListSubjectTeams
	GrantedBy   sql.NullInt64
	CreatedAt   string
}

// NormalizeTeamSlug reuses the organization slug rules.
func NormalizeTeamSlug(slug string) (string, error) {
	return NormalizeOrganizationSlug(slug)
}

func normalizeSubjectRole(role string) (string, error) {
	switch role {
	case SubjectRoleLead, SubjectRoleContributor, SubjectRoleViewer:
		return role, nil
	default:
		return "", ErrInvalidSubjectRole
	}
}

// CreateTeam inserts a team in the active organization.
func (s *Store) CreateTeam(ctx context.Context, name, slug, description string) (*OrganizationTeam, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	normalized, err := NormalizeTeamSlug(slug)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO organization_teams (organization_id, name, slug, description, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, orgID, name, normalized, description, now)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrTeamSlugTaken
		}
		return nil, fmt.Errorf("insert organization team: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("organization team last insert id: %w", err)
	}
	return s.TeamByID(ctx, id)
}

// TeamByID loads a team scoped to the active organization.
func (s *Store) TeamByID(ctx context.Context, teamID int64) (*OrganizationTeam, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var team OrganizationTeam
	err = s.db.QueryRowContext(ctx, `
		SELECT id, organization_id, name, slug, description, created_at
		FROM organization_teams
		WHERE id = ? AND organization_id = ?
	`, teamID, orgID).Scan(&team.ID, &team.OrganizationID, &team.Name, &team.Slug, &team.Description, &team.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTeamNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("team by id: %w", err)
	}
	return &team, nil
}

// ListOrganizationTeams lists teams of the active organization ordered by name.
func (s *Store) ListOrganizationTeams(ctx context.Context) ([]OrganizationTeam, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.id, t.organization_id, t.name, t.slug, t.description, t.created_at,
			(SELECT COUNT(*) FROM team_members tm WHERE tm.team_id = t.id)
		FROM organization_teams t
		WHERE t.organization_id = ?
		ORDER BY t.name COLLATE NOCASE, t.id
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("list organization teams: %w", err)
	}
	defer rows.Close()

	var teams []OrganizationTeam
	for rows.Next() {
		var team OrganizationTeam
		if scanErr := rows.Scan(
			&team.ID, &team.OrganizationID, &team.Name, &team.Slug, &team.Description, &team.CreatedAt, &team.MemberCount,
		); scanErr != nil {
			return nil, fmt.Errorf("scan organization team: %w", scanErr)
		}
		teams = append(teams, team)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("list organization teams rows: %w", err)
	}
	return teams, nil
}

// AddTeamMember adds a user to a team in the active organization.
func (s *Store) AddTeamMember(ctx context.Context, teamID, userID int64) error {
	if _, err := s.TeamByID(ctx, teamID); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO team_members (team_id, user_id, created_at)
		VALUES (?, ?, ?)
		ON CONFLICT(team_id, user_id) DO NOTHING
	`, teamID, userID, now)
	if err != nil {
		return fmt.Errorf("add team member: %w", err)
	}
	return nil
}

// RemoveTeamMember removes a user from a team in the active organization.
func (s *Store) RemoveTeamMember(ctx context.Context, teamID, userID int64) error {
	if _, err := s.TeamByID(ctx, teamID); err != nil {
		return err
	}
	res, err := s.db.ExecContext(ctx, `
		DELETE FROM team_members WHERE team_id = ? AND user_id = ?
	`, teamID, userID)
	if err != nil {
		return fmt.Errorf("remove team member: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("remove team member rows: %w", err)
	}
	if n == 0 {
		return ErrTeamMemberNotFound
	}
	return nil
}

// ListTeamMembers lists members of a team in the active organization (login/email).
func (s *Store) ListTeamMembers(ctx context.Context, teamID int64) ([]TeamMember, error) {
	if _, err := s.TeamByID(ctx, teamID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT m.team_id, m.user_id, u.login, u.email, u.display_name, m.created_at
		FROM team_members m
		INNER JOIN users u ON u.id = m.user_id
		WHERE m.team_id = ?
		ORDER BY u.login COLLATE NOCASE, m.user_id
	`, teamID)
	if err != nil {
		return nil, fmt.Errorf("list team members: %w", err)
	}
	defer rows.Close()

	var members []TeamMember
	for rows.Next() {
		var m TeamMember
		if scanErr := rows.Scan(&m.TeamID, &m.UserID, &m.Login, &m.Email, &m.DisplayName, &m.CreatedAt); scanErr != nil {
			return nil, fmt.Errorf("scan team member: %w", scanErr)
		}
		members = append(members, m)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("list team members rows: %w", err)
	}
	return members, nil
}

// ListUserTeams lists teams of the active org that include userID.
func (s *Store) ListUserTeams(ctx context.Context, userID int64) ([]OrganizationTeam, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.id, t.organization_id, t.name, t.slug, t.description, t.created_at
		FROM organization_teams t
		INNER JOIN team_members m ON m.team_id = t.id
		WHERE t.organization_id = ? AND m.user_id = ?
		ORDER BY t.name COLLATE NOCASE, t.id
	`, orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("list user teams: %w", err)
	}
	defer rows.Close()

	var teams []OrganizationTeam
	for rows.Next() {
		var team OrganizationTeam
		if scanErr := rows.Scan(&team.ID, &team.OrganizationID, &team.Name, &team.Slug, &team.Description, &team.CreatedAt); scanErr != nil {
			return nil, fmt.Errorf("scan user team: %w", scanErr)
		}
		teams = append(teams, team)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("list user teams rows: %w", err)
	}
	return teams, nil
}

// UpsertDirectSubjectMember upserts a row in subject_members (active org).
func (s *Store) UpsertDirectSubjectMember(ctx context.Context, subjectID, userID int64, role string) error {
	role, err := normalizeSubjectRole(role)
	if err != nil {
		return err
	}
	if _, err = s.SubjectByID(ctx, subjectID); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO subject_members (subject_id, user_id, role, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(subject_id, user_id) DO UPDATE SET role = excluded.role
	`, subjectID, userID, role, now)
	if err != nil {
		return fmt.Errorf("upsert direct subject member: %w", err)
	}
	return nil
}

// RemoveDirectSubjectMember removes a row from subject_members.
func (s *Store) RemoveDirectSubjectMember(ctx context.Context, subjectID, userID int64) error {
	if _, err := s.SubjectByID(ctx, subjectID); err != nil {
		return err
	}
	res, err := s.db.ExecContext(ctx, `
		DELETE FROM subject_members WHERE subject_id = ? AND user_id = ?
	`, subjectID, userID)
	if err != nil {
		return fmt.Errorf("remove direct subject member: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("remove direct subject member rows: %w", err)
	}
	if n == 0 {
		return ErrDirectSubjectMemberNotFound
	}
	return nil
}

// ListDirectSubjectMembers lists subject_members rows for a subject in the active org.
func (s *Store) ListDirectSubjectMembers(ctx context.Context, subjectID int64) ([]DirectSubjectMember, error) {
	if _, err := s.SubjectByID(ctx, subjectID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT sm.subject_id, sm.user_id, sm.role, u.login, u.email, u.display_name, sm.created_at
		FROM subject_members sm
		INNER JOIN users u ON u.id = sm.user_id
		WHERE sm.subject_id = ?
		ORDER BY u.login, sm.user_id
	`, subjectID)
	if err != nil {
		return nil, fmt.Errorf("list direct subject members: %w", err)
	}
	defer rows.Close()

	var members []DirectSubjectMember
	for rows.Next() {
		var m DirectSubjectMember
		if scanErr := rows.Scan(&m.SubjectID, &m.UserID, &m.Role, &m.Login, &m.Email, &m.DisplayName, &m.CreatedAt); scanErr != nil {
			return nil, fmt.Errorf("scan direct subject member: %w", scanErr)
		}
		members = append(members, m)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("list direct subject members rows: %w", err)
	}
	return members, nil
}

// GrantTeamSubjectRole assigns or updates a team's role on a subject (same org).
func (s *Store) GrantTeamSubjectRole(ctx context.Context, teamID, subjectID int64, role string, grantedBy int64) error {
	role, err := normalizeSubjectRole(role)
	if err != nil {
		return err
	}
	if _, err = s.TeamByID(ctx, teamID); err != nil {
		return err
	}
	if _, err = s.SubjectByID(ctx, subjectID); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	var grantedByVal sql.NullInt64
	if grantedBy > 0 {
		grantedByVal = sql.NullInt64{Int64: grantedBy, Valid: true}
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO team_subject_roles (team_id, subject_id, role, granted_by, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(team_id, subject_id) DO UPDATE SET
			role = excluded.role,
			granted_by = excluded.granted_by
	`, teamID, subjectID, role, grantedByVal, now)
	if err != nil {
		return fmt.Errorf("grant team subject role: %w", err)
	}
	return nil
}

// RevokeTeamSubjectRole removes a team's role on a subject.
func (s *Store) RevokeTeamSubjectRole(ctx context.Context, teamID, subjectID int64) error {
	if _, err := s.TeamByID(ctx, teamID); err != nil {
		return err
	}
	if _, err := s.SubjectByID(ctx, subjectID); err != nil {
		return err
	}
	res, err := s.db.ExecContext(ctx, `
		DELETE FROM team_subject_roles WHERE team_id = ? AND subject_id = ?
	`, teamID, subjectID)
	if err != nil {
		return fmt.Errorf("revoke team subject role: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("revoke team subject role rows: %w", err)
	}
	if n == 0 {
		return ErrTeamSubjectRoleNotFound
	}
	return nil
}

// ListTeamSubjects lists subjects a team can access in the active org.
func (s *Store) ListTeamSubjects(ctx context.Context, teamID int64) ([]TeamSubjectRole, error) {
	if _, err := s.TeamByID(ctx, teamID); err != nil {
		return nil, err
	}
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.team_id, t.name, r.subject_id, r.role, r.granted_by, r.created_at
		FROM team_subject_roles r
		INNER JOIN organization_teams t ON t.id = r.team_id
		INNER JOIN subjects s ON s.id = r.subject_id
		WHERE r.team_id = ? AND s.organization_id = ?
		ORDER BY r.subject_id
	`, teamID, orgID)
	if err != nil {
		return nil, fmt.Errorf("list team subjects: %w", err)
	}
	defer rows.Close()

	var roles []TeamSubjectRole
	for rows.Next() {
		var r TeamSubjectRole
		if scanErr := rows.Scan(&r.TeamID, &r.TeamName, &r.SubjectID, &r.Role, &r.GrantedBy, &r.CreatedAt); scanErr != nil {
			return nil, fmt.Errorf("scan team subject role: %w", scanErr)
		}
		roles = append(roles, r)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("list team subjects rows: %w", err)
	}
	return roles, nil
}

// ListSubjectTeams lists teams assigned to a subject in the active org.
func (s *Store) ListSubjectTeams(ctx context.Context, subjectID int64) ([]TeamSubjectRole, error) {
	if _, err := s.SubjectByID(ctx, subjectID); err != nil {
		return nil, err
	}
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.team_id, t.name, r.subject_id, r.role,
			(SELECT COUNT(*) FROM team_members tm WHERE tm.team_id = r.team_id),
			r.granted_by, r.created_at
		FROM team_subject_roles r
		INNER JOIN organization_teams t ON t.id = r.team_id
		WHERE r.subject_id = ? AND t.organization_id = ?
		ORDER BY t.name COLLATE NOCASE, r.team_id
	`, subjectID, orgID)
	if err != nil {
		return nil, fmt.Errorf("list subject teams: %w", err)
	}
	defer rows.Close()

	var roles []TeamSubjectRole
	for rows.Next() {
		var r TeamSubjectRole
		if scanErr := rows.Scan(
			&r.TeamID, &r.TeamName, &r.SubjectID, &r.Role, &r.MemberCount, &r.GrantedBy, &r.CreatedAt,
		); scanErr != nil {
			return nil, fmt.Errorf("scan subject team role: %w", scanErr)
		}
		roles = append(roles, r)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("list subject teams rows: %w", err)
	}
	return roles, nil
}
