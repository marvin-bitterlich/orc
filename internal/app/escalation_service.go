package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// EscalationServiceImpl implements the EscalationService interface.
type EscalationServiceImpl struct {
	escalationRepo secondary.EscalationRepository
}

// NewEscalationService creates a new EscalationService with injected dependencies.
func NewEscalationService(escalationRepo secondary.EscalationRepository) *EscalationServiceImpl {
	return &EscalationServiceImpl{
		escalationRepo: escalationRepo,
	}
}

// GetEscalation retrieves an escalation by ID.
func (s *EscalationServiceImpl) GetEscalation(ctx context.Context, escalationID string) (*primary.Escalation, error) {
	record, err := s.escalationRepo.GetByID(ctx, escalationID)
	if err != nil {
		return nil, err
	}
	return s.recordToEscalation(record), nil
}

// ListEscalations lists escalations with optional filters.
func (s *EscalationServiceImpl) ListEscalations(ctx context.Context, filters primary.EscalationFilters) ([]*primary.Escalation, error) {
	records, err := s.escalationRepo.List(ctx, secondary.EscalationFilters{
		TaskID:        filters.TaskID,
		Status:        filters.Status,
		TargetActorID: filters.TargetActorID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list escalations: %w", err)
	}

	escalations := make([]*primary.Escalation, len(records))
	for i, r := range records {
		escalations[i] = s.recordToEscalation(r)
	}
	return escalations, nil
}

// ResolveEscalation resolves an escalation with an outcome.
func (s *EscalationServiceImpl) ResolveEscalation(ctx context.Context, req primary.ResolveEscalationRequest) error {
	// Validate escalation exists and is pending
	escalation, err := s.escalationRepo.GetByID(ctx, req.EscalationID)
	if err != nil {
		return fmt.Errorf("escalation not found: %w", err)
	}

	if escalation.Status != primary.EscalationStatusPending {
		return fmt.Errorf("escalation %s is not pending (current status: %s)", req.EscalationID, escalation.Status)
	}

	// Validate outcome
	if req.Outcome != "approved" && req.Outcome != "rejected" {
		return fmt.Errorf("invalid outcome: %s (must be 'approved' or 'rejected')", req.Outcome)
	}

	// Determine final status based on outcome
	status := primary.EscalationStatusResolved
	if req.Outcome == "rejected" {
		status = primary.EscalationStatusDismissed
	}

	// Resolve in repository
	if err := s.escalationRepo.Resolve(ctx, req.EscalationID, req.Resolution, req.ResolvedBy); err != nil {
		return fmt.Errorf("failed to resolve escalation: %w", err)
	}

	// Update status
	if err := s.escalationRepo.UpdateStatus(ctx, req.EscalationID, status, true); err != nil {
		return fmt.Errorf("failed to update escalation status: %w", err)
	}

	return nil
}

// Helper methods

func (s *EscalationServiceImpl) recordToEscalation(r *secondary.EscalationRecord) *primary.Escalation {
	return &primary.Escalation{
		ID:            r.ID,
		ApprovalID:    r.ApprovalID,
		PlanID:        r.PlanID,
		TaskID:        r.TaskID,
		Reason:        r.Reason,
		Status:        r.Status,
		RoutingRule:   r.RoutingRule,
		OriginActorID: r.OriginActorID,
		TargetActorID: r.TargetActorID,
		Resolution:    r.Resolution,
		ResolvedBy:    r.ResolvedBy,
		CreatedAt:     r.CreatedAt,
		ResolvedAt:    r.ResolvedAt,
	}
}

// Ensure EscalationServiceImpl implements the interface
var _ primary.EscalationService = (*EscalationServiceImpl)(nil)
