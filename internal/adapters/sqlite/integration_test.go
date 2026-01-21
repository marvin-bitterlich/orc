package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

// Integration tests verify cross-repository workflows and constraints.

// ============================================================================
// Mission Lifecycle Tests
// ============================================================================

func TestIntegration_MissionWithShipmentsAndTasks(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()

	missionRepo := sqlite.NewMissionRepository(db)
	shipmentRepo := sqlite.NewShipmentRepository(db)
	taskRepo := sqlite.NewTaskRepository(db)

	// Create mission with status pre-populated (required by repo)
	mission := &secondary.MissionRecord{
		ID:     "MISSION-001",
		Title:  "Integration Test Mission",
		Status: "active",
	}
	if err := missionRepo.Create(ctx, mission); err != nil {
		t.Fatalf("Create mission failed: %v", err)
	}

	// Create shipments under mission
	ship1 := &secondary.ShipmentRecord{
		ID:        "SHIP-001",
		MissionID: "MISSION-001",
		Title:     "Shipment 1",
	}
	ship2 := &secondary.ShipmentRecord{
		ID:        "SHIP-002",
		MissionID: "MISSION-001",
		Title:     "Shipment 2",
	}
	if err := shipmentRepo.Create(ctx, ship1); err != nil {
		t.Fatalf("Create shipment 1 failed: %v", err)
	}
	if err := shipmentRepo.Create(ctx, ship2); err != nil {
		t.Fatalf("Create shipment 2 failed: %v", err)
	}

	// Create tasks under shipments
	task1 := &secondary.TaskRecord{
		ID:         "TASK-001",
		MissionID:  "MISSION-001",
		ShipmentID: "SHIP-001",
		Title:      "Task 1",
	}
	task2 := &secondary.TaskRecord{
		ID:         "TASK-002",
		MissionID:  "MISSION-001",
		ShipmentID: "SHIP-001",
		Title:      "Task 2",
	}
	if err := taskRepo.Create(ctx, task1); err != nil {
		t.Fatalf("Create task 1 failed: %v", err)
	}
	if err := taskRepo.Create(ctx, task2); err != nil {
		t.Fatalf("Create task 2 failed: %v", err)
	}

	// Verify mission exists check works
	exists, _ := shipmentRepo.MissionExists(ctx, "MISSION-001")
	if !exists {
		t.Error("expected mission to exist")
	}

	// Verify shipment listing by mission
	shipments, err := shipmentRepo.List(ctx, secondary.ShipmentFilters{MissionID: "MISSION-001"})
	if err != nil {
		t.Fatalf("List shipments failed: %v", err)
	}
	if len(shipments) != 2 {
		t.Errorf("expected 2 shipments, got %d", len(shipments))
	}

	// Verify task listing by shipment
	tasks, err := taskRepo.GetByShipment(ctx, "SHIP-001")
	if err != nil {
		t.Fatalf("GetByShipment failed: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks for SHIP-001, got %d", len(tasks))
	}
}

func TestIntegration_MissionExistsConstraint(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()

	shipmentRepo := sqlite.NewShipmentRepository(db)

	// Verify mission doesn't exist
	exists, _ := shipmentRepo.MissionExists(ctx, "MISSION-999")
	if exists {
		t.Error("expected mission to not exist")
	}
}

// ============================================================================
// Shipment Workflow Tests
// ============================================================================

