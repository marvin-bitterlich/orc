package models

import (
	"database/sql"
	"time"
)

type Note struct {
	ID               string
	CommissionID     string
	Title            string
	Content          sql.NullString
	Type             sql.NullString // learning, concern, finding, frq, bug
	Status           string         // open, closed
	ShipmentID       sql.NullString
	ConclaveID       sql.NullString
	TomeID           sql.NullString
	Pinned           bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ClosedAt         sql.NullTime
	PromotedFromID   sql.NullString
	PromotedFromType sql.NullString
}

// Note types
const (
	NoteTypeLearning = "learning"
	NoteTypeConcern  = "concern"
	NoteTypeFinding  = "finding"
	NoteTypeFRQ      = "frq"
	NoteTypeBug      = "bug"
)

// Note statuses
const (
	NoteStatusOpen   = "open"
	NoteStatusClosed = "closed"
)
