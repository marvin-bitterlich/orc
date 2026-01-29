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
// Mock Implementations for Workbench Service
// ============================================================================

// mockWorkbenchRepository implements secondary.WorkbenchRepository for testing.
type mockWorkbenchRepository struct {
	workbenches    map[string]*secondary.WorkbenchRecord
	workshopExists map[string]bool
	nextID         string
	createErr      error
	getErr         error
	updateErr      error
	deleteErr      error
	listErr        error
	renameErr      error
	updatePathErr  error
}

func newMockWorkbenchRepository() *mockWorkbenchRepository {
	return &mockWorkbenchRepository{
		workbenches:    make(map[string]*secondary.WorkbenchRecord),
		workshopExists: make(map[string]bool),
		nextID:         "BENCH-001",
	}
}

func (m *mockWorkbenchRepository) Create(ctx context.Context, workbench *secondary.WorkbenchRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	workbench.ID = m.nextID
	m.workbenches[workbench.ID] = workbench
	return nil
}

func (m *mockWorkbenchRepository) GetByID(ctx context.Context, id string) (*secondary.WorkbenchRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if workbench, ok := m.workbenches[id]; ok {
		return workbench, nil
	}
	return nil, errors.New("workbench not found")
}

func (m *mockWorkbenchRepository) GetByPath(ctx context.Context, path string) (*secondary.WorkbenchRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, wb := range m.workbenches {
		if wb.WorktreePath == path {
			return wb, nil
		}
	}
	return nil, errors.New("workbench not found at path")
}

func (m *mockWorkbenchRepository) GetByWorkshop(ctx context.Context, workshopID string) ([]*secondary.WorkbenchRecord, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*secondary.WorkbenchRecord
	for _, wb := range m.workbenches {
		if wb.WorkshopID == workshopID && wb.Status == "active" {
			result = append(result, wb)
		}
	}
	return result, nil
}

func (m *mockWorkbenchRepository) List(ctx context.Context, workshopID string) ([]*secondary.WorkbenchRecord, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*secondary.WorkbenchRecord
	for _, wb := range m.workbenches {
		if workshopID == "" || wb.WorkshopID == workshopID {
			result = append(result, wb)
		}
	}
	return result, nil
}

func (m *mockWorkbenchRepository) Update(ctx context.Context, workbench *secondary.WorkbenchRecord) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.workbenches[workbench.ID]; !ok {
		return errors.New("workbench not found")
	}
	m.workbenches[workbench.ID] = workbench
	return nil
}

func (m *mockWorkbenchRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.workbenches[id]; !ok {
		return errors.New("workbench not found")
	}
	delete(m.workbenches, id)
	return nil
}

func (m *mockWorkbenchRepository) Rename(ctx context.Context, id, newName string) error {
	if m.renameErr != nil {
		return m.renameErr
	}
	if wb, ok := m.workbenches[id]; ok {
		wb.Name = newName
		return nil
	}
	return errors.New("workbench not found")
}

func (m *mockWorkbenchRepository) UpdatePath(ctx context.Context, id, newPath string) error {
	if m.updatePathErr != nil {
		return m.updatePathErr
	}
	if wb, ok := m.workbenches[id]; ok {
		wb.WorktreePath = newPath
		return nil
	}
	return errors.New("workbench not found")
}

func (m *mockWorkbenchRepository) GetNextID(ctx context.Context) (string, error) {
	return m.nextID, nil
}

func (m *mockWorkbenchRepository) WorkshopExists(ctx context.Context, workshopID string) (bool, error) {
	return m.workshopExists[workshopID], nil
}

func (m *mockWorkbenchRepository) UpdateFocusedID(ctx context.Context, id, focusedID string) error {
	if wb, ok := m.workbenches[id]; ok {
		wb.FocusedID = focusedID
		return nil
	}
	return errors.New("workbench not found")
}

// mockWorkshopRepositoryForWorkbench implements secondary.WorkshopRepository minimally.
type mockWorkshopRepositoryForWorkbench struct {
	workshops map[string]*secondary.WorkshopRecord
}

func newMockWorkshopRepositoryForWorkbench() *mockWorkshopRepositoryForWorkbench {
	return &mockWorkshopRepositoryForWorkbench{
		workshops: make(map[string]*secondary.WorkshopRecord),
	}
}

