package runs_test

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/store"
)

func TestSHA256Hex(t *testing.T) {
	got := runs.SHA256Hex([]byte("hello"))
	want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if got != want {
		t.Fatalf("SHA256Hex() = %q, want %q", got, want)
	}
}

func TestBuildEvidenceZIP(t *testing.T) {
	csvData := []byte("subject,revue\nAlpha,Revue\n")
	hash := runs.SHA256Hex(csvData)
	manifest := runs.EvidenceManifest{
		RunID:        42,
		SubjectName:  "Alpha",
		TemplateName: "Release",
		Version:      1,
		Status:       store.RunStatusDone,
		CompletedAt:  "2026-07-16T12:00:00Z",
		ClosedBy:     "alice",
		CSVSHA256:    hash,
		GeneratedAt:  "2026-07-16T12:00:00Z",
	}
	zipBytes, err := runs.BuildEvidenceZIP(42, csvData, manifest)
	if err != nil {
		t.Fatalf("BuildEvidenceZIP(): %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatalf("zip.NewReader(): %v", err)
	}
	files := map[string][]byte{}
	for _, f := range zr.File {
		rc, openErr := f.Open()
		if openErr != nil {
			t.Fatalf("open %s: %v", f.Name, openErr)
		}
		data, readErr := io.ReadAll(rc)
		_ = rc.Close()
		if readErr != nil {
			t.Fatalf("read %s: %v", f.Name, readErr)
		}
		files[f.Name] = data
	}

	csvPath := "revue-42/revue.csv"
	if !bytes.Equal(files[csvPath], csvData) {
		t.Fatalf("csv = %q, want %q", files[csvPath], csvData)
	}
	sumPath := "revue-42/sha256sum.txt"
	if got := string(files[sumPath]); got != hash+"  revue.csv\n" {
		t.Fatalf("sha256sum.txt = %q", got)
	}
	manifestPath := "revue-42/manifest.json"
	body := string(files[manifestPath])
	for _, want := range []string{`"run_id": 42`, `"csv_sha256": "` + hash + `"`, `"subject_name": "Alpha"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("manifest missing %q in %s", want, body)
		}
	}
}
