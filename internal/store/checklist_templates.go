package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrChecklistTemplateNotFound is returned when a template id does not exist.
var ErrChecklistTemplateNotFound = errors.New("checklist template not found")

// ChecklistTemplate is a versioned checklist model container.
type ChecklistTemplate struct {
	ID         int64
	ProjectID  int64
	Name       string
	ArchivedAt sql.NullString
	CreatedAt  string
}

// ChecklistTemplateSummary includes latest version metadata for listings.
type ChecklistTemplateSummary struct {
	ChecklistTemplate
	LatestVersion int
	ItemCount     int
}

// TemplateVersion is an immutable snapshot of template content.
type TemplateVersion struct {
	ID          int64
	TemplateID  int64
	Version     int
	PublishedAt string
	CreatedBy   sql.NullInt64
}

// TemplateItem is an ordered checklist point within a version.
type TemplateItem struct {
	ID        int64
	VersionID int64
	Section   string
	Position  int
	Label     string
	HelpText  string
	Required  bool
}

// TemplateItemInput is input for creating template items.
type TemplateItemInput struct {
	Section  string
	Label    string
	HelpText string
	Required bool
}

// CreateChecklistTemplate inserts a template with version 1 and its items.
func (s *Store) CreateChecklistTemplate(ctx context.Context, projectID int64, name string, createdBy int64, items []TemplateItemInput) (*ChecklistTemplate, *TemplateVersion, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO checklist_templates (project_id, name, created_at)
		VALUES (?, ?, ?)
	`, projectID, name, now)
	if err != nil {
		return nil, nil, fmt.Errorf("insert checklist template: %w", err)
	}

	templateID, err := res.LastInsertId()
	if err != nil {
		return nil, nil, fmt.Errorf("template id: %w", err)
	}

	version, err := insertTemplateVersionTx(ctx, tx, templateID, 1, now, createdBy, items)
	if err != nil {
		return nil, nil, err
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return nil, nil, fmt.Errorf("commit create checklist template: %w", commitErr)
	}

	template, err := s.ChecklistTemplateByID(ctx, templateID)
	if err != nil {
		return nil, nil, err
	}

	return template, version, nil
}

// ChecklistTemplateByID loads a template by primary key in the active organization.
func (s *Store) ChecklistTemplateByID(ctx context.Context, id int64) (*ChecklistTemplate, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var t ChecklistTemplate
	err = s.db.QueryRowContext(ctx, `
		SELECT t.id, t.project_id, t.name, t.archived_at, t.created_at
		FROM checklist_templates t
		INNER JOIN projects p ON p.id = t.project_id
		WHERE t.id = ? AND p.organization_id = ?
	`, id, orgID).Scan(&t.ID, &t.ProjectID, &t.Name, &t.ArchivedAt, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrChecklistTemplateNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("checklist template by id: %w", err)
	}
	return &t, nil
}

// ListChecklistTemplates returns active templates for a project with latest version info.
func (s *Store) ListChecklistTemplates(ctx context.Context, projectID int64) ([]ChecklistTemplateSummary, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			t.id, t.project_id, t.name, t.archived_at, t.created_at,
			v.version,
			COUNT(i.id) AS item_count
		FROM checklist_templates t
		INNER JOIN projects p ON p.id = t.project_id
		INNER JOIN template_versions v ON v.template_id = t.id
		LEFT JOIN template_items i ON i.version_id = v.id
		WHERE t.project_id = ? AND p.organization_id = ? AND t.archived_at IS NULL
		  AND v.version = (
			SELECT MAX(v2.version) FROM template_versions v2 WHERE v2.template_id = t.id
		  )
		GROUP BY t.id, t.project_id, t.name, t.archived_at, t.created_at, v.version
		ORDER BY t.name
	`, projectID, orgID)
	if err != nil {
		return nil, fmt.Errorf("list checklist templates: %w", err)
	}
	defer rows.Close()

	var templates []ChecklistTemplateSummary
	for rows.Next() {
		var summary ChecklistTemplateSummary
		if err := rows.Scan(
			&summary.ID, &summary.ProjectID, &summary.Name, &summary.ArchivedAt, &summary.CreatedAt,
			&summary.LatestVersion, &summary.ItemCount,
		); err != nil {
			return nil, fmt.Errorf("scan checklist template: %w", err)
		}
		templates = append(templates, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate checklist templates: %w", err)
	}

	return templates, nil
}

// UpdateChecklistTemplateName changes the display name of a template.
func (s *Store) UpdateChecklistTemplateName(ctx context.Context, id int64, name string) error {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	res, err := s.db.ExecContext(ctx, `
		UPDATE checklist_templates
		SET name = ?
		WHERE id = ? AND archived_at IS NULL
		  AND project_id IN (SELECT id FROM projects WHERE organization_id = ?)
	`, name, id, orgID)
	if err != nil {
		return fmt.Errorf("update checklist template name: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update checklist template rows: %w", err)
	}
	if n == 0 {
		return ErrChecklistTemplateNotFound
	}
	return nil
}

