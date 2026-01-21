package models

import (
	"database/sql"
	"time"
)

// Plan represents an implementation strategy in the ORC ledger.
// Status can be: draft, approved
type Plan struct {
	ID               string
	ShipmentID       sql.NullString
	MissionID        string
	Title            string
	Description      sql.NullString
	Status           string
	Content          sql.NullString
	Pinned           bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ApprovedAt       sql.NullTime
	ConclaveID       sql.NullString
	PromotedFromID   sql.NullString
	PromotedFromType sql.NullString
}

// Plan status constants
const (
	PlanStatusDraft    = "draft"
	PlanStatusApproved = "approved"
)
