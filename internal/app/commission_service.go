package app

import (
	"context"
	"fmt"
	"time"

	corecommission "github.com/example/orc/internal/core/commission"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// CommissionServiceImpl implements the CommissionService interface.
type CommissionServiceImpl struct {
	commissionRepo secondary.CommissionRepository
	agentProvider  secondary.AgentIdentityProvider
	executor       EffectExecutor
}

// NewCommissionService creates a new CommissionService with injected dependencies.
func NewCommissionService(
	commissionRepo secondary.CommissionRepository,
	agentProvider secondary.AgentIdentityProvider,
	executor EffectExecutor,
) *CommissionServiceImpl {
	return &CommissionServiceImpl{
		commissionRepo: commissionRepo,
		agentProvider:  agentProvider,
		executor:       executor,
	}
}

// CreateCommission creates a new commission.
func (s *CommissionServiceImpl) CreateCommission(ctx context.Context, req primary.CreateCommissionRequest) (*primary.CreateCommissionResponse, error) {
	// 1. Get agent identity for guard
	identity, err := s.agentProvider.GetCurrentIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent identity: %w", err)
	}

	// 2. Check guard
	guardCtx := corecommission.GuardContext{
		AgentType: corecommission.AgentType(identity.Type),
		AgentID:   identity.FullID,
		// CommissionID resolved via DB when needed, not from identity
	}
	if result := corecommission.CanCreateCommission(guardCtx); !result.Allowed {
		return nil, result.Error()
	}

	// 3. Generate ID using core business rule
	nextID, err := s.commissionRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate commission ID: %w", err)
	}

	// 4. Create commission record with pre-populated ID and initial status from core
	record := &secondary.CommissionRecord{
		ID:          nextID,
		Title:       req.Title,
		Description: req.Description,
		Status:      string(corecommission.InitialStatus()),
	}

	if err := s.commissionRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create commission: %w", err)
	}

	// 5. Return response
	return &primary.CreateCommissionResponse{
		CommissionID: record.ID,
		Commission:   s.recordToCommission(record),
	}, nil
}

// StartCommission starts a commission.
func (s *CommissionServiceImpl) StartCommission(ctx context.Context, req primary.StartCommissionRequest) (*primary.StartCommissionResponse, error) {
	// 1. Guard check
	identity, err := s.agentProvider.GetCurrentIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent identity: %w", err)
	}

	guardCtx := corecommission.GuardContext{
		AgentType: corecommission.AgentType(identity.Type),
		AgentID:   identity.FullID,
		// CommissionID resolved via DB when needed, not from identity
	}
	if result := corecommission.CanStartCommission(guardCtx); !result.Allowed {
		return nil, result.Error()
	}

	// 2. Fetch commission
	commission, err := s.commissionRepo.GetByID(ctx, req.CommissionID)
	if err != nil {
		return nil, fmt.Errorf("commission not found: %w", err)
	}

	return &primary.StartCommissionResponse{
		Commission: s.recordToCommission(commission),
	}, nil
}

// LaunchCommission creates and starts commission infrastructure.
func (s *CommissionServiceImpl) LaunchCommission(ctx context.Context, req primary.LaunchCommissionRequest) (*primary.LaunchCommissionResponse, error) {
	// 1. Guard check
	identity, err := s.agentProvider.GetCurrentIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent identity: %w", err)
	}

	guardCtx := corecommission.GuardContext{
		AgentType: corecommission.AgentType(identity.Type),
		AgentID:   identity.FullID,
		// CommissionID resolved via DB when needed, not from identity
	}
	if result := corecommission.CanLaunchCommission(guardCtx); !result.Allowed {
		return nil, result.Error()
	}

	// 2. Create commission first
	createResp, err := s.CreateCommission(ctx, primary.CreateCommissionRequest(req))
	if err != nil {
		return nil, err
	}

	// 3. Start commission
	_, err = s.StartCommission(ctx, primary.StartCommissionRequest{
		CommissionID: createResp.CommissionID,
	})
	if err != nil {
		return nil, err
	}

	return &primary.LaunchCommissionResponse{
		CommissionID: createResp.CommissionID,
		Commission:   createResp.Commission,
	}, nil
}

// GetCommission retrieves a commission by ID.
func (s *CommissionServiceImpl) GetCommission(ctx context.Context, commissionID string) (*primary.Commission, error) {
	record, err := s.commissionRepo.GetByID(ctx, commissionID)
	if err != nil {
		return nil, fmt.Errorf("commission not found: %w", err)
	}
	return s.recordToCommission(record), nil
}

