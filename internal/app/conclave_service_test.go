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

// mockConclaveRepository implements secondary.ConclaveRepository for testing.
type mockConclaveRepository struct {
	conclaves           map[string]*secondary.ConclaveRecord
	createErr           error
	getErr              error
	updateErr           error
	deleteErr           error
	listErr             error
	updateStatusErr     error
	missionExistsResult bool
	missionExistsErr    error
}

func newMockConclaveRepository() *mockConclaveRepository {
	return &mockConclaveRepository{
		conclaves:           make(map[string]*secondary.ConclaveRecord),
		missionExistsResult: true,
	}
}

func (m *mockConclaveRepository) Create(ctx context.Context, conclave *secondary.ConclaveRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.conclaves[conclave.ID] = conclave
	return nil
}

func (m *mockConclaveRepository) GetByID(ctx context.Context, id string) (*secondary.ConclaveRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if conclave, ok := m.conclaves[id]; ok {
		return conclave, nil
	}
	return nil, errors.New("conclave not found")
}

func (m *mockConclaveRepository) List(ctx context.Context, filters secondary.ConclaveFilters) ([]*secondary.ConclaveRecord, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*secondary.ConclaveRecord
	for _, c := range m.conclaves {
		if filters.MissionID != "" && c.MissionID != filters.MissionID {
			continue
		}
		if filters.Status != "" && c.Status != filters.Status {
			continue
		}
		result = append(result, c)
	}
	return result, nil
}

func (m *mockConclaveRepository) Update(ctx context.Context, conclave *secondary.ConclaveRecord) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if existing, ok := m.conclaves[conclave.ID]; ok {
		if conclave.Title != "" {
			existing.Title = conclave.Title
		}
		if conclave.Description != "" {
			existing.Description = conclave.Description
		}
	}
	return nil
}

func (m *mockConclaveRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.conclaves, id)
	return nil
}

func (m *mockConclaveRepository) Pin(ctx context.Context, id string) error {
	if conclave, ok := m.conclaves[id]; ok {
		conclave.Pinned = true
	}
	return nil
}

func (m *mockConclaveRepository) Unpin(ctx context.Context, id string) error {
	if conclave, ok := m.conclaves[id]; ok {
		conclave.Pinned = false
	}
	return nil
}

func (m *mockConclaveRepository) GetNextID(ctx context.Context) (string, error) {
	return "CON-001", nil
}

func (m *mockConclaveRepository) UpdateStatus(ctx context.Context, id, status string, setCompleted bool) error {
	if m.updateStatusErr != nil {
		return m.updateStatusErr
	}
	if conclave, ok := m.conclaves[id]; ok {
		conclave.Status = status
		if setCompleted {
			conclave.CompletedAt = "2026-01-20T10:00:00Z"
		}
	}
	return nil
}

