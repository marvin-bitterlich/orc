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

// StatusTransitionContext provides context for pause/resume/deploy guards.
type StatusTransitionContext struct {
	ShipmentID    string
	Status        string // "active", "paused", "complete"
	OpenTaskCount int    // count of non-complete tasks (for deploy guard)
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
// - Status must be "implementing" or "auto_implementing"
func CanPauseShipment(ctx StatusTransitionContext) GuardResult {
	if ctx.Status != "implementing" && ctx.Status != "auto_implementing" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only pause implementing shipments (current status: %s)", ctx.Status),
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

// CanDeployShipment evaluates whether a shipment can be marked as deployed.
// Rules:
// - Status must be "implementing", "auto_implementing", "implemented", or "complete"
// - All tasks must be complete (OpenTaskCount == 0)
func CanDeployShipment(ctx StatusTransitionContext) GuardResult {
	validStatuses := map[string]bool{
		"implementing":      true,
		"auto_implementing": true,
		"implemented":       true,
		"complete":          true,
	}
	if !validStatuses[ctx.Status] {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only deploy implementing/implemented shipments (current status: %s)", ctx.Status),
		}
	}

	if ctx.OpenTaskCount > 0 {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("cannot deploy: %d task(s) still open", ctx.OpenTaskCount),
		}
	}

	return GuardResult{Allowed: true}
}

// CanVerifyShipment evaluates whether a shipment can be marked as verified.
// Rules:
// - Status must be "deployed"
func CanVerifyShipment(ctx StatusTransitionContext) GuardResult {
	if ctx.Status != "deployed" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only verify deployed shipments (current status: %s)", ctx.Status),
		}
	}

	return GuardResult{Allowed: true}
}

// AutoTransitionContext provides context for automatic status transitions.
type AutoTransitionContext struct {
	CurrentStatus      string
	TriggerEvent       string // "focus", "task_created", "task_claimed", "task_completed", "deploy", "verify"
	TaskCount          int
	CompletedTaskCount int
}

// GetAutoTransitionStatus returns the new status for automatic transitions.
// Returns empty string if no transition should occur.
func GetAutoTransitionStatus(ctx AutoTransitionContext) string {
	switch ctx.TriggerEvent {
	case "focus":
		if ctx.CurrentStatus == "draft" {
			return "exploring"
		}
	case "task_created":
		if ctx.CurrentStatus == "draft" || ctx.CurrentStatus == "exploring" || ctx.CurrentStatus == "specced" {
			return "tasked"
		}
	case "task_claimed":
		if ctx.CurrentStatus == "tasked" || ctx.CurrentStatus == "ready_for_imp" || ctx.CurrentStatus == "exploring" || ctx.CurrentStatus == "specced" {
			return "implementing"
		}
	case "deploy":
		if ctx.CurrentStatus == "implementing" || ctx.CurrentStatus == "auto_implementing" || ctx.CurrentStatus == "implemented" || ctx.CurrentStatus == "complete" {
			return "deployed"
		}
	case "verify":
		if ctx.CurrentStatus == "deployed" {
			return "verified"
		}
	}
	return ""
}

// OverrideStatusContext provides context for status override guards.
type OverrideStatusContext struct {
	ShipmentID    string
	CurrentStatus string
	NewStatus     string
	Force         bool
}

// statusOrder defines the progression order for shipment statuses.
// Lower index = earlier in lifecycle.
var statusOrder = map[string]int{
	"draft":             0,
	"exploring":         1,
	"synthesizing":      2,
	"specced":           3,
	"planned":           4,
	"tasked":            5,
	"ready_for_imp":     6,
	"implementing":      7,
	"auto_implementing": 8,
	"implemented":       9,
	"deployed":          10,
	"verified":          11,
	"complete":          12,
}

// ValidStatuses returns all valid shipment statuses.
func ValidStatuses() []string {
	return []string{
		"draft", "exploring", "synthesizing", "specced", "planned",
		"tasked", "ready_for_imp", "implementing", "auto_implementing",
		"implemented", "deployed", "verified", "complete",
	}
}

// CanOverrideStatus evaluates whether a shipment status can be overridden.
// Rules:
// - New status must be valid
// - Backwards transitions require --force flag
func CanOverrideStatus(ctx OverrideStatusContext) GuardResult {
	// Rule 1: New status must be valid
	if _, ok := statusOrder[ctx.NewStatus]; !ok {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("invalid status '%s'. Valid statuses: %s", ctx.NewStatus, strings.Join(ValidStatuses(), ", ")),
		}
	}

	// Rule 2: Check for backwards transition
	currentIdx, currentOk := statusOrder[ctx.CurrentStatus]
	newIdx := statusOrder[ctx.NewStatus]

	// If current status is unknown, allow transition
	if !currentOk {
		return GuardResult{Allowed: true}
	}

	// Backwards transition requires force
	if newIdx < currentIdx && !ctx.Force {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("backwards transition from '%s' to '%s' requires --force flag", ctx.CurrentStatus, ctx.NewStatus),
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
