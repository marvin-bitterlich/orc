// Package models contains domain types for ORC entities.
// SQL persistence has been moved to internal/adapters/sqlite/*.go
package models

import (
	"database/sql"
	"time"
)

// Grove represents a grove entity.
// This is the domain type used within the models package.
// For persistence, use the repository interfaces in ports/secondary.
type Grove struct {
	ID        string
	MissionID string
	Name      string
	Path      string
	Repos     sql.NullString
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Grove status constants
const (
	GroveStatusActive   = "active"
	GroveStatusArchived = "archived"
)
