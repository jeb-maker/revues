package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

const maxTagLen = 64

// NormalizeTags trims, lowercases, deduplicates and drops empty tag strings.
func NormalizeTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, raw := range tags {
		tag := normalizeTag(raw)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}

// ParseTagsCSV splits a comma-separated tag field.
func ParseTagsCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	return NormalizeTags(parts)
}

// FormatTagsCSV joins tags for form display.
func FormatTagsCSV(tags []string) string {
	return strings.Join(tags, ", ")
}

func normalizeTag(raw string) string {
	tag := strings.ToLower(strings.TrimSpace(raw))
	if len(tag) > maxTagLen {
		tag = tag[:maxTagLen]
	}
	return tag
}

// ListProjectTags returns tags for a project ordered alphabetically.
func (s *Store) ListProjectTags(ctx context.Context, projectID int64) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT tag FROM project_tags WHERE project_id = ? ORDER BY tag
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("scan project tag: %w", err)
		}
		tags = append(tags, tag)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project tags: %w", err)
	}
	return tags, nil
}

// SetProjectTags replaces all tags on a project.
func (s *Store) SetProjectTags(ctx context.Context, projectID int64, tags []string) error {
	tags = NormalizeTags(tags)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, `DELETE FROM project_tags WHERE project_id = ?`, projectID); err != nil {
		return fmt.Errorf("delete project tags: %w", err)
	}
	for _, tag := range tags {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO project_tags (project_id, tag) VALUES (?, ?)
		`, projectID, tag); err != nil {
			return fmt.Errorf("insert project tag: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit project tags: %w", err)
	}
	return nil
}

// ListTemplateTags returns tags for a template ordered alphabetically.
func (s *Store) ListTemplateTags(ctx context.Context, templateID int64) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT tag FROM template_tags WHERE template_id = ? ORDER BY tag
	`, templateID)
	if err != nil {
		return nil, fmt.Errorf("list template tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("scan template tag: %w", err)
		}
		tags = append(tags, tag)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate template tags: %w", err)
	}
	return tags, nil
}

// SetTemplateTags replaces all tags on a template.
func (s *Store) SetTemplateTags(ctx context.Context, templateID int64, tags []string) error {
	tags = NormalizeTags(tags)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, `DELETE FROM template_tags WHERE template_id = ?`, templateID); err != nil {
		return fmt.Errorf("delete template tags: %w", err)
	}
	for _, tag := range tags {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO template_tags (template_id, tag) VALUES (?, ?)
		`, templateID, tag); err != nil {
			return fmt.Errorf("insert template tag: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit template tags: %w", err)
	}
	return nil
}

// TemplateMatchesProject reports whether a template is eligible for a project.
// Untagged templates match all projects; otherwise at least one shared tag is required.
func (s *Store) TemplateMatchesProject(ctx context.Context, projectID, templateID int64) (bool, error) {
	var templateTagCount int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM template_tags WHERE template_id = ?
	`, templateID).Scan(&templateTagCount)
	if err != nil {
		return false, fmt.Errorf("count template tags: %w", err)
	}
	if templateTagCount == 0 {
		return true, nil
	}

	var shared int
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM template_tags tt
		INNER JOIN project_tags pt ON pt.tag = tt.tag AND pt.project_id = ?
		WHERE tt.template_id = ?
	`, projectID, templateID).Scan(&shared)
	if err != nil {
		return false, fmt.Errorf("count shared tags: %w", err)
	}
	return shared > 0, nil
}

func setProjectTagsTx(ctx context.Context, tx *sql.Tx, projectID int64, tags []string) error {
	tags = NormalizeTags(tags)
	if _, err := tx.ExecContext(ctx, `DELETE FROM project_tags WHERE project_id = ?`, projectID); err != nil {
		return fmt.Errorf("delete project tags: %w", err)
	}
	for _, tag := range tags {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO project_tags (project_id, tag) VALUES (?, ?)
		`, projectID, tag); err != nil {
			return fmt.Errorf("insert project tag: %w", err)
		}
	}
	return nil
}

func setTemplateTagsTx(ctx context.Context, tx *sql.Tx, templateID int64, tags []string) error {
	tags = NormalizeTags(tags)
	if _, err := tx.ExecContext(ctx, `DELETE FROM template_tags WHERE template_id = ?`, templateID); err != nil {
		return fmt.Errorf("delete template tags: %w", err)
	}
	for _, tag := range tags {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO template_tags (template_id, tag) VALUES (?, ?)
		`, templateID, tag); err != nil {
			return fmt.Errorf("insert template tag: %w", err)
		}
	}
	return nil
}
