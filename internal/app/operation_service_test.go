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

// mockOperationRepository implements secondary.OperationRepository for testing.
type mockOperationRepository struct {
	operations          map[string]*secondary.OperationRecord
	createErr           error
	getErr              error
	listErr             error
	updateStatusErr     error
	missionExistsResult bool
	missionExistsErr    error
}

func newMockOperationRepository() *mockOperationRepository {
	return &mockOperationRepository{
		operations:          make(map[string]*secondary.OperationRecord),
		missionExistsResult: true,
	}
}

func (m *mockOperationRepository) Create(ctx context.Context, operation *secondary.OperationRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.operations[operation.ID] = operation
	return nil
}

func (m *mockOperationRepository) GetByID(ctx context.Context, id string) (*secondary.OperationRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if operation, ok := m.operations[id]; ok {
		return operation, nil
	}
	return nil, errors.New("operation not found")
}

func (m *mockOperationRepository) List(ctx context.Context, filters secondary.OperationFilters) ([]*secondary.OperationRecord, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*secondary.OperationRecord
	for _, op := range m.operations {
		if filters.MissionID != "" && op.MissionID != filters.MissionID {
			continue
		}
		if filters.Status != "" && op.Status != filters.Status {
			continue
		}
		result = append(result, op)
	}
	return result, nil
}

func (m *mockOperationRepository) UpdateStatus(ctx context.Context, id, status string, setCompleted bool) error {
	if m.updateStatusErr != nil {
		return m.updateStatusErr
	}
	if operation, ok := m.operations[id]; ok {
		operation.Status = status
		if setCompleted {
			operation.CompletedAt = "2026-01-20T10:00:00Z"
		}
	}
	return nil
}

func (m *mockOperationRepository) GetNextID(ctx context.Context) (string, error) {
	return "OP-001", nil
}

func (m *mockOperationRepository) MissionExists(ctx context.Context, missionID string) (bool, error) {
	if m.missionExistsErr != nil {
		return false, m.missionExistsErr
	}
	return m.missionExistsResult, nil
}

// ============================================================================
// Test Helper
// ============================================================================

func newTestOperationService() (*OperationServiceImpl, *mockOperationRepository) {
	operationRepo := newMockOperationRepository()
	service := NewOperationService(operationRepo)
	return service, operationRepo
}

// ============================================================================
// CreateOperation Tests
// ============================================================================

