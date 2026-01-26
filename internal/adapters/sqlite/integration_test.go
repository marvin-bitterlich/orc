package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

// Integration tests verify cross-repository workflows and constraints.

// ============================================================================
// Commission Lifecycle Tests
// ============================================================================

func TestIntegration_CommissionWithShipmentsAndTasks(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	shipmentRepo := sqlite.NewShipmentRepository(db)
	taskRepo := sqlite.NewTaskRepository(db)

	// Commission doesn't exist - should return false
	exists, err := shipmentRepo.CommissionExists(ctx, "COMM-NONEXISTENT")
	if err != nil {
		t.Fatalf("CommissionExists failed: %v", err)
	}
	if exists {
		t.Error("expected CommissionExists to return false for non-existent commission")
	}

	// Create a real commission
	seedCommission(t, db, "COMM-001", "Test")

	exists, err = shipmentRepo.CommissionExists(ctx, "COMM-001")
	if err != nil {
		t.Fatalf("CommissionExists failed: %v", err)
	}
	if !exists {
		t.Error("expected CommissionExists to return true for existing commission")
	}

	// Create shipment and tasks
	if err := shipmentRepo.Create(ctx, &secondary.ShipmentRecord{ID: "SHIP-001", CommissionID: "COMM-001", Title: "Test"}); err != nil {
		t.Fatalf("Create shipment failed: %v", err)
	}
	if err := taskRepo.Create(ctx, &secondary.TaskRecord{ID: "TASK-001", CommissionID: "COMM-001", ShipmentID: "SHIP-001", Title: "Test"}); err != nil {
		t.Fatalf("Create task failed: %v", err)
	}

	// Verify retrieval
	tasks, err := taskRepo.GetByShipment(ctx, "SHIP-001")
	if err != nil {
		t.Fatalf("GetByShipment failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}
}

func TestIntegration_CommissionExistsConstraint(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	taskRepo := sqlite.NewTaskRepository(db)

	// Check non-existent commission
	exists, err := taskRepo.CommissionExists(ctx, "COMM-NONE")
	if err != nil {
		t.Fatalf("CommissionExists failed: %v", err)
	}
	if exists {
		t.Error("expected false for non-existent commission")
	}

	// Create commission and verify
	seedCommission(t, db, "COMM-001", "Test")
	exists, err = taskRepo.CommissionExists(ctx, "COMM-001")
	if err != nil {
		t.Fatalf("CommissionExists failed: %v", err)
	}
	if !exists {
		t.Error("expected true for existing commission")
	}
}

// ============================================================================
// Shipment Workflow Tests
// ============================================================================

func TestIntegration_ShipmentWithPlanAndTasks(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	seedCommission(t, db, "COMM-001", "Test Commission")

	shipmentRepo := sqlite.NewShipmentRepository(db)
	planRepo := sqlite.NewPlanRepository(db)
	taskRepo := sqlite.NewTaskRepository(db)

	// Create shipment
	shipment := &secondary.ShipmentRecord{
		ID:           "SHIP-001",
		CommissionID: "COMM-001",
		Title:        "Feature Shipment",
	}
	if err := shipmentRepo.Create(ctx, shipment); err != nil {
		t.Fatalf("Create shipment failed: %v", err)
	}

	// Create plan for shipment
	plan := &secondary.PlanRecord{
		ID:           "PLAN-001",
		CommissionID: "COMM-001",
		ShipmentID:   "SHIP-001",
		Title:        "Implementation Plan",
		Content:      "Plan content...",
	}
	if err := planRepo.Create(ctx, plan); err != nil {
		t.Fatalf("Create plan failed: %v", err)
	}

	// Verify active plan exists (HasActivePlanForShipment checks for draft status)
	hasActive, _ := planRepo.HasActivePlanForShipment(ctx, "SHIP-001")
	if !hasActive {
		t.Error("expected shipment to have active (draft) plan")
	}

	// Approve plan
	if err := planRepo.Approve(ctx, "PLAN-001"); err != nil {
		t.Fatalf("Approve plan failed: %v", err)
	}

	// After approval, no more draft plans
	hasActive, _ = planRepo.HasActivePlanForShipment(ctx, "SHIP-001")
	if hasActive {
		t.Error("expected no draft plan after approval")
	}

	// Create tasks for shipment
	task := &secondary.TaskRecord{
		ID:           "TASK-001",
		CommissionID: "COMM-001",
		ShipmentID:   "SHIP-001",
		Title:        "Implementation Task",
	}
	if err := taskRepo.Create(ctx, task); err != nil {
		t.Fatalf("Create task failed: %v", err)
	}

	// Verify tasks by shipment
	tasks, _ := taskRepo.GetByShipment(ctx, "SHIP-001")
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}
}

// ============================================================================
// Workbench Assignment Tests
// ============================================================================

func TestIntegration_WorkbenchAssignmentPropagation(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	seedCommission(t, db, "COMM-001", "Test")
	seedWorkbench(t, db, "BENCH-001", "COMM-001", "test-bench")

	shipmentRepo := sqlite.NewShipmentRepository(db)
	taskRepo := sqlite.NewTaskRepository(db)

	// Create shipment and tasks
	if err := shipmentRepo.Create(ctx, &secondary.ShipmentRecord{ID: "SHIP-001", CommissionID: "COMM-001", Title: "Test"}); err != nil {
		t.Fatalf("Create shipment failed: %v", err)
	}
	if err := taskRepo.Create(ctx, &secondary.TaskRecord{ID: "TASK-001", CommissionID: "COMM-001", ShipmentID: "SHIP-001", Title: "Task 1"}); err != nil {
		t.Fatalf("Create task 1 failed: %v", err)
	}
	if err := taskRepo.Create(ctx, &secondary.TaskRecord{ID: "TASK-002", CommissionID: "COMM-001", ShipmentID: "SHIP-001", Title: "Task 2"}); err != nil {
		t.Fatalf("Create task 2 failed: %v", err)
	}

	// Assign workbench to shipment
	if err := shipmentRepo.AssignWorkbench(ctx, "SHIP-001", "BENCH-001"); err != nil {
		t.Fatalf("AssignWorkbench failed: %v", err)
	}

	// Propagate to tasks
	if err := taskRepo.AssignWorkbenchByShipment(ctx, "SHIP-001", "BENCH-001"); err != nil {
		t.Fatalf("AssignWorkbenchByShipment failed: %v", err)
	}

	// Verify shipment has workbench
	shipments, err := shipmentRepo.GetByWorkbench(ctx, "BENCH-001")
	if err != nil {
		t.Fatalf("GetByWorkbench for shipments failed: %v", err)
	}
	if len(shipments) != 1 {
		t.Errorf("expected 1 shipment assigned to workbench, got %d", len(shipments))
	}

	// Verify tasks have workbench
	tasks, err := taskRepo.GetByWorkbench(ctx, "BENCH-001")
	if err != nil {
		t.Fatalf("GetByWorkbench for tasks failed: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks assigned to workbench, got %d", len(tasks))
	}
}

func TestIntegration_MultipleEntitiesAssignedToWorkbench(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	seedCommission(t, db, "COMM-001", "Test")
	seedWorkbench(t, db, "BENCH-001", "COMM-001", "shared-bench")

	shipmentRepo := sqlite.NewShipmentRepository(db)
	taskRepo := sqlite.NewTaskRepository(db)

	// Create multiple shipments
	if err := shipmentRepo.Create(ctx, &secondary.ShipmentRecord{ID: "SHIP-001", CommissionID: "COMM-001", Title: "Shipment 1"}); err != nil {
		t.Fatalf("Create shipment 1 failed: %v", err)
	}
	if err := shipmentRepo.Create(ctx, &secondary.ShipmentRecord{ID: "SHIP-002", CommissionID: "COMM-001", Title: "Shipment 2"}); err != nil {
		t.Fatalf("Create shipment 2 failed: %v", err)
	}

	// Create tasks for each shipment
	if err := taskRepo.Create(ctx, &secondary.TaskRecord{ID: "TASK-001", CommissionID: "COMM-001", ShipmentID: "SHIP-001", Title: "Task 1"}); err != nil {
		t.Fatalf("Create task 1 failed: %v", err)
	}
	if err := taskRepo.Create(ctx, &secondary.TaskRecord{ID: "TASK-002", CommissionID: "COMM-001", ShipmentID: "SHIP-002", Title: "Task 2"}); err != nil {
		t.Fatalf("Create task 2 failed: %v", err)
	}

	// Assign same workbench to both shipments
	if err := shipmentRepo.AssignWorkbench(ctx, "SHIP-001", "BENCH-001"); err != nil {
		t.Fatalf("AssignWorkbench to shipment 1 failed: %v", err)
	}
	if err := shipmentRepo.AssignWorkbench(ctx, "SHIP-002", "BENCH-001"); err != nil {
		t.Fatalf("AssignWorkbench to shipment 2 failed: %v", err)
	}
	if err := taskRepo.AssignWorkbenchByShipment(ctx, "SHIP-001", "BENCH-001"); err != nil {
		t.Fatalf("AssignWorkbenchByShipment for shipment 1 failed: %v", err)
	}
	if err := taskRepo.AssignWorkbenchByShipment(ctx, "SHIP-002", "BENCH-001"); err != nil {
		t.Fatalf("AssignWorkbenchByShipment for shipment 2 failed: %v", err)
	}

	// Verify all entities retrievable by workbench
	shipments, err := shipmentRepo.GetByWorkbench(ctx, "BENCH-001")
	if err != nil {
		t.Fatalf("GetByWorkbench for shipments failed: %v", err)
	}
	if len(shipments) != 2 {
		t.Errorf("expected 2 shipments, got %d", len(shipments))
	}

	tasks, err := taskRepo.GetByWorkbench(ctx, "BENCH-001")
	if err != nil {
		t.Fatalf("GetByWorkbench for tasks failed: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

// ============================================================================
// Conclave Workflow Tests
// ============================================================================

func TestIntegration_ConclaveWithTasksQuestionsPlans(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	seedCommission(t, db, "COMM-001", "Test Commission")

	// Create a shipment for the conclave
	_, _ = db.Exec(`INSERT INTO shipments (id, commission_id, title, status) VALUES ('SHIP-001', 'COMM-001', 'Test Shipment', 'open')`)

	conclaveRepo := sqlite.NewConclaveRepository(db)

	// Create conclave linked to shipment
	conclave := &secondary.ConclaveRecord{
		ID:           "CON-001",
		CommissionID: "COMM-001",
		ShipmentID:   "SHIP-001",
		Title:        "Architecture Review",
	}
	if err := conclaveRepo.Create(ctx, conclave); err != nil {
		t.Fatalf("Create conclave failed: %v", err)
	}

	// Tasks and plans link via shipment
	_, _ = db.Exec(`INSERT INTO tasks (id, shipment_id, commission_id, title, status) VALUES ('TASK-001', 'SHIP-001', 'COMM-001', 'Review Task', 'ready')`)
	_, _ = db.Exec(`INSERT INTO plans (id, shipment_id, commission_id, title, status) VALUES ('PLAN-001', 'SHIP-001', 'COMM-001', 'Review Plan', 'draft')`)

	// Verify entities linked to conclave
	tasks, _ := conclaveRepo.GetTasksByConclave(ctx, "CON-001")
	if len(tasks) != 1 {
		t.Errorf("expected 1 task in conclave, got %d", len(tasks))
	}

	plans, _ := conclaveRepo.GetPlansByConclave(ctx, "CON-001")
	if len(plans) != 1 {
		t.Errorf("expected 1 plan in conclave, got %d", len(plans))
	}
}

// ============================================================================
// Tag System Tests
// ============================================================================

func TestIntegration_TagAcrossEntities(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	seedCommission(t, db, "COMM-001", "Test Commission")
	seedTag(t, db, "TAG-001", "urgent")

	taskRepo := sqlite.NewTaskRepository(db)
	tagRepo := sqlite.NewTagRepository(db)

	// Create tasks
	task1 := &secondary.TaskRecord{ID: "TASK-001", CommissionID: "COMM-001", Title: "Task 1"}
	task2 := &secondary.TaskRecord{ID: "TASK-002", CommissionID: "COMM-001", Title: "Task 2"}
	task3 := &secondary.TaskRecord{ID: "TASK-003", CommissionID: "COMM-001", Title: "Task 3"}
	_ = taskRepo.Create(ctx, task1)
	_ = taskRepo.Create(ctx, task2)
	_ = taskRepo.Create(ctx, task3)

	// Tag task1 and task2
	_ = taskRepo.AddTag(ctx, "TASK-001", "TAG-001")
	_ = taskRepo.AddTag(ctx, "TASK-002", "TAG-001")

	// Query tasks by tag
	taggedTasks, err := taskRepo.ListByTag(ctx, "TAG-001")
	if err != nil {
		t.Fatalf("ListByTag failed: %v", err)
	}
	if len(taggedTasks) != 2 {
		t.Errorf("expected 2 tagged tasks, got %d", len(taggedTasks))
	}

	// Verify tag exists
	tag, err := tagRepo.GetByID(ctx, "TAG-001")
	if err != nil {
		t.Fatalf("GetByID tag failed: %v", err)
	}
	if tag.Name != "urgent" {
		t.Errorf("expected tag name 'urgent', got '%s'", tag.Name)
	}

	// Remove all tags from one task
	if err := taskRepo.RemoveTag(ctx, "TASK-001"); err != nil {
		t.Fatalf("RemoveTag failed: %v", err)
	}

	// Verify only one task remains tagged
	taggedTasks, _ = taskRepo.ListByTag(ctx, "TAG-001")
	if len(taggedTasks) != 1 {
		t.Errorf("expected 1 tagged task after removal, got %d", len(taggedTasks))
	}
}

// ============================================================================
// Handoff Continuity Tests
// ============================================================================

func TestIntegration_HandoffWithContext(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	seedCommission(t, db, "COMM-001", "Test")
	seedWorkbench(t, db, "BENCH-001", "COMM-001", "test-bench")

	handoffRepo := sqlite.NewHandoffRepository(db)

	// Create multiple handoffs for workbench
	h1 := &secondary.HandoffRecord{
		ID:                 "HO-001",
		ActiveCommissionID: "COMM-001",
		ActiveWorkbenchID:  "BENCH-001",
		HandoffNote:        "First handoff",
	}
	h2 := &secondary.HandoffRecord{
		ID:                 "HO-002",
		ActiveCommissionID: "COMM-001",
		ActiveWorkbenchID:  "BENCH-001",
		HandoffNote:        "Second handoff",
	}

	if err := handoffRepo.Create(ctx, h1); err != nil {
		t.Fatalf("Create handoff 1 failed: %v", err)
	}
	// Sleep to ensure different created_at timestamps (SQLite DATETIME has second precision)
	time.Sleep(1100 * time.Millisecond)
	if err := handoffRepo.Create(ctx, h2); err != nil {
		t.Fatalf("Create handoff 2 failed: %v", err)
	}

	// Get latest should return most recent
	latest, err := handoffRepo.GetLatestForWorkbench(ctx, "BENCH-001")
	if err != nil {
		t.Fatalf("GetLatestForWorkbench failed: %v", err)
	}
	if latest.ID != "HO-002" {
		t.Errorf("expected latest handoff HO-002, got %s", latest.ID)
	}
}

// ============================================================================
// Note Container Tests
// ============================================================================

func TestIntegration_NotesAcrossContainers(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	seedCommission(t, db, "COMM-001", "Test Commission")

	noteRepo := sqlite.NewNoteRepository(db)
	shipmentRepo := sqlite.NewShipmentRepository(db)
	investigationRepo := sqlite.NewInvestigationRepository(db)
	conclaveRepo := sqlite.NewConclaveRepository(db)
	tomeRepo := sqlite.NewTomeRepository(db)

	// Create containers
	_ = shipmentRepo.Create(ctx, &secondary.ShipmentRecord{ID: "SHIP-001", CommissionID: "COMM-001", Title: "Shipment"})
	_ = investigationRepo.Create(ctx, &secondary.InvestigationRecord{ID: "INV-001", CommissionID: "COMM-001", Title: "Investigation"})
	_ = conclaveRepo.Create(ctx, &secondary.ConclaveRecord{ID: "CON-001", CommissionID: "COMM-001", Title: "Conclave"})
	_ = tomeRepo.Create(ctx, &secondary.TomeRecord{ID: "TOME-001", CommissionID: "COMM-001", Title: "Tome"})

	// Create notes for each container
	shipmentNote := &secondary.NoteRecord{ID: "NOTE-001", CommissionID: "COMM-001", Title: "Shipment Note", ShipmentID: "SHIP-001"}
	invNote := &secondary.NoteRecord{ID: "NOTE-002", CommissionID: "COMM-001", Title: "Investigation Note", InvestigationID: "INV-001"}
	conclaveNote := &secondary.NoteRecord{ID: "NOTE-003", CommissionID: "COMM-001", Title: "Conclave Note", ConclaveID: "CON-001"}
	tomeNote := &secondary.NoteRecord{ID: "NOTE-004", CommissionID: "COMM-001", Title: "Tome Note", TomeID: "TOME-001"}

	_ = noteRepo.Create(ctx, shipmentNote)
	_ = noteRepo.Create(ctx, invNote)
	_ = noteRepo.Create(ctx, conclaveNote)
	_ = noteRepo.Create(ctx, tomeNote)

	// Verify notes by container
	shipmentNotes, _ := noteRepo.GetByContainer(ctx, "shipment", "SHIP-001")
	if len(shipmentNotes) != 1 {
		t.Errorf("expected 1 shipment note, got %d", len(shipmentNotes))
	}

	invNotes, _ := noteRepo.GetByContainer(ctx, "investigation", "INV-001")
	if len(invNotes) != 1 {
		t.Errorf("expected 1 investigation note, got %d", len(invNotes))
	}

	conclaveNotes, _ := noteRepo.GetByContainer(ctx, "conclave", "CON-001")
	if len(conclaveNotes) != 1 {
		t.Errorf("expected 1 conclave note, got %d", len(conclaveNotes))
	}

	tomeNotes, _ := noteRepo.GetByContainer(ctx, "tome", "TOME-001")
	if len(tomeNotes) != 1 {
		t.Errorf("expected 1 tome note, got %d", len(tomeNotes))
	}

	// Verify all notes in commission
	allNotes, _ := noteRepo.List(ctx, secondary.NoteFilters{CommissionID: "COMM-001"})
	if len(allNotes) != 4 {
		t.Errorf("expected 4 notes total, got %d", len(allNotes))
	}
}

// ============================================================================
// Message Conversation Tests
// ============================================================================

func TestIntegration_MessageConversations(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	seedCommission(t, db, "COMM-001", "Test Commission")

	messageRepo := sqlite.NewMessageRepository(db)

	// Create conversation between ORC and IMP
	msg1 := &secondary.MessageRecord{
		ID:           "MSG-001",
		CommissionID: "COMM-001",
		Sender:       "ORC",
		Recipient:    "IMP-001",
		Subject:      "Task Assignment",
		Body:         "Please work on...",
	}
	msg2 := &secondary.MessageRecord{
		ID:           "MSG-002",
		CommissionID: "COMM-001",
		Sender:       "IMP-001",
		Recipient:    "ORC",
		Subject:      "Re: Task Assignment",
		Body:         "I'm on it",
	}
	msg3 := &secondary.MessageRecord{
		ID:           "MSG-003",
		CommissionID: "COMM-001",
		Sender:       "ORC",
		Recipient:    "IMP-002",
		Subject:      "Different IMP",
		Body:         "Different conversation",
	}

	_ = messageRepo.Create(ctx, msg1)
	_ = messageRepo.Create(ctx, msg2)
	_ = messageRepo.Create(ctx, msg3)

	// Get conversation between ORC and IMP-001
	conversation, err := messageRepo.GetConversation(ctx, "ORC", "IMP-001")
	if err != nil {
		t.Fatalf("GetConversation failed: %v", err)
	}
	if len(conversation) != 2 {
		t.Errorf("expected 2 messages in conversation, got %d", len(conversation))
	}

	// Verify unread count for IMP-001
	unread, _ := messageRepo.GetUnreadCount(ctx, "IMP-001")
	if unread != 1 {
		t.Errorf("expected 1 unread for IMP-001, got %d", unread)
	}

	// Mark as read
	_ = messageRepo.MarkRead(ctx, "MSG-001")

	// Verify unread count updated
	unread, _ = messageRepo.GetUnreadCount(ctx, "IMP-001")
	if unread != 0 {
		t.Errorf("expected 0 unread after marking read, got %d", unread)
	}
}

// ============================================================================
// Status Workflow Tests
// ============================================================================

func TestIntegration_EntityStatusWorkflows(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	seedCommission(t, db, "COMM-001", "Test Commission")

	shipmentRepo := sqlite.NewShipmentRepository(db)
	taskRepo := sqlite.NewTaskRepository(db)
	operationRepo := sqlite.NewOperationRepository(db)

	// Create entities
	shipment := &secondary.ShipmentRecord{ID: "SHIP-001", CommissionID: "COMM-001", Title: "Shipment"}
	task := &secondary.TaskRecord{ID: "TASK-001", CommissionID: "COMM-001", Title: "Task"}
	operation := &secondary.OperationRecord{ID: "OP-001", CommissionID: "COMM-001", Title: "Operation"}

	_ = shipmentRepo.Create(ctx, shipment)
	_ = taskRepo.Create(ctx, task)
	_ = operationRepo.Create(ctx, operation)

	// Verify initial status (defaults from table schema)
	s, _ := shipmentRepo.GetByID(ctx, "SHIP-001")
	if s.Status != "active" {
		t.Errorf("expected shipment status 'active', got '%s'", s.Status)
	}

	tsk, _ := taskRepo.GetByID(ctx, "TASK-001")
	if tsk.Status != "ready" {
		t.Errorf("expected task status 'ready', got '%s'", tsk.Status)
	}

	op, _ := operationRepo.GetByID(ctx, "OP-001")
	if op.Status != "ready" {
		t.Errorf("expected operation status 'ready', got '%s'", op.Status)
	}

	// Transition to in_progress
	_ = shipmentRepo.UpdateStatus(ctx, "SHIP-001", "in_progress", false)
	_ = taskRepo.UpdateStatus(ctx, "TASK-001", "in_progress", false, false)
	_ = operationRepo.UpdateStatus(ctx, "OP-001", "in_progress", false)

	// Verify in_progress
	s, _ = shipmentRepo.GetByID(ctx, "SHIP-001")
	if s.Status != "in_progress" {
		t.Errorf("expected shipment status 'in_progress', got '%s'", s.Status)
	}

	// Transition to complete with timestamp
	_ = shipmentRepo.UpdateStatus(ctx, "SHIP-001", "complete", true)
	_ = taskRepo.UpdateStatus(ctx, "TASK-001", "complete", false, true)
	_ = operationRepo.UpdateStatus(ctx, "OP-001", "complete", true)

	// Verify complete with CompletedAt
	s, _ = shipmentRepo.GetByID(ctx, "SHIP-001")
	if s.Status != "complete" {
		t.Errorf("expected shipment status 'complete', got '%s'", s.Status)
	}
	if s.CompletedAt == "" {
		t.Error("expected CompletedAt to be set")
	}
}

// ============================================================================
// Pin System Tests
// ============================================================================

func TestIntegration_PinAcrossEntities(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	seedCommission(t, db, "COMM-001", "Test Commission")

	shipmentRepo := sqlite.NewShipmentRepository(db)
	taskRepo := sqlite.NewTaskRepository(db)
	noteRepo := sqlite.NewNoteRepository(db)

	// Create entities
	_ = shipmentRepo.Create(ctx, &secondary.ShipmentRecord{ID: "SHIP-001", CommissionID: "COMM-001", Title: "Shipment"})
	_ = taskRepo.Create(ctx, &secondary.TaskRecord{ID: "TASK-001", CommissionID: "COMM-001", Title: "Task"})
	_ = noteRepo.Create(ctx, &secondary.NoteRecord{ID: "NOTE-001", CommissionID: "COMM-001", Title: "Note"})

	// Pin entities
	_ = shipmentRepo.Pin(ctx, "SHIP-001")
	_ = taskRepo.Pin(ctx, "TASK-001")
	_ = noteRepo.Pin(ctx, "NOTE-001")

	// Verify pinned
	s, _ := shipmentRepo.GetByID(ctx, "SHIP-001")
	if !s.Pinned {
		t.Error("expected shipment to be pinned")
	}

	tsk, _ := taskRepo.GetByID(ctx, "TASK-001")
	if !tsk.Pinned {
		t.Error("expected task to be pinned")
	}

	note, _ := noteRepo.GetByID(ctx, "NOTE-001")
	if !note.Pinned {
		t.Error("expected note to be pinned")
	}

	// Unpin entities
	_ = shipmentRepo.Unpin(ctx, "SHIP-001")
	_ = taskRepo.Unpin(ctx, "TASK-001")
	_ = noteRepo.Unpin(ctx, "NOTE-001")

	// Verify unpinned
	s, _ = shipmentRepo.GetByID(ctx, "SHIP-001")
	if s.Pinned {
		t.Error("expected shipment to be unpinned")
	}
}

// ============================================================================
// ID Generation Tests
// ============================================================================

func TestIntegration_IDGenerationAcrossRepositories(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	seedCommission(t, db, "COMM-001", "Test Commission")

	shipmentRepo := sqlite.NewShipmentRepository(db)
	taskRepo := sqlite.NewTaskRepository(db)
	planRepo := sqlite.NewPlanRepository(db)
	investigationRepo := sqlite.NewInvestigationRepository(db)

	// Get initial IDs
	shipID, _ := shipmentRepo.GetNextID(ctx)
	taskID, _ := taskRepo.GetNextID(ctx)
	planID, _ := planRepo.GetNextID(ctx)
	invID, _ := investigationRepo.GetNextID(ctx)

	// Verify expected formats
	if shipID != "SHIP-001" {
		t.Errorf("expected SHIP-001, got %s", shipID)
	}
	if taskID != "TASK-001" {
		t.Errorf("expected TASK-001, got %s", taskID)
	}
	if planID != "PLAN-001" {
		t.Errorf("expected PLAN-001, got %s", planID)
	}
	if invID != "INV-001" {
		t.Errorf("expected INV-001, got %s", invID)
	}

	// Create entities
	_ = shipmentRepo.Create(ctx, &secondary.ShipmentRecord{ID: shipID, CommissionID: "COMM-001", Title: "S1"})
	_ = taskRepo.Create(ctx, &secondary.TaskRecord{ID: taskID, CommissionID: "COMM-001", Title: "T1"})

	// Get next IDs
	shipID2, _ := shipmentRepo.GetNextID(ctx)
	taskID2, _ := taskRepo.GetNextID(ctx)

	if shipID2 != "SHIP-002" {
		t.Errorf("expected SHIP-002, got %s", shipID2)
	}
	if taskID2 != "TASK-002" {
		t.Errorf("expected TASK-002, got %s", taskID2)
	}
}
