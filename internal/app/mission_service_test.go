package app

import (
	"context"
	"errors"
	"testing"

	"github.com/example/orc/internal/core/effects"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// ============================================================================
// Mock Implementations
// ============================================================================

// mockMissionRepository implements secondary.MissionRepository for testing.
type mockMissionRepository struct {
	missions      map[string]*secondary.MissionRecord
	shipmentCount map[string]int
	createErr     error
	getErr        error
	updateErr     error
	deleteErr     error
	listErr       error
}

func newMockMissionRepository() *mockMissionRepository {
	return &mockMissionRepository{
		missions:      make(map[string]*secondary.MissionRecord),
		shipmentCount: make(map[string]int),
	}
}

func (m *mockMissionRepository) Create(ctx context.Context, mission *secondary.MissionRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	mission.ID = "MISSION-001"
	m.missions[mission.ID] = mission
	return nil
}

func (m *mockMissionRepository) GetByID(ctx context.Context, id string) (*secondary.MissionRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if mission, ok := m.missions[id]; ok {
		return mission, nil
	}
	return nil, errors.New("mission not found")
}

func (m *mockMissionRepository) Update(ctx context.Context, mission *secondary.MissionRecord) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.missions[mission.ID] = mission
	return nil
}

func (m *mockMissionRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.missions, id)
	return nil
}

func (m *mockMissionRepository) List(ctx context.Context, filters secondary.MissionFilters) ([]*secondary.MissionRecord, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*secondary.MissionRecord
	for _, mission := range m.missions {
		if filters.Status == "" || mission.Status == filters.Status {
			result = append(result, mission)
		}
	}
	return result, nil
}

func (m *mockMissionRepository) GetNextID(ctx context.Context) (string, error) {
	return "MISSION-001", nil
}

func (m *mockMissionRepository) CountShipments(ctx context.Context, missionID string) (int, error) {
	return m.shipmentCount[missionID], nil
}

func (m *mockMissionRepository) Pin(ctx context.Context, id string) error {
	if mission, ok := m.missions[id]; ok {
		mission.Pinned = true
	}
	return nil
}

func (m *mockMissionRepository) Unpin(ctx context.Context, id string) error {
	if mission, ok := m.missions[id]; ok {
		mission.Pinned = false
	}
	return nil
}

// mockGroveRepository implements secondary.GroveRepository for testing.
type mockGroveRepository struct {
	groves map[string][]*secondary.GroveRecord
}

func newMockGroveRepository() *mockGroveRepository {
	return &mockGroveRepository{
		groves: make(map[string][]*secondary.GroveRecord),
	}
}

func (m *mockGroveRepository) Create(ctx context.Context, grove *secondary.GroveRecord) error {
	grove.ID = "GROVE-001"
	m.groves[grove.MissionID] = append(m.groves[grove.MissionID], grove)
	return nil
}

func (m *mockGroveRepository) GetByID(ctx context.Context, id string) (*secondary.GroveRecord, error) {
	for _, groves := range m.groves {
		for _, g := range groves {
			if g.ID == id {
				return g, nil
			}
		}
	}
	return nil, errors.New("grove not found")
}

func (m *mockGroveRepository) GetByMission(ctx context.Context, missionID string) ([]*secondary.GroveRecord, error) {
	return m.groves[missionID], nil
}

func (m *mockGroveRepository) Update(ctx context.Context, grove *secondary.GroveRecord) error {
	return nil
}

func (m *mockGroveRepository) Delete(ctx context.Context, id string) error {
	for missionID, groves := range m.groves {
		for i, g := range groves {
			if g.ID == id {
				m.groves[missionID] = append(groves[:i], groves[i+1:]...)
				return nil
			}
		}
	}
	return nil
}

func (m *mockGroveRepository) GetNextID(ctx context.Context) (string, error) {
	return "GROVE-001", nil
}

// mockAgentProvider implements secondary.AgentIdentityProvider for testing.
type mockAgentProvider struct {
	identity *secondary.AgentIdentity
	err      error
}

func newMockAgentProvider(agentType secondary.AgentType) *mockAgentProvider {
	return &mockAgentProvider{
		identity: &secondary.AgentIdentity{
			Type:      agentType,
			ID:        "001",
			FullID:    string(agentType) + "-001",
			MissionID: "",
		},
	}
}

func (m *mockAgentProvider) GetCurrentIdentity(ctx context.Context) (*secondary.AgentIdentity, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.identity, nil
}

// mockEffectExecutor implements EffectExecutor for testing.
type mockEffectExecutor struct {
	executedEffects []effects.Effect
	executeErr      error
}

func newMockEffectExecutor() *mockEffectExecutor {
	return &mockEffectExecutor{
		executedEffects: []effects.Effect{},
	}
}

