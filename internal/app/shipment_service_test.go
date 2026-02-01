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

// mockShipmentRepository implements secondary.ShipmentRepository for testing.
type mockShipmentRepository struct {
	shipments              map[string]*secondary.ShipmentRecord
	workbenchAssignments   map[string]string // workbenchID -> shipmentID
	createErr              error
	getErr                 error
	updateErr              error
	deleteErr              error
	listErr                error
	updateStatusErr        error
	assignWorkbenchErr     error
	commissionExistsResult bool
	commissionExistsErr    error
}

func newMockShipmentRepository() *mockShipmentRepository {
	return &mockShipmentRepository{
		shipments:              make(map[string]*secondary.ShipmentRecord),
		workbenchAssignments:   make(map[string]string),
		commissionExistsResult: true,
	}
}

func (m *mockShipmentRepository) Create(ctx context.Context, shipment *secondary.ShipmentRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.shipments[shipment.ID] = shipment
	return nil
}

func (m *mockShipmentRepository) GetByID(ctx context.Context, id string) (*secondary.ShipmentRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if shipment, ok := m.shipments[id]; ok {
		return shipment, nil
	}
	return nil, errors.New("shipment not found")
}

func (m *mockShipmentRepository) List(ctx context.Context, filters secondary.ShipmentFilters) ([]*secondary.ShipmentRecord, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*secondary.ShipmentRecord
	for _, s := range m.shipments {
		if filters.CommissionID != "" && s.CommissionID != filters.CommissionID {
			continue
		}
		if filters.Status != "" && s.Status != filters.Status {
			continue
		}
		result = append(result, s)
	}
	return result, nil
}

func (m *mockShipmentRepository) Update(ctx context.Context, shipment *secondary.ShipmentRecord) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if existing, ok := m.shipments[shipment.ID]; ok {
		if shipment.Title != "" {
			existing.Title = shipment.Title
		}
		if shipment.Description != "" {
			existing.Description = shipment.Description
		}
	}
	return nil
}

func (m *mockShipmentRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.shipments, id)
	return nil
}

func (m *mockShipmentRepository) Pin(ctx context.Context, id string) error {
	if shipment, ok := m.shipments[id]; ok {
		shipment.Pinned = true
	}
	return nil
}

func (m *mockShipmentRepository) Unpin(ctx context.Context, id string) error {
	if shipment, ok := m.shipments[id]; ok {
		shipment.Pinned = false
	}
	return nil
}

func (m *mockShipmentRepository) GetNextID(ctx context.Context) (string, error) {
	return "SHIPMENT-001", nil
}

