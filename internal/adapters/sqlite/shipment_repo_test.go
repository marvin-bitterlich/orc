package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

// setupShipmentTestDB creates the test database with required seed data.
func setupShipmentTestDB(t *testing.T) *sql.DB {
	t.Helper()
	testDB := setupTestDB(t)
	seedCommission(t, testDB, "COMM-001", "Test Commission")
	return testDB
}

// createTestShipment is a helper that creates a shipment with a generated ID.
func createTestShipment(t *testing.T, repo *sqlite.ShipmentRepository, ctx context.Context, commissionID, title, description string) *secondary.ShipmentRecord {
	t.Helper()

	nextID, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}

	shipment := &secondary.ShipmentRecord{
		ID:           nextID,
		CommissionID: commissionID,
		Title:        title,
		Description:  description,
	}

	err = repo.Create(ctx, shipment)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	return shipment
}

func TestShipmentRepository_Create(t *testing.T) {
	db := setupShipmentTestDB(t)
	repo := sqlite.NewShipmentRepository(db)
	ctx := context.Background()

	shipment := &secondary.ShipmentRecord{
		ID:           "SHIP-001",
		CommissionID: "COMM-001",
		Title:        "Test Shipment",
		Description:  "A test shipment description",
	}

	err := repo.Create(ctx, shipment)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify shipment was created
	retrieved, err := repo.GetByID(ctx, "SHIP-001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Title != "Test Shipment" {
		t.Errorf("expected title 'Test Shipment', got '%s'", retrieved.Title)
	}
	if retrieved.Status != "active" {
		t.Errorf("expected status 'active', got '%s'", retrieved.Status)
	}
}

func TestShipmentRepository_GetByID(t *testing.T) {
	db := setupShipmentTestDB(t)
	repo := sqlite.NewShipmentRepository(db)
	ctx := context.Background()

	// Create a shipment using helper
	shipment := createTestShipment(t, repo, ctx, "COMM-001", "Test Shipment", "Description")

	// Retrieve it
	retrieved, err := repo.GetByID(ctx, shipment.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.Title != "Test Shipment" {
		t.Errorf("expected title 'Test Shipment', got '%s'", retrieved.Title)
	}

	if retrieved.Description != "Description" {
		t.Errorf("expected description 'Description', got '%s'", retrieved.Description)
	}

	if retrieved.Status != "active" {
		t.Errorf("expected status 'active', got '%s'", retrieved.Status)
	}
}

func TestShipmentRepository_GetByID_NotFound(t *testing.T) {
	db := setupShipmentTestDB(t)
	repo := sqlite.NewShipmentRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "SHIP-999")
	if err == nil {
		t.Error("expected error for non-existent shipment")
	}
}

