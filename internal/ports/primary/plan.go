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

	// ApprovePlan approves a plan (marks it as approved).
	ApprovePlan(ctx context.Context, planID string) error

	// UpdatePlan updates a plan's title, description, and/or content.
	UpdatePlan(ctx context.Context, req UpdatePlanRequest) error

	// PinPlan pins a plan.
	PinPlan(ctx context.Context, planID string) error

	// UnpinPlan unpins a plan.
	UnpinPlan(ctx context.Context, planID string) error

	// DeletePlan deletes a plan.
	DeletePlan(ctx context.Context, planID string) error

	// GetShipmentActivePlan retrieves the active (draft) plan for a shipment.
	GetShipmentActivePlan(ctx context.Context, shipmentID string) (*Plan, error)
}

// CreatePlanRequest contains parameters for creating a plan.
type CreatePlanRequest struct {
	CommissionID string
	ShipmentID   string // Optional
	Title        string
	Description  string
	Content      string
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
	ShipmentID       string
	CommissionID     string
	Title            string
	Description      string
	Status           string
	Content          string
	Pinned           bool
	CreatedAt        string
	UpdatedAt        string
	ApprovedAt       string
	ConclaveID       string
	PromotedFromID   string
	PromotedFromType string
}

// PlanFilters contains filter options for listing plans.
type PlanFilters struct {
	ShipmentID   string
	CommissionID string
	Status       string
}
