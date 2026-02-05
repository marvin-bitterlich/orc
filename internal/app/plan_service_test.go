package app

import (
	"context"
	"errors"
	"fmt"
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
// Additional Mock Implementations for PlanService Dependencies
// ============================================================================

// mockEscalationRepoForPlan is a minimal mock for testing PlanService.
type mockEscalationRepoForPlan struct {
	escalations map[string]*secondary.EscalationRecord
	nextID      int
}

func newMockEscalationRepoForPlan() *mockEscalationRepoForPlan {
	return &mockEscalationRepoForPlan{
		escalations: make(map[string]*secondary.EscalationRecord),
		nextID:      1,
	}
}

func (m *mockEscalationRepoForPlan) Create(ctx context.Context, e *secondary.EscalationRecord) error {
	m.escalations[e.ID] = e
	return nil
}

func (m *mockEscalationRepoForPlan) GetByID(ctx context.Context, id string) (*secondary.EscalationRecord, error) {
	if e, ok := m.escalations[id]; ok {
		return e, nil
	}
	return nil, errors.New("not found")
}

func (m *mockEscalationRepoForPlan) List(ctx context.Context, filters secondary.EscalationFilters) ([]*secondary.EscalationRecord, error) {
	return nil, nil
}

func (m *mockEscalationRepoForPlan) Update(ctx context.Context, e *secondary.EscalationRecord) error {
	return nil
}

func (m *mockEscalationRepoForPlan) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockEscalationRepoForPlan) GetNextID(ctx context.Context) (string, error) {
	id := m.nextID
	m.nextID++
	return fmt.Sprintf("ESC-%03d", id), nil
}

func (m *mockEscalationRepoForPlan) UpdateStatus(ctx context.Context, id, status string, setResolved bool) error {
	return nil
}

func (m *mockEscalationRepoForPlan) Resolve(ctx context.Context, id, resolution, resolvedBy string) error {
	return nil
}

func (m *mockEscalationRepoForPlan) PlanExists(ctx context.Context, planID string) (bool, error) {
	return true, nil
}

func (m *mockEscalationRepoForPlan) TaskExists(ctx context.Context, taskID string) (bool, error) {
	return true, nil
}

func (m *mockEscalationRepoForPlan) ApprovalExists(ctx context.Context, approvalID string) (bool, error) {
	return true, nil
}

// mockWorkbenchRepoForPlan is a minimal mock for testing PlanService.
type mockWorkbenchRepoForPlan struct {
	workbenches map[string]*secondary.WorkbenchRecord
}

func newMockWorkbenchRepoForPlan() *mockWorkbenchRepoForPlan {
	return &mockWorkbenchRepoForPlan{
		workbenches: make(map[string]*secondary.WorkbenchRecord),
	}
}

func (m *mockWorkbenchRepoForPlan) Create(ctx context.Context, w *secondary.WorkbenchRecord) error {
	return nil
}

func (m *mockWorkbenchRepoForPlan) GetByID(ctx context.Context, id string) (*secondary.WorkbenchRecord, error) {
	if w, ok := m.workbenches[id]; ok {
		return w, nil
	}
	return nil, errors.New("not found")
}

func (m *mockWorkbenchRepoForPlan) GetByPath(ctx context.Context, path string) (*secondary.WorkbenchRecord, error) {
	return nil, errors.New("not found")
}

func (m *mockWorkbenchRepoForPlan) GetByWorkshop(ctx context.Context, workshopID string) ([]*secondary.WorkbenchRecord, error) {
	return nil, nil
}

func (m *mockWorkbenchRepoForPlan) List(ctx context.Context, workshopID string) ([]*secondary.WorkbenchRecord, error) {
	return nil, nil
}

func (m *mockWorkbenchRepoForPlan) Update(ctx context.Context, w *secondary.WorkbenchRecord) error {
	return nil
}

func (m *mockWorkbenchRepoForPlan) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockWorkbenchRepoForPlan) GetNextID(ctx context.Context) (string, error) {
	return "BENCH-001", nil
}

func (m *mockWorkbenchRepoForPlan) UpdateStatus(ctx context.Context, id, status string) error {
	return nil
}

func (m *mockWorkbenchRepoForPlan) Rename(ctx context.Context, id, name string) error {
	return nil
}

func (m *mockWorkbenchRepoForPlan) UpdatePath(ctx context.Context, id, path string) error {
	return nil
}

func (m *mockWorkbenchRepoForPlan) UpdateFocusedID(ctx context.Context, id, focusedID string) error {
	return nil
}

func (m *mockWorkbenchRepoForPlan) GetByFocusedID(ctx context.Context, focusedID string) ([]*secondary.WorkbenchRecord, error) {
	return nil, nil
}

