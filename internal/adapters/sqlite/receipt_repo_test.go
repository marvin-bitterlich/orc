package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestReceiptRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewReceiptRepository(db)
	ctx := context.Background()

	// Create test fixtures: commission -> shipment
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test Shipment", "active")

	t.Run("creates receipt successfully", func(t *testing.T) {
		record := &secondary.ReceiptRecord{
			ID:               "REC-001",
			ShipmentID:       "SHIP-001",
			DeliveredOutcome: "Delivered feature X",
			Status:           "submitted",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "REC-001")
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

	t.Run("creates receipt with evidence", func(t *testing.T) {
		db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-002", "COMM-001", "Test 2", "active")

		record := &secondary.ReceiptRecord{
			ID:                "REC-002",
			ShipmentID:        "SHIP-002",
			DeliveredOutcome:  "Delivered feature Y",
			Evidence:          "commit: abc123",
			VerificationNotes: "All tests pass",
			Status:            "submitted",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "REC-002")
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

	t.Run("enforces unique shipment constraint", func(t *testing.T) {
		record := &secondary.ReceiptRecord{
			ID:               "REC-003",
			ShipmentID:       "SHIP-001", // Same shipment as REC-001
			DeliveredOutcome: "Duplicate",
			Status:           "submitted",
		}

		err := repo.Create(ctx, record)
		if err == nil {
			t.Fatal("Expected error for duplicate shipment, got nil")
		}
	})
}

func TestReceiptRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewReceiptRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")

	repo.Create(ctx, &secondary.ReceiptRecord{
		ID:               "REC-001",
		ShipmentID:       "SHIP-001",
		DeliveredOutcome: "Test outcome",
		Status:           "submitted",
	})

	t.Run("finds receipt by ID", func(t *testing.T) {
		got, err := repo.GetByID(ctx, "REC-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if got.ID != "REC-001" {
			t.Errorf("ID = %q, want %q", got.ID, "REC-001")
		}
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "REC-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestReceiptRepository_GetByShipment(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewReceiptRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")

	repo.Create(ctx, &secondary.ReceiptRecord{
		ID:               "REC-001",
		ShipmentID:       "SHIP-001",
		DeliveredOutcome: "Test outcome",
		Status:           "submitted",
	})

	t.Run("finds receipt by shipment", func(t *testing.T) {
		got, err := repo.GetByShipment(ctx, "SHIP-001")
		if err != nil {
			t.Fatalf("GetByShipment failed: %v", err)
		}
		if got == nil {
			t.Fatal("expected receipt, got nil")
		}
		if got.ID != "REC-001" {
			t.Errorf("ID = %q, want %q", got.ID, "REC-001")
		}
	})

	t.Run("returns error for shipment without receipt", func(t *testing.T) {
		_, err := repo.GetByShipment(ctx, "SHIP-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestReceiptRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewReceiptRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test 1", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-002", "COMM-001", "Test 2", "active")

	repo.Create(ctx, &secondary.ReceiptRecord{ID: "REC-001", ShipmentID: "SHIP-001", DeliveredOutcome: "Test 1", Status: "submitted"})
	repo.Create(ctx, &secondary.ReceiptRecord{ID: "REC-002", ShipmentID: "SHIP-002", DeliveredOutcome: "Test 2", Status: "verified"})

	t.Run("lists all receipts", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.ReceiptFilters{})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("len = %d, want 2", len(list))
		}
	})

	t.Run("filters by shipment_id", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.ReceiptFilters{ShipmentID: "SHIP-002"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "REC-002" {
			t.Errorf("ID = %q, want %q", list[0].ID, "REC-002")
		}
	})

	t.Run("filters by status", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.ReceiptFilters{Status: "verified"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "REC-002" {
			t.Errorf("ID = %q, want %q", list[0].ID, "REC-002")
		}
	})
}

func TestReceiptRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewReceiptRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")

	repo.Create(ctx, &secondary.ReceiptRecord{
		ID:               "REC-001",
		ShipmentID:       "SHIP-001",
		DeliveredOutcome: "Original outcome",
		Status:           "submitted",
	})

	t.Run("updates delivered outcome", func(t *testing.T) {
		err := repo.Update(ctx, &secondary.ReceiptRecord{
			ID:               "REC-001",
			DeliveredOutcome: "Updated outcome",
		})
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "REC-001")
		if got.DeliveredOutcome != "Updated outcome" {
			t.Errorf("DeliveredOutcome = %q, want %q", got.DeliveredOutcome, "Updated outcome")
		}
	})

	t.Run("returns error for non-existent REC", func(t *testing.T) {
		err := repo.Update(ctx, &secondary.ReceiptRecord{
			ID:               "REC-999",
			DeliveredOutcome: "Will fail",
		})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestReceiptRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewReceiptRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")

	repo.Create(ctx, &secondary.ReceiptRecord{ID: "REC-001", ShipmentID: "SHIP-001", DeliveredOutcome: "Test", Status: "submitted"})

	t.Run("deletes receipt", func(t *testing.T) {
		err := repo.Delete(ctx, "REC-001")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, err = repo.GetByID(ctx, "REC-001")
		if err == nil {
			t.Error("expected error after delete, got nil")
		}
	})

	t.Run("returns error for non-existent REC", func(t *testing.T) {
		err := repo.Delete(ctx, "REC-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestReceiptRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewReceiptRepository(db)
	ctx := context.Background()

	t.Run("returns REC-001 for empty table", func(t *testing.T) {
		id, err := repo.GetNextID(ctx)
		if err != nil {
			t.Fatalf("GetNextID failed: %v", err)
		}
		if id != "REC-001" {
			t.Errorf("ID = %q, want %q", id, "REC-001")
		}
	})
}

func TestReceiptRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewReceiptRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")

	repo.Create(ctx, &secondary.ReceiptRecord{
		ID:               "REC-001",
		ShipmentID:       "SHIP-001",
		DeliveredOutcome: "Test",
		Status:           "submitted",
	})

	t.Run("updates to verified", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "REC-001", "verified")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "REC-001")
		if got.Status != "verified" {
			t.Errorf("Status = %q, want %q", got.Status, "verified")
		}
	})

	t.Run("updates to submitted", func(t *testing.T) {
		// First reset to draft
		repo.UpdateStatus(ctx, "REC-001", "draft")

		err := repo.UpdateStatus(ctx, "REC-001", "submitted")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "REC-001")
		if got.Status != "submitted" {
			t.Errorf("Status = %q, want %q", got.Status, "submitted")
		}
	})

	t.Run("returns error for non-existent REC", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "REC-999", "verified")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestReceiptRepository_ShipmentExists(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewReceiptRepository(db)
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

func TestReceiptRepository_ShipmentHasREC(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewReceiptRepository(db)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test 1", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-002", "COMM-001", "Test 2", "active")

	repo.Create(ctx, &secondary.ReceiptRecord{
		ID:               "REC-001",
		ShipmentID:       "SHIP-001",
		DeliveredOutcome: "Test",
		Status:           "submitted",
	})

	t.Run("returns true when shipment has REC", func(t *testing.T) {
		has, err := repo.ShipmentHasREC(ctx, "SHIP-001")
		if err != nil {
			t.Fatalf("ShipmentHasREC failed: %v", err)
		}
		if !has {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns false when shipment has no REC", func(t *testing.T) {
		has, err := repo.ShipmentHasREC(ctx, "SHIP-002")
		if err != nil {
			t.Fatalf("ShipmentHasREC failed: %v", err)
		}
		if has {
			t.Error("expected false, got true")
		}
	})
}