func (m *mockWorkshopRepositoryForWorkbench) Create(ctx context.Context, workshop *secondary.WorkshopRecord) error {
	m.workshops[workshop.ID] = workshop
	return nil
}

func (m *mockWorkshopRepositoryForWorkbench) GetByID(ctx context.Context, id string) (*secondary.WorkshopRecord, error) {
	if ws, ok := m.workshops[id]; ok {
		return ws, nil
	}
	return nil, errors.New("workshop not found")
}

func (m *mockWorkshopRepositoryForWorkbench) List(ctx context.Context, filters secondary.WorkshopFilters) ([]*secondary.WorkshopRecord, error) {
	var result []*secondary.WorkshopRecord
	for _, ws := range m.workshops {
		result = append(result, ws)
	}
	return result, nil
}

func (m *mockWorkshopRepositoryForWorkbench) Update(ctx context.Context, workshop *secondary.WorkshopRecord) error {
	m.workshops[workshop.ID] = workshop
	return nil
}

func (m *mockWorkshopRepositoryForWorkbench) Delete(ctx context.Context, id string) error {
	delete(m.workshops, id)
	return nil
}

func (m *mockWorkshopRepositoryForWorkbench) GetNextID(ctx context.Context) (string, error) {
	return "WORK-001", nil
}

func (m *mockWorkshopRepositoryForWorkbench) CountWorkbenches(ctx context.Context, workshopID string) (int, error) {
	return 0, nil
}

func (m *mockWorkshopRepositoryForWorkbench) CountByFactory(ctx context.Context, factoryID string) (int, error) {
	return 0, nil
}

func (m *mockWorkshopRepositoryForWorkbench) FactoryExists(ctx context.Context, factoryID string) (bool, error) {
	return true, nil
}

func (m *mockWorkshopRepositoryForWorkbench) UpdateFocusedConclaveID(ctx context.Context, id, conclaveID string) error {
	if ws, ok := m.workshops[id]; ok {
		ws.FocusedConclaveID = conclaveID
		return nil
	}
	return errors.New("workshop not found")
}

func (m *mockWorkshopRepositoryForWorkbench) SetActiveCommissionID(ctx context.Context, workshopID, commissionID string) error {
	if ws, ok := m.workshops[workshopID]; ok {
		ws.ActiveCommissionID = commissionID
		return nil
	}
	return errors.New("workshop not found")
}

// ============================================================================
// Test Helper
// ============================================================================

func newTestWorkbenchService() (*WorkbenchServiceImpl, *mockWorkbenchRepository, *mockWorkshopRepositoryForWorkbench, *mockEffectExecutor) {
	workbenchRepo := newMockWorkbenchRepository()
	workshopRepo := newMockWorkshopRepositoryForWorkbench()
	agentProvider := newMockAgentProvider(secondary.AgentTypeORC)
	executor := newMockEffectExecutor()

	service := NewWorkbenchService(workbenchRepo, workshopRepo, agentProvider, executor)
	return service, workbenchRepo, workshopRepo, executor
}

// ============================================================================
// CreateWorkbench Tests
// ============================================================================

