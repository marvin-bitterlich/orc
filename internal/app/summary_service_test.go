package app

import (
	"context"
	"testing"

	"github.com/example/orc/internal/ports/primary"
)

// ============================================================================
// Mock Implementations for Summary Service
// ============================================================================

// mockCommissionServiceForSummary implements primary.CommissionService for testing.
type mockCommissionServiceForSummary struct {
	commissions map[string]*primary.Commission
	getErr      error
}

func newMockCommissionServiceForSummary() *mockCommissionServiceForSummary {
	return &mockCommissionServiceForSummary{
		commissions: make(map[string]*primary.Commission),
	}
}

func (m *mockCommissionServiceForSummary) CreateCommission(_ context.Context, _ primary.CreateCommissionRequest) (*primary.CreateCommissionResponse, error) {
	return nil, nil
}

func (m *mockCommissionServiceForSummary) StartCommission(_ context.Context, _ primary.StartCommissionRequest) (*primary.StartCommissionResponse, error) {
	return nil, nil
}

func (m *mockCommissionServiceForSummary) LaunchCommission(_ context.Context, _ primary.LaunchCommissionRequest) (*primary.LaunchCommissionResponse, error) {
	return nil, nil
}

func (m *mockCommissionServiceForSummary) GetCommission(_ context.Context, id string) (*primary.Commission, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if comm, ok := m.commissions[id]; ok {
		return comm, nil
	}
	return nil, nil
}

func (m *mockCommissionServiceForSummary) ListCommissions(_ context.Context, _ primary.CommissionFilters) ([]*primary.Commission, error) {
	var result []*primary.Commission
	for _, c := range m.commissions {
		result = append(result, c)
	}
	return result, nil
}

func (m *mockCommissionServiceForSummary) ArchiveCommission(_ context.Context, _ string) error {
	return nil
}

func (m *mockCommissionServiceForSummary) CompleteCommission(_ context.Context, _ string) error {
	return nil
}

func (m *mockCommissionServiceForSummary) UpdateCommission(_ context.Context, _ primary.UpdateCommissionRequest) error {
	return nil
}

func (m *mockCommissionServiceForSummary) DeleteCommission(_ context.Context, _ primary.DeleteCommissionRequest) error {
	return nil
}

func (m *mockCommissionServiceForSummary) PinCommission(_ context.Context, _ string) error {
	return nil
}

func (m *mockCommissionServiceForSummary) UnpinCommission(_ context.Context, _ string) error {
	return nil
}

// mockConclaveServiceForSummary implements primary.ConclaveService for testing.
type mockConclaveServiceForSummary struct {
	conclaves map[string]*primary.Conclave
}

func newMockConclaveServiceForSummary() *mockConclaveServiceForSummary {
	return &mockConclaveServiceForSummary{
		conclaves: make(map[string]*primary.Conclave),
	}
}

func (m *mockConclaveServiceForSummary) CreateConclave(_ context.Context, _ primary.CreateConclaveRequest) (*primary.CreateConclaveResponse, error) {
	return nil, nil
}

func (m *mockConclaveServiceForSummary) GetConclave(_ context.Context, id string) (*primary.Conclave, error) {
	if con, ok := m.conclaves[id]; ok {
		return con, nil
	}
	return nil, nil
}

func (m *mockConclaveServiceForSummary) ListConclaves(_ context.Context, filters primary.ConclaveFilters) ([]*primary.Conclave, error) {
	var result []*primary.Conclave
	for _, c := range m.conclaves {
		if filters.CommissionID != "" && c.CommissionID != filters.CommissionID {
			continue
		}
		result = append(result, c)
	}
	return result, nil
}

func (m *mockConclaveServiceForSummary) CompleteConclave(_ context.Context, _ string) error {
	return nil
}

func (m *mockConclaveServiceForSummary) PauseConclave(_ context.Context, _ string) error {
	return nil
}

func (m *mockConclaveServiceForSummary) ResumeConclave(_ context.Context, _ string) error {
	return nil
}

func (m *mockConclaveServiceForSummary) UpdateConclave(_ context.Context, _ primary.UpdateConclaveRequest) error {
	return nil
}

func (m *mockConclaveServiceForSummary) PinConclave(_ context.Context, _ string) error {
	return nil
}

func (m *mockConclaveServiceForSummary) UnpinConclave(_ context.Context, _ string) error {
	return nil
}

