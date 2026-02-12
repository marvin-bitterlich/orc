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
	nextID                 string
}

func newMockPlanRepository() *mockPlanRepository {
	return &mockPlanRepository{
		plans:                  make(map[string]*secondary.PlanRecord),
		commissionExistsResult: true,
		taskExistsResult:       true,
		nextID:                 "PLAN-001",
	}
}

func (m *mockPlanRepository) Create(ctx context.Context, plan *secondary.PlanRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.plans[plan.ID] = plan
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
		if filters.TaskID != "" && p.TaskID != filters.TaskID {
			continue
		}
		if filters.CommissionID != "" && p.CommissionID != filters.CommissionID {
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
	return m.nextID, nil
}

func (m *mockPlanRepository) Approve(ctx context.Context, id string) error {
	if m.approveErr != nil {
		return m.approveErr
	}
	if plan, ok := m.plans[id]; ok {
		plan.Status = "approved"
		plan.ApprovedAt = "2026-01-20T10:00:00Z"
	}
	return nil
}

func (m *mockPlanRepository) GetActivePlanForTask(ctx context.Context, taskID string) (*secondary.PlanRecord, error) {
	for _, p := range m.plans {
		if p.TaskID == taskID && p.Status == "draft" {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockPlanRepository) HasActivePlanForTask(ctx context.Context, taskID string) (bool, error) {
	if m.hasActivePlanErr != nil {
		return false, m.hasActivePlanErr
	}
	return m.hasActivePlanResult, nil
}

func (m *mockPlanRepository) UpdateStatus(ctx context.Context, id, status string) error {
	if plan, ok := m.plans[id]; ok {
		plan.Status = status
	}
	return nil
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
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	resp, err := service.CreatePlan(ctx, primary.CreatePlanRequest{
		CommissionID: "COMM-001",
		TaskID:       "TASK-001",
		Title:        "Test Plan",
		Description:  "A test plan",
		Content:      "Plan content here",
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
	if _, ok := planRepo.plans["PLAN-001"]; !ok {
		t.Error("expected plan to be stored in repository")
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
	})

	if err == nil {
		t.Fatal("expected error for non-existent task, got nil")
	}
}

func TestCreatePlan_TaskAlreadyHasActivePlan(t *testing.T) {
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	planRepo.hasActivePlanResult = true

	_, err := service.CreatePlan(ctx, primary.CreatePlanRequest{
		CommissionID: "COMM-001",
		TaskID:       "TASK-001",
		Title:        "Test Plan",
	})

	if err == nil {
		t.Fatal("expected error for task with active plan, got nil")
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
		TaskID:       "TASK-001",
		CommissionID: "COMM-001",
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

func TestListPlans_FilterByTask(t *testing.T) {
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:           "PLAN-001",
		TaskID:       "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Plan 1",
		Status:       "draft",
	}
	planRepo.plans["PLAN-002"] = &secondary.PlanRecord{
		ID:           "PLAN-002",
		TaskID:       "TASK-002",
		CommissionID: "COMM-001",
		Title:        "Plan 2",
		Status:       "draft",
	}

	plans, err := service.ListPlans(ctx, primary.PlanFilters{TaskID: "TASK-001"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(plans) != 1 {
		t.Errorf("expected 1 plan, got %d", len(plans))
	}
}

func TestListPlans_FilterByStatus(t *testing.T) {
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:           "PLAN-001",
		TaskID:       "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Draft Plan",
		Status:       "draft",
	}
	planRepo.plans["PLAN-002"] = &secondary.PlanRecord{
		ID:           "PLAN-002",
		TaskID:       "TASK-002",
		CommissionID: "COMM-001",
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
		TaskID:       "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Draft Plan",
		Status:       "draft",
		Pinned:       false,
	}

	err := service.ApprovePlan(ctx, "PLAN-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if planRepo.plans["PLAN-001"].Status != "approved" {
		t.Errorf("expected status 'approved', got '%s'", planRepo.plans["PLAN-001"].Status)
	}
}

func TestApprovePlan_NotDraft(t *testing.T) {
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:     "PLAN-001",
		Status: "approved",
	}

	err := service.ApprovePlan(ctx, "PLAN-001")

	if err == nil {
		t.Fatal("expected error for approving non-draft plan, got nil")
	}
}

func TestApprovePlan_PinnedBlocked(t *testing.T) {
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:     "PLAN-001",
		Status: "draft",
		Pinned: true,
	}

	err := service.ApprovePlan(ctx, "PLAN-001")

	if err == nil {
		t.Fatal("expected error for approving pinned plan, got nil")
	}
}

func TestApprovePlan_NotFound(t *testing.T) {
	service, _ := newTestPlanService()
	ctx := context.Background()

	err := service.ApprovePlan(ctx, "PLAN-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent plan, got nil")
	}
}

// ============================================================================
// UpdatePlan Tests
// ============================================================================

func TestUpdatePlan_Success(t *testing.T) {
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:    "PLAN-001",
		Title: "Old Title",
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

// ============================================================================
// Pin/Unpin Tests
// ============================================================================

func TestPinPlan(t *testing.T) {
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:     "PLAN-001",
		Pinned: false,
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
		ID:     "PLAN-001",
		Pinned: true,
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
// DeletePlan Tests
// ============================================================================

func TestDeletePlan_Success(t *testing.T) {
	service, planRepo := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:     "PLAN-001",
		Pinned: false,
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
		TaskID:       "TASK-001",
		CommissionID: "COMM-001",
		Title:        "Active Plan",
		Status:       "draft",
	}

	plan, err := service.GetTaskActivePlan(ctx, "TASK-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if plan == nil {
		t.Fatal("expected to find active plan")
	}
	if plan.Title != "Active Plan" {
		t.Errorf("expected title 'Active Plan', got '%s'", plan.Title)
	}
}

func TestGetTaskActivePlan_NoneFound(t *testing.T) {
	service, _ := newTestPlanService()
	ctx := context.Background()

	plan, err := service.GetTaskActivePlan(ctx, "TASK-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if plan != nil {
		t.Error("expected nil plan when no active plan exists")
	}
}
