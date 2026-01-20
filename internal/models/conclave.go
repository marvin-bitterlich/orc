package models

import (
	"database/sql"
	"time"
)

// Conclave represents an ideation session container.
// All database operations are handled by ConclaveRepository in the adapters layer.
type Conclave struct {
	ID              string
	MissionID       string
	Title           string
	Description     sql.NullString
	Status          string
	AssignedGroveID sql.NullString
	Pinned          bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CompletedAt     sql.NullTime
}

// Conclave status constants
const (
	ConclaveStatusActive   = "active"
	ConclaveStatusPaused   = "paused"
	ConclaveStatusComplete = "complete"
)
