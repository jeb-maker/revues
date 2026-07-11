package store_test

import (
	"context"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

// Compare pool sizes:
//
//	go test ./internal/store/ -bench=BenchmarkConcurrentListProjects -benchmem -count=3
func BenchmarkConcurrentListProjects_Pool1(b *testing.B) {
	benchmarkConcurrentListProjects(b, 1)
}

func BenchmarkConcurrentListProjects_Pool10(b *testing.B) {
	benchmarkConcurrentListProjects(b, store.DefaultMaxOpenConns)
}

func benchmarkConcurrentListProjects(b *testing.B, maxOpen int) {
	ctx := context.Background()
	st, user := seedLoadFixturePool(b, maxOpen)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := st.ListProjects(ctx, user.ID, true); err != nil {
				b.Fatal(err)
			}
			if _, err := st.ListActiveRunSummaries(ctx, user.ID, true); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func seedLoadFixturePool(b *testing.B, maxOpen int) (*store.Store, *store.User) {
	b.Helper()
	ctx := context.Background()
	st := openWALFileStoreBench(b, maxOpen)
	user, err := st.UpsertGitHubUser(ctx, 1, "bench", "bench@example.com", "Bench", "", auth.RoleAdmin)
	if err != nil {
		b.Fatalf("UpsertGitHubUser(): %v", err)
	}
	if _, err = st.CreateProject(ctx, "Bench", "", user.ID); err != nil {
		b.Fatalf("CreateProject(): %v", err)
	}
	return st, user
}

func openWALFileStoreBench(b *testing.B, maxOpen int) *store.Store {
	b.Helper()
	ctx := context.Background()
	db, err := store.Open(ctx, b.TempDir()+"/bench.db", maxOpen)
	if err != nil {
		b.Fatalf("Open(): %v", err)
	}
	b.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			b.Errorf("Close(): %v", closeErr)
		}
	})
	if migrateErr := store.Migrate(ctx, db); migrateErr != nil {
		b.Fatalf("Migrate(): %v", migrateErr)
	}
	return store.New(db)
}
