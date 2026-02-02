package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	corecommission "github.com/example/orc/internal/core/commission"
	"github.com/example/orc/internal/ports/secondary"
)

// createTestCommission is a helper that simulates service-layer behavior:
// gets next ID, sets initial status, then creates.
func createTestCommission(t *testing.T, repo *sqlite.CommissionRepository, ctx context.Context, title, description string) *secondary.CommissionRecord {
	t.Helper()

	// Service would call GetNextID first
	nextID, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}

	commission := &secondary.CommissionRecord{
		ID:          nextID,
		Title:       title,
		Description: description,
		Status:      string(corecommission.InitialStatus()),
	}

	err = repo.Create(ctx, commission)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	return commission
}

func TestCommissionRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCommissionRepository(db)
	ctx := context.Background()

	// Pre-populate ID and Status as service layer would
	commission := &secondary.CommissionRecord{
		ID:          "COMM-001",
		Title:       "Test Commission",
		Description: "A test commission description",
		Status:      "active",
	}

	err := repo.Create(ctx, commission)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify commission was created
	retrieved, err := repo.GetByID(ctx, "COMM-001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Title != "Test Commission" {
		t.Errorf("expected title 'Test Commission', got '%s'", retrieved.Title)
	}
}

func TestCommissionRepository_Create_WithWorkshopID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCommissionRepository(db)
	ctx := context.Background()

	// Create a workshop first (FK constraint)
	_, _ = db.Exec("INSERT INTO factories (id, name) VALUES (?, ?)", "FACT-001", "Test Factory")
	_, _ = db.Exec("INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)", "WORK-001", "FACT-001", "Test Workshop", "active")

	// Create commission with workshop_id
	commission := &secondary.CommissionRecord{
		ID:         "COMM-001",
		WorkshopID: "WORK-001",
		Title:      "Workshop Commission",
		Status:     "active",
	}

	err := repo.Create(ctx, commission)
	if err != nil {
		t.Fatalf("Create with workshop_id failed: %v", err)
	}

	// Verify workshop_id was persisted
	retrieved, err := repo.GetByID(ctx, "COMM-001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.WorkshopID != "WORK-001" {
		t.Errorf("expected workshop_id 'WORK-001', got '%s'", retrieved.WorkshopID)
	}
}

func TestCommissionRepository_Create_WithoutWorkshopID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCommissionRepository(db)
	ctx := context.Background()

	// Create commission without workshop_id (should be allowed - nullable)
	commission := &secondary.CommissionRecord{
		ID:     "COMM-001",
		Title:  "No Workshop Commission",
		Status: "active",
	}

	err := repo.Create(ctx, commission)
	if err != nil {
		t.Fatalf("Create without workshop_id failed: %v", err)
	}

	// Verify workshop_id is empty
	retrieved, err := repo.GetByID(ctx, "COMM-001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.WorkshopID != "" {
		t.Errorf("expected empty workshop_id, got '%s'", retrieved.WorkshopID)
	}
}

func TestCommissionRepository_Create_RequiresID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCommissionRepository(db)
	ctx := context.Background()

	// Missing ID should fail
	commission := &secondary.CommissionRecord{
		Title:  "Test Commission",
		Status: "active",
	}

	err := repo.Create(ctx, commission)
	if err == nil {
		t.Error("expected error for missing ID")
	}
}

func TestCommissionRepository_Create_RequiresStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCommissionRepository(db)
	ctx := context.Background()

	// Missing Status should fail
	commission := &secondary.CommissionRecord{
		ID:    "COMM-001",
		Title: "Test Commission",
	}

	err := repo.Create(ctx, commission)
	if err == nil {
		t.Error("expected error for missing Status")
	}
}

func TestCommissionRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCommissionRepository(db)
	ctx := context.Background()

	// Create a commission using helper
	commission := createTestCommission(t, repo, ctx, "Test Commission", "Description")

	// Retrieve it
	retrieved, err := repo.GetByID(ctx, commission.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.Title != "Test Commission" {
		t.Errorf("expected title 'Test Commission', got '%s'", retrieved.Title)
	}

	if retrieved.Description != "Description" {
		t.Errorf("expected description 'Description', got '%s'", retrieved.Description)
	}

	if retrieved.Status != "active" {
		t.Errorf("expected status 'active', got '%s'", retrieved.Status)
	}
}

func TestCommissionRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCommissionRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "COMM-999")
	if err == nil {
		t.Error("expected error for non-existent commission")
	}
}

func TestCommissionRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCommissionRepository(db)
	ctx := context.Background()

	// Create multiple commissions using helper
	createTestCommission(t, repo, ctx, "Commission 1", "")
	createTestCommission(t, repo, ctx, "Commission 2", "")
	createTestCommission(t, repo, ctx, "Commission 3", "")

	commissions, err := repo.List(ctx, secondary.CommissionFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(commissions) != 3 {
		t.Errorf("expected 3 commissions, got %d", len(commissions))
	}
}

func TestCommissionRepository_List_FilterByStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCommissionRepository(db)
	ctx := context.Background()

	// Create active commission
	createTestCommission(t, repo, ctx, "Active Commission", "")

	// Create and complete a commission
	m2 := createTestCommission(t, repo, ctx, "Complete Commission", "")
	_ = repo.Update(ctx, &secondary.CommissionRecord{ID: m2.ID, Status: "complete"})

	// List only active
	commissions, err := repo.List(ctx, secondary.CommissionFilters{Status: "active"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(commissions) != 1 {
		t.Errorf("expected 1 active commission, got %d", len(commissions))
	}
}

func TestCommissionRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCommissionRepository(db)
	ctx := context.Background()

	// Create a commission
	commission := createTestCommission(t, repo, ctx, "Original Title", "")

	// Update it
	err := repo.Update(ctx, &secondary.CommissionRecord{
		ID:    commission.ID,
		Title: "Updated Title",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	retrieved, _ := repo.GetByID(ctx, commission.ID)
	if retrieved.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", retrieved.Title)
	}
}

func TestCommissionRepository_Update_CompletedAt(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCommissionRepository(db)
	ctx := context.Background()

	// Create a commission
	commission := createTestCommission(t, repo, ctx, "Test", "")

	// Update with CompletedAt (service layer would set this)
	err := repo.Update(ctx, &secondary.CommissionRecord{
		ID:          commission.ID,
		Status:      "complete",
		CompletedAt: "2026-01-20T12:00:00Z",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify CompletedAt was set
	retrieved, _ := repo.GetByID(ctx, commission.ID)
	if retrieved.CompletedAt == "" {
		t.Error("expected CompletedAt to be set")
	}
	if retrieved.Status != "complete" {
		t.Errorf("expected status 'complete', got '%s'", retrieved.Status)
	}
}

func TestCommissionRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCommissionRepository(db)
	ctx := context.Background()

	// Create a commission
	commission := createTestCommission(t, repo, ctx, "To Delete", "")

	// Delete it
	err := repo.Delete(ctx, commission.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(ctx, commission.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestCommissionRepository_Pin_Unpin(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCommissionRepository(db)
	ctx := context.Background()

	// Create a commission
	commission := createTestCommission(t, repo, ctx, "Pin Test", "")

	// Pin it
	err := repo.Pin(ctx, commission.ID)
	if err != nil {
		t.Fatalf("Pin failed: %v", err)
	}

	// Verify pinned
	retrieved, _ := repo.GetByID(ctx, commission.ID)
	if !retrieved.Pinned {
		t.Error("expected commission to be pinned")
	}

	// Unpin it
	err = repo.Unpin(ctx, commission.ID)
	if err != nil {
		t.Fatalf("Unpin failed: %v", err)
	}

	// Verify unpinned
	retrieved, _ = repo.GetByID(ctx, commission.ID)
	if retrieved.Pinned {
		t.Error("expected commission to be unpinned")
	}
}

func TestCommissionRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCommissionRepository(db)
	ctx := context.Background()

	// First ID should be COMM-001
	id, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "COMM-001" {
		t.Errorf("expected COMM-001, got %s", id)
	}

	// Create a commission using that ID
	createTestCommission(t, repo, ctx, "Test", "")

	// Next ID should be COMM-002
	id, err = repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "COMM-002" {
		t.Errorf("expected COMM-002, got %s", id)
	}
}

func TestCommissionRepository_CountShipments(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCommissionRepository(db)
	ctx := context.Background()

	// Create a commission
	commission := createTestCommission(t, repo, ctx, "Test", "")

	// Count should be 0
	count, err := repo.CountShipments(ctx, commission.ID)
	if err != nil {
		t.Fatalf("CountShipments failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 shipments, got %d", count)
	}

	// Add shipments
	_, _ = db.Exec("INSERT INTO shipments (id, commission_id, title) VALUES (?, ?, ?)", "SHIP-001", commission.ID, "Ship 1")
	_, _ = db.Exec("INSERT INTO shipments (id, commission_id, title) VALUES (?, ?, ?)", "SHIP-002", commission.ID, "Ship 2")

	// Count should be 2
	count, err = repo.CountShipments(ctx, commission.ID)
	if err != nil {
		t.Fatalf("CountShipments failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 shipments, got %d", count)
	}
}
