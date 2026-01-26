package app

import (
	"context"
	"errors"
	"testing"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// ============================================================================
// Mock Implementations
// ============================================================================

// mockTaskRepository implements secondary.TaskRepository for testing.
type mockTaskRepository struct {
	tasks                  map[string]*secondary.TaskRecord
	tags                   map[string]*secondary.TagRecord // taskID -> tag
	createErr              error
	getErr                 error
	updateErr              error
	deleteErr              error
	listErr                error
	updateStatusErr        error
	claimErr               error
	commissionExistsResult bool
	commissionExistsErr    error
	shipmentExistsResult   bool
	shipmentExistsErr      error
}

func newMockTaskRepository() *mockTaskRepository {
	return &mockTaskRepository{
		tasks:                  make(map[string]*secondary.TaskRecord),
		tags:                   make(map[string]*secondary.TagRecord),
		commissionExistsResult: true,
		shipmentExistsResult:   true,
	}
}

func (m *mockTaskRepository) Create(ctx context.Context, task *secondary.TaskRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskRepository) GetByID(ctx context.Context, id string) (*secondary.TaskRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if task, ok := m.tasks[id]; ok {
		return task, nil
	}
	return nil, errors.New("task not found")
}

func (m *mockTaskRepository) List(ctx context.Context, filters secondary.TaskFilters) ([]*secondary.TaskRecord, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*secondary.TaskRecord
	for _, t := range m.tasks {
		if filters.ShipmentID != "" && t.ShipmentID != filters.ShipmentID {
			continue
		}
		if filters.CommissionID != "" && t.CommissionID != filters.CommissionID {
			continue
		}
		if filters.Status != "" && t.Status != filters.Status {
			continue
		}
		result = append(result, t)
	}
	return result, nil
}

func (m *mockTaskRepository) Update(ctx context.Context, task *secondary.TaskRecord) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if existing, ok := m.tasks[task.ID]; ok {
		if task.Title != "" {
			existing.Title = task.Title
		}
		if task.Description != "" {
			existing.Description = task.Description
		}
	}
	return nil
}

func (m *mockTaskRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.tasks, id)
	return nil
}

func (m *mockTaskRepository) Pin(ctx context.Context, id string) error {
	if task, ok := m.tasks[id]; ok {
		task.Pinned = true
	}
	return nil
}

func (m *mockTaskRepository) Unpin(ctx context.Context, id string) error {
	if task, ok := m.tasks[id]; ok {
		task.Pinned = false
	}
	return nil
}

func (m *mockTaskRepository) GetNextID(ctx context.Context) (string, error) {
	return "TASK-001", nil
}