func (m *mockShipmentRepository) GetByWorkbench(ctx context.Context, workbenchID string) ([]*secondary.ShipmentRecord, error) {
	var result []*secondary.ShipmentRecord
	for _, s := range m.shipments {
		if s.AssignedWorkbenchID == workbenchID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *mockShipmentRepository) AssignWorkbench(ctx context.Context, shipmentID, workbenchID string) error {
	if m.assignWorkbenchErr != nil {
		return m.assignWorkbenchErr
	}
	if shipment, ok := m.shipments[shipmentID]; ok {
		shipment.AssignedWorkbenchID = workbenchID
		m.workbenchAssignments[workbenchID] = shipmentID
	}
	return nil
}

func (m *mockShipmentRepository) UpdateStatus(ctx context.Context, id, status string, setCompleted bool) error {
	if m.updateStatusErr != nil {
		return m.updateStatusErr
	}
	if shipment, ok := m.shipments[id]; ok {
		shipment.Status = status
		if setCompleted {
			shipment.CompletedAt = "2026-01-20T10:00:00Z"
		}
	}
	return nil
}

func (m *mockShipmentRepository) CommissionExists(ctx context.Context, commissionID string) (bool, error) {
	if m.commissionExistsErr != nil {
		return false, m.commissionExistsErr
	}
	return m.commissionExistsResult, nil
}

func (m *mockShipmentRepository) WorkbenchAssignedToOther(ctx context.Context, workbenchID, excludeShipmentID string) (string, error) {
	if otherID, ok := m.workbenchAssignments[workbenchID]; ok && otherID != excludeShipmentID {
		return otherID, nil
	}
	return "", nil
}

func (m *mockShipmentRepository) UpdateContainer(ctx context.Context, id, containerID, containerType string) error {
	if shipment, ok := m.shipments[id]; ok {
		shipment.ContainerID = containerID
		shipment.ContainerType = containerType
	}
	return nil
}

func (m *mockShipmentRepository) ListShipyardQueue(ctx context.Context, commissionID string) ([]*secondary.ShipyardQueueEntry, error) {
	var entries []*secondary.ShipyardQueueEntry
	for _, s := range m.shipments {
		if s.ContainerType != "shipyard" {
			continue
		}
		if commissionID != "" && s.CommissionID != commissionID {
			continue
		}
		if s.Status == "complete" || s.Status == "merged" {
			continue
		}
		entries = append(entries, &secondary.ShipyardQueueEntry{
			ID:           s.ID,
			CommissionID: s.CommissionID,
			Title:        s.Title,
			Priority:     s.Priority,
			TaskCount:    0,
			DoneCount:    0,
			CreatedAt:    s.CreatedAt,
		})
	}
	return entries, nil
}

func (m *mockShipmentRepository) UpdatePriority(ctx context.Context, id string, priority *int) error {
	if shipment, ok := m.shipments[id]; ok {
		shipment.Priority = priority
	}
	return nil
}

// mockTaskRepositoryForShipment implements minimal TaskRepository for shipment tests.
type mockTaskRepositoryForShipment struct {
	tasks     map[string]*secondary.TaskRecord
	assignErr error
}

func newMockTaskRepositoryForShipment() *mockTaskRepositoryForShipment {
	return &mockTaskRepositoryForShipment{
		tasks: make(map[string]*secondary.TaskRecord),
	}
}

func (m *mockTaskRepositoryForShipment) Create(ctx context.Context, task *secondary.TaskRecord) error {
	return nil
}

func (m *mockTaskRepositoryForShipment) GetByID(ctx context.Context, id string) (*secondary.TaskRecord, error) {
	return nil, errors.New("not implemented")
}

func (m *mockTaskRepositoryForShipment) List(ctx context.Context, filters secondary.TaskFilters) ([]*secondary.TaskRecord, error) {
	var result []*secondary.TaskRecord
	for _, t := range m.tasks {
		if filters.ShipmentID != "" && t.ShipmentID != filters.ShipmentID {
			continue
		}
		result = append(result, t)
	}
	return result, nil
}

func (m *mockTaskRepositoryForShipment) Update(ctx context.Context, task *secondary.TaskRecord) error {
	return nil
}

func (m *mockTaskRepositoryForShipment) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockTaskRepositoryForShipment) Pin(ctx context.Context, id string) error {
	return nil
}

func (m *mockTaskRepositoryForShipment) Unpin(ctx context.Context, id string) error {
	return nil
}

func (m *mockTaskRepositoryForShipment) GetNextID(ctx context.Context) (string, error) {
	return "TASK-001", nil
}

func (m *mockTaskRepositoryForShipment) GetByWorkbench(ctx context.Context, workbenchID string) ([]*secondary.TaskRecord, error) {
	return nil, nil
}

func (m *mockTaskRepositoryForShipment) GetByShipment(ctx context.Context, shipmentID string) ([]*secondary.TaskRecord, error) {
	return []*secondary.TaskRecord{}, nil
}

func (m *mockTaskRepositoryForShipment) UpdateStatus(ctx context.Context, id, status string, setClaimed, setCompleted bool) error {
	return nil
}

func (m *mockTaskRepositoryForShipment) Claim(ctx context.Context, id, workbenchID string) error {
	return nil
}

func (m *mockTaskRepositoryForShipment) AssignWorkbenchByShipment(ctx context.Context, shipmentID, workbenchID string) error {
	return m.assignErr
}

func (m *mockTaskRepositoryForShipment) CommissionExists(ctx context.Context, commissionID string) (bool, error) {
	return true, nil
}

func (m *mockTaskRepositoryForShipment) ShipmentExists(ctx context.Context, shipmentID string) (bool, error) {
	return true, nil
}

func (m *mockTaskRepositoryForShipment) TomeExists(ctx context.Context, tomeID string) (bool, error) {
	return true, nil
}

func (m *mockTaskRepositoryForShipment) ConclaveExists(ctx context.Context, conclaveID string) (bool, error) {
	return true, nil
}

func (m *mockTaskRepositoryForShipment) GetTag(ctx context.Context, taskID string) (*secondary.TagRecord, error) {
	return nil, nil
}

func (m *mockTaskRepositoryForShipment) AddTag(ctx context.Context, taskID, tagID string) error {
	return nil
}

