// Package investigation contains the pure business logic for investigation operations.
// Guards are pure functions that evaluate preconditions without side effects.
package investigation

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

// CreateInvestigationContext provides context for investigation creation guards.
type CreateInvestigationContext struct {
	MissionID     string
	MissionExists bool
}

// CompleteInvestigationContext provides context for investigation completion guards.
type CompleteInvestigationContext struct {
	InvestigationID string
	IsPinned        bool
}

// StatusTransitionContext provides context for pause/resume guards.
type StatusTransitionContext struct {
	InvestigationID string
	Status          string // "active", "paused", "complete"
}

// CanCreateInvestigation evaluates whether an investigation can be created.
// Rules:
// - Mission must exist
func CanCreateInvestigation(ctx CreateInvestigationContext) GuardResult {
	if !ctx.MissionExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("mission %s not found", ctx.MissionID),
		}
	}

	return GuardResult{Allowed: true}
}

// CanCompleteInvestigation evaluates whether an investigation can be completed.
// Rules:
// - Investigation must not be pinned
func CanCompleteInvestigation(ctx CompleteInvestigationContext) GuardResult {
	if ctx.IsPinned {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("cannot complete pinned investigation %s. Unpin first with: orc investigation unpin %s", ctx.InvestigationID, ctx.InvestigationID),
		}
	}

	return GuardResult{Allowed: true}
}

// CanPauseInvestigation evaluates whether an investigation can be paused.
// Rules:
// - Status must be "active"
func CanPauseInvestigation(ctx StatusTransitionContext) GuardResult {
	if ctx.Status != "active" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only pause active investigations (current status: %s)", ctx.Status),
		}
	}

	return GuardResult{Allowed: true}
}

// CanResumeInvestigation evaluates whether an investigation can be resumed.
// Rules:
// - Status must be "paused"
func CanResumeInvestigation(ctx StatusTransitionContext) GuardResult {
	if ctx.Status != "paused" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only resume paused investigations (current status: %s)", ctx.Status),
		}
	}

	return GuardResult{Allowed: true}
}
