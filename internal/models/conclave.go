package models

import (
	"database/sql"
	"time"
)

// Conclave represents an ideation session container.
// All database operations are handled by ConclaveRepository in the adapters layer.
type Conclave struct {
	ID                  string
	CommissionID        string
	Title               string
	Description         sql.NullString
	Status              string
	AssignedWorkbenchID sql.NullString
	Pinned              bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
	CompletedAt         sql.NullTime
}

// Conclave status constants
const (
	ConclaveStatusOpen   = "open"
	ConclaveStatusPaused = "paused"
	ConclaveStatusClosed = "closed"
)
