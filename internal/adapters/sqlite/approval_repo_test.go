package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestApprovalRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewApprovalRepository(db, nil)
	ctx := context.Background()

	// Create test fixtures: commission -> shipment -> task, plan
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test Shipment", "draft")
	db.ExecContext(ctx, "INSERT INTO tasks (id, shipment_id, commission_id, title, status) VALUES (?, ?, ?, ?, ?)", "TASK-001", "SHIP-001", "COMM-001", "Test Task", "ready")
	db.ExecContext(ctx, "INSERT INTO plans (id, commission_id, task_id, title, status) VALUES (?, ?, ?, ?, ?)", "PLAN-001", "COMM-001", "TASK-001", "Test Plan", "draft")

	t.Run("creates approval successfully", func(t *testing.T) {
		record := &secondary.ApprovalRecord{
			ID:        "APPR-001",
			PlanID:    "PLAN-001",
			TaskID:    "TASK-001",
			Mechanism: "subagent",
			Outcome:   "approved",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "APPR-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.PlanID != "PLAN-001" {
			t.Errorf("PlanID = %q, want %q", got.PlanID, "PLAN-001")
		}
		if got.Mechanism != "subagent" {
			t.Errorf("Mechanism = %q, want %q", got.Mechanism, "subagent")
		}
		if got.Outcome != "approved" {
			t.Errorf("Outcome = %q, want %q", got.Outcome, "approved")
		}
	})

	t.Run("creates approval with reviewer input/output", func(t *testing.T) {
		db.ExecContext(ctx, "INSERT INTO plans (id, commission_id, task_id, title, status) VALUES (?, ?, ?, ?, ?)", "PLAN-002", "COMM-001", "TASK-001", "Test Plan 2", "draft")

		record := &secondary.ApprovalRecord{
			ID:             "APPR-002",
			PlanID:         "PLAN-002",
			TaskID:         "TASK-001",
			Mechanism:      "manual",
			ReviewerInput:  "LGTM",
			ReviewerOutput: "Approved with suggestions",
			Outcome:        "approved",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "APPR-002")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.ReviewerInput != "LGTM" {
			t.Errorf("ReviewerInput = %q, want %q", got.ReviewerInput, "LGTM")
		}
		if got.ReviewerOutput != "Approved with suggestions" {
			t.Errorf("ReviewerOutput = %q, want %q", got.ReviewerOutput, "Approved with suggestions")
		}
	})

	t.Run("enforces unique plan constraint", func(t *testing.T) {
		record := &secondary.ApprovalRecord{
			ID:        "APPR-003",
			PlanID:    "PLAN-001", // Same plan as APPR-001
			TaskID:    "TASK-001",
			Mechanism: "manual",
			Outcome:   "escalated",
		}

		err := repo.Create(ctx, record)
		if err == nil {
			t.Fatal("Expected error for duplicate plan, got nil")
		}
	})
}

func TestApprovalRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewApprovalRepository(db, nil)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "draft")
	db.ExecContext(ctx, "INSERT INTO tasks (id, shipment_id, commission_id, title, status) VALUES (?, ?, ?, ?, ?)", "TASK-001", "SHIP-001", "COMM-001", "Test", "ready")
	db.ExecContext(ctx, "INSERT INTO plans (id, commission_id, task_id, title, status) VALUES (?, ?, ?, ?, ?)", "PLAN-001", "COMM-001", "TASK-001", "Test", "draft")

	repo.Create(ctx, &secondary.ApprovalRecord{
		ID:        "APPR-001",
		PlanID:    "PLAN-001",
		TaskID:    "TASK-001",
		Mechanism: "subagent",
		Outcome:   "approved",
	})

	t.Run("finds approval by ID", func(t *testing.T) {
		got, err := repo.GetByID(ctx, "APPR-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if got.ID != "APPR-001" {
			t.Errorf("ID = %q, want %q", got.ID, "APPR-001")
		}
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "APPR-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestApprovalRepository_GetByPlan(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewApprovalRepository(db, nil)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "draft")
	db.ExecContext(ctx, "INSERT INTO tasks (id, shipment_id, commission_id, title, status) VALUES (?, ?, ?, ?, ?)", "TASK-001", "SHIP-001", "COMM-001", "Test", "ready")
	db.ExecContext(ctx, "INSERT INTO plans (id, commission_id, task_id, title, status) VALUES (?, ?, ?, ?, ?)", "PLAN-001", "COMM-001", "TASK-001", "Test", "draft")

	repo.Create(ctx, &secondary.ApprovalRecord{
		ID:        "APPR-001",
		PlanID:    "PLAN-001",
		TaskID:    "TASK-001",
		Mechanism: "subagent",
		Outcome:   "approved",
	})

	t.Run("finds approval by plan", func(t *testing.T) {
		got, err := repo.GetByPlan(ctx, "PLAN-001")
		if err != nil {
			t.Fatalf("GetByPlan failed: %v", err)
		}
		if got == nil {
			t.Fatal("expected approval, got nil")
		}
		if got.ID != "APPR-001" {
			t.Errorf("ID = %q, want %q", got.ID, "APPR-001")
		}
	})

	t.Run("returns error for plan without approval", func(t *testing.T) {
		_, err := repo.GetByPlan(ctx, "PLAN-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestApprovalRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewApprovalRepository(db, nil)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "draft")
	db.ExecContext(ctx, "INSERT INTO tasks (id, shipment_id, commission_id, title, status) VALUES (?, ?, ?, ?, ?)", "TASK-001", "SHIP-001", "COMM-001", "Test 1", "ready")
	db.ExecContext(ctx, "INSERT INTO tasks (id, shipment_id, commission_id, title, status) VALUES (?, ?, ?, ?, ?)", "TASK-002", "SHIP-001", "COMM-001", "Test 2", "ready")
	db.ExecContext(ctx, "INSERT INTO plans (id, commission_id, task_id, title, status) VALUES (?, ?, ?, ?, ?)", "PLAN-001", "COMM-001", "TASK-001", "Test 1", "draft")
	db.ExecContext(ctx, "INSERT INTO plans (id, commission_id, task_id, title, status) VALUES (?, ?, ?, ?, ?)", "PLAN-002", "COMM-001", "TASK-002", "Test 2", "draft")

	repo.Create(ctx, &secondary.ApprovalRecord{ID: "APPR-001", PlanID: "PLAN-001", TaskID: "TASK-001", Mechanism: "subagent", Outcome: "approved"})
	repo.Create(ctx, &secondary.ApprovalRecord{ID: "APPR-002", PlanID: "PLAN-002", TaskID: "TASK-002", Mechanism: "manual", Outcome: "escalated"})

	t.Run("lists all approvals", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.ApprovalFilters{})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("len = %d, want 2", len(list))
		}
	})

	t.Run("filters by task_id", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.ApprovalFilters{TaskID: "TASK-002"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "APPR-002" {
			t.Errorf("ID = %q, want %q", list[0].ID, "APPR-002")
		}
	})

	t.Run("filters by outcome", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.ApprovalFilters{Outcome: "escalated"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "APPR-002" {
			t.Errorf("ID = %q, want %q", list[0].ID, "APPR-002")
		}
	})
}

func TestApprovalRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewApprovalRepository(db, nil)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "draft")
	db.ExecContext(ctx, "INSERT INTO tasks (id, shipment_id, commission_id, title, status) VALUES (?, ?, ?, ?, ?)", "TASK-001", "SHIP-001", "COMM-001", "Test", "ready")
	db.ExecContext(ctx, "INSERT INTO plans (id, commission_id, task_id, title, status) VALUES (?, ?, ?, ?, ?)", "PLAN-001", "COMM-001", "TASK-001", "Test", "draft")

	repo.Create(ctx, &secondary.ApprovalRecord{ID: "APPR-001", PlanID: "PLAN-001", TaskID: "TASK-001", Mechanism: "subagent", Outcome: "approved"})

	t.Run("deletes approval", func(t *testing.T) {
		err := repo.Delete(ctx, "APPR-001")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, err = repo.GetByID(ctx, "APPR-001")
		if err == nil {
			t.Error("expected error after delete, got nil")
		}
	})

	t.Run("returns error for non-existent approval", func(t *testing.T) {
		err := repo.Delete(ctx, "APPR-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestApprovalRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewApprovalRepository(db, nil)
	ctx := context.Background()

	t.Run("returns APPR-001 for empty table", func(t *testing.T) {
		id, err := repo.GetNextID(ctx)
		if err != nil {
			t.Fatalf("GetNextID failed: %v", err)
		}
		if id != "APPR-001" {
			t.Errorf("ID = %q, want %q", id, "APPR-001")
		}
	})
}

func TestApprovalRepository_PlanExists(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewApprovalRepository(db, nil)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO tasks (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "TASK-001", "COMM-001", "Test", "ready")
	db.ExecContext(ctx, "INSERT INTO plans (id, commission_id, task_id, title, status) VALUES (?, ?, ?, ?, ?)", "PLAN-001", "COMM-001", "TASK-001", "Test", "draft")

	t.Run("returns true for existing plan", func(t *testing.T) {
		exists, err := repo.PlanExists(ctx, "PLAN-001")
		if err != nil {
			t.Fatalf("PlanExists failed: %v", err)
		}
		if !exists {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns false for non-existent plan", func(t *testing.T) {
		exists, err := repo.PlanExists(ctx, "PLAN-999")
		if err != nil {
			t.Fatalf("PlanExists failed: %v", err)
		}
		if exists {
			t.Error("expected false, got true")
		}
	})
}

func TestApprovalRepository_TaskExists(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewApprovalRepository(db, nil)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO tasks (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "TASK-001", "COMM-001", "Test", "ready")

	t.Run("returns true for existing task", func(t *testing.T) {
		exists, err := repo.TaskExists(ctx, "TASK-001")
		if err != nil {
			t.Fatalf("TaskExists failed: %v", err)
		}
		if !exists {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns false for non-existent task", func(t *testing.T) {
		exists, err := repo.TaskExists(ctx, "TASK-999")
		if err != nil {
			t.Fatalf("TaskExists failed: %v", err)
		}
		if exists {
			t.Error("expected false, got true")
		}
	})
}

func TestApprovalRepository_PlanHasApproval(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewApprovalRepository(db, nil)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "draft")
	db.ExecContext(ctx, "INSERT INTO tasks (id, shipment_id, commission_id, title, status) VALUES (?, ?, ?, ?, ?)", "TASK-001", "SHIP-001", "COMM-001", "Test", "ready")
	db.ExecContext(ctx, "INSERT INTO plans (id, commission_id, task_id, title, status) VALUES (?, ?, ?, ?, ?)", "PLAN-001", "COMM-001", "TASK-001", "Test 1", "draft")
	db.ExecContext(ctx, "INSERT INTO plans (id, commission_id, task_id, title, status) VALUES (?, ?, ?, ?, ?)", "PLAN-002", "COMM-001", "TASK-001", "Test 2", "draft")

	repo.Create(ctx, &secondary.ApprovalRecord{
		ID:        "APPR-001",
		PlanID:    "PLAN-001",
		TaskID:    "TASK-001",
		Mechanism: "subagent",
		Outcome:   "approved",
	})

	t.Run("returns true when plan has approval", func(t *testing.T) {
		has, err := repo.PlanHasApproval(ctx, "PLAN-001")
		if err != nil {
			t.Fatalf("PlanHasApproval failed: %v", err)
		}
		if !has {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns false when plan has no approval", func(t *testing.T) {
		has, err := repo.PlanHasApproval(ctx, "PLAN-002")
		if err != nil {
			t.Fatalf("PlanHasApproval failed: %v", err)
		}
		if has {
			t.Error("expected false, got true")
		}
	})
}
