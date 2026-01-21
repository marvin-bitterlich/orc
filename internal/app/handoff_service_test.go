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

// mockHandoffRepository implements secondary.HandoffRepository for testing.
type mockHandoffRepository struct {
	handoffs  map[string]*secondary.HandoffRecord
	latest    string // ID of latest handoff
	createErr error
	getErr    error
	listErr   error
}

func newMockHandoffRepository() *mockHandoffRepository {
	return &mockHandoffRepository{
		handoffs: make(map[string]*secondary.HandoffRecord),
	}
}

func (m *mockHandoffRepository) Create(ctx context.Context, handoff *secondary.HandoffRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.handoffs[handoff.ID] = handoff
	m.latest = handoff.ID
	return nil
}

func (m *mockHandoffRepository) GetByID(ctx context.Context, id string) (*secondary.HandoffRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if handoff, ok := m.handoffs[id]; ok {
		return handoff, nil
	}
	return nil, errors.New("handoff not found")
}

func (m *mockHandoffRepository) GetLatest(ctx context.Context) (*secondary.HandoffRecord, error) {
	if m.latest == "" {
		return nil, errors.New("no handoffs found")
	}
	return m.handoffs[m.latest], nil
}

func (m *mockHandoffRepository) GetLatestForGrove(ctx context.Context, groveID string) (*secondary.HandoffRecord, error) {
	// Find latest handoff for the grove
	for _, h := range m.handoffs {
		if h.ActiveGroveID == groveID {
			return h, nil
		}
	}
	return nil, errors.New("no handoffs found for grove")
}

func (m *mockHandoffRepository) List(ctx context.Context, limit int) ([]*secondary.HandoffRecord, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*secondary.HandoffRecord
	count := 0
	for _, h := range m.handoffs {
		if limit > 0 && count >= limit {
			break
		}
		result = append(result, h)
		count++
	}
	return result, nil
}

func (m *mockHandoffRepository) GetNextID(ctx context.Context) (string, error) {
	return "HANDOFF-001", nil
}

// ============================================================================
// Test Helper
// ============================================================================

func newTestHandoffService() (*HandoffServiceImpl, *mockHandoffRepository) {
	handoffRepo := newMockHandoffRepository()
	service := NewHandoffService(handoffRepo)
	return service, handoffRepo
}

// ============================================================================
// CreateHandoff Tests
// ============================================================================

