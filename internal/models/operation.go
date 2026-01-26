package models

import (
	"database/sql"
	"time"
)

// Operation represents a unit of work within a commission.
// All database operations are handled by OperationRepository in the adapters layer.
type Operation struct {
	ID           string
	CommissionID string
	Title        string
	Description  sql.NullString
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CompletedAt  sql.NullTime
}

// Operation status constants
const (
	OperationStatusReady    = "ready"
	OperationStatusComplete = "complete"
)
