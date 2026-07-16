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

// ListSubjectTags returns descriptive labels for a subject ordered alphabetically.
func (s *Store) ListSubjectTags(ctx context.Context, subjectID int64) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT tag FROM subject_tags WHERE subject_id = ? ORDER BY tag
	`, subjectID)
	if err != nil {
		return nil, fmt.Errorf("list subject tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("scan subject tag: %w", err)
		}
		tags = append(tags, tag)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subject tags: %w", err)
	}
	return tags, nil
}

// SetSubjectTags replaces all descriptive labels on a subject.
func (s *Store) SetSubjectTags(ctx context.Context, subjectID int64, tags []string) error {
	tags = NormalizeTags(tags)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, `DELETE FROM subject_tags WHERE subject_id = ?`, subjectID); err != nil {
		return fmt.Errorf("delete subject tags: %w", err)
	}
	for _, tag := range tags {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO subject_tags (subject_id, tag) VALUES (?, ?)
		`, subjectID, tag); err != nil {
			return fmt.Errorf("insert subject tag: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit subject tags: %w", err)
	}
	return nil
}

// ListSubjectDomains returns matching domains for a subject ordered alphabetically.
func (s *Store) ListSubjectDomains(ctx context.Context, subjectID int64) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT tag FROM subject_domains WHERE subject_id = ? ORDER BY tag
	`, subjectID)
	if err != nil {
		return nil, fmt.Errorf("list subject domains: %w", err)
	}
	defer rows.Close()

	var domains []string
	for rows.Next() {
		var domain string
		if err := rows.Scan(&domain); err != nil {
			return nil, fmt.Errorf("scan subject domain: %w", err)
		}
		domains = append(domains, domain)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subject domains: %w", err)
	}
	return domains, nil
}

// SetSubjectDomains replaces all matching domains on a subject.
func (s *Store) SetSubjectDomains(ctx context.Context, subjectID int64, domains []string) error {
	tags := NormalizeTags(domains)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, `DELETE FROM subject_domains WHERE subject_id = ?`, subjectID); err != nil {
		return fmt.Errorf("delete subject domains: %w", err)
	}
	for _, tag := range tags {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO subject_domains (subject_id, tag) VALUES (?, ?)
		`, subjectID, tag); err != nil {
			return fmt.Errorf("insert subject domain: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit subject domains: %w", err)
	}
	return nil
}

// ListTemplateDomains returns matching domains for a template ordered alphabetically.
func (s *Store) ListTemplateDomains(ctx context.Context, templateID int64) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT tag FROM template_domains WHERE template_id = ? ORDER BY tag
	`, templateID)
	if err != nil {
		return nil, fmt.Errorf("list template domains: %w", err)
	}
	defer rows.Close()

	var domains []string
	for rows.Next() {
		var domain string
		if err := rows.Scan(&domain); err != nil {
			return nil, fmt.Errorf("scan template domain: %w", err)
		}
		domains = append(domains, domain)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate template domains: %w", err)
	}
	return domains, nil
}

// SetTemplateDomains replaces all matching domains on a template.
func (s *Store) SetTemplateDomains(ctx context.Context, templateID int64, domains []string) error {
	tags := NormalizeTags(domains)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, `DELETE FROM template_domains WHERE template_id = ?`, templateID); err != nil {
		return fmt.Errorf("delete template domains: %w", err)
	}
	for _, tag := range tags {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO template_domains (template_id, tag) VALUES (?, ?)
		`, templateID, tag); err != nil {
			return fmt.Errorf("insert template domain: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit template domains: %w", err)
	}
	return nil
}

// TemplateMatchesSubject reports whether a template is eligible for a subject.
// Templates without domains match all subjects; otherwise at least one shared domain is required.
func (s *Store) TemplateMatchesSubject(ctx context.Context, subjectID, templateID int64) (bool, error) {
	var templateDomainCount int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM template_domains WHERE template_id = ?
	`, templateID).Scan(&templateDomainCount)
	if err != nil {
		return false, fmt.Errorf("count template domains: %w", err)
	}
	if templateDomainCount == 0 {
		return true, nil
	}

	var shared int
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM template_domains td
		INNER JOIN subject_domains sd ON sd.tag = td.tag AND sd.subject_id = ?
		WHERE td.template_id = ?
	`, subjectID, templateID).Scan(&shared)
	if err != nil {
		return false, fmt.Errorf("count shared domains: %w", err)
	}
	return shared > 0, nil
}

func setSubjectDomainsTx(ctx context.Context, tx *sql.Tx, subjectID int64, domains []string) error {
	tags := NormalizeTags(domains)
	if _, err := tx.ExecContext(ctx, `DELETE FROM subject_domains WHERE subject_id = ?`, subjectID); err != nil {
		return fmt.Errorf("delete subject domains: %w", err)
	}
	for _, tag := range tags {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO subject_domains (subject_id, tag) VALUES (?, ?)
		`, subjectID, tag); err != nil {
			return fmt.Errorf("insert subject domain: %w", err)
		}
	}
	return nil
}

func setTemplateDomainsTx(ctx context.Context, tx *sql.Tx, templateID int64, domains []string) error {
	tags := NormalizeTags(domains)
	if _, err := tx.ExecContext(ctx, `DELETE FROM template_domains WHERE template_id = ?`, templateID); err != nil {
		return fmt.Errorf("delete template domains: %w", err)
	}
	for _, tag := range tags {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO template_domains (template_id, tag) VALUES (?, ?)
		`, templateID, tag); err != nil {
			return fmt.Errorf("insert template domain: %w", err)
		}
	}
	return nil
}

// Deprecated tag/domain aliases.

func (s *Store) ListProjectTags(ctx context.Context, subjectID int64) ([]string, error) {
	return s.ListSubjectDomains(ctx, subjectID)
}

func (s *Store) SetProjectTags(ctx context.Context, subjectID int64, domains []string) error {
	return s.SetSubjectDomains(ctx, subjectID, domains)
}

func (s *Store) ListTemplateTags(ctx context.Context, templateID int64) ([]string, error) {
	return s.ListTemplateDomains(ctx, templateID)
}

func (s *Store) SetTemplateTags(ctx context.Context, templateID int64, domains []string) error {
	return s.SetTemplateDomains(ctx, templateID, domains)
}

func (s *Store) TemplateMatchesProject(ctx context.Context, subjectID, templateID int64) (bool, error) {
	return s.TemplateMatchesSubject(ctx, subjectID, templateID)
}
