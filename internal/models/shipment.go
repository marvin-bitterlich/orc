// Package models contains domain types for ORC entities.
// SQL persistence has been moved to internal/adapters/sqlite/*.go
package models

import (
	"database/sql"
	"time"
)

// Shipment represents a shipment entity.
// This is the domain type used within the models package.
// For persistence, use the repository interfaces in ports/secondary.
type Shipment struct {
	ID                  string
	CommissionID        string
	Title               string
	Description         sql.NullString
	Status              string
	AssignedWorkbenchID sql.NullString
	Pinned              bool
	ContainerID         sql.NullString // CON-xxx or YARD-xxx
	ContainerType       sql.NullString // "conclave" or "shipyard"
	SpecNoteID          sql.NullString // NOTE-xxx - spec note that generated this shipment
	CreatedAt           time.Time
	UpdatedAt           time.Time
	CompletedAt         sql.NullTime
}

// Shipment status constants - work state lifecycle
const (
	ShipmentStatusDraft      = "draft"
	ShipmentStatusExploring  = "exploring"
	ShipmentStatusSpecced    = "specced"
	ShipmentStatusTasked     = "tasked"
	ShipmentStatusInProgress = "in_progress"
	ShipmentStatusComplete   = "complete"
)
