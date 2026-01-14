package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/looneym/orc/internal/db"
)

type Mission struct {
	ID          string
	Title       string
	Description sql.NullString
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt sql.NullTime
}

// CreateMission creates a new mission
func CreateMission(title, description string) (*Mission, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	// Generate mission ID
	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM missions").Scan(&count)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("MISSION-%03d", count+1)

	var desc sql.NullString
	if description != "" {
		desc = sql.NullString{String: description, Valid: true}
	}

	_, err = database.Exec(
		"INSERT INTO missions (id, title, description, status) VALUES (?, ?, ?, ?)",
		id, title, desc, "active",
	)
	if err != nil {
		return nil, err
	}

	return GetMission(id)
}

// GetMission retrieves a mission by ID
func GetMission(id string) (*Mission, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	m := &Mission{}
	err = database.QueryRow(
		"SELECT id, title, description, status, created_at, updated_at, completed_at FROM missions WHERE id = ?",
		id,
	).Scan(&m.ID, &m.Title, &m.Description, &m.Status, &m.CreatedAt, &m.UpdatedAt, &m.CompletedAt)

	if err != nil {
		return nil, err
	}

	return m, nil
}

// ListMissions retrieves all missions, optionally filtered by status
func ListMissions(status string) ([]*Mission, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, title, description, status, created_at, updated_at, completed_at FROM missions"
	args := []interface{}{}

	if status != "" {
		query += " WHERE status = ?"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := database.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var missions []*Mission
	for rows.Next() {
		m := &Mission{}
		err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.Status, &m.CreatedAt, &m.UpdatedAt, &m.CompletedAt)
		if err != nil {
			return nil, err
		}
		missions = append(missions, m)
	}

	return missions, nil
}

// UpdateMissionStatus updates the status of a mission
func UpdateMissionStatus(id, status string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var completedAt sql.NullTime
	if status == "complete" {
		completedAt = sql.NullTime{Time: time.Now(), Valid: true}
	}

	_, err = database.Exec(
		"UPDATE missions SET status = ?, updated_at = CURRENT_TIMESTAMP, completed_at = ? WHERE id = ?",
		status, completedAt, id,
	)

	return err
}

// UpdateMission updates the title and/or description of a mission
func UpdateMission(id, title, description string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	// Build dynamic query based on what's being updated
	query := "UPDATE missions SET updated_at = CURRENT_TIMESTAMP"
	args := []interface{}{}

	if title != "" {
		query += ", title = ?"
		args = append(args, title)
	}

	if description != "" {
		query += ", description = ?"
		args = append(args, sql.NullString{String: description, Valid: true})
	}

	query += " WHERE id = ?"
	args = append(args, id)

	_, err = database.Exec(query, args...)
	return err
}

// DeleteMission deletes a mission and all associated data
func DeleteMission(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	// TODO: Add cascade delete for work_orders, groves, handoffs when implemented

	_, err = database.Exec("DELETE FROM missions WHERE id = ?", id)
	return err
}
