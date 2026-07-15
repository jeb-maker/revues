package store_test

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jeb-maker/revues/internal/auth"
	runs "github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
)

const (
	loadWorkers       = 32
	loadIterations    = 50
	maxDualWriteLocks = 0.05 // 5% — two writers on distinct items
	maxWriteLatency   = 500 * time.Millisecond
)

func TestConcurrentReadsNoLock(t *testing.T) {
	ctx, st, user := seedLoadFixture(t)

	var lockErrors atomic.Int64
	var wg sync.WaitGroup
	wg.Add(loadWorkers)

	for w := 0; w < loadWorkers; w++ {
		go func() {
			defer wg.Done()
			for range loadIterations {
				if _, err := st.ListProjects(ctx, user.ID, true, ""); err != nil {
					if isSQLiteLockErr(err) {
						lockErrors.Add(1)
					}
					t.Errorf("ListProjects(): %v", err)
					return
				}
				if _, err := st.ListActiveRunSummaries(ctx, user.ID, true); err != nil {
					if isSQLiteLockErr(err) {
						lockErrors.Add(1)
					}
					t.Errorf("ListActiveRunSummaries(): %v", err)
					return
				}
			}
		}()
	}

	wg.Wait()
	if n := lockErrors.Load(); n > 0 {
		t.Fatalf("database lock errors under read load: %d", n)
	}
}

func TestConcurrentReadsWithWriterNoLock(t *testing.T) {
	ctx, st, run, itemIDs := seedInProgressRunFileDB(t)

	userID := int64(1)
	statuses := []string{runs.StatusOK, runs.StatusPending, runs.StatusNA, runs.StatusNOK}

	var lockErrors atomic.Int64
	var wg sync.WaitGroup

	readWorkers := 32
	wg.Add(readWorkers + 1)

	go func() {
		defer wg.Done()
		for i := range loadIterations {
			itemID := itemIDs[i%len(itemIDs)]
			status := statuses[i%len(statuses)]
			comment := ""
			if status == runs.StatusNOK {
				comment = fmt.Sprintf("nok-%d", i)
			}
			if err := st.UpdateRunItemStatus(ctx, run.ID, itemID, userID, status, comment); err != nil {
				if isSQLiteLockErr(err) {
					lockErrors.Add(1)
				}
				t.Errorf("UpdateRunItemStatus(): %v", err)
				return
			}
		}
	}()

	for range readWorkers {
		go func() {
			defer wg.Done()
			for range loadIterations {
				if _, err := st.ListRunItems(ctx, run.ID); err != nil {
					if isSQLiteLockErr(err) {
						lockErrors.Add(1)
					}
					t.Errorf("ListRunItems(): %v", err)
					return
				}
				if _, err := st.ListActiveRunSummaries(ctx, userID, true); err != nil {
					if isSQLiteLockErr(err) {
						lockErrors.Add(1)
					}
					t.Errorf("ListActiveRunSummaries(): %v", err)
					return
				}
			}
		}()
	}

	wg.Wait()
	if n := lockErrors.Load(); n > 0 {
		t.Fatalf("database lock errors under read+writer load: %d", n)
	}
}

func TestWriterLatencyUnderReadLoad(t *testing.T) {
	ctx, st, run, itemIDs := seedInProgressRunFileDB(t)

	const readWorkers = 16
	var slowWrites atomic.Int64
	var wg sync.WaitGroup
	wg.Add(readWorkers + 1)

	go func() {
		defer wg.Done()
		for i := range loadIterations {
			itemID := itemIDs[i%len(itemIDs)]
			start := time.Now()
			if err := st.UpdateRunItemStatus(ctx, run.ID, itemID, 1, runs.StatusOK, ""); err != nil {
				t.Errorf("UpdateRunItemStatus(): %v", err)
				return
			}
			if time.Since(start) > maxWriteLatency {
				slowWrites.Add(1)
			}
		}
	}()

	for range readWorkers {
		go func() {
			defer wg.Done()
			for range loadIterations {
				if _, err := st.ListRunItems(ctx, run.ID); err != nil {
					t.Errorf("ListRunItems(): %v", err)
					return
				}
			}
		}()
	}

	wg.Wait()
	if n := slowWrites.Load(); n > 0 {
		t.Fatalf("writes slower than %s under read load: %d", maxWriteLatency, n)
	}
}

