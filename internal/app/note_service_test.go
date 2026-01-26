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

// mockNoteRepository implements secondary.NoteRepository for testing.
type mockNoteRepository struct {
	notes                  map[string]*secondary.NoteRecord
	createErr              error
	getErr                 error
	updateErr              error
	deleteErr              error
	listErr                error
	commissionExistsResult bool
	commissionExistsErr    error
}

func newMockNoteRepository() *mockNoteRepository {
	return &mockNoteRepository{
		notes:                  make(map[string]*secondary.NoteRecord),
		commissionExistsResult: true,
	}
}

func (m *mockNoteRepository) Create(ctx context.Context, note *secondary.NoteRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.notes[note.ID] = note
	return nil
}

func (m *mockNoteRepository) GetByID(ctx context.Context, id string) (*secondary.NoteRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if note, ok := m.notes[id]; ok {
		return note, nil
	}
	return nil, errors.New("note not found")
}

func (m *mockNoteRepository) List(ctx context.Context, filters secondary.NoteFilters) ([]*secondary.NoteRecord, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*secondary.NoteRecord
	for _, n := range m.notes {
		if filters.CommissionID != "" && n.CommissionID != filters.CommissionID {
			continue
		}
		if filters.Type != "" && n.Type != filters.Type {
			continue
		}
		result = append(result, n)
	}
	return result, nil
}

func (m *mockNoteRepository) Update(ctx context.Context, note *secondary.NoteRecord) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if existing, ok := m.notes[note.ID]; ok {
		if note.Title != "" {
			existing.Title = note.Title
		}
		if note.Content != "" {
			existing.Content = note.Content
		}
	}
	return nil
}

func (m *mockNoteRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.notes, id)
	return nil
}

func (m *mockNoteRepository) Pin(ctx context.Context, id string) error {
	if note, ok := m.notes[id]; ok {
		note.Pinned = true
	}
	return nil
}

func (m *mockNoteRepository) Unpin(ctx context.Context, id string) error {
	if note, ok := m.notes[id]; ok {
		note.Pinned = false
	}
	return nil
}

func (m *mockNoteRepository) GetNextID(ctx context.Context) (string, error) {
	return "NOTE-001", nil
}

func (m *mockNoteRepository) GetByContainer(ctx context.Context, containerType, containerID string) ([]*secondary.NoteRecord, error) {
	var result []*secondary.NoteRecord
	for _, n := range m.notes {
		switch containerType {
		case "shipment":
			if n.ShipmentID == containerID {
				result = append(result, n)
			}
		case "investigation":
			if n.InvestigationID == containerID {
				result = append(result, n)
			}
		case "conclave":
			if n.ConclaveID == containerID {
				result = append(result, n)
			}
		case "tome":
			if n.TomeID == containerID {
				result = append(result, n)
			}
		}
	}
	return result, nil
}

func (m *mockNoteRepository) CommissionExists(ctx context.Context, commissionID string) (bool, error) {
	if m.commissionExistsErr != nil {
		return false, m.commissionExistsErr
	}
	return m.commissionExistsResult, nil
}

func (m *mockNoteRepository) ShipmentExists(ctx context.Context, shipmentID string) (bool, error) {
	return true, nil
}

func (m *mockNoteRepository) TomeExists(ctx context.Context, tomeID string) (bool, error) {
	return true, nil
}

func (m *mockNoteRepository) ConclaveExists(ctx context.Context, conclaveID string) (bool, error) {
	return true, nil
}

func (m *mockNoteRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	if note, ok := m.notes[id]; ok {
		note.Status = status
		return nil
	}
	return errors.New("note not found")
}

// ============================================================================
// Test Helper
// ============================================================================

func newTestNoteService() (*NoteServiceImpl, *mockNoteRepository) {
	noteRepo := newMockNoteRepository()
	service := NewNoteService(noteRepo)
	return service, noteRepo
}

