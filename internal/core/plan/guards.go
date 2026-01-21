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
	MissionID             string
	MissionExists         bool
	ShipmentID            string // Optional - empty string means no shipment
	ShipmentExists        bool   // Only checked if ShipmentID != ""
	ShipmentHasActivePlan bool   // Only checked if ShipmentID != ""
}

// ApprovePlanContext provides context for plan approval guards.
type ApprovePlanContext struct {
	PlanID   string
	Status   string // "draft", "approved"
	IsPinned bool
}

// DeletePlanContext provides context for plan deletion guards.
type DeletePlanContext struct {
	PlanID   string
	IsPinned bool
}

// CanCreatePlan evaluates whether a plan can be created.
// Rules:
// - Mission must exist
// - Shipment must exist if provided
// - No active (draft) plan for shipment if provided
func CanCreatePlan(ctx CreatePlanContext) GuardResult {
	// Check mission exists
	if !ctx.MissionExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("mission %s not found", ctx.MissionID),
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

// CanApprovePlan evaluates whether a plan can be approved.
// Rules:
// - Status must be "draft" (cannot re-approve)
// - Plan must not be pinned
func CanApprovePlan(ctx ApprovePlanContext) GuardResult {
	// Check status is draft (must check first - approved plans can't be approved regardless of pinned)
	if ctx.Status != "draft" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only approve draft plans (current status: %s)", ctx.Status),
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
