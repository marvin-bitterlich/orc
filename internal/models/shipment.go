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
	RepoID              sql.NullString // REPO-xxx - linked repository for branch ownership
	Branch              sql.NullString // Owned branch (e.g., ml/SHIP-001-feature-name)
	Pinned              bool
	SpecNoteID          sql.NullString // NOTE-xxx - spec note that generated this shipment
	CreatedAt           time.Time
	UpdatedAt           time.Time
	CompletedAt         sql.NullTime
}

// Shipment status constants - work state lifecycle
const (
	ShipmentStatusDraft            = "draft"
	ShipmentStatusExploring        = "exploring"
	ShipmentStatusSpecced          = "specced"
	ShipmentStatusTasked           = "tasked"
	ShipmentStatusReadyForImp      = "ready_for_imp"
	ShipmentStatusImplementing     = "implementing"
	ShipmentStatusAutoImplementing = "auto_implementing"
	ShipmentStatusComplete         = "complete"
)
