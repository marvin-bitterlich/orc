package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestPatrolRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPatrolRepository(db)
	ctx := context.Background()

	// Create test fixtures: factory -> workshop -> workbench -> kennel
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test Factory", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test Workshop", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test Workbench", "/test/path", "active")
	db.ExecContext(ctx, "INSERT INTO kennels (id, workbench_id, status) VALUES (?, ?, ?)", "KENNEL-001", "BENCH-001", "occupied")

	t.Run("creates patrol successfully", func(t *testing.T) {
		record := &secondary.PatrolRecord{
			ID:       "PATROL-001",
			KennelID: "KENNEL-001",
			Target:   "workshop:bench.2",
			Status:   "active",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "PATROL-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.Target != "workshop:bench.2" {
			t.Errorf("Target = %q, want %q", got.Target, "workshop:bench.2")
		}
		if got.Status != "active" {
			t.Errorf("Status = %q, want %q", got.Status, "active")
		}
	})
}

func TestPatrolRepository_GetActiveByKennel(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPatrolRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test", "/test", "active")
	db.ExecContext(ctx, "INSERT INTO kennels (id, workbench_id, status) VALUES (?, ?, ?)", "KENNEL-001", "BENCH-001", "occupied")

	// Create completed patrol
	repo.Create(ctx, &secondary.PatrolRecord{
		ID:       "PATROL-001",
		KennelID: "KENNEL-001",
		Target:   "workshop:bench.2",
		Status:   "completed",
	})

	// Create active patrol
	repo.Create(ctx, &secondary.PatrolRecord{
		ID:       "PATROL-002",
		KennelID: "KENNEL-001",
		Target:   "workshop:bench.2",
		Status:   "active",
	})

	t.Run("returns active patrol", func(t *testing.T) {
		active, err := repo.GetActiveByKennel(ctx, "KENNEL-001")
		if err != nil {
			t.Fatalf("GetActiveByKennel failed: %v", err)
		}
		if active == nil {
			t.Fatal("expected active patrol, got nil")
		}
		if active.ID != "PATROL-002" {
			t.Errorf("ID = %q, want PATROL-002", active.ID)
		}
	})
}

func TestPatrolRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPatrolRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test", "/test", "active")
	db.ExecContext(ctx, "INSERT INTO kennels (id, workbench_id, status) VALUES (?, ?, ?)", "KENNEL-001", "BENCH-001", "occupied")

	repo.Create(ctx, &secondary.PatrolRecord{
		ID:       "PATROL-001",
		KennelID: "KENNEL-001",
		Target:   "workshop:bench.2",
		Status:   "active",
	})

	t.Run("updates status", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "PATROL-001", "completed")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		patrol, _ := repo.GetByID(ctx, "PATROL-001")
		if patrol.Status != "completed" {
			t.Errorf("Status = %q, want completed", patrol.Status)
		}
	})
}

func TestPatrolRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPatrolRepository(db)
	ctx := context.Background()

	t.Run("returns first ID when empty", func(t *testing.T) {
		id, err := repo.GetNextID(ctx)
		if err != nil {
			t.Fatalf("GetNextID failed: %v", err)
		}
		if id != "PATROL-001" {
			t.Errorf("ID = %q, want PATROL-001", id)
		}
	})
}

func TestPatrolRepository_KennelExists(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPatrolRepository(db)
	ctx := context.Background()

	t.Run("returns false for non-existent kennel", func(t *testing.T) {
		exists, err := repo.KennelExists(ctx, "KENNEL-999")
		if err != nil {
			t.Fatalf("KennelExists failed: %v", err)
		}
		if exists {
			t.Error("expected false, got true")
		}
	})

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test", "/test", "active")
	db.ExecContext(ctx, "INSERT INTO kennels (id, workbench_id, status) VALUES (?, ?, ?)", "KENNEL-001", "BENCH-001", "vacant")

	t.Run("returns true for existing kennel", func(t *testing.T) {
		exists, err := repo.KennelExists(ctx, "KENNEL-001")
		if err != nil {
			t.Fatalf("KennelExists failed: %v", err)
		}
		if !exists {
			t.Error("expected true, got false")
		}
	})
}