// ArchiveChecklistTemplate marks a template archived.
func (s *Store) ArchiveChecklistTemplate(ctx context.Context, id int64) error {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, `
		UPDATE checklist_templates
		SET archived_at = ?
		WHERE id = ? AND archived_at IS NULL
		  AND project_id IN (SELECT id FROM projects WHERE organization_id = ?)
	`, now, id, orgID)
	if err != nil {
		return fmt.Errorf("archive checklist template: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive checklist template rows: %w", err)
	}
	if n == 0 {
		return ErrChecklistTemplateNotFound
	}
	return nil
}

// LatestTemplateVersion returns the highest version for a template.
func (s *Store) LatestTemplateVersion(ctx context.Context, templateID int64) (*TemplateVersion, error) {
	var v TemplateVersion
	err := s.db.QueryRowContext(ctx, `
		SELECT id, template_id, version, published_at, created_by
		FROM template_versions
		WHERE template_id = ?
		ORDER BY version DESC
		LIMIT 1
	`, templateID).Scan(&v.ID, &v.TemplateID, &v.Version, &v.PublishedAt, &v.CreatedBy)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("latest template version: %w", err)
	}
	return &v, nil
}

// ListTemplateItems returns ordered items for a version.
func (s *Store) ListTemplateItems(ctx context.Context, versionID int64) ([]TemplateItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, version_id, section, position, label, help_text, required
		FROM template_items
		WHERE version_id = ?
		ORDER BY position
	`, versionID)
	if err != nil {
		return nil, fmt.Errorf("list template items: %w", err)
	}
	defer rows.Close()

	var items []TemplateItem
	for rows.Next() {
		var item TemplateItem
		var required int
		if err := rows.Scan(&item.ID, &item.VersionID, &item.Section, &item.Position, &item.Label, &item.HelpText, &required); err != nil {
			return nil, fmt.Errorf("scan template item: %w", err)
		}
		item.Required = required == 1
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate template items: %w", err)
	}

	return items, nil
}

// CreateTemplateVersion appends a new version with items (never mutates prior versions).
func (s *Store) CreateTemplateVersion(ctx context.Context, templateID, createdBy int64, items []TemplateItemInput) (*TemplateVersion, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var nextVersion int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0) + 1 FROM template_versions WHERE template_id = ?
	`, templateID).Scan(&nextVersion)
	if err != nil {
		return nil, fmt.Errorf("next template version: %w", err)
	}

	version, err := insertTemplateVersionTx(ctx, tx, templateID, nextVersion, now, createdBy, items)
	if err != nil {
		return nil, err
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return nil, fmt.Errorf("commit template version: %w", commitErr)
	}

	return version, nil
}

// TemplateVersionInfo links a version to its template metadata.
type TemplateVersionInfo struct {
	TemplateID int64
	Name       string
	Version    int
}

// TemplateVersionInfo loads template metadata for a version id in the active organization.
func (s *Store) TemplateVersionInfo(ctx context.Context, versionID int64) (*TemplateVersionInfo, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var info TemplateVersionInfo
	err = s.db.QueryRowContext(ctx, `
		SELECT t.id, t.name, v.version
		FROM template_versions v
		INNER JOIN checklist_templates t ON t.id = v.template_id
		INNER JOIN projects p ON p.id = t.project_id
		WHERE v.id = ? AND p.organization_id = ?
	`, versionID, orgID).Scan(&info.TemplateID, &info.Name, &info.Version)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("template version info: %w", err)
	}
	return &info, nil
}

func insertTemplateVersionTx(ctx context.Context, tx *sql.Tx, templateID int64, versionNum int, now string, createdBy int64, items []TemplateItemInput) (*TemplateVersion, error) {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO template_versions (template_id, version, published_at, created_by)
		VALUES (?, ?, ?, ?)
	`, templateID, versionNum, now, createdBy)
	if err != nil {
		return nil, fmt.Errorf("insert template version: %w", err)
	}

	versionID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("template version id: %w", err)
	}

	for i, item := range items {
		required := 0
		if item.Required {
			required = 1
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO template_items (version_id, section, position, label, help_text, required)
			VALUES (?, ?, ?, ?, ?, ?)
		`, versionID, item.Section, i+1, item.Label, item.HelpText, required)
		if err != nil {
			return nil, fmt.Errorf("insert template item: %w", err)
		}
	}

	return &TemplateVersion{
		ID:          versionID,
		TemplateID:  templateID,
		Version:     versionNum,
		PublishedAt: now,
		CreatedBy:   sql.NullInt64{Int64: createdBy, Valid: true},
	}, nil
}
