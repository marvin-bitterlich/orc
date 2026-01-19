package app

import (
	"context"
	"testing"

	"github.com/example/orc/internal/core/effects"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// ============================================================================
// Test Helper
// ============================================================================

func newTestGroveService(agentType secondary.AgentType) (*GroveServiceImpl, *mockMissionRepository, *mockGroveRepository, *mockEffectExecutor) {
	missionRepo := newMockMissionRepository()
	groveRepo := newMockGroveRepository()
	agentProvider := newMockAgentProvider(agentType)
	executor := newMockEffectExecutor()

	service := NewGroveService(groveRepo, missionRepo, agentProvider, executor)
	return service, missionRepo, groveRepo, executor
}

// ============================================================================
// CreateGrove Tests
// ============================================================================

func TestCreateGrove_ORCCanCreate(t *testing.T) {
	service, missionRepo, _, executor := newTestGroveService(secondary.AgentTypeORC)
	ctx := context.Background()

	// Setup: mission exists
	missionRepo.missions["MISSION-001"] = &secondary.MissionRecord{
		ID:     "MISSION-001",
		Title:  "Test Mission",
		Status: "active",
	}

	resp, err := service.CreateGrove(ctx, primary.CreateGroveRequest{
		Name:      "auth-backend",
		MissionID: "MISSION-001",
		Repos:     []string{},
		BasePath:  "/tmp/worktrees",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.GroveID == "" {
		t.Error("expected grove ID to be set")
	}
	if resp.Path == "" {
		t.Error("expected path to be set")
	}
	if len(executor.executedEffects) == 0 {
		t.Error("expected effects to be executed")
	}
}

func TestCreateGrove_IMPCannotCreate(t *testing.T) {
	service, missionRepo, _, _ := newTestGroveService(secondary.AgentTypeIMP)
	ctx := context.Background()

	missionRepo.missions["MISSION-001"] = &secondary.MissionRecord{
		ID:     "MISSION-001",
		Title:  "Test Mission",
		Status: "active",
	}

	_, err := service.CreateGrove(ctx, primary.CreateGroveRequest{
		Name:      "test-grove",
		MissionID: "MISSION-001",
	})

	if err == nil {
		t.Fatal("expected error for IMP creating grove, got nil")
	}
}

func TestCreateGrove_MissionMustExist(t *testing.T) {
	service, _, _, _ := newTestGroveService(secondary.AgentTypeORC)
	ctx := context.Background()

	// No mission setup - mission doesn't exist

	_, err := service.CreateGrove(ctx, primary.CreateGroveRequest{
		Name:      "test-grove",
		MissionID: "MISSION-NONEXISTENT",
	})

	if err == nil {
		t.Fatal("expected error for non-existent mission, got nil")
	}
}

func TestCreateGrove_GeneratesEffects(t *testing.T) {
	service, missionRepo, _, executor := newTestGroveService(secondary.AgentTypeORC)
	ctx := context.Background()

	missionRepo.missions["MISSION-001"] = &secondary.MissionRecord{
		ID:     "MISSION-001",
		Title:  "Test Mission",
		Status: "active",
	}

	_, err := service.CreateGrove(ctx, primary.CreateGroveRequest{
		Name:      "backend",
		MissionID: "MISSION-001",
		BasePath:  "/tmp/worktrees",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should have filesystem effects (mkdir, write config)
	hasFileEffect := false
	for _, eff := range executor.executedEffects {
		if _, ok := eff.(effects.FileEffect); ok {
			hasFileEffect = true
			break
		}
	}
	if !hasFileEffect {
		t.Error("expected FileEffect to be executed")
	}
}

// ============================================================================
// GetGrove Tests
// ============================================================================

func TestGetGrove_Found(t *testing.T) {
	service, _, groveRepo, _ := newTestGroveService(secondary.AgentTypeORC)
	ctx := context.Background()

	groveRepo.groves["MISSION-001"] = []*secondary.GroveRecord{
		{ID: "GROVE-001", Name: "backend", MissionID: "MISSION-001", WorktreePath: "/tmp/grove"},
	}

	grove, err := service.GetGrove(ctx, "GROVE-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if grove.Name != "backend" {
		t.Errorf("expected name 'backend', got '%s'", grove.Name)
	}
}

func TestGetGrove_NotFound(t *testing.T) {
	service, _, _, _ := newTestGroveService(secondary.AgentTypeORC)
	ctx := context.Background()

	_, err := service.GetGrove(ctx, "GROVE-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent grove, got nil")
	}
}

// ============================================================================
// ListGroves Tests
// ============================================================================

func TestListGroves_FilterByMission(t *testing.T) {
	service, _, groveRepo, _ := newTestGroveService(secondary.AgentTypeORC)
	ctx := context.Background()

	groveRepo.groves["MISSION-001"] = []*secondary.GroveRecord{
		{ID: "GROVE-001", Name: "backend", MissionID: "MISSION-001"},
		{ID: "GROVE-002", Name: "frontend", MissionID: "MISSION-001"},
	}
	groveRepo.groves["MISSION-002"] = []*secondary.GroveRecord{
		{ID: "GROVE-003", Name: "other", MissionID: "MISSION-002"},
	}

	groves, err := service.ListGroves(ctx, primary.GroveFilters{MissionID: "MISSION-001"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(groves) != 2 {
		t.Errorf("expected 2 groves for MISSION-001, got %d", len(groves))
	}
}

func TestListGroves_Empty(t *testing.T) {
	service, _, _, _ := newTestGroveService(secondary.AgentTypeORC)
	ctx := context.Background()

	groves, err := service.ListGroves(ctx, primary.GroveFilters{MissionID: "MISSION-EMPTY"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if groves == nil {
		t.Error("expected empty slice, got nil")
	}
}

// ============================================================================
// RenameGrove Tests
// ============================================================================

func TestRenameGrove_Success(t *testing.T) {
	service, _, groveRepo, _ := newTestGroveService(secondary.AgentTypeORC)
	ctx := context.Background()

	groveRepo.groves["MISSION-001"] = []*secondary.GroveRecord{
		{ID: "GROVE-001", Name: "old-name", MissionID: "MISSION-001"},
	}

	err := service.RenameGrove(ctx, primary.RenameGroveRequest{
		GroveID: "GROVE-001",
		NewName: "new-name",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRenameGrove_NotFound(t *testing.T) {
	service, _, _, _ := newTestGroveService(secondary.AgentTypeORC)
	ctx := context.Background()

	err := service.RenameGrove(ctx, primary.RenameGroveRequest{
		GroveID: "GROVE-NONEXISTENT",
		NewName: "new-name",
	})

	if err == nil {
		t.Fatal("expected error for non-existent grove, got nil")
	}
}

// ============================================================================
// DeleteGrove Tests
// ============================================================================

func TestDeleteGrove_NoActiveTasks(t *testing.T) {
	service, _, groveRepo, _ := newTestGroveService(secondary.AgentTypeORC)
	ctx := context.Background()

	groveRepo.groves["MISSION-001"] = []*secondary.GroveRecord{
		{ID: "GROVE-001", Name: "backend", MissionID: "MISSION-001"},
	}

	err := service.DeleteGrove(ctx, primary.DeleteGroveRequest{
		GroveID: "GROVE-001",
		Force:   false,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeleteGrove_WithForce(t *testing.T) {
	service, _, groveRepo, _ := newTestGroveService(secondary.AgentTypeORC)
	ctx := context.Background()

	groveRepo.groves["MISSION-001"] = []*secondary.GroveRecord{
		{ID: "GROVE-001", Name: "backend", MissionID: "MISSION-001"},
	}

	err := service.DeleteGrove(ctx, primary.DeleteGroveRequest{
		GroveID: "GROVE-001",
		Force:   true,
	})

	if err != nil {
		t.Fatalf("expected no error with force, got %v", err)
	}
}

func TestDeleteGrove_NotFound(t *testing.T) {
	service, _, _, _ := newTestGroveService(secondary.AgentTypeORC)
	ctx := context.Background()

	err := service.DeleteGrove(ctx, primary.DeleteGroveRequest{
		GroveID: "GROVE-NONEXISTENT",
		Force:   false,
	})

	if err == nil {
		t.Fatal("expected error for non-existent grove, got nil")
	}
}

// ============================================================================
// OpenGrove Tests - Note: These test guard logic, actual TMux execution is mocked
// ============================================================================

func TestOpenGrove_GroveNotFound(t *testing.T) {
	service, _, _, _ := newTestGroveService(secondary.AgentTypeORC)
	ctx := context.Background()

	_, err := service.OpenGrove(ctx, primary.OpenGroveRequest{
		GroveID: "GROVE-NONEXISTENT",
	})

	if err == nil {
		t.Fatal("expected error for non-existent grove, got nil")
	}
}

// Note: Full OpenGrove tests would require mocking TMux environment
// which is outside the scope of unit tests. The guard logic is tested
// in guards_test.go.
