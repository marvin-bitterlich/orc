// Package shipment contains the pure business logic for shipment operations.
// Guards are pure functions that evaluate preconditions without side effects.
package shipment

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

// CreateShipmentContext provides context for shipment creation guards.
type CreateShipmentContext struct {
	MissionID     string
	MissionExists bool
}

// CompleteShipmentContext provides context for shipment completion guards.
type CompleteShipmentContext struct {
	ShipmentID string
	IsPinned   bool
}

// StatusTransitionContext provides context for pause/resume guards.
type StatusTransitionContext struct {
	ShipmentID string
	Status     string // "active", "paused", "complete"
}

// AssignGroveContext provides context for grove assignment guards.
type AssignGroveContext struct {
	ShipmentID        string
	GroveID           string
	ShipmentExists    bool
	GroveAssignedToID string // ID of shipment grove is assigned to, empty if unassigned
}

// CanCreateShipment evaluates whether a shipment can be created.
// Rules:
// - Mission must exist
func CanCreateShipment(ctx CreateShipmentContext) GuardResult {
	if !ctx.MissionExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("mission %s not found", ctx.MissionID),
		}
	}

	return GuardResult{Allowed: true}
}

// CanCompleteShipment evaluates whether a shipment can be completed.
// Rules:
// - Shipment must not be pinned
func CanCompleteShipment(ctx CompleteShipmentContext) GuardResult {
	if ctx.IsPinned {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("cannot complete pinned shipment %s. Unpin first with: orc shipment unpin %s", ctx.ShipmentID, ctx.ShipmentID),
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

// CanAssignGrove evaluates whether a grove can be assigned to a shipment.
// Rules:
// - Shipment must exist
// - Grove must not be assigned to another shipment
func CanAssignGrove(ctx AssignGroveContext) GuardResult {
	// Rule 1: Shipment must exist
	if !ctx.ShipmentExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("shipment %s not found", ctx.ShipmentID),
		}
	}

	// Rule 2: Grove must not be assigned to another shipment
	// If the grove is assigned to this same shipment, that's OK (idempotent)
	if ctx.GroveAssignedToID != "" && ctx.GroveAssignedToID != ctx.ShipmentID {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("grove already assigned to shipment %s", ctx.GroveAssignedToID),
		}
	}

	return GuardResult{Allowed: true}
}