func (m *mockTaskRepository) GetByWorkbench(ctx context.Context, workbenchID string) ([]*secondary.TaskRecord, error) {
	var result []*secondary.TaskRecord
	for _, t := range m.tasks {
		if t.AssignedWorkbenchID == workbenchID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockTaskRepository) GetByShipment(ctx context.Context, shipmentID string) ([]*secondary.TaskRecord, error) {
	var result []*secondary.TaskRecord
	for _, t := range m.tasks {
		if t.ShipmentID == shipmentID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockTaskRepository) GetByInvestigation(ctx context.Context, investigationID string) ([]*secondary.TaskRecord, error) {
	var result []*secondary.TaskRecord
	for _, t := range m.tasks {
		if t.InvestigationID == investigationID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockTaskRepository) UpdateStatus(ctx context.Context, id, status string, setClaimed, setCompleted bool) error {
	if m.updateStatusErr != nil {
		return m.updateStatusErr
	}
	if task, ok := m.tasks[id]; ok {
		task.Status = status
		if setClaimed {
			task.ClaimedAt = "2026-01-20T10:00:00Z"
		}
		if setCompleted {
			task.CompletedAt = "2026-01-20T10:00:00Z"
		}
	}
	return nil
}

func (m *mockTaskRepository) Claim(ctx context.Context, id, workbenchID string) error {
	if m.claimErr != nil {
		return m.claimErr
	}
	if task, ok := m.tasks[id]; ok {
		task.AssignedWorkbenchID = workbenchID
		task.Status = "in_progress"
		task.ClaimedAt = "2026-01-20T10:00:00Z"
	}
	return nil
}

func (m *mockTaskRepository) AssignWorkbenchByShipment(ctx context.Context, shipmentID, workbenchID string) error {
	return nil
}

func (m *mockTaskRepository) CommissionExists(ctx context.Context, commissionID string) (bool, error) {
	if m.commissionExistsErr != nil {
		return false, m.commissionExistsErr
	}
	return m.commissionExistsResult, nil
}

func (m *mockTaskRepository) ShipmentExists(ctx context.Context, shipmentID string) (bool, error) {
	if m.shipmentExistsErr != nil {
		return false, m.shipmentExistsErr
	}
	return m.shipmentExistsResult, nil
}

func (m *mockTaskRepository) TomeExists(ctx context.Context, tomeID string) (bool, error) {
	return true, nil
}

func (m *mockTaskRepository) ConclaveExists(ctx context.Context, conclaveID string) (bool, error) {
	return true, nil
}

func (m *mockTaskRepository) GetTag(ctx context.Context, taskID string) (*secondary.TagRecord, error) {
	if tag, ok := m.tags[taskID]; ok {
		return tag, nil
	}
	return nil, nil
}

func (m *mockTaskRepository) AddTag(ctx context.Context, taskID, tagID string) error {
	// In real impl this would add the tag association
	return nil
}

func (m *mockTaskRepository) RemoveTag(ctx context.Context, taskID string) error {
	delete(m.tags, taskID)
	return nil
}

func (m *mockTaskRepository) ListByTag(ctx context.Context, tagID string) ([]*secondary.TaskRecord, error) {
	// Simplified implementation
	var result []*secondary.TaskRecord
	for taskID, tag := range m.tags {
		if tag.ID == tagID {
			if task, ok := m.tasks[taskID]; ok {
				result = append(result, task)
			}
		}
	}
	return result, nil
}

func (m *mockTaskRepository) GetNextEntityTagID(ctx context.Context) (string, error) {
	return "ENTITY-TAG-001", nil
}

// mockTagRepositoryForTask implements minimal TagRepository for task tests.
type mockTagRepositoryForTask struct {
	tags map[string]*secondary.TagRecord
}

func newMockTagRepositoryForTask() *mockTagRepositoryForTask {
	return &mockTagRepositoryForTask{
		tags: make(map[string]*secondary.TagRecord),
	}
}

func (m *mockTagRepositoryForTask) Create(ctx context.Context, tag *secondary.TagRecord) error {
	m.tags[tag.ID] = tag
	return nil
}

func (m *mockTagRepositoryForTask) GetByID(ctx context.Context, id string) (*secondary.TagRecord, error) {
	if tag, ok := m.tags[id]; ok {
		return tag, nil
	}
	return nil, errors.New("tag not found")
}

func (m *mockTagRepositoryForTask) GetByName(ctx context.Context, name string) (*secondary.TagRecord, error) {
	for _, tag := range m.tags {
		if tag.Name == name {
			return tag, nil
		}
	}
	return nil, errors.New("tag not found")
}

func (m *mockTagRepositoryForTask) List(ctx context.Context) ([]*secondary.TagRecord, error) {
	var result []*secondary.TagRecord
	for _, tag := range m.tags {
		result = append(result, tag)
	}
	return result, nil
}

func (m *mockTagRepositoryForTask) Delete(ctx context.Context, id string) error {
	delete(m.tags, id)
	return nil
}

func (m *mockTagRepositoryForTask) GetNextID(ctx context.Context) (string, error) {
	return "TAG-001", nil
}

func (m *mockTagRepositoryForTask) GetEntityTag(ctx context.Context, entityID, entityType string) (*secondary.TagRecord, error) {
	return nil, nil
}

// ============================================================================
// Test Helper
// ============================================================================

func newTestTaskService() (*TaskServiceImpl, *mockTaskRepository, *mockTagRepositoryForTask) {
	taskRepo := newMockTaskRepository()
	tagRepo := newMockTagRepositoryForTask()
	service := NewTaskService(taskRepo, tagRepo)
	return service, taskRepo, tagRepo
}

// ============================================================================
// CreateTask Tests
// ============================================================================

func TestCreateTask_Success(t *testing.T) {
	service, _, _ := newTestTaskService()
	ctx := context.Background()

	resp, err := service.CreateTask(ctx, primary.CreateTaskRequest{
		CommissionID: "COMM-001",
		Title:        "Test Task",
		Description:  "A test task",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.TaskID == "" {
		t.Error("expected task ID to be set")
	}
	if resp.Task.Title != "Test Task" {
		t.Errorf("expected title 'Test Task', got '%s'", resp.Task.Title)
	}
	if resp.Task.Status != "ready" {
		t.Errorf("expected status 'ready', got '%s'", resp.Task.Status)
	}
}

func TestCreateTask_WithShipment(t *testing.T) {
	service, _, _ := newTestTaskService()
	ctx := context.Background()

	resp, err := service.CreateTask(ctx, primary.CreateTaskRequest{
		CommissionID: "COMM-001",
		ShipmentID:   "SHIPMENT-001",
		Title:        "Shipment Task",
		Description:  "A task for a shipment",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Task.ShipmentID != "SHIPMENT-001" {
		t.Errorf("expected shipment ID 'SHIPMENT-001', got '%s'", resp.Task.ShipmentID)
	}
}

func TestCreateTask_CommissionNotFound(t *testing.T) {
	service, taskRepo, _ := newTestTaskService()
	ctx := context.Background()

	taskRepo.commissionExistsResult = false

	_, err := service.CreateTask(ctx, primary.CreateTaskRequest{
		CommissionID: "COMM-NONEXISTENT",
		Title:        "Test Task",
		Description:  "A test task",
	})

	if err == nil {
		t.Fatal("expected error for non-existent commission, got nil")
	}
}

func TestCreateTask_ShipmentNotFound(t *testing.T) {
	service, taskRepo, _ := newTestTaskService()
	ctx := context.Background()

	taskRepo.shipmentExistsResult = false

	_, err := service.CreateTask(ctx, primary.CreateTaskRequest{
		CommissionID: "COMM-001",
		ShipmentID:   "SHIPMENT-NONEXISTENT",
		Title:        "Test Task",
		Description:  "A test task",
	})

	if err == nil {
		t.Fatal("expected error for non-existent shipment, got nil")
	}
}

// ============================================================================
// GetTask Tests
// ============================================================================

func TestGetTask_Found(t *testing.T) {
	service, taskRepo, _ := newTestTaskService()
	ctx := context.Background()

	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{
		ID:           "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Test Task",
		Status:       "ready",
	}

	task, err := service.GetTask(ctx, "TASK-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if task.Title != "Test Task" {
		t.Errorf("expected title 'Test Task', got '%s'", task.Title)
	}
}

func TestGetTask_NotFound(t *testing.T) {
	service, _, _ := newTestTaskService()
	ctx := context.Background()

	_, err := service.GetTask(ctx, "TASK-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent task, got nil")
	}
}

func TestGetTask_WithTag(t *testing.T) {
	service, taskRepo, _ := newTestTaskService()
	ctx := context.Background()

	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{
		ID:           "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Tagged Task",
		Status:       "ready",
	}
	taskRepo.tags["TASK-001"] = &secondary.TagRecord{
		ID:   "TAG-001",
		Name: "urgent",
	}

	task, err := service.GetTask(ctx, "TASK-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if task.Tag == nil {
		t.Fatal("expected task to have tag")
	}
	if task.Tag.Name != "urgent" {
		t.Errorf("expected tag name 'urgent', got '%s'", task.Tag.Name)
	}
}

// ============================================================================
// ListTasks Tests
// ============================================================================

func TestListTasks_FilterByCommission(t *testing.T) {
	service, taskRepo, _ := newTestTaskService()
	ctx := context.Background()

	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{
		ID:           "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Task 1",
		Status:       "ready",
	}
	taskRepo.tasks["TASK-002"] = &secondary.TaskRecord{
		ID:           "TASK-002",
		CommissionID: "COMM-002",
		Title:        "Task 2",
		Status:       "ready",
	}

	tasks, err := service.ListTasks(ctx, primary.TaskFilters{CommissionID: "COMM-001"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}
}

func TestListTasks_FilterByStatus(t *testing.T) {
	service, taskRepo, _ := newTestTaskService()
	ctx := context.Background()

	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{
		ID:           "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Ready Task",
		Status:       "ready",
	}
	taskRepo.tasks["TASK-002"] = &secondary.TaskRecord{
		ID:           "TASK-002",
		CommissionID: "COMM-001",
		Title:        "In Progress Task",
		Status:       "in_progress",
	}

	tasks, err := service.ListTasks(ctx, primary.TaskFilters{Status: "ready"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 ready task, got %d", len(tasks))
	}
}

// ============================================================================
// ClaimTask Tests
// ============================================================================

func TestClaimTask_Success(t *testing.T) {
	service, taskRepo, _ := newTestTaskService()
	ctx := context.Background()

	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{
		ID:           "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Test Task",
		Status:       "ready",
	}

	err := service.ClaimTask(ctx, primary.ClaimTaskRequest{
		TaskID:      "TASK-001",
		WorkbenchID: "BENCH-001",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if taskRepo.tasks["TASK-001"].AssignedWorkbenchID != "BENCH-001" {
		t.Errorf("expected workbench ID 'BENCH-001', got '%s'", taskRepo.tasks["TASK-001"].AssignedWorkbenchID)
	}
}

func TestClaimTask_TaskNotFound(t *testing.T) {
	service, _, _ := newTestTaskService()
	ctx := context.Background()

	err := service.ClaimTask(ctx, primary.ClaimTaskRequest{
		TaskID:      "TASK-NONEXISTENT",
		WorkbenchID: "BENCH-001",
	})

	if err == nil {
		t.Fatal("expected error for non-existent task, got nil")
	}
}

// ============================================================================
// CompleteTask Tests
// ============================================================================

func TestCompleteTask_UnpinnedAllowed(t *testing.T) {
	service, taskRepo, _ := newTestTaskService()
	ctx := context.Background()

	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{
		ID:           "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Test Task",
		Status:       "in_progress",
		Pinned:       false,
	}

	err := service.CompleteTask(ctx, "TASK-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if taskRepo.tasks["TASK-001"].Status != "complete" {
		t.Errorf("expected status 'complete', got '%s'", taskRepo.tasks["TASK-001"].Status)
	}
}

func TestCompleteTask_PinnedBlocked(t *testing.T) {
	service, taskRepo, _ := newTestTaskService()
	ctx := context.Background()

	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{
		ID:           "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Pinned Task",
		Status:       "in_progress",
		Pinned:       true,
	}

	err := service.CompleteTask(ctx, "TASK-001")

	if err == nil {
		t.Fatal("expected error for completing pinned task, got nil")
	}
}

func TestCompleteTask_NotFound(t *testing.T) {
	service, _, _ := newTestTaskService()
	ctx := context.Background()

	err := service.CompleteTask(ctx, "TASK-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent task, got nil")
	}
}

// ============================================================================
// PauseTask Tests
// ============================================================================

func TestPauseTask_InProgressAllowed(t *testing.T) {
	service, taskRepo, _ := newTestTaskService()
	ctx := context.Background()

	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{
		ID:           "TASK-001",
		CommissionID: "COMM-001",
		Title:        "In Progress Task",
		Status:       "in_progress",
	}

	err := service.PauseTask(ctx, "TASK-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if taskRepo.tasks["TASK-001"].Status != "paused" {
		t.Errorf("expected status 'paused', got '%s'", taskRepo.tasks["TASK-001"].Status)
	}
}

func TestPauseTask_NotInProgressBlocked(t *testing.T) {
	service, taskRepo, _ := newTestTaskService()
	ctx := context.Background()

	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{
		ID:           "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Ready Task",
		Status:       "ready",
	}

	err := service.PauseTask(ctx, "TASK-001")

	if err == nil {
		t.Fatal("expected error for pausing non-in_progress task, got nil")
	}
}

// ============================================================================
// ResumeTask Tests
// ============================================================================

func TestResumeTask_PausedAllowed(t *testing.T) {
	service, taskRepo, _ := newTestTaskService()
	ctx := context.Background()

	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{
		ID:           "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Paused Task",
		Status:       "paused",
	}

	err := service.ResumeTask(ctx, "TASK-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if taskRepo.tasks["TASK-001"].Status != "in_progress" {
		t.Errorf("expected status 'in_progress', got '%s'", taskRepo.tasks["TASK-001"].Status)
	}
}

func TestResumeTask_NotPausedBlocked(t *testing.T) {
	service, taskRepo, _ := newTestTaskService()
	ctx := context.Background()

	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{
		ID:           "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Ready Task",
		Status:       "ready",
	}

	err := service.ResumeTask(ctx, "TASK-001")

	if err == nil {
		t.Fatal("expected error for resuming non-paused task, got nil")
	}
}

// ============================================================================
// Pin/Unpin Tests
// ============================================================================

func TestPinTask(t *testing.T) {
	service, taskRepo, _ := newTestTaskService()
	ctx := context.Background()

	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{
		ID:           "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Test Task",
		Status:       "ready",
		Pinned:       false,
	}

	err := service.PinTask(ctx, "TASK-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !taskRepo.tasks["TASK-001"].Pinned {
		t.Error("expected task to be pinned")
	}
}

func TestUnpinTask(t *testing.T) {
	service, taskRepo, _ := newTestTaskService()
	ctx := context.Background()

	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{
		ID:           "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Pinned Task",
		Status:       "ready",
		Pinned:       true,
	}

	err := service.UnpinTask(ctx, "TASK-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if taskRepo.tasks["TASK-001"].Pinned {
		t.Error("expected task to be unpinned")
	}
}

// ============================================================================
// TagTask Tests
// ============================================================================

func TestTagTask_Success(t *testing.T) {
	service, taskRepo, tagRepo := newTestTaskService()
	ctx := context.Background()

	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{
		ID:           "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Test Task",
		Status:       "ready",
	}
	tagRepo.tags["TAG-001"] = &secondary.TagRecord{
		ID:   "TAG-001",
		Name: "urgent",
	}

	err := service.TagTask(ctx, "TASK-001", "urgent")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestTagTask_TaskNotFound(t *testing.T) {
	service, _, tagRepo := newTestTaskService()
	ctx := context.Background()

	tagRepo.tags["TAG-001"] = &secondary.TagRecord{
		ID:   "TAG-001",
		Name: "urgent",
	}

	err := service.TagTask(ctx, "TASK-NONEXISTENT", "urgent")

	if err == nil {
		t.Fatal("expected error for non-existent task, got nil")
	}
}

func TestTagTask_TagNotFound(t *testing.T) {
	service, taskRepo, _ := newTestTaskService()
	ctx := context.Background()

	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{
		ID:           "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Test Task",
		Status:       "ready",
	}

	err := service.TagTask(ctx, "TASK-001", "nonexistent-tag")

	if err == nil {
		t.Fatal("expected error for non-existent tag, got nil")
	}
}

func TestTagTask_AlreadyTagged(t *testing.T) {
	service, taskRepo, tagRepo := newTestTaskService()
	ctx := context.Background()

	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{
		ID:           "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Test Task",
		Status:       "ready",
	}
	taskRepo.tags["TASK-001"] = &secondary.TagRecord{
		ID:   "TAG-001",
		Name: "existing-tag",
	}
	tagRepo.tags["TAG-002"] = &secondary.TagRecord{
		ID:   "TAG-002",
		Name: "new-tag",
	}

	err := service.TagTask(ctx, "TASK-001", "new-tag")

	if err == nil {
		t.Fatal("expected error for already tagged task, got nil")
	}
}

// ============================================================================
// UntagTask Tests
// ============================================================================

func TestUntagTask_Success(t *testing.T) {
	service, taskRepo, _ := newTestTaskService()
	ctx := context.Background()

	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{
		ID:           "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Test Task",
		Status:       "ready",
	}
	taskRepo.tags["TASK-001"] = &secondary.TagRecord{
		ID:   "TAG-001",
		Name: "urgent",
	}

	err := service.UntagTask(ctx, "TASK-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUntagTask_NoTag(t *testing.T) {
	service, taskRepo, _ := newTestTaskService()
	ctx := context.Background()

	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{
		ID:           "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Test Task",
		Status:       "ready",
	}

	err := service.UntagTask(ctx, "TASK-001")

	if err == nil {
		t.Fatal("expected error for task without tag, got nil")
	}
}

// ============================================================================
// GetTasksByGrove Tests
// ============================================================================

func TestGetTasksByWorkbench_Success(t *testing.T) {
	service, taskRepo, _ := newTestTaskService()
	ctx := context.Background()

	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{
		ID:                  "TASK-001",
		CommissionID:        "COMM-001",
		Title:               "Assigned Task",
		Status:              "in_progress",
		AssignedWorkbenchID: "BENCH-001",
	}
	taskRepo.tasks["TASK-002"] = &secondary.TaskRecord{
		ID:           "TASK-002",
		CommissionID: "COMM-001",
		Title:        "Unassigned Task",
		Status:       "ready",
	}

	tasks, err := service.GetTasksByWorkbench(ctx, "BENCH-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}
}

// ============================================================================
// DiscoverTasks Tests
// ============================================================================

func TestDiscoverTasks_FindsReadyTasks(t *testing.T) {
	service, taskRepo, _ := newTestTaskService()
	ctx := context.Background()

	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{
		ID:                  "TASK-001",
		CommissionID:        "COMM-001",
		Title:               "Ready Task",
		Status:              "ready",
		AssignedWorkbenchID: "BENCH-001",
	}
	taskRepo.tasks["TASK-002"] = &secondary.TaskRecord{
		ID:                  "TASK-002",
		CommissionID:        "COMM-001",
		Title:               "In Progress Task",
		Status:              "in_progress",
		AssignedWorkbenchID: "BENCH-001",
	}

	tasks, err := service.DiscoverTasks(ctx, "BENCH-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 ready task, got %d", len(tasks))
	}
}

// ============================================================================
// UpdateTask Tests
// ============================================================================

func TestUpdateTask_Title(t *testing.T) {
	service, taskRepo, _ := newTestTaskService()
	ctx := context.Background()

	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{
		ID:           "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Old Title",
		Description:  "Original description",
		Status:       "ready",
	}

	err := service.UpdateTask(ctx, primary.UpdateTaskRequest{
		TaskID: "TASK-001",
		Title:  "New Title",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if taskRepo.tasks["TASK-001"].Title != "New Title" {
		t.Errorf("expected title 'New Title', got '%s'", taskRepo.tasks["TASK-001"].Title)
	}
}

// ============================================================================
// DeleteTask Tests
// ============================================================================

func TestDeleteTask_Success(t *testing.T) {
	service, taskRepo, _ := newTestTaskService()
	ctx := context.Background()

	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{
		ID:           "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Test Task",
		Status:       "ready",
	}

	err := service.DeleteTask(ctx, "TASK-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, exists := taskRepo.tasks["TASK-001"]; exists {
		t.Error("expected task to be deleted")
	}
}
