package store

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestWithSQLiteBusyRetrySucceedsAfterBusy(t *testing.T) {
	ctx := context.Background()
	attempts := 0
	err := withSQLiteBusyRetry(ctx, func() error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("update run item: database is locked (5) (SQLITE_BUSY)")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("withSQLiteBusyRetry() = %v", err)
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
}

func TestWithSQLiteBusyRetryReturnsNonBusyImmediately(t *testing.T) {
	ctx := context.Background()
	want := errors.New("constraint failed")
	err := withSQLiteBusyRetry(ctx, func() error {
		return want
	})
	if !errors.Is(err, want) {
		t.Fatalf("err = %v, want %v", err, want)
	}
}

func TestIsSQLiteBusy(t *testing.T) {
	if !isSQLiteBusy(fmt.Errorf("database is locked (5) (SQLITE_BUSY)")) {
		t.Fatal("expected busy")
	}
	if isSQLiteBusy(errors.New("constraint failed")) {
		t.Fatal("expected not busy")
	}
}
