// Package models contains domain types for ORC entities.
// SQL persistence has been moved to internal/adapters/sqlite/*.go
package models

import (
	"database/sql"
	"time"
)

// Task represents a task entity.
// This is the domain type used within the models package.
// For persistence, use the repository interfaces in ports/secondary.
type Task struct {
	ID               string
	ShipmentID       sql.NullString
	MissionID        string
	Title            string
	Description      sql.NullString
	Type             sql.NullString
	Status           string
	Priority         sql.NullString
	AssignedGroveID  sql.NullString
	Pinned           bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ClaimedAt        sql.NullTime
	CompletedAt      sql.NullTime
	ConclaveID       sql.NullString
	PromotedFromID   sql.NullString
	PromotedFromType sql.NullString
}

// Task status constants
const (
	TaskStatusReady      = "ready"
	TaskStatusInProgress = "in_progress"
	TaskStatusPaused     = "paused"
	TaskStatusComplete   = "complete"
	TaskStatusBlocked    = "blocked"
)