func TestIntegration_ShipmentWithPlanAndTasks(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()

	seedMission(t, db, "MISSION-001", "Test Mission")

	shipmentRepo := sqlite.NewShipmentRepository(db)
	planRepo := sqlite.NewPlanRepository(db)
	taskRepo := sqlite.NewTaskRepository(db)

	// Create shipment
	shipment := &secondary.ShipmentRecord{
		ID:        "SHIP-001",
		MissionID: "MISSION-001",
		Title:     "Feature Shipment",
	}
	if err := shipmentRepo.Create(ctx, shipment); err != nil {
		t.Fatalf("Create shipment failed: %v", err)
	}

	// Create plan for shipment
	plan := &secondary.PlanRecord{
		ID:         "PLAN-001",
		MissionID:  "MISSION-001",
		ShipmentID: "SHIP-001",
		Title:      "Implementation Plan",
		Content:    "Plan content...",
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
		ID:         "TASK-001",
		MissionID:  "MISSION-001",
		ShipmentID: "SHIP-001",
		Title:      "Implementation Task",
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
// Grove Assignment Tests
// ============================================================================

func TestIntegration_GroveAssignmentPropagation(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()

	seedMission(t, db, "MISSION-001", "Test Mission")
	seedGrove(t, db, "GROVE-001", "MISSION-001", "feature-grove")

	shipmentRepo := sqlite.NewShipmentRepository(db)
	taskRepo := sqlite.NewTaskRepository(db)

	// Create shipment
	shipment := &secondary.ShipmentRecord{
		ID:        "SHIP-001",
		MissionID: "MISSION-001",
		Title:     "Feature Shipment",
	}
	if err := shipmentRepo.Create(ctx, shipment); err != nil {
		t.Fatalf("Create shipment failed: %v", err)
	}

	// Create tasks
	task1 := &secondary.TaskRecord{
		ID:         "TASK-001",
		MissionID:  "MISSION-001",
		ShipmentID: "SHIP-001",
		Title:      "Task 1",
	}
	task2 := &secondary.TaskRecord{
		ID:         "TASK-002",
		MissionID:  "MISSION-001",
		ShipmentID: "SHIP-001",
		Title:      "Task 2",
	}
	_ = taskRepo.Create(ctx, task1)
	_ = taskRepo.Create(ctx, task2)

	// Assign grove to shipment
	if err := shipmentRepo.AssignGrove(ctx, "SHIP-001", "GROVE-001"); err != nil {
		t.Fatalf("AssignGrove to shipment failed: %v", err)
	}

	// Assign grove to tasks via shipment
	if err := taskRepo.AssignGroveByShipment(ctx, "SHIP-001", "GROVE-001"); err != nil {
		t.Fatalf("AssignGroveByShipment failed: %v", err)
	}

	// Verify all tasks have grove assigned
	tasks, _ := taskRepo.GetByGrove(ctx, "GROVE-001")
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks assigned to grove, got %d", len(tasks))
	}

	// Verify shipment has grove assigned
	shipments, _ := shipmentRepo.GetByGrove(ctx, "GROVE-001")
	if len(shipments) != 1 {
		t.Errorf("expected 1 shipment assigned to grove, got %d", len(shipments))
	}
}

func TestIntegration_MultipleEntitiesAssignedToGrove(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()

	seedMission(t, db, "MISSION-001", "Test Mission")
	seedGrove(t, db, "GROVE-001", "MISSION-001", "multi-entity-grove")

	shipmentRepo := sqlite.NewShipmentRepository(db)
	investigationRepo := sqlite.NewInvestigationRepository(db)
	tomeRepo := sqlite.NewTomeRepository(db)

	// Create and assign shipment
	shipment := &secondary.ShipmentRecord{ID: "SHIP-001", MissionID: "MISSION-001", Title: "Shipment"}
	_ = shipmentRepo.Create(ctx, shipment)
	_ = shipmentRepo.AssignGrove(ctx, "SHIP-001", "GROVE-001")

	// Create and assign investigation
	inv := &secondary.InvestigationRecord{ID: "INV-001", MissionID: "MISSION-001", Title: "Investigation"}
	_ = investigationRepo.Create(ctx, inv)
	_ = investigationRepo.AssignGrove(ctx, "INV-001", "GROVE-001")

	// Create and assign tome
	tome := &secondary.TomeRecord{ID: "TOME-001", MissionID: "MISSION-001", Title: "Tome"}
	_ = tomeRepo.Create(ctx, tome)
	_ = tomeRepo.AssignGrove(ctx, "TOME-001", "GROVE-001")

	// Verify all entities are assigned to grove
	shipments, _ := shipmentRepo.GetByGrove(ctx, "GROVE-001")
	if len(shipments) != 1 {
		t.Errorf("expected 1 shipment in grove, got %d", len(shipments))
	}

	investigations, _ := investigationRepo.GetByGrove(ctx, "GROVE-001")
	if len(investigations) != 1 {
		t.Errorf("expected 1 investigation in grove, got %d", len(investigations))
	}

	tomes, _ := tomeRepo.GetByGrove(ctx, "GROVE-001")
	if len(tomes) != 1 {
		t.Errorf("expected 1 tome in grove, got %d", len(tomes))
	}
}

// ============================================================================
// Investigation Flow Tests
// ============================================================================

func TestIntegration_InvestigationWithQuestions(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()

	seedMission(t, db, "MISSION-001", "Test Mission")

	investigationRepo := sqlite.NewInvestigationRepository(db)
	questionRepo := sqlite.NewQuestionRepository(db)

	// Create investigation
	inv := &secondary.InvestigationRecord{
		ID:        "INV-001",
		MissionID: "MISSION-001",
		Title:     "Performance Investigation",
	}
	if err := investigationRepo.Create(ctx, inv); err != nil {
		t.Fatalf("Create investigation failed: %v", err)
	}

	// Create questions for investigation
	q1 := &secondary.QuestionRecord{
		ID:              "Q-001",
		MissionID:       "MISSION-001",
		InvestigationID: "INV-001",
		Title:           "What is the root cause?",
	}
	q2 := &secondary.QuestionRecord{
		ID:              "Q-002",
		MissionID:       "MISSION-001",
		InvestigationID: "INV-001",
		Title:           "What are the affected components?",
	}
	_ = questionRepo.Create(ctx, q1)
	_ = questionRepo.Create(ctx, q2)

	// Verify questions by investigation
	questions, _ := investigationRepo.GetQuestionsByInvestigation(ctx, "INV-001")
	if len(questions) != 2 {
		t.Errorf("expected 2 questions, got %d", len(questions))
	}

	// Answer a question
	if err := questionRepo.Answer(ctx, "Q-001", "Database lock contention"); err != nil {
		t.Fatalf("Answer failed: %v", err)
	}

	// Verify answer
	answered, _ := questionRepo.GetByID(ctx, "Q-001")
	if answered.Answer != "Database lock contention" {
		t.Errorf("expected answer set, got '%s'", answered.Answer)
	}
	if answered.Status != "answered" {
		t.Errorf("expected status 'answered', got '%s'", answered.Status)
	}
}

// ============================================================================
// Conclave Workflow Tests
// ============================================================================

func TestIntegration_ConclaveWithTasksQuestionsPlans(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()

	seedMission(t, db, "MISSION-001", "Test Mission")

	conclaveRepo := sqlite.NewConclaveRepository(db)

	// Create conclave
	conclave := &secondary.ConclaveRecord{
		ID:        "CON-001",
		MissionID: "MISSION-001",
		Title:     "Architecture Review",
	}
	if err := conclaveRepo.Create(ctx, conclave); err != nil {
		t.Fatalf("Create conclave failed: %v", err)
	}

	// Create entities linked to conclave via direct SQL (Create methods don't support conclave_id)
	_, _ = db.Exec(`INSERT INTO tasks (id, mission_id, title, status, conclave_id) VALUES ('TASK-001', 'MISSION-001', 'Review Task', 'ready', 'CON-001')`)
	_, _ = db.Exec(`INSERT INTO questions (id, mission_id, title, status, conclave_id) VALUES ('Q-001', 'MISSION-001', 'Review Question', 'open', 'CON-001')`)
	_, _ = db.Exec(`INSERT INTO plans (id, mission_id, title, status, conclave_id) VALUES ('PLAN-001', 'MISSION-001', 'Review Plan', 'draft', 'CON-001')`)

	// Verify all entities linked to conclave
	tasks, _ := conclaveRepo.GetTasksByConclave(ctx, "CON-001")
	if len(tasks) != 1 {
		t.Errorf("expected 1 task in conclave, got %d", len(tasks))
	}

	questions, _ := conclaveRepo.GetQuestionsByConclave(ctx, "CON-001")
	if len(questions) != 1 {
		t.Errorf("expected 1 question in conclave, got %d", len(questions))
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
	db := setupIntegrationDB(t)
	ctx := context.Background()

	seedMission(t, db, "MISSION-001", "Test Mission")
	seedTag(t, db, "TAG-001", "urgent")

	taskRepo := sqlite.NewTaskRepository(db)
	tagRepo := sqlite.NewTagRepository(db)

	// Create tasks
	task1 := &secondary.TaskRecord{ID: "TASK-001", MissionID: "MISSION-001", Title: "Task 1"}
	task2 := &secondary.TaskRecord{ID: "TASK-002", MissionID: "MISSION-001", Title: "Task 2"}
	task3 := &secondary.TaskRecord{ID: "TASK-003", MissionID: "MISSION-001", Title: "Task 3"}
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
	db := setupIntegrationDB(t)
	ctx := context.Background()

	seedMission(t, db, "MISSION-001", "Test Mission")
	seedGrove(t, db, "GROVE-001", "MISSION-001", "feature-grove")

	handoffRepo := sqlite.NewHandoffRepository(db)

	// Create handoffs with explicit timestamps
	_, _ = db.Exec(`INSERT INTO handoffs (id, handoff_note, active_mission_id, active_grove_id, todos_snapshot, created_at)
		VALUES ('HO-001', 'First session complete', 'MISSION-001', 'GROVE-001', '[{"task":"Task 1"}]', '2024-01-01 10:00:00')`)
	_, _ = db.Exec(`INSERT INTO handoffs (id, handoff_note, active_mission_id, active_grove_id, todos_snapshot, created_at)
		VALUES ('HO-002', 'Second session - continued work', 'MISSION-001', 'GROVE-001', '[{"task":"Task 2"}]', '2024-01-01 11:00:00')`)

	// Get latest handoff
	latest, err := handoffRepo.GetLatest(ctx)
	if err != nil {
		t.Fatalf("GetLatest failed: %v", err)
	}
	if latest.ID != "HO-002" {
		t.Errorf("expected latest HO-002, got %s", latest.ID)
	}
	if latest.ActiveMissionID != "MISSION-001" {
		t.Errorf("expected mission MISSION-001, got %s", latest.ActiveMissionID)
	}
	if latest.ActiveGroveID != "GROVE-001" {
		t.Errorf("expected grove GROVE-001, got %s", latest.ActiveGroveID)
	}

	// Get latest for grove
	latestForGrove, err := handoffRepo.GetLatestForGrove(ctx, "GROVE-001")
	if err != nil {
		t.Fatalf("GetLatestForGrove failed: %v", err)
	}
	if latestForGrove.ID != "HO-002" {
		t.Errorf("expected HO-002 for grove, got %s", latestForGrove.ID)
	}
}

// ============================================================================
// Note Container Tests
// ============================================================================

func TestIntegration_NotesAcrossContainers(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()

	seedMission(t, db, "MISSION-001", "Test Mission")

	noteRepo := sqlite.NewNoteRepository(db)
	shipmentRepo := sqlite.NewShipmentRepository(db)
	investigationRepo := sqlite.NewInvestigationRepository(db)
	conclaveRepo := sqlite.NewConclaveRepository(db)
	tomeRepo := sqlite.NewTomeRepository(db)

	// Create containers
	_ = shipmentRepo.Create(ctx, &secondary.ShipmentRecord{ID: "SHIP-001", MissionID: "MISSION-001", Title: "Shipment"})
	_ = investigationRepo.Create(ctx, &secondary.InvestigationRecord{ID: "INV-001", MissionID: "MISSION-001", Title: "Investigation"})
	_ = conclaveRepo.Create(ctx, &secondary.ConclaveRecord{ID: "CON-001", MissionID: "MISSION-001", Title: "Conclave"})
	_ = tomeRepo.Create(ctx, &secondary.TomeRecord{ID: "TOME-001", MissionID: "MISSION-001", Title: "Tome"})

	// Create notes for each container
	shipmentNote := &secondary.NoteRecord{ID: "NOTE-001", MissionID: "MISSION-001", Title: "Shipment Note", ShipmentID: "SHIP-001"}
	invNote := &secondary.NoteRecord{ID: "NOTE-002", MissionID: "MISSION-001", Title: "Investigation Note", InvestigationID: "INV-001"}
	conclaveNote := &secondary.NoteRecord{ID: "NOTE-003", MissionID: "MISSION-001", Title: "Conclave Note", ConclaveID: "CON-001"}
	tomeNote := &secondary.NoteRecord{ID: "NOTE-004", MissionID: "MISSION-001", Title: "Tome Note", TomeID: "TOME-001"}

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

	// Verify all notes in mission
	allNotes, _ := noteRepo.List(ctx, secondary.NoteFilters{MissionID: "MISSION-001"})
	if len(allNotes) != 4 {
		t.Errorf("expected 4 notes total, got %d", len(allNotes))
	}
}

// ============================================================================
// Message Conversation Tests
// ============================================================================

func TestIntegration_MessageConversations(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()

	seedMission(t, db, "MISSION-001", "Test Mission")

	messageRepo := sqlite.NewMessageRepository(db)

	// Create conversation between ORC and IMP
	msg1 := &secondary.MessageRecord{
		ID:        "MSG-001",
		MissionID: "MISSION-001",
		Sender:    "ORC",
		Recipient: "IMP-001",
		Subject:   "Task Assignment",
		Body:      "Please work on...",
	}
	msg2 := &secondary.MessageRecord{
		ID:        "MSG-002",
		MissionID: "MISSION-001",
		Sender:    "IMP-001",
		Recipient: "ORC",
		Subject:   "Re: Task Assignment",
		Body:      "I'm on it",
	}
	msg3 := &secondary.MessageRecord{
		ID:        "MSG-003",
		MissionID: "MISSION-001",
		Sender:    "ORC",
		Recipient: "IMP-002",
		Subject:   "Different IMP",
		Body:      "Different conversation",
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
	db := setupIntegrationDB(t)
	ctx := context.Background()

	seedMission(t, db, "MISSION-001", "Test Mission")

	shipmentRepo := sqlite.NewShipmentRepository(db)
	taskRepo := sqlite.NewTaskRepository(db)
	operationRepo := sqlite.NewOperationRepository(db)

	// Create entities
	shipment := &secondary.ShipmentRecord{ID: "SHIP-001", MissionID: "MISSION-001", Title: "Shipment"}
	task := &secondary.TaskRecord{ID: "TASK-001", MissionID: "MISSION-001", Title: "Task"}
	operation := &secondary.OperationRecord{ID: "OP-001", MissionID: "MISSION-001", Title: "Operation"}

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
	db := setupIntegrationDB(t)
	ctx := context.Background()

	seedMission(t, db, "MISSION-001", "Test Mission")

	shipmentRepo := sqlite.NewShipmentRepository(db)
	taskRepo := sqlite.NewTaskRepository(db)
	noteRepo := sqlite.NewNoteRepository(db)

	// Create entities
	_ = shipmentRepo.Create(ctx, &secondary.ShipmentRecord{ID: "SHIP-001", MissionID: "MISSION-001", Title: "Shipment"})
	_ = taskRepo.Create(ctx, &secondary.TaskRecord{ID: "TASK-001", MissionID: "MISSION-001", Title: "Task"})
	_ = noteRepo.Create(ctx, &secondary.NoteRecord{ID: "NOTE-001", MissionID: "MISSION-001", Title: "Note"})

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
	db := setupIntegrationDB(t)
	ctx := context.Background()

	seedMission(t, db, "MISSION-001", "Test Mission")

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
	_ = shipmentRepo.Create(ctx, &secondary.ShipmentRecord{ID: shipID, MissionID: "MISSION-001", Title: "S1"})
	_ = taskRepo.Create(ctx, &secondary.TaskRecord{ID: taskID, MissionID: "MISSION-001", Title: "T1"})

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
