package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestPRRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	prRepo := sqlite.NewPRRepository(db)
	repoRepo := sqlite.NewRepoRepository(db)
	ctx := context.Background()

	// Create a test repo first
	repoRecord := &secondary.RepoRecord{
		ID:   "REPO-001",
		Name: "test-repo",
	}
	if err := repoRepo.Create(ctx, repoRecord); err != nil {
		t.Fatalf("Failed to create repo: %v", err)
	}

	// Create a test shipment (assuming shipments table exists)
	_, err := db.ExecContext(ctx,
		"INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)",
		"SHIP-001", "COMM-001", "Test Shipment", "active",
	)
	if err != nil {
		t.Fatalf("Failed to create shipment: %v", err)
	}

	// Create a test commission (if needed)
	_, err = db.ExecContext(ctx,
		"INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)",
		"COMM-001", "Test Commission", "active",
	)
	if err != nil {
		t.Fatalf("Failed to create commission: %v", err)
	}

	t.Run("creates PR successfully", func(t *testing.T) {
		record := &secondary.PRRecord{
			ID:           "PR-001",
			ShipmentID:   "SHIP-001",
			RepoID:       "REPO-001",
			CommissionID: "COMM-001",
			Title:        "Test PR",
			Branch:       "feature/test",
			Status:       "open",
		}

		err := prRepo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		// Verify it was created
		got, err := prRepo.GetByID(ctx, "PR-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.Title != "Test PR" {
			t.Errorf("Title = %q, want %q", got.Title, "Test PR")
		}
		if got.Status != "open" {
			t.Errorf("Status = %q, want %q", got.Status, "open")
		}
	})

	t.Run("creates draft PR", func(t *testing.T) {
		// Create another shipment for this test
		_, err := db.ExecContext(ctx,
			"INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)",
			"SHIP-002", "COMM-001", "Test Shipment 2", "active",
		)
		if err != nil {
			t.Fatalf("Failed to create shipment: %v", err)
		}

		record := &secondary.PRRecord{
			ID:           "PR-002",
			ShipmentID:   "SHIP-002",
			RepoID:       "REPO-001",
			CommissionID: "COMM-001",
			Title:        "Draft PR",
			Branch:       "feature/draft",
			Status:       "draft",
		}

		err = prRepo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := prRepo.GetByID(ctx, "PR-002")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.Status != "draft" {
			t.Errorf("Status = %q, want %q", got.Status, "draft")
		}
	})
}

func TestPRRepository_GetByShipment(t *testing.T) {
	db := setupTestDB(t)
	prRepo := sqlite.NewPRRepository(db)
	repoRepo := sqlite.NewRepoRepository(db)
	ctx := context.Background()

	// Setup
	repoRepo.Create(ctx, &secondary.RepoRecord{ID: "REPO-001", Name: "test-repo"})
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")

	prRepo.Create(ctx, &secondary.PRRecord{
		ID:           "PR-001",
		ShipmentID:   "SHIP-001",
		RepoID:       "REPO-001",
		CommissionID: "COMM-001",
		Title:        "Test PR",
		Branch:       "feature/test",
	})

	t.Run("finds PR by shipment", func(t *testing.T) {
		got, err := prRepo.GetByShipment(ctx, "SHIP-001")
		if err != nil {
			t.Fatalf("GetByShipment failed: %v", err)
		}
		if got == nil {
			t.Fatal("expected PR, got nil")
		}
		if got.ID != "PR-001" {
			t.Errorf("ID = %q, want %q", got.ID, "PR-001")
		}
	})

	t.Run("returns nil for shipment without PR", func(t *testing.T) {
		got, err := prRepo.GetByShipment(ctx, "SHIP-999")
		if err != nil {
			t.Fatalf("GetByShipment failed: %v", err)
		}
		if got != nil {
			t.Errorf("expected nil, got %+v", got)
		}
	})
}

func TestPRRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	prRepo := sqlite.NewPRRepository(db)
	repoRepo := sqlite.NewRepoRepository(db)
	ctx := context.Background()

	// Setup
	repoRepo.Create(ctx, &secondary.RepoRecord{ID: "REPO-001", Name: "test-repo"})
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "active")

	prRepo.Create(ctx, &secondary.PRRecord{
		ID:           "PR-001",
		ShipmentID:   "SHIP-001",
		RepoID:       "REPO-001",
		CommissionID: "COMM-001",
		Title:        "Test PR",
		Branch:       "feature/test",
		Status:       "open",
	})

	t.Run("updates to approved", func(t *testing.T) {
		err := prRepo.UpdateStatus(ctx, "PR-001", "approved", false, false)
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		got, _ := prRepo.GetByID(ctx, "PR-001")
		if got.Status != "approved" {
			t.Errorf("Status = %q, want %q", got.Status, "approved")
		}
	})

	t.Run("updates to merged with timestamp", func(t *testing.T) {
		err := prRepo.UpdateStatus(ctx, "PR-001", "merged", true, false)
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		got, _ := prRepo.GetByID(ctx, "PR-001")
		if got.Status != "merged" {
			t.Errorf("Status = %q, want %q", got.Status, "merged")
		}
		if got.MergedAt == "" {
			t.Error("MergedAt should be set")
		}
	})
}

func TestPRRepository_ShipmentHasPR(t *testing.T) {
	db := setupTestDB(t)
	prRepo := sqlite.NewPRRepository(db)
	repoRepo := sqlite.NewRepoRepository(db)
	ctx := context.Background()

	// Setup
	repoRepo.Create(ctx, &secondary.RepoRecord{ID: "REPO-001", Name: "test-repo"})
	db.ExecContext(ctx, "INSERT OR IGNORE INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test 1", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-002", "COMM-001", "Test 2", "active")

	prRepo.Create(ctx, &secondary.PRRecord{
		ID:           "PR-001",
		ShipmentID:   "SHIP-001",
		RepoID:       "REPO-001",
		CommissionID: "COMM-001",
		Title:        "Test PR",
		Branch:       "feature/test",
	})

	t.Run("returns true when shipment has PR", func(t *testing.T) {
		has, err := prRepo.ShipmentHasPR(ctx, "SHIP-001")
		if err != nil {
			t.Fatalf("ShipmentHasPR failed: %v", err)
		}
		if !has {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns false when shipment has no PR", func(t *testing.T) {
		has, err := prRepo.ShipmentHasPR(ctx, "SHIP-002")
		if err != nil {
			t.Fatalf("ShipmentHasPR failed: %v", err)
		}
		if has {
			t.Error("expected false, got true")
		}
	})
}

func TestPRRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	prRepo := sqlite.NewPRRepository(db)
	ctx := context.Background()

	t.Run("returns PR-001 for empty table", func(t *testing.T) {
		id, err := prRepo.GetNextID(ctx)
		if err != nil {
			t.Fatalf("GetNextID failed: %v", err)
		}
		if id != "PR-001" {
			t.Errorf("ID = %q, want %q", id, "PR-001")
		}
	})
}
