package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// TomeServiceImpl implements the TomeService interface.
type TomeServiceImpl struct {
	tomeRepo    secondary.TomeRepository
	noteService primary.NoteService
}

// NewTomeService creates a new TomeService with injected dependencies.
func NewTomeService(
	tomeRepo secondary.TomeRepository,
	noteService primary.NoteService,
) *TomeServiceImpl {
	return &TomeServiceImpl{
		tomeRepo:    tomeRepo,
		noteService: noteService,
	}
}

// CreateTome creates a new tome (knowledge container).
func (s *TomeServiceImpl) CreateTome(ctx context.Context, req primary.CreateTomeRequest) (*primary.CreateTomeResponse, error) {
	// Validate container type if provided
	if req.ContainerType != "" && req.ContainerType != "conclave" {
		return nil, fmt.Errorf("invalid container type %q: must be 'conclave'", req.ContainerType)
	}

	// Validate commission exists
	exists, err := s.tomeRepo.CommissionExists(ctx, req.CommissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate commission: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("commission %s not found", req.CommissionID)
	}

	// Get next ID
	nextID, err := s.tomeRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tome ID: %w", err)
	}

	// Set ConclaveID for backwards compatibility if container is a conclave
	conclaveID := req.ConclaveID
	if req.ContainerType == "conclave" && conclaveID == "" {
		conclaveID = req.ContainerID
	}

	// Create record
	record := &secondary.TomeRecord{
		ID:            nextID,
		CommissionID:  req.CommissionID,
		ConclaveID:    conclaveID,
		Title:         req.Title,
		Description:   req.Description,
		Status:        "open",
		ContainerID:   req.ContainerID,
		ContainerType: req.ContainerType,
	}

	if err := s.tomeRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create tome: %w", err)
	}

	// Fetch created tome
	created, err := s.tomeRepo.GetByID(ctx, nextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created tome: %w", err)
	}

	return &primary.CreateTomeResponse{
		TomeID: created.ID,
		Tome:   s.recordToTome(created),
	}, nil
}

// GetTome retrieves a tome by ID.
func (s *TomeServiceImpl) GetTome(ctx context.Context, tomeID string) (*primary.Tome, error) {
	record, err := s.tomeRepo.GetByID(ctx, tomeID)
	if err != nil {
		return nil, err
	}
	return s.recordToTome(record), nil
}

// ListTomes lists tomes with optional filters.
func (s *TomeServiceImpl) ListTomes(ctx context.Context, filters primary.TomeFilters) ([]*primary.Tome, error) {
	records, err := s.tomeRepo.List(ctx, secondary.TomeFilters{
		CommissionID: filters.CommissionID,
		ConclaveID:   filters.ConclaveID,
		Status:       filters.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list tomes: %w", err)
	}

	tomes := make([]*primary.Tome, len(records))
	for i, r := range records {
		tomes[i] = s.recordToTome(r)
	}
	return tomes, nil
}

// CloseTome marks a tome as closed.
func (s *TomeServiceImpl) CloseTome(ctx context.Context, tomeID string) error {
	record, err := s.tomeRepo.GetByID(ctx, tomeID)
	if err != nil {
		return err
	}

	// Guard: cannot close pinned tome
	if record.Pinned {
		return fmt.Errorf("cannot close pinned tome %s. Unpin first with: orc tome unpin %s", tomeID, tomeID)
	}

	return s.tomeRepo.UpdateStatus(ctx, tomeID, "closed", true)
}

// UpdateTome updates a tome's title and/or description.
func (s *TomeServiceImpl) UpdateTome(ctx context.Context, req primary.UpdateTomeRequest) error {
	record := &secondary.TomeRecord{
		ID:          req.TomeID,
		Title:       req.Title,
		Description: req.Description,
	}
	return s.tomeRepo.Update(ctx, record)
}

// PinTome pins a tome.
func (s *TomeServiceImpl) PinTome(ctx context.Context, tomeID string) error {
	return s.tomeRepo.Pin(ctx, tomeID)
}

// UnpinTome unpins a tome.
func (s *TomeServiceImpl) UnpinTome(ctx context.Context, tomeID string) error {
	return s.tomeRepo.Unpin(ctx, tomeID)
}

// DeleteTome deletes a tome.
func (s *TomeServiceImpl) DeleteTome(ctx context.Context, tomeID string) error {
	return s.tomeRepo.Delete(ctx, tomeID)
}

// AssignTomeToWorkbench assigns a tome to a workbench.
func (s *TomeServiceImpl) AssignTomeToWorkbench(ctx context.Context, tomeID, workbenchID string) error {
	// Verify tome exists
	_, err := s.tomeRepo.GetByID(ctx, tomeID)
	if err != nil {
		return err
	}

	return s.tomeRepo.AssignWorkbench(ctx, tomeID, workbenchID)
}

// GetTomesByWorkbench retrieves tomes assigned to a workbench.
func (s *TomeServiceImpl) GetTomesByWorkbench(ctx context.Context, workbenchID string) ([]*primary.Tome, error) {
	records, err := s.tomeRepo.GetByWorkbench(ctx, workbenchID)
	if err != nil {
		return nil, err
	}

	tomes := make([]*primary.Tome, len(records))
	for i, r := range records {
		tomes[i] = s.recordToTome(r)
	}
	return tomes, nil
}

// GetTomeNotes retrieves all notes in a tome.
// Delegates to NoteService.GetNotesByContainer.
func (s *TomeServiceImpl) GetTomeNotes(ctx context.Context, tomeID string) ([]*primary.Note, error) {
	return s.noteService.GetNotesByContainer(ctx, "tome", tomeID)
}

// UnparkTome moves a tome from commission root to a specific Conclave.
func (s *TomeServiceImpl) UnparkTome(ctx context.Context, tomeID, conclaveID string) error {
	// Get tome to verify it exists
	_, err := s.tomeRepo.GetByID(ctx, tomeID)
	if err != nil {
		return err
	}

	// Update container
	return s.tomeRepo.UpdateContainer(ctx, tomeID, conclaveID, "conclave")
}

// Helper methods

func (s *TomeServiceImpl) recordToTome(r *secondary.TomeRecord) *primary.Tome {
	return &primary.Tome{
		ID:                  r.ID,
		CommissionID:        r.CommissionID,
		ConclaveID:          r.ConclaveID,
		Title:               r.Title,
		Description:         r.Description,
		Status:              r.Status,
		AssignedWorkbenchID: r.AssignedWorkbenchID,
		Pinned:              r.Pinned,
		ContainerID:         r.ContainerID,
		ContainerType:       r.ContainerType,
		CreatedAt:           r.CreatedAt,
		UpdatedAt:           r.UpdatedAt,
		ClosedAt:            r.ClosedAt,
	}
}

// Ensure TomeServiceImpl implements the interface
var _ primary.TomeService = (*TomeServiceImpl)(nil)
