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

// mockPlanRepository implements secondary.PlanRepository for testing.
type mockPlanRepository struct {
	plans                  map[string]*secondary.PlanRecord
	activePlanForTask      map[string]string // taskID -> planID
	createErr              error
	getErr                 error
	updateErr              error
	deleteErr              error
	listErr                error
	approveErr             error
	commissionExistsResult bool
	commissionExistsErr    error
	taskExistsResult       bool
	taskExistsErr          error
	hasActivePlanResult    bool
	hasActivePlanErr       error
}

func newMockPlanRepository() *mockPlanRepository {
	return &mockPlanRepository{
		plans:                  make(map[string]*secondary.PlanRecord),
		activePlanForTask:      make(map[string]string),
		commissionExistsResult: true,
		taskExistsResult:       true,
		hasActivePlanResult:    false,
	}
}

func (m *mockPlanRepository) Create(ctx context.Context, plan *secondary.PlanRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.plans[plan.ID] = plan
	if plan.TaskID != "" && plan.Status == "draft" {
		m.activePlanForTask[plan.TaskID] = plan.ID
	}
	return nil
}

func (m *mockPlanRepository) GetByID(ctx context.Context, id string) (*secondary.PlanRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if plan, ok := m.plans[id]; ok {
		return plan, nil
	}
	return nil, errors.New("plan not found")
}

func (m *mockPlanRepository) List(ctx context.Context, filters secondary.PlanFilters) ([]*secondary.PlanRecord, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*secondary.PlanRecord
	for _, p := range m.plans {
		if filters.CommissionID != "" && p.CommissionID != filters.CommissionID {
			continue
		}
		if filters.TaskID != "" && p.TaskID != filters.TaskID {
			continue
		}
		if filters.Status != "" && p.Status != filters.Status {
			continue
		}
		result = append(result, p)
	}
	return result, nil
}

func (m *mockPlanRepository) Update(ctx context.Context, plan *secondary.PlanRecord) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if existing, ok := m.plans[plan.ID]; ok {
		if plan.Title != "" {
			existing.Title = plan.Title
		}
		if plan.Description != "" {
			existing.Description = plan.Description
		}
		if plan.Content != "" {
			existing.Content = plan.Content
		}
	}
	return nil
}

func (m *mockPlanRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if plan, ok := m.plans[id]; ok {
		if plan.TaskID != "" {
			delete(m.activePlanForTask, plan.TaskID)
		}
	}
	delete(m.plans, id)
	return nil
}

func (m *mockPlanRepository) Pin(ctx context.Context, id string) error {
	if plan, ok := m.plans[id]; ok {
		plan.Pinned = true
	}
	return nil
}

func (m *mockPlanRepository) Unpin(ctx context.Context, id string) error {
	if plan, ok := m.plans[id]; ok {
		plan.Pinned = false
	}
	return nil
}

func (m *mockPlanRepository) GetNextID(ctx context.Context) (string, error) {
	return "PLAN-001", nil
}

func (m *mockPlanRepository) Approve(ctx context.Context, id string) error {
	if m.approveErr != nil {
		return m.approveErr
	}
	if plan, ok := m.plans[id]; ok {
		plan.Status = "approved"
		plan.ApprovedAt = "2026-01-20T10:00:00Z"
		// Remove from active plan tracking
		if plan.TaskID != "" {
			delete(m.activePlanForTask, plan.TaskID)
		}
	}
	return nil
}

func (m *mockPlanRepository) UpdateStatus(ctx context.Context, id, status string) error {
	if plan, ok := m.plans[id]; ok {
		plan.Status = status
	}
	return nil
}

func (m *mockPlanRepository) GetActivePlanForTask(ctx context.Context, taskID string) (*secondary.PlanRecord, error) {
	if planID, ok := m.activePlanForTask[taskID]; ok {
		if plan, ok := m.plans[planID]; ok {
			return plan, nil
		}
	}
	return nil, nil
}

func (m *mockPlanRepository) HasActivePlanForTask(ctx context.Context, taskID string) (bool, error) {
	if m.hasActivePlanErr != nil {
		return false, m.hasActivePlanErr
	}
	_, exists := m.activePlanForTask[taskID]
	return exists || m.hasActivePlanResult, nil
}

func (m *mockPlanRepository) CommissionExists(ctx context.Context, commissionID string) (bool, error) {
	if m.commissionExistsErr != nil {
		return false, m.commissionExistsErr
	}
	return m.commissionExistsResult, nil
}

func (m *mockPlanRepository) TaskExists(ctx context.Context, taskID string) (bool, error) {
	if m.taskExistsErr != nil {
		return false, m.taskExistsErr
	}
	return m.taskExistsResult, nil
}

// ============================================================================
// Test Helper
// ============================================================================

func newTestPlanService() (*PlanServiceImpl, *mockPlanRepository) {
	planRepo := newMockPlanRepository()
	service := NewPlanService(planRepo)
	return service, planRepo
}

// ============================================================================
// CreatePlan Tests
// ============================================================================

