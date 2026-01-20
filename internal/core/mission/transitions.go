// Package mission contains the pure business logic for mission operations.
// This is part of the Functional Core - no I/O, only pure functions.
package mission

import "time"

// MissionStatus represents the possible states of a mission.
type MissionStatus string

const (
	StatusActive   MissionStatus = "active"
	StatusPaused   MissionStatus = "paused"
	StatusComplete MissionStatus = "complete"
	StatusArchived MissionStatus = "archived"
)

// StatusTransitionResult contains the result of a status transition.
// This is a value object that captures both the new status and any
// side effects (like setting CompletedAt timestamp).
type StatusTransitionResult struct {
	NewStatus   MissionStatus
	CompletedAt *time.Time // Set when transitioning to complete status
}

// ApplyStatusTransition applies a status transition and returns the result.
// This is a pure function that captures the business rule:
// - When status becomes "complete", CompletedAt should be set to the current time.
// The caller should pass the current time to enable testing.
func ApplyStatusTransition(newStatus MissionStatus, now time.Time) StatusTransitionResult {
	result := StatusTransitionResult{
		NewStatus: newStatus,
	}

	if newStatus == StatusComplete {
		result.CompletedAt = &now
	}

	return result
}

// InitialStatus returns the initial status for a new mission.
// This is a pure function that defines the business rule for new missions.
func InitialStatus() MissionStatus {
	return StatusActive
}
