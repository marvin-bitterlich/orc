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

func (m *mockShipmentServiceForSummary) UpdateStatus(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockShipmentServiceForSummary) TriggerAutoTransition(_ context.Context, _, _ string) (string, error) {
	return "", nil
}

func (m *mockShipmentServiceForSummary) DeployShipment(_ context.Context, _ string) error {
	return nil
}

func (m *mockShipmentServiceForSummary) VerifyShipment(_ context.Context, _ string) error {
	return nil
}

func (m *mockShipmentServiceForSummary) SetStatus(_ context.Context, _, _ string, _ bool) error {
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

func (m *mockTaskServiceForSummary) DeleteTask(_ context.Context, _ string, _ bool) error {
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

func (m *mockNoteServiceForSummary) SetNoteInFlight(_ context.Context, _ string) error {
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

func (m *mockWorkbenchServiceForSummary) ArchiveWorkbench(_ context.Context, _ string) error {
	return nil
}

func (m *mockWorkbenchServiceForSummary) GetWorkbenchesByFocusedID(_ context.Context, _ string) ([]*primary.Workbench, error) {
	return nil, nil
}

// ============================================================================
// Tests for Flat Summary Structure
// ============================================================================

func TestSummaryService_GetCommissionSummary_FlatStructure(t *testing.T) {
	// Setup mocks
	commissionSvc := newMockCommissionServiceForSummary()
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

	// Create tome directly under commission
	tomeSvc.tomes["TOME-001"] = &primary.Tome{
		ID:           "TOME-001",
		CommissionID: "COMM-001",
		Title:        "Design Notes",
		Status:       "open",
	}

	// Add notes to tome
	tomeSvc.tomeNotes["TOME-001"] = []*primary.Note{
		{ID: "NOTE-001", Title: "Note 1", Status: "open"},
		{ID: "NOTE-002", Title: "Note 2", Status: "open"},
		{ID: "NOTE-003", Title: "Note 3", Status: "closed"},
	}

	// Create shipment directly under commission with assigned workbench
	shipmentSvc.shipments["SHIP-001"] = &primary.Shipment{
		ID:                  "SHIP-001",
		CommissionID:        "COMM-001",
		Title:               "Bug Fixes",
		Status:              "active",
		AssignedWorkbenchID: "BENCH-001",
	}

	// Add tasks to shipment
	shipmentSvc.shipmentTasks["SHIP-001"] = []*primary.Task{
		{ID: "TASK-001", Title: "Task 1", Status: "closed"},
		{ID: "TASK-002", Title: "Task 2", Status: "open"},
		{ID: "TASK-003", Title: "Task 3", Status: "open"},
	}

	// Add workbench
	workbenchSvc.workbenches["BENCH-001"] = &primary.Workbench{
		ID:   "BENCH-001",
		Name: "bench-alpha",
	}

	// Create another tome
	tomeSvc.tomes["TOME-002"] = &primary.Tome{
		ID:           "TOME-002",
		CommissionID: "COMM-001",
		Title:        "Root Notes",
		Status:       "open",
	}

	// Create service
	svc := NewSummaryService(commissionSvc, tomeSvc, shipmentSvc, taskSvc, noteSvc, workbenchSvc, nil, nil, nil)

	// Request summary
	req := primary.SummaryRequest{
		CommissionID: "COMM-001",
		WorkshopID:   "WORK-001",
		FocusID:      "SHIP-001",
	}

	summary, err := svc.GetCommissionSummary(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify commission
	if summary.ID != "COMM-001" {
		t.Errorf("expected commission ID COMM-001, got %s", summary.ID)
	}

	// Verify flat shipments list
	if len(summary.Shipments) != 1 {
		t.Errorf("expected 1 shipment, got %d", len(summary.Shipments))
	}
	if summary.Shipments[0].ID != "SHIP-001" {
		t.Errorf("expected shipment ID SHIP-001, got %s", summary.Shipments[0].ID)
	}
	if !summary.Shipments[0].IsFocused {
		t.Error("expected shipment to be focused")
	}
	if summary.Shipments[0].BenchID != "BENCH-001" {
		t.Errorf("expected bench ID BENCH-001, got %s", summary.Shipments[0].BenchID)
	}
	if summary.Shipments[0].TasksDone != 1 {
		t.Errorf("expected 1 task done, got %d", summary.Shipments[0].TasksDone)
	}
	if summary.Shipments[0].TasksTotal != 3 {
		t.Errorf("expected 3 total tasks, got %d", summary.Shipments[0].TasksTotal)
	}

	// Verify flat tomes list
	if len(summary.Tomes) != 2 {
		t.Errorf("expected 2 tomes, got %d", len(summary.Tomes))
	}
}

func TestSummaryService_GetCommissionSummary_ShowsAllShipments(t *testing.T) {
	// Setup mocks
	commissionSvc := newMockCommissionServiceForSummary()
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

	// Create shipment assigned to one workbench
	shipmentSvc.shipments["SHIP-001"] = &primary.Shipment{
		ID:                  "SHIP-001",
		CommissionID:        "COMM-001",
		Title:               "My Shipment",
		Status:              "active",
		AssignedWorkbenchID: "BENCH-001",
	}

	// Create shipment assigned to different workbench
	shipmentSvc.shipments["SHIP-002"] = &primary.Shipment{
		ID:                  "SHIP-002",
		CommissionID:        "COMM-001",
		Title:               "Other Shipment",
		Status:              "active",
		AssignedWorkbenchID: "BENCH-002",
	}

	// Create unassigned shipment
	shipmentSvc.shipments["SHIP-003"] = &primary.Shipment{
		ID:                  "SHIP-003",
		CommissionID:        "COMM-001",
		Title:               "Unassigned Shipment",
		Status:              "active",
		AssignedWorkbenchID: "",
	}

	// Create service
	svc := NewSummaryService(commissionSvc, tomeSvc, shipmentSvc, taskSvc, noteSvc, workbenchSvc, nil, nil, nil)

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

	// All 3 shipments should be visible in flat list
	if len(summary.Shipments) != 3 {
		t.Errorf("expected 3 visible shipments, got %d", len(summary.Shipments))
	}
}

func TestSummaryService_GetCommissionSummary_TaskCounting(t *testing.T) {
	// Setup mocks
	commissionSvc := newMockCommissionServiceForSummary()
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

	shipmentSvc.shipments["SHIP-001"] = &primary.Shipment{
		ID:           "SHIP-001",
		CommissionID: "COMM-001",
		Title:        "Feature Work",
		Status:       "active",
	}

	// 3 closed, 5 not closed = 8 total, 3 done
	shipmentSvc.shipmentTasks["SHIP-001"] = []*primary.Task{
		{ID: "TASK-001", Status: "closed"},
		{ID: "TASK-002", Status: "closed"},
		{ID: "TASK-003", Status: "closed"},
		{ID: "TASK-004", Status: "open"},
		{ID: "TASK-005", Status: "in-progress"},
		{ID: "TASK-006", Status: "blocked"},
		{ID: "TASK-007", Status: "open"},
		{ID: "TASK-008", Status: "open"},
	}

	svc := NewSummaryService(commissionSvc, tomeSvc, shipmentSvc, taskSvc, noteSvc, workbenchSvc, nil, nil, nil)

	req := primary.SummaryRequest{
		CommissionID: "COMM-001",
	}

	summary, err := svc.GetCommissionSummary(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ship := summary.Shipments[0]
	if ship.TasksDone != 3 {
		t.Errorf("expected 3 tasks done, got %d", ship.TasksDone)
	}
	if ship.TasksTotal != 8 {
		t.Errorf("expected 8 total tasks, got %d", ship.TasksTotal)
	}
}

func TestSummaryService_GetCommissionSummary_HidesClosedAndComplete(t *testing.T) {
	// Setup mocks
	commissionSvc := newMockCommissionServiceForSummary()
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

	// Open tomes (should appear)
	tomeSvc.tomes["TOME-001"] = &primary.Tome{
		ID:           "TOME-001",
		CommissionID: "COMM-001",
		Title:        "Open Tome 1",
		Status:       "open",
	}
	tomeSvc.tomes["TOME-002"] = &primary.Tome{
		ID:           "TOME-002",
		CommissionID: "COMM-001",
		Title:        "Open Tome 2",
		Status:       "open",
	}
	// Closed tome (should be hidden)
	tomeSvc.tomes["TOME-003"] = &primary.Tome{
		ID:           "TOME-003",
		CommissionID: "COMM-001",
		Title:        "Closed Tome",
		Status:       "closed",
	}

	// Active shipment (should appear)
	shipmentSvc.shipments["SHIP-001"] = &primary.Shipment{
		ID:           "SHIP-001",
		CommissionID: "COMM-001",
		Title:        "Active Shipment",
		Status:       "active",
	}
	// Complete shipment (should be hidden)
	shipmentSvc.shipments["SHIP-002"] = &primary.Shipment{
		ID:           "SHIP-002",
		CommissionID: "COMM-001",
		Title:        "Complete Shipment",
		Status:       "closed",
	}

	svc := NewSummaryService(commissionSvc, tomeSvc, shipmentSvc, taskSvc, noteSvc, workbenchSvc, nil, nil, nil)

	req := primary.SummaryRequest{
		CommissionID: "COMM-001",
	}

	summary, err := svc.GetCommissionSummary(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 2 open tomes (not the closed one)
	if len(summary.Tomes) != 2 {
		t.Errorf("expected 2 tomes, got %d", len(summary.Tomes))
	}

	// Should have 1 active shipment (not the complete one)
	if len(summary.Shipments) != 1 {
		t.Errorf("expected 1 shipment, got %d", len(summary.Shipments))
	}
}

func TestSummaryService_GetCommissionSummary_FocusedCommission(t *testing.T) {
	// Setup mocks
	commissionSvc := newMockCommissionServiceForSummary()
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

	shipmentSvc.shipments["SHIP-001"] = &primary.Shipment{
		ID:           "SHIP-001",
		CommissionID: "COMM-001",
		Title:        "Test Shipment",
		Status:       "active",
	}

	svc := NewSummaryService(commissionSvc, tomeSvc, shipmentSvc, taskSvc, noteSvc, workbenchSvc, nil, nil, nil)

	// Test with focus on shipment in this commission
	req := primary.SummaryRequest{
		CommissionID: "COMM-001",
		FocusID:      "SHIP-001",
	}

	summary, err := svc.GetCommissionSummary(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Commission should be marked as focused
	if !summary.IsFocusedCommission {
		t.Error("expected IsFocusedCommission to be true when focus is in this commission")
	}

	// Test without focus
	reqNoFocus := primary.SummaryRequest{
		CommissionID: "COMM-001",
	}

	summaryNoFocus, err := svc.GetCommissionSummary(context.Background(), reqNoFocus)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Commission should not be marked as focused
	if summaryNoFocus.IsFocusedCommission {
		t.Error("expected IsFocusedCommission to be false when no focus")
	}
}

func TestSummaryService_GetCommissionSummary_TomeNoteCount(t *testing.T) {
	// Setup mocks
	commissionSvc := newMockCommissionServiceForSummary()
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

	tomeSvc.tomes["TOME-001"] = &primary.Tome{
		ID:           "TOME-001",
		CommissionID: "COMM-001",
		Title:        "Notes Tome",
		Status:       "open",
	}

	// 2 open notes, 1 closed note (should only count open)
	tomeSvc.tomeNotes["TOME-001"] = []*primary.Note{
		{ID: "NOTE-001", Title: "Open Note 1", Status: "open"},
		{ID: "NOTE-002", Title: "Open Note 2", Status: "open"},
		{ID: "NOTE-003", Title: "Closed Note", Status: "closed"},
	}

	svc := NewSummaryService(commissionSvc, tomeSvc, shipmentSvc, taskSvc, noteSvc, workbenchSvc, nil, nil, nil)

	req := primary.SummaryRequest{
		CommissionID: "COMM-001",
	}

	summary, err := svc.GetCommissionSummary(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(summary.Tomes) != 1 {
		t.Fatalf("expected 1 tome, got %d", len(summary.Tomes))
	}

	// Should count only open notes
	if summary.Tomes[0].NoteCount != 2 {
		t.Errorf("expected 2 open notes, got %d", summary.Tomes[0].NoteCount)
	}
}

func TestSummaryService_GetCommissionSummary_TomeExpansion(t *testing.T) {
	tests := []struct {
		name             string
		focusID          string
		wantNotesLen     int // number of note summaries in TOME-001
		wantTomeExpanded bool
	}{
		{
			name:             "no focus does not expand tome notes",
			focusID:          "",
			wantNotesLen:     0,
			wantTomeExpanded: false,
		},
		{
			name:             "focus on tome expands its notes",
			focusID:          "TOME-001",
			wantNotesLen:     2,
			wantTomeExpanded: true,
		},
		{
			name:             "focus on commission expands tome notes",
			focusID:          "COMM-001",
			wantNotesLen:     2,
			wantTomeExpanded: false, // tome itself not focused, but notes expanded
		},
		{
			name:             "focus on shipment expands tome notes",
			focusID:          "SHIP-001",
			wantNotesLen:     2,
			wantTomeExpanded: false, // tome itself not focused, but notes expanded
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commissionSvc := newMockCommissionServiceForSummary()
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

			tomeSvc.tomes["TOME-001"] = &primary.Tome{
				ID:           "TOME-001",
				CommissionID: "COMM-001",
				Title:        "Design Notes",
				Status:       "open",
			}

			tomeSvc.tomeNotes["TOME-001"] = []*primary.Note{
				{ID: "NOTE-001", Title: "Note 1", Status: "open"},
				{ID: "NOTE-002", Title: "Note 2", Status: "open"},
				{ID: "NOTE-003", Title: "Note 3", Status: "closed"},
			}

			shipmentSvc.shipments["SHIP-001"] = &primary.Shipment{
				ID:           "SHIP-001",
				CommissionID: "COMM-001",
				Title:        "Feature Work",
				Status:       "active",
			}

			svc := NewSummaryService(commissionSvc, tomeSvc, shipmentSvc, taskSvc, noteSvc, workbenchSvc, nil, nil, nil)

			req := primary.SummaryRequest{
				CommissionID: "COMM-001",
				FocusID:      tt.focusID,
			}

			summary, err := svc.GetCommissionSummary(context.Background(), req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(summary.Tomes) != 1 {
				t.Fatalf("expected 1 tome, got %d", len(summary.Tomes))
			}

			tome := summary.Tomes[0]
			if len(tome.Notes) != tt.wantNotesLen {
				t.Errorf("expected %d expanded notes, got %d", tt.wantNotesLen, len(tome.Notes))
			}
			if tome.IsFocused != tt.wantTomeExpanded {
				t.Errorf("expected IsFocused=%v, got %v", tt.wantTomeExpanded, tome.IsFocused)
			}
		})
	}
}

// Ensure interface compliance
var _ primary.SummaryService = (*SummaryServiceImpl)(nil)
