package store

import (
	"context"
	"fmt"
)

// RunExportRow is one CSV line for a closed review export.
type RunExportRow struct {
	ProjectName string
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
		SELECT p.name, cr.title, COALESCE(cr.completed_at, ''), ri.label, ri.status, ri.comment, COALESCE(u.login, '')
		FROM run_items ri
		INNER JOIN checklist_runs cr ON cr.id = ri.run_id
		INNER JOIN projects p ON p.id = cr.project_id
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
		if scanErr := rows.Scan(
			&row.ProjectName, &row.RunTitle, &row.RunDate, &row.PointLabel,
			&row.Status, &row.Comment, &row.AuthorLogin,
		); scanErr != nil {
			return nil, fmt.Errorf("scan run export row: %w", scanErr)
		}
		exportRows = append(exportRows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate run export rows: %w", err)
	}

	return exportRows, nil
}