func TestCreateOperation_Success(t *testing.T) {
	service, _ := newTestOperationService()
	ctx := context.Background()

	resp, err := service.CreateOperation(ctx, primary.CreateOperationRequest{
		MissionID:   "MISSION-001",
		Title:       "Test Operation",
		Description: "A test operation",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.OperationID == "" {
		t.Error("expected operation ID to be set")
	}
	if resp.Operation.Title != "Test Operation" {
		t.Errorf("expected title 'Test Operation', got '%s'", resp.Operation.Title)
	}
	if resp.Operation.Status != "ready" {
		t.Errorf("expected status 'ready', got '%s'", resp.Operation.Status)
	}
}

func TestCreateOperation_MissionNotFound(t *testing.T) {
	service, operationRepo := newTestOperationService()
	ctx := context.Background()

	operationRepo.missionExistsResult = false

	_, err := service.CreateOperation(ctx, primary.CreateOperationRequest{
		MissionID:   "MISSION-NONEXISTENT",
		Title:       "Test Operation",
		Description: "A test operation",
	})

	if err == nil {
		t.Fatal("expected error for non-existent mission, got nil")
	}
}

func TestCreateOperation_MissionValidationError(t *testing.T) {
	service, operationRepo := newTestOperationService()
	ctx := context.Background()

	operationRepo.missionExistsErr = errors.New("database error")

	_, err := service.CreateOperation(ctx, primary.CreateOperationRequest{
		MissionID:   "MISSION-001",
		Title:       "Test Operation",
		Description: "A test operation",
	})

	if err == nil {
		t.Fatal("expected error for mission validation failure, got nil")
	}
}

// ============================================================================
// GetOperation Tests
// ============================================================================

func TestGetOperation_Found(t *testing.T) {
	service, operationRepo := newTestOperationService()
	ctx := context.Background()

	operationRepo.operations["OP-001"] = &secondary.OperationRecord{
		ID:        "OP-001",
		MissionID: "MISSION-001",
		Title:     "Test Operation",
		Status:    "ready",
	}

	operation, err := service.GetOperation(ctx, "OP-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if operation.Title != "Test Operation" {
		t.Errorf("expected title 'Test Operation', got '%s'", operation.Title)
	}
}

func TestGetOperation_NotFound(t *testing.T) {
	service, _ := newTestOperationService()
	ctx := context.Background()

	_, err := service.GetOperation(ctx, "OP-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent operation, got nil")
	}
}

// ============================================================================
// ListOperations Tests
// ============================================================================

func TestListOperations_FilterByMission(t *testing.T) {
	service, operationRepo := newTestOperationService()
	ctx := context.Background()

	operationRepo.operations["OP-001"] = &secondary.OperationRecord{
		ID:        "OP-001",
		MissionID: "MISSION-001",
		Title:     "Operation 1",
		Status:    "ready",
	}
	operationRepo.operations["OP-002"] = &secondary.OperationRecord{
		ID:        "OP-002",
		MissionID: "MISSION-002",
		Title:     "Operation 2",
		Status:    "ready",
	}

	operations, err := service.ListOperations(ctx, primary.OperationFilters{MissionID: "MISSION-001"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(operations) != 1 {
		t.Errorf("expected 1 operation, got %d", len(operations))
	}
}

func TestListOperations_FilterByStatus(t *testing.T) {
	service, operationRepo := newTestOperationService()
	ctx := context.Background()

	operationRepo.operations["OP-001"] = &secondary.OperationRecord{
		ID:        "OP-001",
		MissionID: "MISSION-001",
		Title:     "Ready Operation",
		Status:    "ready",
	}
	operationRepo.operations["OP-002"] = &secondary.OperationRecord{
		ID:        "OP-002",
		MissionID: "MISSION-001",
		Title:     "Complete Operation",
		Status:    "complete",
	}

	operations, err := service.ListOperations(ctx, primary.OperationFilters{Status: "ready"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(operations) != 1 {
		t.Errorf("expected 1 ready operation, got %d", len(operations))
	}
}

// ============================================================================
// UpdateOperationStatus Tests
// ============================================================================

func TestUpdateOperationStatus_ToInProgress(t *testing.T) {
	service, operationRepo := newTestOperationService()
	ctx := context.Background()

	operationRepo.operations["OP-001"] = &secondary.OperationRecord{
		ID:        "OP-001",
		MissionID: "MISSION-001",
		Title:     "Test Operation",
		Status:    "ready",
	}

	err := service.UpdateOperationStatus(ctx, "OP-001", "in_progress")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if operationRepo.operations["OP-001"].Status != "in_progress" {
		t.Errorf("expected status 'in_progress', got '%s'", operationRepo.operations["OP-001"].Status)
	}
}

func TestUpdateOperationStatus_ToComplete(t *testing.T) {
	service, operationRepo := newTestOperationService()
	ctx := context.Background()

	operationRepo.operations["OP-001"] = &secondary.OperationRecord{
		ID:        "OP-001",
		MissionID: "MISSION-001",
		Title:     "Test Operation",
		Status:    "in_progress",
	}

	err := service.UpdateOperationStatus(ctx, "OP-001", "complete")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if operationRepo.operations["OP-001"].Status != "complete" {
		t.Errorf("expected status 'complete', got '%s'", operationRepo.operations["OP-001"].Status)
	}
	if operationRepo.operations["OP-001"].CompletedAt == "" {
		t.Error("expected completed_at to be set")
	}
}

// ============================================================================
// CompleteOperation Tests
// ============================================================================

func TestCompleteOperation_Success(t *testing.T) {
	service, operationRepo := newTestOperationService()
	ctx := context.Background()

	operationRepo.operations["OP-001"] = &secondary.OperationRecord{
		ID:        "OP-001",
		MissionID: "MISSION-001",
		Title:     "Test Operation",
		Status:    "in_progress",
	}

	err := service.CompleteOperation(ctx, "OP-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if operationRepo.operations["OP-001"].Status != "complete" {
		t.Errorf("expected status 'complete', got '%s'", operationRepo.operations["OP-001"].Status)
	}
}
