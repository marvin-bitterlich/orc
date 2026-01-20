// Package conclave contains the pure business logic for conclave operations.
// Guards are pure functions that evaluate preconditions without side effects.
package conclave

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

// CreateConclaveContext provides context for conclave creation guards.
type CreateConclaveContext struct {
	MissionID     string
	MissionExists bool
}

// CompleteConclaveContext provides context for conclave completion guards.
type CompleteConclaveContext struct {
	ConclaveID string
	IsPinned   bool
}

// StatusTransitionContext provides context for pause/resume guards.
type StatusTransitionContext struct {
	ConclaveID string
	Status     string // "active", "paused", "complete"
}

// CanCreateConclave evaluates whether a conclave can be created.
// Rules:
// - Mission must exist
func CanCreateConclave(ctx CreateConclaveContext) GuardResult {
	if !ctx.MissionExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("mission %s not found", ctx.MissionID),
		}
	}

	return GuardResult{Allowed: true}
}

// CanCompleteConclave evaluates whether a conclave can be completed.
// Rules:
// - Conclave must not be pinned
func CanCompleteConclave(ctx CompleteConclaveContext) GuardResult {
	if ctx.IsPinned {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("cannot complete pinned conclave %s. Unpin first with: orc conclave unpin %s", ctx.ConclaveID, ctx.ConclaveID),
		}
	}

	return GuardResult{Allowed: true}
}

// CanPauseConclave evaluates whether a conclave can be paused.
// Rules:
// - Status must be "active"
func CanPauseConclave(ctx StatusTransitionContext) GuardResult {
	if ctx.Status != "active" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only pause active conclaves (current status: %s)", ctx.Status),
		}
	}

	return GuardResult{Allowed: true}
}

// CanResumeConclave evaluates whether a conclave can be resumed.
// Rules:
// - Status must be "paused"
func CanResumeConclave(ctx StatusTransitionContext) GuardResult {
	if ctx.Status != "paused" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only resume paused conclaves (current status: %s)", ctx.Status),
		}
	}

	return GuardResult{Allowed: true}
}
