package store

import (
	"context"
	"fmt"
)

// RunExportRow is one CSV line for a closed review export.
type RunExportRow struct {
	SubjectName string
	RunTitle    string
	RunDate     string
	PointLabel  string
	Status      string
	Comment     string
	AuthorLogin string
}

// ListRunExportRows returns ordered export rows for a run.
func (s *Store) ListRunExportRows(ctx context.Context, runID int64) ([]RunExportRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.name, t.name, r.created_at, r.id, COALESCE(r.completed_at, ''), ri.label, ri.status, ri.comment, COALESCE(u.login, '')
		FROM run_items ri
		INNER JOIN checklist_runs r ON r.id = ri.run_id
		INNER JOIN subjects p ON p.id = r.subject_id
		INNER JOIN template_versions tv ON tv.id = r.template_version_id
		INNER JOIN checklist_templates t ON t.id = tv.template_id
		LEFT JOIN users u ON u.id = ri.checked_by
		WHERE ri.run_id = ?
		ORDER BY ri.position
	`, runID)
	if err != nil {
		return nil, fmt.Errorf("list run export rows: %w", err)
	}
	defer rows.Close()

	var exportRows []RunExportRow
	for rows.Next() {
		var row RunExportRow
		var templateName, createdAt string
		var runID int64
		if scanErr := rows.Scan(
			&row.SubjectName, &templateName, &createdAt, &runID, &row.RunDate, &row.PointLabel,
			&row.Status, &row.Comment, &row.AuthorLogin,
		); scanErr != nil {
			return nil, fmt.Errorf("scan run export row: %w", scanErr)
		}
		row.RunTitle = RunDisplayLabel(templateName, row.SubjectName, createdAt, runID)
		exportRows = append(exportRows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate run export rows: %w", err)
	}

	return exportRows, nil
}