func (m *mockWorkbenchRepoForPlan) WorkshopExists(ctx context.Context, workshopID string) (bool, error) {
	return true, nil
}

// mockGatehouseRepoForPlan is a minimal mock for testing PlanService.
type mockGatehouseRepoForPlan struct {
	gatehousesByWorkshop map[string]*secondary.GatehouseRecord
}

func newMockGatehouseRepoForPlan() *mockGatehouseRepoForPlan {
	return &mockGatehouseRepoForPlan{
		gatehousesByWorkshop: make(map[string]*secondary.GatehouseRecord),
	}
}

func (m *mockGatehouseRepoForPlan) Create(ctx context.Context, g *secondary.GatehouseRecord) error {
	return nil
}

func (m *mockGatehouseRepoForPlan) GetByID(ctx context.Context, id string) (*secondary.GatehouseRecord, error) {
	return nil, errors.New("not found")
}

func (m *mockGatehouseRepoForPlan) GetByWorkshop(ctx context.Context, workshopID string) (*secondary.GatehouseRecord, error) {
	if g, ok := m.gatehousesByWorkshop[workshopID]; ok {
		return g, nil
	}
	return nil, errors.New("not found")
}

func (m *mockGatehouseRepoForPlan) List(ctx context.Context, filters secondary.GatehouseFilters) ([]*secondary.GatehouseRecord, error) {
	return nil, nil
}

func (m *mockGatehouseRepoForPlan) Update(ctx context.Context, g *secondary.GatehouseRecord) error {
	return nil
}

func (m *mockGatehouseRepoForPlan) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockGatehouseRepoForPlan) GetNextID(ctx context.Context) (string, error) {
	return "GATE-001", nil
}

func (m *mockGatehouseRepoForPlan) UpdateStatus(ctx context.Context, id, status string) error {
	return nil
}

func (m *mockGatehouseRepoForPlan) WorkshopExists(ctx context.Context, workshopID string) (bool, error) {
	return true, nil
}

func (m *mockGatehouseRepoForPlan) WorkshopHasGatehouse(ctx context.Context, workshopID string) (bool, error) {
	_, ok := m.gatehousesByWorkshop[workshopID]
	return ok, nil
}

func (m *mockGatehouseRepoForPlan) UpdateFocusedID(ctx context.Context, id, focusedID string) error {
	for _, g := range m.gatehousesByWorkshop {
		if g.ID == id {
			g.FocusedID = focusedID
			return nil
		}
	}
	return fmt.Errorf("gatehouse %s not found", id)
}

// mockMessageServiceForPlan is a minimal mock for testing PlanService.
type mockMessageServiceForPlan struct {
	messages []*primary.CreateMessageRequest
}

func newMockMessageServiceForPlan() *mockMessageServiceForPlan {
	return &mockMessageServiceForPlan{
		messages: make([]*primary.CreateMessageRequest, 0),
	}
}

func (m *mockMessageServiceForPlan) CreateMessage(ctx context.Context, req primary.CreateMessageRequest) (*primary.CreateMessageResponse, error) {
	m.messages = append(m.messages, &req)
	return &primary.CreateMessageResponse{MessageID: "MSG-001"}, nil
}

func (m *mockMessageServiceForPlan) GetMessage(ctx context.Context, messageID string) (*primary.Message, error) {
	return nil, errors.New("not found")
}

func (m *mockMessageServiceForPlan) ListMessages(ctx context.Context, recipient string, unreadOnly bool) ([]*primary.Message, error) {
	return nil, nil
}

func (m *mockMessageServiceForPlan) MarkRead(ctx context.Context, messageID string) error {
	return nil
}

func (m *mockMessageServiceForPlan) GetConversation(ctx context.Context, actor1, actor2 string) ([]*primary.Message, error) {
	return nil, nil
}

func (m *mockMessageServiceForPlan) GetUnreadCount(ctx context.Context, recipient string) (int, error) {
	return 0, nil
}

// mockTMuxAdapterForPlan is a minimal mock for testing PlanService.
type mockTMuxAdapterForPlan struct{}

func newMockTMuxAdapterForPlan() *mockTMuxAdapterForPlan {
	return &mockTMuxAdapterForPlan{}
}

