package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// MockCommissionService for testing
type mockCommissionService struct {
	commissions map[string]*primary.Commission
}

func newMockCommissionService() *mockCommissionService {
	return &mockCommissionService{
		commissions: make(map[string]*primary.Commission),
	}
}

func (m *mockCommissionService) GetCommission(ctx context.Context, id string) (*primary.Commission, error) {
	commission, ok := m.commissions[id]
	if !ok {
		return nil, os.ErrNotExist
	}
	return commission, nil
}

func (m *mockCommissionService) CreateCommission(ctx context.Context, req primary.CreateCommissionRequest) (*primary.CreateCommissionResponse, error) {
	return nil, nil
}

func (m *mockCommissionService) StartCommission(ctx context.Context, req primary.StartCommissionRequest) (*primary.StartCommissionResponse, error) {
	return nil, nil
}

func (m *mockCommissionService) LaunchCommission(ctx context.Context, req primary.LaunchCommissionRequest) (*primary.LaunchCommissionResponse, error) {
	return nil, nil
}

func (m *mockCommissionService) ListCommissions(ctx context.Context, filters primary.CommissionFilters) ([]*primary.Commission, error) {
	return nil, nil
}

func (m *mockCommissionService) CompleteCommission(ctx context.Context, commissionID string) error {
	return nil
}

func (m *mockCommissionService) ArchiveCommission(ctx context.Context, commissionID string) error {
	return nil
}

func (m *mockCommissionService) UpdateCommission(ctx context.Context, req primary.UpdateCommissionRequest) error {
	return nil
}

func (m *mockCommissionService) DeleteCommission(ctx context.Context, req primary.DeleteCommissionRequest) error {
	return nil
}

func (m *mockCommissionService) PinCommission(ctx context.Context, commissionID string) error {
	return nil
}

func (m *mockCommissionService) UnpinCommission(ctx context.Context, commissionID string) error {
	return nil
}

// MockGroveService for testing
type mockGroveService struct {
	groves map[string][]*primary.Grove
}

func newMockGroveService() *mockGroveService {
	return &mockGroveService{
		groves: make(map[string][]*primary.Grove),
	}
}

func (m *mockGroveService) CreateGrove(ctx context.Context, req primary.CreateGroveRequest) (*primary.CreateGroveResponse, error) {
	return nil, nil
}

func (m *mockGroveService) OpenGrove(ctx context.Context, req primary.OpenGroveRequest) (*primary.OpenGroveResponse, error) {
	return nil, nil
}

func (m *mockGroveService) GetGrove(ctx context.Context, workbenchID string) (*primary.Grove, error) {
	return nil, nil
}

func (m *mockGroveService) GetGroveByPath(ctx context.Context, path string) (*primary.Grove, error) {
	return nil, nil
}

func (m *mockGroveService) ListGroves(ctx context.Context, filters primary.GroveFilters) ([]*primary.Grove, error) {
	groves, ok := m.groves[filters.CommissionID]
	if !ok {
		return []*primary.Grove{}, nil
	}
	return groves, nil
}

func (m *mockGroveService) RenameGrove(ctx context.Context, req primary.RenameGroveRequest) error {
	return nil
}

func (m *mockGroveService) UpdateGrovePath(ctx context.Context, workbenchID, newPath string) error {
	return nil
}

func (m *mockGroveService) DeleteGrove(ctx context.Context, req primary.DeleteGroveRequest) error {
	return nil
}

func TestCommissionOrchestrationService_LoadCommissionState(t *testing.T) {
	ctx := context.Background()

	commissionSvc := newMockCommissionService()
	groveSvc := newMockGroveService()
	agentProvider := newMockAgentProvider(secondary.AgentTypeORC)

	commissionSvc.commissions["COMM-001"] = &primary.Commission{
		ID:    "COMM-001",
		Title: "Test Commission",
	}

	groveSvc.groves["COMM-001"] = []*primary.Grove{
		{ID: "GROVE-001", Name: "grove-a", CommissionID: "COMM-001"},
		{ID: "GROVE-002", Name: "grove-b", CommissionID: "COMM-001"},
	}

	svc := NewCommissionOrchestrationService(commissionSvc, groveSvc, agentProvider)

	state, err := svc.LoadCommissionState(ctx, "COMM-001")
	if err != nil {
		t.Fatalf("LoadCommissionState failed: %v", err)
	}

	if state.Commission.ID != "COMM-001" {
		t.Errorf("expected commission ID COMM-001, got %s", state.Commission.ID)
	}

	if len(state.Groves) != 2 {
		t.Errorf("expected 2 groves, got %d", len(state.Groves))
	}
}

func TestCommissionOrchestrationService_LoadCommissionState_NotFound(t *testing.T) {
	ctx := context.Background()

	commissionSvc := newMockCommissionService()
	groveSvc := newMockGroveService()
	agentProvider := newMockAgentProvider(secondary.AgentTypeORC)

	svc := NewCommissionOrchestrationService(commissionSvc, groveSvc, agentProvider)

	_, err := svc.LoadCommissionState(ctx, "COMM-999")
	if err == nil {
		t.Error("expected error for non-existent commission")
	}
}

