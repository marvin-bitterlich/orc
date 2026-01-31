// Package shipment contains the pure business logic for shipment operations.
// Guards are pure functions that evaluate preconditions without side effects.
package shipment

import (
	"fmt"
	"strings"
)

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

// CreateShipmentContext provides context for shipment creation guards.
type CreateShipmentContext struct {
	CommissionID     string
	CommissionExists bool
}

// TaskSummary contains minimal task info for guard evaluation.
type TaskSummary struct {
	ID     string
	Status string
}

// CompleteShipmentContext provides context for shipment completion guards.
type CompleteShipmentContext struct {
	ShipmentID      string
	IsPinned        bool
	Tasks           []TaskSummary
	ForceCompletion bool // Skip task check if explicitly forced
}

// StatusTransitionContext provides context for pause/resume guards.
type StatusTransitionContext struct {
	ShipmentID string
	Status     string // "active", "paused", "complete"
}

// AssignWorkbenchContext provides context for workbench assignment guards.
type AssignWorkbenchContext struct {
	ShipmentID            string
	WorkbenchID           string
	ShipmentExists        bool
	WorkbenchAssignedToID string // ID of shipment workbench is assigned to, empty if unassigned
}

// CanCreateShipment evaluates whether a shipment can be created.
// Rules:
// - Commission must exist
func CanCreateShipment(ctx CreateShipmentContext) GuardResult {
	if !ctx.CommissionExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("commission %s not found", ctx.CommissionID),
		}
	}

	return GuardResult{Allowed: true}
}

// CanCompleteShipment evaluates whether a shipment can be completed.
// Rules:
// - Shipment must not be pinned
// - All tasks must be complete (unless forced)
func CanCompleteShipment(ctx CompleteShipmentContext) GuardResult {
	if ctx.IsPinned {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("cannot complete pinned shipment %s. Unpin first with: orc shipment unpin %s", ctx.ShipmentID, ctx.ShipmentID),
		}
	}

	// Check for incomplete tasks (unless force flag is set)
	if !ctx.ForceCompletion {
		var incomplete []string
		for _, t := range ctx.Tasks {
			if t.Status != "complete" {
				incomplete = append(incomplete, t.ID)
			}
		}
		if len(incomplete) > 0 {
			return GuardResult{
				Allowed: false,
				Reason: fmt.Sprintf("cannot complete shipment: %d task(s) incomplete (%s). Use --force to complete anyway",
					len(incomplete), strings.Join(incomplete, ", ")),
			}
		}
	}

	return GuardResult{Allowed: true}
}

// CanPauseShipment evaluates whether a shipment can be paused.
// Rules:
// - Status must be "active"
func CanPauseShipment(ctx StatusTransitionContext) GuardResult {
	if ctx.Status != "active" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only pause active shipments (current status: %s)", ctx.Status),
		}
	}

	return GuardResult{Allowed: true}
}

// CanResumeShipment evaluates whether a shipment can be resumed.
// Rules:
// - Status must be "paused"
func CanResumeShipment(ctx StatusTransitionContext) GuardResult {
	if ctx.Status != "paused" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only resume paused shipments (current status: %s)", ctx.Status),
		}
	}

	return GuardResult{Allowed: true}
}

// CanAssignWorkbench evaluates whether a workbench can be assigned to a shipment.
// Rules:
// - Shipment must exist
// - Workbench must not be assigned to another shipment
func CanAssignWorkbench(ctx AssignWorkbenchContext) GuardResult {
	// Rule 1: Shipment must exist
	if !ctx.ShipmentExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("shipment %s not found", ctx.ShipmentID),
		}
	}

	// Rule 2: Workbench must not be assigned to another shipment
	// If the workbench is assigned to this same shipment, that's OK (idempotent)
	if ctx.WorkbenchAssignedToID != "" && ctx.WorkbenchAssignedToID != ctx.ShipmentID {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("workbench already assigned to shipment %s", ctx.WorkbenchAssignedToID),
		}
	}

	return GuardResult{Allowed: true}
}