func (m *mockTaskRepositoryForShipment) RemoveTag(ctx context.Context, taskID string) error {
	return nil
}

func (m *mockTaskRepositoryForShipment) ListByTag(ctx context.Context, tagID string) ([]*secondary.TaskRecord, error) {
	return nil, nil
}

func (m *mockTaskRepositoryForShipment) GetNextEntityTagID(ctx context.Context) (string, error) {
	return "ENTITY-TAG-001", nil
}

// mockShipyardRepository implements secondary.ShipyardRepository for testing.
type mockShipyardRepository struct {
	shipyards map[string]*secondary.ShipyardRecord
}

func newMockShipyardRepository() *mockShipyardRepository {
	return &mockShipyardRepository{
		shipyards: map[string]*secondary.ShipyardRecord{
			"YARD-001": {ID: "YARD-001", CommissionID: "COMM-001"},
		},
	}
}

func (m *mockShipyardRepository) Create(ctx context.Context, shipyard *secondary.ShipyardRecord) error {
	m.shipyards[shipyard.ID] = shipyard
	return nil
}

func (m *mockShipyardRepository) GetByID(ctx context.Context, id string) (*secondary.ShipyardRecord, error) {
	if yard, ok := m.shipyards[id]; ok {
		return yard, nil
	}
	return nil, errors.New("shipyard not found")
}

func (m *mockShipyardRepository) GetByCommissionID(ctx context.Context, commissionID string) (*secondary.ShipyardRecord, error) {
	for _, yard := range m.shipyards {
		if yard.CommissionID == commissionID {
			return yard, nil
		}
	}
	return nil, errors.New("shipyard not found for commission")
}

func (m *mockShipyardRepository) GetNextID(ctx context.Context) (string, error) {
	return "YARD-002", nil
}

func (m *mockShipyardRepository) CommissionExists(ctx context.Context, commissionID string) (bool, error) {
	return true, nil
}

// ============================================================================
// Test Helper
// ============================================================================

func newTestShipmentService() (*ShipmentServiceImpl, *mockShipmentRepository, *mockTaskRepositoryForShipment) {
	shipmentRepo := newMockShipmentRepository()
	taskRepo := newMockTaskRepositoryForShipment()
	shipyardRepo := newMockShipyardRepository()
	service := NewShipmentService(shipmentRepo, taskRepo, shipyardRepo)
	return service, shipmentRepo, taskRepo
}

// ============================================================================
// CreateShipment Tests
// ============================================================================

func TestCreateShipment_Success(t *testing.T) {
	service, _, _ := newTestShipmentService()
	ctx := context.Background()

	resp, err := service.CreateShipment(ctx, primary.CreateShipmentRequest{
		CommissionID:  "COMM-001",
		Title:         "Test Shipment",
		Description:   "A test shipment",
		ContainerID:   "YARD-001",
		ContainerType: "shipyard",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.ShipmentID == "" {
		t.Error("expected shipment ID to be set")
	}
	if resp.Shipment.Title != "Test Shipment" {
		t.Errorf("expected title 'Test Shipment', got '%s'", resp.Shipment.Title)
	}
	if resp.Shipment.Status != "active" {
		t.Errorf("expected status 'active', got '%s'", resp.Shipment.Status)
	}
	if resp.Shipment.ContainerID != "YARD-001" {
		t.Errorf("expected container ID 'YARD-001', got '%s'", resp.Shipment.ContainerID)
	}
	if resp.Shipment.ContainerType != "shipyard" {
		t.Errorf("expected container type 'shipyard', got '%s'", resp.Shipment.ContainerType)
	}
}

func TestCreateShipment_MissingContainer(t *testing.T) {
	service, _, _ := newTestShipmentService()
	ctx := context.Background()

	_, err := service.CreateShipment(ctx, primary.CreateShipmentRequest{
		CommissionID: "COMM-001",
		Title:        "Test Shipment",
		Description:  "A test shipment",
	})

	if err == nil {
		t.Fatal("expected error for missing container, got nil")
	}
}

func TestCreateShipment_CommissionNotFound(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.commissionExistsResult = false

	_, err := service.CreateShipment(ctx, primary.CreateShipmentRequest{
		CommissionID:  "COMM-NONEXISTENT",
		Title:         "Test Shipment",
		Description:   "A test shipment",
		ContainerID:   "YARD-001",
		ContainerType: "shipyard",
	})

	if err == nil {
		t.Fatal("expected error for non-existent commission, got nil")
	}
}

func TestCreateShipment_CommissionValidationError(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.commissionExistsErr = errors.New("database error")

	_, err := service.CreateShipment(ctx, primary.CreateShipmentRequest{
		CommissionID:  "COMM-001",
		Title:         "Test Shipment",
		Description:   "A test shipment",
		ContainerID:   "YARD-001",
		ContainerType: "shipyard",
	})

	if err == nil {
		t.Fatal("expected error for commission validation failure, got nil")
	}
}

// ============================================================================
// GetShipment Tests
// ============================================================================

func TestGetShipment_Found(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:           "SHIPMENT-001",
		CommissionID: "COMM-001",
		Title:        "Test Shipment",
		Status:       "active",
	}

	shipment, err := service.GetShipment(ctx, "SHIPMENT-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if shipment.Title != "Test Shipment" {
		t.Errorf("expected title 'Test Shipment', got '%s'", shipment.Title)
	}
}

func TestGetShipment_NotFound(t *testing.T) {
	service, _, _ := newTestShipmentService()
	ctx := context.Background()

	_, err := service.GetShipment(ctx, "SHIPMENT-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent shipment, got nil")
	}
}

