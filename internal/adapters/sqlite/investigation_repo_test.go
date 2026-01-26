package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

// setupInvestigationTestDB creates the test database with required seed data.
func setupInvestigationTestDB(t *testing.T) *sql.DB {
	t.Helper()
	testDB := setupTestDB(t)
	seedCommission(t, testDB, "COMM-001", "Test Commission")
	return testDB
}

// createTestInvestigation is a helper that creates an investigation with a generated ID.
func createTestInvestigation(t *testing.T, repo *sqlite.InvestigationRepository, ctx context.Context, commissionID, title, description string) *secondary.InvestigationRecord {
	t.Helper()

	nextID, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}

	inv := &secondary.InvestigationRecord{
		ID:           nextID,
		CommissionID: commissionID,
		Title:        title,
		Description:  description,
	}

	err = repo.Create(ctx, inv)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	return inv
}

func TestInvestigationRepository_Create(t *testing.T) {
	db := setupInvestigationTestDB(t)
	repo := sqlite.NewInvestigationRepository(db)
	ctx := context.Background()

	inv := &secondary.InvestigationRecord{
		ID:           "INV-001",
		CommissionID: "COMM-001",
		Title:        "Test Investigation",
		Description:  "A test investigation description",
	}

	err := repo.Create(ctx, inv)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify investigation was created
	retrieved, err := repo.GetByID(ctx, "INV-001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Title != "Test Investigation" {
		t.Errorf("expected title 'Test Investigation', got '%s'", retrieved.Title)
	}
	if retrieved.Status != "active" {
		t.Errorf("expected status 'active', got '%s'", retrieved.Status)
	}
}

func TestInvestigationRepository_GetByID(t *testing.T) {
	db := setupInvestigationTestDB(t)
	repo := sqlite.NewInvestigationRepository(db)
	ctx := context.Background()

	inv := createTestInvestigation(t, repo, ctx, "COMM-001", "Test Investigation", "Description")

	retrieved, err := repo.GetByID(ctx, inv.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.Title != "Test Investigation" {
		t.Errorf("expected title 'Test Investigation', got '%s'", retrieved.Title)
	}
	if retrieved.Description != "Description" {
		t.Errorf("expected description 'Description', got '%s'", retrieved.Description)
	}
}

func TestInvestigationRepository_GetByID_NotFound(t *testing.T) {
	db := setupInvestigationTestDB(t)
	repo := sqlite.NewInvestigationRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "INV-999")
	if err == nil {
		t.Error("expected error for non-existent investigation")
	}
}

