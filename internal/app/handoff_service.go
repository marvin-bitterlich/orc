package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// HandoffServiceImpl implements the HandoffService interface.
// Handoffs are immutable - no update or delete operations.
type HandoffServiceImpl struct {
	handoffRepo secondary.HandoffRepository
}

// NewHandoffService creates a new HandoffService with injected dependencies.
func NewHandoffService(handoffRepo secondary.HandoffRepository) *HandoffServiceImpl {
	return &HandoffServiceImpl{
		handoffRepo: handoffRepo,
	}
}

// CreateHandoff creates a new handoff note for session continuity.
func (s *HandoffServiceImpl) CreateHandoff(ctx context.Context, req primary.CreateHandoffRequest) (*primary.CreateHandoffResponse, error) {
	// Get next ID
	nextID, err := s.handoffRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate handoff ID: %w", err)
	}

	// Create record
	record := &secondary.HandoffRecord{
		ID:              nextID,
		HandoffNote:     req.HandoffNote,
		ActiveMissionID: req.ActiveMissionID,
		ActiveGroveID:   req.ActiveGroveID,
		TodosSnapshot:   req.TodosSnapshot,
	}

	if err := s.handoffRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create handoff: %w", err)
	}

	// Fetch created handoff
	created, err := s.handoffRepo.GetByID(ctx, nextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created handoff: %w", err)
	}

	return &primary.CreateHandoffResponse{
		HandoffID: created.ID,
		Handoff:   s.recordToHandoff(created),
	}, nil
}

// GetHandoff retrieves a handoff by ID.
func (s *HandoffServiceImpl) GetHandoff(ctx context.Context, handoffID string) (*primary.Handoff, error) {
	record, err := s.handoffRepo.GetByID(ctx, handoffID)
	if err != nil {
		return nil, err
	}
	return s.recordToHandoff(record), nil
}

// GetLatestHandoff retrieves the most recent handoff.
func (s *HandoffServiceImpl) GetLatestHandoff(ctx context.Context) (*primary.Handoff, error) {
	record, err := s.handoffRepo.GetLatest(ctx)
	if err != nil {
		return nil, err
	}
	return s.recordToHandoff(record), nil
}

// GetLatestHandoffForGrove retrieves the most recent handoff for a grove.
func (s *HandoffServiceImpl) GetLatestHandoffForGrove(ctx context.Context, groveID string) (*primary.Handoff, error) {
	record, err := s.handoffRepo.GetLatestForGrove(ctx, groveID)
	if err != nil {
		return nil, err
	}
	return s.recordToHandoff(record), nil
}

// ListHandoffs lists handoffs with optional limit.
func (s *HandoffServiceImpl) ListHandoffs(ctx context.Context, limit int) ([]*primary.Handoff, error) {
	records, err := s.handoffRepo.List(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list handoffs: %w", err)
	}

	handoffs := make([]*primary.Handoff, len(records))
	for i, r := range records {
		handoffs[i] = s.recordToHandoff(r)
	}
	return handoffs, nil
}

// Helper methods

func (s *HandoffServiceImpl) recordToHandoff(r *secondary.HandoffRecord) *primary.Handoff {
	return &primary.Handoff{
		ID:              r.ID,
		CreatedAt:       r.CreatedAt,
		HandoffNote:     r.HandoffNote,
		ActiveMissionID: r.ActiveMissionID,
		ActiveGroveID:   r.ActiveGroveID,
		TodosSnapshot:   r.TodosSnapshot,
	}
}

// Ensure HandoffServiceImpl implements the interface
var _ primary.HandoffService = (*HandoffServiceImpl)(nil)
