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

// mockTomeRepository implements secondary.TomeRepository for testing.
type mockTomeRepository struct {
	tomes                  map[string]*secondary.TomeRecord
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

func newMockTomeRepository() *mockTomeRepository {
	return &mockTomeRepository{
		tomes:                  make(map[string]*secondary.TomeRecord),
		commissionExistsResult: true,
	}
}

func (m *mockTomeRepository) Create(ctx context.Context, tome *secondary.TomeRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.tomes[tome.ID] = tome
	return nil
}

func (m *mockTomeRepository) GetByID(ctx context.Context, id string) (*secondary.TomeRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if tome, ok := m.tomes[id]; ok {
		return tome, nil
	}
	return nil, errors.New("tome not found")
}

func (m *mockTomeRepository) List(ctx context.Context, filters secondary.TomeFilters) ([]*secondary.TomeRecord, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*secondary.TomeRecord
	for _, t := range m.tomes {
		if filters.CommissionID != "" && t.CommissionID != filters.CommissionID {
			continue
		}
		if filters.ConclaveID != "" && t.ConclaveID != filters.ConclaveID {
			continue
		}
		if filters.Status != "" && t.Status != filters.Status {
			continue
		}
		result = append(result, t)
	}
	return result, nil
}

func (m *mockTomeRepository) Update(ctx context.Context, tome *secondary.TomeRecord) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if existing, ok := m.tomes[tome.ID]; ok {
		if tome.Title != "" {
			existing.Title = tome.Title
		}
		if tome.Description != "" {
			existing.Description = tome.Description
		}
	}
	return nil
}

func (m *mockTomeRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.tomes, id)
	return nil
}

func (m *mockTomeRepository) Pin(ctx context.Context, id string) error {
	if tome, ok := m.tomes[id]; ok {
		tome.Pinned = true
	}
	return nil
}

func (m *mockTomeRepository) Unpin(ctx context.Context, id string) error {
	if tome, ok := m.tomes[id]; ok {
		tome.Pinned = false
	}
	return nil
}

func (m *mockTomeRepository) GetNextID(ctx context.Context) (string, error) {
	return "TOME-001", nil
}

func (m *mockTomeRepository) UpdateStatus(ctx context.Context, id, status string, setCompleted bool) error {
	if m.updateStatusErr != nil {
		return m.updateStatusErr
	}
	if tome, ok := m.tomes[id]; ok {
		tome.Status = status
		if setCompleted {
			tome.ClosedAt = "2026-01-20T10:00:00Z"
		}
	}
	return nil
}