func (m *mockConclaveServiceForSummary) DeleteConclave(_ context.Context, _ string) error {
	return nil
}

func (m *mockConclaveServiceForSummary) GetConclavesByShipment(_ context.Context, _ string) ([]*primary.Conclave, error) {
	return nil, nil
}

func (m *mockConclaveServiceForSummary) GetConclaveTasks(_ context.Context, _ string) ([]*primary.ConclaveTask, error) {
	return nil, nil
}

func (m *mockConclaveServiceForSummary) GetConclavePlans(_ context.Context, _ string) ([]*primary.ConclavePlan, error) {
	return nil, nil
}

// mockTomeServiceForSummary implements primary.TomeService for testing.
type mockTomeServiceForSummary struct {
	tomes     map[string]*primary.Tome
	tomeNotes map[string][]*primary.Note
}

func newMockTomeServiceForSummary() *mockTomeServiceForSummary {
	return &mockTomeServiceForSummary{
		tomes:     make(map[string]*primary.Tome),
		tomeNotes: make(map[string][]*primary.Note),
	}
}

func (m *mockTomeServiceForSummary) CreateTome(_ context.Context, _ primary.CreateTomeRequest) (*primary.CreateTomeResponse, error) {
	return nil, nil
}

func (m *mockTomeServiceForSummary) GetTome(_ context.Context, id string) (*primary.Tome, error) {
	if t, ok := m.tomes[id]; ok {
		return t, nil
	}
	return nil, nil
}

func (m *mockTomeServiceForSummary) ListTomes(_ context.Context, filters primary.TomeFilters) ([]*primary.Tome, error) {
	var result []*primary.Tome
	for _, t := range m.tomes {
		if filters.CommissionID != "" && t.CommissionID != filters.CommissionID {
			continue
		}
		result = append(result, t)
	}
	return result, nil
}

func (m *mockTomeServiceForSummary) CloseTome(_ context.Context, _ string) error {
	return nil
}

func (m *mockTomeServiceForSummary) UpdateTome(_ context.Context, _ primary.UpdateTomeRequest) error {
	return nil
}

func (m *mockTomeServiceForSummary) PinTome(_ context.Context, _ string) error {
	return nil
}

func (m *mockTomeServiceForSummary) UnpinTome(_ context.Context, _ string) error {
	return nil
}

func (m *mockTomeServiceForSummary) DeleteTome(_ context.Context, _ string) error {
	return nil
}

func (m *mockTomeServiceForSummary) AssignTomeToWorkbench(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockTomeServiceForSummary) GetTomesByWorkbench(_ context.Context, _ string) ([]*primary.Tome, error) {
	return nil, nil
}

func (m *mockTomeServiceForSummary) GetTomeNotes(_ context.Context, tomeID string) ([]*primary.Note, error) {
	if notes, ok := m.tomeNotes[tomeID]; ok {
		return notes, nil
	}
	return []*primary.Note{}, nil
}

func (m *mockTomeServiceForSummary) ParkTome(_ context.Context, _ string) error {
	return nil
}

func (m *mockTomeServiceForSummary) UnparkTome(_ context.Context, _, _ string) error {
	return nil
}

// mockShipmentServiceForSummary implements primary.ShipmentService for testing.
type mockShipmentServiceForSummary struct {
	shipments     map[string]*primary.Shipment
	shipmentTasks map[string][]*primary.Task
}

func newMockShipmentServiceForSummary() *mockShipmentServiceForSummary {
	return &mockShipmentServiceForSummary{
		shipments:     make(map[string]*primary.Shipment),
		shipmentTasks: make(map[string][]*primary.Task),
	}
}

func (m *mockShipmentServiceForSummary) CreateShipment(_ context.Context, _ primary.CreateShipmentRequest) (*primary.CreateShipmentResponse, error) {
	return nil, nil
}

func (m *mockShipmentServiceForSummary) GetShipment(_ context.Context, id string) (*primary.Shipment, error) {
	if s, ok := m.shipments[id]; ok {
		return s, nil
	}
	return nil, nil
}

func (m *mockShipmentServiceForSummary) ListShipments(_ context.Context, filters primary.ShipmentFilters) ([]*primary.Shipment, error) {
	var result []*primary.Shipment
	for _, s := range m.shipments {
		if filters.CommissionID != "" && s.CommissionID != filters.CommissionID {
			continue
		}
		result = append(result, s)
	}
	return result, nil
}

