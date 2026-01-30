package app

import (
	"context"
	"fmt"

	plancore "github.com/example/orc/internal/core/plan"
	"github.com/example/orc/internal/models"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// PlanServiceImpl implements the PlanService interface.
type PlanServiceImpl struct {
	planRepo       secondary.PlanRepository
	approvalRepo   secondary.ApprovalRepository
	escalationRepo secondary.EscalationRepository
	workbenchRepo  secondary.WorkbenchRepository
	gatehouseRepo  secondary.GatehouseRepository
	messageService primary.MessageService
	tmuxAdapter    secondary.TMuxAdapter
}

// NewPlanService creates a new PlanService with injected dependencies.
func NewPlanService(
	planRepo secondary.PlanRepository,
	approvalRepo secondary.ApprovalRepository,
	escalationRepo secondary.EscalationRepository,
	workbenchRepo secondary.WorkbenchRepository,
	gatehouseRepo secondary.GatehouseRepository,
	messageService primary.MessageService,
	tmuxAdapter secondary.TMuxAdapter,
) *PlanServiceImpl {
	return &PlanServiceImpl{
		planRepo:       planRepo,
		approvalRepo:   approvalRepo,
		escalationRepo: escalationRepo,
		workbenchRepo:  workbenchRepo,
		gatehouseRepo:  gatehouseRepo,
		messageService: messageService,
		tmuxAdapter:    tmuxAdapter,
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

	// Validate task exists
	taskExists, err := s.planRepo.TaskExists(ctx, req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate task: %w", err)
	}
	if !taskExists {
		return nil, fmt.Errorf("task %s not found", req.TaskID)
	}

	// Check if task already has an active plan
	hasActive, err := s.planRepo.HasActivePlanForTask(ctx, req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to check active plan: %w", err)
	}
	if hasActive {
		return nil, fmt.Errorf("task %s already has an active plan", req.TaskID)
	}

	// Get next ID
	nextID, err := s.planRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plan ID: %w", err)
	}

	// Create record
	record := &secondary.PlanRecord{
		ID:               nextID,
		CommissionID:     req.CommissionID,
		TaskID:           req.TaskID,
		Title:            req.Title,
		Description:      req.Description,
		Content:          req.Content,
		Status:           "draft",
		SupersedesPlanID: req.SupersedesPlanID,
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
		TaskID:       filters.TaskID,
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

// SubmitPlan submits a plan for review (draft → pending_review).
func (s *PlanServiceImpl) SubmitPlan(ctx context.Context, planID string) error {
	plan, err := s.planRepo.GetByID(ctx, planID)
	if err != nil {
		return err
	}

	guardResult := plancore.CanSubmitPlan(plancore.SubmitPlanContext{
		PlanID:     planID,
		Status:     plan.Status,
		HasContent: plan.Content != "",
	})
	if err := guardResult.Error(); err != nil {
		return err
	}

	return s.planRepo.UpdateStatus(ctx, planID, models.PlanStatusPendingReview)
}

// ApprovePlan approves a plan (pending_review → approved), creating an approval record.
func (s *PlanServiceImpl) ApprovePlan(ctx context.Context, planID string) (*primary.Approval, error) {
	plan, err := s.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, err
	}

	guardResult := plancore.CanApprovePlan(plancore.ApprovePlanContext{
		PlanID:   planID,
		Status:   plan.Status,
		IsPinned: plan.Pinned,
	})
	if err := guardResult.Error(); err != nil {
		return nil, err
	}

	// Create approval record
	approvalID, err := s.approvalRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate approval ID: %w", err)
	}

	approvalRecord := &secondary.ApprovalRecord{
		ID:        approvalID,
		PlanID:    planID,
		TaskID:    plan.TaskID,
		Mechanism: primary.ApprovalMechanismManual,
		Outcome:   primary.ApprovalOutcomeApproved,
	}
	if err := s.approvalRepo.Create(ctx, approvalRecord); err != nil {
		return nil, fmt.Errorf("failed to create approval: %w", err)
	}

	// Update plan status
	if err := s.planRepo.Approve(ctx, planID); err != nil {
		return nil, fmt.Errorf("failed to approve plan: %w", err)
	}

	return &primary.Approval{
		ID:        approvalID,
		PlanID:    planID,
		TaskID:    plan.TaskID,
		Mechanism: primary.ApprovalMechanismManual,
		Outcome:   primary.ApprovalOutcomeApproved,
	}, nil
}

// EscalatePlan escalates a plan for human review, creating approval and escalation records.
func (s *PlanServiceImpl) EscalatePlan(ctx context.Context, req primary.EscalatePlanRequest) (*primary.EscalatePlanResponse, error) {
	plan, err := s.planRepo.GetByID(ctx, req.PlanID)
	if err != nil {
		return nil, err
	}

	guardResult := plancore.CanEscalatePlan(plancore.EscalatePlanContext{
		PlanID:     req.PlanID,
		Status:     plan.Status,
		HasContent: plan.Content != "",
		HasReason:  req.Reason != "",
	})
	if err := guardResult.Error(); err != nil {
		return nil, err
	}

	// Get workbench to find workshop
	workbench, err := s.workbenchRepo.GetByID(ctx, req.OriginActorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workbench: %w", err)
	}

	// Get gatehouse for the workshop
	gatehouse, err := s.gatehouseRepo.GetByWorkshop(ctx, workbench.WorkshopID)
	if err != nil {
		return nil, fmt.Errorf("failed to get gatehouse for workshop: %w", err)
	}

	// Create approval record (outcome=escalated)
	approvalID, err := s.approvalRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate approval ID: %w", err)
	}

	approvalRecord := &secondary.ApprovalRecord{
		ID:        approvalID,
		PlanID:    req.PlanID,
		TaskID:    plan.TaskID,
		Mechanism: primary.ApprovalMechanismManual,
		Outcome:   primary.ApprovalOutcomeEscalated,
	}
	if err := s.approvalRepo.Create(ctx, approvalRecord); err != nil {
		return nil, fmt.Errorf("failed to create approval: %w", err)
	}

	// Create escalation record
	escalationID, err := s.escalationRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate escalation ID: %w", err)
	}

	escalationRecord := &secondary.EscalationRecord{
		ID:            escalationID,
		ApprovalID:    approvalID,
		PlanID:        req.PlanID,
		TaskID:        plan.TaskID,
		Reason:        req.Reason,
		Status:        primary.EscalationStatusPending,
		RoutingRule:   "workshop_gatehouse",
		OriginActorID: req.OriginActorID,
		TargetActorID: gatehouse.ID,
	}
	if err := s.escalationRepo.Create(ctx, escalationRecord); err != nil {
		return nil, fmt.Errorf("failed to create escalation: %w", err)
	}

	// Update plan status to escalated
	if err := s.planRepo.UpdateStatus(ctx, req.PlanID, models.PlanStatusEscalated); err != nil {
		return nil, fmt.Errorf("failed to update plan status: %w", err)
	}

	// Send mail to gatehouse
	mailBody := fmt.Sprintf("Plan %s has been escalated.\n\nReason: %s\n\nTask: %s\nFrom: %s\n\nUse 'orc plan show %s' to review.",
		req.PlanID, req.Reason, plan.TaskID, req.OriginActorID, req.PlanID)
	_, err = s.messageService.CreateMessage(ctx, primary.CreateMessageRequest{
		Sender:    req.OriginActorID,
		Recipient: gatehouse.ID,
		Subject:   fmt.Sprintf("Escalation: %s", req.PlanID),
		Body:      mailBody,
	})
	if err != nil {
		// Log but don't fail - escalation is already created
		fmt.Printf("Warning: failed to send escalation mail: %v\n", err)
	}

	// Nudge gatehouse (best effort - don't fail if tmux not running)
	nudgeMsg := fmt.Sprintf("New escalation: %s - %s", escalationID, req.Reason)
	sessionName := s.tmuxAdapter.FindSessionByWorkshopID(ctx, workbench.WorkshopID)
	if sessionName != "" {
		target := fmt.Sprintf("%s:1.1", sessionName) // Gatehouse is window 1, pane 1
		_ = s.tmuxAdapter.NudgeSession(ctx, target, nudgeMsg)
	}

	return &primary.EscalatePlanResponse{
		ApprovalID:   approvalID,
		EscalationID: escalationID,
		TargetActor:  gatehouse.ID,
	}, nil
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

// GetTaskActivePlan retrieves the active (draft) plan for a task.
func (s *PlanServiceImpl) GetTaskActivePlan(ctx context.Context, taskID string) (*primary.Plan, error) {
	record, err := s.planRepo.GetActivePlanForTask(ctx, taskID)
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
		TaskID:           r.TaskID,
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
		SupersedesPlanID: r.SupersedesPlanID,
	}
}

// Ensure PlanServiceImpl implements the interface
var _ primary.PlanService = (*PlanServiceImpl)(nil)
