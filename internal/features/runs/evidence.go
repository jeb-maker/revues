package runs

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path"
	"strconv"
)

// EvidenceAttachmentRef describes a run-item attachment without embedding binaries.
type EvidenceAttachmentRef struct {
	RunItemID   int64  `json:"run_item_id"`
	Filename    string `json:"filename"`
	StoragePath string `json:"storage_path"`
	SizeBytes   int64  `json:"size_bytes"`
	SHA256      string `json:"sha256,omitempty"`
}

// EvidenceManifest is the JSON metadata sealed with a completed run export.
type EvidenceManifest struct {
	RunID        int64                   `json:"run_id"`
	SubjectName  string                  `json:"subject_name"`
	TemplateName string                  `json:"template_name"`
	Version      int                     `json:"version"`
	Status       string                  `json:"status"`
	CompletedAt  string                  `json:"completed_at"`
	ClosedBy     string                  `json:"closed_by,omitempty"`
	CSVSHA256    string                  `json:"csv_sha256"`
	GeneratedAt  string                  `json:"generated_at"`
	Attachments  []EvidenceAttachmentRef `json:"attachments,omitempty"`
}

// SHA256Hex returns the lowercase hex SHA-256 digest of data.
func SHA256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// BuildEvidenceZIP packs CSV + manifest + sha256sum.txt under revue-{id}/.
func BuildEvidenceZIP(runID int64, csvData []byte, manifest EvidenceManifest) ([]byte, error) {
	dir := "revue-" + strconv.FormatInt(runID, 10)
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	if err := writeZipFile(zw, path.Join(dir, "revue.csv"), csvData); err != nil {
		_ = zw.Close()
		return nil, err
	}

	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		_ = zw.Close()
		return nil, fmt.Errorf("marshal evidence manifest: %w", err)
	}
	manifestJSON = append(manifestJSON, '\n')
	if err := writeZipFile(zw, path.Join(dir, "manifest.json"), manifestJSON); err != nil {
		_ = zw.Close()
		return nil, err
	}

	sumLine := []byte(manifest.CSVSHA256 + "  revue.csv\n")
	if err := writeZipFile(zw, path.Join(dir, "sha256sum.txt"), sumLine); err != nil {
		_ = zw.Close()
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("close evidence zip: %w", err)
	}
	return buf.Bytes(), nil
}

func writeZipFile(zw *zip.Writer, name string, data []byte) error {
	w, err := zw.Create(name)
	if err != nil {
		return fmt.Errorf("create zip entry %s: %w", name, err)
	}
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("write zip entry %s: %w", name, err)
	}
	return nil
}
