package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/db"
)

type Handoff struct {
	ID              string
	CreatedAt       time.Time
	HandoffNote     string
	ActiveMissionID sql.NullString
	ActiveGroveID   sql.NullString
	TodosSnapshot   sql.NullString
}

// CreateHandoff creates a new handoff with a narrative note
func CreateHandoff(note string, activeMissionID string, todosJSON, activeGroveID string) (*Handoff, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	// Generate handoff ID by finding max existing ID
	var maxID int
	err = database.QueryRow("SELECT COALESCE(MAX(CAST(SUBSTR(id, 4) AS INTEGER)), 0) FROM handoffs").Scan(&maxID)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("HO-%03d", maxID+1)

	// Handle nullable strings
	var missionID, todos, groveID sql.NullString

	if activeMissionID != "" {
		missionID = sql.NullString{String: activeMissionID, Valid: true}
	}

	if todosJSON != "" {
		todos = sql.NullString{String: todosJSON, Valid: true}
	}
	if activeGroveID != "" {
		groveID = sql.NullString{String: activeGroveID, Valid: true}
	}

	_, err = database.Exec(
		`INSERT INTO handoffs (id, handoff_note, active_mission_id, active_grove_id, todos_snapshot)
		 VALUES (?, ?, ?, ?, ?)`,
		id, note, missionID, groveID, todos,
	)
	if err != nil {
		return nil, err
	}

	return GetHandoff(id)
}

// GetHandoff retrieves a handoff by ID
func GetHandoff(id string) (*Handoff, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	h := &Handoff{}
	err = database.QueryRow(
		`SELECT id, created_at, handoff_note, active_mission_id, active_grove_id, todos_snapshot
		 FROM handoffs WHERE id = ?`,
		id,
	).Scan(&h.ID, &h.CreatedAt, &h.HandoffNote, &h.ActiveMissionID, &h.ActiveGroveID, &h.TodosSnapshot)

	if err != nil {
		return nil, err
	}

	return h, nil
}

// GetLatestHandoff retrieves the most recent handoff
func GetLatestHandoff() (*Handoff, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	h := &Handoff{}
	err = database.QueryRow(
		`SELECT id, created_at, handoff_note, active_mission_id, active_grove_id, todos_snapshot
		 FROM handoffs ORDER BY created_at DESC LIMIT 1`,
	).Scan(&h.ID, &h.CreatedAt, &h.HandoffNote, &h.ActiveMissionID, &h.ActiveGroveID, &h.TodosSnapshot)

	if err != nil {
		return nil, err
	}

	return h, nil
}

// GetLatestHandoffForGrove retrieves the most recent handoff for a specific grove
func GetLatestHandoffForGrove(groveID string) (*Handoff, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	h := &Handoff{}
	err = database.QueryRow(
		`SELECT id, created_at, handoff_note, active_mission_id, active_grove_id, todos_snapshot
		 FROM handoffs WHERE active_grove_id = ? ORDER BY created_at DESC LIMIT 1`,
		groveID,
	).Scan(&h.ID, &h.CreatedAt, &h.HandoffNote, &h.ActiveMissionID, &h.ActiveGroveID, &h.TodosSnapshot)

	if err != nil {
		return nil, err
	}

	return h, nil
}

// ListHandoffs retrieves all handoffs ordered by creation date
func ListHandoffs(limit int) ([]*Handoff, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := `SELECT id, created_at, handoff_note, active_mission_id, active_grove_id, todos_snapshot
	          FROM handoffs ORDER BY created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := database.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var handoffs []*Handoff
	for rows.Next() {
		h := &Handoff{}
		err := rows.Scan(&h.ID, &h.CreatedAt, &h.HandoffNote, &h.ActiveMissionID, &h.ActiveGroveID, &h.TodosSnapshot)
		if err != nil {
			return nil, err
		}
		handoffs = append(handoffs, h)
	}

	return handoffs, nil
}
