// Package pr contains the pure business logic for pull request operations.
// Guards are pure functions that evaluate preconditions without side effects.
package pr

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

// CreatePRContext provides context for PR creation guards.
type CreatePRContext struct {
	ShipmentID     string
	RepoID         string
	ShipmentExists bool
	ShipmentStatus string // "active", "paused", "complete"
	ShipmentHasPR  bool
	RepoExists     bool
}

// OpenPRContext provides context for opening a draft PR.
type OpenPRContext struct {
	PRID   string
	Status string
}

// ApprovePRContext provides context for approving a PR.
type ApprovePRContext struct {
	PRID   string
	Status string
}

// MergePRContext provides context for merging a PR.
type MergePRContext struct {
	PRID   string
	Status string
}

// ClosePRContext provides context for closing a PR.
type ClosePRContext struct {
	PRID   string
	Status string
}

// CanCreatePR evaluates whether a PR can be created.
// Rules:
// - Shipment must exist
// - Shipment must be active
// - Shipment must not already have a PR
// - Repository must exist
func CanCreatePR(ctx CreatePRContext) GuardResult {
	// Rule 1: Shipment must exist
	if !ctx.ShipmentExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("shipment %s not found", ctx.ShipmentID),
		}
	}

	// Rule 2: Shipment must be active
	if ctx.ShipmentStatus != "active" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only create PR for active shipments (current status: %s)", ctx.ShipmentStatus),
		}
	}

	// Rule 3: Shipment must not already have a PR
	if ctx.ShipmentHasPR {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("shipment %s already has a PR", ctx.ShipmentID),
		}
	}

	// Rule 4: Repository must exist
	if !ctx.RepoExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("repository %s not found", ctx.RepoID),
		}
	}

	return GuardResult{Allowed: true}
}

// CanOpenPR evaluates whether a PR can be opened.
// Rules:
// - Status must be "draft"
func CanOpenPR(ctx OpenPRContext) GuardResult {
	if ctx.Status != "draft" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only open draft PRs (current status: %s)", ctx.Status),
		}
	}

	return GuardResult{Allowed: true}
}

// CanApprovePR evaluates whether a PR can be approved.
// Rules:
// - Status must be "open"
func CanApprovePR(ctx ApprovePRContext) GuardResult {
	if ctx.Status != "open" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only approve open PRs (current status: %s)", ctx.Status),
		}
	}

	return GuardResult{Allowed: true}
}

// CanMergePR evaluates whether a PR can be merged.
// Rules:
// - Status must be "open" or "approved"
func CanMergePR(ctx MergePRContext) GuardResult {
	if ctx.Status != "open" && ctx.Status != "approved" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only merge open or approved PRs (current status: %s)", ctx.Status),
		}
	}

	return GuardResult{Allowed: true}
}

// CanClosePR evaluates whether a PR can be closed.
// Rules:
// - Status must be "open" or "approved" (not merged, not draft)
func CanClosePR(ctx ClosePRContext) GuardResult {
	if ctx.Status != "open" && ctx.Status != "approved" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only close open or approved PRs (current status: %s)", ctx.Status),
		}
	}

	return GuardResult{Allowed: true}
}
