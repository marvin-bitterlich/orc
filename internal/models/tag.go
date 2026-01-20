package models

import (
	"database/sql"
	"time"
)

// Tag represents a classification label
type Tag struct {
	ID          string
	Name        string
	Description sql.NullString
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
