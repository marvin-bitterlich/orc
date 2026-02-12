package app

import (
	"context"
	"errors"
	"testing"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// ============================================================================
// Mock Implementations for Workshop Service
// ============================================================================

// mockWorkshopRepository implements secondary.WorkshopRepository for testing.
type mockWorkshopRepository struct {
	workshops      map[string]*secondary.WorkshopRecord
	workbenchCount map[string]int
	factoryExists  map[string]bool
	nextID         string
	createErr      error
	getErr         error
	updateErr      error
	deleteErr      error
	listErr        error
}

func newMockWorkshopRepository() *mockWorkshopRepository {
	return &mockWorkshopRepository{
		workshops:      make(map[string]*secondary.WorkshopRecord),
		workbenchCount: make(map[string]int),
		factoryExists:  make(map[string]bool),
		nextID:         "WORK-001",
	}
}

func (m *mockWorkshopRepository) Create(ctx context.Context, workshop *secondary.WorkshopRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	workshop.ID = m.nextID
	m.workshops[workshop.ID] = workshop
	return nil
}

func (m *mockWorkshopRepository) GetByID(ctx context.Context, id string) (*secondary.WorkshopRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if workshop, ok := m.workshops[id]; ok {
		return workshop, nil
	}
	return nil, errors.New("workshop not found")
}

func (m *mockWorkshopRepository) List(ctx context.Context, filters secondary.WorkshopFilters) ([]*secondary.WorkshopRecord, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*secondary.WorkshopRecord
	for _, ws := range m.workshops {
		if filters.FactoryID != "" && ws.FactoryID != filters.FactoryID {
			continue
		}
		if filters.Status != "" && ws.Status != filters.Status {
			continue
		}
		result = append(result, ws)
	}
	return result, nil
}

func (m *mockWorkshopRepository) Update(ctx context.Context, workshop *secondary.WorkshopRecord) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.workshops[workshop.ID]; !ok {
		return errors.New("workshop not found")
	}
	// Merge only non-empty fields
	existing := m.workshops[workshop.ID]
	if workshop.Name != "" {
		existing.Name = workshop.Name
	}
	if workshop.Status != "" {
		existing.Status = workshop.Status
	}
	return nil
}

func (m *mockWorkshopRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.workshops[id]; !ok {
		return errors.New("workshop not found")
	}
	delete(m.workshops, id)
	return nil
}

func (m *mockWorkshopRepository) GetNextID(ctx context.Context) (string, error) {
	return m.nextID, nil
}

func (m *mockWorkshopRepository) CountWorkbenches(ctx context.Context, workshopID string) (int, error) {
	return m.workbenchCount[workshopID], nil
}

func (m *mockWorkshopRepository) CountByFactory(ctx context.Context, factoryID string) (int, error) {
	count := 0
	for _, ws := range m.workshops {
		if ws.FactoryID == factoryID {
			count++
		}
	}
	return count, nil
}

func (m *mockWorkshopRepository) FactoryExists(ctx context.Context, factoryID string) (bool, error) {
	return m.factoryExists[factoryID], nil
}

func (m *mockWorkshopRepository) SetActiveCommissionID(ctx context.Context, workshopID, commissionID string) error {
	if ws, ok := m.workshops[workshopID]; ok {
		ws.ActiveCommissionID = commissionID
		return nil
	}
	return errors.New("workshop not found")
}

func (m *mockWorkshopRepository) GetActiveCommissions(ctx context.Context, workshopID string) ([]string, error) {
	return nil, nil
}

// mockFactoryRepository implements secondary.FactoryRepository for testing.
type mockFactoryRepository struct {
	factories    map[string]*secondary.FactoryRecord
	nextID       string
	getByNameErr error
}

func newMockFactoryRepository() *mockFactoryRepository {
	return &mockFactoryRepository{
		factories: make(map[string]*secondary.FactoryRecord),
		nextID:    "FACT-001",
	}
}

func (m *mockFactoryRepository) Create(ctx context.Context, factory *secondary.FactoryRecord) error {
	m.factories[factory.ID] = factory
	return nil
}

func (m *mockFactoryRepository) GetByID(ctx context.Context, id string) (*secondary.FactoryRecord, error) {
	if factory, ok := m.factories[id]; ok {
		return factory, nil
	}
	return nil, errors.New("factory not found")
}

