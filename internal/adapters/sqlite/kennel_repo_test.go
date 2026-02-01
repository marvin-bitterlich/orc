package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestKennelRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewKennelRepository(db)
	ctx := context.Background()

	// Create test fixtures: factory -> workshop -> workbench
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test Factory", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test Workshop", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test Workbench", "/test/path", "active")

	t.Run("creates kennel successfully", func(t *testing.T) {
		record := &secondary.KennelRecord{
			ID:          "KENNEL-001",
			WorkbenchID: "BENCH-001",
			Status:      "vacant",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "KENNEL-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.WorkbenchID != "BENCH-001" {
			t.Errorf("WorkbenchID = %q, want %q", got.WorkbenchID, "BENCH-001")
		}
		if got.Status != "vacant" {
			t.Errorf("Status = %q, want %q", got.Status, "vacant")
		}
	})
}

func TestKennelRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewKennelRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test", "/test", "active")

	repo.Create(ctx, &secondary.KennelRecord{
		ID:          "KENNEL-001",
		WorkbenchID: "BENCH-001",
		Status:      "vacant",
	})

	t.Run("finds kennel by ID", func(t *testing.T) {
		got, err := repo.GetByID(ctx, "KENNEL-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if got.ID != "KENNEL-001" {
			t.Errorf("ID = %q, want %q", got.ID, "KENNEL-001")
		}
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "KENNEL-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestKennelRepository_GetByWorkbench(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewKennelRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test", "/test", "active")

	repo.Create(ctx, &secondary.KennelRecord{
		ID:          "KENNEL-001",
		WorkbenchID: "BENCH-001",
		Status:      "vacant",
	})

	t.Run("finds kennel by workbench", func(t *testing.T) {
		got, err := repo.GetByWorkbench(ctx, "BENCH-001")
		if err != nil {
			t.Fatalf("GetByWorkbench failed: %v", err)
		}
		if got == nil {
			t.Fatal("expected kennel, got nil")
		}
		if got.ID != "KENNEL-001" {
			t.Errorf("ID = %q, want %q", got.ID, "KENNEL-001")
		}
	})

	t.Run("returns error for workbench without kennel", func(t *testing.T) {
		_, err := repo.GetByWorkbench(ctx, "BENCH-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestKennelRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewKennelRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test 1", "/test1", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-002", "WORK-001", "Test 2", "/test2", "active")

	repo.Create(ctx, &secondary.KennelRecord{ID: "KENNEL-001", WorkbenchID: "BENCH-001", Status: "vacant"})
	repo.Create(ctx, &secondary.KennelRecord{ID: "KENNEL-002", WorkbenchID: "BENCH-002", Status: "occupied"})

	t.Run("lists all kennels", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.KennelFilters{})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("len = %d, want 2", len(list))
		}
	})

	t.Run("filters by workbench_id", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.KennelFilters{WorkbenchID: "BENCH-002"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "KENNEL-002" {
			t.Errorf("ID = %q, want %q", list[0].ID, "KENNEL-002")
		}
	})

	t.Run("filters by status", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.KennelFilters{Status: "occupied"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "KENNEL-002" {
			t.Errorf("ID = %q, want %q", list[0].ID, "KENNEL-002")
		}
	})
}

func TestKennelRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewKennelRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test", "/test", "active")

	repo.Create(ctx, &secondary.KennelRecord{ID: "KENNEL-001", WorkbenchID: "BENCH-001", Status: "vacant"})

	t.Run("deletes kennel", func(t *testing.T) {
		err := repo.Delete(ctx, "KENNEL-001")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, err = repo.GetByID(ctx, "KENNEL-001")
		if err == nil {
			t.Error("expected error after delete, got nil")
		}
	})

	t.Run("returns error for non-existent kennel", func(t *testing.T) {
		err := repo.Delete(ctx, "KENNEL-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestKennelRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewKennelRepository(db)
	ctx := context.Background()

	t.Run("returns KENNEL-001 for empty table", func(t *testing.T) {
		id, err := repo.GetNextID(ctx)
		if err != nil {
			t.Fatalf("GetNextID failed: %v", err)
		}
		if id != "KENNEL-001" {
			t.Errorf("ID = %q, want %q", id, "KENNEL-001")
		}
	})
}

func TestKennelRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewKennelRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test", "/test", "active")

	repo.Create(ctx, &secondary.KennelRecord{
		ID:          "KENNEL-001",
		WorkbenchID: "BENCH-001",
		Status:      "vacant",
	})

	t.Run("updates status to occupied", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "KENNEL-001", "occupied")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "KENNEL-001")
		if got.Status != "occupied" {
			t.Errorf("Status = %q, want %q", got.Status, "occupied")
		}
	})

	t.Run("updates status to away", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "KENNEL-001", "away")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "KENNEL-001")
		if got.Status != "away" {
			t.Errorf("Status = %q, want %q", got.Status, "away")
		}
	})

	t.Run("returns error for non-existent kennel", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "KENNEL-999", "occupied")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestKennelRepository_WorkbenchExists(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewKennelRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test", "/test", "active")

	t.Run("returns true for existing workbench", func(t *testing.T) {
		exists, err := repo.WorkbenchExists(ctx, "BENCH-001")
		if err != nil {
			t.Fatalf("WorkbenchExists failed: %v", err)
		}
		if !exists {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns false for non-existent workbench", func(t *testing.T) {
		exists, err := repo.WorkbenchExists(ctx, "BENCH-999")
		if err != nil {
			t.Fatalf("WorkbenchExists failed: %v", err)
		}
		if exists {
			t.Error("expected false, got true")
		}
	})
}

func TestKennelRepository_WorkbenchHasKennel(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewKennelRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO factories (id, name, status) VALUES (?, ?, ?)", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-001", "WORK-001", "Test 1", "/test1", "active")
	db.ExecContext(ctx, "INSERT INTO workbenches (id, workshop_id, name, path, status) VALUES (?, ?, ?, ?, ?)", "BENCH-002", "WORK-001", "Test 2", "/test2", "active")

	repo.Create(ctx, &secondary.KennelRecord{
		ID:          "KENNEL-001",
		WorkbenchID: "BENCH-001",
		Status:      "vacant",
	})

	t.Run("returns true when workbench has kennel", func(t *testing.T) {
		has, err := repo.WorkbenchHasKennel(ctx, "BENCH-001")
		if err != nil {
			t.Fatalf("WorkbenchHasKennel failed: %v", err)
		}
		if !has {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns false when workbench has no kennel", func(t *testing.T) {
		has, err := repo.WorkbenchHasKennel(ctx, "BENCH-002")
		if err != nil {
			t.Fatalf("WorkbenchHasKennel failed: %v", err)
		}
		if has {
			t.Error("expected false, got true")
		}
	})
}
