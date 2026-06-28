package runs

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/jeb-maker/revues/internal/store"
)

var runExportHeaders = []string{"projet", "revue", "date", "points", "statuts", "commentaires", "auteur"}

// BuildRunCSV encodes export rows as CSV with a header row.
func BuildRunCSV(rows []store.RunExportRow) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if err := w.Write(runExportHeaders); err != nil {
		return nil, fmt.Errorf("write csv header: %w", err)
	}

	for _, row := range rows {
		record := []string{
			row.ProjectName,
			row.RunTitle,
			row.RunDate,
			row.PointLabel,
			row.Status,
			row.Comment,
			row.AuthorLogin,
		}
		if err := w.Write(record); err != nil {
			return nil, fmt.Errorf("write csv row: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("flush csv: %w", err)
	}

	return buf.Bytes(), nil
}