func (m *mockShipmentServiceForSummary) CompleteShipment(_ context.Context, _ string, _ bool) error {
	return nil
}

func (m *mockShipmentServiceForSummary) PauseShipment(_ context.Context, _ string) error {
	return nil
}

func (m *mockShipmentServiceForSummary) ResumeShipment(_ context.Context, _ string) error {
	return nil
}

func (m *mockShipmentServiceForSummary) UpdateShipment(_ context.Context, _ primary.UpdateShipmentRequest) error {
	return nil
}

func (m *mockShipmentServiceForSummary) PinShipment(_ context.Context, _ string) error {
	return nil
}

func (m *mockShipmentServiceForSummary) UnpinShipment(_ context.Context, _ string) error {
	return nil
}

func (m *mockShipmentServiceForSummary) AssignShipmentToWorkbench(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockShipmentServiceForSummary) GetShipmentsByWorkbench(_ context.Context, _ string) ([]*primary.Shipment, error) {
	return nil, nil
}

func (m *mockShipmentServiceForSummary) GetShipmentTasks(_ context.Context, shipmentID string) ([]*primary.Task, error) {
	if tasks, ok := m.shipmentTasks[shipmentID]; ok {
		return tasks, nil
	}
	return []*primary.Task{}, nil
}

func (m *mockShipmentServiceForSummary) DeleteShipment(_ context.Context, _ string) error {
	return nil
}

func (m *mockShipmentServiceForSummary) ParkShipment(_ context.Context, _ string) error {
	return nil
}

func (m *mockShipmentServiceForSummary) UnparkShipment(_ context.Context, _, _ string) error {
	return nil
}

// mockTaskServiceForSummary implements primary.TaskService for testing.
type mockTaskServiceForSummary struct{}

func newMockTaskServiceForSummary() *mockTaskServiceForSummary {
	return &mockTaskServiceForSummary{}
}

func (m *mockTaskServiceForSummary) CreateTask(_ context.Context, _ primary.CreateTaskRequest) (*primary.CreateTaskResponse, error) {
	return nil, nil
}

func (m *mockTaskServiceForSummary) GetTask(_ context.Context, _ string) (*primary.Task, error) {
	return nil, nil
}

func (m *mockTaskServiceForSummary) ListTasks(_ context.Context, _ primary.TaskFilters) ([]*primary.Task, error) {
	return nil, nil
}

func (m *mockTaskServiceForSummary) ClaimTask(_ context.Context, _ primary.ClaimTaskRequest) error {
	return nil
}

func (m *mockTaskServiceForSummary) CompleteTask(_ context.Context, _ string) error {
	return nil
}

func (m *mockTaskServiceForSummary) PauseTask(_ context.Context, _ string) error {
	return nil
}

func (m *mockTaskServiceForSummary) ResumeTask(_ context.Context, _ string) error {
	return nil
}

func (m *mockTaskServiceForSummary) UpdateTask(_ context.Context, _ primary.UpdateTaskRequest) error {
	return nil
}

func (m *mockTaskServiceForSummary) PinTask(_ context.Context, _ string) error {
	return nil
}

func (m *mockTaskServiceForSummary) UnpinTask(_ context.Context, _ string) error {
	return nil
}

func (m *mockTaskServiceForSummary) DeleteTask(_ context.Context, _ string) error {
	return nil
}

func (m *mockTaskServiceForSummary) GetTasksByWorkbench(_ context.Context, _ string) ([]*primary.Task, error) {
	return nil, nil
}

func (m *mockTaskServiceForSummary) TagTask(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockTaskServiceForSummary) UntagTask(_ context.Context, _ string) error {
	return nil
}

func (m *mockTaskServiceForSummary) ListTasksByTag(_ context.Context, _ string) ([]*primary.Task, error) {
	return nil, nil
}

func (m *mockTaskServiceForSummary) DiscoverTasks(_ context.Context, _ string) ([]*primary.Task, error) {
	return nil, nil
}

func (m *mockTaskServiceForSummary) MoveTask(_ context.Context, _ primary.MoveTaskRequest) error {
	return nil
}

// mockNoteServiceForSummary implements primary.NoteService for testing.
type mockNoteServiceForSummary struct{}

func newMockNoteServiceForSummary() *mockNoteServiceForSummary {
	return &mockNoteServiceForSummary{}
}

