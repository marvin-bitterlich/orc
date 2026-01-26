package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

// setupConclaveTestDB creates the test database with required seed data.
func setupConclaveTestDB(t *testing.T) *sql.DB {
	t.Helper()
	testDB := setupTestDB(t)
	seedCommission(t, testDB, "COMM-001", "Test Commission")
	seedShipment(t, testDB, "SHIP-001", "COMM-001", "Test Shipment")
	return testDB
}

// createTestConclave is a helper that creates a conclave with a generated ID.
func createTestConclave(t *testing.T, repo *sqlite.ConclaveRepository, ctx context.Context, commissionID, title, description string) *secondary.ConclaveRecord {
	t.Helper()

	nextID, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}

	conclave := &secondary.ConclaveRecord{
		ID:           nextID,
		CommissionID: commissionID,
		Title:        title,
		Description:  description,
	}

	err = repo.Create(ctx, conclave)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	return conclave
}

func TestConclaveRepository_Create(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	conclave := &secondary.ConclaveRecord{
		ID:           "CON-001",
		CommissionID: "COMM-001",
		Title:        "Test Conclave",
		Description:  "A test conclave description",
	}

	err := repo.Create(ctx, conclave)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify conclave was created
	retrieved, err := repo.GetByID(ctx, "CON-001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Title != "Test Conclave" {
		t.Errorf("expected title 'Test Conclave', got '%s'", retrieved.Title)
	}
	if retrieved.Status != "open" {
		t.Errorf("expected status 'open', got '%s'", retrieved.Status)
	}
}

func TestConclaveRepository_GetByID(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	conclave := createTestConclave(t, repo, ctx, "COMM-001", "Test Conclave", "Description")

	retrieved, err := repo.GetByID(ctx, conclave.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.Title != "Test Conclave" {
		t.Errorf("expected title 'Test Conclave', got '%s'", retrieved.Title)
	}
	if retrieved.Description != "Description" {
		t.Errorf("expected description 'Description', got '%s'", retrieved.Description)
	}
}

func TestConclaveRepository_GetByID_NotFound(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "CON-999")
	if err == nil {
		t.Error("expected error for non-existent conclave")
	}
}