func (m *mockEffectExecutor) Execute(ctx context.Context, effs []effects.Effect) error {
	if m.executeErr != nil {
		return m.executeErr
	}
	m.executedEffects = append(m.executedEffects, effs...)
	return nil
}

// ============================================================================
// Test Helper
// ============================================================================

func newTestService(agentType secondary.AgentType) (*MissionServiceImpl, *mockMissionRepository, *mockGroveRepository, *mockEffectExecutor) {
	missionRepo := newMockMissionRepository()
	groveRepo := newMockGroveRepository()
	agentProvider := newMockAgentProvider(agentType)
	executor := newMockEffectExecutor()

	service := NewMissionService(missionRepo, groveRepo, agentProvider, executor)
	return service, missionRepo, groveRepo, executor
}

// ============================================================================
// CreateMission Tests
// ============================================================================

func TestCreateMission_ORCCanCreate(t *testing.T) {
	service, _, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	resp, err := service.CreateMission(ctx, primary.CreateMissionRequest{
		Title:       "Test Mission",
		Description: "A test mission",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.MissionID == "" {
		t.Error("expected mission ID to be set")
	}
	if resp.Mission.Title != "Test Mission" {
		t.Errorf("expected title 'Test Mission', got '%s'", resp.Mission.Title)
	}
}

func TestCreateMission_IMPCannotCreate(t *testing.T) {
	service, _, _, _ := newTestService(secondary.AgentTypeIMP)
	ctx := context.Background()

	_, err := service.CreateMission(ctx, primary.CreateMissionRequest{
		Title:       "Test Mission",
		Description: "A test mission",
	})

	if err == nil {
		t.Fatal("expected error for IMP creating mission, got nil")
	}
}

// Note: Only ORC and IMP agent types are defined. ORC can create, IMP cannot.
// Additional agent types could be added in the future.

// ============================================================================
// CompleteMission Tests
// ============================================================================

func TestCompleteMission_UnpinnedAllowed(t *testing.T) {
	service, missionRepo, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	// Setup: create an unpinned mission
	missionRepo.missions["MISSION-001"] = &secondary.MissionRecord{
		ID:     "MISSION-001",
		Title:  "Test Mission",
		Status: "active",
		Pinned: false,
	}

	err := service.CompleteMission(ctx, "MISSION-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if missionRepo.missions["MISSION-001"].Status != "complete" {
		t.Errorf("expected status 'complete', got '%s'", missionRepo.missions["MISSION-001"].Status)
	}
}

func TestCompleteMission_PinnedBlocked(t *testing.T) {
	service, missionRepo, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	// Setup: create a pinned mission
	missionRepo.missions["MISSION-001"] = &secondary.MissionRecord{
		ID:     "MISSION-001",
		Title:  "Pinned Mission",
		Status: "active",
		Pinned: true,
	}

	err := service.CompleteMission(ctx, "MISSION-001")

	if err == nil {
		t.Fatal("expected error for completing pinned mission, got nil")
	}
}

func TestCompleteMission_NotFound(t *testing.T) {
	service, _, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	err := service.CompleteMission(ctx, "MISSION-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent mission, got nil")
	}
}

// ============================================================================
// ArchiveMission Tests
// ============================================================================

func TestArchiveMission_UnpinnedAllowed(t *testing.T) {
	service, missionRepo, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	// Setup: create an unpinned mission
	missionRepo.missions["MISSION-001"] = &secondary.MissionRecord{
		ID:     "MISSION-001",
		Title:  "Test Mission",
		Status: "complete",
		Pinned: false,
	}

	err := service.ArchiveMission(ctx, "MISSION-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if missionRepo.missions["MISSION-001"].Status != "archived" {
		t.Errorf("expected status 'archived', got '%s'", missionRepo.missions["MISSION-001"].Status)
	}
}

func TestArchiveMission_PinnedBlocked(t *testing.T) {
	service, missionRepo, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	// Setup: create a pinned mission
	missionRepo.missions["MISSION-001"] = &secondary.MissionRecord{
		ID:     "MISSION-001",
		Title:  "Pinned Mission",
		Status: "complete",
		Pinned: true,
	}

	err := service.ArchiveMission(ctx, "MISSION-001")

	if err == nil {
		t.Fatal("expected error for archiving pinned mission, got nil")
	}
}

// ============================================================================
// DeleteMission Tests
// ============================================================================

func TestDeleteMission_NoDependents(t *testing.T) {
	service, missionRepo, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	// Setup: create a mission with no dependents
	missionRepo.missions["MISSION-001"] = &secondary.MissionRecord{
		ID:     "MISSION-001",
		Title:  "Empty Mission",
		Status: "active",
	}
	missionRepo.shipmentCount["MISSION-001"] = 0

	err := service.DeleteMission(ctx, primary.DeleteMissionRequest{
		MissionID: "MISSION-001",
		Force:     false,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, exists := missionRepo.missions["MISSION-001"]; exists {
		t.Error("expected mission to be deleted")
	}
}

func TestDeleteMission_HasDependentsNoForce(t *testing.T) {
	service, missionRepo, groveRepo, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	// Setup: create a mission with shipments
	missionRepo.missions["MISSION-001"] = &secondary.MissionRecord{
		ID:     "MISSION-001",
		Title:  "Mission with Shipments",
		Status: "active",
	}
	missionRepo.shipmentCount["MISSION-001"] = 3
	groveRepo.groves["MISSION-001"] = []*secondary.GroveRecord{
		{ID: "GROVE-001", MissionID: "MISSION-001", Name: "main-grove"},
	}

	err := service.DeleteMission(ctx, primary.DeleteMissionRequest{
		MissionID: "MISSION-001",
		Force:     false,
	})

	if err == nil {
		t.Fatal("expected error for deleting mission with dependents without force, got nil")
	}
	if _, exists := missionRepo.missions["MISSION-001"]; !exists {
		t.Error("mission should not have been deleted")
	}
}

func TestDeleteMission_HasDependentsWithForce(t *testing.T) {
	service, missionRepo, groveRepo, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	// Setup: create a mission with shipments
	missionRepo.missions["MISSION-001"] = &secondary.MissionRecord{
		ID:     "MISSION-001",
		Title:  "Mission with Shipments",
		Status: "active",
	}
	missionRepo.shipmentCount["MISSION-001"] = 3
	groveRepo.groves["MISSION-001"] = []*secondary.GroveRecord{
		{ID: "GROVE-001", MissionID: "MISSION-001", Name: "main-grove"},
	}

	err := service.DeleteMission(ctx, primary.DeleteMissionRequest{
		MissionID: "MISSION-001",
		Force:     true,
	})

	if err != nil {
		t.Fatalf("expected no error with force flag, got %v", err)
	}
	if _, exists := missionRepo.missions["MISSION-001"]; exists {
		t.Error("expected mission to be deleted with force")
	}
}

// ============================================================================
// StartMission Tests
// ============================================================================

func TestStartMission_ORCCanStart(t *testing.T) {
	service, missionRepo, groveRepo, executor := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	// Setup: create a mission with groves
	missionRepo.missions["MISSION-001"] = &secondary.MissionRecord{
		ID:     "MISSION-001",
		Title:  "Test Mission",
		Status: "active",
	}
	groveRepo.groves["MISSION-001"] = []*secondary.GroveRecord{
		{ID: "GROVE-001", MissionID: "MISSION-001", Name: "main-grove", WorktreePath: "/tmp/worktree"},
	}

	resp, err := service.StartMission(ctx, primary.StartMissionRequest{
		MissionID: "MISSION-001",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Mission.ID != "MISSION-001" {
		t.Errorf("expected mission ID 'MISSION-001', got '%s'", resp.Mission.ID)
	}
	// Verify effects were generated
	if len(executor.executedEffects) == 0 {
		t.Error("expected effects to be executed")
	}
}

func TestStartMission_IMPCannotStart(t *testing.T) {
	service, missionRepo, _, _ := newTestService(secondary.AgentTypeIMP)
	ctx := context.Background()

	// Setup: create a mission
	missionRepo.missions["MISSION-001"] = &secondary.MissionRecord{
		ID:     "MISSION-001",
		Title:  "Test Mission",
		Status: "active",
	}

	_, err := service.StartMission(ctx, primary.StartMissionRequest{
		MissionID: "MISSION-001",
	})

	if err == nil {
		t.Fatal("expected error for IMP starting mission, got nil")
	}
}

func TestStartMission_GeneratesTMuxEffects(t *testing.T) {
	service, missionRepo, groveRepo, executor := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	// Setup
	missionRepo.missions["MISSION-001"] = &secondary.MissionRecord{
		ID:     "MISSION-001",
		Title:  "Test Mission",
		Status: "active",
	}
	groveRepo.groves["MISSION-001"] = []*secondary.GroveRecord{
		{ID: "GROVE-001", MissionID: "MISSION-001", Name: "grove1", WorktreePath: "/path/to/worktree"},
	}

	_, err := service.StartMission(ctx, primary.StartMissionRequest{
		MissionID: "MISSION-001",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check that TMux effects were generated
	hasTMuxEffect := false
	for _, eff := range executor.executedEffects {
		if _, ok := eff.(effects.TMuxEffect); ok {
			hasTMuxEffect = true
			break
		}
	}
	if !hasTMuxEffect {
		t.Error("expected TMux effects to be generated for start mission")
	}
}

// ============================================================================
// LaunchMission Tests
// ============================================================================

func TestLaunchMission_ORCCanLaunch(t *testing.T) {
	service, _, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	resp, err := service.LaunchMission(ctx, primary.LaunchMissionRequest{
		Title:       "New Mission",
		Description: "A launched mission",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.MissionID == "" {
		t.Error("expected mission ID to be set")
	}
}

func TestLaunchMission_IMPCannotLaunch(t *testing.T) {
	service, _, _, _ := newTestService(secondary.AgentTypeIMP)
	ctx := context.Background()

	_, err := service.LaunchMission(ctx, primary.LaunchMissionRequest{
		Title:       "New Mission",
		Description: "A launched mission",
	})

	if err == nil {
		t.Fatal("expected error for IMP launching mission, got nil")
	}
}

// ============================================================================
// GetMission / ListMissions Tests
// ============================================================================

func TestGetMission_Found(t *testing.T) {
	service, missionRepo, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	missionRepo.missions["MISSION-001"] = &secondary.MissionRecord{
		ID:     "MISSION-001",
		Title:  "Test Mission",
		Status: "active",
	}

	mission, err := service.GetMission(ctx, "MISSION-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mission.Title != "Test Mission" {
		t.Errorf("expected title 'Test Mission', got '%s'", mission.Title)
	}
}

func TestGetMission_NotFound(t *testing.T) {
	service, _, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	_, err := service.GetMission(ctx, "MISSION-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent mission, got nil")
	}
}

func TestListMissions_FilterByStatus(t *testing.T) {
	service, missionRepo, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	missionRepo.missions["MISSION-001"] = &secondary.MissionRecord{
		ID:     "MISSION-001",
		Title:  "Active Mission",
		Status: "active",
	}
	missionRepo.missions["MISSION-002"] = &secondary.MissionRecord{
		ID:     "MISSION-002",
		Title:  "Complete Mission",
		Status: "complete",
	}

	missions, err := service.ListMissions(ctx, primary.MissionFilters{Status: "active"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(missions) != 1 {
		t.Errorf("expected 1 active mission, got %d", len(missions))
	}
}

// ============================================================================
// Pin/Unpin Tests
// ============================================================================

func TestPinMission(t *testing.T) {
	service, missionRepo, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	missionRepo.missions["MISSION-001"] = &secondary.MissionRecord{
		ID:     "MISSION-001",
		Title:  "Test Mission",
		Status: "active",
		Pinned: false,
	}

	err := service.PinMission(ctx, "MISSION-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !missionRepo.missions["MISSION-001"].Pinned {
		t.Error("expected mission to be pinned")
	}
}

func TestUnpinMission(t *testing.T) {
	service, missionRepo, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	missionRepo.missions["MISSION-001"] = &secondary.MissionRecord{
		ID:     "MISSION-001",
		Title:  "Test Mission",
		Status: "active",
		Pinned: true,
	}

	err := service.UnpinMission(ctx, "MISSION-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if missionRepo.missions["MISSION-001"].Pinned {
		t.Error("expected mission to be unpinned")
	}
}

// ============================================================================
// UpdateMission Tests
// ============================================================================

func TestUpdateMission_Title(t *testing.T) {
	service, missionRepo, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	missionRepo.missions["MISSION-001"] = &secondary.MissionRecord{
		ID:          "MISSION-001",
		Title:       "Old Title",
		Description: "Original description",
		Status:      "active",
	}

	err := service.UpdateMission(ctx, primary.UpdateMissionRequest{
		MissionID: "MISSION-001",
		Title:     "New Title",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if missionRepo.missions["MISSION-001"].Title != "New Title" {
		t.Errorf("expected title 'New Title', got '%s'", missionRepo.missions["MISSION-001"].Title)
	}
}

func TestUpdateMission_Description(t *testing.T) {
	service, missionRepo, _, _ := newTestService(secondary.AgentTypeORC)
	ctx := context.Background()

	missionRepo.missions["MISSION-001"] = &secondary.MissionRecord{
		ID:          "MISSION-001",
		Title:       "Test Mission",
		Description: "Old description",
		Status:      "active",
	}

	err := service.UpdateMission(ctx, primary.UpdateMissionRequest{
		MissionID:   "MISSION-001",
		Description: "New description",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if missionRepo.missions["MISSION-001"].Description != "New description" {
		t.Errorf("expected description 'New description', got '%s'", missionRepo.missions["MISSION-001"].Description)
	}
}
