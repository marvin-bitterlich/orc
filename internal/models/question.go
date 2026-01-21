package models

import (
	"database/sql"
	"time"
)

// Question represents a question in the ORC ledger.
// Status can be: open, answered
type Question struct {
	ID               string
	InvestigationID  sql.NullString
	MissionID        string
	Title            string
	Description      sql.NullString
	Status           string
	Answer           sql.NullString
	Pinned           bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	AnsweredAt       sql.NullTime
	ConclaveID       sql.NullString
	PromotedFromID   sql.NullString
	PromotedFromType sql.NullString
}

// Question status constants
const (
	QuestionStatusOpen     = "open"
	QuestionStatusAnswered = "answered"
)
