// Package grove contains the pure business logic for grove operations.
// Guards are pure functions that evaluate preconditions without side effects.
package grove

import "fmt"

// AgentType represents the type of agent in the grove domain.
// Mirrored from mission to avoid import cycles.
type AgentType string

const (
	// AgentTypeORC represents the orchestrator agent.
	AgentTypeORC AgentType = "ORC"
	// AgentTypeIMP represents an implementation agent in a grove.
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
	AgentType AgentType
	AgentID   string
	MissionID string
}

// CreateGroveContext provides context for grove creation guards.
type CreateGroveContext struct {
	GuardContext
	MissionExists bool
}

// OpenGroveContext provides context for grove open guards.
type OpenGroveContext struct {
	GroveID       string
	GroveExists   bool
	PathExists    bool
	InTMuxSession bool
}

// DeleteGroveContext provides context for grove deletion guards.
type DeleteGroveContext struct {
	GroveID         string
	ActiveTaskCount int
	ForceDelete     bool
}

// CanCreateGrove evaluates whether a grove can be created.
// Rules:
// - Only ORC can create groves (IMPs work within existing groves)
// - Mission must exist
func CanCreateGrove(ctx CreateGroveContext) GuardResult {
	// Rule 1: Only ORC can create groves
	if ctx.AgentType == AgentTypeIMP {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("IMPs cannot create groves - only ORC can create groves (agent: %s)", ctx.AgentID),
		}
	}

	// Rule 2: Mission must exist
	if !ctx.MissionExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("cannot create grove: mission %s not found", ctx.MissionID),
		}
	}

	return GuardResult{Allowed: true}
}

// CanOpenGrove evaluates whether a grove can be opened in TMux.
// Rules:
// - Grove must exist in database
// - Grove path must exist on filesystem
// - Must be in a TMux session
func CanOpenGrove(ctx OpenGroveContext) GuardResult {
	if !ctx.GroveExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("grove %s not found", ctx.GroveID),
		}
	}

	if !ctx.PathExists {
		return GuardResult{
			Allowed: false,
			Reason:  "grove worktree not found - run 'orc grove create' to materialize",
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

// CanDeleteGrove evaluates whether a grove can be deleted.
// Rules:
// - Groves with active tasks require --force
func CanDeleteGrove(ctx DeleteGroveContext) GuardResult {
	if ctx.ActiveTaskCount > 0 && !ctx.ForceDelete {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("grove %s has %d active tasks. Use --force to delete anyway", ctx.GroveID, ctx.ActiveTaskCount),
		}
	}

	return GuardResult{Allowed: true}
}

// CanRenameGrove evaluates whether a grove can be renamed.
// Rules:
// - Grove must exist
func CanRenameGrove(groveExists bool, groveID string) GuardResult {
	if !groveExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("grove %s not found", groveID),
		}
	}
	return GuardResult{Allowed: true}
}
