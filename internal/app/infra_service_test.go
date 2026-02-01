package app

import (
	"context"
	"errors"
	"testing"

	"github.com/example/orc/internal/core/effects"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// mockInfraFactoryRepo implements secondary.FactoryRepository for infra tests.
type mockInfraFactoryRepo struct {
	factories map[string]*secondary.FactoryRecord
}

func (m *mockInfraFactoryRepo) GetByID(ctx context.Context, id string) (*secondary.FactoryRecord, error) {
	if f, ok := m.factories[id]; ok {
		return f, nil
	}
	return nil, errors.New("not found")
}

func (m *mockInfraFactoryRepo) GetByName(ctx context.Context, name string) (*secondary.FactoryRecord, error) {
	return nil, errors.New("not implemented")
}

func (m *mockInfraFactoryRepo) Create(ctx context.Context, f *secondary.FactoryRecord) error {
	return nil
}

func (m *mockInfraFactoryRepo) List(ctx context.Context, filters secondary.FactoryFilters) ([]*secondary.FactoryRecord, error) {
	return nil, nil
}

func (m *mockInfraFactoryRepo) Update(ctx context.Context, f *secondary.FactoryRecord) error {
	return nil
}

func (m *mockInfraFactoryRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockInfraFactoryRepo) GetNextID(ctx context.Context) (string, error) {
	return "FACT-001", nil
}

func (m *mockInfraFactoryRepo) CountWorkshops(ctx context.Context, factoryID string) (int, error) {
	return 0, nil
}

func (m *mockInfraFactoryRepo) CountCommissions(ctx context.Context, factoryID string) (int, error) {
	return 0, nil
}

// mockInfraWorkshopRepo implements secondary.WorkshopRepository for infra tests.
type mockInfraWorkshopRepo struct {
	workshops map[string]*secondary.WorkshopRecord
}

func (m *mockInfraWorkshopRepo) GetByID(ctx context.Context, id string) (*secondary.WorkshopRecord, error) {
	if w, ok := m.workshops[id]; ok {
		return w, nil
	}
	return nil, errors.New("not found")
}

func (m *mockInfraWorkshopRepo) Create(ctx context.Context, w *secondary.WorkshopRecord) error {
	return nil
}

func (m *mockInfraWorkshopRepo) List(ctx context.Context, filters secondary.WorkshopFilters) ([]*secondary.WorkshopRecord, error) {
	return nil, nil
}

func (m *mockInfraWorkshopRepo) Update(ctx context.Context, w *secondary.WorkshopRecord) error {
	return nil
}

func (m *mockInfraWorkshopRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockInfraWorkshopRepo) GetNextID(ctx context.Context) (string, error) {
	return "WORK-001", nil
}

func (m *mockInfraWorkshopRepo) FactoryExists(ctx context.Context, factoryID string) (bool, error) {
	return true, nil
}

func (m *mockInfraWorkshopRepo) CountWorkbenches(ctx context.Context, workshopID string) (int, error) {
	return 0, nil
}

func (m *mockInfraWorkshopRepo) UpdateFocusedConclaveID(ctx context.Context, workshopID, conclaveID string) error {
	return nil
}

func (m *mockInfraWorkshopRepo) SetActiveCommissionID(ctx context.Context, workshopID, commissionID string) error {
	return nil
}

func (m *mockInfraWorkshopRepo) CountByFactory(ctx context.Context, factoryID string) (int, error) {
	return 0, nil
}

// mockInfraWorkbenchRepo implements secondary.WorkbenchRepository for infra tests.
type mockInfraWorkbenchRepo struct {
	workbenches []*secondary.WorkbenchRecord
}

func (m *mockInfraWorkbenchRepo) List(ctx context.Context, workshopID string) ([]*secondary.WorkbenchRecord, error) {
	return m.workbenches, nil
}

func (m *mockInfraWorkbenchRepo) GetByID(ctx context.Context, id string) (*secondary.WorkbenchRecord, error) {
	for _, wb := range m.workbenches {
		if wb.ID == id {
			return wb, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockInfraWorkbenchRepo) Create(ctx context.Context, wb *secondary.WorkbenchRecord) error {
	return nil
}

func (m *mockInfraWorkbenchRepo) Update(ctx context.Context, wb *secondary.WorkbenchRecord) error {
	return nil
}

func (m *mockInfraWorkbenchRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockInfraWorkbenchRepo) GetNextID(ctx context.Context) (string, error) {
	return "BENCH-001", nil
}

func (m *mockInfraWorkbenchRepo) GetByPath(ctx context.Context, path string) (*secondary.WorkbenchRecord, error) {
	return nil, errors.New("not found")
}

func (m *mockInfraWorkbenchRepo) WorkshopExists(ctx context.Context, workshopID string) (bool, error) {
	return true, nil
}

func (m *mockInfraWorkbenchRepo) ListAll(ctx context.Context) ([]*secondary.WorkbenchRecord, error) {
	return m.workbenches, nil
}

func (m *mockInfraWorkbenchRepo) ListByCommission(ctx context.Context, commissionID string) ([]*secondary.WorkbenchRecord, error) {
	return nil, nil
}

func (m *mockInfraWorkbenchRepo) GetByWorkshop(ctx context.Context, workshopID string) ([]*secondary.WorkbenchRecord, error) {
	return m.workbenches, nil
}

func (m *mockInfraWorkbenchRepo) Rename(ctx context.Context, id, newName string) error {
	return nil
}

func (m *mockInfraWorkbenchRepo) UpdatePath(ctx context.Context, id, newPath string) error {
	return nil
}

func (m *mockInfraWorkbenchRepo) UpdateFocusedID(ctx context.Context, id, focusedID string) error {
	return nil
}

// mockInfraRepoRepo implements secondary.RepoRepository for infra tests.
type mockInfraRepoRepo struct{}

func (m *mockInfraRepoRepo) GetByID(ctx context.Context, id string) (*secondary.RepoRecord, error) {
	return &secondary.RepoRecord{ID: id, Name: "test-repo"}, nil
}

func (m *mockInfraRepoRepo) Create(ctx context.Context, r *secondary.RepoRecord) error {
	return nil
}

func (m *mockInfraRepoRepo) List(ctx context.Context, filters secondary.RepoFilters) ([]*secondary.RepoRecord, error) {
	return nil, nil
}

func (m *mockInfraRepoRepo) Update(ctx context.Context, r *secondary.RepoRecord) error {
	return nil
}

func (m *mockInfraRepoRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockInfraRepoRepo) GetNextID(ctx context.Context) (string, error) {
	return "REPO-001", nil
}

func (m *mockInfraRepoRepo) GetByName(ctx context.Context, name string) (*secondary.RepoRecord, error) {
	return nil, errors.New("not found")
}

func (m *mockInfraRepoRepo) UpdateStatus(ctx context.Context, id, status string) error {
	return nil
}

func (m *mockInfraRepoRepo) HasActivePRs(ctx context.Context, repoID string) (bool, error) {
	return false, nil
}

// mockInfraGatehouseRepo implements secondary.GatehouseRepository for infra tests.
type mockInfraGatehouseRepo struct {
	gatehouses map[string]*secondary.GatehouseRecord
}

func (m *mockInfraGatehouseRepo) GetByWorkshop(ctx context.Context, workshopID string) (*secondary.GatehouseRecord, error) {
	if g, ok := m.gatehouses[workshopID]; ok {
		return g, nil
	}
	return nil, errors.New("not found")
}

func (m *mockInfraGatehouseRepo) GetByID(ctx context.Context, id string) (*secondary.GatehouseRecord, error) {
	for _, g := range m.gatehouses {
		if g.ID == id {
			return g, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockInfraGatehouseRepo) Create(ctx context.Context, g *secondary.GatehouseRecord) error {
	return nil
}

func (m *mockInfraGatehouseRepo) List(ctx context.Context, filters secondary.GatehouseFilters) ([]*secondary.GatehouseRecord, error) {
	return nil, nil
}

func (m *mockInfraGatehouseRepo) Update(ctx context.Context, g *secondary.GatehouseRecord) error {
	return nil
}

func (m *mockInfraGatehouseRepo) UpdateStatus(ctx context.Context, id, status string) error {
	return nil
}

func (m *mockInfraGatehouseRepo) WorkshopExists(ctx context.Context, workshopID string) (bool, error) {
	return true, nil
}

func (m *mockInfraGatehouseRepo) WorkshopHasGatehouse(ctx context.Context, workshopID string) (bool, error) {
	_, ok := m.gatehouses[workshopID]
	return ok, nil
}

func (m *mockInfraGatehouseRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockInfraGatehouseRepo) GetNextID(ctx context.Context) (string, error) {
	return "GATE-001", nil
}

// mockInfraWorkspaceAdapter implements secondary.WorkspaceAdapter for testing.
type mockInfraWorkspaceAdapter struct{}

func (m *mockInfraWorkspaceAdapter) WorktreeExists(ctx context.Context, path string) (bool, error) {
	return false, nil
}

func (m *mockInfraWorkspaceAdapter) CreateWorktree(ctx context.Context, repoPath, branch, worktreePath string) error {
	return nil
}

func (m *mockInfraWorkspaceAdapter) RemoveWorktree(ctx context.Context, path string) error {
	return nil
}

func (m *mockInfraWorkspaceAdapter) CreateDirectory(ctx context.Context, path string) error {
	return nil
}

func (m *mockInfraWorkspaceAdapter) RemoveDirectory(ctx context.Context, path string) error {
	return nil
}

func (m *mockInfraWorkspaceAdapter) DirectoryExists(ctx context.Context, path string) (bool, error) {
	return false, nil
}

func (m *mockInfraWorkspaceAdapter) GetWorktreesBasePath() string {
	return "/tmp/worktrees"
}

func (m *mockInfraWorkspaceAdapter) GetRepoPath(repoName string) string {
	return "/tmp/repos/" + repoName
}

func (m *mockInfraWorkspaceAdapter) ResolveWorkbenchPath(workbenchName string) string {
	return "/tmp/worktrees/" + workbenchName
}

// mockInfraEffectExecutor implements EffectExecutor for testing.
type mockInfraEffectExecutor struct{}

func (m *mockInfraEffectExecutor) Execute(ctx context.Context, effs []effects.Effect) error {
	return nil
}

func newTestInfraService() *InfraServiceImpl {
	factoryRepo := &mockInfraFactoryRepo{
		factories: map[string]*secondary.FactoryRecord{
			"FACT-001": {ID: "FACT-001", Name: "default", Status: "active"},
		},
	}
	workshopRepo := &mockInfraWorkshopRepo{
		workshops: map[string]*secondary.WorkshopRecord{
			"WORK-001": {ID: "WORK-001", Name: "Test Workshop", FactoryID: "FACT-001", Status: "active"},
		},
	}
	workbenchRepo := &mockInfraWorkbenchRepo{
		workbenches: []*secondary.WorkbenchRecord{
			{ID: "BENCH-001", Name: "test-bench", WorktreePath: "/tmp/test-bench", WorkshopID: "WORK-001"},
		},
	}
	repoRepo := &mockInfraRepoRepo{}
	gatehouseRepo := &mockInfraGatehouseRepo{
		gatehouses: map[string]*secondary.GatehouseRecord{
			"WORK-001": {ID: "GATE-001", WorkshopID: "WORK-001", Status: "active"},
		},
	}
	workspaceAdapter := &mockInfraWorkspaceAdapter{}
	executor := &mockInfraEffectExecutor{}

	return NewInfraService(factoryRepo, workshopRepo, workbenchRepo, repoRepo, gatehouseRepo, workspaceAdapter, executor)
}

func TestInfraService_PlanInfra(t *testing.T) {
	service := newTestInfraService()
	ctx := context.Background()

	plan, err := service.PlanInfra(ctx, primary.InfraPlanRequest{
		WorkshopID: "WORK-001",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if plan.WorkshopID != "WORK-001" {
		t.Errorf("expected workshop ID 'WORK-001', got %q", plan.WorkshopID)
	}

	if plan.WorkshopName != "Test Workshop" {
		t.Errorf("expected workshop name 'Test Workshop', got %q", plan.WorkshopName)
	}

	if plan.FactoryID != "FACT-001" {
		t.Errorf("expected factory ID 'FACT-001', got %q", plan.FactoryID)
	}

	if plan.Gatehouse == nil {
		t.Fatal("expected gatehouse op to be set")
	}

	if plan.Gatehouse.ID != "GATE-001" {
		t.Errorf("expected gatehouse ID 'GATE-001', got %q", plan.Gatehouse.ID)
	}

	if len(plan.Workbenches) != 1 {
		t.Errorf("expected 1 workbench, got %d", len(plan.Workbenches))
	}

	if len(plan.Workbenches) > 0 && plan.Workbenches[0].ID != "BENCH-001" {
		t.Errorf("expected workbench ID 'BENCH-001', got %q", plan.Workbenches[0].ID)
	}
}

func TestInfraService_PlanInfra_WorkshopNotFound(t *testing.T) {
	service := newTestInfraService()
	ctx := context.Background()

	_, err := service.PlanInfra(ctx, primary.InfraPlanRequest{
		WorkshopID: "WORK-999",
	})

	if err == nil {
		t.Error("expected error for non-existent workshop")
	}
}

func TestInfraService_PlanInfra_OpStatusCreate(t *testing.T) {
	// Test that non-existent paths result in CREATE status
	service := newTestInfraService()
	ctx := context.Background()

	plan, err := service.PlanInfra(ctx, primary.InfraPlanRequest{
		WorkshopID: "WORK-001",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// The gatehouse path won't exist in tests, so it should be CREATE
	if plan.Gatehouse.Status != primary.OpCreate {
		t.Errorf("expected gatehouse status CREATE, got %s", plan.Gatehouse.Status)
	}

	// The workbench path won't exist either
	if len(plan.Workbenches) > 0 && plan.Workbenches[0].Status != primary.OpMissing {
		t.Errorf("expected workbench status MISSING, got %s", plan.Workbenches[0].Status)
	}
}
