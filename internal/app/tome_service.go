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
	// Validate mission exists
	exists, err := s.tomeRepo.MissionExists(ctx, req.MissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate mission: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("mission %s not found", req.MissionID)
	}

	// Get next ID
	nextID, err := s.tomeRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tome ID: %w", err)
	}

	// Create record
	record := &secondary.TomeRecord{
		ID:          nextID,
		MissionID:   req.MissionID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "active",
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
		MissionID: filters.MissionID,
		Status:    filters.Status,
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

// CompleteTome marks a tome as complete.
func (s *TomeServiceImpl) CompleteTome(ctx context.Context, tomeID string) error {
	record, err := s.tomeRepo.GetByID(ctx, tomeID)
	if err != nil {
		return err
	}

	// Guard: cannot complete pinned tome
	if record.Pinned {
		return fmt.Errorf("cannot complete pinned tome %s. Unpin first with: orc tome unpin %s", tomeID, tomeID)
	}

	return s.tomeRepo.UpdateStatus(ctx, tomeID, "complete", true)
}

// PauseTome pauses an active tome.
func (s *TomeServiceImpl) PauseTome(ctx context.Context, tomeID string) error {
	record, err := s.tomeRepo.GetByID(ctx, tomeID)
	if err != nil {
		return err
	}

	// Guard: can only pause active tomes
	if record.Status != "active" {
		return fmt.Errorf("can only pause active tomes (current status: %s)", record.Status)
	}

	return s.tomeRepo.UpdateStatus(ctx, tomeID, "paused", false)
}

// ResumeTome resumes a paused tome.
func (s *TomeServiceImpl) ResumeTome(ctx context.Context, tomeID string) error {
	record, err := s.tomeRepo.GetByID(ctx, tomeID)
	if err != nil {
		return err
	}

	// Guard: can only resume paused tomes
	if record.Status != "paused" {
		return fmt.Errorf("can only resume paused tomes (current status: %s)", record.Status)
	}

	return s.tomeRepo.UpdateStatus(ctx, tomeID, "active", false)
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

// AssignTomeToGrove assigns a tome to a grove.
func (s *TomeServiceImpl) AssignTomeToGrove(ctx context.Context, tomeID, groveID string) error {
	// Verify tome exists
	_, err := s.tomeRepo.GetByID(ctx, tomeID)
	if err != nil {
		return err
	}

	return s.tomeRepo.AssignGrove(ctx, tomeID, groveID)
}

// GetTomesByGrove retrieves tomes assigned to a grove.
func (s *TomeServiceImpl) GetTomesByGrove(ctx context.Context, groveID string) ([]*primary.Tome, error) {
	records, err := s.tomeRepo.GetByGrove(ctx, groveID)
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

// Helper methods

func (s *TomeServiceImpl) recordToTome(r *secondary.TomeRecord) *primary.Tome {
	return &primary.Tome{
		ID:              r.ID,
		MissionID:       r.MissionID,
		Title:           r.Title,
		Description:     r.Description,
		Status:          r.Status,
		AssignedGroveID: r.AssignedGroveID,
		Pinned:          r.Pinned,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
		CompletedAt:     r.CompletedAt,
	}
}

// Ensure TomeServiceImpl implements the interface
var _ primary.TomeService = (*TomeServiceImpl)(nil)
