package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestStuckRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewStuckRepository(db)
	ctx := context.Background()

	// Create test fixtures: factory -> workshop -> workbench -> kennel -> patrol
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test Factory", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test Workshop", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test Workbench", "/test/path", "active")
	db.ExecContext(ctx, "INSERT INTO kennels (id, workbench_id, status) VALUES (?, ?, ?)", "KENNEL-001", "BENCH-001", "occupied")
	db.ExecContext(ctx, "INSERT INTO patrols (id, kennel_id, target, status) VALUES (?, ?, ?, ?)", "PATROL-001", "KENNEL-001", "workshop:bench.2", "active")

	t.Run("creates stuck successfully", func(t *testing.T) {
		record := &secondary.StuckRecord{
			ID:         "STUCK-001",
			PatrolID:   "PATROL-001",
			CheckCount: 1,
			Status:     "open",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "STUCK-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.Status != "open" {
			t.Errorf("Status = %q, want %q", got.Status, "open")
		}
		if got.CheckCount != 1 {
			t.Errorf("CheckCount = %d, want 1", got.CheckCount)
		}
	})
}

func TestStuckRepository_GetOpenByPatrol(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewStuckRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test", "/test", "active")
	db.ExecContext(ctx, "INSERT INTO kennels (id, workbench_id, status) VALUES (?, ?, ?)", "KENNEL-001", "BENCH-001", "occupied")
	db.ExecContext(ctx, "INSERT INTO patrols (id, kennel_id, target, status) VALUES (?, ?, ?, ?)", "PATROL-001", "KENNEL-001", "workshop:bench.2", "active")

	// Create resolved stuck
	repo.Create(ctx, &secondary.StuckRecord{
		ID:         "STUCK-001",
		PatrolID:   "PATROL-001",
		CheckCount: 3,
		Status:     "resolved",
	})

	// Create open stuck
	repo.Create(ctx, &secondary.StuckRecord{
		ID:         "STUCK-002",
		PatrolID:   "PATROL-001",
		CheckCount: 1,
		Status:     "open",
	})

	t.Run("returns open stuck", func(t *testing.T) {
		open, err := repo.GetOpenByPatrol(ctx, "PATROL-001")
		if err != nil {
			t.Fatalf("GetOpenByPatrol failed: %v", err)
		}
		if open == nil {
			t.Fatal("expected open stuck, got nil")
		}
		if open.ID != "STUCK-002" {
			t.Errorf("ID = %q, want STUCK-002", open.ID)
		}
	})
}

func TestStuckRepository_IncrementCount(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewStuckRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test", "/test", "active")
	db.ExecContext(ctx, "INSERT INTO kennels (id, workbench_id, status) VALUES (?, ?, ?)", "KENNEL-001", "BENCH-001", "occupied")
	db.ExecContext(ctx, "INSERT INTO patrols (id, kennel_id, target, status) VALUES (?, ?, ?, ?)", "PATROL-001", "KENNEL-001", "workshop:bench.2", "active")

	repo.Create(ctx, &secondary.StuckRecord{
		ID:         "STUCK-001",
		PatrolID:   "PATROL-001",
		CheckCount: 1,
		Status:     "open",
	})

	t.Run("increments count", func(t *testing.T) {
		err := repo.IncrementCount(ctx, "STUCK-001")
		if err != nil {
			t.Fatalf("IncrementCount failed: %v", err)
		}

		stuck, _ := repo.GetByID(ctx, "STUCK-001")
		if stuck.CheckCount != 2 {
			t.Errorf("CheckCount = %d, want 2", stuck.CheckCount)
		}
	})
}

func TestStuckRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewStuckRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test", "/test", "active")
	db.ExecContext(ctx, "INSERT INTO kennels (id, workbench_id, status) VALUES (?, ?, ?)", "KENNEL-001", "BENCH-001", "occupied")
	db.ExecContext(ctx, "INSERT INTO patrols (id, kennel_id, target, status) VALUES (?, ?, ?, ?)", "PATROL-001", "KENNEL-001", "workshop:bench.2", "active")

	repo.Create(ctx, &secondary.StuckRecord{
		ID:         "STUCK-001",
		PatrolID:   "PATROL-001",
		CheckCount: 5,
		Status:     "open",
	})

	t.Run("updates status", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "STUCK-001", "escalated")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		stuck, _ := repo.GetByID(ctx, "STUCK-001")
		if stuck.Status != "escalated" {
			t.Errorf("Status = %q, want escalated", stuck.Status)
		}
	})
}

func TestStuckRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewStuckRepository(db)
	ctx := context.Background()

	t.Run("returns first ID when empty", func(t *testing.T) {
		id, err := repo.GetNextID(ctx)
		if err != nil {
			t.Fatalf("GetNextID failed: %v", err)
		}
		if id != "STUCK-001" {
			t.Errorf("ID = %q, want STUCK-001", id)
		}
	})
}

func TestStuckRepository_PatrolExists(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewStuckRepository(db)
	ctx := context.Background()

	t.Run("returns false for non-existent patrol", func(t *testing.T) {
		exists, err := repo.PatrolExists(ctx, "PATROL-999")
		if err != nil {
			t.Fatalf("PatrolExists failed: %v", err)
		}
		if exists {
			t.Error("expected false, got true")
		}
	})

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test", "/test", "active")
	db.ExecContext(ctx, "INSERT INTO kennels (id, workbench_id, status) VALUES (?, ?, ?)", "KENNEL-001", "BENCH-001", "occupied")
	db.ExecContext(ctx, "INSERT INTO patrols (id, kennel_id, target, status) VALUES (?, ?, ?, ?)", "PATROL-001", "KENNEL-001", "workshop:bench.2", "active")

	t.Run("returns true for existing patrol", func(t *testing.T) {
		exists, err := repo.PatrolExists(ctx, "PATROL-001")
		if err != nil {
			t.Fatalf("PatrolExists failed: %v", err)
		}
		if !exists {
			t.Error("expected true, got false")
		}
	})
}
