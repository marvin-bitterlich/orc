package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

// setupHandoffTestDB creates the test database (no seed data required).
func setupHandoffTestDB(t *testing.T) *sql.DB {
	t.Helper()
	return setupTestDB(t)
}

// createTestHandoff is a helper that creates a handoff with a generated ID.
func createTestHandoff(t *testing.T, repo *sqlite.HandoffRepository, ctx context.Context, note, commissionID, workbenchID, todos string) *secondary.HandoffRecord {
	t.Helper()

	nextID, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}

	handoff := &secondary.HandoffRecord{
		ID:                 nextID,
		HandoffNote:        note,
		ActiveCommissionID: commissionID,
		ActiveWorkbenchID:  workbenchID,
		TodosSnapshot:      todos,
	}

	err = repo.Create(ctx, handoff)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	return handoff
}

func TestHandoffRepository_Create(t *testing.T) {
	db := setupHandoffTestDB(t)
	repo := sqlite.NewHandoffRepository(db)
	ctx := context.Background()

	handoff := &secondary.HandoffRecord{
		ID:                 "HO-001",
		HandoffNote:        "Session complete. Next steps: review plan.",
		ActiveCommissionID: "COMM-001",
		ActiveWorkbenchID:  "WB-001",
		TodosSnapshot:      "[]",
	}

	err := repo.Create(ctx, handoff)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify handoff was created
	retrieved, err := repo.GetByID(ctx, "HO-001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.HandoffNote != "Session complete. Next steps: review plan." {
		t.Errorf("expected correct handoff note, got '%s'", retrieved.HandoffNote)
	}
	if retrieved.ActiveCommissionID != "COMM-001" {
		t.Errorf("expected commission 'COMM-001', got '%s'", retrieved.ActiveCommissionID)
	}
	if retrieved.ActiveWorkbenchID != "WB-001" {
		t.Errorf("expected workbench 'WB-001', got '%s'", retrieved.ActiveWorkbenchID)
	}
}

func TestHandoffRepository_Create_MinimalData(t *testing.T) {
	db := setupHandoffTestDB(t)
	repo := sqlite.NewHandoffRepository(db)
	ctx := context.Background()

	handoff := &secondary.HandoffRecord{
		ID:          "HO-001",
		HandoffNote: "Minimal handoff",
	}

	err := repo.Create(ctx, handoff)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, "HO-001")
	if retrieved.ActiveCommissionID != "" {
		t.Errorf("expected empty commission ID, got '%s'", retrieved.ActiveCommissionID)
	}
	if retrieved.ActiveWorkbenchID != "" {
		t.Errorf("expected empty workbench ID, got '%s'", retrieved.ActiveWorkbenchID)
	}
}

func TestHandoffRepository_GetByID(t *testing.T) {
	db := setupHandoffTestDB(t)
	repo := sqlite.NewHandoffRepository(db)
	ctx := context.Background()

	handoff := createTestHandoff(t, repo, ctx, "Test note", "COMM-001", "WB-001", "[]")

	retrieved, err := repo.GetByID(ctx, handoff.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.HandoffNote != "Test note" {
		t.Errorf("expected note 'Test note', got '%s'", retrieved.HandoffNote)
	}
	if retrieved.CreatedAt == "" {
		t.Error("expected CreatedAt to be set")
	}
}

func TestHandoffRepository_GetByID_NotFound(t *testing.T) {
	db := setupHandoffTestDB(t)
	repo := sqlite.NewHandoffRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "HO-999")
	if err == nil {
		t.Error("expected error for non-existent handoff")
	}
}

func TestHandoffRepository_GetLatest(t *testing.T) {
	db := setupHandoffTestDB(t)
	repo := sqlite.NewHandoffRepository(db)
	ctx := context.Background()

	// Create handoffs with explicit timestamps to ensure ordering
	_, _ = db.Exec(`INSERT INTO handoffs (id, handoff_note, created_at) VALUES ('HO-001', 'First handoff', '2024-01-01 10:00:00')`)
	_, _ = db.Exec(`INSERT INTO handoffs (id, handoff_note, created_at) VALUES ('HO-002', 'Second handoff', '2024-01-01 11:00:00')`)
	_, _ = db.Exec(`INSERT INTO handoffs (id, handoff_note, active_commission_id, created_at) VALUES ('HO-003', 'Latest handoff', 'COMM-001', '2024-01-01 12:00:00')`)

	latest, err := repo.GetLatest(ctx)
	if err != nil {
		t.Fatalf("GetLatest failed: %v", err)
	}

	if latest.ID != "HO-003" {
		t.Errorf("expected latest handoff ID 'HO-003', got '%s'", latest.ID)
	}
	if latest.HandoffNote != "Latest handoff" {
		t.Errorf("expected note 'Latest handoff', got '%s'", latest.HandoffNote)
	}
}

