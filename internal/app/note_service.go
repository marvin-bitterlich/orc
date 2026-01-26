package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// NoteServiceImpl implements the NoteService interface.
type NoteServiceImpl struct {
	noteRepo secondary.NoteRepository
}

// NewNoteService creates a new NoteService with injected dependencies.
func NewNoteService(noteRepo secondary.NoteRepository) *NoteServiceImpl {
	return &NoteServiceImpl{
		noteRepo: noteRepo,
	}
}

// CreateNote creates a new note.
func (s *NoteServiceImpl) CreateNote(ctx context.Context, req primary.CreateNoteRequest) (*primary.CreateNoteResponse, error) {
	// Validate commission exists
	exists, err := s.noteRepo.CommissionExists(ctx, req.CommissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate commission: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("commission %s not found", req.CommissionID)
	}

	// Get next ID
	nextID, err := s.noteRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate note ID: %w", err)
	}

	// Build record with container assignment
	record := &secondary.NoteRecord{
		ID:           nextID,
		CommissionID: req.CommissionID,
		Title:        req.Title,
		Content:      req.Content,
		Type:         req.Type,
	}

	// Set appropriate container FK based on container type
	if req.ContainerID != "" {
		switch req.ContainerType {
		case "shipment":
			record.ShipmentID = req.ContainerID
		case "investigation":
			record.InvestigationID = req.ContainerID
		case "conclave":
			record.ConclaveID = req.ContainerID
		case "tome":
			record.TomeID = req.ContainerID
		}
	}

	if err := s.noteRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	// Fetch created note
	created, err := s.noteRepo.GetByID(ctx, nextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created note: %w", err)
	}

	return &primary.CreateNoteResponse{
		NoteID: created.ID,
		Note:   s.recordToNote(created),
	}, nil
}

// GetNote retrieves a note by ID.
func (s *NoteServiceImpl) GetNote(ctx context.Context, noteID string) (*primary.Note, error) {
	record, err := s.noteRepo.GetByID(ctx, noteID)
	if err != nil {
		return nil, err
	}
	return s.recordToNote(record), nil
}

