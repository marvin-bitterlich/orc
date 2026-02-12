package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

// setupTomeTestDB creates the test database with required seed data.
func setupTomeTestDB(t *testing.T) *sql.DB {
	t.Helper()
	testDB := setupTestDB(t)
	seedCommission(t, testDB, "COMM-001", "Test Commission")
	return testDB
}

// createTestTome is a helper that creates a tome with a generated ID.
func createTestTome(t *testing.T, repo *sqlite.TomeRepository, ctx context.Context, commissionID, title, description string) *secondary.TomeRecord {
	t.Helper()

	nextID, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}

	tome := &secondary.TomeRecord{
		ID:           nextID,
		CommissionID: commissionID,
		Title:        title,
		Description:  description,
	}

	err = repo.Create(ctx, tome)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	return tome
}

func TestTomeRepository_Create(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db, nil)
	ctx := context.Background()

	tome := &secondary.TomeRecord{
		ID:           "TOME-001",
		CommissionID: "COMM-001",
		Title:        "Test Tome",
		Description:  "A test tome description",
	}

	err := repo.Create(ctx, tome)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify tome was created
	retrieved, err := repo.GetByID(ctx, "TOME-001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Title != "Test Tome" {
		t.Errorf("expected title 'Test Tome', got '%s'", retrieved.Title)
	}
	if retrieved.Status != "open" {
		t.Errorf("expected status 'active', got '%s'", retrieved.Status)
	}
}

func TestTomeRepository_GetByID(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db, nil)
	ctx := context.Background()

	tome := createTestTome(t, repo, ctx, "COMM-001", "Test Tome", "Description")

	retrieved, err := repo.GetByID(ctx, tome.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.Title != "Test Tome" {
		t.Errorf("expected title 'Test Tome', got '%s'", retrieved.Title)
	}
	if retrieved.Description != "Description" {
		t.Errorf("expected description 'Description', got '%s'", retrieved.Description)
	}
}

func TestTomeRepository_GetByID_NotFound(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db, nil)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "TOME-999")
	if err == nil {
		t.Error("expected error for non-existent tome")
	}
}

func TestTomeRepository_List(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db, nil)
	ctx := context.Background()

	createTestTome(t, repo, ctx, "COMM-001", "Tome 1", "")
	createTestTome(t, repo, ctx, "COMM-001", "Tome 2", "")
	createTestTome(t, repo, ctx, "COMM-001", "Tome 3", "")

	tomes, err := repo.List(ctx, secondary.TomeFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(tomes) != 3 {
		t.Errorf("expected 3 tomes, got %d", len(tomes))
	}
}

func TestTomeRepository_List_FilterByCommission(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db, nil)
	ctx := context.Background()

	// Add another commission
	_, _ = db.Exec("INSERT INTO commissions (id, title, status) VALUES ('COMM-002', 'Commission 2', 'active')")

	createTestTome(t, repo, ctx, "COMM-001", "Tome 1", "")
	createTestTome(t, repo, ctx, "COMM-002", "Tome 2", "")

	tomes, err := repo.List(ctx, secondary.TomeFilters{CommissionID: "COMM-001"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(tomes) != 1 {
		t.Errorf("expected 1 tome for COMM-001, got %d", len(tomes))
	}
}

func TestTomeRepository_List_FilterByStatus(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db, nil)
	ctx := context.Background()

	t1 := createTestTome(t, repo, ctx, "COMM-001", "Active Tome", "")
	createTestTome(t, repo, ctx, "COMM-001", "Another Active", "")

	// Complete t1
	_ = repo.UpdateStatus(ctx, t1.ID, "closed", true)

	tomes, err := repo.List(ctx, secondary.TomeFilters{Status: "open"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(tomes) != 1 {
		t.Errorf("expected 1 active tome, got %d", len(tomes))
	}
}

func TestTomeRepository_Update(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db, nil)
	ctx := context.Background()

	tome := createTestTome(t, repo, ctx, "COMM-001", "Original Title", "")

	err := repo.Update(ctx, &secondary.TomeRecord{
		ID:    tome.ID,
		Title: "Updated Title",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, tome.ID)
	if retrieved.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", retrieved.Title)
	}
}

func TestTomeRepository_Update_NotFound(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db, nil)
	ctx := context.Background()

	err := repo.Update(ctx, &secondary.TomeRecord{
		ID:    "TOME-999",
		Title: "Updated Title",
	})
	if err == nil {
		t.Error("expected error for non-existent tome")
	}
}