func TestCreatePlan_Success(t *testing.T) {
	service, _ := newTestPlanService()
	ctx := context.Background()

	resp, err := service.CreatePlan(ctx, primary.CreatePlanRequest{
		CommissionID: "COMM-001",
		TaskID:       "TASK-001",
		Title:        "Test Plan",
		Description:  "A test plan",
		Content:      "## Plan Content\n\n- Step 1\n- Step 2",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.PlanID == "" {
		t.Error("expected plan ID to be set")
	}
	if resp.Plan.Title != "Test Plan" {
		t.Errorf("expected title 'Test Plan', got '%s'", resp.Plan.Title)
	}
	if resp.Plan.Status != "draft" {
		t.Errorf("expected status 'draft', got '%s'", resp.Plan.Status)
	}
}

func TestCreatePlan_WithTask(t *testing.T) {
	service, _ := newTestPlanService()
	ctx := context.Background()

	resp, err := service.CreatePlan(ctx, primary.CreatePlanRequest{
		CommissionID: "COMM-001",
		TaskID:       "TASK-001",
		Title:        "Task Plan",
		Description:  "A plan for a task",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Plan.TaskID != "TASK-001" {
		t.Errorf("expected task ID 'TASK-001', got '%s'", resp.Plan.TaskID)
	}
}

func TestCreatePlan_CommissionNotFound(t *testing.T) {
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	planRepo.commissionExistsResult = false

	_, err := service.CreatePlan(ctx, primary.CreatePlanRequest{
		CommissionID: "COMM-NONEXISTENT",
		TaskID:       "TASK-001",
		Title:        "Test Plan",
		Description:  "A test plan",
	})

	if err == nil {
		t.Fatal("expected error for non-existent commission, got nil")
	}
}

func TestCreatePlan_TaskNotFound(t *testing.T) {
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	planRepo.taskExistsResult = false

	_, err := service.CreatePlan(ctx, primary.CreatePlanRequest{
		CommissionID: "COMM-001",
		TaskID:       "TASK-NONEXISTENT",
		Title:        "Test Plan",
		Description:  "A test plan",
	})

	if err == nil {
		t.Fatal("expected error for non-existent task, got nil")
	}
}

func TestCreatePlan_TaskAlreadyHasActivePlan(t *testing.T) {
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	// Create existing active plan
	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:           "PLAN-001",
		CommissionID: "COMM-001",
		TaskID:       "TASK-001",
		Title:        "Existing Plan",
		Status:       "draft",
	}
	planRepo.activePlanForTask["TASK-001"] = "PLAN-001"

	_, err := service.CreatePlan(ctx, primary.CreatePlanRequest{
		CommissionID: "COMM-001",
		TaskID:       "TASK-001",
		Title:        "New Plan",
		Description:  "Should fail",
	})

	if err == nil {
		t.Fatal("expected error for task with existing active plan, got nil")
	}
}

// ============================================================================
// GetPlan Tests
// ============================================================================

func TestGetPlan_Found(t *testing.T) {
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:           "PLAN-001",
		CommissionID: "COMM-001",
		TaskID:       "TASK-001",
		Title:        "Test Plan",
		Status:       "draft",
	}

	plan, err := service.GetPlan(ctx, "PLAN-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if plan.Title != "Test Plan" {
		t.Errorf("expected title 'Test Plan', got '%s'", plan.Title)
	}
}

func TestGetPlan_NotFound(t *testing.T) {
	service, _ := newTestPlanService()
	ctx := context.Background()

	_, err := service.GetPlan(ctx, "PLAN-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent plan, got nil")
	}
}

// ============================================================================
// ListPlans Tests
// ============================================================================

func TestListPlans_FilterByCommission(t *testing.T) {
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:           "PLAN-001",
		CommissionID: "COMM-001",
		TaskID:       "TASK-001",
		Title:        "Plan 1",
		Status:       "draft",
	}
	planRepo.plans["PLAN-002"] = &secondary.PlanRecord{
		ID:           "PLAN-002",
		CommissionID: "COMM-002",
		TaskID:       "TASK-002",
		Title:        "Plan 2",
		Status:       "draft",
	}

	plans, err := service.ListPlans(ctx, primary.PlanFilters{CommissionID: "COMM-001"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(plans) != 1 {
		t.Errorf("expected 1 plan, got %d", len(plans))
	}
}

func TestListPlans_FilterByTask(t *testing.T) {
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:           "PLAN-001",
		CommissionID: "COMM-001",
		TaskID:       "TASK-001",
		Title:        "Task Plan",
		Status:       "draft",
	}
	planRepo.plans["PLAN-002"] = &secondary.PlanRecord{
		ID:           "PLAN-002",
		CommissionID: "COMM-001",
		TaskID:       "TASK-002",
		Title:        "Another Plan",
		Status:       "draft",
	}

	plans, err := service.ListPlans(ctx, primary.PlanFilters{TaskID: "TASK-001"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(plans) != 1 {
		t.Errorf("expected 1 task plan, got %d", len(plans))
	}
}

func TestListPlans_FilterByStatus(t *testing.T) {
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:           "PLAN-001",
		CommissionID: "COMM-001",
		TaskID:       "TASK-001",
		Title:        "Draft Plan",
		Status:       "draft",
	}
	planRepo.plans["PLAN-002"] = &secondary.PlanRecord{
		ID:           "PLAN-002",
		CommissionID: "COMM-001",
		TaskID:       "TASK-002",
		Title:        "Approved Plan",
		Status:       "approved",
	}

	plans, err := service.ListPlans(ctx, primary.PlanFilters{Status: "draft"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(plans) != 1 {
		t.Errorf("expected 1 draft plan, got %d", len(plans))
	}
}

// ============================================================================
// ApprovePlan Tests
// ============================================================================

func TestApprovePlan_Success(t *testing.T) {
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:           "PLAN-001",
		CommissionID: "COMM-001",
		TaskID:       "TASK-001",
		Title:        "Test Plan",
		Status:       "draft",
	}

	err := service.ApprovePlan(ctx, "PLAN-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if planRepo.plans["PLAN-001"].Status != "approved" {
		t.Errorf("expected status 'approved', got '%s'", planRepo.plans["PLAN-001"].Status)
	}
}

func TestApprovePlan_AlreadyApproved(t *testing.T) {
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:           "PLAN-001",
		CommissionID: "COMM-001",
		TaskID:       "TASK-001",
		Title:        "Test Plan",
		Status:       "approved",
	}

	err := service.ApprovePlan(ctx, "PLAN-001")

	if err == nil {
		t.Fatal("expected error for already approved plan, got nil")
	}
}

// ============================================================================
// Pin/Unpin Tests
// ============================================================================

func TestPinPlan(t *testing.T) {
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:           "PLAN-001",
		CommissionID: "COMM-001",
		TaskID:       "TASK-001",
		Title:        "Test Plan",
		Status:       "draft",
		Pinned:       false,
	}

	err := service.PinPlan(ctx, "PLAN-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !planRepo.plans["PLAN-001"].Pinned {
		t.Error("expected plan to be pinned")
	}
}

func TestUnpinPlan(t *testing.T) {
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:           "PLAN-001",
		CommissionID: "COMM-001",
		TaskID:       "TASK-001",
		Title:        "Pinned Plan",
		Status:       "draft",
		Pinned:       true,
	}

	err := service.UnpinPlan(ctx, "PLAN-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if planRepo.plans["PLAN-001"].Pinned {
		t.Error("expected plan to be unpinned")
	}
}