// ListCommissions lists commissions with optional filters.
func (s *CommissionServiceImpl) ListCommissions(ctx context.Context, filters primary.CommissionFilters) ([]*primary.Commission, error) {
	records, err := s.commissionRepo.List(ctx, secondary.CommissionFilters{
		Status: filters.Status,
		Limit:  filters.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list commissions: %w", err)
	}

	commissions := make([]*primary.Commission, len(records))
	for i, r := range records {
		commissions[i] = s.recordToCommission(r)
	}
	return commissions, nil
}

// CompleteCommission marks a commission as complete.
func (s *CommissionServiceImpl) CompleteCommission(ctx context.Context, commissionID string) error {
	// 1. Fetch commission to check state
	record, err := s.commissionRepo.GetByID(ctx, commissionID)
	if err != nil {
		return fmt.Errorf("commission not found: %w", err)
	}

	// 2. Guard check
	stateCtx := corecommission.CommissionStateContext{
		CommissionID: commissionID,
		IsPinned:     record.Pinned,
	}
	if result := corecommission.CanCompleteCommission(stateCtx); !result.Allowed {
		return result.Error()
	}

	// 3. Apply status transition using core business logic
	transition := corecommission.ApplyStatusTransition(corecommission.StatusComplete, time.Now())
	record.Status = string(transition.NewStatus)
	if transition.CompletedAt != nil {
		record.CompletedAt = transition.CompletedAt.Format(time.RFC3339)
	}

	return s.commissionRepo.Update(ctx, record)
}

// ArchiveCommission archives a completed commission.
func (s *CommissionServiceImpl) ArchiveCommission(ctx context.Context, commissionID string) error {
	// 1. Fetch commission to check state
	record, err := s.commissionRepo.GetByID(ctx, commissionID)
	if err != nil {
		return fmt.Errorf("commission not found: %w", err)
	}

	// 2. Guard check
	stateCtx := corecommission.CommissionStateContext{
		CommissionID: commissionID,
		IsPinned:     record.Pinned,
	}
	if result := corecommission.CanArchiveCommission(stateCtx); !result.Allowed {
		return result.Error()
	}

	// 3. Apply status transition using core business logic
	transition := corecommission.ApplyStatusTransition(corecommission.StatusArchived, time.Now())
	record.Status = string(transition.NewStatus)

	return s.commissionRepo.Update(ctx, record)
}

// UpdateCommission updates commission title and/or description.
func (s *CommissionServiceImpl) UpdateCommission(ctx context.Context, req primary.UpdateCommissionRequest) error {
	record, err := s.commissionRepo.GetByID(ctx, req.CommissionID)
	if err != nil {
		return fmt.Errorf("commission not found: %w", err)
	}

	if req.Title != "" {
		record.Title = req.Title
	}
	if req.Description != "" {
		record.Description = req.Description
	}

	return s.commissionRepo.Update(ctx, record)
}

// DeleteCommission deletes a commission.
func (s *CommissionServiceImpl) DeleteCommission(ctx context.Context, req primary.DeleteCommissionRequest) error {
	// 1. Count dependents
	shipmentCount, err := s.commissionRepo.CountShipments(ctx, req.CommissionID)
	if err != nil {
		return fmt.Errorf("failed to count shipments: %w", err)
	}

	// 2. Guard check
	deleteCtx := corecommission.DeleteContext{
		CommissionID:   req.CommissionID,
		ShipmentCount:  shipmentCount,
		WorkbenchCount: 0, // Workbenches tracked separately
		ForceDelete:    req.Force,
	}
	if result := corecommission.CanDeleteCommission(deleteCtx); !result.Allowed {
		return result.Error()
	}

	// 3. Delete
	return s.commissionRepo.Delete(ctx, req.CommissionID)
}

// PinCommission pins a commission.
func (s *CommissionServiceImpl) PinCommission(ctx context.Context, commissionID string) error {
	// 1. Check if commission exists
	record, err := s.commissionRepo.GetByID(ctx, commissionID)
	commissionExists := err == nil

	// 2. Guard check
	pinCtx := corecommission.PinContext{
		CommissionID:     commissionID,
		CommissionExists: commissionExists,
		IsPinned:         commissionExists && record.Pinned,
	}
	if result := corecommission.CanPinCommission(pinCtx); !result.Allowed {
		return result.Error()
	}

	// 3. Pin the commission
	return s.commissionRepo.Pin(ctx, commissionID)
}

// UnpinCommission unpins a commission.
func (s *CommissionServiceImpl) UnpinCommission(ctx context.Context, commissionID string) error {
	// 1. Check if commission exists
	record, err := s.commissionRepo.GetByID(ctx, commissionID)
	commissionExists := err == nil

	// 2. Guard check
	pinCtx := corecommission.PinContext{
		CommissionID:     commissionID,
		CommissionExists: commissionExists,
		IsPinned:         commissionExists && record.Pinned,
	}
	if result := corecommission.CanUnpinCommission(pinCtx); !result.Allowed {
		return result.Error()
	}

	// 3. Unpin the commission
	return s.commissionRepo.Unpin(ctx, commissionID)
}

// Helper methods

func (s *CommissionServiceImpl) recordToCommission(r *secondary.CommissionRecord) *primary.Commission {
	return &primary.Commission{
		ID:          r.ID,
		Title:       r.Title,
		Description: r.Description,
		Status:      r.Status,
		CreatedAt:   r.CreatedAt,
		StartedAt:   r.StartedAt,
		CompletedAt: r.CompletedAt,
	}
}

// Ensure CommissionServiceImpl implements the interface
var _ primary.CommissionService = (*CommissionServiceImpl)(nil)
