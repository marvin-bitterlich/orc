package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// PlanServiceImpl implements the PlanService interface.
type PlanServiceImpl struct {
	planRepo     secondary.PlanRepository
	cycleService primary.CycleService
}

// NewPlanService creates a new PlanService with injected dependencies.
func NewPlanService(
	planRepo secondary.PlanRepository,
	cycleService primary.CycleService,
) *PlanServiceImpl {
	return &PlanServiceImpl{
		planRepo:     planRepo,
		cycleService: cycleService,
	}
}

// CreatePlan creates a new plan.
func (s *PlanServiceImpl) CreatePlan(ctx context.Context, req primary.CreatePlanRequest) (*primary.CreatePlanResponse, error) {
	// Validate commission exists
	exists, err := s.planRepo.CommissionExists(ctx, req.CommissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate commission: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("commission %s not found", req.CommissionID)
	}

	// Validate shipment exists if provided
	if req.ShipmentID != "" {
		shipExists, err := s.planRepo.ShipmentExists(ctx, req.ShipmentID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate shipment: %w", err)
		}
		if !shipExists {
			return nil, fmt.Errorf("shipment %s not found", req.ShipmentID)
		}

		// Check if shipment already has an active plan
		hasActive, err := s.planRepo.HasActivePlanForShipment(ctx, req.ShipmentID)
		if err != nil {
			return nil, fmt.Errorf("failed to check active plan: %w", err)
		}
		if hasActive {
			return nil, fmt.Errorf("shipment %s already has an active plan", req.ShipmentID)
		}
	}

	// Validate cycle exists if provided
	if req.CycleID != "" {
		cycleExists, err := s.planRepo.CycleExists(ctx, req.CycleID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate cycle: %w", err)
		}
		if !cycleExists {
			return nil, fmt.Errorf("cycle %s not found", req.CycleID)
		}
	}

	// Get next ID
	nextID, err := s.planRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plan ID: %w", err)
	}

	// Create record
	record := &secondary.PlanRecord{
		ID:           nextID,
		CommissionID: req.CommissionID,
		ShipmentID:   req.ShipmentID,
		CycleID:      req.CycleID,
		Title:        req.Title,
		Description:  req.Description,
		Content:      req.Content,
		Status:       "draft",
	}

	if err := s.planRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create plan: %w", err)
	}

	// Fetch created plan
	created, err := s.planRepo.GetByID(ctx, nextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created plan: %w", err)
	}

	return &primary.CreatePlanResponse{
		PlanID: created.ID,
		Plan:   s.recordToPlan(created),
	}, nil
}

// GetPlan retrieves a plan by ID.
func (s *PlanServiceImpl) GetPlan(ctx context.Context, planID string) (*primary.Plan, error) {
	record, err := s.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, err
	}
	return s.recordToPlan(record), nil
}

// ListPlans lists plans with optional filters.
func (s *PlanServiceImpl) ListPlans(ctx context.Context, filters primary.PlanFilters) ([]*primary.Plan, error) {
	records, err := s.planRepo.List(ctx, secondary.PlanFilters{
		ShipmentID:   filters.ShipmentID,
		CycleID:      filters.CycleID,
		CommissionID: filters.CommissionID,
		Status:       filters.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list plans: %w", err)
	}

	plans := make([]*primary.Plan, len(records))
	for i, r := range records {
		plans[i] = s.recordToPlan(r)
	}
	return plans, nil
}

// ApprovePlan approves a plan (marks it as approved).
// Also cascades: updates parent Cycle status to "implementing" if the plan has a CycleID.
func (s *PlanServiceImpl) ApprovePlan(ctx context.Context, planID string) error {
	// Get the plan to check for CycleID
	plan, err := s.planRepo.GetByID(ctx, planID)
	if err != nil {
		return fmt.Errorf("failed to get plan: %w", err)
	}

	// Approve the plan
	if err := s.planRepo.Approve(ctx, planID); err != nil {
		return err
	}

	// CASCADE: Update parent Cycle status to "implementing"
	if plan.CycleID != "" && s.cycleService != nil {
		if err := s.cycleService.UpdateCycleStatus(ctx, plan.CycleID, "implementing"); err != nil {
			return fmt.Errorf("failed to cascade cycle status update: %w", err)
		}
	}

	return nil
}

// UpdatePlan updates a plan's title, description, and/or content.
func (s *PlanServiceImpl) UpdatePlan(ctx context.Context, req primary.UpdatePlanRequest) error {
	record := &secondary.PlanRecord{
		ID:          req.PlanID,
		Title:       req.Title,
		Description: req.Description,
		Content:     req.Content,
	}
	return s.planRepo.Update(ctx, record)
}

// PinPlan pins a plan.
func (s *PlanServiceImpl) PinPlan(ctx context.Context, planID string) error {
	return s.planRepo.Pin(ctx, planID)
}

// UnpinPlan unpins a plan.
func (s *PlanServiceImpl) UnpinPlan(ctx context.Context, planID string) error {
	return s.planRepo.Unpin(ctx, planID)
}

// DeletePlan deletes a plan.
func (s *PlanServiceImpl) DeletePlan(ctx context.Context, planID string) error {
	return s.planRepo.Delete(ctx, planID)
}

// GetShipmentActivePlan retrieves the active (draft) plan for a shipment.
func (s *PlanServiceImpl) GetShipmentActivePlan(ctx context.Context, shipmentID string) (*primary.Plan, error) {
	record, err := s.planRepo.GetActivePlanForShipment(ctx, shipmentID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil // No active plan is not an error
	}
	return s.recordToPlan(record), nil
}

// Helper methods

func (s *PlanServiceImpl) recordToPlan(r *secondary.PlanRecord) *primary.Plan {
	return &primary.Plan{
		ID:               r.ID,
		ShipmentID:       r.ShipmentID,
		CycleID:          r.CycleID,
		CommissionID:     r.CommissionID,
		Title:            r.Title,
		Description:      r.Description,
		Status:           r.Status,
		Content:          r.Content,
		Pinned:           r.Pinned,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
		ApprovedAt:       r.ApprovedAt,
		ConclaveID:       r.ConclaveID,
		PromotedFromID:   r.PromotedFromID,
		PromotedFromType: r.PromotedFromType,
	}
}

// Ensure PlanServiceImpl implements the interface
var _ primary.PlanService = (*PlanServiceImpl)(nil)
