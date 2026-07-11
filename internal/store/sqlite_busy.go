package store

import (
	"context"
	"strings"
	"time"
)

const sqliteBusyRetryAttempts = 5

func isSQLiteBusy(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "database is locked") ||
		strings.Contains(msg, "sqlite_busy") ||
		strings.Contains(msg, "database table is locked")
}

func withSQLiteBusyRetry(ctx context.Context, fn func() error) error {
	var err error
	for attempt := range sqliteBusyRetryAttempts {
		err = fn()
		if err == nil {
			return nil
		}
		if !isSQLiteBusy(err) {
			return err
		}
		if attempt == sqliteBusyRetryAttempts-1 {
			break
		}
		wait := time.Duration(10*(attempt+1)) * time.Millisecond
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
	return err
}
