// Package plan contains the pure business logic for plan operations.
// Guards are pure functions that evaluate preconditions without side effects.
package plan

import "fmt"

// GuardResult represents the outcome of a guard evaluation.
type GuardResult struct {
	Allowed bool
	Reason  string
}

// Error converts the guard result to an error if not allowed.
func (r GuardResult) Error() error {
	if r.Allowed {
		return nil
	}
	return fmt.Errorf("%s", r.Reason)
}

// CreatePlanContext provides context for plan creation guards.
type CreatePlanContext struct {
	CommissionID          string
	CommissionExists      bool
	ShipmentID            string // Optional - empty string means no shipment
	ShipmentExists        bool   // Only checked if ShipmentID != ""
	ShipmentHasActivePlan bool   // Only checked if ShipmentID != ""
}

// ApprovePlanContext provides context for plan approval guards.
type ApprovePlanContext struct {
	PlanID   string
	Status   string // "draft", "pending_review", "approved"
	IsPinned bool
}

// SubmitPlanContext provides context for plan submission guards.
type SubmitPlanContext struct {
	PlanID     string
	Status     string
	HasContent bool
}

// DeletePlanContext provides context for plan deletion guards.
type DeletePlanContext struct {
	PlanID   string
	IsPinned bool
}

// EscalatePlanContext provides context for plan escalation guards.
type EscalatePlanContext struct {
	PlanID     string
	Status     string
	HasContent bool
	HasReason  bool
}

// CanCreatePlan evaluates whether a plan can be created.
// Rules:
// - Commission must exist
// - Shipment must exist if provided
// - No active (draft) plan for shipment if provided
func CanCreatePlan(ctx CreatePlanContext) GuardResult {
	// Check commission exists
	if !ctx.CommissionExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("commission %s not found", ctx.CommissionID),
		}
	}

	// Check shipment exists if provided
	if ctx.ShipmentID != "" && !ctx.ShipmentExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("shipment %s not found", ctx.ShipmentID),
		}
	}

	// Check no active plan for shipment if provided
	if ctx.ShipmentID != "" && ctx.ShipmentHasActivePlan {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("shipment %s already has an active plan", ctx.ShipmentID),
		}
	}

	return GuardResult{Allowed: true}
}

// CanSubmitPlan evaluates whether a plan can be submitted for review.
// Rules:
// - Status must be "draft"
// - Plan must have content
func CanSubmitPlan(ctx SubmitPlanContext) GuardResult {
	if ctx.Status != "draft" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only submit draft plans (current status: %s)", ctx.Status),
		}
	}
	if !ctx.HasContent {
		return GuardResult{
			Allowed: false,
			Reason:  "cannot submit plan without content",
		}
	}
	return GuardResult{Allowed: true}
}

// CanApprovePlan evaluates whether a plan can be approved.
// Rules:
// - Status must be "pending_review"
// - Plan must not be pinned
func CanApprovePlan(ctx ApprovePlanContext) GuardResult {
	// Check status is pending_review (must check first - other statuses can't be approved regardless of pinned)
	if ctx.Status != "pending_review" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only approve plans pending review (current status: %s)", ctx.Status),
		}
	}

	// Check not pinned
	if ctx.IsPinned {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("cannot approve pinned plan %s. Unpin first with: orc plan unpin %s", ctx.PlanID, ctx.PlanID),
		}
	}

	return GuardResult{Allowed: true}
}

// CanDeletePlan evaluates whether a plan can be deleted.
// Rules:
// - Plan must not be pinned
func CanDeletePlan(ctx DeletePlanContext) GuardResult {
	if ctx.IsPinned {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("cannot delete pinned plan %s. Unpin first with: orc plan unpin %s", ctx.PlanID, ctx.PlanID),
		}
	}

	return GuardResult{Allowed: true}
}

// CanEscalatePlan evaluates whether a plan can be escalated.
// Rules:
// - Status must be "draft" or "pending_review"
// - Plan must have content
// - Reason must be provided
func CanEscalatePlan(ctx EscalatePlanContext) GuardResult {
	if ctx.Status != "draft" && ctx.Status != "pending_review" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only escalate draft or pending_review plans (current status: %s)", ctx.Status),
		}
	}
	if !ctx.HasContent {
		return GuardResult{
			Allowed: false,
			Reason:  "cannot escalate plan without content",
		}
	}
	if !ctx.HasReason {
		return GuardResult{
			Allowed: false,
			Reason:  "escalation reason is required",
		}
	}
	return GuardResult{Allowed: true}
}
