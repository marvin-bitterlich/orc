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
	shipments           map[string]*secondary.ShipmentRecord
	groveAssignments    map[string]string // groveID -> shipmentID
	createErr           error
	getErr              error
	updateErr           error
	deleteErr           error
	listErr             error
	updateStatusErr     error
	assignGroveErr      error
	missionExistsResult bool
	missionExistsErr    error
}

func newMockShipmentRepository() *mockShipmentRepository {
	return &mockShipmentRepository{
		shipments:           make(map[string]*secondary.ShipmentRecord),
		groveAssignments:    make(map[string]string),
		missionExistsResult: true,
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
		if filters.MissionID != "" && s.MissionID != filters.MissionID {
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

func (m *mockShipmentRepository) GetByGrove(ctx context.Context, groveID string) ([]*secondary.ShipmentRecord, error) {
	var result []*secondary.ShipmentRecord
	for _, s := range m.shipments {
		if s.AssignedGroveID == groveID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *mockShipmentRepository) AssignGrove(ctx context.Context, shipmentID, groveID string) error {
	if m.assignGroveErr != nil {
		return m.assignGroveErr
	}
	if shipment, ok := m.shipments[shipmentID]; ok {
		shipment.AssignedGroveID = groveID
		m.groveAssignments[groveID] = shipmentID
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

func (m *mockShipmentRepository) MissionExists(ctx context.Context, missionID string) (bool, error) {
	if m.missionExistsErr != nil {
		return false, m.missionExistsErr
	}
	return m.missionExistsResult, nil
}

func (m *mockShipmentRepository) GroveAssignedToOther(ctx context.Context, groveID, excludeShipmentID string) (string, error) {
	if otherID, ok := m.groveAssignments[groveID]; ok && otherID != excludeShipmentID {
		return otherID, nil
	}
	return "", nil
}

// mockTaskRepositoryForShipment implements minimal TaskRepository for shipment tests.
type mockTaskRepositoryForShipment struct {
	assignErr error
}

func newMockTaskRepositoryForShipment() *mockTaskRepositoryForShipment {
	return &mockTaskRepositoryForShipment{}
}

func (m *mockTaskRepositoryForShipment) Create(ctx context.Context, task *secondary.TaskRecord) error {
	return nil
}

func (m *mockTaskRepositoryForShipment) GetByID(ctx context.Context, id string) (*secondary.TaskRecord, error) {
	return nil, errors.New("not implemented")
}

func (m *mockTaskRepositoryForShipment) List(ctx context.Context, filters secondary.TaskFilters) ([]*secondary.TaskRecord, error) {
	return nil, nil
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

func (m *mockTaskRepositoryForShipment) GetByGrove(ctx context.Context, groveID string) ([]*secondary.TaskRecord, error) {
	return nil, nil
}

func (m *mockTaskRepositoryForShipment) GetByShipment(ctx context.Context, shipmentID string) ([]*secondary.TaskRecord, error) {
	return []*secondary.TaskRecord{}, nil
}

func (m *mockTaskRepositoryForShipment) UpdateStatus(ctx context.Context, id, status string, setClaimed, setCompleted bool) error {
	return nil
}

func (m *mockTaskRepositoryForShipment) Claim(ctx context.Context, id, groveID string) error {
	return nil
}

func (m *mockTaskRepositoryForShipment) AssignGroveByShipment(ctx context.Context, shipmentID, groveID string) error {
	return m.assignErr
}

func (m *mockTaskRepositoryForShipment) MissionExists(ctx context.Context, missionID string) (bool, error) {
	return true, nil
}

func (m *mockTaskRepositoryForShipment) ShipmentExists(ctx context.Context, shipmentID string) (bool, error) {
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

// ============================================================================
// Test Helper
// ============================================================================

func newTestShipmentService() (*ShipmentServiceImpl, *mockShipmentRepository, *mockTaskRepositoryForShipment) {
	shipmentRepo := newMockShipmentRepository()
	taskRepo := newMockTaskRepositoryForShipment()
	service := NewShipmentService(shipmentRepo, taskRepo)
	return service, shipmentRepo, taskRepo
}

// ============================================================================
// CreateShipment Tests
// ============================================================================

func TestCreateShipment_Success(t *testing.T) {
	service, _, _ := newTestShipmentService()
	ctx := context.Background()

	resp, err := service.CreateShipment(ctx, primary.CreateShipmentRequest{
		MissionID:   "MISSION-001",
		Title:       "Test Shipment",
		Description: "A test shipment",
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
}

func TestCreateShipment_MissionNotFound(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.missionExistsResult = false

	_, err := service.CreateShipment(ctx, primary.CreateShipmentRequest{
		MissionID:   "MISSION-NONEXISTENT",
		Title:       "Test Shipment",
		Description: "A test shipment",
	})

	if err == nil {
		t.Fatal("expected error for non-existent mission, got nil")
	}
}

func TestCreateShipment_MissionValidationError(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.missionExistsErr = errors.New("database error")

	_, err := service.CreateShipment(ctx, primary.CreateShipmentRequest{
		MissionID:   "MISSION-001",
		Title:       "Test Shipment",
		Description: "A test shipment",
	})

	if err == nil {
		t.Fatal("expected error for mission validation failure, got nil")
	}
}

// ============================================================================
// GetShipment Tests
// ============================================================================

func TestGetShipment_Found(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:        "SHIPMENT-001",
		MissionID: "MISSION-001",
		Title:     "Test Shipment",
		Status:    "active",
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

func TestListShipments_FilterByMission(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:        "SHIPMENT-001",
		MissionID: "MISSION-001",
		Title:     "Shipment 1",
		Status:    "active",
	}
	shipmentRepo.shipments["SHIPMENT-002"] = &secondary.ShipmentRecord{
		ID:        "SHIPMENT-002",
		MissionID: "MISSION-002",
		Title:     "Shipment 2",
		Status:    "active",
	}

	shipments, err := service.ListShipments(ctx, primary.ShipmentFilters{MissionID: "MISSION-001"})

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
		ID:        "SHIPMENT-001",
		MissionID: "MISSION-001",
		Title:     "Active Shipment",
		Status:    "active",
	}
	shipmentRepo.shipments["SHIPMENT-002"] = &secondary.ShipmentRecord{
		ID:        "SHIPMENT-002",
		MissionID: "MISSION-001",
		Title:     "Paused Shipment",
		Status:    "paused",
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
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:        "SHIPMENT-001",
		MissionID: "MISSION-001",
		Title:     "Test Shipment",
		Status:    "active",
		Pinned:    false,
	}

	err := service.CompleteShipment(ctx, "SHIPMENT-001")

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
		ID:        "SHIPMENT-001",
		MissionID: "MISSION-001",
		Title:     "Pinned Shipment",
		Status:    "active",
		Pinned:    true,
	}

	err := service.CompleteShipment(ctx, "SHIPMENT-001")

	if err == nil {
		t.Fatal("expected error for completing pinned shipment, got nil")
	}
}

func TestCompleteShipment_NotFound(t *testing.T) {
	service, _, _ := newTestShipmentService()
	ctx := context.Background()

	err := service.CompleteShipment(ctx, "SHIPMENT-NONEXISTENT")

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
		ID:        "SHIPMENT-001",
		MissionID: "MISSION-001",
		Title:     "Active Shipment",
		Status:    "active",
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
		ID:        "SHIPMENT-001",
		MissionID: "MISSION-001",
		Title:     "Paused Shipment",
		Status:    "paused",
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
		ID:        "SHIPMENT-001",
		MissionID: "MISSION-001",
		Title:     "Complete Shipment",
		Status:    "complete",
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
		ID:        "SHIPMENT-001",
		MissionID: "MISSION-001",
		Title:     "Paused Shipment",
		Status:    "paused",
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
		ID:        "SHIPMENT-001",
		MissionID: "MISSION-001",
		Title:     "Active Shipment",
		Status:    "active",
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
		ID:        "SHIPMENT-001",
		MissionID: "MISSION-001",
		Title:     "Test Shipment",
		Status:    "active",
		Pinned:    false,
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
		ID:        "SHIPMENT-001",
		MissionID: "MISSION-001",
		Title:     "Pinned Shipment",
		Status:    "active",
		Pinned:    true,
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
// AssignShipmentToGrove Tests
// ============================================================================

func TestAssignShipmentToGrove_Success(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:        "SHIPMENT-001",
		MissionID: "MISSION-001",
		Title:     "Test Shipment",
		Status:    "active",
	}

	err := service.AssignShipmentToGrove(ctx, "SHIPMENT-001", "GROVE-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if shipmentRepo.shipments["SHIPMENT-001"].AssignedGroveID != "GROVE-001" {
		t.Errorf("expected grove ID 'GROVE-001', got '%s'", shipmentRepo.shipments["SHIPMENT-001"].AssignedGroveID)
	}
}

func TestAssignShipmentToGrove_ShipmentNotFound(t *testing.T) {
	service, _, _ := newTestShipmentService()
	ctx := context.Background()

	err := service.AssignShipmentToGrove(ctx, "SHIPMENT-NONEXISTENT", "GROVE-001")

	if err == nil {
		t.Fatal("expected error for non-existent shipment, got nil")
	}
}

func TestAssignShipmentToGrove_GroveAlreadyAssigned(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:        "SHIPMENT-001",
		MissionID: "MISSION-001",
		Title:     "Shipment 1",
		Status:    "active",
	}
	shipmentRepo.shipments["SHIPMENT-002"] = &secondary.ShipmentRecord{
		ID:              "SHIPMENT-002",
		MissionID:       "MISSION-001",
		Title:           "Shipment 2",
		Status:          "active",
		AssignedGroveID: "GROVE-001",
	}
	shipmentRepo.groveAssignments["GROVE-001"] = "SHIPMENT-002"

	err := service.AssignShipmentToGrove(ctx, "SHIPMENT-001", "GROVE-001")

	if err == nil {
		t.Fatal("expected error for grove already assigned, got nil")
	}
}

// ============================================================================
// GetShipmentsByGrove Tests
// ============================================================================

func TestGetShipmentsByGrove_Success(t *testing.T) {
	service, shipmentRepo, _ := newTestShipmentService()
	ctx := context.Background()

	shipmentRepo.shipments["SHIPMENT-001"] = &secondary.ShipmentRecord{
		ID:              "SHIPMENT-001",
		MissionID:       "MISSION-001",
		Title:           "Assigned Shipment",
		Status:          "active",
		AssignedGroveID: "GROVE-001",
	}
	shipmentRepo.shipments["SHIPMENT-002"] = &secondary.ShipmentRecord{
		ID:        "SHIPMENT-002",
		MissionID: "MISSION-001",
		Title:     "Unassigned Shipment",
		Status:    "active",
	}

	shipments, err := service.GetShipmentsByGrove(ctx, "GROVE-001")

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
		ID:        "SHIPMENT-001",
		MissionID: "MISSION-001",
		Title:     "Test Shipment",
		Status:    "active",
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
		ID:          "SHIPMENT-001",
		MissionID:   "MISSION-001",
		Title:       "Old Title",
		Description: "Original description",
		Status:      "active",
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
		ID:        "SHIPMENT-001",
		MissionID: "MISSION-001",
		Title:     "Test Shipment",
		Status:    "active",
	}

	err := service.DeleteShipment(ctx, "SHIPMENT-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, exists := shipmentRepo.shipments["SHIPMENT-001"]; exists {
		t.Error("expected shipment to be deleted")
	}
}
