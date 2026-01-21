package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func setupTagTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// Create tags table
	_, err = db.Exec(`
		CREATE TABLE tags (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create tags table: %v", err)
	}

	// Create entity_tags table
	_, err = db.Exec(`
		CREATE TABLE entity_tags (
			id TEXT PRIMARY KEY,
			entity_id TEXT NOT NULL,
			entity_type TEXT NOT NULL,
			tag_id TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create entity_tags table: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// createTestTag is a helper that creates a tag with a generated ID.
func createTestTag(t *testing.T, repo *sqlite.TagRepository, ctx context.Context, name, description string) *secondary.TagRecord {
	t.Helper()

	nextID, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}

	tag := &secondary.TagRecord{
		ID:          nextID,
		Name:        name,
		Description: description,
	}

	err = repo.Create(ctx, tag)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	return tag
}

func TestTagRepository_Create(t *testing.T) {
	db := setupTagTestDB(t)
	repo := sqlite.NewTagRepository(db)
	ctx := context.Background()

	tag := &secondary.TagRecord{
		ID:          "TAG-001",
		Name:        "urgent",
		Description: "High priority items",
	}

	err := repo.Create(ctx, tag)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify tag was created
	retrieved, err := repo.GetByID(ctx, "TAG-001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Name != "urgent" {
		t.Errorf("expected name 'urgent', got '%s'", retrieved.Name)
	}
	if retrieved.Description != "High priority items" {
		t.Errorf("expected description 'High priority items', got '%s'", retrieved.Description)
	}
}

func TestTagRepository_Create_DuplicateName(t *testing.T) {
	db := setupTagTestDB(t)
	repo := sqlite.NewTagRepository(db)
	ctx := context.Background()

	createTestTag(t, repo, ctx, "urgent", "")

	// Try to create another with same name
	tag := &secondary.TagRecord{
		ID:   "TAG-002",
		Name: "urgent",
	}

	err := repo.Create(ctx, tag)
	if err == nil {
		t.Error("expected error for duplicate name")
	}
}

func TestTagRepository_GetByID(t *testing.T) {
	db := setupTagTestDB(t)
	repo := sqlite.NewTagRepository(db)
	ctx := context.Background()

	tag := createTestTag(t, repo, ctx, "feature", "Feature requests")

	retrieved, err := repo.GetByID(ctx, tag.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.Name != "feature" {
		t.Errorf("expected name 'feature', got '%s'", retrieved.Name)
	}
}

func TestTagRepository_GetByID_NotFound(t *testing.T) {
	db := setupTagTestDB(t)
	repo := sqlite.NewTagRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "TAG-999")
	if err == nil {
		t.Error("expected error for non-existent tag")
	}
}

func TestTagRepository_GetByName(t *testing.T) {
	db := setupTagTestDB(t)
	repo := sqlite.NewTagRepository(db)
	ctx := context.Background()

	tag := createTestTag(t, repo, ctx, "bug", "Bug fixes")

	retrieved, err := repo.GetByName(ctx, "bug")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}

	if retrieved.ID != tag.ID {
		t.Errorf("expected ID '%s', got '%s'", tag.ID, retrieved.ID)
	}
}

func TestTagRepository_GetByName_NotFound(t *testing.T) {
	db := setupTagTestDB(t)
	repo := sqlite.NewTagRepository(db)
	ctx := context.Background()

	_, err := repo.GetByName(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent tag name")
	}
}

func TestTagRepository_List(t *testing.T) {
	db := setupTagTestDB(t)
	repo := sqlite.NewTagRepository(db)
	ctx := context.Background()

	// Create tags - should be sorted by name
	createTestTag(t, repo, ctx, "zeta", "")
	createTestTag(t, repo, ctx, "alpha", "")
	createTestTag(t, repo, ctx, "beta", "")

	tags, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(tags))
	}

	// Verify sorting by name
	if tags[0].Name != "alpha" {
		t.Errorf("expected first tag 'alpha', got '%s'", tags[0].Name)
	}
	if tags[1].Name != "beta" {
		t.Errorf("expected second tag 'beta', got '%s'", tags[1].Name)
	}
	if tags[2].Name != "zeta" {
		t.Errorf("expected third tag 'zeta', got '%s'", tags[2].Name)
	}
}

func TestTagRepository_Delete(t *testing.T) {
	db := setupTagTestDB(t)
	repo := sqlite.NewTagRepository(db)
	ctx := context.Background()

	tag := createTestTag(t, repo, ctx, "to-delete", "")

	err := repo.Delete(ctx, tag.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.GetByID(ctx, tag.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestTagRepository_Delete_NotFound(t *testing.T) {
	db := setupTagTestDB(t)
	repo := sqlite.NewTagRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, "TAG-999")
	if err == nil {
		t.Error("expected error for non-existent tag")
	}
}

func TestTagRepository_GetNextID(t *testing.T) {
	db := setupTagTestDB(t)
	repo := sqlite.NewTagRepository(db)
	ctx := context.Background()

	id, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "TAG-001" {
		t.Errorf("expected TAG-001, got %s", id)
	}

	createTestTag(t, repo, ctx, "test", "")

	id, err = repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "TAG-002" {
		t.Errorf("expected TAG-002, got %s", id)
	}
}

func TestTagRepository_GetEntityTag(t *testing.T) {
	db := setupTagTestDB(t)
	repo := sqlite.NewTagRepository(db)
	ctx := context.Background()

	tag := createTestTag(t, repo, ctx, "urgent", "")

	// Initially no entity tag
	result, err := repo.GetEntityTag(ctx, "TASK-001", "task")
	if err != nil {
		t.Fatalf("GetEntityTag failed: %v", err)
	}
	if result != nil {
		t.Error("expected no entity tag initially")
	}

	// Add entity tag
	_, _ = db.Exec("INSERT INTO entity_tags (id, entity_id, entity_type, tag_id) VALUES ('ET-001', 'TASK-001', 'task', ?)", tag.ID)

	// Get entity tag
	result, err = repo.GetEntityTag(ctx, "TASK-001", "task")
	if err != nil {
		t.Fatalf("GetEntityTag failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected entity tag to be returned")
	}
	if result.ID != tag.ID {
		t.Errorf("expected tag ID '%s', got '%s'", tag.ID, result.ID)
	}
}

func TestTagRepository_GetEntityTag_WrongType(t *testing.T) {
	db := setupTagTestDB(t)
	repo := sqlite.NewTagRepository(db)
	ctx := context.Background()

	tag := createTestTag(t, repo, ctx, "urgent", "")

	// Add entity tag for task
	_, _ = db.Exec("INSERT INTO entity_tags (id, entity_id, entity_type, tag_id) VALUES ('ET-001', 'TASK-001', 'task', ?)", tag.ID)

	// Get entity tag for different type
	result, err := repo.GetEntityTag(ctx, "TASK-001", "shipment")
	if err != nil {
		t.Fatalf("GetEntityTag failed: %v", err)
	}
	if result != nil {
		t.Error("expected no entity tag for wrong type")
	}
}
