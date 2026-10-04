package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
)

const (
	AccessSourceDirect     = "direct"
	AccessSourceOrgAdmin   = "org_admin"
	accessSourceTeamPrefix = "team:"
)

// SubjectAccess is the resolved visibility and effective role for a subject.
type SubjectAccess struct {
	Visible bool
	Role    string   // lead | contributor | viewer | ""
	Sources []string // "direct", "team:{id}", "org_admin"
}

// HasSource reports whether Sources contains source.
func (a SubjectAccess) HasSource(source string) bool {
	for _, s := range a.Sources {
		if s == source {
			return true
		}
	}
	return false
}

// IsSupervisor is true for org owner/admin (visible without subject role).
func (a SubjectAccess) IsSupervisor() bool {
	return a.HasSource(AccessSourceOrgAdmin)
}

// RoleAtLeast reports whether the effective subject role is at least want.
// Supervisors (org admin) do not invent a subject role here — action
// helpers (CanContributeAccess / CanLeadAccess) apply org-admin write rules.
func (a SubjectAccess) RoleAtLeast(want string) bool {
	return subjectRoleRank(a.Role) >= subjectRoleRank(want)
}

// ResolveSubjectAccess computes access for userID on subjectID in the active org.
// globalRole is ignored (kept for call-site compatibility); product access is
// org owner/admin inheritance + subject_members / team grants.
//
// Org members without a grant do not see the subject. Org owner/admin see all.
func (s *Store) ResolveSubjectAccess(ctx context.Context, userID, subjectID int64, globalRole string) (SubjectAccess, error) {
	_ = globalRole
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return SubjectAccess{}, err
	}

	subject, err := s.SubjectByID(ctx, subjectID)
	if err != nil {
		if errors.Is(err, ErrSubjectNotFound) {
			return SubjectAccess{}, nil
		}
		return SubjectAccess{}, err
	}
	if subject.OrganizationID != orgID {
		return SubjectAccess{}, nil
	}

	orgRole, isMember, err := s.OrganizationMemberRole(ctx, orgID, userID)
	if err != nil {
		return SubjectAccess{}, err
	}

	var access SubjectAccess
	directRole, err := s.directSubjectMemberRole(ctx, subjectID, userID)
	if err != nil {
		return SubjectAccess{}, err
	}
	if directRole != "" {
		access.Role = directRole
		access.Sources = append(access.Sources, AccessSourceDirect)
	}

	teamRoles, err := s.userTeamSubjectRoles(ctx, orgID, userID, subjectID)
	if err != nil {
		return SubjectAccess{}, err
	}
	for _, tr := range teamRoles {
		access.Role = maxSubjectRole(access.Role, tr.Role)
		access.Sources = append(access.Sources, accessSourceTeamPrefix+strconv.FormatInt(tr.TeamID, 10))
	}

	// Org owner/admin: always visible for supervision. Keep any direct/team role
	// so lead actions are not implied by org_admin alone.
	if isMember && (orgRole == OrgRoleOwner || orgRole == OrgRoleAdmin) {
		access.Visible = true
		access.Sources = append([]string{AccessSourceOrgAdmin}, access.Sources...)
		return access, nil
	}

	if access.Role != "" {
		access.Visible = true
		return access, nil
	}

	return SubjectAccess{}, nil
}

type teamSubjectRoleRow struct {
	TeamID int64
	Role   string
}

func (s *Store) directSubjectMemberRole(ctx context.Context, subjectID, userID int64) (string, error) {
	var role string
	err := s.db.QueryRowContext(ctx, `
		SELECT role FROM subject_members
		WHERE subject_id = ? AND user_id = ?
	`, subjectID, userID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("direct subject member role: %w", err)
	}
	return role, nil
}

func (s *Store) userTeamSubjectRoles(ctx context.Context, orgID, userID, subjectID int64) ([]teamSubjectRoleRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.team_id, r.role
		FROM team_subject_roles r
		INNER JOIN organization_teams t ON t.id = r.team_id
		INNER JOIN team_members m ON m.team_id = r.team_id
		WHERE r.subject_id = ?
		  AND m.user_id = ?
		  AND t.organization_id = ?
	`, subjectID, userID, orgID)
	if err != nil {
		return nil, fmt.Errorf("user team subject roles: %w", err)
	}
	defer rows.Close()

	var out []teamSubjectRoleRow
	for rows.Next() {
		var row teamSubjectRoleRow
		if scanErr := rows.Scan(&row.TeamID, &row.Role); scanErr != nil {
			return nil, fmt.Errorf("scan team subject role: %w", scanErr)
		}
		out = append(out, row)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("user team subject roles rows: %w", err)
	}
	return out, nil
}

// subjectVisibleToOrgMemberSQL is an AND fragment for runs/dashboard listings.
// subjectAlias is the subjects table alias (e.g. "p"). Requires organization_members
// joined as "om". Bind args after the fragment: userID, userID, orgID.
func subjectVisibleToOrgMemberSQL(subjectAlias string) string {
	return `
		AND (
			om.role IN ('` + OrgRoleOwner + `', '` + OrgRoleAdmin + `')
			OR EXISTS (
				SELECT 1 FROM subject_members sm
				WHERE sm.subject_id = ` + subjectAlias + `.id AND sm.user_id = ?
			)
			OR EXISTS (
				SELECT 1 FROM team_subject_roles tsr
				INNER JOIN team_members tm ON tm.team_id = tsr.team_id
				INNER JOIN organization_teams ot ON ot.id = tsr.team_id
				WHERE tsr.subject_id = ` + subjectAlias + `.id AND tm.user_id = ? AND ot.organization_id = ?
			)
		)`
}

// subjectVisibleToListedMemberSQL is like subjectVisibleToOrgMemberSQL but uses
// om.user_id / subject.organization_id from the row (no extra bind args).
// Used when listing candidate members for a single subject.
func subjectVisibleToListedMemberSQL(subjectAlias string) string {
	return `
		AND (
			om.role IN ('` + OrgRoleOwner + `', '` + OrgRoleAdmin + `')
			OR EXISTS (
				SELECT 1 FROM subject_members sm
				WHERE sm.subject_id = ` + subjectAlias + `.id AND sm.user_id = om.user_id
			)
			OR EXISTS (
				SELECT 1 FROM team_subject_roles tsr
				INNER JOIN team_members tm ON tm.team_id = tsr.team_id
				INNER JOIN organization_teams ot ON ot.id = tsr.team_id
				WHERE tsr.subject_id = ` + subjectAlias + `.id
				  AND tm.user_id = om.user_id
				  AND ot.organization_id = ` + subjectAlias + `.organization_id
			)
		)`
}

func subjectRoleRank(role string) int {
	switch role {
	case SubjectRoleLead:
		return 3
	case SubjectRoleContributor:
		return 2
	case SubjectRoleViewer:
		return 1
	default:
		return 0
	}
}

func maxSubjectRole(a, b string) string {
	if subjectRoleRank(b) > subjectRoleRank(a) {
		return b
	}
	return a
}
