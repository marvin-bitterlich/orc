package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/core/pr"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// PRServiceImpl implements the PRService interface.
type PRServiceImpl struct {
	prRepo          secondary.PRRepository
	shipmentService primary.ShipmentService
}

// NewPRService creates a new PRService with injected dependencies.
func NewPRService(prRepo secondary.PRRepository, shipmentService primary.ShipmentService) *PRServiceImpl {
	return &PRServiceImpl{
		prRepo:          prRepo,
		shipmentService: shipmentService,
	}
}

// CreatePR creates a new pull request.
func (s *PRServiceImpl) CreatePR(ctx context.Context, req primary.CreatePRRequest) (*primary.CreatePRResponse, error) {
	// Gather context for guards
	shipmentExists, err := s.prRepo.ShipmentExists(ctx, req.ShipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to check shipment: %w", err)
	}

	var shipmentStatus string
	if shipmentExists {
		shipmentStatus, err = s.prRepo.GetShipmentStatus(ctx, req.ShipmentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get shipment status: %w", err)
		}
	}

	shipmentHasPR, err := s.prRepo.ShipmentHasPR(ctx, req.ShipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to check shipment PR: %w", err)
	}

	repoExists, err := s.prRepo.RepoExists(ctx, req.RepoID)
	if err != nil {
		return nil, fmt.Errorf("failed to check repo: %w", err)
	}

	// Evaluate guard
	result := pr.CanCreatePR(pr.CreatePRContext{
		ShipmentID:     req.ShipmentID,
		RepoID:         req.RepoID,
		ShipmentExists: shipmentExists,
		ShipmentStatus: shipmentStatus,
		ShipmentHasPR:  shipmentHasPR,
		RepoExists:     repoExists,
	})
	if err := result.Error(); err != nil {
		return nil, err
	}

	// Get next ID
	nextID, err := s.prRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PR ID: %w", err)
	}

	// Get commission ID from shipment
	shipment, err := s.shipmentService.GetShipment(ctx, req.ShipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shipment: %w", err)
	}

	// Determine initial status
	status := "open"
	if req.Draft {
		status = "draft"
	}

	// Build record
	record := &secondary.PRRecord{
		ID:           nextID,
		ShipmentID:   req.ShipmentID,
		RepoID:       req.RepoID,
		CommissionID: shipment.CommissionID,
		Number:       req.Number,
		Title:        req.Title,
		Description:  req.Description,
		Branch:       req.Branch,
		TargetBranch: req.TargetBranch,
		URL:          req.URL,
		Status:       status,
	}

	if err := s.prRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	// Fetch created PR
	created, err := s.prRepo.GetByID(ctx, nextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created PR: %w", err)
	}

	return &primary.CreatePRResponse{
		PRID: created.ID,
		PR:   s.recordToPR(created),
	}, nil
}

// GetPR retrieves a pull request by ID.
func (s *PRServiceImpl) GetPR(ctx context.Context, prID string) (*primary.PR, error) {
	record, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, err
	}
	return s.recordToPR(record), nil
}

// GetPRByShipment retrieves a pull request by shipment ID.
func (s *PRServiceImpl) GetPRByShipment(ctx context.Context, shipmentID string) (*primary.PR, error) {
	record, err := s.prRepo.GetByShipment(ctx, shipmentID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, fmt.Errorf("no PR found for shipment %s", shipmentID)
	}
	return s.recordToPR(record), nil
}

// ListPRs lists pull requests with optional filters.
func (s *PRServiceImpl) ListPRs(ctx context.Context, filters primary.PRFilters) ([]*primary.PR, error) {
	records, err := s.prRepo.List(ctx, secondary.PRFilters{
		ShipmentID:   filters.ShipmentID,
		RepoID:       filters.RepoID,
		CommissionID: filters.CommissionID,
		Status:       filters.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list PRs: %w", err)
	}

	prs := make([]*primary.PR, len(records))
	for i, r := range records {
		prs[i] = s.recordToPR(r)
	}
	return prs, nil
}

// UpdatePR updates a pull request's metadata.
func (s *PRServiceImpl) UpdatePR(ctx context.Context, req primary.UpdatePRRequest) error {
	// Verify PR exists
	_, err := s.prRepo.GetByID(ctx, req.PRID)
	if err != nil {
		return err
	}

	record := &secondary.PRRecord{
		ID:          req.PRID,
		Title:       req.Title,
		Description: req.Description,
		URL:         req.URL,
		Number:      req.Number,
	}
	return s.prRepo.Update(ctx, record)
}

// OpenPR opens a draft PR for review.
func (s *PRServiceImpl) OpenPR(ctx context.Context, prID string) error {
	// Get current PR
	record, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return err
	}

	// Evaluate guard
	result := pr.CanOpenPR(pr.OpenPRContext{
		PRID:   prID,
		Status: record.Status,
	})
	if err := result.Error(); err != nil {
		return err
	}

	return s.prRepo.UpdateStatus(ctx, prID, "open", false, false)
}