func TestHandoffRepository_GetLatest_NoHandoffs(t *testing.T) {
	db := setupHandoffTestDB(t)
	repo := sqlite.NewHandoffRepository(db)
	ctx := context.Background()

	_, err := repo.GetLatest(ctx)
	if err == nil {
		t.Error("expected error when no handoffs exist")
	}
}

func TestHandoffRepository_GetLatestForWorkbench(t *testing.T) {
	db := setupHandoffTestDB(t)
	repo := sqlite.NewHandoffRepository(db)
	ctx := context.Background()

	// Create handoffs for different workbenches with explicit timestamps
	_, _ = db.Exec(`INSERT INTO handoffs (id, handoff_note, active_workbench_id, created_at) VALUES ('HO-001', 'Workbench 1 first', 'WB-001', '2024-01-01 10:00:00')`)
	_, _ = db.Exec(`INSERT INTO handoffs (id, handoff_note, active_workbench_id, created_at) VALUES ('HO-002', 'Workbench 1 latest', 'WB-001', '2024-01-01 11:00:00')`)
	_, _ = db.Exec(`INSERT INTO handoffs (id, handoff_note, active_workbench_id, created_at) VALUES ('HO-003', 'Workbench 2', 'WB-002', '2024-01-01 12:00:00')`)

	latest, err := repo.GetLatestForWorkbench(ctx, "WB-001")
	if err != nil {
		t.Fatalf("GetLatestForWorkbench failed: %v", err)
	}

	if latest.ID != "HO-002" {
		t.Errorf("expected handoff ID 'HO-002', got '%s'", latest.ID)
	}
	if latest.HandoffNote != "Workbench 1 latest" {
		t.Errorf("expected note 'Workbench 1 latest', got '%s'", latest.HandoffNote)
	}
}

func TestHandoffRepository_GetLatestForWorkbench_NotFound(t *testing.T) {
	db := setupHandoffTestDB(t)
	repo := sqlite.NewHandoffRepository(db)
	ctx := context.Background()

	createTestHandoff(t, repo, ctx, "Different workbench", "", "WB-002", "")

	_, err := repo.GetLatestForWorkbench(ctx, "WB-001")
	if err == nil {
		t.Error("expected error for workbench with no handoffs")
	}
}

func TestHandoffRepository_List(t *testing.T) {
	db := setupHandoffTestDB(t)
	repo := sqlite.NewHandoffRepository(db)
	ctx := context.Background()

	// Create handoffs with explicit timestamps for deterministic ordering
	_, _ = db.Exec(`INSERT INTO handoffs (id, handoff_note, created_at) VALUES ('HO-001', 'Handoff 1', '2024-01-01 10:00:00')`)
	_, _ = db.Exec(`INSERT INTO handoffs (id, handoff_note, created_at) VALUES ('HO-002', 'Handoff 2', '2024-01-01 11:00:00')`)
	_, _ = db.Exec(`INSERT INTO handoffs (id, handoff_note, created_at) VALUES ('HO-003', 'Handoff 3', '2024-01-01 12:00:00')`)

	// List all
	handoffs, err := repo.List(ctx, 0)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(handoffs) != 3 {
		t.Errorf("expected 3 handoffs, got %d", len(handoffs))
	}

	// Verify order (most recent first)
	if handoffs[0].HandoffNote != "Handoff 3" {
		t.Errorf("expected first handoff to be 'Handoff 3', got '%s'", handoffs[0].HandoffNote)
	}
}

func TestHandoffRepository_List_WithLimit(t *testing.T) {
	db := setupHandoffTestDB(t)
	repo := sqlite.NewHandoffRepository(db)
	ctx := context.Background()

	createTestHandoff(t, repo, ctx, "Handoff 1", "", "", "")
	createTestHandoff(t, repo, ctx, "Handoff 2", "", "", "")
	createTestHandoff(t, repo, ctx, "Handoff 3", "", "", "")
	createTestHandoff(t, repo, ctx, "Handoff 4", "", "", "")
	createTestHandoff(t, repo, ctx, "Handoff 5", "", "", "")

	// List with limit
	handoffs, err := repo.List(ctx, 2)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(handoffs) != 2 {
		t.Errorf("expected 2 handoffs, got %d", len(handoffs))
	}
}

func TestHandoffRepository_GetNextID(t *testing.T) {
	db := setupHandoffTestDB(t)
	repo := sqlite.NewHandoffRepository(db)
	ctx := context.Background()

	id, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "HO-001" {
		t.Errorf("expected HO-001, got %s", id)
	}

	createTestHandoff(t, repo, ctx, "Test", "", "", "")

	id, err = repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "HO-002" {
		t.Errorf("expected HO-002, got %s", id)
	}
}
