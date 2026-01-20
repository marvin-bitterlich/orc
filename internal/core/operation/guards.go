// Package operation contains the pure business logic for operation operations.
// Guards are pure functions that evaluate preconditions without side effects.
package operation

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

// CreateOperationContext provides context for operation creation guards.
type CreateOperationContext struct {
	MissionID     string
	MissionExists bool
}

// CompleteOperationContext provides context for operation completion guards.
type CompleteOperationContext struct {
	OperationID string
	Status      string // "ready", "complete"
}

// CanCreateOperation evaluates whether an operation can be created.
// Rules:
// - Mission must exist
func CanCreateOperation(ctx CreateOperationContext) GuardResult {
	if !ctx.MissionExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("mission %s not found", ctx.MissionID),
		}
	}

	return GuardResult{Allowed: true}
}

// CanCompleteOperation evaluates whether an operation can be completed.
// Rules:
// - Status must be "ready" (cannot re-complete)
func CanCompleteOperation(ctx CompleteOperationContext) GuardResult {
	if ctx.Status != "ready" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only complete ready operations (current status: %s)", ctx.Status),
		}
	}

	return GuardResult{Allowed: true}
}
