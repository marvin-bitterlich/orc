package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func setupTomeTestDB(t *testing.T) *sql.DB {
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

	// Create groves table
	_, err = db.Exec(`
		CREATE TABLE groves (
			id TEXT PRIMARY KEY,
			mission_id TEXT NOT NULL,
			name TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create groves table: %v", err)
	}

	// Create tomes table
	_, err = db.Exec(`
		CREATE TABLE tomes (
			id TEXT PRIMARY KEY,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL DEFAULT 'active',
			assigned_grove_id TEXT,
			pinned INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME
		)
	`)
	if err != nil {
		t.Fatalf("failed to create tomes table: %v", err)
	}

	// Insert test data
	_, _ = db.Exec("INSERT INTO missions (id, title, status) VALUES ('MISSION-001', 'Test Mission', 'active')")
	_, _ = db.Exec("INSERT INTO groves (id, mission_id, name, status) VALUES ('GROVE-001', 'MISSION-001', 'test-grove', 'active')")

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// createTestTome is a helper that creates a tome with a generated ID.
func createTestTome(t *testing.T, repo *sqlite.TomeRepository, ctx context.Context, missionID, title, description string) *secondary.TomeRecord {
	t.Helper()

	nextID, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}

	tome := &secondary.TomeRecord{
		ID:          nextID,
		MissionID:   missionID,
		Title:       title,
		Description: description,
	}

	err = repo.Create(ctx, tome)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	return tome
}

func TestTomeRepository_Create(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db)
	ctx := context.Background()

	tome := &secondary.TomeRecord{
		ID:          "TOME-001",
		MissionID:   "MISSION-001",
		Title:       "Test Tome",
		Description: "A test tome description",
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
	if retrieved.Status != "active" {
		t.Errorf("expected status 'active', got '%s'", retrieved.Status)
	}
}

func TestTomeRepository_GetByID(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db)
	ctx := context.Background()

	tome := createTestTome(t, repo, ctx, "MISSION-001", "Test Tome", "Description")

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
	repo := sqlite.NewTomeRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "TOME-999")
	if err == nil {
		t.Error("expected error for non-existent tome")
	}
}

func TestTomeRepository_List(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db)
	ctx := context.Background()

	createTestTome(t, repo, ctx, "MISSION-001", "Tome 1", "")
	createTestTome(t, repo, ctx, "MISSION-001", "Tome 2", "")
	createTestTome(t, repo, ctx, "MISSION-001", "Tome 3", "")

	tomes, err := repo.List(ctx, secondary.TomeFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(tomes) != 3 {
		t.Errorf("expected 3 tomes, got %d", len(tomes))
	}
}

func TestTomeRepository_List_FilterByMission(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db)
	ctx := context.Background()

	// Add another mission
	_, _ = db.Exec("INSERT INTO missions (id, title, status) VALUES ('MISSION-002', 'Mission 2', 'active')")

	createTestTome(t, repo, ctx, "MISSION-001", "Tome 1", "")
	createTestTome(t, repo, ctx, "MISSION-002", "Tome 2", "")

	tomes, err := repo.List(ctx, secondary.TomeFilters{MissionID: "MISSION-001"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(tomes) != 1 {
		t.Errorf("expected 1 tome for MISSION-001, got %d", len(tomes))
	}
}

func TestTomeRepository_List_FilterByStatus(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db)
	ctx := context.Background()

	t1 := createTestTome(t, repo, ctx, "MISSION-001", "Active Tome", "")
	createTestTome(t, repo, ctx, "MISSION-001", "Another Active", "")

	// Complete t1
	_ = repo.UpdateStatus(ctx, t1.ID, "complete", true)

	tomes, err := repo.List(ctx, secondary.TomeFilters{Status: "active"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(tomes) != 1 {
		t.Errorf("expected 1 active tome, got %d", len(tomes))
	}
}

func TestTomeRepository_Update(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db)
	ctx := context.Background()

	tome := createTestTome(t, repo, ctx, "MISSION-001", "Original Title", "")

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
	repo := sqlite.NewTomeRepository(db)
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
	repo := sqlite.NewTomeRepository(db)
	ctx := context.Background()

	tome := createTestTome(t, repo, ctx, "MISSION-001", "To Delete", "")

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
	repo := sqlite.NewTomeRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, "TOME-999")
	if err == nil {
		t.Error("expected error for non-existent tome")
	}
}

func TestTomeRepository_Pin_Unpin(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db)
	ctx := context.Background()

	tome := createTestTome(t, repo, ctx, "MISSION-001", "Pin Test", "")

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
	repo := sqlite.NewTomeRepository(db)
	ctx := context.Background()

	err := repo.Pin(ctx, "TOME-999")
	if err == nil {
		t.Error("expected error for non-existent tome")
	}
}

func TestTomeRepository_GetNextID(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db)
	ctx := context.Background()

	id, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "TOME-001" {
		t.Errorf("expected TOME-001, got %s", id)
	}

	createTestTome(t, repo, ctx, "MISSION-001", "Test", "")

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
	repo := sqlite.NewTomeRepository(db)
	ctx := context.Background()

	tome := createTestTome(t, repo, ctx, "MISSION-001", "Status Test", "")

	// Update status without completed timestamp
	err := repo.UpdateStatus(ctx, tome.ID, "in_progress", false)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, tome.ID)
	if retrieved.Status != "in_progress" {
		t.Errorf("expected status 'in_progress', got '%s'", retrieved.Status)
	}
	if retrieved.CompletedAt != "" {
		t.Error("expected CompletedAt to be empty")
	}

	// Update to complete
	err = repo.UpdateStatus(ctx, tome.ID, "complete", true)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	retrieved, _ = repo.GetByID(ctx, tome.ID)
	if retrieved.Status != "complete" {
		t.Errorf("expected status 'complete', got '%s'", retrieved.Status)
	}
	if retrieved.CompletedAt == "" {
		t.Error("expected CompletedAt to be set")
	}
}

func TestTomeRepository_UpdateStatus_NotFound(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db)
	ctx := context.Background()

	err := repo.UpdateStatus(ctx, "TOME-999", "complete", true)
	if err == nil {
		t.Error("expected error for non-existent tome")
	}
}

func TestTomeRepository_AssignGrove(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db)
	ctx := context.Background()

	tome := createTestTome(t, repo, ctx, "MISSION-001", "Grove Test", "")

	err := repo.AssignGrove(ctx, tome.ID, "GROVE-001")
	if err != nil {
		t.Fatalf("AssignGrove failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, tome.ID)
	if retrieved.AssignedGroveID != "GROVE-001" {
		t.Errorf("expected assigned grove 'GROVE-001', got '%s'", retrieved.AssignedGroveID)
	}
}

func TestTomeRepository_AssignGrove_NotFound(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db)
	ctx := context.Background()

	err := repo.AssignGrove(ctx, "TOME-999", "GROVE-001")
	if err == nil {
		t.Error("expected error for non-existent tome")
	}
}

func TestTomeRepository_GetByGrove(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db)
	ctx := context.Background()

	t1 := createTestTome(t, repo, ctx, "MISSION-001", "Tome 1", "")
	t2 := createTestTome(t, repo, ctx, "MISSION-001", "Tome 2", "")
	createTestTome(t, repo, ctx, "MISSION-001", "Tome 3 (unassigned)", "")

	_ = repo.AssignGrove(ctx, t1.ID, "GROVE-001")
	_ = repo.AssignGrove(ctx, t2.ID, "GROVE-001")

	tomes, err := repo.GetByGrove(ctx, "GROVE-001")
	if err != nil {
		t.Fatalf("GetByGrove failed: %v", err)
	}

	if len(tomes) != 2 {
		t.Errorf("expected 2 tomes for grove, got %d", len(tomes))
	}
}

func TestTomeRepository_MissionExists(t *testing.T) {
	db := setupTomeTestDB(t)
	repo := sqlite.NewTomeRepository(db)
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
