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
	ID             int64
	OrganizationID int64
	Name           string
	ArchivedAt     sql.NullString
	CreatedAt      string
}

// ChecklistTemplateSummary includes latest version metadata for listings.
type ChecklistTemplateSummary struct {
	ChecklistTemplate
	LatestVersion int
	ItemCount     int
	Tags          []string
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

// CreateChecklistTemplate inserts a global template with version 1, items and tags.
func (s *Store) CreateChecklistTemplate(ctx context.Context, name string, createdBy int64, tags []string, items []TemplateItemInput) (*ChecklistTemplate, *TemplateVersion, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	tags = NormalizeTags(tags)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO checklist_templates (organization_id, name, created_at)
		VALUES (?, ?, ?)
	`, orgID, name, now)
	if err != nil {
		return nil, nil, fmt.Errorf("insert checklist template: %w", err)
	}

	templateID, err := res.LastInsertId()
	if err != nil {
		return nil, nil, fmt.Errorf("template id: %w", err)
	}

	if err = setTemplateDomainsTx(ctx, tx, templateID, tags); err != nil {
		return nil, nil, err
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
		SELECT t.id, t.organization_id, t.name, t.archived_at, t.created_at
		FROM checklist_templates t
		WHERE t.id = ? AND t.organization_id = ?
	`, id, orgID).Scan(&t.ID, &t.OrganizationID, &t.Name, &t.ArchivedAt, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrChecklistTemplateNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("checklist template by id: %w", err)
	}
	return &t, nil
}

// ListChecklistTemplates returns active templates matching a subject's domains.
func (s *Store) ListChecklistTemplates(ctx context.Context, subjectID int64) ([]ChecklistTemplateSummary, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			t.id, t.organization_id, t.name, t.archived_at, t.created_at,
			v.version,
			COUNT(i.id) AS item_count
		FROM checklist_templates t
		INNER JOIN template_versions v ON v.template_id = t.id
		LEFT JOIN template_items i ON i.version_id = v.id
		WHERE t.archived_at IS NULL AND t.organization_id = ?
		  AND v.version = (
			SELECT MAX(v2.version) FROM template_versions v2 WHERE v2.template_id = t.id
		  )
		  AND (
			NOT EXISTS (SELECT 1 FROM template_domains td WHERE td.template_id = t.id)
			OR EXISTS (
				SELECT 1 FROM template_domains td
				INNER JOIN subject_domains sd ON sd.tag = td.tag AND sd.subject_id = ?
				WHERE td.template_id = t.id
			)
		  )
		GROUP BY t.id, t.organization_id, t.name, t.archived_at, t.created_at, v.version
		ORDER BY t.name
	`, orgID, subjectID)
	if err != nil {
		return nil, fmt.Errorf("list checklist templates: %w", err)
	}
	defer rows.Close()

	var templates []ChecklistTemplateSummary
	for rows.Next() {
		var summary ChecklistTemplateSummary
		if err := rows.Scan(
			&summary.ID, &summary.OrganizationID, &summary.Name, &summary.ArchivedAt, &summary.CreatedAt,
			&summary.LatestVersion, &summary.ItemCount,
		); err != nil {
			return nil, fmt.Errorf("scan checklist template: %w", err)
		}
		templates = append(templates, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate checklist templates: %w", err)
	}

	for i := range templates {
		tags, err := s.ListTemplateDomains(ctx, templates[i].ID)
		if err != nil {
			return nil, err
		}
		templates[i].Tags = tags
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
		WHERE id = ? AND archived_at IS NULL AND organization_id = ?
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
		WHERE id = ? AND archived_at IS NULL AND organization_id = ?
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

// ErrTemplateVersionNotFound is returned when a template version does not exist
// in the active organization (or the parent template is missing).
var ErrTemplateVersionNotFound = errors.New("template version not found")

// ErrPublishedVersionImmutable is returned when a caller attempts a structural
// mutation of an already published template version. Structural changes must
// go through CreateTemplateVersion. Editorial help_text updates on the latest
// version use UpdateTemplateItemsHelpText instead.
var ErrPublishedVersionImmutable = errors.New("published template version is immutable")

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

// ListTemplateVersions returns all versions of a template (newest first), scoped to the active org.
func (s *Store) ListTemplateVersions(ctx context.Context, templateID int64) ([]TemplateVersion, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT v.id, v.template_id, v.version, v.published_at, v.created_by
		FROM template_versions v
		INNER JOIN checklist_templates t ON t.id = v.template_id
		WHERE v.template_id = ? AND t.organization_id = ?
		ORDER BY v.version DESC
	`, templateID, orgID)
	if err != nil {
		return nil, fmt.Errorf("list template versions: %w", err)
	}
	defer rows.Close()

	var versions []TemplateVersion
	for rows.Next() {
		var v TemplateVersion
		if err := rows.Scan(&v.ID, &v.TemplateID, &v.Version, &v.PublishedAt, &v.CreatedBy); err != nil {
			return nil, fmt.Errorf("scan template version: %w", err)
		}
		versions = append(versions, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate template versions: %w", err)
	}
	return versions, nil
}

// TemplateVersionByNumber loads a version by template id + version number in the active org.
func (s *Store) TemplateVersionByNumber(ctx context.Context, templateID int64, versionNum int) (*TemplateVersion, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var v TemplateVersion
	err = s.db.QueryRowContext(ctx, `
		SELECT v.id, v.template_id, v.version, v.published_at, v.created_by
		FROM template_versions v
		INNER JOIN checklist_templates t ON t.id = v.template_id
		WHERE v.template_id = ? AND v.version = ? AND t.organization_id = ?
	`, templateID, versionNum, orgID).Scan(&v.ID, &v.TemplateID, &v.Version, &v.PublishedAt, &v.CreatedBy)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTemplateVersionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("template version by number: %w", err)
	}
	return &v, nil
}

// ReplaceTemplateItems is intentionally unsupported for structural rewrites.
// Callers must use CreateTemplateVersion for structural changes, or
// UpdateTemplateItemsHelpText for editorial help_text-only updates.
func (s *Store) ReplaceTemplateItems(_ context.Context, _ int64, _ []TemplateItemInput) error {
	return ErrPublishedVersionImmutable
}

// UpdateTemplateItemsHelpText updates only help_text on an existing version.
// Items must match the current structure (count, order, label, section, required).
func (s *Store) UpdateTemplateItemsHelpText(ctx context.Context, versionID int64, items []TemplateItemInput) error {
	current, err := s.ListTemplateItems(ctx, versionID)
	if err != nil {
		return fmt.Errorf("list items for editorial update: %w", err)
	}
	if len(current) != len(items) {
		return ErrPublishedVersionImmutable
	}
	for i := range current {
		if current[i].Label != items[i].Label ||
			current[i].Section != items[i].Section ||
			current[i].Required != items[i].Required {
			return ErrPublishedVersionImmutable
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for i, item := range items {
		res, execErr := tx.ExecContext(ctx, `
			UPDATE template_items SET help_text = ? WHERE id = ? AND version_id = ?
		`, item.HelpText, current[i].ID, versionID)
		if execErr != nil {
			return fmt.Errorf("update help_text: %w", execErr)
		}
		n, _ := res.RowsAffected()
		if n != 1 {
			return fmt.Errorf("update help_text: expected 1 row, got %d", n)
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit editorial update: %w", err)
	}
	return nil
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
		WHERE v.id = ? AND t.organization_id = ?
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
