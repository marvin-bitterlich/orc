// Package cyclereceipt contains the pure business logic for cycle receipt operations.
// Guards are pure functions that evaluate preconditions without side effects.
package cyclereceipt

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

// CreateCRECContext provides context for CREC creation guards.
type CreateCRECContext struct {
	CWOID            string
	CWOExists        bool
	CWOHasCREC       bool
	DeliveredOutcome string
}

// StatusTransitionContext provides context for status transition guards.
type StatusTransitionContext struct {
	CRECID        string
	CurrentStatus string
	CWOStatus     string
	CWOExists     bool
}

// CanCreateCREC evaluates whether a CREC can be created.
// Rules:
// - CWO must exist
// - CWO must not already have a CREC (1:1 constraint)
// - DeliveredOutcome must not be empty
func CanCreateCREC(ctx CreateCRECContext) GuardResult {
	// Rule 1: CWO must exist
	if !ctx.CWOExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("CWO %s not found", ctx.CWOID),
		}
	}

	// Rule 2: CWO must not already have a CREC
	if ctx.CWOHasCREC {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("CWO %s already has a CREC", ctx.CWOID),
		}
	}

	// Rule 3: DeliveredOutcome must not be empty
	if strings.TrimSpace(ctx.DeliveredOutcome) == "" {
		return GuardResult{
			Allowed: false,
			Reason:  "delivered outcome cannot be empty",
		}
	}

	return GuardResult{Allowed: true}
}

// CanSubmit evaluates whether a CREC can be submitted.
// Rules:
// - CREC must be in draft status
// - CWO must be complete
func CanSubmit(ctx StatusTransitionContext) GuardResult {
	// Rule 1: Must be in draft status
	if ctx.CurrentStatus != "draft" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only submit draft CRECs (current status: %s)", ctx.CurrentStatus),
		}
	}

	// Rule 2: CWO must exist
	if !ctx.CWOExists {
		return GuardResult{
			Allowed: false,
			Reason:  "cannot submit CREC: parent CWO no longer exists",
		}
	}

	// Rule 3: CWO must be complete
	if ctx.CWOStatus != "complete" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("cannot submit CREC: parent CWO is not complete (status: %s)", ctx.CWOStatus),
		}
	}

	return GuardResult{Allowed: true}
}

// CanVerify evaluates whether a CREC can be verified.
// Rules:
// - CREC must be in submitted status
func CanVerify(ctx StatusTransitionContext) GuardResult {
	// Rule 1: Must be in submitted status
	if ctx.CurrentStatus != "submitted" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only verify submitted CRECs (current status: %s)", ctx.CurrentStatus),
		}
	}

	return GuardResult{Allowed: true}
}