func (m *mockTMuxAdapterForPlan) CreateSession(ctx context.Context, name, workingDir string) error {
	return nil
}
func (m *mockTMuxAdapterForPlan) SessionExists(ctx context.Context, name string) bool { return false }
func (m *mockTMuxAdapterForPlan) KillSession(ctx context.Context, name string) error  { return nil }
func (m *mockTMuxAdapterForPlan) GetSessionInfo(ctx context.Context, name string) (string, error) {
	return "", nil
}
func (m *mockTMuxAdapterForPlan) CreateOrcWindow(ctx context.Context, sessionName string, workingDir string) error {
	return nil
}
func (m *mockTMuxAdapterForPlan) CreateWorkbenchWindow(ctx context.Context, sessionName string, windowIndex int, windowName string, workingDir string) error {
	return nil
}
func (m *mockTMuxAdapterForPlan) CreateWorkbenchWindowShell(ctx context.Context, sessionName string, windowIndex int, windowName string, workingDir string) error {
	return nil
}
func (m *mockTMuxAdapterForPlan) WindowExists(ctx context.Context, sessionName string, windowName string) bool {
	return false
}
func (m *mockTMuxAdapterForPlan) KillWindow(ctx context.Context, sessionName string, windowName string) error {
	return nil
}
func (m *mockTMuxAdapterForPlan) SendKeys(ctx context.Context, target, keys string) error { return nil }
func (m *mockTMuxAdapterForPlan) GetPaneCount(ctx context.Context, sessionName, windowName string) int {
	return 0
}
func (m *mockTMuxAdapterForPlan) GetPaneCommand(ctx context.Context, sessionName, windowName string, paneNum int) string {
	return ""
}
func (m *mockTMuxAdapterForPlan) GetPaneStartPath(ctx context.Context, sessionName, windowName string, paneNum int) string {
	return ""
}
func (m *mockTMuxAdapterForPlan) GetPaneStartCommand(ctx context.Context, sessionName, windowName string, paneNum int) string {
	return ""
}
func (m *mockTMuxAdapterForPlan) CapturePaneContent(ctx context.Context, target string, lines int) (string, error) {
	return "", nil
}
func (m *mockTMuxAdapterForPlan) SplitVertical(ctx context.Context, target, workingDir string) error {
	return nil
}
func (m *mockTMuxAdapterForPlan) SplitHorizontal(ctx context.Context, target, workingDir string) error {
	return nil
}
func (m *mockTMuxAdapterForPlan) NudgeSession(ctx context.Context, target, message string) error {
	return nil
}
func (m *mockTMuxAdapterForPlan) AttachInstructions(sessionName string) string { return "" }
func (m *mockTMuxAdapterForPlan) SelectWindow(ctx context.Context, sessionName string, index int) error {
	return nil
}
func (m *mockTMuxAdapterForPlan) RenameWindow(ctx context.Context, target, newName string) error {
	return nil
}
func (m *mockTMuxAdapterForPlan) RespawnPane(ctx context.Context, target string, command ...string) error {
	return nil
}
func (m *mockTMuxAdapterForPlan) RenameSession(ctx context.Context, session, newName string) error {
	return nil
}
func (m *mockTMuxAdapterForPlan) ConfigureStatusBar(ctx context.Context, session string, config secondary.StatusBarConfig) error {
	return nil
}
func (m *mockTMuxAdapterForPlan) DisplayPopup(ctx context.Context, session, command string, config secondary.PopupConfig) error {
	return nil
}
func (m *mockTMuxAdapterForPlan) ConfigureSessionBindings(ctx context.Context, session string, bindings []secondary.KeyBinding) error {
	return nil
}
func (m *mockTMuxAdapterForPlan) ConfigureSessionPopupBindings(ctx context.Context, session string, bindings []secondary.PopupKeyBinding) error {
	return nil
}
func (m *mockTMuxAdapterForPlan) GetCurrentSessionName(ctx context.Context) string { return "" }
func (m *mockTMuxAdapterForPlan) SetEnvironment(ctx context.Context, sessionName, key, value string) error {
	return nil
}
func (m *mockTMuxAdapterForPlan) GetEnvironment(ctx context.Context, sessionName, key string) (string, error) {
	return "", nil
}
func (m *mockTMuxAdapterForPlan) ListSessions(ctx context.Context) ([]string, error) { return nil, nil }
func (m *mockTMuxAdapterForPlan) FindSessionByWorkshopID(ctx context.Context, workshopID string) string {
	return ""
}
func (m *mockTMuxAdapterForPlan) ListWindows(ctx context.Context, sessionName string) ([]string, error) {
	return nil, nil
}
func (m *mockTMuxAdapterForPlan) JoinPane(ctx context.Context, source, target string, vertical bool, size int) error {
	return nil
}

// ============================================================================
// Test Helper
// ============================================================================