func (m *mockFactoryRepository) GetByName(ctx context.Context, name string) (*secondary.FactoryRecord, error) {
	if m.getByNameErr != nil {
		return nil, m.getByNameErr
	}
	for _, factory := range m.factories {
		if factory.Name == name {
			return factory, nil
		}
	}
	return nil, errors.New("factory not found")
}

func (m *mockFactoryRepository) List(ctx context.Context, filters secondary.FactoryFilters) ([]*secondary.FactoryRecord, error) {
	var result []*secondary.FactoryRecord
	for _, factory := range m.factories {
		result = append(result, factory)
	}
	return result, nil
}

func (m *mockFactoryRepository) Update(ctx context.Context, factory *secondary.FactoryRecord) error {
	m.factories[factory.ID] = factory
	return nil
}

func (m *mockFactoryRepository) Delete(ctx context.Context, id string) error {
	delete(m.factories, id)
	return nil
}

func (m *mockFactoryRepository) GetNextID(ctx context.Context) (string, error) {
	return m.nextID, nil
}

func (m *mockFactoryRepository) CountWorkshops(ctx context.Context, factoryID string) (int, error) {
	return 0, nil
}

func (m *mockFactoryRepository) CountCommissions(ctx context.Context, factoryID string) (int, error) {
	return 0, nil
}

// mockWorkbenchRepositoryForWorkshop implements secondary.WorkbenchRepository minimally.
type mockWorkbenchRepositoryForWorkshop struct {
	workbenches map[string]*secondary.WorkbenchRecord
}

func newMockWorkbenchRepositoryForWorkshop() *mockWorkbenchRepositoryForWorkshop {
	return &mockWorkbenchRepositoryForWorkshop{
		workbenches: make(map[string]*secondary.WorkbenchRecord),
	}
}

func (m *mockWorkbenchRepositoryForWorkshop) Create(ctx context.Context, workbench *secondary.WorkbenchRecord) error {
	m.workbenches[workbench.ID] = workbench
	return nil
}

func (m *mockWorkbenchRepositoryForWorkshop) GetByID(ctx context.Context, id string) (*secondary.WorkbenchRecord, error) {
	if wb, ok := m.workbenches[id]; ok {
		return wb, nil
	}
	return nil, errors.New("workbench not found")
}

func (m *mockWorkbenchRepositoryForWorkshop) GetByPath(ctx context.Context, path string) (*secondary.WorkbenchRecord, error) {
	return nil, errors.New("not implemented")
}

func (m *mockWorkbenchRepositoryForWorkshop) GetByWorkshop(ctx context.Context, workshopID string) ([]*secondary.WorkbenchRecord, error) {
	var result []*secondary.WorkbenchRecord
	for _, wb := range m.workbenches {
		if wb.WorkshopID == workshopID {
			result = append(result, wb)
		}
	}
	return result, nil
}

func (m *mockWorkbenchRepositoryForWorkshop) List(ctx context.Context, workshopID string) ([]*secondary.WorkbenchRecord, error) {
	var result []*secondary.WorkbenchRecord
	for _, wb := range m.workbenches {
		if workshopID == "" || wb.WorkshopID == workshopID {
			result = append(result, wb)
		}
	}
	return result, nil
}

func (m *mockWorkbenchRepositoryForWorkshop) Update(ctx context.Context, workbench *secondary.WorkbenchRecord) error {
	m.workbenches[workbench.ID] = workbench
	return nil
}

func (m *mockWorkbenchRepositoryForWorkshop) Delete(ctx context.Context, id string) error {
	delete(m.workbenches, id)
	return nil
}

func (m *mockWorkbenchRepositoryForWorkshop) Rename(ctx context.Context, id, newName string) error {
	return nil
}

func (m *mockWorkbenchRepositoryForWorkshop) UpdatePath(ctx context.Context, id, newPath string) error {
	return nil
}

func (m *mockWorkbenchRepositoryForWorkshop) GetNextID(ctx context.Context) (string, error) {
	return "BENCH-001", nil
}

func (m *mockWorkbenchRepositoryForWorkshop) WorkshopExists(ctx context.Context, workshopID string) (bool, error) {
	return true, nil
}

func (m *mockWorkbenchRepositoryForWorkshop) UpdateFocusedID(ctx context.Context, id, focusedID string) error {
	if wb, ok := m.workbenches[id]; ok {
		wb.FocusedID = focusedID
		return nil
	}
	return errors.New("workbench not found")
}