func TestCommissionOrchestrationService_AnalyzeInfrastructure(t *testing.T) {
	commissionSvc := newMockCommissionService()
	groveSvc := newMockGroveService()
	agentProvider := newMockAgentProvider(secondary.AgentTypeORC)
	svc := NewCommissionOrchestrationService(commissionSvc, groveSvc, agentProvider)

	state := &primary.CommissionState{
		Commission: &primary.Commission{ID: "COMM-001", Title: "Test"},
		Groves: []*primary.Grove{
			{ID: "GROVE-001", Name: "grove-a", CommissionID: "COMM-001", Path: "/some/path/grove-a"},
		},
	}

	// Use a temp directory that doesn't exist
	workspacePath := filepath.Join(os.TempDir(), "orc-test-nonexistent")
	defer os.RemoveAll(workspacePath)

	plan := svc.AnalyzeInfrastructure(state, workspacePath)

	if plan.WorkspacePath != workspacePath {
		t.Errorf("expected workspace path %s, got %s", workspacePath, plan.WorkspacePath)
	}

	if !plan.CreateWorkspace {
		t.Error("expected CreateWorkspace to be true for non-existent path")
	}

	if !plan.CreateGrovesDir {
		t.Error("expected CreateGrovesDir to be true for non-existent path")
	}

	if len(plan.GroveActions) != 1 {
		t.Errorf("expected 1 grove action, got %d", len(plan.GroveActions))
	}
}

func TestCommissionOrchestrationService_ApplyInfrastructure(t *testing.T) {
	ctx := context.Background()

	commissionSvc := newMockCommissionService()
	groveSvc := newMockGroveService()
	agentProvider := newMockAgentProvider(secondary.AgentTypeORC)
	svc := NewCommissionOrchestrationService(commissionSvc, groveSvc, agentProvider)

	// Create a temp directory for testing
	tempDir, err := os.MkdirTemp("", "orc-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	workspacePath := filepath.Join(tempDir, "commission-workspace")
	grovesDir := filepath.Join(workspacePath, "groves")

	plan := &primary.InfrastructurePlan{
		WorkspacePath:   workspacePath,
		GrovesDir:       grovesDir,
		CreateWorkspace: true,
		CreateGrovesDir: true,
		GroveActions:    []primary.GroveAction{},
		ConfigWrites:    []primary.ConfigWrite{},
		Cleanups:        []primary.CleanupAction{},
	}

	result := svc.ApplyInfrastructure(ctx, plan)

	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}

	if !result.WorkspaceCreated {
		t.Error("expected WorkspaceCreated to be true")
	}

	if !result.GrovesDirCreated {
		t.Error("expected GrovesDirCreated to be true")
	}

	// Verify directories were created
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		t.Error("workspace directory was not created")
	}

	if _, err := os.Stat(grovesDir); os.IsNotExist(err) {
		t.Error("groves directory was not created")
	}
}

func TestCommissionOrchestrationService_PlanTmuxSession(t *testing.T) {
	commissionSvc := newMockCommissionService()
	groveSvc := newMockGroveService()
	agentProvider := newMockAgentProvider(secondary.AgentTypeORC)
	svc := NewCommissionOrchestrationService(commissionSvc, groveSvc, agentProvider)

	// Create a temp directory with a grove
	tempDir, err := os.MkdirTemp("", "orc-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	grovePath := filepath.Join(tempDir, "groves", "grove-a")
	os.MkdirAll(grovePath, 0755)

	state := &primary.CommissionState{
		Commission: &primary.Commission{ID: "COMM-001", Title: "Test"},
		Groves: []*primary.Grove{
			{ID: "GROVE-001", Name: "grove-a", CommissionID: "COMM-001", Path: grovePath},
		},
	}

	plan := svc.PlanTmuxSession(state, tempDir, "orc-COMM-001", false, nil)

	if plan.SessionName != "orc-COMM-001" {
		t.Errorf("expected session name orc-COMM-001, got %s", plan.SessionName)
	}

	if plan.SessionExists {
		t.Error("expected SessionExists to be false")
	}

	if len(plan.WindowPlans) != 1 {
		t.Errorf("expected 1 window plan, got %d", len(plan.WindowPlans))
	}

	if plan.WindowPlans[0].Action != "create" {
		t.Errorf("expected window action 'create', got %s", plan.WindowPlans[0].Action)
	}
}

func TestCheckLaunchPermission_ORCAllowed(t *testing.T) {
	ctx := context.Background()

	commissionSvc := newMockCommissionService()
	groveSvc := newMockGroveService()
	agentProvider := newMockAgentProvider(secondary.AgentTypeORC)

	svc := NewCommissionOrchestrationService(commissionSvc, groveSvc, agentProvider)

	err := svc.CheckLaunchPermission(ctx)
	if err != nil {
		t.Errorf("expected ORC to be allowed, got error: %v", err)
	}
}

func TestCheckLaunchPermission_IMPDenied(t *testing.T) {
	ctx := context.Background()

	commissionSvc := newMockCommissionService()
	groveSvc := newMockGroveService()
	agentProvider := newMockAgentProvider(secondary.AgentTypeIMP)

	svc := NewCommissionOrchestrationService(commissionSvc, groveSvc, agentProvider)

	err := svc.CheckLaunchPermission(ctx)
	if err == nil {
		t.Error("expected IMP to be denied, got nil error")
	}
	// Verify error message mentions IMPs
	if err != nil && !strings.Contains(err.Error(), "IMP") {
		t.Errorf("expected error to mention IMP, got: %v", err)
	}
}
