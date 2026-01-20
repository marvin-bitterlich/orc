// Package task contains the pure business logic for task operations.
// Guards are pure functions that evaluate preconditions without side effects.
package task

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

// CreateTaskContext provides context for task creation guards.
type CreateTaskContext struct {
	MissionID      string
	MissionExists  bool
	ShipmentID     string // optional, empty if not specified
	ShipmentExists bool   // only checked if ShipmentID != ""
}

// CompleteTaskContext provides context for task completion guards.
type CompleteTaskContext struct {
	TaskID   string
	IsPinned bool
}

// StatusTransitionContext provides context for pause/resume guards.
type StatusTransitionContext struct {
	TaskID string
	Status string // "ready", "in_progress", "paused", "complete"
}

// TagTaskContext provides context for tag operation guards.
type TagTaskContext struct {
	TaskID          string
	ExistingTagID   string // empty if no tag
	ExistingTagName string
}

// CanCreateTask evaluates whether a task can be created.
// Rules:
// - Mission must exist
// - Shipment must exist (if shipment_id provided)
func CanCreateTask(ctx CreateTaskContext) GuardResult {
	// Rule 1: Mission must exist
	if !ctx.MissionExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("mission %s not found", ctx.MissionID),
		}
	}

	// Rule 2: Shipment must exist (if provided)
	if ctx.ShipmentID != "" && !ctx.ShipmentExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("shipment %s not found", ctx.ShipmentID),
		}
	}

	return GuardResult{Allowed: true}
}

// CanCompleteTask evaluates whether a task can be completed.
// Rules:
// - Task must not be pinned
func CanCompleteTask(ctx CompleteTaskContext) GuardResult {
	if ctx.IsPinned {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("cannot complete pinned task %s. Unpin first with: orc task unpin %s", ctx.TaskID, ctx.TaskID),
		}
	}

	return GuardResult{Allowed: true}
}

// CanPauseTask evaluates whether a task can be paused.
// Rules:
// - Status must be "in_progress"
func CanPauseTask(ctx StatusTransitionContext) GuardResult {
	if ctx.Status != "in_progress" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only pause in_progress tasks (current status: %s)", ctx.Status),
		}
	}

	return GuardResult{Allowed: true}
}

// CanResumeTask evaluates whether a task can be resumed.
// Rules:
// - Status must be "paused"
func CanResumeTask(ctx StatusTransitionContext) GuardResult {
	if ctx.Status != "paused" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only resume paused tasks (current status: %s)", ctx.Status),
		}
	}

	return GuardResult{Allowed: true}
}

// CanTagTask evaluates whether a tag can be added to a task.
// Rules:
// - Task must not already have a tag (one tag per task limit)
func CanTagTask(ctx TagTaskContext) GuardResult {
	if ctx.ExistingTagID != "" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("task %s already has tag '%s' (one tag per task limit)\nRemove existing tag first with: orc task untag %s", ctx.TaskID, ctx.ExistingTagName, ctx.TaskID),
		}
	}

	return GuardResult{Allowed: true}
}