func TestCreateHandoff_Success(t *testing.T) {
	service, _ := newTestHandoffService()
	ctx := context.Background()

	resp, err := service.CreateHandoff(ctx, primary.CreateHandoffRequest{
		HandoffNote:     "Session completed. Main task was fixing authentication bug.",
		ActiveMissionID: "MISSION-001",
		ActiveGroveID:   "GROVE-001",
		TodosSnapshot:   "- Fix auth bug [DONE]\n- Update docs [IN PROGRESS]",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.HandoffID == "" {
		t.Error("expected handoff ID to be set")
	}
	if resp.Handoff.HandoffNote != "Session completed. Main task was fixing authentication bug." {
		t.Errorf("unexpected handoff note: %s", resp.Handoff.HandoffNote)
	}
	if resp.Handoff.ActiveMissionID != "MISSION-001" {
		t.Errorf("expected active mission ID 'MISSION-001', got '%s'", resp.Handoff.ActiveMissionID)
	}
}

func TestCreateHandoff_MinimalFields(t *testing.T) {
	service, _ := newTestHandoffService()
	ctx := context.Background()

	resp, err := service.CreateHandoff(ctx, primary.CreateHandoffRequest{
		HandoffNote: "Quick session handoff",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.HandoffID == "" {
		t.Error("expected handoff ID to be set")
	}
}

// ============================================================================
// GetHandoff Tests
// ============================================================================

func TestGetHandoff_Found(t *testing.T) {
	service, handoffRepo := newTestHandoffService()
	ctx := context.Background()

	handoffRepo.handoffs["HANDOFF-001"] = &secondary.HandoffRecord{
		ID:              "HANDOFF-001",
		HandoffNote:     "Test handoff",
		ActiveMissionID: "MISSION-001",
		CreatedAt:       "2026-01-20T10:00:00Z",
	}

	handoff, err := service.GetHandoff(ctx, "HANDOFF-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if handoff.HandoffNote != "Test handoff" {
		t.Errorf("expected handoff note 'Test handoff', got '%s'", handoff.HandoffNote)
	}
}

func TestGetHandoff_NotFound(t *testing.T) {
	service, _ := newTestHandoffService()
	ctx := context.Background()

	_, err := service.GetHandoff(ctx, "HANDOFF-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent handoff, got nil")
	}
}

// ============================================================================
// GetLatestHandoff Tests
// ============================================================================

func TestGetLatestHandoff_Found(t *testing.T) {
	service, handoffRepo := newTestHandoffService()
	ctx := context.Background()

	handoffRepo.handoffs["HANDOFF-001"] = &secondary.HandoffRecord{
		ID:          "HANDOFF-001",
		HandoffNote: "First handoff",
		CreatedAt:   "2026-01-20T09:00:00Z",
	}
	handoffRepo.handoffs["HANDOFF-002"] = &secondary.HandoffRecord{
		ID:          "HANDOFF-002",
		HandoffNote: "Latest handoff",
		CreatedAt:   "2026-01-20T10:00:00Z",
	}
	handoffRepo.latest = "HANDOFF-002"

	handoff, err := service.GetLatestHandoff(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if handoff.HandoffNote != "Latest handoff" {
		t.Errorf("expected handoff note 'Latest handoff', got '%s'", handoff.HandoffNote)
	}
}

func TestGetLatestHandoff_NoHandoffs(t *testing.T) {
	service, _ := newTestHandoffService()
	ctx := context.Background()

	_, err := service.GetLatestHandoff(ctx)

	if err == nil {
		t.Fatal("expected error when no handoffs exist, got nil")
	}
}

// ============================================================================
// GetLatestHandoffForGrove Tests
// ============================================================================

func TestGetLatestHandoffForGrove_Found(t *testing.T) {
	service, handoffRepo := newTestHandoffService()
	ctx := context.Background()

	handoffRepo.handoffs["HANDOFF-001"] = &secondary.HandoffRecord{
		ID:            "HANDOFF-001",
		HandoffNote:   "Grove 1 handoff",
		ActiveGroveID: "GROVE-001",
	}
	handoffRepo.handoffs["HANDOFF-002"] = &secondary.HandoffRecord{
		ID:            "HANDOFF-002",
		HandoffNote:   "Grove 2 handoff",
		ActiveGroveID: "GROVE-002",
	}

	handoff, err := service.GetLatestHandoffForGrove(ctx, "GROVE-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if handoff.HandoffNote != "Grove 1 handoff" {
		t.Errorf("expected handoff note 'Grove 1 handoff', got '%s'", handoff.HandoffNote)
	}
}

func TestGetLatestHandoffForGrove_NotFound(t *testing.T) {
	service, handoffRepo := newTestHandoffService()
	ctx := context.Background()

	handoffRepo.handoffs["HANDOFF-001"] = &secondary.HandoffRecord{
		ID:            "HANDOFF-001",
		HandoffNote:   "Other grove handoff",
		ActiveGroveID: "GROVE-002",
	}

	_, err := service.GetLatestHandoffForGrove(ctx, "GROVE-001")

	if err == nil {
		t.Fatal("expected error for grove with no handoffs, got nil")
	}
}

// ============================================================================
// ListHandoffs Tests
// ============================================================================

func TestListHandoffs_All(t *testing.T) {
	service, handoffRepo := newTestHandoffService()
	ctx := context.Background()

	handoffRepo.handoffs["HANDOFF-001"] = &secondary.HandoffRecord{
		ID:          "HANDOFF-001",
		HandoffNote: "First handoff",
	}
	handoffRepo.handoffs["HANDOFF-002"] = &secondary.HandoffRecord{
		ID:          "HANDOFF-002",
		HandoffNote: "Second handoff",
	}
	handoffRepo.handoffs["HANDOFF-003"] = &secondary.HandoffRecord{
		ID:          "HANDOFF-003",
		HandoffNote: "Third handoff",
	}

	handoffs, err := service.ListHandoffs(ctx, 0)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(handoffs) != 3 {
		t.Errorf("expected 3 handoffs, got %d", len(handoffs))
	}
}

func TestListHandoffs_WithLimit(t *testing.T) {
	service, handoffRepo := newTestHandoffService()
	ctx := context.Background()

	handoffRepo.handoffs["HANDOFF-001"] = &secondary.HandoffRecord{
		ID:          "HANDOFF-001",
		HandoffNote: "First handoff",
	}
	handoffRepo.handoffs["HANDOFF-002"] = &secondary.HandoffRecord{
		ID:          "HANDOFF-002",
		HandoffNote: "Second handoff",
	}
	handoffRepo.handoffs["HANDOFF-003"] = &secondary.HandoffRecord{
		ID:          "HANDOFF-003",
		HandoffNote: "Third handoff",
	}

	handoffs, err := service.ListHandoffs(ctx, 2)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(handoffs) != 2 {
		t.Errorf("expected 2 handoffs with limit, got %d", len(handoffs))
	}
}

func TestListHandoffs_Empty(t *testing.T) {
	service, _ := newTestHandoffService()
	ctx := context.Background()

	handoffs, err := service.ListHandoffs(ctx, 0)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if handoffs == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(handoffs) != 0 {
		t.Errorf("expected 0 handoffs, got %d", len(handoffs))
	}
}