func (m *mockConclaveRepository) GetByGrove(ctx context.Context, groveID string) ([]*secondary.ConclaveRecord, error) {
	var result []*secondary.ConclaveRecord
	for _, c := range m.conclaves {
		if c.AssignedGroveID == groveID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *mockConclaveRepository) MissionExists(ctx context.Context, missionID string) (bool, error) {
	if m.missionExistsErr != nil {
		return false, m.missionExistsErr
	}
	return m.missionExistsResult, nil
}

func (m *mockConclaveRepository) GetTasksByConclave(ctx context.Context, conclaveID string) ([]*secondary.ConclaveTaskRecord, error) {
	return []*secondary.ConclaveTaskRecord{}, nil
}

func (m *mockConclaveRepository) GetQuestionsByConclave(ctx context.Context, conclaveID string) ([]*secondary.ConclaveQuestionRecord, error) {
	return []*secondary.ConclaveQuestionRecord{}, nil
}

func (m *mockConclaveRepository) GetPlansByConclave(ctx context.Context, conclaveID string) ([]*secondary.ConclavePlanRecord, error) {
	return []*secondary.ConclavePlanRecord{}, nil
}

// ============================================================================
// Test Helper
// ============================================================================

func newTestConclaveService() (*ConclaveServiceImpl, *mockConclaveRepository) {
	conclaveRepo := newMockConclaveRepository()
	service := NewConclaveService(conclaveRepo)
	return service, conclaveRepo
}

// ============================================================================
// CreateConclave Tests
// ============================================================================

func TestCreateConclave_Success(t *testing.T) {
	service, _ := newTestConclaveService()
	ctx := context.Background()

	resp, err := service.CreateConclave(ctx, primary.CreateConclaveRequest{
		MissionID:   "MISSION-001",
		Title:       "Test Conclave",
		Description: "A test conclave",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.ConclaveID == "" {
		t.Error("expected conclave ID to be set")
	}
	if resp.Conclave.Title != "Test Conclave" {
		t.Errorf("expected title 'Test Conclave', got '%s'", resp.Conclave.Title)
	}
	if resp.Conclave.Status != "active" {
		t.Errorf("expected status 'active', got '%s'", resp.Conclave.Status)
	}
}

func TestCreateConclave_MissionNotFound(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.missionExistsResult = false

	_, err := service.CreateConclave(ctx, primary.CreateConclaveRequest{
		MissionID:   "MISSION-NONEXISTENT",
		Title:       "Test Conclave",
		Description: "A test conclave",
	})

	if err == nil {
		t.Fatal("expected error for non-existent mission, got nil")
	}
}

// ============================================================================
// GetConclave Tests
// ============================================================================

func TestGetConclave_Found(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.conclaves["CON-001"] = &secondary.ConclaveRecord{
		ID:        "CON-001",
		MissionID: "MISSION-001",
		Title:     "Test Conclave",
		Status:    "active",
	}

	conclave, err := service.GetConclave(ctx, "CON-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if conclave.Title != "Test Conclave" {
		t.Errorf("expected title 'Test Conclave', got '%s'", conclave.Title)
	}
}

func TestGetConclave_NotFound(t *testing.T) {
	service, _ := newTestConclaveService()
	ctx := context.Background()

	_, err := service.GetConclave(ctx, "CON-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent conclave, got nil")
	}
}

// ============================================================================
// ListConclaves Tests
// ============================================================================

func TestListConclaves_FilterByMission(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.conclaves["CON-001"] = &secondary.ConclaveRecord{
		ID:        "CON-001",
		MissionID: "MISSION-001",
		Title:     "Conclave 1",
		Status:    "active",
	}
	conclaveRepo.conclaves["CON-002"] = &secondary.ConclaveRecord{
		ID:        "CON-002",
		MissionID: "MISSION-002",
		Title:     "Conclave 2",
		Status:    "active",
	}

	conclaves, err := service.ListConclaves(ctx, primary.ConclaveFilters{MissionID: "MISSION-001"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(conclaves) != 1 {
		t.Errorf("expected 1 conclave, got %d", len(conclaves))
	}
}

func TestListConclaves_FilterByStatus(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.conclaves["CON-001"] = &secondary.ConclaveRecord{
		ID:        "CON-001",
		MissionID: "MISSION-001",
		Title:     "Active Conclave",
		Status:    "active",
	}
	conclaveRepo.conclaves["CON-002"] = &secondary.ConclaveRecord{
		ID:        "CON-002",
		MissionID: "MISSION-001",
		Title:     "Paused Conclave",
		Status:    "paused",
	}

	conclaves, err := service.ListConclaves(ctx, primary.ConclaveFilters{Status: "active"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(conclaves) != 1 {
		t.Errorf("expected 1 active conclave, got %d", len(conclaves))
	}
}

// ============================================================================
// CompleteConclave Tests
// ============================================================================

func TestCompleteConclave_UnpinnedAllowed(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.conclaves["CON-001"] = &secondary.ConclaveRecord{
		ID:        "CON-001",
		MissionID: "MISSION-001",
		Title:     "Test Conclave",
		Status:    "active",
		Pinned:    false,
	}

	err := service.CompleteConclave(ctx, "CON-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if conclaveRepo.conclaves["CON-001"].Status != "complete" {
		t.Errorf("expected status 'complete', got '%s'", conclaveRepo.conclaves["CON-001"].Status)
	}
}

func TestCompleteConclave_PinnedBlocked(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.conclaves["CON-001"] = &secondary.ConclaveRecord{
		ID:        "CON-001",
		MissionID: "MISSION-001",
		Title:     "Pinned Conclave",
		Status:    "active",
		Pinned:    true,
	}

	err := service.CompleteConclave(ctx, "CON-001")

	if err == nil {
		t.Fatal("expected error for completing pinned conclave, got nil")
	}
}

func TestCompleteConclave_NotFound(t *testing.T) {
	service, _ := newTestConclaveService()
	ctx := context.Background()

	err := service.CompleteConclave(ctx, "CON-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent conclave, got nil")
	}
}

// ============================================================================
// PauseConclave Tests
// ============================================================================

func TestPauseConclave_ActiveAllowed(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.conclaves["CON-001"] = &secondary.ConclaveRecord{
		ID:        "CON-001",
		MissionID: "MISSION-001",
		Title:     "Active Conclave",
		Status:    "active",
	}

	err := service.PauseConclave(ctx, "CON-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if conclaveRepo.conclaves["CON-001"].Status != "paused" {
		t.Errorf("expected status 'paused', got '%s'", conclaveRepo.conclaves["CON-001"].Status)
	}
}

func TestPauseConclave_NotActiveBlocked(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.conclaves["CON-001"] = &secondary.ConclaveRecord{
		ID:        "CON-001",
		MissionID: "MISSION-001",
		Title:     "Paused Conclave",
		Status:    "paused",
	}

	err := service.PauseConclave(ctx, "CON-001")

	if err == nil {
		t.Fatal("expected error for pausing non-active conclave, got nil")
	}
}

func TestPauseConclave_CompleteBlocked(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.conclaves["CON-001"] = &secondary.ConclaveRecord{
		ID:        "CON-001",
		MissionID: "MISSION-001",
		Title:     "Complete Conclave",
		Status:    "complete",
	}

	err := service.PauseConclave(ctx, "CON-001")

	if err == nil {
		t.Fatal("expected error for pausing complete conclave, got nil")
	}
}

// ============================================================================
// ResumeConclave Tests
// ============================================================================

func TestResumeConclave_PausedAllowed(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.conclaves["CON-001"] = &secondary.ConclaveRecord{
		ID:        "CON-001",
		MissionID: "MISSION-001",
		Title:     "Paused Conclave",
		Status:    "paused",
	}

	err := service.ResumeConclave(ctx, "CON-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if conclaveRepo.conclaves["CON-001"].Status != "active" {
		t.Errorf("expected status 'active', got '%s'", conclaveRepo.conclaves["CON-001"].Status)
	}
}

func TestResumeConclave_NotPausedBlocked(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.conclaves["CON-001"] = &secondary.ConclaveRecord{
		ID:        "CON-001",
		MissionID: "MISSION-001",
		Title:     "Active Conclave",
		Status:    "active",
	}

	err := service.ResumeConclave(ctx, "CON-001")

	if err == nil {
		t.Fatal("expected error for resuming non-paused conclave, got nil")
	}
}

// ============================================================================
// Pin/Unpin Tests
// ============================================================================

func TestPinConclave(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.conclaves["CON-001"] = &secondary.ConclaveRecord{
		ID:        "CON-001",
		MissionID: "MISSION-001",
		Title:     "Test Conclave",
		Status:    "active",
		Pinned:    false,
	}

	err := service.PinConclave(ctx, "CON-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !conclaveRepo.conclaves["CON-001"].Pinned {
		t.Error("expected conclave to be pinned")
	}
}

func TestUnpinConclave(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.conclaves["CON-001"] = &secondary.ConclaveRecord{
		ID:        "CON-001",
		MissionID: "MISSION-001",
		Title:     "Pinned Conclave",
		Status:    "active",
		Pinned:    true,
	}

	err := service.UnpinConclave(ctx, "CON-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if conclaveRepo.conclaves["CON-001"].Pinned {
		t.Error("expected conclave to be unpinned")
	}
}

// ============================================================================
// UpdateConclave Tests
// ============================================================================

func TestUpdateConclave_Title(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.conclaves["CON-001"] = &secondary.ConclaveRecord{
		ID:          "CON-001",
		MissionID:   "MISSION-001",
		Title:       "Old Title",
		Description: "Original description",
		Status:      "active",
	}

	err := service.UpdateConclave(ctx, primary.UpdateConclaveRequest{
		ConclaveID: "CON-001",
		Title:      "New Title",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if conclaveRepo.conclaves["CON-001"].Title != "New Title" {
		t.Errorf("expected title 'New Title', got '%s'", conclaveRepo.conclaves["CON-001"].Title)
	}
}

// ============================================================================
// DeleteConclave Tests
// ============================================================================

func TestDeleteConclave_Success(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.conclaves["CON-001"] = &secondary.ConclaveRecord{
		ID:        "CON-001",
		MissionID: "MISSION-001",
		Title:     "Test Conclave",
		Status:    "active",
	}

	err := service.DeleteConclave(ctx, "CON-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, exists := conclaveRepo.conclaves["CON-001"]; exists {
		t.Error("expected conclave to be deleted")
	}
}

// ============================================================================
// GetConclavesByGrove Tests
// ============================================================================

func TestGetConclavesByGrove_Success(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.conclaves["CON-001"] = &secondary.ConclaveRecord{
		ID:              "CON-001",
		MissionID:       "MISSION-001",
		Title:           "Assigned Conclave",
		Status:          "active",
		AssignedGroveID: "GROVE-001",
	}
	conclaveRepo.conclaves["CON-002"] = &secondary.ConclaveRecord{
		ID:        "CON-002",
		MissionID: "MISSION-001",
		Title:     "Unassigned Conclave",
		Status:    "active",
	}

	conclaves, err := service.GetConclavesByGrove(ctx, "GROVE-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(conclaves) != 1 {
		t.Errorf("expected 1 conclave, got %d", len(conclaves))
	}
}

// ============================================================================
// GetConclaveTasks Tests
// ============================================================================

func TestGetConclaveTasks_Success(t *testing.T) {
	service, _ := newTestConclaveService()
	ctx := context.Background()

	tasks, err := service.GetConclaveTasks(ctx, "CON-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// Empty list is valid
	if tasks == nil {
		t.Error("expected non-nil tasks slice")
	}
}

// ============================================================================
// GetConclaveQuestions Tests
// ============================================================================

func TestGetConclaveQuestions_Success(t *testing.T) {
	service, _ := newTestConclaveService()
	ctx := context.Background()

	questions, err := service.GetConclaveQuestions(ctx, "CON-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// Empty list is valid
	if questions == nil {
		t.Error("expected non-nil questions slice")
	}
}

// ============================================================================
// GetConclavePlans Tests
// ============================================================================

func TestGetConclavePlans_Success(t *testing.T) {
	service, _ := newTestConclaveService()
	ctx := context.Background()

	plans, err := service.GetConclavePlans(ctx, "CON-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// Empty list is valid
	if plans == nil {
		t.Error("expected non-nil plans slice")
	}
}
