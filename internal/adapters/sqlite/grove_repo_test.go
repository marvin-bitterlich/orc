package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func setupGroveTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// Create missions table (required for foreign key)
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

	// Create groves table
	_, err = db.Exec(`
		CREATE TABLE groves (
			id TEXT PRIMARY KEY,
			mission_id TEXT NOT NULL,
			name TEXT NOT NULL,
			path TEXT,
			repos TEXT,
			status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create groves table: %v", err)
	}

	// Insert a test mission
	_, err = db.Exec("INSERT INTO missions (id, title) VALUES ('MISSION-001', 'Test Mission')")
	if err != nil {
		t.Fatalf("failed to insert test mission: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func TestGroveRepository_Create(t *testing.T) {
	db := setupGroveTestDB(t)
	repo := sqlite.NewGroveRepository(db)
	ctx := context.Background()

	grove := &secondary.GroveRecord{
		MissionID:    "MISSION-001",
		Name:         "test-grove",
		WorktreePath: "/path/to/worktree",
	}

	err := repo.Create(ctx, grove)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if grove.ID == "" {
		t.Error("expected grove ID to be set")
	}

	if grove.ID != "GROVE-001" {
		t.Errorf("expected ID GROVE-001, got %s", grove.ID)
	}
}

func TestGroveRepository_Create_MissionNotFound(t *testing.T) {
	db := setupGroveTestDB(t)
	repo := sqlite.NewGroveRepository(db)
	ctx := context.Background()

	grove := &secondary.GroveRecord{
		MissionID:    "MISSION-999",
		Name:         "test-grove",
		WorktreePath: "/path/to/worktree",
	}

	err := repo.Create(ctx, grove)
	if err == nil {
		t.Error("expected error for non-existent mission")
	}
}

func TestGroveRepository_GetByID(t *testing.T) {
	db := setupGroveTestDB(t)
	repo := sqlite.NewGroveRepository(db)
	ctx := context.Background()

	// Create a grove first
	grove := &secondary.GroveRecord{
		MissionID:    "MISSION-001",
		Name:         "test-grove",
		WorktreePath: "/path/to/worktree",
	}
	_ = repo.Create(ctx, grove)

	// Retrieve it
	retrieved, err := repo.GetByID(ctx, grove.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.Name != "test-grove" {
		t.Errorf("expected name 'test-grove', got '%s'", retrieved.Name)
	}

	if retrieved.WorktreePath != "/path/to/worktree" {
		t.Errorf("expected path '/path/to/worktree', got '%s'", retrieved.WorktreePath)
	}

	if retrieved.Status != "active" {
		t.Errorf("expected status 'active', got '%s'", retrieved.Status)
	}
}

func TestGroveRepository_GetByID_NotFound(t *testing.T) {
	db := setupGroveTestDB(t)
	repo := sqlite.NewGroveRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "GROVE-999")
	if err == nil {
		t.Error("expected error for non-existent grove")
	}
}

func TestGroveRepository_GetByPath(t *testing.T) {
	db := setupGroveTestDB(t)
	repo := sqlite.NewGroveRepository(db)
	ctx := context.Background()

	// Create a grove
	grove := &secondary.GroveRecord{
		MissionID:    "MISSION-001",
		Name:         "test-grove",
		WorktreePath: "/unique/path/here",
	}
	_ = repo.Create(ctx, grove)

	// Retrieve by path
	retrieved, err := repo.GetByPath(ctx, "/unique/path/here")
	if err != nil {
		t.Fatalf("GetByPath failed: %v", err)
	}

	if retrieved.ID != grove.ID {
		t.Errorf("expected ID %s, got %s", grove.ID, retrieved.ID)
	}
}

func TestGroveRepository_GetByMission(t *testing.T) {
	db := setupGroveTestDB(t)
	repo := sqlite.NewGroveRepository(db)
	ctx := context.Background()

	// Create multiple groves
	_ = repo.Create(ctx, &secondary.GroveRecord{MissionID: "MISSION-001", Name: "grove-1"})
	_ = repo.Create(ctx, &secondary.GroveRecord{MissionID: "MISSION-001", Name: "grove-2"})

	groves, err := repo.GetByMission(ctx, "MISSION-001")
	if err != nil {
		t.Fatalf("GetByMission failed: %v", err)
	}

	if len(groves) != 2 {
		t.Errorf("expected 2 groves, got %d", len(groves))
	}
}

func TestGroveRepository_List(t *testing.T) {
	db := setupGroveTestDB(t)
	repo := sqlite.NewGroveRepository(db)
	ctx := context.Background()

	// Create groves
	_ = repo.Create(ctx, &secondary.GroveRecord{MissionID: "MISSION-001", Name: "grove-1"})
	_ = repo.Create(ctx, &secondary.GroveRecord{MissionID: "MISSION-001", Name: "grove-2"})
	_ = repo.Create(ctx, &secondary.GroveRecord{MissionID: "MISSION-001", Name: "grove-3"})

	// List all
	groves, err := repo.List(ctx, "")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(groves) != 3 {
		t.Errorf("expected 3 groves, got %d", len(groves))
	}

	// List filtered by mission
	groves, err = repo.List(ctx, "MISSION-001")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(groves) != 3 {
		t.Errorf("expected 3 groves for mission, got %d", len(groves))
	}
}

func TestGroveRepository_Update(t *testing.T) {
	db := setupGroveTestDB(t)
	repo := sqlite.NewGroveRepository(db)
	ctx := context.Background()

	// Create a grove
	grove := &secondary.GroveRecord{
		MissionID:    "MISSION-001",
		Name:         "original-name",
		WorktreePath: "/original/path",
	}
	_ = repo.Create(ctx, grove)

	// Update it
	err := repo.Update(ctx, &secondary.GroveRecord{
		ID:           grove.ID,
		Name:         "updated-name",
		WorktreePath: "/updated/path",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	retrieved, _ := repo.GetByID(ctx, grove.ID)
	if retrieved.Name != "updated-name" {
		t.Errorf("expected name 'updated-name', got '%s'", retrieved.Name)
	}
	if retrieved.WorktreePath != "/updated/path" {
		t.Errorf("expected path '/updated/path', got '%s'", retrieved.WorktreePath)
	}
}

func TestGroveRepository_Delete(t *testing.T) {
	db := setupGroveTestDB(t)
	repo := sqlite.NewGroveRepository(db)
	ctx := context.Background()

	// Create a grove
	grove := &secondary.GroveRecord{MissionID: "MISSION-001", Name: "to-delete"}
	_ = repo.Create(ctx, grove)

	// Delete it
	err := repo.Delete(ctx, grove.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(ctx, grove.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestGroveRepository_Rename(t *testing.T) {
	db := setupGroveTestDB(t)
	repo := sqlite.NewGroveRepository(db)
	ctx := context.Background()

	// Create a grove
	grove := &secondary.GroveRecord{MissionID: "MISSION-001", Name: "original"}
	_ = repo.Create(ctx, grove)

	// Rename it
	err := repo.Rename(ctx, grove.ID, "renamed")
	if err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	// Verify rename
	retrieved, _ := repo.GetByID(ctx, grove.ID)
	if retrieved.Name != "renamed" {
		t.Errorf("expected name 'renamed', got '%s'", retrieved.Name)
	}
}

func TestGroveRepository_UpdatePath(t *testing.T) {
	db := setupGroveTestDB(t)
	repo := sqlite.NewGroveRepository(db)
	ctx := context.Background()

	// Create a grove
	grove := &secondary.GroveRecord{
		MissionID:    "MISSION-001",
		Name:         "test",
		WorktreePath: "/old/path",
	}
	_ = repo.Create(ctx, grove)

	// Update path
	err := repo.UpdatePath(ctx, grove.ID, "/new/path")
	if err != nil {
		t.Fatalf("UpdatePath failed: %v", err)
	}

	// Verify update
	retrieved, _ := repo.GetByID(ctx, grove.ID)
	if retrieved.WorktreePath != "/new/path" {
		t.Errorf("expected path '/new/path', got '%s'", retrieved.WorktreePath)
	}
}

func TestGroveRepository_GetNextID(t *testing.T) {
	db := setupGroveTestDB(t)
	repo := sqlite.NewGroveRepository(db)
	ctx := context.Background()

	// First ID should be GROVE-001
	id, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "GROVE-001" {
		t.Errorf("expected GROVE-001, got %s", id)
	}

	// Create a grove
	_ = repo.Create(ctx, &secondary.GroveRecord{MissionID: "MISSION-001", Name: "test"})

	// Next ID should be GROVE-002
	id, err = repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "GROVE-002" {
		t.Errorf("expected GROVE-002, got %s", id)
	}
}
