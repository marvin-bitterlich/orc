package primary

import "context"

// PlanService defines the primary port for plan operations.
type PlanService interface {
	// CreatePlan creates a new plan.
	CreatePlan(ctx context.Context, req CreatePlanRequest) (*CreatePlanResponse, error)

	// GetPlan retrieves a plan by ID.
	GetPlan(ctx context.Context, planID string) (*Plan, error)

	// ListPlans lists plans with optional filters.
	ListPlans(ctx context.Context, filters PlanFilters) ([]*Plan, error)

	// SubmitPlan submits a plan for review (draft → pending_review).
	SubmitPlan(ctx context.Context, planID string) error

	// ApprovePlan approves a plan (pending_review → approved), creating an approval record.
	ApprovePlan(ctx context.Context, planID string) (*Approval, error)

	// EscalatePlan escalates a plan for human review, creating approval and escalation records.
	EscalatePlan(ctx context.Context, req EscalatePlanRequest) (*EscalatePlanResponse, error)

	// UpdatePlan updates a plan's title, description, and/or content.
	UpdatePlan(ctx context.Context, req UpdatePlanRequest) error

	// PinPlan pins a plan.
	PinPlan(ctx context.Context, planID string) error

	// UnpinPlan unpins a plan.
	UnpinPlan(ctx context.Context, planID string) error

	// DeletePlan deletes a plan.
	DeletePlan(ctx context.Context, planID string) error

	// GetTaskActivePlan retrieves the active (draft) plan for a task.
	GetTaskActivePlan(ctx context.Context, taskID string) (*Plan, error)
}

// CreatePlanRequest contains parameters for creating a plan.
type CreatePlanRequest struct {
	CommissionID     string
	TaskID           string
	Title            string
	Description      string
	Content          string
	SupersedesPlanID string // Optional - ID of plan this supersedes
}

// CreatePlanResponse contains the result of creating a plan.
type CreatePlanResponse struct {
	PlanID string
	Plan   *Plan
}

// UpdatePlanRequest contains parameters for updating a plan.
type UpdatePlanRequest struct {
	PlanID      string
	Title       string
	Description string
	Content     string
}

// Plan represents a plan entity at the port boundary.
type Plan struct {
	ID               string
	TaskID           string
	CommissionID     string
	Title            string
	Description      string
	Status           string
	Content          string
	Pinned           bool
	CreatedAt        string
	UpdatedAt        string
	ApprovedAt       string
	PromotedFromID   string
	PromotedFromType string
	SupersedesPlanID string
}

// PlanFilters contains filter options for listing plans.
type PlanFilters struct {
	TaskID       string
	CommissionID string
	Status       string
}

// EscalatePlanRequest contains parameters for escalating a plan.
type EscalatePlanRequest struct {
	PlanID        string
	Reason        string
	OriginActorID string // BENCH-xxx of the escalating IMP
}

// EscalatePlanResponse contains the result of escalating a plan.
type EscalatePlanResponse struct {
	ApprovalID   string
	EscalationID string
	TargetActor  string // GATE-xxx of the target gatehouse
}