// ============================================================================
// CreateNote Tests
// ============================================================================

func TestCreateNote_Success(t *testing.T) {
	service, _ := newTestNoteService()
	ctx := context.Background()

	resp, err := service.CreateNote(ctx, primary.CreateNoteRequest{
		CommissionID: "COMM-001",
		Title:        "Test Note",
		Content:      "Note content",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.NoteID == "" {
		t.Error("expected note ID to be set")
	}
	if resp.Note.Title != "Test Note" {
		t.Errorf("expected title 'Test Note', got '%s'", resp.Note.Title)
	}
}

func TestCreateNote_WithShipmentContainer(t *testing.T) {
	service, noteRepo := newTestNoteService()
	ctx := context.Background()

	resp, err := service.CreateNote(ctx, primary.CreateNoteRequest{
		CommissionID:  "COMM-001",
		Title:         "Shipment Note",
		Content:       "Note for shipment",
		ContainerType: "shipment",
		ContainerID:   "SHIPMENT-001",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if noteRepo.notes[resp.NoteID].ShipmentID != "SHIPMENT-001" {
		t.Errorf("expected shipment ID 'SHIPMENT-001', got '%s'", noteRepo.notes[resp.NoteID].ShipmentID)
	}
}

func TestCreateNote_WithInvestigationContainer(t *testing.T) {
	service, noteRepo := newTestNoteService()
	ctx := context.Background()

	resp, err := service.CreateNote(ctx, primary.CreateNoteRequest{
		CommissionID:  "COMM-001",
		Title:         "Investigation Note",
		Content:       "Note for investigation",
		ContainerType: "investigation",
		ContainerID:   "INV-001",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if noteRepo.notes[resp.NoteID].InvestigationID != "INV-001" {
		t.Errorf("expected investigation ID 'INV-001', got '%s'", noteRepo.notes[resp.NoteID].InvestigationID)
	}
}

func TestCreateNote_WithConclaveContainer(t *testing.T) {
	service, noteRepo := newTestNoteService()
	ctx := context.Background()

	resp, err := service.CreateNote(ctx, primary.CreateNoteRequest{
		CommissionID:  "COMM-001",
		Title:         "Conclave Note",
		Content:       "Note for conclave",
		ContainerType: "conclave",
		ContainerID:   "CON-001",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if noteRepo.notes[resp.NoteID].ConclaveID != "CON-001" {
		t.Errorf("expected conclave ID 'CON-001', got '%s'", noteRepo.notes[resp.NoteID].ConclaveID)
	}
}

func TestCreateNote_WithTomeContainer(t *testing.T) {
	service, noteRepo := newTestNoteService()
	ctx := context.Background()

	resp, err := service.CreateNote(ctx, primary.CreateNoteRequest{
		CommissionID:  "COMM-001",
		Title:         "Tome Note",
		Content:       "Note for tome",
		ContainerType: "tome",
		ContainerID:   "TOME-001",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if noteRepo.notes[resp.NoteID].TomeID != "TOME-001" {
		t.Errorf("expected tome ID 'TOME-001', got '%s'", noteRepo.notes[resp.NoteID].TomeID)
	}
}

func TestCreateNote_CommissionNotFound(t *testing.T) {
	service, noteRepo := newTestNoteService()
	ctx := context.Background()

	noteRepo.commissionExistsResult = false

	_, err := service.CreateNote(ctx, primary.CreateNoteRequest{
		CommissionID: "COMM-NONEXISTENT",
		Title:        "Test Note",
		Content:      "Note content",
	})

	if err == nil {
		t.Fatal("expected error for non-existent commission, got nil")
	}
}

// ============================================================================
// GetNote Tests
// ============================================================================

func TestGetNote_Found(t *testing.T) {
	service, noteRepo := newTestNoteService()
	ctx := context.Background()

	noteRepo.notes["NOTE-001"] = &secondary.NoteRecord{
		ID:           "NOTE-001",
		CommissionID: "COMM-001",
		Title:        "Test Note",
		Content:      "Note content",
	}

	note, err := service.GetNote(ctx, "NOTE-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if note.Title != "Test Note" {
		t.Errorf("expected title 'Test Note', got '%s'", note.Title)
	}
}

func TestGetNote_NotFound(t *testing.T) {
	service, _ := newTestNoteService()
	ctx := context.Background()

	_, err := service.GetNote(ctx, "NOTE-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent note, got nil")
	}
}

// ============================================================================
// ListNotes Tests
// ============================================================================

func TestListNotes_FilterByCommission(t *testing.T) {
	service, noteRepo := newTestNoteService()
	ctx := context.Background()

	noteRepo.notes["NOTE-001"] = &secondary.NoteRecord{
		ID:           "NOTE-001",
		CommissionID: "COMM-001",
		Title:        "Note 1",
	}
	noteRepo.notes["NOTE-002"] = &secondary.NoteRecord{
		ID:           "NOTE-002",
		CommissionID: "COMM-002",
		Title:        "Note 2",
	}

	notes, err := service.ListNotes(ctx, primary.NoteFilters{CommissionID: "COMM-001"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(notes) != 1 {
		t.Errorf("expected 1 note, got %d", len(notes))
	}
}

func TestListNotes_FilterByType(t *testing.T) {
	service, noteRepo := newTestNoteService()
	ctx := context.Background()

	noteRepo.notes["NOTE-001"] = &secondary.NoteRecord{
		ID:           "NOTE-001",
		CommissionID: "COMM-001",
		Title:        "Design Note",
		Type:         "design",
	}
	noteRepo.notes["NOTE-002"] = &secondary.NoteRecord{
		ID:           "NOTE-002",
		CommissionID: "COMM-001",
		Title:        "Code Note",
		Type:         "code",
	}

	notes, err := service.ListNotes(ctx, primary.NoteFilters{Type: "design"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(notes) != 1 {
		t.Errorf("expected 1 design note, got %d", len(notes))
	}
}

// ============================================================================
// UpdateNote Tests
// ============================================================================

func TestUpdateNote_Title(t *testing.T) {
	service, noteRepo := newTestNoteService()
	ctx := context.Background()

	noteRepo.notes["NOTE-001"] = &secondary.NoteRecord{
		ID:           "NOTE-001",
		CommissionID: "COMM-001",
		Title:        "Old Title",
		Content:      "Original content",
	}

	err := service.UpdateNote(ctx, primary.UpdateNoteRequest{
		NoteID: "NOTE-001",
		Title:  "New Title",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if noteRepo.notes["NOTE-001"].Title != "New Title" {
		t.Errorf("expected title 'New Title', got '%s'", noteRepo.notes["NOTE-001"].Title)
	}
}

func TestUpdateNote_Content(t *testing.T) {
	service, noteRepo := newTestNoteService()
	ctx := context.Background()

	noteRepo.notes["NOTE-001"] = &secondary.NoteRecord{
		ID:           "NOTE-001",
		CommissionID: "COMM-001",
		Title:        "Test Note",
		Content:      "Original content",
	}

	err := service.UpdateNote(ctx, primary.UpdateNoteRequest{
		NoteID:  "NOTE-001",
		Content: "Updated content",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if noteRepo.notes["NOTE-001"].Content != "Updated content" {
		t.Errorf("expected content 'Updated content', got '%s'", noteRepo.notes["NOTE-001"].Content)
	}
}

// ============================================================================
// DeleteNote Tests
// ============================================================================

func TestDeleteNote_Success(t *testing.T) {
	service, noteRepo := newTestNoteService()
	ctx := context.Background()

	noteRepo.notes["NOTE-001"] = &secondary.NoteRecord{
		ID:           "NOTE-001",
		CommissionID: "COMM-001",
		Title:        "Test Note",
	}

	err := service.DeleteNote(ctx, "NOTE-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, exists := noteRepo.notes["NOTE-001"]; exists {
		t.Error("expected note to be deleted")
	}
}

// ============================================================================
// Pin/Unpin Tests
// ============================================================================

func TestPinNote(t *testing.T) {
	service, noteRepo := newTestNoteService()
	ctx := context.Background()

	noteRepo.notes["NOTE-001"] = &secondary.NoteRecord{
		ID:           "NOTE-001",
		CommissionID: "COMM-001",
		Title:        "Test Note",
		Pinned:       false,
	}

	err := service.PinNote(ctx, "NOTE-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !noteRepo.notes["NOTE-001"].Pinned {
		t.Error("expected note to be pinned")
	}
}

func TestUnpinNote(t *testing.T) {
	service, noteRepo := newTestNoteService()
	ctx := context.Background()

	noteRepo.notes["NOTE-001"] = &secondary.NoteRecord{
		ID:           "NOTE-001",
		CommissionID: "COMM-001",
		Title:        "Pinned Note",
		Pinned:       true,
	}

	err := service.UnpinNote(ctx, "NOTE-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if noteRepo.notes["NOTE-001"].Pinned {
		t.Error("expected note to be unpinned")
	}
}

// ============================================================================
// GetNotesByContainer Tests
// ============================================================================

func TestGetNotesByContainer_Shipment(t *testing.T) {
	service, noteRepo := newTestNoteService()
	ctx := context.Background()

	noteRepo.notes["NOTE-001"] = &secondary.NoteRecord{
		ID:           "NOTE-001",
		CommissionID: "COMM-001",
		Title:        "Shipment Note",
		ShipmentID:   "SHIPMENT-001",
	}
	noteRepo.notes["NOTE-002"] = &secondary.NoteRecord{
		ID:           "NOTE-002",
		CommissionID: "COMM-001",
		Title:        "Other Note",
	}

	notes, err := service.GetNotesByContainer(ctx, "shipment", "SHIPMENT-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(notes) != 1 {
		t.Errorf("expected 1 shipment note, got %d", len(notes))
	}
}

func TestGetNotesByContainer_Investigation(t *testing.T) {
	service, noteRepo := newTestNoteService()
	ctx := context.Background()

	noteRepo.notes["NOTE-001"] = &secondary.NoteRecord{
		ID:              "NOTE-001",
		CommissionID:    "COMM-001",
		Title:           "Investigation Note",
		InvestigationID: "INV-001",
	}

	notes, err := service.GetNotesByContainer(ctx, "investigation", "INV-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(notes) != 1 {
		t.Errorf("expected 1 investigation note, got %d", len(notes))
	}
}

func TestGetNotesByContainer_Conclave(t *testing.T) {
	service, noteRepo := newTestNoteService()
	ctx := context.Background()

	noteRepo.notes["NOTE-001"] = &secondary.NoteRecord{
		ID:           "NOTE-001",
		CommissionID: "COMM-001",
		Title:        "Conclave Note",
		ConclaveID:   "CON-001",
	}

	notes, err := service.GetNotesByContainer(ctx, "conclave", "CON-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(notes) != 1 {
		t.Errorf("expected 1 conclave note, got %d", len(notes))
	}
}

func TestGetNotesByContainer_Tome(t *testing.T) {
	service, noteRepo := newTestNoteService()
	ctx := context.Background()

	noteRepo.notes["NOTE-001"] = &secondary.NoteRecord{
		ID:           "NOTE-001",
		CommissionID: "COMM-001",
		Title:        "Tome Note",
		TomeID:       "TOME-001",
	}

	notes, err := service.GetNotesByContainer(ctx, "tome", "TOME-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(notes) != 1 {
		t.Errorf("expected 1 tome note, got %d", len(notes))
	}
}