// ============================================================================
// ListShipments Tests
// ============================================================================

func TestListShipments_FilterByCommission(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:           "SHIPMENT-001",
		CommissionID: "COMM-001",
		Title:        "Shipment 1",
		Status:       "active",
	}
	shipmentRepo.shipments["SHIPMENT-002"] = &secondary.ShipmentRecord{
		ID:           "SHIPMENT-002",
		CommissionID: "COMM-002",
		Title:        "Shipment 2",
		Status:       "active",
	}

	shipments, err := service.ListShipments(ctx, primary.ShipmentFilters{CommissionID: "COMM-001"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(shipments) != 1 {
		t.Errorf("expected 1 shipment, got %d", len(shipments))
	}
}

func TestListShipments_FilterByStatus(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:           "SHIPMENT-001",
		CommissionID: "COMM-001",
		Title:        "Active Shipment",
		Status:       "active",
	}
	shipmentRepo.shipments["SHIPMENT-002"] = &secondary.ShipmentRecord{
		ID:           "SHIPMENT-002",
		CommissionID: "COMM-001",
		Title:        "Paused Shipment",
		Status:       "paused",
	}

	shipments, err := service.ListShipments(ctx, primary.ShipmentFilters{Status: "active"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(shipments) != 1 {
		t.Errorf("expected 1 active shipment, got %d", len(shipments))
	}
}

// ============================================================================
// CompleteShipment Tests
// ============================================================================

func TestCompleteShipment_UnpinnedAllowed(t *testing.T) {
	service, shipmentRepo, taskRepo := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:           "SHIPMENT-001",
		CommissionID: "COMM-001",
		Title:        "Test Shipment",
		Status:       "active",
		Pinned:       false,
	}
	// All tasks complete
	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{ID: "TASK-001", ShipmentID: "SHIPMENT-001", Status: "complete"}
	taskRepo.tasks["TASK-002"] = &secondary.TaskRecord{ID: "TASK-002", ShipmentID: "SHIPMENT-001", Status: "complete"}

	err := service.CompleteShipment(ctx, "SHIPMENT-001", false)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if shipmentRepo.shipments["SHIPMENT-001"].Status != "complete" {
		t.Errorf("expected status 'complete', got '%s'", shipmentRepo.shipments["SHIPMENT-001"].Status)
	}
}

func TestCompleteShipment_PinnedBlocked(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:           "SHIPMENT-001",
		CommissionID: "COMM-001",
		Title:        "Pinned Shipment",
		Status:       "active",
		Pinned:       true,
	}

	err := service.CompleteShipment(ctx, "SHIPMENT-001", false)

	if err == nil {
		t.Fatal("expected error for completing pinned shipment, got nil")
	}
}

func TestCompleteShipment_IncompleteTasksBlocked(t *testing.T) {
	service, shipmentRepo, taskRepo := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:           "SHIPMENT-001",
		CommissionID: "COMM-001",
		Title:        "Test Shipment",
		Status:       "active",
		Pinned:       false,
	}
	// One task incomplete
	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{ID: "TASK-001", ShipmentID: "SHIPMENT-001", Status: "complete"}
	taskRepo.tasks["TASK-002"] = &secondary.TaskRecord{ID: "TASK-002", ShipmentID: "SHIPMENT-001", Status: "ready"}

	err := service.CompleteShipment(ctx, "SHIPMENT-001", false)

	if err == nil {
		t.Fatal("expected error for incomplete tasks, got nil")
	}
	if shipmentRepo.shipments["SHIPMENT-001"].Status == "complete" {
		t.Error("shipment should not be completed with incomplete tasks")
	}
}

