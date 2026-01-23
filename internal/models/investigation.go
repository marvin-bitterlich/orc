package models

import (
	"database/sql"
	"time"
)

// Investigation represents a research container in the ORC ledger.
// Status can be: active, paused, complete
type Investigation struct {
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

// Investigation status constants
const (
	InvestigationStatusActive   = "active"
	InvestigationStatusPaused   = "paused"
	InvestigationStatusComplete = "complete"
)
