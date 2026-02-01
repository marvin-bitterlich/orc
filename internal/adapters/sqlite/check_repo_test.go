package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestCheckRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCheckRepository(db)
	ctx := context.Background()

	// Create test fixtures: factory -> workshop -> workbench -> kennel -> patrol
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test Factory", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test Workshop", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test Workbench", "/test/path", "active")
	db.ExecContext(ctx, "INSERT INTO kennels (id, workbench_id, status) VALUES (?, ?, ?)", "KENNEL-001", "BENCH-001", "occupied")
	db.ExecContext(ctx, "INSERT INTO patrols (id, kennel_id, target, status) VALUES (?, ?, ?, ?)", "PATROL-001", "KENNEL-001", "workshop:bench.2", "active")

	t.Run("creates check successfully", func(t *testing.T) {
		record := &secondary.CheckRecord{
			ID:          "CHECK-001",
			PatrolID:    "PATROL-001",
			PaneContent: "✶ Thundering… (esc to interrupt)",
			Outcome:     "working",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "CHECK-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.Outcome != "working" {
			t.Errorf("Outcome = %q, want %q", got.Outcome, "working")
		}
		if got.PaneContent != "✶ Thundering… (esc to interrupt)" {
			t.Errorf("PaneContent = %q, want %q", got.PaneContent, "✶ Thundering… (esc to interrupt)")
		}
	})
}

func TestCheckRepository_GetLatest(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCheckRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test", "/test", "active")
	db.ExecContext(ctx, "INSERT INTO kennels (id, workbench_id, status) VALUES (?, ?, ?)", "KENNEL-001", "BENCH-001", "occupied")
	db.ExecContext(ctx, "INSERT INTO patrols (id, kennel_id, target, status) VALUES (?, ?, ?, ?)", "PATROL-001", "KENNEL-001", "workshop:bench.2", "active")

	// Create multiple checks with explicit timestamps
	db.ExecContext(ctx, `INSERT INTO checks (id, patrol_id, pane_content, outcome, captured_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`, "CHECK-001", "PATROL-001", "content1", "working", "2024-01-01T10:00:00Z", "2024-01-01T10:00:00Z")
	db.ExecContext(ctx, `INSERT INTO checks (id, patrol_id, pane_content, outcome, captured_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`, "CHECK-002", "PATROL-001", "content2", "idle", "2024-01-01T10:01:00Z", "2024-01-01T10:01:00Z")

	t.Run("returns latest check", func(t *testing.T) {
		latest, err := repo.GetLatest(ctx, "PATROL-001")
		if err != nil {
			t.Fatalf("GetLatest failed: %v", err)
		}
		if latest == nil {
			t.Fatal("expected latest check, got nil")
		}
		if latest.ID != "CHECK-002" {
			t.Errorf("ID = %q, want CHECK-002", latest.ID)
		}
	})
}

func TestCheckRepository_GetByPatrol(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCheckRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test", "/test", "active")
	db.ExecContext(ctx, "INSERT INTO kennels (id, workbench_id, status) VALUES (?, ?, ?)", "KENNEL-001", "BENCH-001", "occupied")
	db.ExecContext(ctx, "INSERT INTO patrols (id, kennel_id, target, status) VALUES (?, ?, ?, ?)", "PATROL-001", "KENNEL-001", "workshop:bench.2", "active")

	// Create checks
	repo.Create(ctx, &secondary.CheckRecord{ID: "CHECK-001", PatrolID: "PATROL-001", Outcome: "working"})
	repo.Create(ctx, &secondary.CheckRecord{ID: "CHECK-002", PatrolID: "PATROL-001", Outcome: "idle"})

	t.Run("returns all checks for patrol", func(t *testing.T) {
		checks, err := repo.GetByPatrol(ctx, "PATROL-001")
		if err != nil {
			t.Fatalf("GetByPatrol failed: %v", err)
		}
		if len(checks) != 2 {
			t.Errorf("len(checks) = %d, want 2", len(checks))
		}
	})
}

func TestCheckRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCheckRepository(db)
	ctx := context.Background()

	t.Run("returns first ID when empty", func(t *testing.T) {
		id, err := repo.GetNextID(ctx)
		if err != nil {
			t.Fatalf("GetNextID failed: %v", err)
		}
		if id != "CHECK-001" {
			t.Errorf("ID = %q, want CHECK-001", id)
		}
	})
}

func TestCheckRepository_PatrolExists(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCheckRepository(db)
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