func TestCompleteShipment_IncompleteTasksForced(t *testing.T) {
	service, shipmentRepo, taskRepo := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:           "SHIPMENT-001",
		CommissionID: "COMM-001",
		Title:        "Test Shipment",
		Status:       "active",
		Pinned:       false,
	}
	// One task incomplete
	taskRepo.tasks["TASK-001"] = &secondary.TaskRecord{ID: "TASK-001", ShipmentID: "SHIPMENT-001", Status: "complete"}
	taskRepo.tasks["TASK-002"] = &secondary.TaskRecord{ID: "TASK-002", ShipmentID: "SHIPMENT-001", Status: "ready"}

	// Force completion
	err := service.CompleteShipment(ctx, "SHIPMENT-001", true)

	if err != nil {
		t.Fatalf("expected no error with force=true, got %v", err)
	}
	if shipmentRepo.shipments["SHIPMENT-001"].Status != "complete" {
		t.Errorf("expected status 'complete', got '%s'", shipmentRepo.shipments["SHIPMENT-001"].Status)
	}
}

func TestCompleteShipment_NotFound(t *testing.T) {
	service, _, _ := newTestShipmentService()
	ctx := context.Background()

	err := service.CompleteShipment(ctx, "SHIPMENT-NONEXISTENT", false)

	if err == nil {
		t.Fatal("expected error for non-existent shipment, got nil")
	}
}

// ============================================================================
// PauseShipment Tests
// ============================================================================

func TestPauseShipment_ActiveAllowed(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:           "SHIPMENT-001",
		CommissionID: "COMM-001",
		Title:        "Active Shipment",
		Status:       "active",
	}

	err := service.PauseShipment(ctx, "SHIPMENT-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if shipmentRepo.shipments["SHIPMENT-001"].Status != "paused" {
		t.Errorf("expected status 'paused', got '%s'", shipmentRepo.shipments["SHIPMENT-001"].Status)
	}
}

func TestPauseShipment_NotActiveBlocked(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:           "SHIPMENT-001",
		CommissionID: "COMM-001",
		Title:        "Paused Shipment",
		Status:       "paused",
	}

	err := service.PauseShipment(ctx, "SHIPMENT-001")

	if err == nil {
		t.Fatal("expected error for pausing non-active shipment, got nil")
	}
}

func TestPauseShipment_CompleteBlocked(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:           "SHIPMENT-001",
		CommissionID: "COMM-001",
		Title:        "Complete Shipment",
		Status:       "complete",
	}

	err := service.PauseShipment(ctx, "SHIPMENT-001")

	if err == nil {
		t.Fatal("expected error for pausing complete shipment, got nil")
	}
}

// ============================================================================
// ResumeShipment Tests
// ============================================================================

func TestResumeShipment_PausedAllowed(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:           "SHIPMENT-001",
		CommissionID: "COMM-001",
		Title:        "Paused Shipment",
		Status:       "paused",
	}

	err := service.ResumeShipment(ctx, "SHIPMENT-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if shipmentRepo.shipments["SHIPMENT-001"].Status != "active" {
		t.Errorf("expected status 'active', got '%s'", shipmentRepo.shipments["SHIPMENT-001"].Status)
	}
}

func TestResumeShipment_NotPausedBlocked(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:           "SHIPMENT-001",
		CommissionID: "COMM-001",
		Title:        "Active Shipment",
		Status:       "active",
	}

	err := service.ResumeShipment(ctx, "SHIPMENT-001")

	if err == nil {
		t.Fatal("expected error for resuming non-paused shipment, got nil")
	}
}

// ============================================================================
// Pin/Unpin Tests
// ============================================================================

func TestPinShipment(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:           "SHIPMENT-001",
		CommissionID: "COMM-001",
		Title:        "Test Shipment",
		Status:       "active",
		Pinned:       false,
	}

	err := service.PinShipment(ctx, "SHIPMENT-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !shipmentRepo.shipments["SHIPMENT-001"].Pinned {
		t.Error("expected shipment to be pinned")
	}
}

