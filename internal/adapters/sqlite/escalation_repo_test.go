package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestEscalationRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewEscalationRepository(db, nil)
	ctx := context.Background()

	// Create test fixtures
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test Shipment", "draft")
	db.ExecContext(ctx, "INSERT INTO tasks (id, shipment_id, commission_id, title, status) VALUES (?, ?, ?, ?, ?)", "TASK-001", "SHIP-001", "COMM-001", "Test Task", "ready")
	db.ExecContext(ctx, "INSERT INTO plans (id, commission_id, task_id, title, status) VALUES (?, ?, ?, ?, ?)", "PLAN-001", "COMM-001", "TASK-001", "Test Plan", "draft")

	t.Run("creates escalation successfully", func(t *testing.T) {
		record := &secondary.EscalationRecord{
			ID:            "ESC-001",
			PlanID:        "PLAN-001",
			TaskID:        "TASK-001",
			Reason:        "Plan complexity exceeds threshold",
			Status:        "pending",
			RoutingRule:   "workshop_gatehouse",
			OriginActorID: "IMP-BENCH-001",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "ESC-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.Reason != "Plan complexity exceeds threshold" {
			t.Errorf("Reason = %q, want %q", got.Reason, "Plan complexity exceeds threshold")
		}
		if got.Status != "pending" {
			t.Errorf("Status = %q, want %q", got.Status, "pending")
		}
		if got.OriginActorID != "IMP-BENCH-001" {
			t.Errorf("OriginActorID = %q, want %q", got.OriginActorID, "IMP-BENCH-001")
		}
	})

	t.Run("creates escalation with optional fields", func(t *testing.T) {
		db.ExecContext(ctx, "INSERT INTO plans (id, commission_id, task_id, title, status) VALUES (?, ?, ?, ?, ?)", "PLAN-002", "COMM-001", "TASK-001", "Test Plan 2", "draft")
		db.ExecContext(ctx, "INSERT INTO approvals (id, plan_id, task_id, mechanism, outcome) VALUES (?, ?, ?, ?, ?)", "APPR-001", "PLAN-002", "TASK-001", "subagent", "escalated")

		record := &secondary.EscalationRecord{
			ID:            "ESC-002",
			ApprovalID:    "APPR-001",
			PlanID:        "PLAN-002",
			TaskID:        "TASK-001",
			Reason:        "Requires human review",
			Status:        "pending",
			RoutingRule:   "workshop_gatehouse",
			OriginActorID: "IMP-BENCH-001",
			TargetActorID: "GATE-001",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "ESC-002")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.ApprovalID != "APPR-001" {
			t.Errorf("ApprovalID = %q, want %q", got.ApprovalID, "APPR-001")
		}
		if got.TargetActorID != "GATE-001" {
			t.Errorf("TargetActorID = %q, want %q", got.TargetActorID, "GATE-001")
		}
	})
}

func TestEscalationRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewEscalationRepository(db, nil)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "draft")
	db.ExecContext(ctx, "INSERT INTO tasks (id, shipment_id, commission_id, title, status) VALUES (?, ?, ?, ?, ?)", "TASK-001", "SHIP-001", "COMM-001", "Test", "ready")
	db.ExecContext(ctx, "INSERT INTO plans (id, commission_id, task_id, title, status) VALUES (?, ?, ?, ?, ?)", "PLAN-001", "COMM-001", "TASK-001", "Test", "draft")

	repo.Create(ctx, &secondary.EscalationRecord{
		ID:            "ESC-001",
		PlanID:        "PLAN-001",
		TaskID:        "TASK-001",
		Reason:        "Test reason",
		Status:        "pending",
		RoutingRule:   "workshop_gatehouse",
		OriginActorID: "IMP-BENCH-001",
	})

	t.Run("finds escalation by ID", func(t *testing.T) {
		got, err := repo.GetByID(ctx, "ESC-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if got.ID != "ESC-001" {
			t.Errorf("ID = %q, want %q", got.ID, "ESC-001")
		}
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "ESC-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestEscalationRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewEscalationRepository(db, nil)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "draft")
	db.ExecContext(ctx, "INSERT INTO tasks (id, shipment_id, commission_id, title, status) VALUES (?, ?, ?, ?, ?)", "TASK-001", "SHIP-001", "COMM-001", "Test", "ready")
	db.ExecContext(ctx, "INSERT INTO plans (id, commission_id, task_id, title, status) VALUES (?, ?, ?, ?, ?)", "PLAN-001", "COMM-001", "TASK-001", "Test 1", "draft")
	db.ExecContext(ctx, "INSERT INTO plans (id, commission_id, task_id, title, status) VALUES (?, ?, ?, ?, ?)", "PLAN-002", "COMM-001", "TASK-001", "Test 2", "draft")

	repo.Create(ctx, &secondary.EscalationRecord{ID: "ESC-001", PlanID: "PLAN-001", TaskID: "TASK-001", Reason: "Test 1", Status: "pending", RoutingRule: "workshop_gatehouse", OriginActorID: "IMP-BENCH-001"})
	repo.Create(ctx, &secondary.EscalationRecord{ID: "ESC-002", PlanID: "PLAN-002", TaskID: "TASK-001", Reason: "Test 2", Status: "resolved", RoutingRule: "workshop_gatehouse", OriginActorID: "IMP-BENCH-001", TargetActorID: "GATE-001"})

	t.Run("lists all escalations", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.EscalationFilters{})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("len = %d, want 2", len(list))
		}
	})

	t.Run("filters by status", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.EscalationFilters{Status: "pending"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "ESC-001" {
			t.Errorf("ID = %q, want %q", list[0].ID, "ESC-001")
		}
	})

	t.Run("filters by target_actor_id", func(t *testing.T) {
		list, err := repo.List(ctx, secondary.EscalationFilters{TargetActorID: "GATE-001"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("len = %d, want 1", len(list))
		}
		if list[0].ID != "ESC-002" {
			t.Errorf("ID = %q, want %q", list[0].ID, "ESC-002")
		}
	})
}

func TestEscalationRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewEscalationRepository(db, nil)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "draft")
	db.ExecContext(ctx, "INSERT INTO tasks (id, shipment_id, commission_id, title, status) VALUES (?, ?, ?, ?, ?)", "TASK-001", "SHIP-001", "COMM-001", "Test", "ready")
	db.ExecContext(ctx, "INSERT INTO plans (id, commission_id, task_id, title, status) VALUES (?, ?, ?, ?, ?)", "PLAN-001", "COMM-001", "TASK-001", "Test", "draft")

	repo.Create(ctx, &secondary.EscalationRecord{ID: "ESC-001", PlanID: "PLAN-001", TaskID: "TASK-001", Reason: "Test", Status: "pending", RoutingRule: "workshop_gatehouse", OriginActorID: "IMP-BENCH-001"})

	t.Run("deletes escalation", func(t *testing.T) {
		err := repo.Delete(ctx, "ESC-001")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, err = repo.GetByID(ctx, "ESC-001")
		if err == nil {
			t.Error("expected error after delete, got nil")
		}
	})

	t.Run("returns error for non-existent escalation", func(t *testing.T) {
		err := repo.Delete(ctx, "ESC-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestEscalationRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewEscalationRepository(db, nil)
	ctx := context.Background()

	t.Run("returns ESC-001 for empty table", func(t *testing.T) {
		id, err := repo.GetNextID(ctx)
		if err != nil {
			t.Fatalf("GetNextID failed: %v", err)
		}
		if id != "ESC-001" {
			t.Errorf("ID = %q, want %q", id, "ESC-001")
		}
	})
}

func TestEscalationRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewEscalationRepository(db, nil)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "draft")
	db.ExecContext(ctx, "INSERT INTO tasks (id, shipment_id, commission_id, title, status) VALUES (?, ?, ?, ?, ?)", "TASK-001", "SHIP-001", "COMM-001", "Test", "ready")
	db.ExecContext(ctx, "INSERT INTO plans (id, commission_id, task_id, title, status) VALUES (?, ?, ?, ?, ?)", "PLAN-001", "COMM-001", "TASK-001", "Test", "draft")

	repo.Create(ctx, &secondary.EscalationRecord{
		ID:            "ESC-001",
		PlanID:        "PLAN-001",
		TaskID:        "TASK-001",
		Reason:        "Test",
		Status:        "pending",
		RoutingRule:   "workshop_gatehouse",
		OriginActorID: "IMP-BENCH-001",
	})

	t.Run("updates to resolved with timestamp", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "ESC-001", "resolved", true)
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "ESC-001")
		if got.Status != "resolved" {
			t.Errorf("Status = %q, want %q", got.Status, "resolved")
		}
		if got.ResolvedAt == "" {
			t.Error("expected ResolvedAt to be set")
		}
	})

	t.Run("returns error for non-existent escalation", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "ESC-999", "resolved", true)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestEscalationRepository_Resolve(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewEscalationRepository(db, nil)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "draft")
	db.ExecContext(ctx, "INSERT INTO tasks (id, shipment_id, commission_id, title, status) VALUES (?, ?, ?, ?, ?)", "TASK-001", "SHIP-001", "COMM-001", "Test", "ready")
	db.ExecContext(ctx, "INSERT INTO plans (id, commission_id, task_id, title, status) VALUES (?, ?, ?, ?, ?)", "PLAN-001", "COMM-001", "TASK-001", "Test", "draft")

	repo.Create(ctx, &secondary.EscalationRecord{
		ID:            "ESC-001",
		PlanID:        "PLAN-001",
		TaskID:        "TASK-001",
		Reason:        "Test",
		Status:        "pending",
		RoutingRule:   "workshop_gatehouse",
		OriginActorID: "IMP-BENCH-001",
	})

	t.Run("resolves escalation", func(t *testing.T) {
		err := repo.Resolve(ctx, "ESC-001", "Approved plan with modifications", "GATE-001")
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "ESC-001")
		if got.Status != "resolved" {
			t.Errorf("Status = %q, want %q", got.Status, "resolved")
		}
		if got.Resolution != "Approved plan with modifications" {
			t.Errorf("Resolution = %q, want %q", got.Resolution, "Approved plan with modifications")
		}
		if got.ResolvedBy != "GATE-001" {
			t.Errorf("ResolvedBy = %q, want %q", got.ResolvedBy, "GATE-001")
		}
		if got.ResolvedAt == "" {
			t.Error("expected ResolvedAt to be set")
		}
	})

	t.Run("returns error for non-existent escalation", func(t *testing.T) {
		err := repo.Resolve(ctx, "ESC-999", "Resolution", "GATE-001")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestEscalationRepository_Exists(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewEscalationRepository(db, nil)
	ctx := context.Background()

	// Setup
	db.ExecContext(ctx, "INSERT INTO commissions (id, title, status) VALUES (?, ?, ?)", "COMM-001", "Test", "active")
	db.ExecContext(ctx, "INSERT INTO shipments (id, commission_id, title, status) VALUES (?, ?, ?, ?)", "SHIP-001", "COMM-001", "Test", "draft")
	db.ExecContext(ctx, "INSERT INTO tasks (id, shipment_id, commission_id, title, status) VALUES (?, ?, ?, ?, ?)", "TASK-001", "SHIP-001", "COMM-001", "Test", "ready")
	db.ExecContext(ctx, "INSERT INTO plans (id, commission_id, task_id, title, status) VALUES (?, ?, ?, ?, ?)", "PLAN-001", "COMM-001", "TASK-001", "Test", "draft")
	db.ExecContext(ctx, "INSERT INTO approvals (id, plan_id, task_id, mechanism, outcome) VALUES (?, ?, ?, ?, ?)", "APPR-001", "PLAN-001", "TASK-001", "subagent", "escalated")

	t.Run("PlanExists returns true", func(t *testing.T) {
		exists, err := repo.PlanExists(ctx, "PLAN-001")
		if err != nil {
			t.Fatalf("PlanExists failed: %v", err)
		}
		if !exists {
			t.Error("expected true, got false")
		}
	})

	t.Run("TaskExists returns true", func(t *testing.T) {
		exists, err := repo.TaskExists(ctx, "TASK-001")
		if err != nil {
			t.Fatalf("TaskExists failed: %v", err)
		}
		if !exists {
			t.Error("expected true, got false")
		}
	})

	t.Run("ApprovalExists returns true", func(t *testing.T) {
		exists, err := repo.ApprovalExists(ctx, "APPR-001")
		if err != nil {
			t.Fatalf("ApprovalExists failed: %v", err)
		}
		if !exists {
			t.Error("expected true, got false")
		}
	})
}
