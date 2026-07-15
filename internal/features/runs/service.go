package runs

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jeb-maker/revues/internal/features/subjects"
	"github.com/jeb-maker/revues/internal/store"
)

// --- Access (formerly internal/runs/access.go) ---

// CanView reports whether the user may view a run on a subject.
func CanView(user *store.User, orgMember bool) bool {
	return subjects.CanViewSubject(user, orgMember)
}

// CanLaunch reports whether the user may create or start a run.
func CanLaunch(user *store.User, orgMember bool) bool {
	return subjects.CanLaunchRun(user, orgMember)
}

// CanComplete reports whether the user may close a run (in_progress → done).
func CanComplete(user *store.User, orgRole string, orgMember bool) bool {
	return subjects.CanManageSubject(user, orgRole, orgMember)
}

// --- Due date (formerly internal/runs/due_date.go) ---

// ErrInvalidDueDate is returned when a due date value cannot be parsed.
var ErrInvalidDueDate = errors.New("invalid due date")

// ParseDueDate converts an optional HTML date input (YYYY-MM-DD) to ISO 8601 UTC.
// Empty input returns an empty string and no error.
func ParseDueDate(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return "", fmt.Errorf("%w", ErrInvalidDueDate)
	}
	return t.UTC().Format(time.RFC3339), nil
}

// --- CSV export (formerly internal/runs/export.go) ---

//nolint:misspell // French CSV column headers per issue #31
var runExportHeaders = []string{"subject", "revue", "date", "points", "statuts", "commentaires", "auteur"}

// BuildRunCSV encodes export rows as CSV with a header row.
func BuildRunCSV(rows []store.RunExportRow) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if err := w.Write(runExportHeaders); err != nil {
		return nil, fmt.Errorf("write csv header: %w", err)
	}

	for _, row := range rows {
		record := []string{
			row.SubjectName,
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

// --- Item access (formerly internal/items/access.go) ---

// CanUpdate reports whether the user may change run item statuses.
func CanUpdate(user *store.User, orgMember bool) bool {
	return subjects.CanLaunchRun(user, orgMember)
}

// CanLinkJira reports whether the user may link Jira issues to run items.
func CanLinkJira(user *store.User, orgMember bool) bool {
	return CanUpdate(user, orgMember)
}

// CanAssign reports whether the user may assign run items to members.
func CanAssign(user *store.User, orgRole string, orgMember bool) bool {
	return subjects.CanManageSubject(user, orgRole, orgMember)
}

// --- Item progress (formerly internal/items/progress.go) ---

// Progress counts completed run items (ok or na) and the total item count.
func Progress(runItems []store.RunItem) (done, total int) {
	total = len(runItems)
	for _, item := range runItems {
		if item.Status == StatusOK || item.Status == StatusNA {
			done++
		}
	}
	return done, total
}

// --- Item status (formerly internal/items/status.go) ---

// ErrCommentRequired is returned when status nok has no comment.
var ErrCommentRequired = errors.New("comment required for nok status")

// ErrInvalidStatus is returned for unknown item statuses.
var ErrInvalidStatus = errors.New("invalid item status")

var validStatuses = map[string]struct{}{
	StatusPending: {},
	StatusOK:      {},
	StatusNOK:     {},
	StatusNA:      {},
}

// ValidStatus reports whether status is an allowed run item status.
func ValidStatus(status string) bool {
	_, ok := validStatuses[status]
	return ok
}

// ValidateUpdate checks status and comment business rules.
func ValidateUpdate(status, comment string) error {
	if !ValidStatus(status) {
		return ErrInvalidStatus
	}
	if status == StatusNOK && strings.TrimSpace(comment) == "" {
		return ErrCommentRequired
	}
	return nil
}
