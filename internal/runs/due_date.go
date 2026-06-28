package runs

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

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