func (m *mockTomeRepository) GetByWorkbench(ctx context.Context, workbenchID string) ([]*secondary.TomeRecord, error) {
	var result []*secondary.TomeRecord
	for _, t := range m.tomes {
		if t.AssignedWorkbenchID == workbenchID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockTomeRepository) GetByConclave(ctx context.Context, conclaveID string) ([]*secondary.TomeRecord, error) {
	var result []*secondary.TomeRecord
	for _, t := range m.tomes {
		if t.ConclaveID == conclaveID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockTomeRepository) AssignWorkbench(ctx context.Context, tomeID, workbenchID string) error {
	if m.assignWorkbenchErr != nil {
		return m.assignWorkbenchErr
	}
	if tome, ok := m.tomes[tomeID]; ok {
		tome.AssignedWorkbenchID = workbenchID
	}
	return nil
}

func (m *mockTomeRepository) CommissionExists(ctx context.Context, commissionID string) (bool, error) {
	if m.commissionExistsErr != nil {
		return false, m.commissionExistsErr
	}
	return m.commissionExistsResult, nil
}

// mockNoteServiceForTome implements minimal NoteService for tome tests.
type mockNoteServiceForTome struct {
	notes map[string][]*primary.Note // containerID -> notes
}

func newMockNoteServiceForTome() *mockNoteServiceForTome {
	return &mockNoteServiceForTome{
		notes: make(map[string][]*primary.Note),
	}
}

func (m *mockNoteServiceForTome) CreateNote(ctx context.Context, req primary.CreateNoteRequest) (*primary.CreateNoteResponse, error) {
	return nil, nil
}

func (m *mockNoteServiceForTome) GetNote(ctx context.Context, noteID string) (*primary.Note, error) {
	return nil, nil
}

func (m *mockNoteServiceForTome) ListNotes(ctx context.Context, filters primary.NoteFilters) ([]*primary.Note, error) {
	return nil, nil
}

func (m *mockNoteServiceForTome) UpdateNote(ctx context.Context, req primary.UpdateNoteRequest) error {
	return nil
}

func (m *mockNoteServiceForTome) DeleteNote(ctx context.Context, noteID string) error {
	return nil
}

func (m *mockNoteServiceForTome) PinNote(ctx context.Context, noteID string) error {
	return nil
}

func (m *mockNoteServiceForTome) UnpinNote(ctx context.Context, noteID string) error {
	return nil
}

func (m *mockNoteServiceForTome) GetNotesByContainer(ctx context.Context, containerType, containerID string) ([]*primary.Note, error) {
	if notes, ok := m.notes[containerID]; ok {
		return notes, nil
	}
	return []*primary.Note{}, nil
}

func (m *mockNoteServiceForTome) CloseNote(ctx context.Context, noteID string) error {
	return nil
}

func (m *mockNoteServiceForTome) ReopenNote(ctx context.Context, noteID string) error {
	return nil
}

func (m *mockNoteServiceForTome) MoveNote(ctx context.Context, req primary.MoveNoteRequest) error {
	return nil
}

// ============================================================================
// Test Helper
// ============================================================================

func newTestTomeService() (*TomeServiceImpl, *mockTomeRepository, *mockNoteServiceForTome) {
	tomeRepo := newMockTomeRepository()
	noteService := newMockNoteServiceForTome()
	service := NewTomeService(tomeRepo, noteService)
	return service, tomeRepo, noteService
}

// ============================================================================
// CreateTome Tests
// ============================================================================

func TestCreateTome_Success(t *testing.T) {
	service, _, _ := newTestTomeService()
	ctx := context.Background()

	resp, err := service.CreateTome(ctx, primary.CreateTomeRequest{
		CommissionID: "COMM-001",
		Title:        "Test Tome",
		Description:  "A test tome",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.TomeID == "" {
		t.Error("expected tome ID to be set")
	}
	if resp.Tome.Title != "Test Tome" {
		t.Errorf("expected title 'Test Tome', got '%s'", resp.Tome.Title)
	}
	if resp.Tome.Status != "open" {
		t.Errorf("expected status 'active', got '%s'", resp.Tome.Status)
	}
}

func TestCreateTome_CommissionNotFound(t *testing.T) {
	service, tomeRepo, _ := newTestTomeService()
	ctx := context.Background()

	tomeRepo.commissionExistsResult = false

	_, err := service.CreateTome(ctx, primary.CreateTomeRequest{
		CommissionID: "COMM-NONEXISTENT",
		Title:        "Test Tome",
		Description:  "A test tome",
	})

	if err == nil {
		t.Fatal("expected error for non-existent commission, got nil")
	}
}

// ============================================================================
// GetTome Tests
// ============================================================================

func TestGetTome_Found(t *testing.T) {
	service, tomeRepo, _ := newTestTomeService()
	ctx := context.Background()

	tomeRepo.tomes["TOME-001"] = &secondary.TomeRecord{
		ID:           "TOME-001",
		CommissionID: "COMM-001",
		Title:        "Test Tome",
		Status:       "open",
	}

	tome, err := service.GetTome(ctx, "TOME-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tome.Title != "Test Tome" {
		t.Errorf("expected title 'Test Tome', got '%s'", tome.Title)
	}
}

func TestGetTome_NotFound(t *testing.T) {
	service, _, _ := newTestTomeService()
	ctx := context.Background()

	_, err := service.GetTome(ctx, "TOME-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent tome, got nil")
	}
}

// ============================================================================
// ListTomes Tests
// ============================================================================

func TestListTomes_FilterByCommission(t *testing.T) {
	service, tomeRepo, _ := newTestTomeService()
	ctx := context.Background()

	tomeRepo.tomes["TOME-001"] = &secondary.TomeRecord{
		ID:           "TOME-001",
		CommissionID: "COMM-001",
		Title:        "Tome 1",
		Status:       "open",
	}
	tomeRepo.tomes["TOME-002"] = &secondary.TomeRecord{
		ID:           "TOME-002",
		CommissionID: "COMM-002",
		Title:        "Tome 2",
		Status:       "open",
	}

	tomes, err := service.ListTomes(ctx, primary.TomeFilters{CommissionID: "COMM-001"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tomes) != 1 {
		t.Errorf("expected 1 tome, got %d", len(tomes))
	}
}

func TestListTomes_FilterByStatus(t *testing.T) {
	service, tomeRepo, _ := newTestTomeService()
	ctx := context.Background()

	tomeRepo.tomes["TOME-001"] = &secondary.TomeRecord{
		ID:           "TOME-001",
		CommissionID: "COMM-001",
		Title:        "Active Tome",
		Status:       "open",
	}
	tomeRepo.tomes["TOME-002"] = &secondary.TomeRecord{
		ID:           "TOME-002",
		CommissionID: "COMM-001",
		Title:        "Paused Tome",
		Status:       "paused",
	}

	tomes, err := service.ListTomes(ctx, primary.TomeFilters{Status: "open"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tomes) != 1 {
		t.Errorf("expected 1 active tome, got %d", len(tomes))
	}
}

// ============================================================================
// CloseTome Tests
// ============================================================================

func TestCloseTome_UnpinnedAllowed(t *testing.T) {
	service, tomeRepo, _ := newTestTomeService()
	ctx := context.Background()

	tomeRepo.tomes["TOME-001"] = &secondary.TomeRecord{
		ID:           "TOME-001",
		CommissionID: "COMM-001",
		Title:        "Test Tome",
		Status:       "open",
		Pinned:       false,
	}

	err := service.CloseTome(ctx, "TOME-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tomeRepo.tomes["TOME-001"].Status != "closed" {
		t.Errorf("expected status 'complete', got '%s'", tomeRepo.tomes["TOME-001"].Status)
	}
}

func TestCloseTome_PinnedBlocked(t *testing.T) {
	service, tomeRepo, _ := newTestTomeService()
	ctx := context.Background()

	tomeRepo.tomes["TOME-001"] = &secondary.TomeRecord{
		ID:           "TOME-001",
		CommissionID: "COMM-001",
		Title:        "Pinned Tome",
		Status:       "open",
		Pinned:       true,
	}

	err := service.CloseTome(ctx, "TOME-001")

	if err == nil {
		t.Fatal("expected error for completing pinned tome, got nil")
	}
}

func TestCloseTome_NotFound(t *testing.T) {
	service, _, _ := newTestTomeService()
	ctx := context.Background()

	err := service.CloseTome(ctx, "TOME-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent tome, got nil")
	}
}

// ============================================================================
// Pin/Unpin Tests
// ============================================================================

func TestPinTome(t *testing.T) {
	service, tomeRepo, _ := newTestTomeService()
	ctx := context.Background()

	tomeRepo.tomes["TOME-001"] = &secondary.TomeRecord{
		ID:           "TOME-001",
		CommissionID: "COMM-001",
		Title:        "Test Tome",
		Status:       "open",
		Pinned:       false,
	}

	err := service.PinTome(ctx, "TOME-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !tomeRepo.tomes["TOME-001"].Pinned {
		t.Error("expected tome to be pinned")
	}
}

func TestUnpinTome(t *testing.T) {
	service, tomeRepo, _ := newTestTomeService()
	ctx := context.Background()

	tomeRepo.tomes["TOME-001"] = &secondary.TomeRecord{
		ID:           "TOME-001",
		CommissionID: "COMM-001",
		Title:        "Pinned Tome",
		Status:       "open",
		Pinned:       true,
	}

	err := service.UnpinTome(ctx, "TOME-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tomeRepo.tomes["TOME-001"].Pinned {
		t.Error("expected tome to be unpinned")
	}
}

// ============================================================================
// UpdateTome Tests
// ============================================================================

func TestUpdateTome_Title(t *testing.T) {
	service, tomeRepo, _ := newTestTomeService()
	ctx := context.Background()

	tomeRepo.tomes["TOME-001"] = &secondary.TomeRecord{
		ID:           "TOME-001",
		CommissionID: "COMM-001",
		Title:        "Old Title",
		Description:  "Original description",
		Status:       "open",
	}

	err := service.UpdateTome(ctx, primary.UpdateTomeRequest{
		TomeID: "TOME-001",
		Title:  "New Title",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tomeRepo.tomes["TOME-001"].Title != "New Title" {
		t.Errorf("expected title 'New Title', got '%s'", tomeRepo.tomes["TOME-001"].Title)
	}
}

// ============================================================================
// DeleteTome Tests
// ============================================================================

func TestDeleteTome_Success(t *testing.T) {
	service, tomeRepo, _ := newTestTomeService()
	ctx := context.Background()

	tomeRepo.tomes["TOME-001"] = &secondary.TomeRecord{
		ID:           "TOME-001",
		CommissionID: "COMM-001",
		Title:        "Test Tome",
		Status:       "open",
	}

	err := service.DeleteTome(ctx, "TOME-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, exists := tomeRepo.tomes["TOME-001"]; exists {
		t.Error("expected tome to be deleted")
	}
}

// ============================================================================
// AssignTomeToWorkbench Tests
// ============================================================================

func TestAssignTomeToWorkbench_Success(t *testing.T) {
	service, tomeRepo, _ := newTestTomeService()
	ctx := context.Background()

	tomeRepo.tomes["TOME-001"] = &secondary.TomeRecord{
		ID:           "TOME-001",
		CommissionID: "COMM-001",
		Title:        "Test Tome",
		Status:       "open",
	}

	err := service.AssignTomeToWorkbench(ctx, "TOME-001", "BENCH-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tomeRepo.tomes["TOME-001"].AssignedWorkbenchID != "BENCH-001" {
		t.Errorf("expected workbench ID 'BENCH-001', got '%s'", tomeRepo.tomes["TOME-001"].AssignedWorkbenchID)
	}
}

func TestAssignTomeToWorkbench_TomeNotFound(t *testing.T) {
	service, _, _ := newTestTomeService()
	ctx := context.Background()

	err := service.AssignTomeToWorkbench(ctx, "TOME-NONEXISTENT", "BENCH-001")

	if err == nil {
		t.Fatal("expected error for non-existent tome, got nil")
	}
}

// ============================================================================
// GetTomesByWorkbench Tests
// ============================================================================

func TestGetTomesByWorkbench_Success(t *testing.T) {
	service, tomeRepo, _ := newTestTomeService()
	ctx := context.Background()

	tomeRepo.tomes["TOME-001"] = &secondary.TomeRecord{
		ID:                  "TOME-001",
		CommissionID:        "COMM-001",
		Title:               "Assigned Tome",
		Status:              "active",
		AssignedWorkbenchID: "BENCH-001",
	}
	tomeRepo.tomes["TOME-002"] = &secondary.TomeRecord{
		ID:           "TOME-002",
		CommissionID: "COMM-001",
		Title:        "Unassigned Tome",
		Status:       "open",
	}

	tomes, err := service.GetTomesByWorkbench(ctx, "BENCH-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tomes) != 1 {
		t.Errorf("expected 1 tome, got %d", len(tomes))
	}
}

// ============================================================================
// GetTomeNotes Tests
// ============================================================================

func TestGetTomeNotes_Success(t *testing.T) {
	service, _, noteService := newTestTomeService()
	ctx := context.Background()

	noteService.notes["TOME-001"] = []*primary.Note{
		{ID: "NOTE-001", Title: "Note 1"},
		{ID: "NOTE-002", Title: "Note 2"},
	}

	notes, err := service.GetTomeNotes(ctx, "TOME-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(notes) != 2 {
		t.Errorf("expected 2 notes, got %d", len(notes))
	}
}
