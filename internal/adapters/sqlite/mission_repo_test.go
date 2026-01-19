package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/example/orc/internal/adapters/sqlite"
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

func TestMissionRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewMissionRepository(db)
	ctx := context.Background()

	mission := &secondary.MissionRecord{
		Title:       "Test Mission",
		Description: "A test mission description",
	}

	err := repo.Create(ctx, mission)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if mission.ID == "" {
		t.Error("expected mission ID to be set")
	}

	if mission.ID != "MISSION-001" {
		t.Errorf("expected ID MISSION-001, got %s", mission.ID)
	}
}

func TestMissionRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewMissionRepository(db)
	ctx := context.Background()

	// Create a mission first
	mission := &secondary.MissionRecord{
		Title:       "Test Mission",
		Description: "Description",
	}
	_ = repo.Create(ctx, mission)

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

	// Create multiple missions
	_ = repo.Create(ctx, &secondary.MissionRecord{Title: "Mission 1"})
	_ = repo.Create(ctx, &secondary.MissionRecord{Title: "Mission 2"})
	_ = repo.Create(ctx, &secondary.MissionRecord{Title: "Mission 3"})

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

	// Create missions with different statuses
	m1 := &secondary.MissionRecord{Title: "Active Mission"}
	_ = repo.Create(ctx, m1)

	m2 := &secondary.MissionRecord{Title: "Complete Mission"}
	_ = repo.Create(ctx, m2)
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
	mission := &secondary.MissionRecord{Title: "Original Title"}
	_ = repo.Create(ctx, mission)

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

func TestMissionRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewMissionRepository(db)
	ctx := context.Background()

	// Create a mission
	mission := &secondary.MissionRecord{Title: "To Delete"}
	_ = repo.Create(ctx, mission)

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
	mission := &secondary.MissionRecord{Title: "Pin Test"}
	_ = repo.Create(ctx, mission)

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

	// Create a mission
	_ = repo.Create(ctx, &secondary.MissionRecord{Title: "Test"})

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
	mission := &secondary.MissionRecord{Title: "Test"}
	_ = repo.Create(ctx, mission)

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
