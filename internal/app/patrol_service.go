package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// PatrolServiceImpl implements the PatrolService interface.
type PatrolServiceImpl struct {
	patrolRepo    secondary.PatrolRepository
	kennelRepo    secondary.KennelRepository
	workbenchRepo secondary.WorkbenchRepository
}

// NewPatrolService creates a new PatrolService with injected dependencies.
func NewPatrolService(patrolRepo secondary.PatrolRepository, kennelRepo secondary.KennelRepository, workbenchRepo secondary.WorkbenchRepository) *PatrolServiceImpl {
	return &PatrolServiceImpl{
		patrolRepo:    patrolRepo,
		kennelRepo:    kennelRepo,
		workbenchRepo: workbenchRepo,
	}
}

// StartPatrol starts a new patrol for a workbench.
func (s *PatrolServiceImpl) StartPatrol(ctx context.Context, workbenchID string) (*primary.Patrol, error) {
	// Check workbench exists
	workbench, err := s.workbenchRepo.GetByID(ctx, workbenchID)
	if err != nil {
		return nil, fmt.Errorf("workbench not found: %w", err)
	}

	// Get kennel for workbench
	kennel, err := s.kennelRepo.GetByWorkbench(ctx, workbenchID)
	if err != nil {
		return nil, fmt.Errorf("no kennel for workbench: %w", err)
	}

	// Check no active patrol already exists
	active, err := s.patrolRepo.GetActiveByKennel(ctx, kennel.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check active patrol: %w", err)
	}
	if active != nil {
		return nil, fmt.Errorf("patrol %s already active for kennel %s", active.ID, kennel.ID)
	}

	// Generate next ID
	id, err := s.patrolRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate patrol ID: %w", err)
	}

	// Derive TMux target from workbench name
	// Convention: orc:<workbench-name>.0
	target := fmt.Sprintf("orc:%s.0", workbench.Name)

	// Create patrol record
	record := &secondary.PatrolRecord{
		ID:       id,
		KennelID: kennel.ID,
		Target:   target,
		Status:   primary.PatrolStatusActive,
	}
	if err := s.patrolRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create patrol: %w", err)
	}

	// Fetch the created record to get timestamps
	created, err := s.patrolRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created patrol: %w", err)
	}

	return s.recordToPatrol(created), nil
}

// EndPatrol ends an active patrol.
func (s *PatrolServiceImpl) EndPatrol(ctx context.Context, patrolID string) error {
	// Get patrol to verify it exists and is active
	patrol, err := s.patrolRepo.GetByID(ctx, patrolID)
	if err != nil {
		return fmt.Errorf("patrol not found: %w", err)
	}

	if patrol.Status != primary.PatrolStatusActive {
		return fmt.Errorf("patrol %s is not active (status: %s)", patrolID, patrol.Status)
	}

	// Update status to completed
	return s.patrolRepo.UpdateStatus(ctx, patrolID, primary.PatrolStatusCompleted)
}

// GetPatrol retrieves a patrol by ID.
func (s *PatrolServiceImpl) GetPatrol(ctx context.Context, patrolID string) (*primary.Patrol, error) {
	record, err := s.patrolRepo.GetByID(ctx, patrolID)
	if err != nil {
		return nil, err
	}
	return s.recordToPatrol(record), nil
}

// GetActivePatrolForKennel retrieves the active patrol for a kennel (if any).
func (s *PatrolServiceImpl) GetActivePatrolForKennel(ctx context.Context, kennelID string) (*primary.Patrol, error) {
	record, err := s.patrolRepo.GetActiveByKennel(ctx, kennelID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}
	return s.recordToPatrol(record), nil
}

// ListPatrols lists patrols with optional filters.
func (s *PatrolServiceImpl) ListPatrols(ctx context.Context, filters primary.PatrolFilters) ([]*primary.Patrol, error) {
	records, err := s.patrolRepo.List(ctx, secondary.PatrolFilters{
		KennelID: filters.KennelID,
		Status:   filters.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list patrols: %w", err)
	}

	patrols := make([]*primary.Patrol, len(records))
	for i, r := range records {
		patrols[i] = s.recordToPatrol(r)
	}
	return patrols, nil
}

// Helper methods

func (s *PatrolServiceImpl) recordToPatrol(r *secondary.PatrolRecord) *primary.Patrol {
	return &primary.Patrol{
		ID:        r.ID,
		KennelID:  r.KennelID,
		Target:    r.Target,
		Status:    r.Status,
		Config:    r.Config,
		StartedAt: r.StartedAt,
		EndedAt:   r.EndedAt,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

// Ensure PatrolServiceImpl implements the interface
var _ primary.PatrolService = (*PatrolServiceImpl)(nil)