func TestInvestigationRepository_List(t *testing.T) {
	db := setupInvestigationTestDB(t)
	repo := sqlite.NewInvestigationRepository(db)
	ctx := context.Background()

	createTestInvestigation(t, repo, ctx, "COMM-001", "Investigation 1", "")
	createTestInvestigation(t, repo, ctx, "COMM-001", "Investigation 2", "")
	createTestInvestigation(t, repo, ctx, "COMM-001", "Investigation 3", "")

	investigations, err := repo.List(ctx, secondary.InvestigationFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(investigations) != 3 {
		t.Errorf("expected 3 investigations, got %d", len(investigations))
	}
}

func TestInvestigationRepository_List_FilterByCommission(t *testing.T) {
	db := setupInvestigationTestDB(t)
	repo := sqlite.NewInvestigationRepository(db)
	ctx := context.Background()

	// Add another commission
	_, _ = db.Exec("INSERT INTO commissions (id, title, status) VALUES ('COMM-002', 'Commission 2', 'active')")

	createTestInvestigation(t, repo, ctx, "COMM-001", "Investigation 1", "")
	createTestInvestigation(t, repo, ctx, "COMM-002", "Investigation 2", "")

	investigations, err := repo.List(ctx, secondary.InvestigationFilters{CommissionID: "COMM-001"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(investigations) != 1 {
		t.Errorf("expected 1 investigation for COMM-001, got %d", len(investigations))
	}
}

func TestInvestigationRepository_List_FilterByStatus(t *testing.T) {
	db := setupInvestigationTestDB(t)
	repo := sqlite.NewInvestigationRepository(db)
	ctx := context.Background()

	inv1 := createTestInvestigation(t, repo, ctx, "COMM-001", "Active Investigation", "")
	createTestInvestigation(t, repo, ctx, "COMM-001", "Another Active", "")

	// Complete inv1
	_ = repo.UpdateStatus(ctx, inv1.ID, "complete", true)

	investigations, err := repo.List(ctx, secondary.InvestigationFilters{Status: "active"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(investigations) != 1 {
		t.Errorf("expected 1 active investigation, got %d", len(investigations))
	}
}

func TestInvestigationRepository_Update(t *testing.T) {
	db := setupInvestigationTestDB(t)
	repo := sqlite.NewInvestigationRepository(db)
	ctx := context.Background()

	inv := createTestInvestigation(t, repo, ctx, "COMM-001", "Original Title", "")

	err := repo.Update(ctx, &secondary.InvestigationRecord{
		ID:    inv.ID,
		Title: "Updated Title",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, inv.ID)
	if retrieved.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", retrieved.Title)
	}
}

func TestInvestigationRepository_Update_NotFound(t *testing.T) {
	db := setupInvestigationTestDB(t)
	repo := sqlite.NewInvestigationRepository(db)
	ctx := context.Background()

	err := repo.Update(ctx, &secondary.InvestigationRecord{
		ID:    "INV-999",
		Title: "Updated Title",
	})
	if err == nil {
		t.Error("expected error for non-existent investigation")
	}
}

func TestInvestigationRepository_Delete(t *testing.T) {
	db := setupInvestigationTestDB(t)
	repo := sqlite.NewInvestigationRepository(db)
	ctx := context.Background()

	inv := createTestInvestigation(t, repo, ctx, "COMM-001", "To Delete", "")

	err := repo.Delete(ctx, inv.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.GetByID(ctx, inv.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestInvestigationRepository_Delete_NotFound(t *testing.T) {
	db := setupInvestigationTestDB(t)
	repo := sqlite.NewInvestigationRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, "INV-999")
	if err == nil {
		t.Error("expected error for non-existent investigation")
	}
}

func TestInvestigationRepository_Pin_Unpin(t *testing.T) {
	db := setupInvestigationTestDB(t)
	repo := sqlite.NewInvestigationRepository(db)
	ctx := context.Background()

	inv := createTestInvestigation(t, repo, ctx, "COMM-001", "Pin Test", "")

	// Pin
	err := repo.Pin(ctx, inv.ID)
	if err != nil {
		t.Fatalf("Pin failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, inv.ID)
	if !retrieved.Pinned {
		t.Error("expected investigation to be pinned")
	}

	// Unpin
	err = repo.Unpin(ctx, inv.ID)
	if err != nil {
		t.Fatalf("Unpin failed: %v", err)
	}

	retrieved, _ = repo.GetByID(ctx, inv.ID)
	if retrieved.Pinned {
		t.Error("expected investigation to be unpinned")
	}
}

func TestInvestigationRepository_Pin_NotFound(t *testing.T) {
	db := setupInvestigationTestDB(t)
	repo := sqlite.NewInvestigationRepository(db)
	ctx := context.Background()

	err := repo.Pin(ctx, "INV-999")
	if err == nil {
		t.Error("expected error for non-existent investigation")
	}
}

func TestInvestigationRepository_GetNextID(t *testing.T) {
	db := setupInvestigationTestDB(t)
	repo := sqlite.NewInvestigationRepository(db)
	ctx := context.Background()

	id, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "INV-001" {
		t.Errorf("expected INV-001, got %s", id)
	}

	createTestInvestigation(t, repo, ctx, "COMM-001", "Test", "")

	id, err = repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "INV-002" {
		t.Errorf("expected INV-002, got %s", id)
	}
}

func TestInvestigationRepository_UpdateStatus(t *testing.T) {
	db := setupInvestigationTestDB(t)
	repo := sqlite.NewInvestigationRepository(db)
	ctx := context.Background()

	inv := createTestInvestigation(t, repo, ctx, "COMM-001", "Status Test", "")

	// Update status without completed timestamp
	err := repo.UpdateStatus(ctx, inv.ID, "in_progress", false)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, inv.ID)
	if retrieved.Status != "in_progress" {
		t.Errorf("expected status 'in_progress', got '%s'", retrieved.Status)
	}
	if retrieved.CompletedAt != "" {
		t.Error("expected CompletedAt to be empty")
	}

	// Update to complete
	err = repo.UpdateStatus(ctx, inv.ID, "complete", true)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	retrieved, _ = repo.GetByID(ctx, inv.ID)
	if retrieved.Status != "complete" {
		t.Errorf("expected status 'complete', got '%s'", retrieved.Status)
	}
	if retrieved.CompletedAt == "" {
		t.Error("expected CompletedAt to be set")
	}
}

func TestInvestigationRepository_UpdateStatus_NotFound(t *testing.T) {
	db := setupInvestigationTestDB(t)
	repo := sqlite.NewInvestigationRepository(db)
	ctx := context.Background()

	err := repo.UpdateStatus(ctx, "INV-999", "complete", true)
	if err == nil {
		t.Error("expected error for non-existent investigation")
	}
}

func TestInvestigationRepository_AssignWorkbench(t *testing.T) {
	db := setupInvestigationTestDB(t)
	repo := sqlite.NewInvestigationRepository(db)
	ctx := context.Background()

	inv := createTestInvestigation(t, repo, ctx, "COMM-001", "Workbench Test", "")

	err := repo.AssignWorkbench(ctx, inv.ID, "BENCH-001")
	if err != nil {
		t.Fatalf("AssignWorkbench failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, inv.ID)
	if retrieved.AssignedWorkbenchID != "BENCH-001" {
		t.Errorf("expected assigned workbench 'BENCH-001', got '%s'", retrieved.AssignedWorkbenchID)
	}
}

func TestInvestigationRepository_AssignWorkbench_NotFound(t *testing.T) {
	db := setupInvestigationTestDB(t)
	repo := sqlite.NewInvestigationRepository(db)
	ctx := context.Background()

	err := repo.AssignWorkbench(ctx, "INV-999", "BENCH-001")
	if err == nil {
		t.Error("expected error for non-existent investigation")
	}
}

func TestInvestigationRepository_GetByWorkbench(t *testing.T) {
	db := setupInvestigationTestDB(t)
	repo := sqlite.NewInvestigationRepository(db)
	ctx := context.Background()

	inv1 := createTestInvestigation(t, repo, ctx, "COMM-001", "Investigation 1", "")
	inv2 := createTestInvestigation(t, repo, ctx, "COMM-001", "Investigation 2", "")
	createTestInvestigation(t, repo, ctx, "COMM-001", "Investigation 3 (unassigned)", "")

	_ = repo.AssignWorkbench(ctx, inv1.ID, "BENCH-001")
	_ = repo.AssignWorkbench(ctx, inv2.ID, "BENCH-001")

	investigations, err := repo.GetByWorkbench(ctx, "BENCH-001")
	if err != nil {
		t.Fatalf("GetByWorkbench failed: %v", err)
	}

	if len(investigations) != 2 {
		t.Errorf("expected 2 investigations for workbench, got %d", len(investigations))
	}
}

func TestInvestigationRepository_CommissionExists(t *testing.T) {
	db := setupInvestigationTestDB(t)
	repo := sqlite.NewInvestigationRepository(db)
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