// TestConcurrentDualWriterLockRate models two contributors updating different items.
func TestConcurrentDualWriterLockRate(t *testing.T) {
	ctx, st, run, itemIDs := seedInProgressRunFileDB(t)

	const writeWorkers = 2
	const iterations = 50

	var lockErrors atomic.Int64
	var ops atomic.Int64
	var wg sync.WaitGroup
	wg.Add(writeWorkers)

	for w := range writeWorkers {
		go func(worker int) {
			defer wg.Done()
			itemID := itemIDs[worker]
			for range iterations {
				ops.Add(1)
				if err := st.UpdateRunItemStatus(ctx, run.ID, itemID, 1, runs.StatusOK, ""); err != nil {
					if isSQLiteLockErr(err) {
						lockErrors.Add(1)
						continue
					}
					t.Errorf("UpdateRunItemStatus(): %v", err)
					return
				}
			}
		}(w)
	}

	wg.Wait()

	locks := lockErrors.Load()
	total := ops.Load()
	rate := lockRate(locks, total)
	t.Logf("dual writer: %d lock errors on %d ops (%.1f%%)", locks, total, rate*100)
	if rate > maxDualWriteLocks {
		t.Fatalf("dual writer lock rate %.1f%% exceeds %.0f%% threshold", rate*100, maxDualWriteLocks*100)
	}
}

// TestConcurrentWriteStressCanary logs lock rate under unrealistic parallel write storms.
func TestConcurrentWriteStressCanary(t *testing.T) {
	ctx, st, run, itemIDs := seedInProgressRunFileDB(t)

	const writeWorkers = 4
	const iterations = 30

	var lockErrors atomic.Int64
	var ops atomic.Int64
	var wg sync.WaitGroup
	wg.Add(writeWorkers)

	for w := range writeWorkers {
		go func(worker int) {
			defer wg.Done()
			itemID := itemIDs[worker%len(itemIDs)]
			for range iterations {
				ops.Add(1)
				if err := st.UpdateRunItemStatus(ctx, run.ID, itemID, 1, runs.StatusOK, ""); err != nil {
					if isSQLiteLockErr(err) {
						lockErrors.Add(1)
						continue
					}
					t.Errorf("UpdateRunItemStatus(): %v", err)
					return
				}
			}
		}(w)
	}

	wg.Wait()

	locks := lockErrors.Load()
	total := ops.Load()
	t.Logf("write stress canary: %d lock errors on %d ops (%.1f%%)", locks, total, lockRate(locks, total)*100)
}

func lockRate(locks, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(locks) / float64(total)
}

func seedLoadFixture(t *testing.T) (context.Context, *store.Store, *store.User) {
	t.Helper()

	st := openWALFileStore(t, store.DefaultMaxOpenConns)
	ctx := orgCtxForStore(t, st)

	user, err := st.UpsertGitHubUser(ctx, 1, "load", "load@example.com", "Load", "", auth.RoleAdmin)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(default): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, user.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}
	if _, err = st.CreateProject(ctx, "Load", "desc", user.ID, nil); err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	return ctx, st, user
}

func seedInProgressRunFileDB(t *testing.T) (context.Context, *store.Store, *store.ChecklistRun, []int64) {
	t.Helper()

	st := openWALFileStore(t, store.DefaultMaxOpenConns)
	ctx := orgCtxForStore(t, st)

	lead, err := st.UpsertGitHubUser(ctx, 1, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(default): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, lead.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}
	project, err := st.CreateProject(ctx, "P", "", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, []store.TemplateItemInput{
		{Section: "S", Label: "Point 1", Required: true},
		{Section: "S", Label: "Point 2", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, lead.ID)
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	if err = st.StartRun(ctx, run.ID); err != nil {
		t.Fatalf("StartRun(): %v", err)
	}
	runItems, err := st.ListRunItems(ctx, run.ID)
	if err != nil || len(runItems) < 2 {
		t.Fatalf("ListRunItems() = %v, %v", runItems, err)
	}
	itemIDs := make([]int64, len(runItems))
	for i, item := range runItems {
		itemIDs[i] = item.ID
	}
	return ctx, st, run, itemIDs
}

func orgCtxForStore(t *testing.T, st *store.Store) context.Context {
	t.Helper()

	ctx := context.Background()
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(default): %v", err)
	}
	return orgctx.WithOrganizationID(ctx, defaultOrg.ID)
}

func openWALFileStore(t *testing.T, maxOpen int) *store.Store {
	t.Helper()

	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/concurrent.db", maxOpen)
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Errorf("Close(): %v", closeErr)
		}
	})
	if migrateErr := store.Migrate(ctx, db); migrateErr != nil {
		t.Fatalf("Migrate(): %v", migrateErr)
	}
	return store.New(db)
}

func isSQLiteLockErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "database is locked") ||
		strings.Contains(msg, "sqlite_busy") ||
		strings.Contains(msg, "database table is locked")
}
