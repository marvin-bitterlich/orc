package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestCycleReceiptRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleReceiptRepository(db)
	ctx := context.Background()

	// Create test fixtures: commission -> shipment -> cycle -> CWO
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test Shipment", "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-001", "SHIP-001", 1, "active")
	db.ExecContext(ctx, "INSERT INTO cycle_work_orders (id, cycle_id, shipment_id, outcome, status) VALUES (?, ?, ?, ?, ?)", "CWO-001", "CYC-001", "SHIP-001", "Test outcome", "complete")

	t.Run("creates cycle receipt successfully", func(t *testing.T) {
		record := &secondary.CycleReceiptRecord{
			ID:               "CREC-001",
			CWOID:            "CWO-001",
			ShipmentID:       "SHIP-001",
			DeliveredOutcome: "Delivered feature X",
			Status:           "submitted",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "CREC-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.DeliveredOutcome != "Delivered feature X" {
			t.Errorf("DeliveredOutcome = %q, want %q", got.DeliveredOutcome, "Delivered feature X")
		}
		if got.Status != "submitted" {
			t.Errorf("Status = %q, want %q", got.Status, "submitted")
		}
	})

	t.Run("creates cycle receipt with evidence", func(t *testing.T) {
		db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-002", "SHIP-001", 2, "active")
		db.ExecContext(ctx, "INSERT INTO cycle_work_orders (id, cycle_id, shipment_id, outcome, status) VALUES (?, ?, ?, ?, ?)", "CWO-002", "CYC-002", "SHIP-001", "Test outcome 2", "complete")

		record := &secondary.CycleReceiptRecord{
			ID:                "CREC-002",
			CWOID:             "CWO-002",
			ShipmentID:        "SHIP-001",
			DeliveredOutcome:  "Delivered feature Y",
			Evidence:          "commit: abc123",
			VerificationNotes: "All tests pass",
			Status:            "submitted",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "CREC-002")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.Evidence != "commit: abc123" {
			t.Errorf("Evidence = %q, want %q", got.Evidence, "commit: abc123")
		}
		if got.VerificationNotes != "All tests pass" {
			t.Errorf("VerificationNotes = %q, want %q", got.VerificationNotes, "All tests pass")
		}
	})

	t.Run("enforces unique CWO constraint", func(t *testing.T) {
		record := &secondary.CycleReceiptRecord{
			ID:               "CREC-003",
			CWOID:            "CWO-001", // Same CWO as CREC-001
			ShipmentID:       "SHIP-001",
			DeliveredOutcome: "Duplicate",
			Status:           "submitted",
		}

		err := repo.Create(ctx, record)
		if err == nil {
			t.Fatal("Expected error for duplicate CWO, got nil")
		}
	})
}

func TestCycleReceiptRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleReceiptRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-001", "SHIP-001", 1, "active")
	db.ExecContext(ctx, "INSERT INTO cycle_work_orders (id, cycle_id, shipment_id, outcome, status) VALUES (?, ?, ?, ?, ?)", "CWO-001", "CYC-001", "SHIP-001", "Test", "complete")

	repo.Create(ctx, &secondary.CycleReceiptRecord{
		ID:               "CREC-001",
		CWOID:            "CWO-001",
		ShipmentID:       "SHIP-001",
		DeliveredOutcome: "Test outcome",
		Status:           "submitted",
	})

	t.Run("finds cycle receipt by ID", func(t *testing.T) {
		got, err := repo.GetByID(ctx, "CREC-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if got.ID != "CREC-001" {
			t.Errorf("ID = %q, want %q", got.ID, "CREC-001")
		}
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "CREC-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestCycleReceiptRepository_GetByCWO(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleReceiptRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-001", "SHIP-001", 1, "active")
	db.ExecContext(ctx, "INSERT INTO cycle_work_orders (id, cycle_id, shipment_id, outcome, status) VALUES (?, ?, ?, ?, ?)", "CWO-001", "CYC-001", "SHIP-001", "Test", "complete")

	repo.Create(ctx, &secondary.CycleReceiptRecord{
		ID:               "CREC-001",
		CWOID:            "CWO-001",
		ShipmentID:       "SHIP-001",
		DeliveredOutcome: "Test outcome",
		Status:           "submitted",
	})

	t.Run("finds cycle receipt by CWO", func(t *testing.T) {
		got, err := repo.GetByCWO(ctx, "CWO-001")
		if err != nil {
			t.Fatalf("GetByCWO failed: %v", err)
		}
		if got == nil {
			t.Fatal("expected cycle receipt, got nil")
		}
		if got.ID != "CREC-001" {
			t.Errorf("ID = %q, want %q", got.ID, "CREC-001")
		}
	})

	t.Run("returns error for CWO without CREC", func(t *testing.T) {
		_, err := repo.GetByCWO(ctx, "CWO-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestCycleReceiptRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleReceiptRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test 1", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-002", "COMM-001", "Test 2", "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-001", "SHIP-001", 1, "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-002", "SHIP-002", 1, "active")
	db.ExecContext(ctx, "INSERT INTO cycle_work_orders (id, cycle_id, shipment_id, outcome, status) VALUES (?, ?, ?, ?, ?)", "CWO-001", "CYC-001", "SHIP-001", "Test 1", "complete")
	db.ExecContext(ctx, "INSERT INTO cycle_work_orders (id, cycle_id, shipment_id, outcome, status) VALUES (?, ?, ?, ?, ?)", "CWO-002", "CYC-002", "SHIP-002", "Test 2", "complete")

	repo.Create(ctx, &secondary.CycleReceiptRecord{ID: "CREC-001", CWOID: "CWO-001", ShipmentID: "SHIP-001", DeliveredOutcome: "Test 1", Status: "submitted"})
	repo.Create(ctx, &secondary.CycleReceiptRecord{ID: "CREC-002", CWOID: "CWO-002", ShipmentID: "SHIP-002", DeliveredOutcome: "Test 2", Status: "verified"})

	t.Run("lists all cycle receipts", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.CycleReceiptFilters{})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("len = %d, want 2", len(list))
		}
	})

	t.Run("filters by cwo_id", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.CycleReceiptFilters{CWOID: "CWO-001"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "CREC-001" {
			t.Errorf("ID = %q, want %q", list[0].ID, "CREC-001")
		}
	})

	t.Run("filters by shipment_id", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.CycleReceiptFilters{ShipmentID: "SHIP-002"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "CREC-002" {
			t.Errorf("ID = %q, want %q", list[0].ID, "CREC-002")
		}
	})

	t.Run("filters by status", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.CycleReceiptFilters{Status: "verified"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "CREC-002" {
			t.Errorf("ID = %q, want %q", list[0].ID, "CREC-002")
		}
	})
}

func TestCycleReceiptRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleReceiptRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-001", "SHIP-001", 1, "active")
	db.ExecContext(ctx, "INSERT INTO cycle_work_orders (id, cycle_id, shipment_id, outcome, status) VALUES (?, ?, ?, ?, ?)", "CWO-001", "CYC-001", "SHIP-001", "Test", "complete")

	repo.Create(ctx, &secondary.CycleReceiptRecord{
		ID:               "CREC-001",
		CWOID:            "CWO-001",
		ShipmentID:       "SHIP-001",
		DeliveredOutcome: "Original outcome",
		Status:           "submitted",
	})

	t.Run("updates delivered outcome", func(t *testing.T) {
		err := repo.Update(ctx, &secondary.CycleReceiptRecord{
			ID:               "CREC-001",
			DeliveredOutcome: "Updated outcome",
		})
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "CREC-001")
		if got.DeliveredOutcome != "Updated outcome" {
			t.Errorf("DeliveredOutcome = %q, want %q", got.DeliveredOutcome, "Updated outcome")
		}
	})

	t.Run("returns error for non-existent CREC", func(t *testing.T) {
		err := repo.Update(ctx, &secondary.CycleReceiptRecord{
			ID:               "CREC-999",
			DeliveredOutcome: "Will fail",
		})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestCycleReceiptRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleReceiptRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-001", "SHIP-001", 1, "active")
	db.ExecContext(ctx, "INSERT INTO cycle_work_orders (id, cycle_id, shipment_id, outcome, status) VALUES (?, ?, ?, ?, ?)", "CWO-001", "CYC-001", "SHIP-001", "Test", "complete")

	repo.Create(ctx, &secondary.CycleReceiptRecord{ID: "CREC-001", CWOID: "CWO-001", ShipmentID: "SHIP-001", DeliveredOutcome: "Test", Status: "submitted"})

	t.Run("deletes cycle receipt", func(t *testing.T) {
		err := repo.Delete(ctx, "CREC-001")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, err = repo.GetByID(ctx, "CREC-001")
		if err == nil {
			t.Error("expected error after delete, got nil")
		}
	})

	t.Run("returns error for non-existent CREC", func(t *testing.T) {
		err := repo.Delete(ctx, "CREC-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestCycleReceiptRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleReceiptRepository(db)
	ctx := context.Background()

	t.Run("returns CREC-001 for empty table", func(t *testing.T) {
		id, err := repo.GetNextID(ctx)
		if err != nil {
			t.Fatalf("GetNextID failed: %v", err)
		}
		if id != "CREC-001" {
			t.Errorf("ID = %q, want %q", id, "CREC-001")
		}
	})
}

func TestCycleReceiptRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleReceiptRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-001", "SHIP-001", 1, "active")
	db.ExecContext(ctx, "INSERT INTO cycle_work_orders (id, cycle_id, shipment_id, outcome, status) VALUES (?, ?, ?, ?, ?)", "CWO-001", "CYC-001", "SHIP-001", "Test", "complete")

	repo.Create(ctx, &secondary.CycleReceiptRecord{
		ID:               "CREC-001",
		CWOID:            "CWO-001",
		ShipmentID:       "SHIP-001",
		DeliveredOutcome: "Test",
		Status:           "submitted",
	})

	t.Run("updates to verified", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "CREC-001", "verified")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "CREC-001")
		if got.Status != "verified" {
			t.Errorf("Status = %q, want %q", got.Status, "verified")
		}
	})

	t.Run("updates to submitted", func(t *testing.T) {
		// First reset to draft
		repo.UpdateStatus(ctx, "CREC-001", "draft")

		err := repo.UpdateStatus(ctx, "CREC-001", "submitted")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "CREC-001")
		if got.Status != "submitted" {
			t.Errorf("Status = %q, want %q", got.Status, "submitted")
		}
	})

	t.Run("returns error for non-existent CREC", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "CREC-999", "verified")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestCycleReceiptRepository_CWOExists(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleReceiptRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-001", "SHIP-001", 1, "active")
	db.ExecContext(ctx, "INSERT INTO cycle_work_orders (id, cycle_id, shipment_id, outcome, status) VALUES (?, ?, ?, ?, ?)", "CWO-001", "CYC-001", "SHIP-001", "Test", "complete")

	t.Run("returns true for existing CWO", func(t *testing.T) {
		exists, err := repo.CWOExists(ctx, "CWO-001")
		if err != nil {
			t.Fatalf("CWOExists failed: %v", err)
		}
		if !exists {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns false for non-existent CWO", func(t *testing.T) {
		exists, err := repo.CWOExists(ctx, "CWO-999")
		if err != nil {
			t.Fatalf("CWOExists failed: %v", err)
		}
		if exists {
			t.Error("expected false, got true")
		}
	})
}

func TestCycleReceiptRepository_CWOHasCREC(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleReceiptRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-001", "SHIP-001", 1, "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-002", "SHIP-001", 2, "queued")
	db.ExecContext(ctx, "INSERT INTO cycle_work_orders (id, cycle_id, shipment_id, outcome, status) VALUES (?, ?, ?, ?, ?)", "CWO-001", "CYC-001", "SHIP-001", "Test 1", "complete")
	db.ExecContext(ctx, "INSERT INTO cycle_work_orders (id, cycle_id, shipment_id, outcome, status) VALUES (?, ?, ?, ?, ?)", "CWO-002", "CYC-002", "SHIP-001", "Test 2", "active")

	repo.Create(ctx, &secondary.CycleReceiptRecord{
		ID:               "CREC-001",
		CWOID:            "CWO-001",
		ShipmentID:       "SHIP-001",
		DeliveredOutcome: "Test",
		Status:           "submitted",
	})

	t.Run("returns true when CWO has CREC", func(t *testing.T) {
		has, err := repo.CWOHasCREC(ctx, "CWO-001")
		if err != nil {
			t.Fatalf("CWOHasCREC failed: %v", err)
		}
		if !has {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns false when CWO has no CREC", func(t *testing.T) {
		has, err := repo.CWOHasCREC(ctx, "CWO-002")
		if err != nil {
			t.Fatalf("CWOHasCREC failed: %v", err)
		}
		if has {
			t.Error("expected false, got true")
		}
	})
}

func TestCycleReceiptRepository_GetCWOStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleReceiptRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-001", "SHIP-001", 1, "active")
	db.ExecContext(ctx, "INSERT INTO cycle_work_orders (id, cycle_id, shipment_id, outcome, status) VALUES (?, ?, ?, ?, ?)", "CWO-001", "CYC-001", "SHIP-001", "Test", "complete")

	t.Run("returns status for existing CWO", func(t *testing.T) {
		status, err := repo.GetCWOStatus(ctx, "CWO-001")
		if err != nil {
			t.Fatalf("GetCWOStatus failed: %v", err)
		}
		if status != "complete" {
			t.Errorf("Status = %q, want %q", status, "complete")
		}
	})

	t.Run("returns error for non-existent CWO", func(t *testing.T) {
		_, err := repo.GetCWOStatus(ctx, "CWO-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestCycleReceiptRepository_GetCWOShipmentID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCycleReceiptRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO cycles (id, shipment_id, sequence_number, status) VALUES (?, ?, ?, ?)", "CYC-001", "SHIP-001", 1, "active")
	db.ExecContext(ctx, "INSERT INTO cycle_work_orders (id, cycle_id, shipment_id, outcome, status) VALUES (?, ?, ?, ?, ?)", "CWO-001", "CYC-001", "SHIP-001", "Test", "complete")

	t.Run("returns shipment ID for existing CWO", func(t *testing.T) {
		shipmentID, err := repo.GetCWOShipmentID(ctx, "CWO-001")
		if err != nil {
			t.Fatalf("GetCWOShipmentID failed: %v", err)
		}
		if shipmentID != "SHIP-001" {
			t.Errorf("ShipmentID = %q, want %q", shipmentID, "SHIP-001")
		}
	})

	t.Run("returns error for non-existent CWO", func(t *testing.T) {
		_, err := repo.GetCWOShipmentID(ctx, "CWO-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