func (m *mockWorkbenchRepositoryForWorkshop) GetByFocusedID(ctx context.Context, focusedID string) ([]*secondary.WorkbenchRecord, error) {
	if focusedID == "" {
		return nil, nil
	}
	var result []*secondary.WorkbenchRecord
	for _, wb := range m.workbenches {
		if wb.FocusedID == focusedID && wb.Status == "active" {
			result = append(result, wb)
		}
	}
	return result, nil
}

// mockRepoRepositoryForWorkshop implements secondary.RepoRepository minimally.
type mockRepoRepositoryForWorkshop struct{}

func newMockRepoRepositoryForWorkshop() *mockRepoRepositoryForWorkshop {
	return &mockRepoRepositoryForWorkshop{}
}

func (m *mockRepoRepositoryForWorkshop) Create(ctx context.Context, repo *secondary.RepoRecord) error {
	return nil
}

func (m *mockRepoRepositoryForWorkshop) GetByID(ctx context.Context, id string) (*secondary.RepoRecord, error) {
	return nil, errors.New("not implemented")
}

func (m *mockRepoRepositoryForWorkshop) GetByName(ctx context.Context, name string) (*secondary.RepoRecord, error) {
	return nil, errors.New("not implemented")
}

func (m *mockRepoRepositoryForWorkshop) List(ctx context.Context, filters secondary.RepoFilters) ([]*secondary.RepoRecord, error) {
	return nil, nil
}

func (m *mockRepoRepositoryForWorkshop) Update(ctx context.Context, repo *secondary.RepoRecord) error {
	return nil
}

func (m *mockRepoRepositoryForWorkshop) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockRepoRepositoryForWorkshop) GetNextID(ctx context.Context) (string, error) {
	return "REPO-001", nil
}

func (m *mockRepoRepositoryForWorkshop) UpdateStatus(ctx context.Context, id, status string) error {
	return nil
}

func (m *mockRepoRepositoryForWorkshop) HasActivePRs(ctx context.Context, repoID string) (bool, error) {
	return false, nil
}

// mockTMuxAdapter implements secondary.TMuxAdapter for testing.
type mockTMuxAdapter struct {
	sessions       map[string]bool
	killSessionErr error
}

func newMockTMuxAdapter() *mockTMuxAdapter {
	return &mockTMuxAdapter{
		sessions: make(map[string]bool),
	}
}

func (m *mockTMuxAdapter) SessionExists(ctx context.Context, name string) bool {
	return m.sessions[name]
}

func (m *mockTMuxAdapter) KillSession(ctx context.Context, name string) error {
	if m.killSessionErr != nil {
		return m.killSessionErr
	}
	delete(m.sessions, name)
	return nil
}

func (m *mockTMuxAdapter) GetSessionInfo(ctx context.Context, name string) (string, error) {
	return "", nil
}

func (m *mockTMuxAdapter) WindowExists(ctx context.Context, sessionName string, windowName string) bool {
	return false
}

func (m *mockTMuxAdapter) KillWindow(ctx context.Context, sessionName string, windowName string) error {
	return nil
}

func (m *mockTMuxAdapter) SendKeys(ctx context.Context, target, keys string) error {
	return nil
}

func (m *mockTMuxAdapter) GetPaneCount(ctx context.Context, sessionName, windowName string) int {
	return 0
}

func (m *mockTMuxAdapter) GetPaneCommand(ctx context.Context, sessionName, windowName string, paneNum int) string {
	return ""
}

func (m *mockTMuxAdapter) GetPaneStartPath(ctx context.Context, sessionName, windowName string, paneNum int) string {
	return ""
}

func (m *mockTMuxAdapter) GetPaneStartCommand(ctx context.Context, sessionName, windowName string, paneNum int) string {
	return ""
}

func (m *mockTMuxAdapter) CapturePaneContent(ctx context.Context, target string, lines int) (string, error) {
	return "", nil
}

func (m *mockTMuxAdapter) SplitVertical(ctx context.Context, target, workingDir string) error {
	return nil
}

func (m *mockTMuxAdapter) SplitHorizontal(ctx context.Context, target, workingDir string) error {
	return nil
}

func (m *mockTMuxAdapter) AttachInstructions(sessionName string) string {
	return "tmux attach -t " + sessionName
}

