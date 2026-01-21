package models

import (
	"database/sql"
	"time"
)

type Tome struct {
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

// Tome status constants
const (
	TomeStatusActive   = "active"
	TomeStatusPaused   = "paused"
	TomeStatusComplete = "complete"
)
