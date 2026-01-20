package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/example/orc/internal/adapters/sqlite"
	coremission "github.com/example/orc/internal/core/mission"
	"github.com/example/orc/internal/ports/secondary"
)

func setupTestDB(t *testing.T) *sql.DB {
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
			description TEXT,
			status TEXT NOT NULL DEFAULT 'active',
			pinned INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME
		)
	`)
	if err != nil {
		t.Fatalf("failed to create missions table: %v", err)
	}

	// Create shipments table for CountShipments test
	_, err = db.Exec(`
		CREATE TABLE shipments (
			id TEXT PRIMARY KEY,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active'
		)
	`)
	if err != nil {
		t.Fatalf("failed to create shipments table: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// createTestMission is a helper that simulates service-layer behavior:
// gets next ID, sets initial status, then creates.
func createTestMission(t *testing.T, repo *sqlite.MissionRepository, ctx context.Context, title, description string) *secondary.MissionRecord {
	t.Helper()

	// Service would call GetNextID first
	nextID, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}

	mission := &secondary.MissionRecord{
		ID:          nextID,
		Title:       title,
		Description: description,
		Status:      string(coremission.InitialStatus()),
	}

	err = repo.Create(ctx, mission)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	return mission
}

func TestMissionRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewMissionRepository(db)
	ctx := context.Background()

	// Pre-populate ID and Status as service layer would
	mission := &secondary.MissionRecord{
		ID:          "MISSION-001",
		Title:       "Test Mission",
		Description: "A test mission description",
		Status:      "active",
	}

	err := repo.Create(ctx, mission)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify mission was created
	retrieved, err := repo.GetByID(ctx, "MISSION-001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Title != "Test Mission" {
		t.Errorf("expected title 'Test Mission', got '%s'", retrieved.Title)
	}
}

func TestMissionRepository_Create_RequiresID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewMissionRepository(db)
	ctx := context.Background()

	// Missing ID should fail
	mission := &secondary.MissionRecord{
		Title:  "Test Mission",
		Status: "active",
	}

	err := repo.Create(ctx, mission)
	if err == nil {
		t.Error("expected error for missing ID")
	}
}

func TestMissionRepository_Create_RequiresStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewMissionRepository(db)
	ctx := context.Background()

	// Missing Status should fail
	mission := &secondary.MissionRecord{
		ID:    "MISSION-001",
		Title: "Test Mission",
	}

	err := repo.Create(ctx, mission)
	if err == nil {
		t.Error("expected error for missing Status")
	}
}

func TestMissionRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewMissionRepository(db)
	ctx := context.Background()

	// Create a mission using helper
	mission := createTestMission(t, repo, ctx, "Test Mission", "Description")

	// Retrieve it
	retrieved, err := repo.GetByID(ctx, mission.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.Title != "Test Mission" {
		t.Errorf("expected title 'Test Mission', got '%s'", retrieved.Title)
	}

	if retrieved.Description != "Description" {
		t.Errorf("expected description 'Description', got '%s'", retrieved.Description)
	}

	if retrieved.Status != "active" {
		t.Errorf("expected status 'active', got '%s'", retrieved.Status)
	}
}

func TestMissionRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewMissionRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "MISSION-999")
	if err == nil {
		t.Error("expected error for non-existent mission")
	}
}

func TestMissionRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewMissionRepository(db)
	ctx := context.Background()

	// Create multiple missions using helper
	createTestMission(t, repo, ctx, "Mission 1", "")
	createTestMission(t, repo, ctx, "Mission 2", "")
	createTestMission(t, repo, ctx, "Mission 3", "")

	missions, err := repo.List(ctx, secondary.MissionFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(missions) != 3 {
		t.Errorf("expected 3 missions, got %d", len(missions))
	}
}

func TestMissionRepository_List_FilterByStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewMissionRepository(db)
	ctx := context.Background()

	// Create active mission
	createTestMission(t, repo, ctx, "Active Mission", "")

	// Create and complete a mission
	m2 := createTestMission(t, repo, ctx, "Complete Mission", "")
	_ = repo.Update(ctx, &secondary.MissionRecord{ID: m2.ID, Status: "complete"})

	// List only active
	missions, err := repo.List(ctx, secondary.MissionFilters{Status: "active"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(missions) != 1 {
		t.Errorf("expected 1 active mission, got %d", len(missions))
	}
}

func TestMissionRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewMissionRepository(db)
	ctx := context.Background()

	// Create a mission
	mission := createTestMission(t, repo, ctx, "Original Title", "")

	// Update it
	err := repo.Update(ctx, &secondary.MissionRecord{
		ID:    mission.ID,
		Title: "Updated Title",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	retrieved, _ := repo.GetByID(ctx, mission.ID)
	if retrieved.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", retrieved.Title)
	}
}

func TestMissionRepository_Update_CompletedAt(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewMissionRepository(db)
	ctx := context.Background()

	// Create a mission
	mission := createTestMission(t, repo, ctx, "Test", "")

	// Update with CompletedAt (service layer would set this)
	err := repo.Update(ctx, &secondary.MissionRecord{
		ID:          mission.ID,
		Status:      "complete",
		CompletedAt: "2026-01-20T12:00:00Z",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify CompletedAt was set
	retrieved, _ := repo.GetByID(ctx, mission.ID)
	if retrieved.CompletedAt == "" {
		t.Error("expected CompletedAt to be set")
	}
	if retrieved.Status != "complete" {
		t.Errorf("expected status 'complete', got '%s'", retrieved.Status)
	}
}

func TestMissionRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewMissionRepository(db)
	ctx := context.Background()

	// Create a mission
	mission := createTestMission(t, repo, ctx, "To Delete", "")

	// Delete it
	err := repo.Delete(ctx, mission.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(ctx, mission.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestMissionRepository_Pin_Unpin(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewMissionRepository(db)
	ctx := context.Background()

	// Create a mission
	mission := createTestMission(t, repo, ctx, "Pin Test", "")

	// Pin it
	err := repo.Pin(ctx, mission.ID)
	if err != nil {
		t.Fatalf("Pin failed: %v", err)
	}

	// Verify pinned
	retrieved, _ := repo.GetByID(ctx, mission.ID)
	if !retrieved.Pinned {
		t.Error("expected mission to be pinned")
	}

	// Unpin it
	err = repo.Unpin(ctx, mission.ID)
	if err != nil {
		t.Fatalf("Unpin failed: %v", err)
	}

	// Verify unpinned
	retrieved, _ = repo.GetByID(ctx, mission.ID)
	if retrieved.Pinned {
		t.Error("expected mission to be unpinned")
	}
}

func TestMissionRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewMissionRepository(db)
	ctx := context.Background()

	// First ID should be MISSION-001
	id, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "MISSION-001" {
		t.Errorf("expected MISSION-001, got %s", id)
	}

	// Create a mission using that ID
	createTestMission(t, repo, ctx, "Test", "")

	// Next ID should be MISSION-002
	id, err = repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "MISSION-002" {
		t.Errorf("expected MISSION-002, got %s", id)
	}
}

func TestMissionRepository_CountShipments(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewMissionRepository(db)
	ctx := context.Background()

	// Create a mission
	mission := createTestMission(t, repo, ctx, "Test", "")

	// Count should be 0
	count, err := repo.CountShipments(ctx, mission.ID)
	if err != nil {
		t.Fatalf("CountShipments failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 shipments, got %d", count)
	}

	// Add shipments
	_, _ = db.Exec("INSERT INTO shipments (id, mission_id, title) VALUES (?, ?, ?)", "SHIP-001", mission.ID, "Ship 1")
	_, _ = db.Exec("INSERT INTO shipments (id, mission_id, title) VALUES (?, ?, ?)", "SHIP-002", mission.ID, "Ship 2")

	// Count should be 2
	count, err = repo.CountShipments(ctx, mission.ID)
	if err != nil {
		t.Fatalf("CountShipments failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 shipments, got %d", count)
	}
}
