package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

// setupOperationTestDB creates the test database with required seed data.
func setupOperationTestDB(t *testing.T) *sql.DB {
	t.Helper()
	testDB := setupTestDB(t)
	seedCommission(t, testDB, "COMM-001", "Test Commission")
	return testDB
}

// createTestOperation is a helper that creates an operation with a generated ID.
func createTestOperation(t *testing.T, repo *sqlite.OperationRepository, ctx context.Context, commissionID, title, description string) *secondary.OperationRecord {
	t.Helper()

	nextID, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}

	op := &secondary.OperationRecord{
		ID:           nextID,
		CommissionID: commissionID,
		Title:        title,
		Description:  description,
	}

	err = repo.Create(ctx, op)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	return op
}

func TestOperationRepository_Create(t *testing.T) {
	db := setupOperationTestDB(t)
	repo := sqlite.NewOperationRepository(db)
	ctx := context.Background()

	op := &secondary.OperationRecord{
		ID:           "OP-001",
		CommissionID: "COMM-001",
		Title:        "Test Operation",
		Description:  "A test operation description",
	}

	err := repo.Create(ctx, op)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify operation was created
	retrieved, err := repo.GetByID(ctx, "OP-001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Title != "Test Operation" {
		t.Errorf("expected title 'Test Operation', got '%s'", retrieved.Title)
	}
	if retrieved.Status != "ready" {
		t.Errorf("expected status 'ready', got '%s'", retrieved.Status)
	}
}

func TestOperationRepository_GetByID(t *testing.T) {
	db := setupOperationTestDB(t)
	repo := sqlite.NewOperationRepository(db)
	ctx := context.Background()

	op := createTestOperation(t, repo, ctx, "COMM-001", "Test Operation", "Description")

	retrieved, err := repo.GetByID(ctx, op.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.Title != "Test Operation" {
		t.Errorf("expected title 'Test Operation', got '%s'", retrieved.Title)
	}
	if retrieved.Description != "Description" {
		t.Errorf("expected description 'Description', got '%s'", retrieved.Description)
	}
}

func TestOperationRepository_GetByID_NotFound(t *testing.T) {
	db := setupOperationTestDB(t)
	repo := sqlite.NewOperationRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "OP-999")
	if err == nil {
		t.Error("expected error for non-existent operation")
	}
}

func TestOperationRepository_List(t *testing.T) {
	db := setupOperationTestDB(t)
	repo := sqlite.NewOperationRepository(db)
	ctx := context.Background()

	createTestOperation(t, repo, ctx, "COMM-001", "Operation 1", "")
	createTestOperation(t, repo, ctx, "COMM-001", "Operation 2", "")
	createTestOperation(t, repo, ctx, "COMM-001", "Operation 3", "")

	operations, err := repo.List(ctx, secondary.OperationFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(operations) != 3 {
		t.Errorf("expected 3 operations, got %d", len(operations))
	}
}

func TestOperationRepository_List_FilterByCommission(t *testing.T) {
	db := setupOperationTestDB(t)
	repo := sqlite.NewOperationRepository(db)
	ctx := context.Background()

	// Add another commission
	_, _ = db.Exec("INSERT INTO commissions (id, title, status) VALUES ('COMM-002', 'Commission 2', 'active')")

	createTestOperation(t, repo, ctx, "COMM-001", "Operation 1", "")
	createTestOperation(t, repo, ctx, "COMM-002", "Operation 2", "")

	operations, err := repo.List(ctx, secondary.OperationFilters{CommissionID: "COMM-001"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(operations) != 1 {
		t.Errorf("expected 1 operation for COMM-001, got %d", len(operations))
	}
}

func TestOperationRepository_List_FilterByStatus(t *testing.T) {
	db := setupOperationTestDB(t)
	repo := sqlite.NewOperationRepository(db)
	ctx := context.Background()

	op1 := createTestOperation(t, repo, ctx, "COMM-001", "Ready Operation", "")
	createTestOperation(t, repo, ctx, "COMM-001", "Another Ready", "")

	// Complete op1
	_ = repo.UpdateStatus(ctx, op1.ID, "complete", true)

	operations, err := repo.List(ctx, secondary.OperationFilters{Status: "ready"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(operations) != 1 {
		t.Errorf("expected 1 ready operation, got %d", len(operations))
	}
}

func TestOperationRepository_UpdateStatus(t *testing.T) {
	db := setupOperationTestDB(t)
	repo := sqlite.NewOperationRepository(db)
	ctx := context.Background()

	op := createTestOperation(t, repo, ctx, "COMM-001", "Status Test", "")

	// Update status without completed timestamp
	err := repo.UpdateStatus(ctx, op.ID, "in_progress", false)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, op.ID)
	if retrieved.Status != "in_progress" {
		t.Errorf("expected status 'in_progress', got '%s'", retrieved.Status)
	}
	if retrieved.CompletedAt != "" {
		t.Error("expected CompletedAt to be empty")
	}

	// Update to complete
	err = repo.UpdateStatus(ctx, op.ID, "complete", true)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	retrieved, _ = repo.GetByID(ctx, op.ID)
	if retrieved.Status != "complete" {
		t.Errorf("expected status 'complete', got '%s'", retrieved.Status)
	}
	if retrieved.CompletedAt == "" {
		t.Error("expected CompletedAt to be set")
	}
}

func TestOperationRepository_UpdateStatus_NotFound(t *testing.T) {
	db := setupOperationTestDB(t)
	repo := sqlite.NewOperationRepository(db)
	ctx := context.Background()

	err := repo.UpdateStatus(ctx, "OP-999", "complete", true)
	if err == nil {
		t.Error("expected error for non-existent operation")
	}
}

func TestOperationRepository_GetNextID(t *testing.T) {
	db := setupOperationTestDB(t)
	repo := sqlite.NewOperationRepository(db)
	ctx := context.Background()

	id, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "OP-001" {
		t.Errorf("expected OP-001, got %s", id)
	}

	createTestOperation(t, repo, ctx, "COMM-001", "Test", "")

	id, err = repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "OP-002" {
		t.Errorf("expected OP-002, got %s", id)
	}
}

func TestOperationRepository_CommissionExists(t *testing.T) {
	db := setupOperationTestDB(t)
	repo := sqlite.NewOperationRepository(db)
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