// ============================================================================
// UpdatePlan Tests
// ============================================================================

func TestUpdatePlan_Title(t *testing.T) {
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:           "PLAN-001",
		CommissionID: "COMM-001",
		TaskID:       "TASK-001",
		Title:        "Old Title",
		Description:  "Original description",
		Status:       "draft",
	}

	err := service.UpdatePlan(ctx, primary.UpdatePlanRequest{
		PlanID: "PLAN-001",
		Title:  "New Title",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if planRepo.plans["PLAN-001"].Title != "New Title" {
		t.Errorf("expected title 'New Title', got '%s'", planRepo.plans["PLAN-001"].Title)
	}
}

func TestUpdatePlan_Content(t *testing.T) {
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:           "PLAN-001",
		CommissionID: "COMM-001",
		TaskID:       "TASK-001",
		Title:        "Test Plan",
		Content:      "Original content",
		Status:       "draft",
	}

	err := service.UpdatePlan(ctx, primary.UpdatePlanRequest{
		PlanID:  "PLAN-001",
		Content: "Updated content",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if planRepo.plans["PLAN-001"].Content != "Updated content" {
		t.Errorf("expected content 'Updated content', got '%s'", planRepo.plans["PLAN-001"].Content)
	}
}

// ============================================================================
// DeletePlan Tests
// ============================================================================

func TestDeletePlan_Success(t *testing.T) {
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:           "PLAN-001",
		CommissionID: "COMM-001",
		TaskID:       "TASK-001",
		Title:        "Test Plan",
		Status:       "draft",
	}

	err := service.DeletePlan(ctx, "PLAN-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, exists := planRepo.plans["PLAN-001"]; exists {
		t.Error("expected plan to be deleted")
	}
}

// ============================================================================
// GetTaskActivePlan Tests
// ============================================================================

func TestGetTaskActivePlan_Found(t *testing.T) {
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:           "PLAN-001",
		CommissionID: "COMM-001",
		TaskID:       "TASK-001",
		Title:        "Active Plan",
		Status:       "draft",
	}
	planRepo.activePlanForTask["TASK-001"] = "PLAN-001"

	plan, err := service.GetTaskActivePlan(ctx, "TASK-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if plan == nil {
		t.Fatal("expected plan, got nil")
	}
	if plan.Title != "Active Plan" {
		t.Errorf("expected title 'Active Plan', got '%s'", plan.Title)
	}
}

func TestGetTaskActivePlan_NotFound(t *testing.T) {
	service, _ := newTestPlanService()
	ctx := context.Background()

	plan, err := service.GetTaskActivePlan(ctx, "TASK-NONEXISTENT")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if plan != nil {
		t.Error("expected nil plan for task without active plan")
	}
}