func TestWorkbenchService_CreateWorkbench(t *testing.T) {
	service, workbenchRepo, _, _ := newTestWorkbenchService()
	ctx := context.Background()

	// Setup: workshop exists
	workbenchRepo.workshopExists["WORK-001"] = true

	resp, err := service.CreateWorkbench(ctx, primary.CreateWorkbenchRequest{
		Name:       "test-bench",
		WorkshopID: "WORK-001",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.WorkbenchID == "" {
		t.Error("expected workbench ID to be set")
	}
	if resp.Workbench.Name != "test-bench" {
		t.Errorf("expected name 'test-bench', got '%s'", resp.Workbench.Name)
	}
}

func TestWorkbenchService_CreateWorkbench_WorkshopNotFound(t *testing.T) {
	service, _, _, _ := newTestWorkbenchService()
	ctx := context.Background()

	// No workshop exists
	_, err := service.CreateWorkbench(ctx, primary.CreateWorkbenchRequest{
		Name:       "test-bench",
		WorkshopID: "WORK-999",
	})

	if err == nil {
		t.Fatal("expected error for non-existent workshop, got nil")
	}
}

// ============================================================================
// GetWorkbench Tests
// ============================================================================

func TestWorkbenchService_GetWorkbench(t *testing.T) {
	service, workbenchRepo, _, _ := newTestWorkbenchService()
	ctx := context.Background()

	// Setup: create a workbench
	workbenchRepo.workbenches["BENCH-001"] = &secondary.WorkbenchRecord{
		ID:         "BENCH-001",
		Name:       "test-bench",
		WorkshopID: "WORK-001",
		Status:     "active",
	}

	workbench, err := service.GetWorkbench(ctx, "BENCH-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if workbench.Name != "test-bench" {
		t.Errorf("expected name 'test-bench', got '%s'", workbench.Name)
	}
}

func TestWorkbenchService_GetWorkbench_NotFound(t *testing.T) {
	service, _, _, _ := newTestWorkbenchService()
	ctx := context.Background()

	_, err := service.GetWorkbench(ctx, "BENCH-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent workbench, got nil")
	}
}

// ============================================================================
// GetWorkbenchByPath Tests
// ============================================================================

func TestWorkbenchService_GetWorkbenchByPath(t *testing.T) {
	service, workbenchRepo, _, _ := newTestWorkbenchService()
	ctx := context.Background()

	// Setup: create a workbench
	workbenchRepo.workbenches["BENCH-001"] = &secondary.WorkbenchRecord{
		ID:           "BENCH-001",
		Name:         "test-bench",
		WorkshopID:   "WORK-001",
		WorktreePath: "/home/user/wb/test-bench",
		Status:       "active",
	}

	workbench, err := service.GetWorkbenchByPath(ctx, "/home/user/wb/test-bench")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if workbench.ID != "BENCH-001" {
		t.Errorf("expected ID 'BENCH-001', got '%s'", workbench.ID)
	}
}

func TestWorkbenchService_GetWorkbenchByPath_NotFound(t *testing.T) {
	service, _, _, _ := newTestWorkbenchService()
	ctx := context.Background()

	_, err := service.GetWorkbenchByPath(ctx, "/nonexistent/path")

	if err == nil {
		t.Fatal("expected error for non-existent path, got nil")
	}
}

// ============================================================================
// UpdateWorkbenchPath Tests
// ============================================================================

func TestWorkbenchService_UpdateWorkbenchPath(t *testing.T) {
	service, workbenchRepo, _, _ := newTestWorkbenchService()
	ctx := context.Background()

	// Setup: create a workbench
	workbenchRepo.workbenches["BENCH-001"] = &secondary.WorkbenchRecord{
		ID:           "BENCH-001",
		Name:         "test-bench",
		WorktreePath: "/old/path",
		Status:       "active",
	}

	err := service.UpdateWorkbenchPath(ctx, "BENCH-001", "/new/path")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if workbenchRepo.workbenches["BENCH-001"].WorktreePath != "/new/path" {
		t.Errorf("expected path '/new/path', got '%s'", workbenchRepo.workbenches["BENCH-001"].WorktreePath)
	}
}

func TestWorkbenchService_UpdateWorkbenchPath_NotFound(t *testing.T) {
	service, _, _, _ := newTestWorkbenchService()
	ctx := context.Background()

	err := service.UpdateWorkbenchPath(ctx, "BENCH-NONEXISTENT", "/new/path")

	if err == nil {
		t.Fatal("expected error for non-existent workbench, got nil")
	}
}

// ============================================================================
// ListWorkbenches Tests
// ============================================================================

func TestWorkbenchService_ListWorkbenches(t *testing.T) {
	service, workbenchRepo, _, _ := newTestWorkbenchService()
	ctx := context.Background()

	// Setup: create workbenches
	workbenchRepo.workbenches["BENCH-001"] = &secondary.WorkbenchRecord{
		ID:         "BENCH-001",
		Name:       "bench-1",
		WorkshopID: "WORK-001",
		Status:     "active",
	}
	workbenchRepo.workbenches["BENCH-002"] = &secondary.WorkbenchRecord{
		ID:         "BENCH-002",
		Name:       "bench-2",
		WorkshopID: "WORK-001",
		Status:     "active",
	}

	workbenches, err := service.ListWorkbenches(ctx, primary.WorkbenchFilters{})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(workbenches) != 2 {
		t.Errorf("expected 2 workbenches, got %d", len(workbenches))
	}
}

func TestWorkbenchService_ListWorkbenches_FilterByWorkshop(t *testing.T) {
	service, workbenchRepo, _, _ := newTestWorkbenchService()
	ctx := context.Background()

	// Setup: create workbenches in different workshops
	workbenchRepo.workbenches["BENCH-001"] = &secondary.WorkbenchRecord{
		ID:         "BENCH-001",
		Name:       "bench-1",
		WorkshopID: "WORK-001",
		Status:     "active",
	}
	workbenchRepo.workbenches["BENCH-002"] = &secondary.WorkbenchRecord{
		ID:         "BENCH-002",
		Name:       "bench-2",
		WorkshopID: "WORK-002",
		Status:     "active",
	}

	workbenches, err := service.ListWorkbenches(ctx, primary.WorkbenchFilters{WorkshopID: "WORK-001"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(workbenches) != 1 {
		t.Errorf("expected 1 workbench for WORK-001, got %d", len(workbenches))
	}
}

// ============================================================================
// RenameWorkbench Tests
// ============================================================================

func TestWorkbenchService_RenameWorkbench(t *testing.T) {
	service, workbenchRepo, _, _ := newTestWorkbenchService()
	ctx := context.Background()

	// Setup: create a workbench
	workbenchRepo.workbenches["BENCH-001"] = &secondary.WorkbenchRecord{
		ID:     "BENCH-001",
		Name:   "old-name",
		Status: "active",
	}

	err := service.RenameWorkbench(ctx, primary.RenameWorkbenchRequest{
		WorkbenchID: "BENCH-001",
		NewName:     "new-name",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if workbenchRepo.workbenches["BENCH-001"].Name != "new-name" {
		t.Errorf("expected name 'new-name', got '%s'", workbenchRepo.workbenches["BENCH-001"].Name)
	}
}

func TestWorkbenchService_RenameWorkbench_NotFound(t *testing.T) {
	service, _, _, _ := newTestWorkbenchService()
	ctx := context.Background()

	err := service.RenameWorkbench(ctx, primary.RenameWorkbenchRequest{
		WorkbenchID: "BENCH-NONEXISTENT",
		NewName:     "new-name",
	})

	if err == nil {
		t.Fatal("expected error for non-existent workbench, got nil")
	}
}

// ============================================================================
// DeleteWorkbench Tests
// ============================================================================

func TestWorkbenchService_DeleteWorkbench(t *testing.T) {
	service, workbenchRepo, _, _ := newTestWorkbenchService()
	ctx := context.Background()

	// Setup: create a workbench
	workbenchRepo.workbenches["BENCH-001"] = &secondary.WorkbenchRecord{
		ID:     "BENCH-001",
		Name:   "test-bench",
		Status: "active",
	}

	err := service.DeleteWorkbench(ctx, primary.DeleteWorkbenchRequest{
		WorkbenchID: "BENCH-001",
		Force:       false,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, exists := workbenchRepo.workbenches["BENCH-001"]; exists {
		t.Error("expected workbench to be deleted")
	}
}

func TestWorkbenchService_DeleteWorkbench_NotFound(t *testing.T) {
	service, _, _, _ := newTestWorkbenchService()
	ctx := context.Background()

	err := service.DeleteWorkbench(ctx, primary.DeleteWorkbenchRequest{
		WorkbenchID: "BENCH-NONEXISTENT",
		Force:       false,
	})

	if err == nil {
		t.Fatal("expected error for non-existent workbench, got nil")
	}
}

// ============================================================================
// Effect Executor Verification
// ============================================================================

func TestWorkbenchService_CreateWorkbench_ExecutesEffects(t *testing.T) {
	service, workbenchRepo, _, executor := newTestWorkbenchService()
	ctx := context.Background()

	// Setup: workshop exists
	workbenchRepo.workshopExists["WORK-001"] = true

	_, err := service.CreateWorkbench(ctx, primary.CreateWorkbenchRequest{
		Name:       "test-bench",
		WorkshopID: "WORK-001",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should have executed file effects for .orc/config.json
	if len(executor.executedEffects) == 0 {
		t.Error("expected effects to be executed")
	}

	// Check for mkdir and write effects
	var hasMkdir, hasWrite bool
	for _, eff := range executor.executedEffects {
		if fe, ok := eff.(effects.FileEffect); ok {
			if fe.Operation == "mkdir" {
				hasMkdir = true
			}
			if fe.Operation == "write" {
				hasWrite = true
			}
		}
	}

	if !hasMkdir {
		t.Error("expected mkdir effect for .orc directory")
	}
	if !hasWrite {
		t.Error("expected write effect for config.json")
	}
}