func TestUnpinShipment(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:           "SHIPMENT-001",
		CommissionID: "COMM-001",
		Title:        "Pinned Shipment",
		Status:       "active",
		Pinned:       true,
	}

	err := service.UnpinShipment(ctx, "SHIPMENT-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if shipmentRepo.shipments["SHIPMENT-001"].Pinned {
		t.Error("expected shipment to be unpinned")
	}
}

// ============================================================================
// AssignShipmentToWorkbench Tests
// ============================================================================

func TestAssignShipmentToWorkbench_Success(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:           "SHIPMENT-001",
		CommissionID: "COMM-001",
		Title:        "Test Shipment",
		Status:       "active",
	}

	err := service.AssignShipmentToWorkbench(ctx, "SHIPMENT-001", "BENCH-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if shipmentRepo.shipments["SHIPMENT-001"].AssignedWorkbenchID != "BENCH-001" {
		t.Errorf("expected workbench ID 'BENCH-001', got '%s'", shipmentRepo.shipments["SHIPMENT-001"].AssignedWorkbenchID)
	}
}

func TestAssignShipmentToWorkbench_ShipmentNotFound(t *testing.T) {
	service, _, _ := newTestShipmentService()
	ctx := context.Background()

	err := service.AssignShipmentToWorkbench(ctx, "SHIPMENT-NONEXISTENT", "BENCH-001")

	if err == nil {
		t.Fatal("expected error for non-existent shipment, got nil")
	}
}

func TestAssignShipmentToWorkbench_WorkbenchAlreadyAssigned(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:           "SHIPMENT-001",
		CommissionID: "COMM-001",
		Title:        "Shipment 1",
		Status:       "active",
	}
	shipmentRepo.shipments["SHIPMENT-002"] = &secondary.ShipmentRecord{
		ID:                  "SHIPMENT-002",
		CommissionID:        "COMM-001",
		Title:               "Shipment 2",
		Status:              "active",
		AssignedWorkbenchID: "BENCH-001",
	}
	shipmentRepo.workbenchAssignments["BENCH-001"] = "SHIPMENT-002"

	err := service.AssignShipmentToWorkbench(ctx, "SHIPMENT-001", "BENCH-001")

	if err == nil {
		t.Fatal("expected error for workbench already assigned, got nil")
	}
}

// ============================================================================
// GetShipmentsByWorkbench Tests
// ============================================================================

func TestGetShipmentsByWorkbench_Success(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:                  "SHIPMENT-001",
		CommissionID:        "COMM-001",
		Title:               "Assigned Shipment",
		Status:              "active",
		AssignedWorkbenchID: "BENCH-001",
	}
	shipmentRepo.shipments["SHIPMENT-002"] = &secondary.ShipmentRecord{
		ID:           "SHIPMENT-002",
		CommissionID: "COMM-001",
		Title:        "Unassigned Shipment",
		Status:       "active",
	}

	shipments, err := service.GetShipmentsByWorkbench(ctx, "BENCH-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(shipments) != 1 {
		t.Errorf("expected 1 shipment, got %d", len(shipments))
	}
}

// ============================================================================
// GetShipmentTasks Tests
// ============================================================================

func TestGetShipmentTasks_Success(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:           "SHIPMENT-001",
		CommissionID: "COMM-001",
		Title:        "Test Shipment",
		Status:       "active",
	}

	tasks, err := service.GetShipmentTasks(ctx, "SHIPMENT-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// Empty list is valid
	if tasks == nil {
		t.Error("expected non-nil tasks slice")
	}
}

// ============================================================================
// UpdateShipment Tests
// ============================================================================

func TestUpdateShipment_Title(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:           "SHIPMENT-001",
		CommissionID: "COMM-001",
		Title:        "Old Title",
		Description:  "Original description",
		Status:       "active",
	}

	err := service.UpdateShipment(ctx, primary.UpdateShipmentRequest{
		ShipmentID: "SHIPMENT-001",
		Title:      "New Title",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if shipmentRepo.shipments["SHIPMENT-001"].Title != "New Title" {
		t.Errorf("expected title 'New Title', got '%s'", shipmentRepo.shipments["SHIPMENT-001"].Title)
	}
}

// ============================================================================
// DeleteShipment Tests
// ============================================================================

func TestDeleteShipment_Success(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:           "SHIPMENT-001",
		CommissionID: "COMM-001",
		Title:        "Test Shipment",
		Status:       "active",
	}

	err := service.DeleteShipment(ctx, "SHIPMENT-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, exists := shipmentRepo.shipments["SHIPMENT-001"]; exists {
		t.Error("expected shipment to be deleted")
	}
}
