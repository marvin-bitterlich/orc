package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

// setupPlanTestDB creates the test database with required seed data.
func setupPlanTestDB(t *testing.T) *sql.DB {
	t.Helper()
	testDB := setupTestDB(t)
	seedCommission(t, testDB, "COMM-001", "Test Commission")
	seedShipment(t, testDB, "SHIP-001", "COMM-001", "Test Shipment")
	return testDB
}

// createTestPlan is a helper that creates a plan with a generated ID.
func createTestPlan(t *testing.T, repo *sqlite.PlanRepository, ctx context.Context, commissionID, shipmentID, title string) *secondary.PlanRecord {
	t.Helper()

	nextID, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}

	plan := &secondary.PlanRecord{
		ID:           nextID,
		CommissionID: commissionID,
		ShipmentID:   shipmentID,
		Title:        title,
	}

	err = repo.Create(ctx, plan)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	return plan
}

func TestPlanRepository_Create(t *testing.T) {
	db := setupPlanTestDB(t)
	repo := sqlite.NewPlanRepository(db)
	ctx := context.Background()

	plan := &secondary.PlanRecord{
		ID:           "PLAN-001",
		CommissionID: "COMM-001",
		ShipmentID:   "SHIP-001",
		Title:        "Test Plan",
		Description:  "A test plan description",
		Content:      "## Plan Content\n\n1. Step one\n2. Step two",
	}

	err := repo.Create(ctx, plan)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify plan was created
	retrieved, err := repo.GetByID(ctx, "PLAN-001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Title != "Test Plan" {
		t.Errorf("expected title 'Test Plan', got '%s'", retrieved.Title)
	}
	if retrieved.Status != "draft" {
		t.Errorf("expected status 'draft', got '%s'", retrieved.Status)
	}
	if retrieved.Content != "## Plan Content\n\n1. Step one\n2. Step two" {
		t.Errorf("unexpected content: %s", retrieved.Content)
	}
}

func TestPlanRepository_Create_WithoutShipment(t *testing.T) {
	db := setupPlanTestDB(t)
	repo := sqlite.NewPlanRepository(db)
	ctx := context.Background()

	plan := &secondary.PlanRecord{
		ID:           "PLAN-001",
		CommissionID: "COMM-001",
		Title:        "Standalone Plan",
	}

	err := repo.Create(ctx, plan)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, "PLAN-001")
	if retrieved.ShipmentID != "" {
		t.Errorf("expected empty shipment ID, got '%s'", retrieved.ShipmentID)
	}
}

func TestPlanRepository_GetByID(t *testing.T) {
	db := setupPlanTestDB(t)
	repo := sqlite.NewPlanRepository(db)
	ctx := context.Background()

	plan := createTestPlan(t, repo, ctx, "COMM-001", "SHIP-001", "Test Plan")

	retrieved, err := repo.GetByID(ctx, plan.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.Title != "Test Plan" {
		t.Errorf("expected title 'Test Plan', got '%s'", retrieved.Title)
	}
	if retrieved.CommissionID != "COMM-001" {
		t.Errorf("expected commission 'COMM-001', got '%s'", retrieved.CommissionID)
	}
}

func TestPlanRepository_GetByID_NotFound(t *testing.T) {
	db := setupPlanTestDB(t)
	repo := sqlite.NewPlanRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "PLAN-999")
	if err == nil {
		t.Error("expected error for non-existent plan")
	}
}