func (m *mockTMuxAdapter) SelectWindow(ctx context.Context, sessionName string, index int) error {
	return nil
}

func (m *mockTMuxAdapter) RenameWindow(ctx context.Context, target, newName string) error {
	return nil
}

func (m *mockTMuxAdapter) RespawnPane(ctx context.Context, target string, command ...string) error {
	return nil
}

func (m *mockTMuxAdapter) RenameSession(ctx context.Context, session, newName string) error {
	return nil
}

func (m *mockTMuxAdapter) ConfigureStatusBar(ctx context.Context, session string, config secondary.StatusBarConfig) error {
	return nil
}

func (m *mockTMuxAdapter) DisplayPopup(ctx context.Context, session, command string, config secondary.PopupConfig) error {
	return nil
}

func (m *mockTMuxAdapter) ConfigureSessionBindings(ctx context.Context, session string, bindings []secondary.KeyBinding) error {
	return nil
}

func (m *mockTMuxAdapter) ConfigureSessionPopupBindings(ctx context.Context, session string, bindings []secondary.PopupKeyBinding) error {
	return nil
}

func (m *mockTMuxAdapter) GetCurrentSessionName(ctx context.Context) string {
	return ""
}

func (m *mockTMuxAdapter) SetEnvironment(ctx context.Context, sessionName, key, value string) error {
	return nil
}

func (m *mockTMuxAdapter) GetEnvironment(ctx context.Context, sessionName, key string) (string, error) {
	return "", nil
}