func (m *mockNoteServiceForSummary) CreateNote(_ context.Context, _ primary.CreateNoteRequest) (*primary.CreateNoteResponse, error) {
	return nil, nil
}

func (m *mockNoteServiceForSummary) GetNote(_ context.Context, _ string) (*primary.Note, error) {
	return nil, nil
}

func (m *mockNoteServiceForSummary) ListNotes(_ context.Context, _ primary.NoteFilters) ([]*primary.Note, error) {
	return nil, nil
}

func (m *mockNoteServiceForSummary) CloseNote(_ context.Context, _ primary.CloseNoteRequest) error {
	return nil
}

func (m *mockNoteServiceForSummary) ReopenNote(_ context.Context, _ string) error {
	return nil
}

func (m *mockNoteServiceForSummary) UpdateNote(_ context.Context, _ primary.UpdateNoteRequest) error {
	return nil
}

func (m *mockNoteServiceForSummary) PinNote(_ context.Context, _ string) error {
	return nil
}

func (m *mockNoteServiceForSummary) UnpinNote(_ context.Context, _ string) error {
	return nil
}

func (m *mockNoteServiceForSummary) DeleteNote(_ context.Context, _ string) error {
	return nil
}

func (m *mockNoteServiceForSummary) GetNotesByContainer(_ context.Context, _, _ string) ([]*primary.Note, error) {
	return nil, nil
}

func (m *mockNoteServiceForSummary) MoveNote(_ context.Context, _ primary.MoveNoteRequest) error {
	return nil
}

func (m *mockNoteServiceForSummary) MergeNotes(_ context.Context, _ primary.MergeNoteRequest) error {
	return nil
}

// mockWorkbenchServiceForSummary implements primary.WorkbenchService for testing.
type mockWorkbenchServiceForSummary struct {
	workbenches map[string]*primary.Workbench
}

func newMockWorkbenchServiceForSummary() *mockWorkbenchServiceForSummary {
	return &mockWorkbenchServiceForSummary{
		workbenches: make(map[string]*primary.Workbench),
	}
}

func (m *mockWorkbenchServiceForSummary) CreateWorkbench(_ context.Context, _ primary.CreateWorkbenchRequest) (*primary.CreateWorkbenchResponse, error) {
	return nil, nil
}

func (m *mockWorkbenchServiceForSummary) GetWorkbench(_ context.Context, id string) (*primary.Workbench, error) {
	if wb, ok := m.workbenches[id]; ok {
		return wb, nil
	}
	return nil, nil
}

func (m *mockWorkbenchServiceForSummary) GetWorkbenchByPath(_ context.Context, _ string) (*primary.Workbench, error) {
	return nil, nil
}

func (m *mockWorkbenchServiceForSummary) ListWorkbenches(_ context.Context, _ primary.WorkbenchFilters) ([]*primary.Workbench, error) {
	return nil, nil
}

func (m *mockWorkbenchServiceForSummary) RenameWorkbench(_ context.Context, _ primary.RenameWorkbenchRequest) error {
	return nil
}

func (m *mockWorkbenchServiceForSummary) UpdateWorkbenchPath(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockWorkbenchServiceForSummary) DeleteWorkbench(_ context.Context, _ primary.DeleteWorkbenchRequest) error {
	return nil
}

func (m *mockWorkbenchServiceForSummary) CheckoutBranch(_ context.Context, _ primary.CheckoutBranchRequest) (*primary.CheckoutBranchResponse, error) {
	return nil, nil
}

func (m *mockWorkbenchServiceForSummary) GetWorkbenchStatus(_ context.Context, _ string) (*primary.WorkbenchGitStatus, error) {
	return nil, nil
}

func (m *mockWorkbenchServiceForSummary) UpdateFocusedID(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockWorkbenchServiceForSummary) GetFocusedID(_ context.Context, _ string) (string, error) {
	return "", nil
}

// ============================================================================
// Tests
// ============================================================================