func TestPlanRepository_List(t *testing.T) {
	db := setupPlanTestDB(t)
	repo := sqlite.NewPlanRepository(db)
	ctx := context.Background()

	createTestPlan(t, repo, ctx, "COMM-001", "", "Plan 1")
	createTestPlan(t, repo, ctx, "COMM-001", "", "Plan 2")
	createTestPlan(t, repo, ctx, "COMM-001", "", "Plan 3")

	plans, err := repo.List(ctx, secondary.PlanFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(plans) != 3 {
		t.Errorf("expected 3 plans, got %d", len(plans))
	}
}

func TestPlanRepository_List_FilterByShipment(t *testing.T) {
	db := setupPlanTestDB(t)
	repo := sqlite.NewPlanRepository(db)
	ctx := context.Background()

	// Add another shipment
	_, _ = db.Exec("INSERT INTO shipments (id, commission_id, title) VALUES ('SHIP-002', 'COMM-001', 'Ship 2')")

	createTestPlan(t, repo, ctx, "COMM-001", "SHIP-001", "Plan 1")
	createTestPlan(t, repo, ctx, "COMM-001", "SHIP-002", "Plan 2")

	plans, err := repo.List(ctx, secondary.PlanFilters{ShipmentID: "SHIP-001"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(plans) != 1 {
		t.Errorf("expected 1 plan for SHIP-001, got %d", len(plans))
	}
}

func TestPlanRepository_List_FilterByCommission(t *testing.T) {
	db := setupPlanTestDB(t)
	repo := sqlite.NewPlanRepository(db)
	ctx := context.Background()

	// Add another commission
	_, _ = db.Exec("INSERT INTO commissions (id, title, status) VALUES ('COMM-002', 'Commission 2', 'active')")

	createTestPlan(t, repo, ctx, "COMM-001", "", "Plan 1")
	createTestPlan(t, repo, ctx, "COMM-002", "", "Plan 2")

	plans, err := repo.List(ctx, secondary.PlanFilters{CommissionID: "COMM-001"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(plans) != 1 {
		t.Errorf("expected 1 plan for COMM-001, got %d", len(plans))
	}
}

func TestPlanRepository_List_FilterByStatus(t *testing.T) {
	db := setupPlanTestDB(t)
	repo := sqlite.NewPlanRepository(db)
	ctx := context.Background()

	plan1 := createTestPlan(t, repo, ctx, "COMM-001", "", "Draft Plan")
	createTestPlan(t, repo, ctx, "COMM-001", "", "Another Draft Plan")

	// Approve plan1
	_ = repo.Approve(ctx, plan1.ID)

	plans, err := repo.List(ctx, secondary.PlanFilters{Status: "draft"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(plans) != 1 {
		t.Errorf("expected 1 draft plan, got %d", len(plans))
	}
}

func TestPlanRepository_Update(t *testing.T) {
	db := setupPlanTestDB(t)
	repo := sqlite.NewPlanRepository(db)
	ctx := context.Background()

	plan := createTestPlan(t, repo, ctx, "COMM-001", "", "Original Title")

	err := repo.Update(ctx, &secondary.PlanRecord{
		ID:      plan.ID,
		Title:   "Updated Title",
		Content: "New content",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, plan.ID)
	if retrieved.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", retrieved.Title)
	}
	if retrieved.Content != "New content" {
		t.Errorf("expected content 'New content', got '%s'", retrieved.Content)
	}
}

func TestPlanRepository_Update_NotFound(t *testing.T) {
	db := setupPlanTestDB(t)
	repo := sqlite.NewPlanRepository(db)
	ctx := context.Background()

	err := repo.Update(ctx, &secondary.PlanRecord{
		ID:    "PLAN-999",
		Title: "Updated Title",
	})
	if err == nil {
		t.Error("expected error for non-existent plan")
	}
}

func TestPlanRepository_Delete(t *testing.T) {
	db := setupPlanTestDB(t)
	repo := sqlite.NewPlanRepository(db)
	ctx := context.Background()

	plan := createTestPlan(t, repo, ctx, "COMM-001", "", "To Delete")

	err := repo.Delete(ctx, plan.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.GetByID(ctx, plan.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestPlanRepository_Delete_NotFound(t *testing.T) {
	db := setupPlanTestDB(t)
	repo := sqlite.NewPlanRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, "PLAN-999")
	if err == nil {
		t.Error("expected error for non-existent plan")
	}
}

func TestPlanRepository_Pin_Unpin(t *testing.T) {
	db := setupPlanTestDB(t)
	repo := sqlite.NewPlanRepository(db)
	ctx := context.Background()

	plan := createTestPlan(t, repo, ctx, "COMM-001", "", "Pin Test")

	// Pin
	err := repo.Pin(ctx, plan.ID)
	if err != nil {
		t.Fatalf("Pin failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, plan.ID)
	if !retrieved.Pinned {
		t.Error("expected plan to be pinned")
	}

	// Unpin
	err = repo.Unpin(ctx, plan.ID)
	if err != nil {
		t.Fatalf("Unpin failed: %v", err)
	}

	retrieved, _ = repo.GetByID(ctx, plan.ID)
	if retrieved.Pinned {
		t.Error("expected plan to be unpinned")
	}
}

func TestPlanRepository_Pin_NotFound(t *testing.T) {
	db := setupPlanTestDB(t)
	repo := sqlite.NewPlanRepository(db)
	ctx := context.Background()

	err := repo.Pin(ctx, "PLAN-999")
	if err == nil {
		t.Error("expected error for non-existent plan")
	}
}

func TestPlanRepository_GetNextID(t *testing.T) {
	db := setupPlanTestDB(t)
	repo := sqlite.NewPlanRepository(db)
	ctx := context.Background()

	id, err := repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "PLAN-001" {
		t.Errorf("expected PLAN-001, got %s", id)
	}

	createTestPlan(t, repo, ctx, "COMM-001", "", "Test")

	id, err = repo.GetNextID(ctx)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "PLAN-002" {
		t.Errorf("expected PLAN-002, got %s", id)
	}
}

func TestPlanRepository_Approve(t *testing.T) {
	db := setupPlanTestDB(t)
	repo := sqlite.NewPlanRepository(db)
	ctx := context.Background()

	plan := createTestPlan(t, repo, ctx, "COMM-001", "", "Plan to Approve")

	err := repo.Approve(ctx, plan.ID)
	if err != nil {
		t.Fatalf("Approve failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, plan.ID)
	if retrieved.Status != "approved" {
		t.Errorf("expected status 'approved', got '%s'", retrieved.Status)
	}
	if retrieved.ApprovedAt == "" {
		t.Error("expected ApprovedAt to be set")
	}
}

func TestPlanRepository_Approve_NotFound(t *testing.T) {
	db := setupPlanTestDB(t)
	repo := sqlite.NewPlanRepository(db)
	ctx := context.Background()

	err := repo.Approve(ctx, "PLAN-999")
	if err == nil {
		t.Error("expected error for non-existent plan")
	}
}

func TestPlanRepository_GetActivePlanForShipment(t *testing.T) {
	db := setupPlanTestDB(t)
	repo := sqlite.NewPlanRepository(db)
	ctx := context.Background()

	// Create a draft plan for shipment
	plan := createTestPlan(t, repo, ctx, "COMM-001", "SHIP-001", "Active Plan")

	// Get active plan
	active, err := repo.GetActivePlanForShipment(ctx, "SHIP-001")
	if err != nil {
		t.Fatalf("GetActivePlanForShipment failed: %v", err)
	}
	if active == nil {
		t.Fatal("expected active plan to be returned")
	}
	if active.ID != plan.ID {
		t.Errorf("expected plan ID '%s', got '%s'", plan.ID, active.ID)
	}

	// Approve the plan
	_ = repo.Approve(ctx, plan.ID)

	// Should return nil when no draft plan exists
	active, err = repo.GetActivePlanForShipment(ctx, "SHIP-001")
	if err != nil {
		t.Fatalf("GetActivePlanForShipment failed: %v", err)
	}
	if active != nil {
		t.Error("expected no active plan after approval")
	}
}

func TestPlanRepository_GetActivePlanForShipment_NoPlan(t *testing.T) {
	db := setupPlanTestDB(t)
	repo := sqlite.NewPlanRepository(db)
	ctx := context.Background()

	active, err := repo.GetActivePlanForShipment(ctx, "SHIP-001")
	if err != nil {
		t.Fatalf("GetActivePlanForShipment failed: %v", err)
	}
	if active != nil {
		t.Error("expected no active plan")
	}
}

func TestPlanRepository_HasActivePlanForShipment(t *testing.T) {
	db := setupPlanTestDB(t)
	repo := sqlite.NewPlanRepository(db)
	ctx := context.Background()

	// Initially no active plan
	has, err := repo.HasActivePlanForShipment(ctx, "SHIP-001")
	if err != nil {
		t.Fatalf("HasActivePlanForShipment failed: %v", err)
	}
	if has {
		t.Error("expected no active plan")
	}

	// Create a draft plan
	plan := createTestPlan(t, repo, ctx, "COMM-001", "SHIP-001", "Draft Plan")

	has, err = repo.HasActivePlanForShipment(ctx, "SHIP-001")
	if err != nil {
		t.Fatalf("HasActivePlanForShipment failed: %v", err)
	}
	if !has {
		t.Error("expected active plan to exist")
	}

	// Approve the plan
	_ = repo.Approve(ctx, plan.ID)

	has, err = repo.HasActivePlanForShipment(ctx, "SHIP-001")
	if err != nil {
		t.Fatalf("HasActivePlanForShipment failed: %v", err)
	}
	if has {
		t.Error("expected no active plan after approval")
	}
}

func TestPlanRepository_CommissionExists(t *testing.T) {
	db := setupPlanTestDB(t)
	repo := sqlite.NewPlanRepository(db)
	ctx := context.Background()

	exists, err := repo.CommissionExists(ctx, "COMM-001")
	if err != nil {
		t.Fatalf("CommissionExists failed: %v", err)
	}
	if !exists {
		t.Error("expected commission to exist")
	}

	exists, err = repo.CommissionExists(ctx, "COMM-999")
	if err != nil {
		t.Fatalf("CommissionExists failed: %v", err)
	}
	if exists {
		t.Error("expected commission to not exist")
	}
}

func TestPlanRepository_ShipmentExists(t *testing.T) {
	db := setupPlanTestDB(t)
	repo := sqlite.NewPlanRepository(db)
	ctx := context.Background()

	exists, err := repo.ShipmentExists(ctx, "SHIP-001")
	if err != nil {
		t.Fatalf("ShipmentExists failed: %v", err)
	}
	if !exists {
		t.Error("expected shipment to exist")
	}

	exists, err = repo.ShipmentExists(ctx, "SHIP-999")
	if err != nil {
		t.Fatalf("ShipmentExists failed: %v", err)
	}
	if exists {
		t.Error("expected shipment to not exist")
	}
}