func newTestPlanService() (*PlanServiceImpl, *mockPlanRepository, *mockApprovalRepository) {
	planRepo := newMockPlanRepository()
	approvalRepo := newMockApprovalRepository()
	escalationRepo := newMockEscalationRepoForPlan()
	workbenchRepo := newMockWorkbenchRepoForPlan()
	gatehouseRepo := newMockGatehouseRepoForPlan()
	messageService := newMockMessageServiceForPlan()
	tmuxAdapter := newMockTMuxAdapterForPlan()
	service := NewPlanService(planRepo, approvalRepo, escalationRepo, workbenchRepo, gatehouseRepo, messageService, tmuxAdapter)
	return service, planRepo, approvalRepo
}

// ============================================================================
// CreatePlan Tests
// ============================================================================

func TestCreatePlan_Success(t *testing.T) {
	service, _, _ := newTestPlanService()
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
	service, _, _ := newTestPlanService()
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
	service, planRepo, _ := newTestPlanService()
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
	service, planRepo, _ := newTestPlanService()
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
	service, planRepo, _ := newTestPlanService()
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
	service, planRepo, _ := newTestPlanService()
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
	service, _, _ := newTestPlanService()
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
	service, planRepo, _ := newTestPlanService()
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
	service, planRepo, _ := newTestPlanService()
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
	service, planRepo, _ := newTestPlanService()
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

func TestSubmitPlan_Success(t *testing.T) {
	service, planRepo, _ := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:           "PLAN-001",
		CommissionID: "COMM-001",
		TaskID:       "TASK-001",
		Title:        "Test Plan",
		Content:      "Some plan content",
		Status:       "draft",
	}

	err := service.SubmitPlan(ctx, "PLAN-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if planRepo.plans["PLAN-001"].Status != "pending_review" {
		t.Errorf("expected status 'pending_review', got '%s'", planRepo.plans["PLAN-001"].Status)
	}
}

func TestSubmitPlan_NoContent(t *testing.T) {
	service, planRepo, _ := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:           "PLAN-001",
		CommissionID: "COMM-001",
		TaskID:       "TASK-001",
		Title:        "Test Plan",
		Content:      "",
		Status:       "draft",
	}

	err := service.SubmitPlan(ctx, "PLAN-001")

	if err == nil {
		t.Fatal("expected error for plan without content, got nil")
	}
}

func TestSubmitPlan_NotDraft(t *testing.T) {
	service, planRepo, _ := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:           "PLAN-001",
		CommissionID: "COMM-001",
		TaskID:       "TASK-001",
		Title:        "Test Plan",
		Content:      "Some content",
		Status:       "pending_review",
	}

	err := service.SubmitPlan(ctx, "PLAN-001")

	if err == nil {
		t.Fatal("expected error for non-draft plan, got nil")
	}
}

func TestApprovePlan_Success(t *testing.T) {
	service, planRepo, _ := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:           "PLAN-001",
		CommissionID: "COMM-001",
		TaskID:       "TASK-001",
		Title:        "Test Plan",
		Status:       "pending_review",
	}

	approval, err := service.ApprovePlan(ctx, "PLAN-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if approval == nil {
		t.Fatal("expected approval, got nil")
	}
	if approval.PlanID != "PLAN-001" {
		t.Errorf("expected approval planID 'PLAN-001', got '%s'", approval.PlanID)
	}
	if planRepo.plans["PLAN-001"].Status != "approved" {
		t.Errorf("expected status 'approved', got '%s'", planRepo.plans["PLAN-001"].Status)
	}
}

func TestApprovePlan_NotPendingReview(t *testing.T) {
	service, planRepo, _ := newTestPlanService()
	ctx := context.Background()

	planRepo.plans["PLAN-001"] = &secondary.PlanRecord{
		ID:           "PLAN-001",
		CommissionID: "COMM-001",
		TaskID:       "TASK-001",
		Title:        "Test Plan",
		Status:       "draft",
	}

	_, err := service.ApprovePlan(ctx, "PLAN-001")

	if err == nil {
		t.Fatal("expected error for plan not in pending_review status, got nil")
	}
}

// ============================================================================
// Pin/Unpin Tests
// ============================================================================

func TestPinPlan(t *testing.T) {
	service, planRepo, _ := newTestPlanService()
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
	service, planRepo, _ := newTestPlanService()
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
	service, planRepo, _ := newTestPlanService()
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
	service, planRepo, _ := newTestPlanService()
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
	service, planRepo, _ := newTestPlanService()
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
	service, planRepo, _ := newTestPlanService()
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
	service, _, _ := newTestPlanService()
	ctx := context.Background()

	plan, err := service.GetTaskActivePlan(ctx, "TASK-NONEXISTENT")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if plan != nil {
		t.Error("expected nil plan for task without active plan")
	}
}
