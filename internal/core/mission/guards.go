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

// GuardContext provides the context needed for guard evaluation.
// This is the input to all guard functions.
type GuardContext struct {
	AgentType AgentType
	AgentID   string // Full agent ID (e.g., "ORC" or "IMP-GROVE-001")
	MissionID string // Current mission context (may be empty)
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