func TestSummaryService_GetCommissionSummary_GoblinFullTree(t *testing.T) {
	// Setup mocks
	commissionSvc := newMockCommissionServiceForSummary()
	conclaveSvc := newMockConclaveServiceForSummary()
	tomeSvc := newMockTomeServiceForSummary()
	shipmentSvc := newMockShipmentServiceForSummary()
	taskSvc := newMockTaskServiceForSummary()
	noteSvc := newMockNoteServiceForSummary()
	workbenchSvc := newMockWorkbenchServiceForSummary()

	// Create commission
	commissionSvc.commissions["COMM-001"] = &primary.Commission{
		ID:     "COMM-001",
		Title:  "Test Commission",
		Status: "active",
	}

	// Create conclave
	conclaveSvc.conclaves["CON-001"] = &primary.Conclave{
		ID:           "CON-001",
		CommissionID: "COMM-001",
		Title:        "Design Conclave",
		Status:       "active",
	}

	// Create tome under conclave
	tomeSvc.tomes["TOME-001"] = &primary.Tome{
		ID:            "TOME-001",
		CommissionID:  "COMM-001",
		ContainerID:   "CON-001",
		ContainerType: "conclave",
		Title:         "Design Notes",
		Status:        "open",
	}

	// Add notes to tome
	tomeSvc.tomeNotes["TOME-001"] = []*primary.Note{
		{ID: "NOTE-001", Title: "Note 1", Status: "open"},
		{ID: "NOTE-002", Title: "Note 2", Status: "open"},
		{ID: "NOTE-003", Title: "Note 3", Status: "closed"},
	}

	// Create shipment under conclave with assigned workbench
	shipmentSvc.shipments["SHIP-001"] = &primary.Shipment{
		ID:                  "SHIP-001",
		CommissionID:        "COMM-001",
		ContainerID:         "CON-001",
		ContainerType:       "conclave",
		Title:               "Bug Fixes",
		Status:              "active",
		AssignedWorkbenchID: "BENCH-001",
	}

	// Add tasks to shipment
	shipmentSvc.shipmentTasks["SHIP-001"] = []*primary.Task{
		{ID: "TASK-001", Title: "Task 1", Status: "complete"},
		{ID: "TASK-002", Title: "Task 2", Status: "ready"},
		{ID: "TASK-003", Title: "Task 3", Status: "ready"},
	}

	// Add workbench
	workbenchSvc.workbenches["BENCH-001"] = &primary.Workbench{
		ID:   "BENCH-001",
		Name: "bench-alpha",
	}

	// Create library tome
	tomeSvc.tomes["TOME-002"] = &primary.Tome{
		ID:            "TOME-002",
		CommissionID:  "COMM-001",
		ContainerID:   "LIB-001",
		ContainerType: "library",
		Title:         "Archived Notes",
		Status:        "open",
	}

	// Create service
	svc := NewSummaryService(commissionSvc, conclaveSvc, tomeSvc, shipmentSvc, taskSvc, noteSvc, workbenchSvc, nil, nil, nil, nil)

	// Request summary
	req := primary.SummaryRequest{
		CommissionID: "COMM-001",
		WorkshopID:   "WORK-001",
		FocusID:      "CON-001",
	}

	summary, err := svc.GetCommissionSummary(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify commission
	if summary.ID != "COMM-001" {
		t.Errorf("expected commission ID COMM-001, got %s", summary.ID)
	}

	// Verify conclaves
	if len(summary.Conclaves) != 1 {
		t.Errorf("expected 1 conclave, got %d", len(summary.Conclaves))
	}
	if summary.Conclaves[0].ID != "CON-001" {
		t.Errorf("expected conclave ID CON-001, got %s", summary.Conclaves[0].ID)
	}
	if !summary.Conclaves[0].IsFocused {
		t.Error("expected conclave to be focused")
	}

	// Verify tomes under conclave
	if len(summary.Conclaves[0].Tomes) != 1 {
		t.Errorf("expected 1 tome under conclave, got %d", len(summary.Conclaves[0].Tomes))
	}
	if summary.Conclaves[0].Tomes[0].NoteCount != 2 {
		t.Errorf("expected 2 open notes, got %d", summary.Conclaves[0].Tomes[0].NoteCount)
	}

	// Verify shipments under conclave
	if len(summary.Conclaves[0].Shipments) != 1 {
		t.Errorf("expected 1 shipment under conclave, got %d", len(summary.Conclaves[0].Shipments))
	}
	if summary.Conclaves[0].Shipments[0].BenchID != "BENCH-001" {
		t.Errorf("expected bench ID BENCH-001, got %s", summary.Conclaves[0].Shipments[0].BenchID)
	}
	if summary.Conclaves[0].Shipments[0].TasksDone != 1 {
		t.Errorf("expected 1 task done, got %d", summary.Conclaves[0].Shipments[0].TasksDone)
	}
	if summary.Conclaves[0].Shipments[0].TasksTotal != 3 {
		t.Errorf("expected 3 total tasks, got %d", summary.Conclaves[0].Shipments[0].TasksTotal)
	}

	// Verify library
	if summary.Library.TomeCount != 1 {
		t.Errorf("expected 1 library tome, got %d", summary.Library.TomeCount)
	}
}

func TestSummaryService_GetCommissionSummary_ShowsAllWorkbenchShipments(t *testing.T) {
	// Setup mocks
	commissionSvc := newMockCommissionServiceForSummary()
	conclaveSvc := newMockConclaveServiceForSummary()
	tomeSvc := newMockTomeServiceForSummary()
	shipmentSvc := newMockShipmentServiceForSummary()
	taskSvc := newMockTaskServiceForSummary()
	noteSvc := newMockNoteServiceForSummary()
	workbenchSvc := newMockWorkbenchServiceForSummary()

	// Create commission
	commissionSvc.commissions["COMM-001"] = &primary.Commission{
		ID:     "COMM-001",
		Title:  "Test Commission",
		Status: "active",
	}

	// Create conclave
	conclaveSvc.conclaves["CON-001"] = &primary.Conclave{
		ID:           "CON-001",
		CommissionID: "COMM-001",
		Title:        "Design Conclave",
		Status:       "active",
	}

	// Create shipment assigned to one workbench
	shipmentSvc.shipments["SHIP-001"] = &primary.Shipment{
		ID:                  "SHIP-001",
		CommissionID:        "COMM-001",
		ContainerID:         "CON-001",
		ContainerType:       "conclave",
		Title:               "My Shipment",
		Status:              "active",
		AssignedWorkbenchID: "BENCH-001",
	}

	// Create shipment assigned to different workbench
	shipmentSvc.shipments["SHIP-002"] = &primary.Shipment{
		ID:                  "SHIP-002",
		CommissionID:        "COMM-001",
		ContainerID:         "CON-001",
		ContainerType:       "conclave",
		Title:               "Other Shipment",
		Status:              "active",
		AssignedWorkbenchID: "BENCH-002",
	}

	// Create unassigned shipment
	shipmentSvc.shipments["SHIP-003"] = &primary.Shipment{
		ID:                  "SHIP-003",
		CommissionID:        "COMM-001",
		ContainerID:         "CON-001",
		ContainerType:       "conclave",
		Title:               "Unassigned Shipment",
		Status:              "active",
		AssignedWorkbenchID: "",
	}

	// Create service
	svc := NewSummaryService(commissionSvc, conclaveSvc, tomeSvc, shipmentSvc, taskSvc, noteSvc, workbenchSvc, nil, nil, nil, nil)

	// Request summary - all shipments should be visible regardless of workbench assignment
	req := primary.SummaryRequest{
		CommissionID: "COMM-001",
		WorkbenchID:  "BENCH-001",
		FocusID:      "SHIP-001",
	}

	summary, err := svc.GetCommissionSummary(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All 3 shipments should be visible (no more IMP filtering)
	if len(summary.Conclaves[0].Shipments) != 3 {
		t.Errorf("expected 3 visible shipments, got %d", len(summary.Conclaves[0].Shipments))
	}
}

func TestSummaryService_GetCommissionSummary_TaskCounting(t *testing.T) {
	// Setup mocks
	commissionSvc := newMockCommissionServiceForSummary()
	conclaveSvc := newMockConclaveServiceForSummary()
	tomeSvc := newMockTomeServiceForSummary()
	shipmentSvc := newMockShipmentServiceForSummary()
	taskSvc := newMockTaskServiceForSummary()
	noteSvc := newMockNoteServiceForSummary()
	workbenchSvc := newMockWorkbenchServiceForSummary()

	commissionSvc.commissions["COMM-001"] = &primary.Commission{
		ID:     "COMM-001",
		Title:  "Test Commission",
		Status: "active",
	}

	conclaveSvc.conclaves["CON-001"] = &primary.Conclave{
		ID:           "CON-001",
		CommissionID: "COMM-001",
		Title:        "Design",
		Status:       "active",
	}

	shipmentSvc.shipments["SHIP-001"] = &primary.Shipment{
		ID:            "SHIP-001",
		CommissionID:  "COMM-001",
		ContainerID:   "CON-001",
		ContainerType: "conclave",
		Title:         "Feature Work",
		Status:        "active",
	}

	// 3 complete, 5 not complete = 8 total, 3 done
	shipmentSvc.shipmentTasks["SHIP-001"] = []*primary.Task{
		{ID: "TASK-001", Status: "complete"},
		{ID: "TASK-002", Status: "complete"},
		{ID: "TASK-003", Status: "complete"},
		{ID: "TASK-004", Status: "ready"},
		{ID: "TASK-005", Status: "in_progress"},
		{ID: "TASK-006", Status: "blocked"},
		{ID: "TASK-007", Status: "ready"},
		{ID: "TASK-008", Status: "ready"},
	}

	svc := NewSummaryService(commissionSvc, conclaveSvc, tomeSvc, shipmentSvc, taskSvc, noteSvc, workbenchSvc, nil, nil, nil, nil)

	req := primary.SummaryRequest{
		CommissionID: "COMM-001",
	}

	summary, err := svc.GetCommissionSummary(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ship := summary.Conclaves[0].Shipments[0]
	if ship.TasksDone != 3 {
		t.Errorf("expected 3 tasks done, got %d", ship.TasksDone)
	}
	if ship.TasksTotal != 8 {
		t.Errorf("expected 8 total tasks, got %d", ship.TasksTotal)
	}
}

func TestSummaryService_GetCommissionSummary_LibraryAndShipyard(t *testing.T) {
	// Setup mocks
	commissionSvc := newMockCommissionServiceForSummary()
	conclaveSvc := newMockConclaveServiceForSummary()
	tomeSvc := newMockTomeServiceForSummary()
	shipmentSvc := newMockShipmentServiceForSummary()
	taskSvc := newMockTaskServiceForSummary()
	noteSvc := newMockNoteServiceForSummary()
	workbenchSvc := newMockWorkbenchServiceForSummary()

	commissionSvc.commissions["COMM-001"] = &primary.Commission{
		ID:     "COMM-001",
		Title:  "Test Commission",
		Status: "active",
	}

	// Library tomes (2 open, 1 closed)
	tomeSvc.tomes["TOME-L1"] = &primary.Tome{
		ID:            "TOME-L1",
		CommissionID:  "COMM-001",
		ContainerID:   "LIB-001",
		ContainerType: "library",
		Title:         "Parked Tome 1",
		Status:        "open",
	}
	tomeSvc.tomes["TOME-L2"] = &primary.Tome{
		ID:            "TOME-L2",
		CommissionID:  "COMM-001",
		ContainerID:   "LIB-001",
		ContainerType: "library",
		Title:         "Parked Tome 2",
		Status:        "open",
	}
	tomeSvc.tomes["TOME-L3"] = &primary.Tome{
		ID:            "TOME-L3",
		CommissionID:  "COMM-001",
		ContainerID:   "LIB-001",
		ContainerType: "library",
		Title:         "Closed Library Tome",
		Status:        "closed",
	}

	// Shipyard shipments (1 active, 1 complete)
	shipmentSvc.shipments["SHIP-Y1"] = &primary.Shipment{
		ID:            "SHIP-Y1",
		CommissionID:  "COMM-001",
		ContainerID:   "YARD-001",
		ContainerType: "shipyard",
		Title:         "Parked Shipment",
		Status:        "active",
	}
	shipmentSvc.shipments["SHIP-Y2"] = &primary.Shipment{
		ID:            "SHIP-Y2",
		CommissionID:  "COMM-001",
		ContainerID:   "YARD-001",
		ContainerType: "shipyard",
		Title:         "Complete Parked Shipment",
		Status:        "complete",
	}

	svc := NewSummaryService(commissionSvc, conclaveSvc, tomeSvc, shipmentSvc, taskSvc, noteSvc, workbenchSvc, nil, nil, nil, nil)

	req := primary.SummaryRequest{
		CommissionID: "COMM-001",
	}

	summary, err := svc.GetCommissionSummary(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Library should count 2 open tomes (not the closed one)
	if summary.Library.TomeCount != 2 {
		t.Errorf("expected 2 library tomes, got %d", summary.Library.TomeCount)
	}

	// Shipyard should count 1 active shipment (not the complete one)
	if summary.Shipyard.ShipmentCount != 1 {
		t.Errorf("expected 1 shipyard shipment, got %d", summary.Shipyard.ShipmentCount)
	}
}

// Ensure interface compliance
var _ primary.SummaryService = (*SummaryServiceImpl)(nil)