// ListNotes lists notes with optional filters.
func (s *NoteServiceImpl) ListNotes(ctx context.Context, filters primary.NoteFilters) ([]*primary.Note, error) {
	records, err := s.noteRepo.List(ctx, secondary.NoteFilters{
		Type:         filters.Type,
		CommissionID: filters.CommissionID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list notes: %w", err)
	}

	notes := make([]*primary.Note, len(records))
	for i, r := range records {
		notes[i] = s.recordToNote(r)
	}
	return notes, nil
}

// UpdateNote updates a note's title and/or content.
func (s *NoteServiceImpl) UpdateNote(ctx context.Context, req primary.UpdateNoteRequest) error {
	record := &secondary.NoteRecord{
		ID:      req.NoteID,
		Title:   req.Title,
		Content: req.Content,
	}
	return s.noteRepo.Update(ctx, record)
}

// DeleteNote deletes a note.
func (s *NoteServiceImpl) DeleteNote(ctx context.Context, noteID string) error {
	return s.noteRepo.Delete(ctx, noteID)
}

// PinNote pins a note.
func (s *NoteServiceImpl) PinNote(ctx context.Context, noteID string) error {
	return s.noteRepo.Pin(ctx, noteID)
}

// UnpinNote unpins a note.
func (s *NoteServiceImpl) UnpinNote(ctx context.Context, noteID string) error {
	return s.noteRepo.Unpin(ctx, noteID)
}

// GetNotesByContainer retrieves notes for a specific container.
func (s *NoteServiceImpl) GetNotesByContainer(ctx context.Context, containerType, containerID string) ([]*primary.Note, error) {
	records, err := s.noteRepo.GetByContainer(ctx, containerType, containerID)
	if err != nil {
		return nil, err
	}

	notes := make([]*primary.Note, len(records))
	for i, r := range records {
		notes[i] = s.recordToNote(r)
	}
	return notes, nil
}

// CloseNote closes a note.
func (s *NoteServiceImpl) CloseNote(ctx context.Context, noteID string) error {
	// Get current note to verify it exists and check status
	note, err := s.noteRepo.GetByID(ctx, noteID)
	if err != nil {
		return err
	}

	if note.Status == "closed" {
		return fmt.Errorf("note %s is already closed", noteID)
	}

	return s.noteRepo.UpdateStatus(ctx, noteID, "closed")
}

// ReopenNote reopens a closed note.
func (s *NoteServiceImpl) ReopenNote(ctx context.Context, noteID string) error {
	// Get current note to verify it exists and check status
	note, err := s.noteRepo.GetByID(ctx, noteID)
	if err != nil {
		return err
	}

	if note.Status == "open" {
		return fmt.Errorf("note %s is already open", noteID)
	}

	return s.noteRepo.UpdateStatus(ctx, noteID, "open")
}

// MoveNote moves a note to a different container.
func (s *NoteServiceImpl) MoveNote(ctx context.Context, req primary.MoveNoteRequest) error {
	// Verify note exists
	_, err := s.noteRepo.GetByID(ctx, req.NoteID)
	if err != nil {
		return err
	}

	// Count how many targets are specified - exactly one required
	targetCount := 0
	if req.ToTomeID != "" {
		targetCount++
	}
	if req.ToShipmentID != "" {
		targetCount++
	}
	if req.ToConclaveID != "" {
		targetCount++
	}

	if targetCount == 0 {
		return fmt.Errorf("must specify exactly one target container (--to-tome, --to-shipment, or --to-conclave)")
	}
	if targetCount > 1 {
		return fmt.Errorf("cannot specify multiple target containers")
	}

	// Validate target container exists and build update record
	record := &secondary.NoteRecord{ID: req.NoteID}

	if req.ToTomeID != "" {
		exists, err := s.noteRepo.TomeExists(ctx, req.ToTomeID)
		if err != nil {
			return fmt.Errorf("failed to validate tome: %w", err)
		}
		if !exists {
			return fmt.Errorf("tome %s not found", req.ToTomeID)
		}
		record.TomeID = req.ToTomeID
	}

	if req.ToShipmentID != "" {
		exists, err := s.noteRepo.ShipmentExists(ctx, req.ToShipmentID)
		if err != nil {
			return fmt.Errorf("failed to validate shipment: %w", err)
		}
		if !exists {
			return fmt.Errorf("shipment %s not found", req.ToShipmentID)
		}
		record.ShipmentID = req.ToShipmentID
	}

	if req.ToConclaveID != "" {
		exists, err := s.noteRepo.ConclaveExists(ctx, req.ToConclaveID)
		if err != nil {
			return fmt.Errorf("failed to validate conclave: %w", err)
		}
		if !exists {
			return fmt.Errorf("conclave %s not found", req.ToConclaveID)
		}
		record.ConclaveID = req.ToConclaveID
	}

	return s.noteRepo.Update(ctx, record)
}

// Helper methods

func (s *NoteServiceImpl) recordToNote(r *secondary.NoteRecord) *primary.Note {
	return &primary.Note{
		ID:               r.ID,
		CommissionID:     r.CommissionID,
		Title:            r.Title,
		Content:          r.Content,
		Type:             r.Type,
		Status:           r.Status,
		ShipmentID:       r.ShipmentID,
		InvestigationID:  r.InvestigationID,
		ConclaveID:       r.ConclaveID,
		TomeID:           r.TomeID,
		Pinned:           r.Pinned,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
		ClosedAt:         r.ClosedAt,
		PromotedFromID:   r.PromotedFromID,
		PromotedFromType: r.PromotedFromType,
	}
}

// Ensure NoteServiceImpl implements the interface
var _ primary.NoteService = (*NoteServiceImpl)(nil)
