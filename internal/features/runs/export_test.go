package runs_test

import (
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/store"
)

func TestBuildRunCSV(t *testing.T) {
	t.Parallel()

	data, err := runs.BuildRunCSV([]store.RunExportRow{
		{
			SubjectName: "Alpha",
			RunTitle:    "Revue Q2",
			RunDate:     "2025-06-01T10:00:00Z",
			PointLabel:  "Backup OK",
			Status:      "ok",
			Comment:     "",
			AuthorLogin: "marie",
		},
		{
			SubjectName: "Alpha",
			RunTitle:    "Revue Q2",
			RunDate:     "2025-06-01T10:00:00Z",
			PointLabel:  "Logs",
			Status:      "nok",
			Comment:     "Manque rotation",
			AuthorLogin: "thomas",
		},
	})
	if err != nil {
		t.Fatalf("BuildRunCSV() error = %v", err)
	}

	csv := string(data)
	//nolint:misspell // French CSV column headers per issue #31
	if !strings.HasPrefix(csv, "subject,revue,date,points,statuts,commentaires,auteur\n") {
		t.Fatalf("unexpected header: %q", csv)
	}
	if !strings.Contains(csv, "Alpha,Revue Q2,2025-06-01T10:00:00Z,Backup OK,ok,,marie") {
		t.Fatalf("missing first row in %q", csv)
	}
	if !strings.Contains(csv, "Manque rotation") {
		t.Fatalf("missing comment in %q", csv)
	}
}