func TestShipmentRepository_List(t *testing.T) {
	db := setupShipmentTestDB(t)
	repo := sqlite.NewShipmentRepository(db)
	ctx := context.Background()

	// Create multiple shipments
	createTestShipment(t, repo, ctx, "COMM-001", "Shipment 1", "")
	createTestShipment(t, repo, ctx, "COMM-001", "Shipment 2", "")
	createTestShipment(t, repo, ctx, "COMM-001", "Shipment 3", "")

	shipments, err := repo.List(ctx, secondary.ShipmentFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(shipments) != 3 {
		t.Errorf("expected 3 shipments, got %d", len(shipments))
	}
}

func TestShipmentRepository_List_FilterByCommission(t *testing.T) {
	db := setupShipmentTestDB(t)
	repo := sqlite.NewShipmentRepository(db)
	ctx := context.Background()

	// Add another commission
	_, _ = db.Exec("INSERT INTO commissions (id, title, status) VALUES ('COMM-002', 'Test Commission 2', 'active')")

	// Create shipments for different commissions
	createTestShipment(t, repo, ctx, "COMM-001", "Shipment 1", "")
	createTestShipment(t, repo, ctx, "COMM-002", "Shipment 2", "")

	// List only COMM-001 shipments
	shipments, err := repo.List(ctx, secondary.ShipmentFilters{CommissionID: "COMM-001"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(shipments) != 1 {
		t.Errorf("expected 1 shipment for COMM-001, got %d", len(shipments))
	}
}

func TestShipmentRepository_List_FilterByStatus(t *testing.T) {
	db := setupShipmentTestDB(t)
	repo := sqlite.NewShipmentRepository(db)
	ctx := context.Background()

	// Create active shipment
	createTestShipment(t, repo, ctx, "COMM-001", "Active Shipment", "")

	// Create and complete a shipment
	s2 := createTestShipment(t, repo, ctx, "COMM-001", "Complete Shipment", "")
	_ = repo.UpdateStatus(ctx, s2.ID, "complete", true)

	// List only active
	shipments, err := repo.List(ctx, secondary.ShipmentFilters{Status: "active"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(shipments) != 1 {
		t.Errorf("expected 1 active shipment, got %d", len(shipments))
	}
}

func TestShipmentRepository_Update(t *testing.T) {
	db := setupShipmentTestDB(t)
	repo := sqlite.NewShipmentRepository(db)
	ctx := context.Background()

	// Create a shipment
	shipment := createTestShipment(t, repo, ctx, "COMM-001", "Original Title", "")

	// Update it
	err := repo.Update(ctx, &secondary.ShipmentRecord{
		ID:    shipment.ID,
		Title: "Updated Title",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	retrieved, _ := repo.GetByID(ctx, shipment.ID)
	if retrieved.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", retrieved.Title)
	}
}

func TestShipmentRepository_Update_NotFound(t *testing.T) {
	db := setupShipmentTestDB(t)
	repo := sqlite.NewShipmentRepository(db)
	ctx := context.Background()

	err := repo.Update(ctx, &secondary.ShipmentRecord{
		ID:    "SHIP-999",
		Title: "Updated Title",
	})
	if err == nil {
		t.Error("expected error for non-existent shipment")
	}
}

func TestShipmentRepository_Delete(t *testing.T) {
	db := setupShipmentTestDB(t)
	repo := sqlite.NewShipmentRepository(db)
	ctx := context.Background()

	// Create a shipment
	shipment := createTestShipment(t, repo, ctx, "COMM-001", "To Delete", "")

	// Delete it
	err := repo.Delete(ctx, shipment.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(ctx, shipment.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestShipmentRepository_Delete_NotFound(t *testing.T) {
	db := setupShipmentTestDB(t)
	repo := sqlite.NewShipmentRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, "SHIP-999")
	if err == nil {
		t.Error("expected error for non-existent shipment")
	}
}

func TestShipmentRepository_Pin_Unpin(t *testing.T) {
	db := setupShipmentTestDB(t)
	repo := sqlite.NewShipmentRepository(db)
	ctx := context.Background()

	// Create a shipment
	shipment := createTestShipment(t, repo, ctx, "COMM-001", "Pin Test", "")

	// Pin it
	err := repo.Pin(ctx, shipment.ID)
	if err != nil {
		t.Fatalf("Pin failed: %v", err)
	}

	// Verify pinned
	retrieved, _ := repo.GetByID(ctx, shipment.ID)
	if !retrieved.Pinned {
		t.Error("expected shipment to be pinned")
	}

	// Unpin it
	err = repo.Unpin(ctx, shipment.ID)
	if err != nil {
		t.Fatalf("Unpin failed: %v", err)
	}

	// Verify unpinned
	retrieved, _ = repo.GetByID(ctx, shipment.ID)
	if retrieved.Pinned {
		t.Error("expected shipment to be unpinned")
	}
}

func TestShipmentRepository_Pin_NotFound(t *testing.T) {
	db := setupShipmentTestDB(t)
	repo := sqlite.NewShipmentRepository(db)
	ctx := context.Background()

	err := repo.Pin(ctx, "SHIP-999")
	if err == nil {
		t.Error("expected error for non-existent shipment")
	}
}

func TestShipmentRepository_GetNextID(t *testing.T) {
	db := setupShipmentTestDB(t)
	repo := sqlite.NewShipmentRepository(db)
	ctx := context.Background()

	// First ID should be SHIP-001
	id, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "SHIP-001" {
		t.Errorf("expected SHIP-001, got %s", id)
	}

	// Create a shipment
	createTestShipment(t, repo, ctx, "COMM-001", "Test", "")

	// Next ID should be SHIP-002
	id, err = repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "SHIP-002" {
		t.Errorf("expected SHIP-002, got %s", id)
	}
}

func TestShipmentRepository_UpdateStatus(t *testing.T) {
	db := setupShipmentTestDB(t)
	repo := sqlite.NewShipmentRepository(db)
	ctx := context.Background()

	// Create a shipment
	shipment := createTestShipment(t, repo, ctx, "COMM-001", "Status Test", "")

	// Update status without completed timestamp
	err := repo.UpdateStatus(ctx, shipment.ID, "in_progress", false)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, shipment.ID)
	if retrieved.Status != "in_progress" {
		t.Errorf("expected status 'in_progress', got '%s'", retrieved.Status)
	}
	if retrieved.CompletedAt != "" {
		t.Error("expected CompletedAt to be empty")
	}

	// Update to complete with timestamp
	err = repo.UpdateStatus(ctx, shipment.ID, "complete", true)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	retrieved, _ = repo.GetByID(ctx, shipment.ID)
	if retrieved.Status != "complete" {
		t.Errorf("expected status 'complete', got '%s'", retrieved.Status)
	}
	if retrieved.CompletedAt == "" {
		t.Error("expected CompletedAt to be set")
	}
}

func TestShipmentRepository_UpdateStatus_NotFound(t *testing.T) {
	db := setupShipmentTestDB(t)
	repo := sqlite.NewShipmentRepository(db)
	ctx := context.Background()

	err := repo.UpdateStatus(ctx, "SHIP-999", "complete", true)
	if err == nil {
		t.Error("expected error for non-existent shipment")
	}
}

func TestShipmentRepository_AssignWorkbench(t *testing.T) {
	db := setupShipmentTestDB(t)
	repo := sqlite.NewShipmentRepository(db)
	ctx := context.Background()

	// Insert a test workbench
	_, _ = db.Exec("INSERT INTO groves (id, commission_id, name, status) VALUES ('BENCH-001', 'COMM-001', 'test-workbench', 'active')")

	// Create a shipment
	shipment := createTestShipment(t, repo, ctx, "COMM-001", "Workbench Test", "")

	// Assign workbench
	err := repo.AssignWorkbench(ctx, shipment.ID, "BENCH-001")
	if err != nil {
		t.Fatalf("AssignWorkbench failed: %v", err)
	}

	// Verify assignment
	retrieved, _ := repo.GetByID(ctx, shipment.ID)
	if retrieved.AssignedWorkbenchID != "BENCH-001" {
		t.Errorf("expected assigned workbench 'BENCH-001', got '%s'", retrieved.AssignedWorkbenchID)
	}
}

func TestShipmentRepository_AssignWorkbench_NotFound(t *testing.T) {
	db := setupShipmentTestDB(t)
	repo := sqlite.NewShipmentRepository(db)
	ctx := context.Background()

	err := repo.AssignWorkbench(ctx, "SHIP-999", "BENCH-001")
	if err == nil {
		t.Error("expected error for non-existent shipment")
	}
}

func TestShipmentRepository_GetByWorkbench(t *testing.T) {
	db := setupShipmentTestDB(t)
	repo := sqlite.NewShipmentRepository(db)
	ctx := context.Background()

	// Insert a test workbench
	_, _ = db.Exec("INSERT INTO groves (id, commission_id, name, status) VALUES ('BENCH-001', 'COMM-001', 'test-workbench', 'active')")

	// Create shipments and assign to workbench
	s1 := createTestShipment(t, repo, ctx, "COMM-001", "Ship 1", "")
	s2 := createTestShipment(t, repo, ctx, "COMM-001", "Ship 2", "")
	createTestShipment(t, repo, ctx, "COMM-001", "Ship 3 (unassigned)", "")

	_ = repo.AssignWorkbench(ctx, s1.ID, "BENCH-001")
	_ = repo.AssignWorkbench(ctx, s2.ID, "BENCH-001")

	// Get by workbench
	shipments, err := repo.GetByWorkbench(ctx, "BENCH-001")
	if err != nil {
		t.Fatalf("GetByWorkbench failed: %v", err)
	}

	if len(shipments) != 2 {
		t.Errorf("expected 2 shipments for workbench, got %d", len(shipments))
	}
}

func TestShipmentRepository_CommissionExists(t *testing.T) {
	db := setupShipmentTestDB(t)
	repo := sqlite.NewShipmentRepository(db)
	ctx := context.Background()

	// Existing commission
	exists, err := repo.CommissionExists(ctx, "COMM-001")
	if err != nil {
		t.Fatalf("CommissionExists failed: %v", err)
	}
	if !exists {
		t.Error("expected commission to exist")
	}

	// Non-existing commission
	exists, err = repo.CommissionExists(ctx, "COMM-999")
	if err != nil {
		t.Fatalf("CommissionExists failed: %v", err)
	}
	if exists {
		t.Error("expected commission to not exist")
	}
}

func TestShipmentRepository_WorkbenchAssignedToOther(t *testing.T) {
	db := setupShipmentTestDB(t)
	repo := sqlite.NewShipmentRepository(db)
	ctx := context.Background()

	// Insert a test workbench
	_, _ = db.Exec("INSERT INTO groves (id, commission_id, name, status) VALUES ('BENCH-001', 'COMM-001', 'test-workbench', 'active')")

	// Create two shipments
	s1 := createTestShipment(t, repo, ctx, "COMM-001", "Ship 1", "")
	s2 := createTestShipment(t, repo, ctx, "COMM-001", "Ship 2", "")

	// Assign workbench to s1
	_ = repo.AssignWorkbench(ctx, s1.ID, "BENCH-001")

	// Check if workbench is assigned to another shipment (excluding s1)
	otherID, err := repo.WorkbenchAssignedToOther(ctx, "BENCH-001", s1.ID)
	if err != nil {
		t.Fatalf("WorkbenchAssignedToOther failed: %v", err)
	}
	if otherID != "" {
		t.Error("expected no other shipment to have the workbench")
	}

	// Check from s2's perspective
	otherID, err = repo.WorkbenchAssignedToOther(ctx, "BENCH-001", s2.ID)
	if err != nil {
		t.Fatalf("WorkbenchAssignedToOther failed: %v", err)
	}
	if otherID != s1.ID {
		t.Errorf("expected s1 to have the workbench, got '%s'", otherID)
	}
}