// ApprovePR marks a PR as approved.
func (s *PRServiceImpl) ApprovePR(ctx context.Context, prID string) error {
	// Get current PR
	record, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return err
	}

	// Evaluate guard
	result := pr.CanApprovePR(pr.ApprovePRContext{
		PRID:   prID,
		Status: record.Status,
	})
	if err := result.Error(); err != nil {
		return err
	}

	return s.prRepo.UpdateStatus(ctx, prID, "approved", false, false)
}

// MergePR merges a PR and cascades to complete the shipment.
func (s *PRServiceImpl) MergePR(ctx context.Context, prID string) error {
	// Get current PR
	record, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return err
	}

	// Evaluate guard
	result := pr.CanMergePR(pr.MergePRContext{
		PRID:   prID,
		Status: record.Status,
	})
	if err := result.Error(); err != nil {
		return err
	}

	// Update PR status to merged
	if err := s.prRepo.UpdateStatus(ctx, prID, "merged", true, false); err != nil {
		return fmt.Errorf("failed to update PR status: %w", err)
	}

	// Cascade: complete the shipment
	if err := s.shipmentService.CompleteShipment(ctx, record.ShipmentID); err != nil {
		// Log but don't fail if shipment completion fails (e.g., already complete)
		// The PR is still marked as merged
		fmt.Printf("Warning: failed to complete shipment %s: %v\n", record.ShipmentID, err)
	}

	return nil
}

// ClosePR closes a PR without merging.
func (s *PRServiceImpl) ClosePR(ctx context.Context, prID string) error {
	// Get current PR
	record, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return err
	}

	// Evaluate guard
	result := pr.CanClosePR(pr.ClosePRContext{
		PRID:   prID,
		Status: record.Status,
	})
	if err := result.Error(); err != nil {
		return err
	}

	return s.prRepo.UpdateStatus(ctx, prID, "closed", false, true)
}

// LinkPR links an existing external PR to a shipment.
func (s *PRServiceImpl) LinkPR(ctx context.Context, shipmentID, url string, number int) (*primary.PR, error) {
	// Get shipment to derive other fields
	shipment, err := s.shipmentService.GetShipment(ctx, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shipment: %w", err)
	}

	// Check if shipment already has a PR
	hasPR, err := s.prRepo.ShipmentHasPR(ctx, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to check shipment PR: %w", err)
	}
	if hasPR {
		return nil, fmt.Errorf("shipment %s already has a PR", shipmentID)
	}

	// Create a linked PR (use shipment title as PR title)
	resp, err := s.CreatePR(ctx, primary.CreatePRRequest{
		ShipmentID: shipmentID,
		RepoID:     "", // Will need to be provided or inferred
		Title:      shipment.Title,
		Branch:     fmt.Sprintf("shipment/%s", shipmentID),
		URL:        url,
		Number:     number,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to link PR: %w", err)
	}

	return resp.PR, nil
}

// Helper methods

func (s *PRServiceImpl) recordToPR(r *secondary.PRRecord) *primary.PR {
	return &primary.PR{
		ID:           r.ID,
		ShipmentID:   r.ShipmentID,
		RepoID:       r.RepoID,
		CommissionID: r.CommissionID,
		Number:       r.Number,
		Title:        r.Title,
		Description:  r.Description,
		Branch:       r.Branch,
		TargetBranch: r.TargetBranch,
		URL:          r.URL,
		Status:       r.Status,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
		MergedAt:     r.MergedAt,
		ClosedAt:     r.ClosedAt,
	}
}

// Ensure PRServiceImpl implements the interface
var _ primary.PRService = (*PRServiceImpl)(nil)
