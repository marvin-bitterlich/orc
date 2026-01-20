package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/example/orc/internal/ports/primary"
)

// MockMissionService for testing
type mockMissionService struct {
	missions map[string]*primary.Mission
}

func newMockMissionService() *mockMissionService {
	return &mockMissionService{
		missions: make(map[string]*primary.Mission),
	}
}

func (m *mockMissionService) GetMission(ctx context.Context, id string) (*primary.Mission, error) {
	mission, ok := m.missions[id]
	if !ok {
		return nil, os.ErrNotExist
	}
	return mission, nil
}

func (m *mockMissionService) CreateMission(ctx context.Context, req primary.CreateMissionRequest) (*primary.CreateMissionResponse, error) {
	return nil, nil
}

func (m *mockMissionService) StartMission(ctx context.Context, req primary.StartMissionRequest) (*primary.StartMissionResponse, error) {
	return nil, nil
}

func (m *mockMissionService) LaunchMission(ctx context.Context, req primary.LaunchMissionRequest) (*primary.LaunchMissionResponse, error) {
	return nil, nil
}

func (m *mockMissionService) ListMissions(ctx context.Context, filters primary.MissionFilters) ([]*primary.Mission, error) {
	return nil, nil
}

func (m *mockMissionService) CompleteMission(ctx context.Context, missionID string) error {
	return nil
}

func (m *mockMissionService) ArchiveMission(ctx context.Context, missionID string) error {
	return nil
}

func (m *mockMissionService) UpdateMission(ctx context.Context, req primary.UpdateMissionRequest) error {
	return nil
}

func (m *mockMissionService) DeleteMission(ctx context.Context, req primary.DeleteMissionRequest) error {
	return nil
}

func (m *mockMissionService) PinMission(ctx context.Context, missionID string) error {
	return nil
}

func (m *mockMissionService) UnpinMission(ctx context.Context, missionID string) error {
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

func (m *mockGroveService) GetGrove(ctx context.Context, groveID string) (*primary.Grove, error) {
	return nil, nil
}

func (m *mockGroveService) GetGroveByPath(ctx context.Context, path string) (*primary.Grove, error) {
	return nil, nil
}

func (m *mockGroveService) ListGroves(ctx context.Context, filters primary.GroveFilters) ([]*primary.Grove, error) {
	groves, ok := m.groves[filters.MissionID]
	if !ok {
		return []*primary.Grove{}, nil
	}
	return groves, nil
}

func (m *mockGroveService) RenameGrove(ctx context.Context, req primary.RenameGroveRequest) error {
	return nil
}

func (m *mockGroveService) UpdateGrovePath(ctx context.Context, groveID, newPath string) error {
	return nil
}

func (m *mockGroveService) DeleteGrove(ctx context.Context, req primary.DeleteGroveRequest) error {
	return nil
}

func TestMissionOrchestrationService_LoadMissionState(t *testing.T) {
	ctx := context.Background()

	missionSvc := newMockMissionService()
	groveSvc := newMockGroveService()

	missionSvc.missions["MISSION-001"] = &primary.Mission{
		ID:    "MISSION-001",
		Title: "Test Mission",
	}

	groveSvc.groves["MISSION-001"] = []*primary.Grove{
		{ID: "GROVE-001", Name: "grove-a", MissionID: "MISSION-001"},
		{ID: "GROVE-002", Name: "grove-b", MissionID: "MISSION-001"},
	}

	svc := NewMissionOrchestrationService(missionSvc, groveSvc)

	state, err := svc.LoadMissionState(ctx, "MISSION-001")
	if err != nil {
		t.Fatalf("LoadMissionState failed: %v", err)
	}

	if state.Mission.ID != "MISSION-001" {
		t.Errorf("expected mission ID MISSION-001, got %s", state.Mission.ID)
	}

	if len(state.Groves) != 2 {
		t.Errorf("expected 2 groves, got %d", len(state.Groves))
	}
}

func TestMissionOrchestrationService_LoadMissionState_NotFound(t *testing.T) {
	ctx := context.Background()

	missionSvc := newMockMissionService()
	groveSvc := newMockGroveService()

	svc := NewMissionOrchestrationService(missionSvc, groveSvc)

	_, err := svc.LoadMissionState(ctx, "MISSION-999")
	if err == nil {
		t.Error("expected error for non-existent mission")
	}
}

func TestMissionOrchestrationService_AnalyzeInfrastructure(t *testing.T) {
	missionSvc := newMockMissionService()
	groveSvc := newMockGroveService()
	svc := NewMissionOrchestrationService(missionSvc, groveSvc)

	state := &MissionState{
		Mission: &primary.Mission{ID: "MISSION-001", Title: "Test"},
		Groves: []*primary.Grove{
			{ID: "GROVE-001", Name: "grove-a", MissionID: "MISSION-001", Path: "/some/path/grove-a"},
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

func TestMissionOrchestrationService_ApplyInfrastructure(t *testing.T) {
	ctx := context.Background()

	missionSvc := newMockMissionService()
	groveSvc := newMockGroveService()
	svc := NewMissionOrchestrationService(missionSvc, groveSvc)

	// Create a temp directory for testing
	tempDir, err := os.MkdirTemp("", "orc-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	workspacePath := filepath.Join(tempDir, "mission-workspace")
	grovesDir := filepath.Join(workspacePath, "groves")

	plan := &InfrastructurePlan{
		WorkspacePath:   workspacePath,
		GrovesDir:       grovesDir,
		CreateWorkspace: true,
		CreateGrovesDir: true,
		GroveActions:    []GroveAction{},
		ConfigWrites:    []ConfigWrite{},
		Cleanups:        []CleanupAction{},
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

func TestMissionOrchestrationService_PlanTmuxSession(t *testing.T) {
	missionSvc := newMockMissionService()
	groveSvc := newMockGroveService()
	svc := NewMissionOrchestrationService(missionSvc, groveSvc)

	// Create a temp directory with a grove
	tempDir, err := os.MkdirTemp("", "orc-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	grovePath := filepath.Join(tempDir, "groves", "grove-a")
	os.MkdirAll(grovePath, 0755)

	state := &MissionState{
		Mission: &primary.Mission{ID: "MISSION-001", Title: "Test"},
		Groves: []*primary.Grove{
			{ID: "GROVE-001", Name: "grove-a", MissionID: "MISSION-001", Path: grovePath},
		},
	}

	plan := svc.PlanTmuxSession(state, tempDir, "orc-MISSION-001", false, nil)

	if plan.SessionName != "orc-MISSION-001" {
		t.Errorf("expected session name orc-MISSION-001, got %s", plan.SessionName)
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

func TestDefaultWorkspacePath(t *testing.T) {
	path, err := DefaultWorkspacePath("MISSION-001")
	if err != nil {
		t.Fatalf("DefaultWorkspacePath failed: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "src", "missions", "MISSION-001")

	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}
