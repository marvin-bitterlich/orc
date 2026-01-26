// Package workbench contains the pure business logic for workbench operations.
// Guards are pure functions that evaluate preconditions without side effects.
package workbench

import "fmt"

// AgentType represents the type of agent in the workbench domain.
type AgentType string

const (
	// AgentTypeORC represents the orchestrator agent.
	AgentTypeORC AgentType = "ORC"
	// AgentTypeIMP represents an implementation agent in a workbench.
	AgentTypeIMP AgentType = "IMP"
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

// GuardContext provides context for agent-based guard evaluation.
type GuardContext struct {
	AgentType  AgentType
	AgentID    string
	WorkshopID string
}

// CreateWorkbenchContext provides context for workbench creation guards.
type CreateWorkbenchContext struct {
	GuardContext
	WorkshopExists bool
}

// CanCreateWorkbench evaluates whether a workbench can be created.
// Rules:
// - Workshop must exist
func CanCreateWorkbench(ctx CreateWorkbenchContext) GuardResult {
	// Workshop must exist
	if !ctx.WorkshopExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("cannot create workbench: workshop %s not found", ctx.WorkshopID),
		}
	}

	return GuardResult{Allowed: true}
}

// OpenWorkbenchContext provides context for workbench open guards.
type OpenWorkbenchContext struct {
	WorkbenchID     string
	WorkbenchExists bool
	PathExists      bool
	InTMuxSession   bool
}

// CanOpenWorkbench evaluates whether a workbench can be opened in TMux.
// Rules:
// - Workbench must exist in database
// - Workbench path must exist on filesystem
// - Must be in a TMux session
func CanOpenWorkbench(ctx OpenWorkbenchContext) GuardResult {
	if !ctx.WorkbenchExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("workbench %s not found", ctx.WorkbenchID),
		}
	}

	if !ctx.PathExists {
		return GuardResult{
			Allowed: false,
			Reason:  "workbench worktree not found - run 'orc workbench create' to materialize",
		}
	}

	if !ctx.InTMuxSession {
		return GuardResult{
			Allowed: false,
			Reason:  "not in a TMux session - run this command from within a TMux session",
		}
	}

	return GuardResult{Allowed: true}
}

// DeleteWorkbenchContext provides context for workbench deletion guards.
type DeleteWorkbenchContext struct {
	WorkbenchID     string
	ActiveTaskCount int
	ForceDelete     bool
}

// CanDeleteWorkbench evaluates whether a workbench can be deleted.
// Rules:
// - Workbenches with active tasks require --force
func CanDeleteWorkbench(ctx DeleteWorkbenchContext) GuardResult {
	if ctx.ActiveTaskCount > 0 && !ctx.ForceDelete {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("workbench %s has %d active tasks. Use --force to delete anyway", ctx.WorkbenchID, ctx.ActiveTaskCount),
		}
	}

	return GuardResult{Allowed: true}
}

// CanRenameWorkbench evaluates whether a workbench can be renamed.
// Rules:
// - Workbench must exist
func CanRenameWorkbench(workbenchExists bool, workbenchID string) GuardResult {
	if !workbenchExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("workbench %s not found", workbenchID),
		}
	}
	return GuardResult{Allowed: true}
}
