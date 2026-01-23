package models

import (
	"database/sql"
	"time"
)

type Tome struct {
	ID                  string
	CommissionID        string
	ConclaveID          sql.NullString // Optional parent conclave
	Title               string
	Description         sql.NullString
	Status              string
	AssignedWorkbenchID sql.NullString
	Pinned              bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
	ClosedAt            sql.NullTime
}

// Tome status constants
const (
	TomeStatusOpen   = "open"
	TomeStatusClosed = "closed"
)
