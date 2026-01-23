// Package commission contains the pure business logic for commission operations.
// This is part of the Functional Core - no I/O, only pure functions.
package commission

import "fmt"

// AgentType represents the type of agent in the commission domain.
// Defined here to avoid import cycles with internal/agent.
type AgentType string

const (
	// AgentTypeORC represents the orchestrator agent.
	AgentTypeORC AgentType = "ORC"
	// AgentTypeIMP represents an implementation agent in a workbench.
	AgentTypeIMP AgentType = "IMP"
)

// GuardContext provides the context needed for agent-based guard evaluation.
// This is the input to agent permission guards.
type GuardContext struct {
	AgentType    AgentType
	AgentID      string // Full agent ID (e.g., "ORC" or "IMP-BENCH-001")
	CommissionID string // Current commission context (may be empty)
}

// CommissionStateContext provides context for state-based commission guards.
// Used when checking if commission state allows a transition.
type CommissionStateContext struct {
	CommissionID string
	IsPinned     bool
}

// DeleteContext provides context for commission deletion guards.
// Populated by the caller with pre-fetched dependency counts.
type DeleteContext struct {
	CommissionID   string
	ShipmentCount  int
	WorkbenchCount int
	ForceDelete    bool
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

// CanCreateCommission evaluates whether the current agent can create a commission.
// Rule: Only ORC can create commissions. IMPs work within existing commissions.
func CanCreateCommission(ctx GuardContext) GuardResult {
	if ctx.AgentType == AgentTypeIMP {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("IMPs cannot create commissions - only ORC can create commissions (agent: %s)", ctx.AgentID),
		}
	}
	return GuardResult{Allowed: true}
}

// CanStartCommission evaluates whether the current agent can start a commission.
// Rule: Only ORC can start commissions. IMPs cannot control commission lifecycle.
func CanStartCommission(ctx GuardContext) GuardResult {
	if ctx.AgentType == AgentTypeIMP {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("IMPs cannot start commissions - only ORC can start commissions (agent: %s)", ctx.AgentID),
		}
	}
	return GuardResult{Allowed: true}
}

// CanLaunchCommission evaluates whether the current agent can launch a commission.
// Rule: Only ORC can launch commissions. Launch = create + start.
func CanLaunchCommission(ctx GuardContext) GuardResult {
	if ctx.AgentType == AgentTypeIMP {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("IMPs cannot launch commissions - only ORC can launch commissions (agent: %s)", ctx.AgentID),
		}
	}
	return GuardResult{Allowed: true}
}

// CanCompleteCommission evaluates whether a commission can be marked complete.
// Rule: Pinned commissions cannot be completed.
func CanCompleteCommission(ctx CommissionStateContext) GuardResult {
	if ctx.IsPinned {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("Cannot complete pinned commission %s. Unpin first with: orc commission unpin %s", ctx.CommissionID, ctx.CommissionID),
		}
	}
	return GuardResult{Allowed: true}
}

// CanArchiveCommission evaluates whether a commission can be archived.
// Rule: Pinned commissions cannot be archived.
func CanArchiveCommission(ctx CommissionStateContext) GuardResult {
	if ctx.IsPinned {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("Cannot archive pinned commission %s. Unpin first with: orc commission unpin %s", ctx.CommissionID, ctx.CommissionID),
		}
	}
	return GuardResult{Allowed: true}
}

// CanDeleteCommission evaluates whether a commission can be deleted.
// Rule: Commissions with dependents require --force flag.
func CanDeleteCommission(ctx DeleteContext) GuardResult {
	hasDependents := ctx.ShipmentCount > 0 || ctx.WorkbenchCount > 0
	if hasDependents && !ctx.ForceDelete {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("Commission %s has %d shipments and %d workbenches. Use --force to delete anyway", ctx.CommissionID, ctx.ShipmentCount, ctx.WorkbenchCount),
		}
	}
	return GuardResult{Allowed: true}
}

// PinContext provides context for commission pin/unpin guards.
type PinContext struct {
	CommissionID     string
	CommissionExists bool
	IsPinned         bool
}

// CanPinCommission evaluates whether a commission can be pinned.
// Rule: Commission must exist to be pinned.
func CanPinCommission(ctx PinContext) GuardResult {
	if !ctx.CommissionExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("Commission %s not found", ctx.CommissionID),
		}
	}
	// If already pinned, this is a no-op (still allowed, just does nothing)
	return GuardResult{Allowed: true}
}

// CanUnpinCommission evaluates whether a commission can be unpinned.
// Rule: Commission must exist to be unpinned.
func CanUnpinCommission(ctx PinContext) GuardResult {
	if !ctx.CommissionExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("Commission %s not found", ctx.CommissionID),
		}
	}
	// If already unpinned, this is a no-op (still allowed, just does nothing)
	return GuardResult{Allowed: true}
}