func TestConclaveRepository_List(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	createTestConclave(t, repo, ctx, "COMM-001", "Conclave 1", "")
	createTestConclave(t, repo, ctx, "COMM-001", "Conclave 2", "")
	createTestConclave(t, repo, ctx, "COMM-001", "Conclave 3", "")

	conclaves, err := repo.List(ctx, secondary.ConclaveFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(conclaves) != 3 {
		t.Errorf("expected 3 conclaves, got %d", len(conclaves))
	}
}

func TestConclaveRepository_List_FilterByCommission(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	// Add another commission
	_, _ = db.Exec("INSERT INTO commissions (id, title, status) VALUES ('COMM-002', 'Commission 2', 'active')")

	createTestConclave(t, repo, ctx, "COMM-001", "Conclave 1", "")
	createTestConclave(t, repo, ctx, "COMM-002", "Conclave 2", "")

	conclaves, err := repo.List(ctx, secondary.ConclaveFilters{CommissionID: "COMM-001"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(conclaves) != 1 {
		t.Errorf("expected 1 conclave for COMM-001, got %d", len(conclaves))
	}
}

func TestConclaveRepository_List_FilterByStatus(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	c1 := createTestConclave(t, repo, ctx, "COMM-001", "Active Conclave", "")
	createTestConclave(t, repo, ctx, "COMM-001", "Another Active", "")

	// Close c1 (conclave statuses are: open, paused, closed)
	_ = repo.UpdateStatus(ctx, c1.ID, "closed", true)

	conclaves, err := repo.List(ctx, secondary.ConclaveFilters{Status: "open"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(conclaves) != 1 {
		t.Errorf("expected 1 open conclave, got %d", len(conclaves))
	}
}

func TestConclaveRepository_Update(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	conclave := createTestConclave(t, repo, ctx, "COMM-001", "Original Title", "")

	err := repo.Update(ctx, &secondary.ConclaveRecord{
		ID:    conclave.ID,
		Title: "Updated Title",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, conclave.ID)
	if retrieved.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", retrieved.Title)
	}
}

func TestConclaveRepository_Update_NotFound(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	err := repo.Update(ctx, &secondary.ConclaveRecord{
		ID:    "CON-999",
		Title: "Updated Title",
	})
	if err == nil {
		t.Error("expected error for non-existent conclave")
	}
}

func TestConclaveRepository_Delete(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	conclave := createTestConclave(t, repo, ctx, "COMM-001", "To Delete", "")

	err := repo.Delete(ctx, conclave.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.GetByID(ctx, conclave.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestConclaveRepository_Delete_NotFound(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, "CON-999")
	if err == nil {
		t.Error("expected error for non-existent conclave")
	}
}

func TestConclaveRepository_Pin_Unpin(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	conclave := createTestConclave(t, repo, ctx, "COMM-001", "Pin Test", "")

	// Pin
	err := repo.Pin(ctx, conclave.ID)
	if err != nil {
		t.Fatalf("Pin failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, conclave.ID)
	if !retrieved.Pinned {
		t.Error("expected conclave to be pinned")
	}

	// Unpin
	err = repo.Unpin(ctx, conclave.ID)
	if err != nil {
		t.Fatalf("Unpin failed: %v", err)
	}

	retrieved, _ = repo.GetByID(ctx, conclave.ID)
	if retrieved.Pinned {
		t.Error("expected conclave to be unpinned")
	}
}

func TestConclaveRepository_Pin_NotFound(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	err := repo.Pin(ctx, "CON-999")
	if err == nil {
		t.Error("expected error for non-existent conclave")
	}
}

func TestConclaveRepository_GetNextID(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	id, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "CON-001" {
		t.Errorf("expected CON-001, got %s", id)
	}

	createTestConclave(t, repo, ctx, "COMM-001", "Test", "")

	id, err = repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "CON-002" {
		t.Errorf("expected CON-002, got %s", id)
	}
}

func TestConclaveRepository_UpdateStatus(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	conclave := createTestConclave(t, repo, ctx, "COMM-001", "Status Test", "")

	// Update status without completed timestamp
	err := repo.UpdateStatus(ctx, conclave.ID, "paused", false)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, conclave.ID)
	if retrieved.Status != "paused" {
		t.Errorf("expected status 'paused', got '%s'", retrieved.Status)
	}
	if retrieved.DecidedAt != "" {
		t.Error("expected DecidedAt to be empty")
	}

	// Update to closed (with decided timestamp)
	err = repo.UpdateStatus(ctx, conclave.ID, "closed", true)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	retrieved, _ = repo.GetByID(ctx, conclave.ID)
	if retrieved.Status != "closed" {
		t.Errorf("expected status 'closed', got '%s'", retrieved.Status)
	}
	if retrieved.DecidedAt == "" {
		t.Error("expected DecidedAt to be set")
	}
}

func TestConclaveRepository_UpdateStatus_NotFound(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	err := repo.UpdateStatus(ctx, "CON-999", "complete", true)
	if err == nil {
		t.Error("expected error for non-existent conclave")
	}
}

// Note: GetByWorkbench was removed - conclaves are now tied to shipments, not workbenches

func TestConclaveRepository_CommissionExists(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
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

// Multi-entity query tests

func TestConclaveRepository_GetTasksByConclave(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	conclave := createTestConclave(t, repo, ctx, "COMM-001", "Conclave with Tasks", "")

	// Link conclave to shipment SHIP-001
	_, _ = db.Exec("UPDATE conclaves SET shipment_id = 'SHIP-001' WHERE id = ?", conclave.ID)

	// Insert tasks for the shipment (tasks link to shipments, not conclaves directly)
	_, _ = db.Exec(`INSERT INTO tasks (id, shipment_id, commission_id, title, status) VALUES ('TASK-001', 'SHIP-001', 'COMM-001', 'Task 1', 'ready')`)
	_, _ = db.Exec(`INSERT INTO tasks (id, shipment_id, commission_id, title, status) VALUES ('TASK-002', 'SHIP-001', 'COMM-001', 'Task 2', 'ready')`)
	_, _ = db.Exec(`INSERT INTO tasks (id, commission_id, title, status) VALUES ('TASK-003', 'COMM-001', 'Task 3 (no shipment)', 'ready')`)

	tasks, err := repo.GetTasksByConclave(ctx, conclave.ID)
	if err != nil {
		t.Fatalf("GetTasksByConclave failed: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks for conclave, got %d", len(tasks))
	}

	// Verify task data
	if len(tasks) > 0 && tasks[0].Title != "Task 1" {
		t.Errorf("expected title 'Task 1', got '%s'", tasks[0].Title)
	}
}

func TestConclaveRepository_GetPlansByConclave(t *testing.T) {
	db := setupConclaveTestDB(t)
	repo := sqlite.NewConclaveRepository(db)
	ctx := context.Background()

	conclave := createTestConclave(t, repo, ctx, "COMM-001", "Conclave with Plans", "")

	// Link conclave to shipment SHIP-001
	_, _ = db.Exec("UPDATE conclaves SET shipment_id = 'SHIP-001' WHERE id = ?", conclave.ID)

	// Insert plans for the shipment (plans link to shipments, not conclaves directly)
	_, _ = db.Exec(`INSERT INTO plans (id, shipment_id, commission_id, title, status) VALUES ('PLAN-001', 'SHIP-001', 'COMM-001', 'Plan 1', 'draft')`)
	_, _ = db.Exec(`INSERT INTO plans (id, shipment_id, commission_id, title, status) VALUES ('PLAN-002', 'SHIP-001', 'COMM-001', 'Plan 2', 'draft')`)
	_, _ = db.Exec(`INSERT INTO plans (id, commission_id, title, status) VALUES ('PLAN-003', 'COMM-001', 'Plan 3 (no shipment)', 'draft')`)

	plans, err := repo.GetPlansByConclave(ctx, conclave.ID)
	if err != nil {
		t.Fatalf("GetPlansByConclave failed: %v", err)
	}

	if len(plans) != 2 {
		t.Errorf("expected 2 plans for conclave, got %d", len(plans))
	}

	// Verify plan data
	if len(plans) > 0 && plans[0].Title != "Plan 1" {
		t.Errorf("expected title 'Plan 1', got '%s'", plans[0].Title)
	}
}