func (m *mockTMuxAdapter) ListSessions(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (m *mockTMuxAdapter) FindSessionByWorkshopID(ctx context.Context, workshopID string) string {
	return ""
}

func (m *mockTMuxAdapter) ListWindows(ctx context.Context, sessionName string) ([]string, error) {
	return nil, nil
}

func (m *mockTMuxAdapter) JoinPane(ctx context.Context, source, target string, vertical bool, size int) error {
	return nil
}
func (m *mockTMuxAdapter) GetWindowOption(ctx context.Context, target, option string) string {
	return ""
}
func (m *mockTMuxAdapter) SetWindowOption(ctx context.Context, target, option, value string) error {
	return nil
}
func (m *mockTMuxAdapter) SetupGoblinPane(ctx context.Context, sessionName, windowName string) error {
	return nil
}

// ============================================================================
// Test Helper
// ============================================================================

func newTestWorkshopService() (*WorkshopServiceImpl, *mockWorkshopRepository, *mockFactoryRepository, *mockTMuxAdapter) {
	factoryRepo := newMockFactoryRepository()
	workshopRepo := newMockWorkshopRepository()
	workbenchRepo := newMockWorkbenchRepositoryForWorkshop()
	repoRepo := newMockRepoRepositoryForWorkshop()
	tmuxAdapter := newMockTMuxAdapter()
	workspaceAdapter := newMockWorkspaceAdapter()
	executor := newMockEffectExecutor()

	service := NewWorkshopService(
		factoryRepo,
		workshopRepo,
		workbenchRepo,
		repoRepo,
		tmuxAdapter,
		workspaceAdapter,
		executor,
	)
	return service, workshopRepo, factoryRepo, tmuxAdapter
}

// ============================================================================
// CreateWorkshop Tests
// ============================================================================

func TestWorkshopService_CreateWorkshop(t *testing.T) {
	service, workshopRepo, _, _ := newTestWorkshopService()
	ctx := context.Background()

	// Setup: factory exists
	workshopRepo.factoryExists["FACT-001"] = true

	resp, err := service.CreateWorkshop(ctx, primary.CreateWorkshopRequest{
		FactoryID: "FACT-001",
		Name:      "test-workshop",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.WorkshopID == "" {
		t.Error("expected workshop ID to be set")
	}
	if resp.Workshop.Name != "test-workshop" {
		t.Errorf("expected name 'test-workshop', got '%s'", resp.Workshop.Name)
	}
}

func TestWorkshopService_CreateWorkshop_FactoryNotFound(t *testing.T) {
	service, _, _, _ := newTestWorkshopService()
	ctx := context.Background()

	// No factory exists
	_, err := service.CreateWorkshop(ctx, primary.CreateWorkshopRequest{
		FactoryID: "FACT-999",
		Name:      "test-workshop",
	})

	if err == nil {
		t.Fatal("expected error for non-existent factory, got nil")
	}
}

func TestWorkshopService_CreateWorkshop_DefaultFactory(t *testing.T) {
	service, workshopRepo, factoryRepo, _ := newTestWorkshopService()
	ctx := context.Background()

	// No factory specified - should create default
	factoryRepo.getByNameErr = errors.New("not found") // Force default factory creation
	workshopRepo.factoryExists["FACT-001"] = true      // After creation, it exists

	resp, err := service.CreateWorkshop(ctx, primary.CreateWorkshopRequest{
		Name: "test-workshop",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.WorkshopID == "" {
		t.Error("expected workshop ID to be set")
	}
	// Verify default factory was created
	if _, exists := factoryRepo.factories["FACT-001"]; !exists {
		t.Error("expected default factory to be created")
	}
}

// ============================================================================
// GetWorkshop Tests
// ============================================================================

func TestWorkshopService_GetWorkshop(t *testing.T) {
	service, workshopRepo, _, _ := newTestWorkshopService()
	ctx := context.Background()

	// Setup: create a workshop
	workshopRepo.workshops["WORK-001"] = &secondary.WorkshopRecord{
		ID:        "WORK-001",
		Name:      "test-workshop",
		FactoryID: "FACT-001",
		Status:    "active",
	}

	workshop, err := service.GetWorkshop(ctx, "WORK-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if workshop.Name != "test-workshop" {
		t.Errorf("expected name 'test-workshop', got '%s'", workshop.Name)
	}
}

func TestWorkshopService_GetWorkshop_NotFound(t *testing.T) {
	service, _, _, _ := newTestWorkshopService()
	ctx := context.Background()

	_, err := service.GetWorkshop(ctx, "WORK-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent workshop, got nil")
	}
}

// ============================================================================
// ListWorkshops Tests
// ============================================================================

func TestWorkshopService_ListWorkshops(t *testing.T) {
	service, workshopRepo, _, _ := newTestWorkshopService()
	ctx := context.Background()

	// Setup: create workshops
	workshopRepo.workshops["WORK-001"] = &secondary.WorkshopRecord{
		ID:        "WORK-001",
		Name:      "workshop-1",
		FactoryID: "FACT-001",
		Status:    "active",
	}
	workshopRepo.workshops["WORK-002"] = &secondary.WorkshopRecord{
		ID:        "WORK-002",
		Name:      "workshop-2",
		FactoryID: "FACT-001",
		Status:    "active",
	}

	workshops, err := service.ListWorkshops(ctx, primary.WorkshopFilters{})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(workshops) != 2 {
		t.Errorf("expected 2 workshops, got %d", len(workshops))
	}
}

func TestWorkshopService_ListWorkshops_FilterByFactory(t *testing.T) {
	service, workshopRepo, _, _ := newTestWorkshopService()
	ctx := context.Background()

	// Setup: create workshops in different factories
	workshopRepo.workshops["WORK-001"] = &secondary.WorkshopRecord{
		ID:        "WORK-001",
		Name:      "workshop-1",
		FactoryID: "FACT-001",
		Status:    "active",
	}
	workshopRepo.workshops["WORK-002"] = &secondary.WorkshopRecord{
		ID:        "WORK-002",
		Name:      "workshop-2",
		FactoryID: "FACT-002",
		Status:    "active",
	}

	workshops, err := service.ListWorkshops(ctx, primary.WorkshopFilters{FactoryID: "FACT-001"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(workshops) != 1 {
		t.Errorf("expected 1 workshop for FACT-001, got %d", len(workshops))
	}
}

// ============================================================================
// UpdateWorkshop Tests
// ============================================================================

func TestWorkshopService_UpdateWorkshop(t *testing.T) {
	service, workshopRepo, _, _ := newTestWorkshopService()
	ctx := context.Background()

	// Setup: create a workshop
	workshopRepo.workshops["WORK-001"] = &secondary.WorkshopRecord{
		ID:        "WORK-001",
		Name:      "old-name",
		FactoryID: "FACT-001",
		Status:    "active",
	}

	err := service.UpdateWorkshop(ctx, primary.UpdateWorkshopRequest{
		WorkshopID: "WORK-001",
		Name:       "new-name",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if workshopRepo.workshops["WORK-001"].Name != "new-name" {
		t.Errorf("expected name 'new-name', got '%s'", workshopRepo.workshops["WORK-001"].Name)
	}
}

func TestWorkshopService_UpdateWorkshop_NotFound(t *testing.T) {
	service, _, _, _ := newTestWorkshopService()
	ctx := context.Background()

	err := service.UpdateWorkshop(ctx, primary.UpdateWorkshopRequest{
		WorkshopID: "WORK-NONEXISTENT",
		Name:       "new-name",
	})

	if err == nil {
		t.Fatal("expected error for non-existent workshop, got nil")
	}
}

// ============================================================================
// DeleteWorkshop Tests
// ============================================================================

func TestWorkshopService_DeleteWorkshop(t *testing.T) {
	service, workshopRepo, _, _ := newTestWorkshopService()
	ctx := context.Background()

	// Setup: create a workshop with no workbenches
	workshopRepo.workshops["WORK-001"] = &secondary.WorkshopRecord{
		ID:        "WORK-001",
		Name:      "test-workshop",
		FactoryID: "FACT-001",
		Status:    "active",
	}
	workshopRepo.workbenchCount["WORK-001"] = 0

	err := service.DeleteWorkshop(ctx, primary.DeleteWorkshopRequest{
		WorkshopID: "WORK-001",
		Force:      false,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, exists := workshopRepo.workshops["WORK-001"]; exists {
		t.Error("expected workshop to be deleted")
	}
}

func TestWorkshopService_DeleteWorkshop_HasWorkbenches_NoForce(t *testing.T) {
	service, workshopRepo, _, _ := newTestWorkshopService()
	ctx := context.Background()

	// Setup: create a workshop with workbenches
	workshopRepo.workshops["WORK-001"] = &secondary.WorkshopRecord{
		ID:        "WORK-001",
		Name:      "test-workshop",
		FactoryID: "FACT-001",
		Status:    "active",
	}
	workshopRepo.workbenchCount["WORK-001"] = 3

	err := service.DeleteWorkshop(ctx, primary.DeleteWorkshopRequest{
		WorkshopID: "WORK-001",
		Force:      false,
	})

	if err == nil {
		t.Fatal("expected error for workshop with workbenches, got nil")
	}
	if _, exists := workshopRepo.workshops["WORK-001"]; !exists {
		t.Error("workshop should not have been deleted")
	}
}

func TestWorkshopService_DeleteWorkshop_HasWorkbenches_Force(t *testing.T) {
	service, workshopRepo, _, _ := newTestWorkshopService()
	ctx := context.Background()

	// Setup: create a workshop with workbenches
	workshopRepo.workshops["WORK-001"] = &secondary.WorkshopRecord{
		ID:        "WORK-001",
		Name:      "test-workshop",
		FactoryID: "FACT-001",
		Status:    "active",
	}
	workshopRepo.workbenchCount["WORK-001"] = 3

	err := service.DeleteWorkshop(ctx, primary.DeleteWorkshopRequest{
		WorkshopID: "WORK-001",
		Force:      true,
	})

	if err != nil {
		t.Fatalf("expected no error with force, got %v", err)
	}
	if _, exists := workshopRepo.workshops["WORK-001"]; exists {
		t.Error("expected workshop to be deleted with force")
	}
}

func TestWorkshopService_DeleteWorkshop_NotFound(t *testing.T) {
	service, _, _, _ := newTestWorkshopService()
	ctx := context.Background()

	err := service.DeleteWorkshop(ctx, primary.DeleteWorkshopRequest{
		WorkshopID: "WORK-NONEXISTENT",
		Force:      false,
	})

	if err == nil {
		t.Fatal("expected error for non-existent workshop, got nil")
	}
}

// ============================================================================
// CloseWorkshop Tests
// ============================================================================

func TestWorkshopService_CloseWorkshop(t *testing.T) {
	service, _, _, tmuxAdapter := newTestWorkshopService()
	ctx := context.Background()

	// Setup: session exists
	tmuxAdapter.sessions["WORK-001"] = true

	err := service.CloseWorkshop(ctx, "WORK-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tmuxAdapter.sessions["WORK-001"] {
		t.Error("expected session to be killed")
	}
}

func TestWorkshopService_CloseWorkshop_NoSession(t *testing.T) {
	service, _, _, _ := newTestWorkshopService()
	ctx := context.Background()

	// No session exists - should be no-op
	err := service.CloseWorkshop(ctx, "WORK-001")

	if err != nil {
		t.Fatalf("expected no error for non-existent session, got %v", err)
	}
}
