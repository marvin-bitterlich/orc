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

// Shipment status constants
const (
	ShipmentStatusActive   = "active"
	ShipmentStatusPaused   = "paused"
	ShipmentStatusComplete = "complete"
)
