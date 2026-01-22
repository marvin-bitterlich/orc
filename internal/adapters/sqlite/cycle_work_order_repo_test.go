package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestCycleWorkOrderRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleWorkOrderRepository(db)
	ctx := context.Background()

	// Create test fixtures: commission -> shipment -> cycle
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test Shipment", "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-001", "SHIP-001", 1, "active")

	t.Run("creates cycle work order successfully", func(t *testing.T) {
		record := &secondary.CycleWorkOrderRecord{
			ID:         "CWO-001",
			CycleID:    "CYC-001",
			ShipmentID: "SHIP-001",
			Outcome:    "Implement feature X",
			Status:     "draft",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "CWO-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.Outcome != "Implement feature X" {
			t.Errorf("Outcome = %q, want %q", got.Outcome, "Implement feature X")
		}
		if got.Status != "draft" {
			t.Errorf("Status = %q, want %q", got.Status, "draft")
		}
	})

	t.Run("creates cycle work order with acceptance criteria", func(t *testing.T) {
		db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-002", "SHIP-001", 2, "queued")

		record := &secondary.CycleWorkOrderRecord{
			ID:                 "CWO-002",
			CycleID:            "CYC-002",
			ShipmentID:         "SHIP-001",
			Outcome:            "Implement feature Y",
			AcceptanceCriteria: `["Tests pass", "No lint errors"]`,
			Status:             "draft",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "CWO-002")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.AcceptanceCriteria != `["Tests pass", "No lint errors"]` {
			t.Errorf("AcceptanceCriteria = %q, want %q", got.AcceptanceCriteria, `["Tests pass", "No lint errors"]`)
		}
	})

	t.Run("enforces unique cycle constraint", func(t *testing.T) {
		record := &secondary.CycleWorkOrderRecord{
			ID:         "CWO-003",
			CycleID:    "CYC-001", // Same cycle as CWO-001
			ShipmentID: "SHIP-001",
			Outcome:    "Duplicate",
			Status:     "draft",
		}

		err := repo.Create(ctx, record)
		if err == nil {
			t.Fatal("Expected error for duplicate cycle, got nil")
		}
	})
}

func TestCycleWorkOrderRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleWorkOrderRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-001", "SHIP-001", 1, "active")

	repo.Create(ctx, &secondary.CycleWorkOrderRecord{
		ID:         "CWO-001",
		CycleID:    "CYC-001",
		ShipmentID: "SHIP-001",
		Outcome:    "Test outcome",
		Status:     "draft",
	})

	t.Run("finds cycle work order by ID", func(t *testing.T) {
		got, err := repo.GetByID(ctx, "CWO-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if got.ID != "CWO-001" {
			t.Errorf("ID = %q, want %q", got.ID, "CWO-001")
		}
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "CWO-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestCycleWorkOrderRepository_GetByCycle(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleWorkOrderRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-001", "SHIP-001", 1, "active")

	repo.Create(ctx, &secondary.CycleWorkOrderRecord{
		ID:         "CWO-001",
		CycleID:    "CYC-001",
		ShipmentID: "SHIP-001",
		Outcome:    "Test outcome",
		Status:     "draft",
	})

	t.Run("finds cycle work order by cycle", func(t *testing.T) {
		got, err := repo.GetByCycle(ctx, "CYC-001")
		if err != nil {
			t.Fatalf("GetByCycle failed: %v", err)
		}
		if got == nil {
			t.Fatal("expected cycle work order, got nil")
		}
		if got.ID != "CWO-001" {
			t.Errorf("ID = %q, want %q", got.ID, "CWO-001")
		}
	})

	t.Run("returns error for cycle without CWO", func(t *testing.T) {
		_, err := repo.GetByCycle(ctx, "CYC-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestCycleWorkOrderRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleWorkOrderRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test 1", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-002", "COMM-001", "Test 2", "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-001", "SHIP-001", 1, "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-002", "SHIP-002", 1, "active")

	repo.Create(ctx, &secondary.CycleWorkOrderRecord{ID: "CWO-001", CycleID: "CYC-001", ShipmentID: "SHIP-001", Outcome: "Test 1", Status: "draft"})
	repo.Create(ctx, &secondary.CycleWorkOrderRecord{ID: "CWO-002", CycleID: "CYC-002", ShipmentID: "SHIP-002", Outcome: "Test 2", Status: "active"})

	t.Run("lists all cycle work orders", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.CycleWorkOrderFilters{})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("len = %d, want 2", len(list))
		}
	})

	t.Run("filters by cycle_id", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.CycleWorkOrderFilters{CycleID: "CYC-001"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "CWO-001" {
			t.Errorf("ID = %q, want %q", list[0].ID, "CWO-001")
		}
	})

	t.Run("filters by shipment_id", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.CycleWorkOrderFilters{ShipmentID: "SHIP-002"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "CWO-002" {
			t.Errorf("ID = %q, want %q", list[0].ID, "CWO-002")
		}
	})

	t.Run("filters by status", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.CycleWorkOrderFilters{Status: "draft"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "CWO-001" {
			t.Errorf("ID = %q, want %q", list[0].ID, "CWO-001")
		}
	})
}

func TestCycleWorkOrderRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleWorkOrderRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-001", "SHIP-001", 1, "active")

	repo.Create(ctx, &secondary.CycleWorkOrderRecord{
		ID:         "CWO-001",
		CycleID:    "CYC-001",
		ShipmentID: "SHIP-001",
		Outcome:    "Original outcome",
		Status:     "draft",
	})

	t.Run("updates outcome", func(t *testing.T) {
		err := repo.Update(ctx, &secondary.CycleWorkOrderRecord{
			ID:      "CWO-001",
			Outcome: "Updated outcome",
		})
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "CWO-001")
		if got.Outcome != "Updated outcome" {
			t.Errorf("Outcome = %q, want %q", got.Outcome, "Updated outcome")
		}
	})

	t.Run("returns error for non-existent CWO", func(t *testing.T) {
		err := repo.Update(ctx, &secondary.CycleWorkOrderRecord{
			ID:      "CWO-999",
			Outcome: "Will fail",
		})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestCycleWorkOrderRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleWorkOrderRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-001", "SHIP-001", 1, "active")

	repo.Create(ctx, &secondary.CycleWorkOrderRecord{ID: "CWO-001", CycleID: "CYC-001", ShipmentID: "SHIP-001", Outcome: "Test", Status: "draft"})

	t.Run("deletes cycle work order", func(t *testing.T) {
		err := repo.Delete(ctx, "CWO-001")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, err = repo.GetByID(ctx, "CWO-001")
		if err == nil {
			t.Error("expected error after delete, got nil")
		}
	})

	t.Run("returns error for non-existent CWO", func(t *testing.T) {
		err := repo.Delete(ctx, "CWO-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestCycleWorkOrderRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleWorkOrderRepository(db)
	ctx := context.Background()

	t.Run("returns CWO-001 for empty table", func(t *testing.T) {
		id, err := repo.GetNextID(ctx)
		if err != nil {
			t.Fatalf("GetNextID failed: %v", err)
		}
		if id != "CWO-001" {
			t.Errorf("ID = %q, want %q", id, "CWO-001")
		}
	})
}

func TestCycleWorkOrderRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleWorkOrderRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-001", "SHIP-001", 1, "active")

	repo.Create(ctx, &secondary.CycleWorkOrderRecord{
		ID:         "CWO-001",
		CycleID:    "CYC-001",
		ShipmentID: "SHIP-001",
		Outcome:    "Test",
		Status:     "draft",
	})

	t.Run("updates to active", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "CWO-001", "active")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "CWO-001")
		if got.Status != "active" {
			t.Errorf("Status = %q, want %q", got.Status, "active")
		}
	})

	t.Run("updates to complete", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "CWO-001", "complete")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "CWO-001")
		if got.Status != "complete" {
			t.Errorf("Status = %q, want %q", got.Status, "complete")
		}
	})

	t.Run("returns error for non-existent CWO", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "CWO-999", "active")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestCycleWorkOrderRepository_CycleExists(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleWorkOrderRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-001", "SHIP-001", 1, "active")

	t.Run("returns true for existing cycle", func(t *testing.T) {
		exists, err := repo.CycleExists(ctx, "CYC-001")
		if err != nil {
			t.Fatalf("CycleExists failed: %v", err)
		}
		if !exists {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns false for non-existent cycle", func(t *testing.T) {
		exists, err := repo.CycleExists(ctx, "CYC-999")
		if err != nil {
			t.Fatalf("CycleExists failed: %v", err)
		}
		if exists {
			t.Error("expected false, got true")
		}
	})
}

func TestCycleWorkOrderRepository_ShipmentExists(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleWorkOrderRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")

	t.Run("returns true for existing shipment", func(t *testing.T) {
		exists, err := repo.ShipmentExists(ctx, "SHIP-001")
		if err != nil {
			t.Fatalf("ShipmentExists failed: %v", err)
		}
		if !exists {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns false for non-existent shipment", func(t *testing.T) {
		exists, err := repo.ShipmentExists(ctx, "SHIP-999")
		if err != nil {
			t.Fatalf("ShipmentExists failed: %v", err)
		}
		if exists {
			t.Error("expected false, got true")
		}
	})
}

func TestCycleWorkOrderRepository_CycleHasCWO(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleWorkOrderRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-001", "SHIP-001", 1, "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-002", "SHIP-001", 2, "queued")

	repo.Create(ctx, &secondary.CycleWorkOrderRecord{
		ID:         "CWO-001",
		CycleID:    "CYC-001",
		ShipmentID: "SHIP-001",
		Outcome:    "Test",
		Status:     "draft",
	})

	t.Run("returns true when cycle has CWO", func(t *testing.T) {
		has, err := repo.CycleHasCWO(ctx, "CYC-001")
		if err != nil {
			t.Fatalf("CycleHasCWO failed: %v", err)
		}
		if !has {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns false when cycle has no CWO", func(t *testing.T) {
		has, err := repo.CycleHasCWO(ctx, "CYC-002")
		if err != nil {
			t.Fatalf("CycleHasCWO failed: %v", err)
		}
		if has {
			t.Error("expected false, got true")
		}
	})
}

func TestCycleWorkOrderRepository_GetCycleStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleWorkOrderRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-001", "SHIP-001", 1, "active")

	t.Run("returns status for existing cycle", func(t *testing.T) {
		status, err := repo.GetCycleStatus(ctx, "CYC-001")
		if err != nil {
			t.Fatalf("GetCycleStatus failed: %v", err)
		}
		if status != "active" {
			t.Errorf("Status = %q, want %q", status, "active")
		}
	})

	t.Run("returns error for non-existent cycle", func(t *testing.T) {
		_, err := repo.GetCycleStatus(ctx, "CYC-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestCycleWorkOrderRepository_GetCycleShipmentID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleWorkOrderRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-001", "SHIP-001", 1, "active")

	t.Run("returns shipment ID for existing cycle", func(t *testing.T) {
		shipmentID, err := repo.GetCycleShipmentID(ctx, "CYC-001")
		if err != nil {
			t.Fatalf("GetCycleShipmentID failed: %v", err)
		}
		if shipmentID != "SHIP-001" {
			t.Errorf("ShipmentID = %q, want %q", shipmentID, "SHIP-001")
		}
	})

	t.Run("returns error for non-existent cycle", func(t *testing.T) {
		_, err := repo.GetCycleShipmentID(ctx, "CYC-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
