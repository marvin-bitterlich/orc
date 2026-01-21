package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func setupOperationTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// Create missions table
	_, err = db.Exec(`
		CREATE TABLE missions (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create missions table: %v", err)
	}

	// Create operations table
	_, err = db.Exec(`
		CREATE TABLE operations (
			id TEXT PRIMARY KEY,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL DEFAULT 'ready',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME
		)
	`)
	if err != nil {
		t.Fatalf("failed to create operations table: %v", err)
	}

	// Insert test mission
	_, _ = db.Exec("INSERT INTO missions (id, title, status) VALUES ('MISSION-001', 'Test Mission', 'active')")

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// createTestOperation is a helper that creates an operation with a generated ID.
func createTestOperation(t *testing.T, repo *sqlite.OperationRepository, ctx context.Context, missionID, title, description string) *secondary.OperationRecord {
	t.Helper()

	nextID, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}

	op := &secondary.OperationRecord{
		ID:          nextID,
		MissionID:   missionID,
		Title:       title,
		Description: description,
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
		ID:          "OP-001",
		MissionID:   "MISSION-001",
		Title:       "Test Operation",
		Description: "A test operation description",
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

	op := createTestOperation(t, repo, ctx, "MISSION-001", "Test Operation", "Description")

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

	createTestOperation(t, repo, ctx, "MISSION-001", "Operation 1", "")
	createTestOperation(t, repo, ctx, "MISSION-001", "Operation 2", "")
	createTestOperation(t, repo, ctx, "MISSION-001", "Operation 3", "")

	operations, err := repo.List(ctx, secondary.OperationFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(operations) != 3 {
		t.Errorf("expected 3 operations, got %d", len(operations))
	}
}

func TestOperationRepository_List_FilterByMission(t *testing.T) {
	db := setupOperationTestDB(t)
	repo := sqlite.NewOperationRepository(db)
	ctx := context.Background()

	// Add another mission
	_, _ = db.Exec("INSERT INTO missions (id, title, status) VALUES ('MISSION-002', 'Mission 2', 'active')")

	createTestOperation(t, repo, ctx, "MISSION-001", "Operation 1", "")
	createTestOperation(t, repo, ctx, "MISSION-002", "Operation 2", "")

	operations, err := repo.List(ctx, secondary.OperationFilters{MissionID: "MISSION-001"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(operations) != 1 {
		t.Errorf("expected 1 operation for MISSION-001, got %d", len(operations))
	}
}

func TestOperationRepository_List_FilterByStatus(t *testing.T) {
	db := setupOperationTestDB(t)
	repo := sqlite.NewOperationRepository(db)
	ctx := context.Background()

	op1 := createTestOperation(t, repo, ctx, "MISSION-001", "Ready Operation", "")
	createTestOperation(t, repo, ctx, "MISSION-001", "Another Ready", "")

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

	op := createTestOperation(t, repo, ctx, "MISSION-001", "Status Test", "")

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

	createTestOperation(t, repo, ctx, "MISSION-001", "Test", "")

	id, err = repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "OP-002" {
		t.Errorf("expected OP-002, got %s", id)
	}
}

func TestOperationRepository_MissionExists(t *testing.T) {
	db := setupOperationTestDB(t)
	repo := sqlite.NewOperationRepository(db)
	ctx := context.Background()

	exists, err := repo.MissionExists(ctx, "MISSION-001")
	if err != nil {
		t.Fatalf("MissionExists failed: %v", err)
	}
	if !exists {
		t.Error("expected mission to exist")
	}

	exists, err = repo.MissionExists(ctx, "MISSION-999")
	if err != nil {
		t.Fatalf("MissionExists failed: %v", err)
	}
	if exists {
		t.Error("expected mission to not exist")
	}
}
