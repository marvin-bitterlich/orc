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
	ID                  string
	ShipmentID          sql.NullString
	CommissionID        string
	Title               string
	Description         sql.NullString
	Type                sql.NullString
	Status              string
	Priority            sql.NullString
	AssignedWorkbenchID sql.NullString
	Pinned              bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
	ClaimedAt           sql.NullTime
	CompletedAt         sql.NullTime
	PromotedFromID      sql.NullString
	PromotedFromType    sql.NullString
}

// Task status constants - simplified lifecycle
// Flow: open → in-progress → closed (with blocked as a lateral state)
const (
	TaskStatusOpen       = "open"
	TaskStatusInProgress = "in-progress"
	TaskStatusBlocked    = "blocked"
	TaskStatusClosed     = "closed"
)