func TestTomeRepository_Delete(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db, nil)
	ctx := context.Background()

	tome := createTestTome(t, repo, ctx, "COMM-001", "To Delete", "")

	err := repo.Delete(ctx, tome.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.GetByID(ctx, tome.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestTomeRepository_Delete_NotFound(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db, nil)
	ctx := context.Background()

	err := repo.Delete(ctx, "TOME-999")
	if err == nil {
		t.Error("expected error for non-existent tome")
	}
}

func TestTomeRepository_Pin_Unpin(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db, nil)
	ctx := context.Background()

	tome := createTestTome(t, repo, ctx, "COMM-001", "Pin Test", "")

	// Pin
	err := repo.Pin(ctx, tome.ID)
	if err != nil {
		t.Fatalf("Pin failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, tome.ID)
	if !retrieved.Pinned {
		t.Error("expected tome to be pinned")
	}

	// Unpin
	err = repo.Unpin(ctx, tome.ID)
	if err != nil {
		t.Fatalf("Unpin failed: %v", err)
	}

	retrieved, _ = repo.GetByID(ctx, tome.ID)
	if retrieved.Pinned {
		t.Error("expected tome to be unpinned")
	}
}

func TestTomeRepository_Pin_NotFound(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db, nil)
	ctx := context.Background()

	err := repo.Pin(ctx, "TOME-999")
	if err == nil {
		t.Error("expected error for non-existent tome")
	}
}

func TestTomeRepository_GetNextID(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db, nil)
	ctx := context.Background()

	id, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "TOME-001" {
		t.Errorf("expected TOME-001, got %s", id)
	}

	createTestTome(t, repo, ctx, "COMM-001", "Test", "")

	id, err = repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "TOME-002" {
		t.Errorf("expected TOME-002, got %s", id)
	}
}

func TestTomeRepository_UpdateStatus(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db, nil)
	ctx := context.Background()

	tome := createTestTome(t, repo, ctx, "COMM-001", "Status Test", "")

	// Verify initial status is open
	retrieved, _ := repo.GetByID(ctx, tome.ID)
	if retrieved.Status != "open" {
		t.Errorf("expected initial status 'open', got '%s'", retrieved.Status)
	}
	if retrieved.ClosedAt != "" {
		t.Error("expected ClosedAt to be empty")
	}

	// Update to closed
	err := repo.UpdateStatus(ctx, tome.ID, "closed", true)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	retrieved, _ = repo.GetByID(ctx, tome.ID)
	if retrieved.Status != "closed" {
		t.Errorf("expected status 'complete', got '%s'", retrieved.Status)
	}
	if retrieved.ClosedAt == "" {
		t.Error("expected ClosedAt to be set")
	}
}

func TestTomeRepository_UpdateStatus_NotFound(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db, nil)
	ctx := context.Background()

	err := repo.UpdateStatus(ctx, "TOME-999", "closed", true)
	if err == nil {
		t.Error("expected error for non-existent tome")
	}
}

func TestTomeRepository_AssignWorkbench(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db, nil)
	ctx := context.Background()

	tome := createTestTome(t, repo, ctx, "COMM-001", "Workbench Test", "")

	err := repo.AssignWorkbench(ctx, tome.ID, "BENCH-001")
	if err != nil {
		t.Fatalf("AssignWorkbench failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, tome.ID)
	if retrieved.AssignedWorkbenchID != "BENCH-001" {
		t.Errorf("expected assigned workbench 'BENCH-001', got '%s'", retrieved.AssignedWorkbenchID)
	}
}

func TestTomeRepository_AssignWorkbench_NotFound(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db, nil)
	ctx := context.Background()

	err := repo.AssignWorkbench(ctx, "TOME-999", "BENCH-001")
	if err == nil {
		t.Error("expected error for non-existent tome")
	}
}

func TestTomeRepository_GetByWorkbench(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db, nil)
	ctx := context.Background()

	t1 := createTestTome(t, repo, ctx, "COMM-001", "Tome 1", "")
	t2 := createTestTome(t, repo, ctx, "COMM-001", "Tome 2", "")
	createTestTome(t, repo, ctx, "COMM-001", "Tome 3 (unassigned)", "")

	_ = repo.AssignWorkbench(ctx, t1.ID, "BENCH-001")
	_ = repo.AssignWorkbench(ctx, t2.ID, "BENCH-001")

	tomes, err := repo.GetByWorkbench(ctx, "BENCH-001")
	if err != nil {
		t.Fatalf("GetByWorkbench failed: %v", err)
	}

	if len(tomes) != 2 {
		t.Errorf("expected 2 tomes for workbench, got %d", len(tomes))
	}
}

func TestTomeRepository_CommissionExists(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db, nil)
	ctx := context.Background()

	exists, err := repo.CommissionExists(ctx, "COMM-001")
	if err != nil {
		t.Fatalf("CommissionExists failed: %v", err)
	}
	if !exists {
		t.Error("expected commission to exist")
	}

	exists, err = repo.CommissionExists(ctx, "COMM-999")
	if err != nil {
		t.Fatalf("CommissionExists failed: %v", err)
	}
	if exists {
		t.Error("expected commission to not exist")
	}
}
