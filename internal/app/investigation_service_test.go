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

// mockInvestigationRepository implements secondary.InvestigationRepository for testing.
type mockInvestigationRepository struct {
	investigations      map[string]*secondary.InvestigationRecord
	createErr           error
	getErr              error
	updateErr           error
	deleteErr           error
	listErr             error
	updateStatusErr     error
	missionExistsResult bool
	missionExistsErr    error
}

func newMockInvestigationRepository() *mockInvestigationRepository {
	return &mockInvestigationRepository{
		investigations:      make(map[string]*secondary.InvestigationRecord),
		missionExistsResult: true,
	}
}

func (m *mockInvestigationRepository) Create(ctx context.Context, investigation *secondary.InvestigationRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.investigations[investigation.ID] = investigation
	return nil
}

func (m *mockInvestigationRepository) GetByID(ctx context.Context, id string) (*secondary.InvestigationRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if investigation, ok := m.investigations[id]; ok {
		return investigation, nil
	}
	return nil, errors.New("investigation not found")
}

func (m *mockInvestigationRepository) List(ctx context.Context, filters secondary.InvestigationFilters) ([]*secondary.InvestigationRecord, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*secondary.InvestigationRecord
	for _, inv := range m.investigations {
		if filters.MissionID != "" && inv.MissionID != filters.MissionID {
			continue
		}
		if filters.Status != "" && inv.Status != filters.Status {
			continue
		}
		result = append(result, inv)
	}
	return result, nil
}

func (m *mockInvestigationRepository) Update(ctx context.Context, investigation *secondary.InvestigationRecord) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if existing, ok := m.investigations[investigation.ID]; ok {
		if investigation.Title != "" {
			existing.Title = investigation.Title
		}
		if investigation.Description != "" {
			existing.Description = investigation.Description
		}
	}
	return nil
}

func (m *mockInvestigationRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.investigations, id)
	return nil
}

func (m *mockInvestigationRepository) Pin(ctx context.Context, id string) error {
	if investigation, ok := m.investigations[id]; ok {
		investigation.Pinned = true
	}
	return nil
}

func (m *mockInvestigationRepository) Unpin(ctx context.Context, id string) error {
	if investigation, ok := m.investigations[id]; ok {
		investigation.Pinned = false
	}
	return nil
}

func (m *mockInvestigationRepository) GetNextID(ctx context.Context) (string, error) {
	return "INV-001", nil
}

func (m *mockInvestigationRepository) UpdateStatus(ctx context.Context, id, status string, setCompleted bool) error {
	if m.updateStatusErr != nil {
		return m.updateStatusErr
	}
	if investigation, ok := m.investigations[id]; ok {
		investigation.Status = status
		if setCompleted {
			investigation.CompletedAt = "2026-01-20T10:00:00Z"
		}
	}
	return nil
}

func (m *mockInvestigationRepository) GetByGrove(ctx context.Context, groveID string) ([]*secondary.InvestigationRecord, error) {
	var result []*secondary.InvestigationRecord
	for _, inv := range m.investigations {
		if inv.AssignedGroveID == groveID {
			result = append(result, inv)
		}
	}
	return result, nil
}

func (m *mockInvestigationRepository) AssignGrove(ctx context.Context, investigationID, groveID string) error {
	if investigation, ok := m.investigations[investigationID]; ok {
		investigation.AssignedGroveID = groveID
	}
	return nil
}

func (m *mockInvestigationRepository) MissionExists(ctx context.Context, missionID string) (bool, error) {
	if m.missionExistsErr != nil {
		return false, m.missionExistsErr
	}
	return m.missionExistsResult, nil
}

func (m *mockInvestigationRepository) GetQuestionsByInvestigation(ctx context.Context, investigationID string) ([]*secondary.InvestigationQuestionRecord, error) {
	return []*secondary.InvestigationQuestionRecord{}, nil
}

// ============================================================================
// Test Helper
// ============================================================================

func newTestInvestigationService() (*InvestigationServiceImpl, *mockInvestigationRepository) {
	investigationRepo := newMockInvestigationRepository()
	service := NewInvestigationService(investigationRepo)
	return service, investigationRepo
}

// ============================================================================
// CreateInvestigation Tests
// ============================================================================

