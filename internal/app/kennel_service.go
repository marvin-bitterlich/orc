package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// KennelServiceImpl implements the KennelService interface.
type KennelServiceImpl struct {
	kennelRepo    secondary.KennelRepository
	workbenchRepo secondary.WorkbenchRepository
}

// NewKennelService creates a new KennelService with injected dependencies.
func NewKennelService(kennelRepo secondary.KennelRepository, workbenchRepo ...secondary.WorkbenchRepository) *KennelServiceImpl {
	service := &KennelServiceImpl{
		kennelRepo: kennelRepo,
	}
	// Optional workbench repo for EnsureAllWorkbenchesHaveKennels
	if len(workbenchRepo) > 0 {
		service.workbenchRepo = workbenchRepo[0]
	}
	return service
}

// GetKennel retrieves a kennel by ID.
func (s *KennelServiceImpl) GetKennel(ctx context.Context, kennelID string) (*primary.Kennel, error) {
	record, err := s.kennelRepo.GetByID(ctx, kennelID)
	if err != nil {
		return nil, err
	}
	return s.recordToKennel(record), nil
}

// GetKennelByWorkbench retrieves a kennel by workbench ID.
func (s *KennelServiceImpl) GetKennelByWorkbench(ctx context.Context, workbenchID string) (*primary.Kennel, error) {
	record, err := s.kennelRepo.GetByWorkbench(ctx, workbenchID)
	if err != nil {
		return nil, err
	}
	return s.recordToKennel(record), nil
}

// ListKennels lists kennels with optional filters.
func (s *KennelServiceImpl) ListKennels(ctx context.Context, filters primary.KennelFilters) ([]*primary.Kennel, error) {
	records, err := s.kennelRepo.List(ctx, secondary.KennelFilters{
		WorkbenchID: filters.WorkbenchID,
		Status:      filters.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list kennels: %w", err)
	}

	kennels := make([]*primary.Kennel, len(records))
	for i, r := range records {
		kennels[i] = s.recordToKennel(r)
	}
	return kennels, nil
}

// CreateKennel creates a new kennel for a workbench.
// Returns error if workbench already has a kennel.
func (s *KennelServiceImpl) CreateKennel(ctx context.Context, workbenchID string) (*primary.Kennel, error) {
	// Check workbench exists
	exists, err := s.kennelRepo.WorkbenchExists(ctx, workbenchID)
	if err != nil {
		return nil, fmt.Errorf("failed to check workbench exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("workbench %s not found", workbenchID)
	}

	// Check no existing kennel
	hasKennel, err := s.kennelRepo.WorkbenchHasKennel(ctx, workbenchID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing kennel: %w", err)
	}
	if hasKennel {
		return nil, fmt.Errorf("workbench %s already has a kennel", workbenchID)
	}

	// Generate next ID
	id, err := s.kennelRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate kennel ID: %w", err)
	}

	// Create in DB
	record := &secondary.KennelRecord{
		ID:          id,
		WorkbenchID: workbenchID,
		Status:      primary.KennelStatusVacant,
	}
	if err := s.kennelRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create kennel: %w", err)
	}

	return s.recordToKennel(record), nil
}

// UpdateKennelStatus updates the status of a kennel.
func (s *KennelServiceImpl) UpdateKennelStatus(ctx context.Context, kennelID, status string) error {
	// Validate status
	switch status {
	case primary.KennelStatusVacant, primary.KennelStatusOccupied, primary.KennelStatusAway:
		// Valid status
	default:
		return fmt.Errorf("invalid kennel status: %s (valid: vacant, occupied, away)", status)
	}

	return s.kennelRepo.UpdateStatus(ctx, kennelID, status)
}

// EnsureAllWorkbenchesHaveKennels creates kennels for any workbenches missing them.
// Returns a list of newly created kennel IDs.
func (s *KennelServiceImpl) EnsureAllWorkbenchesHaveKennels(ctx context.Context) ([]string, error) {
	if s.workbenchRepo == nil {
		return nil, fmt.Errorf("workbench repository not available")
	}

	workbenches, err := s.workbenchRepo.List(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list workbenches: %w", err)
	}

	var created []string
	for _, wb := range workbenches {
		hasKennel, err := s.kennelRepo.WorkbenchHasKennel(ctx, wb.ID)
		if err != nil {
			return created, fmt.Errorf("failed to check kennel for %s: %w", wb.ID, err)
		}
		if hasKennel {
			continue
		}

		// Create kennel
		kennel, err := s.CreateKennel(ctx, wb.ID)
		if err != nil {
			return created, fmt.Errorf("failed to create kennel for %s: %w", wb.ID, err)
		}
		created = append(created, kennel.ID)
	}

	return created, nil
}

// Helper methods

func (s *KennelServiceImpl) recordToKennel(r *secondary.KennelRecord) *primary.Kennel {
	return &primary.Kennel{
		ID:          r.ID,
		WorkbenchID: r.WorkbenchID,
		Status:      r.Status,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// Ensure KennelServiceImpl implements the interface
var _ primary.KennelService = (*KennelServiceImpl)(nil)
