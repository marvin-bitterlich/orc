package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestShipyardRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewShipyardRepository(db)
	ctx := context.Background()

	// Seed required factory
	seedFactory(t, db, "FACT-001", "Test Factory")

	shipyard := &secondary.ShipyardRecord{
		FactoryID: "FACT-001",
	}

	err := repo.Create(ctx, shipyard)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify ID was generated
	if shipyard.ID == "" {
		t.Error("expected ID to be generated")
	}
	if shipyard.ID != "YARD-001" {
		t.Errorf("expected ID 'YARD-001', got %q", shipyard.ID)
	}

	// Verify round-trip
	got, err := repo.GetByID(ctx, shipyard.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.FactoryID != "FACT-001" {
		t.Errorf("expected factory_id 'FACT-001', got %q", got.FactoryID)
	}
}

func TestShipyardRepository_Create_FactoryNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewShipyardRepository(db)
	ctx := context.Background()

	shipyard := &secondary.ShipyardRecord{
		FactoryID: "FACT-999",
	}

	err := repo.Create(ctx, shipyard)
	if err == nil {
		t.Error("expected error for non-existent factory")
	}
}

func TestShipyardRepository_Create_DuplicateFactory(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewShipyardRepository(db)
	ctx := context.Background()

	// Seed required factory
	seedFactory(t, db, "FACT-001", "Test Factory")

	// Create first shipyard
	shipyard1 := &secondary.ShipyardRecord{FactoryID: "FACT-001"}
	err := repo.Create(ctx, shipyard1)
	if err != nil {
		t.Fatalf("First Create failed: %v", err)
	}

	// Try to create second shipyard for same factory
	shipyard2 := &secondary.ShipyardRecord{FactoryID: "FACT-001"}
	err = repo.Create(ctx, shipyard2)
	if err == nil {
		t.Error("expected error for duplicate factory")
	}
}

func TestShipyardRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewShipyardRepository(db)
	ctx := context.Background()

	// Seed factory and shipyard
	seedFactory(t, db, "FACT-001", "Test Factory")
	seedShipyard(t, db, "YARD-001", "FACT-001")

	got, err := repo.GetByID(ctx, "YARD-001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.FactoryID != "FACT-001" {
		t.Errorf("expected factory_id 'FACT-001', got %q", got.FactoryID)
	}
	if got.CreatedAt == "" {
		t.Error("expected CreatedAt to be set")
	}
	if got.UpdatedAt == "" {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestShipyardRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewShipyardRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "YARD-999")
	if err == nil {
		t.Error("expected error for non-existent shipyard")
	}
}

func TestShipyardRepository_GetByFactoryID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewShipyardRepository(db)
	ctx := context.Background()

	// Seed factory and shipyard
	seedFactory(t, db, "FACT-001", "Test Factory")
	seedShipyard(t, db, "YARD-001", "FACT-001")

	got, err := repo.GetByFactoryID(ctx, "FACT-001")
	if err != nil {
		t.Fatalf("GetByFactoryID failed: %v", err)
	}
	if got.ID != "YARD-001" {
		t.Errorf("expected ID 'YARD-001', got %q", got.ID)
	}
}

func TestShipyardRepository_GetByFactoryID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewShipyardRepository(db)
	ctx := context.Background()

	// Seed factory without shipyard
	seedFactory(t, db, "FACT-001", "Test Factory")

	_, err := repo.GetByFactoryID(ctx, "FACT-001")
	if err == nil {
		t.Error("expected error for factory without shipyard")
	}
}

func TestShipyardRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewShipyardRepository(db)
	ctx := context.Background()

	// First ID with empty table
	nextID, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if nextID != "YARD-001" {
		t.Errorf("expected 'YARD-001', got %q", nextID)
	}

	// Seed some shipyards
	seedFactory(t, db, "FACT-001", "Test Factory 1")
	seedFactory(t, db, "FACT-002", "Test Factory 2")
	seedShipyard(t, db, "YARD-001", "FACT-001")
	seedShipyard(t, db, "YARD-002", "FACT-002")

	// Next ID after existing
	nextID, err = repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if nextID != "YARD-003" {
		t.Errorf("expected 'YARD-003', got %q", nextID)
	}
}

func TestShipyardRepository_FactoryExists(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewShipyardRepository(db)
	ctx := context.Background()

	// Factory doesn't exist
	exists, err := repo.FactoryExists(ctx, "FACT-001")
	if err != nil {
		t.Fatalf("FactoryExists failed: %v", err)
	}
	if exists {
		t.Error("expected factory to not exist")
	}

	// Seed factory
	seedFactory(t, db, "FACT-001", "Test Factory")

	// Factory exists
	exists, err = repo.FactoryExists(ctx, "FACT-001")
	if err != nil {
		t.Fatalf("FactoryExists failed: %v", err)
	}
	if !exists {
		t.Error("expected factory to exist")
	}
}

// seedShipyard inserts a test shipyard.
func seedShipyard(t *testing.T, db *sql.DB, id, factoryID string) {
	t.Helper()
	_, err := db.Exec("INSERT INTO shipyards (id, factory_id) VALUES (?, ?)", id, factoryID)
	if err != nil {
		t.Fatalf("failed to seed shipyard: %v", err)
	}
}