func TestCreateInvestigation_Success(t *testing.T) {
	service, _ := newTestInvestigationService()
	ctx := context.Background()

	resp, err := service.CreateInvestigation(ctx, primary.CreateInvestigationRequest{
		MissionID:   "MISSION-001",
		Title:       "Test Investigation",
		Description: "A test investigation",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.InvestigationID == "" {
		t.Error("expected investigation ID to be set")
	}
	if resp.Investigation.Title != "Test Investigation" {
		t.Errorf("expected title 'Test Investigation', got '%s'", resp.Investigation.Title)
	}
	if resp.Investigation.Status != "active" {
		t.Errorf("expected status 'active', got '%s'", resp.Investigation.Status)
	}
}

func TestCreateInvestigation_MissionNotFound(t *testing.T) {
	service, investigationRepo := newTestInvestigationService()
	ctx := context.Background()

	investigationRepo.missionExistsResult = false

	_, err := service.CreateInvestigation(ctx, primary.CreateInvestigationRequest{
		MissionID:   "MISSION-NONEXISTENT",
		Title:       "Test Investigation",
		Description: "A test investigation",
	})

	if err == nil {
		t.Fatal("expected error for non-existent mission, got nil")
	}
}

// ============================================================================
// GetInvestigation Tests
// ============================================================================

func TestGetInvestigation_Found(t *testing.T) {
	service, investigationRepo := newTestInvestigationService()
	ctx := context.Background()

	investigationRepo.investigations["INV-001"] = &secondary.InvestigationRecord{
		ID:        "INV-001",
		MissionID: "MISSION-001",
		Title:     "Test Investigation",
		Status:    "active",
	}

	investigation, err := service.GetInvestigation(ctx, "INV-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if investigation.Title != "Test Investigation" {
		t.Errorf("expected title 'Test Investigation', got '%s'", investigation.Title)
	}
}

func TestGetInvestigation_NotFound(t *testing.T) {
	service, _ := newTestInvestigationService()
	ctx := context.Background()

	_, err := service.GetInvestigation(ctx, "INV-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent investigation, got nil")
	}
}

// ============================================================================
// ListInvestigations Tests
// ============================================================================

func TestListInvestigations_FilterByMission(t *testing.T) {
	service, investigationRepo := newTestInvestigationService()
	ctx := context.Background()

	investigationRepo.investigations["INV-001"] = &secondary.InvestigationRecord{
		ID:        "INV-001",
		MissionID: "MISSION-001",
		Title:     "Investigation 1",
		Status:    "active",
	}
	investigationRepo.investigations["INV-002"] = &secondary.InvestigationRecord{
		ID:        "INV-002",
		MissionID: "MISSION-002",
		Title:     "Investigation 2",
		Status:    "active",
	}

	investigations, err := service.ListInvestigations(ctx, primary.InvestigationFilters{MissionID: "MISSION-001"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(investigations) != 1 {
		t.Errorf("expected 1 investigation, got %d", len(investigations))
	}
}

func TestListInvestigations_FilterByStatus(t *testing.T) {
	service, investigationRepo := newTestInvestigationService()
	ctx := context.Background()

	investigationRepo.investigations["INV-001"] = &secondary.InvestigationRecord{
		ID:        "INV-001",
		MissionID: "MISSION-001",
		Title:     "Active Investigation",
		Status:    "active",
	}
	investigationRepo.investigations["INV-002"] = &secondary.InvestigationRecord{
		ID:        "INV-002",
		MissionID: "MISSION-001",
		Title:     "Paused Investigation",
		Status:    "paused",
	}

	investigations, err := service.ListInvestigations(ctx, primary.InvestigationFilters{Status: "active"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(investigations) != 1 {
		t.Errorf("expected 1 active investigation, got %d", len(investigations))
	}
}

// ============================================================================
// CompleteInvestigation Tests
// ============================================================================

func TestCompleteInvestigation_UnpinnedAllowed(t *testing.T) {
	service, investigationRepo := newTestInvestigationService()
	ctx := context.Background()

	investigationRepo.investigations["INV-001"] = &secondary.InvestigationRecord{
		ID:        "INV-001",
		MissionID: "MISSION-001",
		Title:     "Test Investigation",
		Status:    "active",
		Pinned:    false,
	}

	err := service.CompleteInvestigation(ctx, "INV-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if investigationRepo.investigations["INV-001"].Status != "complete" {
		t.Errorf("expected status 'complete', got '%s'", investigationRepo.investigations["INV-001"].Status)
	}
}

func TestCompleteInvestigation_PinnedBlocked(t *testing.T) {
	service, investigationRepo := newTestInvestigationService()
	ctx := context.Background()

	investigationRepo.investigations["INV-001"] = &secondary.InvestigationRecord{
		ID:        "INV-001",
		MissionID: "MISSION-001",
		Title:     "Pinned Investigation",
		Status:    "active",
		Pinned:    true,
	}

	err := service.CompleteInvestigation(ctx, "INV-001")

	if err == nil {
		t.Fatal("expected error for completing pinned investigation, got nil")
	}
}

func TestCompleteInvestigation_NotFound(t *testing.T) {
	service, _ := newTestInvestigationService()
	ctx := context.Background()

	err := service.CompleteInvestigation(ctx, "INV-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent investigation, got nil")
	}
}

// ============================================================================
// PauseInvestigation Tests
// ============================================================================

func TestPauseInvestigation_ActiveAllowed(t *testing.T) {
	service, investigationRepo := newTestInvestigationService()
	ctx := context.Background()

	investigationRepo.investigations["INV-001"] = &secondary.InvestigationRecord{
		ID:        "INV-001",
		MissionID: "MISSION-001",
		Title:     "Active Investigation",
		Status:    "active",
	}

	err := service.PauseInvestigation(ctx, "INV-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if investigationRepo.investigations["INV-001"].Status != "paused" {
		t.Errorf("expected status 'paused', got '%s'", investigationRepo.investigations["INV-001"].Status)
	}
}

func TestPauseInvestigation_NotActiveBlocked(t *testing.T) {
	service, investigationRepo := newTestInvestigationService()
	ctx := context.Background()

	investigationRepo.investigations["INV-001"] = &secondary.InvestigationRecord{
		ID:        "INV-001",
		MissionID: "MISSION-001",
		Title:     "Paused Investigation",
		Status:    "paused",
	}

	err := service.PauseInvestigation(ctx, "INV-001")

	if err == nil {
		t.Fatal("expected error for pausing non-active investigation, got nil")
	}
}

func TestPauseInvestigation_CompleteBlocked(t *testing.T) {
	service, investigationRepo := newTestInvestigationService()
	ctx := context.Background()

	investigationRepo.investigations["INV-001"] = &secondary.InvestigationRecord{
		ID:        "INV-001",
		MissionID: "MISSION-001",
		Title:     "Complete Investigation",
		Status:    "complete",
	}

	err := service.PauseInvestigation(ctx, "INV-001")

	if err == nil {
		t.Fatal("expected error for pausing complete investigation, got nil")
	}
}

// ============================================================================
// ResumeInvestigation Tests
// ============================================================================

func TestResumeInvestigation_PausedAllowed(t *testing.T) {
	service, investigationRepo := newTestInvestigationService()
	ctx := context.Background()

	investigationRepo.investigations["INV-001"] = &secondary.InvestigationRecord{
		ID:        "INV-001",
		MissionID: "MISSION-001",
		Title:     "Paused Investigation",
		Status:    "paused",
	}

	err := service.ResumeInvestigation(ctx, "INV-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if investigationRepo.investigations["INV-001"].Status != "active" {
		t.Errorf("expected status 'active', got '%s'", investigationRepo.investigations["INV-001"].Status)
	}
}

func TestResumeInvestigation_NotPausedBlocked(t *testing.T) {
	service, investigationRepo := newTestInvestigationService()
	ctx := context.Background()

	investigationRepo.investigations["INV-001"] = &secondary.InvestigationRecord{
		ID:        "INV-001",
		MissionID: "MISSION-001",
		Title:     "Active Investigation",
		Status:    "active",
	}

	err := service.ResumeInvestigation(ctx, "INV-001")

	if err == nil {
		t.Fatal("expected error for resuming non-paused investigation, got nil")
	}
}

// ============================================================================
// Pin/Unpin Tests
// ============================================================================

func TestPinInvestigation(t *testing.T) {
	service, investigationRepo := newTestInvestigationService()
	ctx := context.Background()

	investigationRepo.investigations["INV-001"] = &secondary.InvestigationRecord{
		ID:        "INV-001",
		MissionID: "MISSION-001",
		Title:     "Test Investigation",
		Status:    "active",
		Pinned:    false,
	}

	err := service.PinInvestigation(ctx, "INV-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !investigationRepo.investigations["INV-001"].Pinned {
		t.Error("expected investigation to be pinned")
	}
}

func TestUnpinInvestigation(t *testing.T) {
	service, investigationRepo := newTestInvestigationService()
	ctx := context.Background()

	investigationRepo.investigations["INV-001"] = &secondary.InvestigationRecord{
		ID:        "INV-001",
		MissionID: "MISSION-001",
		Title:     "Pinned Investigation",
		Status:    "active",
		Pinned:    true,
	}

	err := service.UnpinInvestigation(ctx, "INV-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if investigationRepo.investigations["INV-001"].Pinned {
		t.Error("expected investigation to be unpinned")
	}
}

// ============================================================================
// UpdateInvestigation Tests
// ============================================================================

func TestUpdateInvestigation_Title(t *testing.T) {
	service, investigationRepo := newTestInvestigationService()
	ctx := context.Background()

	investigationRepo.investigations["INV-001"] = &secondary.InvestigationRecord{
		ID:          "INV-001",
		MissionID:   "MISSION-001",
		Title:       "Old Title",
		Description: "Original description",
		Status:      "active",
	}

	err := service.UpdateInvestigation(ctx, primary.UpdateInvestigationRequest{
		InvestigationID: "INV-001",
		Title:           "New Title",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if investigationRepo.investigations["INV-001"].Title != "New Title" {
		t.Errorf("expected title 'New Title', got '%s'", investigationRepo.investigations["INV-001"].Title)
	}
}

// ============================================================================
// DeleteInvestigation Tests
// ============================================================================

func TestDeleteInvestigation_Success(t *testing.T) {
	service, investigationRepo := newTestInvestigationService()
	ctx := context.Background()

	investigationRepo.investigations["INV-001"] = &secondary.InvestigationRecord{
		ID:        "INV-001",
		MissionID: "MISSION-001",
		Title:     "Test Investigation",
		Status:    "active",
	}

	err := service.DeleteInvestigation(ctx, "INV-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, exists := investigationRepo.investigations["INV-001"]; exists {
		t.Error("expected investigation to be deleted")
	}
}

// ============================================================================
// AssignInvestigationToGrove Tests
// ============================================================================

func TestAssignInvestigationToGrove_Success(t *testing.T) {
	service, investigationRepo := newTestInvestigationService()
	ctx := context.Background()

	investigationRepo.investigations["INV-001"] = &secondary.InvestigationRecord{
		ID:        "INV-001",
		MissionID: "MISSION-001",
		Title:     "Test Investigation",
		Status:    "active",
	}

	err := service.AssignInvestigationToGrove(ctx, "INV-001", "GROVE-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if investigationRepo.investigations["INV-001"].AssignedGroveID != "GROVE-001" {
		t.Errorf("expected grove ID 'GROVE-001', got '%s'", investigationRepo.investigations["INV-001"].AssignedGroveID)
	}
}

// ============================================================================
// GetInvestigationsByGrove Tests
// ============================================================================

func TestGetInvestigationsByGrove_Success(t *testing.T) {
	service, investigationRepo := newTestInvestigationService()
	ctx := context.Background()

	investigationRepo.investigations["INV-001"] = &secondary.InvestigationRecord{
		ID:              "INV-001",
		MissionID:       "MISSION-001",
		Title:           "Assigned Investigation",
		Status:          "active",
		AssignedGroveID: "GROVE-001",
	}
	investigationRepo.investigations["INV-002"] = &secondary.InvestigationRecord{
		ID:        "INV-002",
		MissionID: "MISSION-001",
		Title:     "Unassigned Investigation",
		Status:    "active",
	}

	investigations, err := service.GetInvestigationsByGrove(ctx, "GROVE-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(investigations) != 1 {
		t.Errorf("expected 1 investigation, got %d", len(investigations))
	}
}

// ============================================================================
// GetInvestigationQuestions Tests
// ============================================================================

func TestGetInvestigationQuestions_Success(t *testing.T) {
	service, _ := newTestInvestigationService()
	ctx := context.Background()

	questions, err := service.GetInvestigationQuestions(ctx, "INV-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// Empty list is valid
	if questions == nil {
		t.Error("expected non-nil questions slice")
	}
}
