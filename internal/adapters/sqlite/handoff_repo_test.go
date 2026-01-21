package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func setupHandoffTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// Create handoffs table
	_, err = db.Exec(`
		CREATE TABLE handoffs (
			id TEXT PRIMARY KEY,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			handoff_note TEXT NOT NULL,
			active_mission_id TEXT,
			active_grove_id TEXT,
			todos_snapshot TEXT
		)
	`)
	if err != nil {
		t.Fatalf("failed to create handoffs table: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// createTestHandoff is a helper that creates a handoff with a generated ID.
func createTestHandoff(t *testing.T, repo *sqlite.HandoffRepository, ctx context.Context, note, missionID, groveID, todos string) *secondary.HandoffRecord {
	t.Helper()

	nextID, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}

	handoff := &secondary.HandoffRecord{
		ID:              nextID,
		HandoffNote:     note,
		ActiveMissionID: missionID,
		ActiveGroveID:   groveID,
		TodosSnapshot:   todos,
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
		ID:              "HO-001",
		HandoffNote:     "Session complete. Next steps: review plan.",
		ActiveMissionID: "MISSION-001",
		ActiveGroveID:   "GROVE-001",
		TodosSnapshot:   "[]",
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
	if retrieved.ActiveMissionID != "MISSION-001" {
		t.Errorf("expected mission 'MISSION-001', got '%s'", retrieved.ActiveMissionID)
	}
	if retrieved.ActiveGroveID != "GROVE-001" {
		t.Errorf("expected grove 'GROVE-001', got '%s'", retrieved.ActiveGroveID)
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
	if retrieved.ActiveMissionID != "" {
		t.Errorf("expected empty mission ID, got '%s'", retrieved.ActiveMissionID)
	}
	if retrieved.ActiveGroveID != "" {
		t.Errorf("expected empty grove ID, got '%s'", retrieved.ActiveGroveID)
	}
}

func TestHandoffRepository_GetByID(t *testing.T) {
	db := setupHandoffTestDB(t)
	repo := sqlite.NewHandoffRepository(db)
	ctx := context.Background()

	handoff := createTestHandoff(t, repo, ctx, "Test note", "MISSION-001", "GROVE-001", "[]")

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
	_, _ = db.Exec(`INSERT INTO handoffs (id, handoff_note, active_mission_id, created_at) VALUES ('HO-003', 'Latest handoff', 'MISSION-001', '2024-01-01 12:00:00')`)

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

func TestHandoffRepository_GetLatestForGrove(t *testing.T) {
	db := setupHandoffTestDB(t)
	repo := sqlite.NewHandoffRepository(db)
	ctx := context.Background()

	// Create handoffs for different groves with explicit timestamps
	_, _ = db.Exec(`INSERT INTO handoffs (id, handoff_note, active_grove_id, created_at) VALUES ('HO-001', 'Grove 1 first', 'GROVE-001', '2024-01-01 10:00:00')`)
	_, _ = db.Exec(`INSERT INTO handoffs (id, handoff_note, active_grove_id, created_at) VALUES ('HO-002', 'Grove 1 latest', 'GROVE-001', '2024-01-01 11:00:00')`)
	_, _ = db.Exec(`INSERT INTO handoffs (id, handoff_note, active_grove_id, created_at) VALUES ('HO-003', 'Grove 2', 'GROVE-002', '2024-01-01 12:00:00')`)

	latest, err := repo.GetLatestForGrove(ctx, "GROVE-001")
	if err != nil {
		t.Fatalf("GetLatestForGrove failed: %v", err)
	}

	if latest.ID != "HO-002" {
		t.Errorf("expected handoff ID 'HO-002', got '%s'", latest.ID)
	}
	if latest.HandoffNote != "Grove 1 latest" {
		t.Errorf("expected note 'Grove 1 latest', got '%s'", latest.HandoffNote)
	}
}

func TestHandoffRepository_GetLatestForGrove_NotFound(t *testing.T) {
	db := setupHandoffTestDB(t)
	repo := sqlite.NewHandoffRepository(db)
	ctx := context.Background()

	createTestHandoff(t, repo, ctx, "Different grove", "", "GROVE-002", "")

	_, err := repo.GetLatestForGrove(ctx, "GROVE-001")
	if err == nil {
		t.Error("expected error for grove with no handoffs")
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
