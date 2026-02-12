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

func (m *mockInfraWorkshopRepo) SetActiveCommissionID(ctx context.Context, workshopID, commissionID string) error {
	return nil
}

func (m *mockInfraWorkshopRepo) GetActiveCommissions(ctx context.Context, workshopID string) ([]string, error) {
	return nil, nil
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

func (m *mockInfraWorkbenchRepo) GetByFocusedID(ctx context.Context, focusedID string) ([]*secondary.WorkbenchRecord, error) {
	return nil, nil
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
// Captures executed effects for verification.
type mockInfraEffectExecutor struct {
	executedEffects []effects.Effect
	executeErr      error
}

func (m *mockInfraEffectExecutor) Execute(ctx context.Context, effs []effects.Effect) error {
	if m.executeErr != nil {
		return m.executeErr
	}
	m.executedEffects = append(m.executedEffects, effs...)
	return nil
}

// hasFileEffectWithOperation checks if any executed effect is a FileEffect with the given operation.
func (m *mockInfraEffectExecutor) hasFileEffectWithOperation(op string) bool {
	for _, e := range m.executedEffects {
		if fe, ok := e.(effects.FileEffect); ok && fe.Operation == op {
			return true
		}
	}
	return false
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
	workspaceAdapter := &mockInfraWorkspaceAdapter{}
	executor := &mockInfraEffectExecutor{}

	return NewInfraService(factoryRepo, workshopRepo, workbenchRepo, repoRepo, workspaceAdapter, nil, executor)
}

// newTestInfraServiceWithExecutor creates a test infra service with a custom effect executor.
func newTestInfraServiceWithExecutor(executor *mockInfraEffectExecutor) *InfraServiceImpl {
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
			{ID: "BENCH-001", Name: "test-bench", WorktreePath: "/tmp/test-bench", WorkshopID: "WORK-001", Status: "active"},
		},
	}
	repoRepo := &mockInfraRepoRepo{}
	workspaceAdapter := &mockInfraWorkspaceAdapter{}

	return NewInfraService(factoryRepo, workshopRepo, workbenchRepo, repoRepo, workspaceAdapter, nil, executor)
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

	if plan.WorkshopDir == nil {
		t.Fatal("expected workshop dir op to be set")
	}

	if plan.WorkshopDir.ID != "WORK-001" {
		t.Errorf("expected workshop dir ID 'WORK-001', got %q", plan.WorkshopDir.ID)
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

	// The workshop dir path won't exist in tests, so it should be CREATE
	if plan.WorkshopDir.Status != primary.OpCreate {
		t.Errorf("expected workshop dir status CREATE, got %s", plan.WorkshopDir.Status)
	}

	// The workbench path won't exist either
	if len(plan.Workbenches) > 0 && plan.Workbenches[0].Status != primary.OpMissing {
		t.Errorf("expected workbench status MISSING, got %s", plan.Workbenches[0].Status)
	}
}

func TestInfraService_PlanInfra_ArchivedWorkbenchExcluded(t *testing.T) {
	// Test that archived workbenches are excluded from the plan
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
	// Include both active and archived workbenches
	workbenchRepo := &mockInfraWorkbenchRepo{
		workbenches: []*secondary.WorkbenchRecord{
			{ID: "BENCH-001", Name: "active-bench", WorktreePath: "/tmp/active-bench", WorkshopID: "WORK-001", Status: "active"},
			{ID: "BENCH-002", Name: "archived-bench", WorktreePath: "/tmp/archived-bench", WorkshopID: "WORK-001", Status: "archived"},
		},
	}
	repoRepo := &mockInfraRepoRepo{}
	workspaceAdapter := &mockInfraWorkspaceAdapter{}
	executor := &mockInfraEffectExecutor{}

	service := NewInfraService(factoryRepo, workshopRepo, workbenchRepo, repoRepo, workspaceAdapter, nil, executor)
	ctx := context.Background()

	plan, err := service.PlanInfra(ctx, primary.InfraPlanRequest{
		WorkshopID: "WORK-001",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should only include active workbench, not archived
	if len(plan.Workbenches) != 1 {
		t.Errorf("expected 1 workbench (active only), got %d", len(plan.Workbenches))
	}

	if len(plan.Workbenches) > 0 && plan.Workbenches[0].ID != "BENCH-001" {
		t.Errorf("expected active workbench BENCH-001, got %q", plan.Workbenches[0].ID)
	}

	// Archived workbench should NOT appear in orphan list either
	// (archived workbenches have DB records, so they're not orphans)
	for _, orphan := range plan.OrphanWorkbenches {
		if orphan.ID == "BENCH-002" {
			t.Error("archived workbench should NOT appear in orphan list")
		}
	}
}

func TestInfraService_ApplyInfra_NoFilesystemDeletion(t *testing.T) {
	// Test that ApplyInfra does NOT execute any filesystem deletion effects
	executor := &mockInfraEffectExecutor{}
	service := newTestInfraServiceWithExecutor(executor)
	ctx := context.Background()

	// First get a plan
	plan, err := service.PlanInfra(ctx, primary.InfraPlanRequest{
		WorkshopID: "WORK-001",
	})
	if err != nil {
		t.Fatalf("PlanInfra failed: %v", err)
	}

	// Apply the plan
	_, err = service.ApplyInfra(ctx, plan)
	if err != nil {
		t.Fatalf("ApplyInfra failed: %v", err)
	}

	// Verify no deletion effects were executed
	if executor.hasFileEffectWithOperation("rmdir") {
		t.Error("ApplyInfra should NOT execute rmdir effects")
	}
	if executor.hasFileEffectWithOperation("rm") {
		t.Error("ApplyInfra should NOT execute rm effects")
	}
	if executor.hasFileEffectWithOperation("remove") {
		t.Error("ApplyInfra should NOT execute remove effects")
	}
}
