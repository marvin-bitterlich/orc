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
	conclaves              map[string]*secondary.ConclaveRecord
	createErr              error
	getErr                 error
	updateErr              error
	deleteErr              error
	listErr                error
	updateStatusErr        error
	commissionExistsResult bool
	commissionExistsErr    error
}

func newMockConclaveRepository() *mockConclaveRepository {
	return &mockConclaveRepository{
		conclaves:              make(map[string]*secondary.ConclaveRecord),
		commissionExistsResult: true,
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
		if filters.CommissionID != "" && c.CommissionID != filters.CommissionID {
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

func (m *mockConclaveRepository) UpdateStatus(ctx context.Context, id, status string, setDecided bool) error {
	if m.updateStatusErr != nil {
		return m.updateStatusErr
	}
	if conclave, ok := m.conclaves[id]; ok {
		conclave.Status = status
		if setDecided {
			conclave.DecidedAt = "2026-01-20T10:00:00Z"
		}
	}
	return nil
}

func (m *mockConclaveRepository) CommissionExists(ctx context.Context, commissionID string) (bool, error) {
	if m.commissionExistsErr != nil {
		return false, m.commissionExistsErr
	}
	return m.commissionExistsResult, nil
}

func (m *mockConclaveRepository) GetTasksByConclave(ctx context.Context, conclaveID string) ([]*secondary.ConclaveTaskRecord, error) {
	return []*secondary.ConclaveTaskRecord{}, nil
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
		CommissionID: "COMM-001",
		Title:        "Test Conclave",
		Description:  "A test conclave",
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
	if resp.Conclave.Status != "open" {
		t.Errorf("expected status 'open', got '%s'", resp.Conclave.Status)
	}
}

func TestCreateConclave_MissionNotFound(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.commissionExistsResult = false

	_, err := service.CreateConclave(ctx, primary.CreateConclaveRequest{
		CommissionID: "COMM-NONEXISTENT",
		Title:        "Test Conclave",
		Description:  "A test conclave",
	})

	if err == nil {
		t.Fatal("expected error for non-existent commission, got nil")
	}
}

// ============================================================================
// GetConclave Tests
// ============================================================================

func TestGetConclave_Found(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.conclaves["CON-001"] = &secondary.ConclaveRecord{
		ID:           "CON-001",
		CommissionID: "COMM-001",
		Title:        "Test Conclave",
		Status:       "open",
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
		ID:           "CON-001",
		CommissionID: "COMM-001",
		Title:        "Conclave 1",
		Status:       "open",
	}
	conclaveRepo.conclaves["CON-002"] = &secondary.ConclaveRecord{
		ID:           "CON-002",
		CommissionID: "COMM-002",
		Title:        "Conclave 2",
		Status:       "open",
	}

	conclaves, err := service.ListConclaves(ctx, primary.ConclaveFilters{CommissionID: "COMM-001"})

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
		ID:           "CON-001",
		CommissionID: "COMM-001",
		Title:        "Active Conclave",
		Status:       "open",
	}
	conclaveRepo.conclaves["CON-002"] = &secondary.ConclaveRecord{
		ID:           "CON-002",
		CommissionID: "COMM-001",
		Title:        "Paused Conclave",
		Status:       "paused",
	}

	conclaves, err := service.ListConclaves(ctx, primary.ConclaveFilters{Status: "open"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(conclaves) != 1 {
		t.Errorf("expected 1 open conclave, got %d", len(conclaves))
	}
}

// ============================================================================
// CompleteConclave Tests
// ============================================================================

func TestCompleteConclave_UnpinnedAllowed(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.conclaves["CON-001"] = &secondary.ConclaveRecord{
		ID:           "CON-001",
		CommissionID: "COMM-001",
		Title:        "Test Conclave",
		Status:       "open",
		Pinned:       false,
	}

	err := service.CompleteConclave(ctx, "CON-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if conclaveRepo.conclaves["CON-001"].Status != "closed" {
		t.Errorf("expected status 'closed', got '%s'", conclaveRepo.conclaves["CON-001"].Status)
	}
}

func TestCompleteConclave_PinnedBlocked(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.conclaves["CON-001"] = &secondary.ConclaveRecord{
		ID:           "CON-001",
		CommissionID: "COMM-001",
		Title:        "Pinned Conclave",
		Status:       "open",
		Pinned:       true,
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
		ID:           "CON-001",
		CommissionID: "COMM-001",
		Title:        "Active Conclave",
		Status:       "open",
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
		ID:           "CON-001",
		CommissionID: "COMM-001",
		Title:        "Paused Conclave",
		Status:       "paused",
	}

	err := service.PauseConclave(ctx, "CON-001")

	if err == nil {
		t.Fatal("expected error for pausing non-open conclave, got nil")
	}
}

func TestPauseConclave_CompleteBlocked(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.conclaves["CON-001"] = &secondary.ConclaveRecord{
		ID:           "CON-001",
		CommissionID: "COMM-001",
		Title:        "Complete Conclave",
		Status:       "closed",
	}

	err := service.PauseConclave(ctx, "CON-001")

	if err == nil {
		t.Fatal("expected error for pausing closed conclave, got nil")
	}
}

// ============================================================================
// ResumeConclave Tests
// ============================================================================

func TestResumeConclave_PausedAllowed(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.conclaves["CON-001"] = &secondary.ConclaveRecord{
		ID:           "CON-001",
		CommissionID: "COMM-001",
		Title:        "Paused Conclave",
		Status:       "paused",
	}

	err := service.ResumeConclave(ctx, "CON-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if conclaveRepo.conclaves["CON-001"].Status != "open" {
		t.Errorf("expected status 'open', got '%s'", conclaveRepo.conclaves["CON-001"].Status)
	}
}

func TestResumeConclave_NotPausedBlocked(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.conclaves["CON-001"] = &secondary.ConclaveRecord{
		ID:           "CON-001",
		CommissionID: "COMM-001",
		Title:        "Active Conclave",
		Status:       "open",
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
		ID:           "CON-001",
		CommissionID: "COMM-001",
		Title:        "Test Conclave",
		Status:       "open",
		Pinned:       false,
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
		ID:           "CON-001",
		CommissionID: "COMM-001",
		Title:        "Pinned Conclave",
		Status:       "open",
		Pinned:       true,
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
		ID:           "CON-001",
		CommissionID: "COMM-001",
		Title:        "Old Title",
		Description:  "Original description",
		Status:       "open",
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
		ID:           "CON-001",
		CommissionID: "COMM-001",
		Title:        "Test Conclave",
		Status:       "open",
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
// GetConclavesByShipment Tests
// ============================================================================

func TestGetConclavesByShipment_Success(t *testing.T) {
	service, conclaveRepo := newTestConclaveService()
	ctx := context.Background()

	conclaveRepo.conclaves["CON-001"] = &secondary.ConclaveRecord{
		ID:           "CON-001",
		CommissionID: "COMM-001",
		ShipmentID:   "SHIP-001",
		Title:        "Assigned Conclave",
		Status:       "open",
	}
	conclaveRepo.conclaves["CON-002"] = &secondary.ConclaveRecord{
		ID:           "CON-002",
		CommissionID: "COMM-001",
		Title:        "Unassigned Conclave",
		Status:       "open",
	}

	conclaves, err := service.GetConclavesByShipment(ctx, "SHIP-001")

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
