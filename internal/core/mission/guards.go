// Package mission contains the pure business logic for mission operations.
// This is part of the Functional Core - no I/O, only pure functions.
package mission

import "fmt"

// AgentType represents the type of agent in the mission domain.
// Defined here to avoid import cycles with internal/agent.
type AgentType string

const (
	// AgentTypeORC represents the orchestrator agent.
	AgentTypeORC AgentType = "ORC"
	// AgentTypeIMP represents an implementation agent in a grove.
	AgentTypeIMP AgentType = "IMP"
)

// GuardContext provides the context needed for agent-based guard evaluation.
// This is the input to agent permission guards.
type GuardContext struct {
	AgentType AgentType
	AgentID   string // Full agent ID (e.g., "ORC" or "IMP-GROVE-001")
	MissionID string // Current mission context (may be empty)
}

// MissionStateContext provides context for state-based mission guards.
// Used when checking if mission state allows a transition.
type MissionStateContext struct {
	MissionID string
	IsPinned  bool
}

// DeleteContext provides context for mission deletion guards.
// Populated by the caller with pre-fetched dependency counts.
type DeleteContext struct {
	MissionID     string
	ShipmentCount int
	GroveCount    int
	ForceDelete   bool
}

// GuardResult represents the outcome of a guard evaluation.
type GuardResult struct {
	Allowed bool
	Reason  string // Human-readable reason (populated when not allowed)
}

// Error returns the guard result as an error if not allowed, nil otherwise.
func (r GuardResult) Error() error {
	if r.Allowed {
		return nil
	}
	return fmt.Errorf("%s", r.Reason)
}

// CanCreateMission evaluates whether the current agent can create a mission.
// Rule: Only ORC can create missions. IMPs work within existing missions.
func CanCreateMission(ctx GuardContext) GuardResult {
	if ctx.AgentType == AgentTypeIMP {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("IMPs cannot create missions - only ORC can create missions (agent: %s)", ctx.AgentID),
		}
	}
	return GuardResult{Allowed: true}
}

// CanStartMission evaluates whether the current agent can start a mission.
// Rule: Only ORC can start missions. IMPs cannot control mission lifecycle.
func CanStartMission(ctx GuardContext) GuardResult {
	if ctx.AgentType == AgentTypeIMP {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("IMPs cannot start missions - only ORC can start missions (agent: %s)", ctx.AgentID),
		}
	}
	return GuardResult{Allowed: true}
}

// CanLaunchMission evaluates whether the current agent can launch a mission.
// Rule: Only ORC can launch missions. Launch = create + start.
func CanLaunchMission(ctx GuardContext) GuardResult {
	if ctx.AgentType == AgentTypeIMP {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("IMPs cannot launch missions - only ORC can launch missions (agent: %s)", ctx.AgentID),
		}
	}
	return GuardResult{Allowed: true}
}

// CanCompleteMission evaluates whether a mission can be marked complete.
// Rule: Pinned missions cannot be completed.
func CanCompleteMission(ctx MissionStateContext) GuardResult {
	if ctx.IsPinned {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("Cannot complete pinned mission %s. Unpin first with: orc mission unpin %s", ctx.MissionID, ctx.MissionID),
		}
	}
	return GuardResult{Allowed: true}
}

// CanArchiveMission evaluates whether a mission can be archived.
// Rule: Pinned missions cannot be archived.
func CanArchiveMission(ctx MissionStateContext) GuardResult {
	if ctx.IsPinned {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("Cannot archive pinned mission %s. Unpin first with: orc mission unpin %s", ctx.MissionID, ctx.MissionID),
		}
	}
	return GuardResult{Allowed: true}
}

// CanDeleteMission evaluates whether a mission can be deleted.
// Rule: Missions with dependents require --force flag.
func CanDeleteMission(ctx DeleteContext) GuardResult {
	hasDependents := ctx.ShipmentCount > 0 || ctx.GroveCount > 0
	if hasDependents && !ctx.ForceDelete {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("Mission %s has %d shipments and %d groves. Use --force to delete anyway", ctx.MissionID, ctx.ShipmentCount, ctx.GroveCount),
		}
	}
	return GuardResult{Allowed: true}
}

// PinContext provides context for mission pin/unpin guards.
type PinContext struct {
	MissionID     string
	MissionExists bool
	IsPinned      bool
}

// CanPinMission evaluates whether a mission can be pinned.
// Rule: Mission must exist to be pinned.
func CanPinMission(ctx PinContext) GuardResult {
	if !ctx.MissionExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("Mission %s not found", ctx.MissionID),
		}
	}
	// If already pinned, this is a no-op (still allowed, just does nothing)
	return GuardResult{Allowed: true}
}

// CanUnpinMission evaluates whether a mission can be unpinned.
// Rule: Mission must exist to be unpinned.
func CanUnpinMission(ctx PinContext) GuardResult {
	if !ctx.MissionExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("Mission %s not found", ctx.MissionID),
		}
	}
	// If already unpinned, this is a no-op (still allowed, just does nothing)
	return GuardResult{Allowed: true}
}
