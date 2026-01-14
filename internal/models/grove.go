package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/looneym/orc/internal/db"
)

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

// CreateGrove creates a new grove
func CreateGrove(missionID, name, path string, repos []string) (*Grove, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	// Verify mission exists
	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM missions WHERE id = ?", missionID).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists == 0 {
		return nil, fmt.Errorf("mission %s not found", missionID)
	}

	// Generate grove ID
	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM groves").Scan(&count)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("GROVE-%03d", count+1)

	// Convert repos array to JSON
	var reposJSON sql.NullString
	if len(repos) > 0 {
		// Simple JSON array format
		reposStr := "["
		for i, repo := range repos {
			if i > 0 {
				reposStr += ","
			}
			reposStr += fmt.Sprintf(`"%s"`, repo)
		}
		reposStr += "]"
		reposJSON = sql.NullString{String: reposStr, Valid: true}
	}

	_, err = database.Exec(
		"INSERT INTO groves (id, mission_id, name, path, repos, status) VALUES (?, ?, ?, ?, ?, ?)",
		id, missionID, name, path, reposJSON, "active",
	)
	if err != nil {
		return nil, err
	}

	return GetGrove(id)
}

// GetGrove retrieves a grove by ID
func GetGrove(id string) (*Grove, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	grove := &Grove{}
	err = database.QueryRow(
		"SELECT id, mission_id, name, path, repos, status, created_at, updated_at FROM groves WHERE id = ?",
		id,
	).Scan(&grove.ID, &grove.MissionID, &grove.Name, &grove.Path, &grove.Repos, &grove.Status, &grove.CreatedAt, &grove.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return grove, nil
}

// ListGroves retrieves all groves, optionally filtered by mission
func ListGroves(missionID string) ([]*Grove, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, mission_id, name, path, repos, status, created_at, updated_at FROM groves WHERE 1=1"
	args := []interface{}{}

	if missionID != "" {
		query += " AND mission_id = ?"
		args = append(args, missionID)
	}

	query += " ORDER BY created_at DESC"

	rows, err := database.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groves []*Grove
	for rows.Next() {
		grove := &Grove{}
		err := rows.Scan(&grove.ID, &grove.MissionID, &grove.Name, &grove.Path, &grove.Repos, &grove.Status, &grove.CreatedAt, &grove.UpdatedAt)
		if err != nil {
			return nil, err
		}
		groves = append(groves, grove)
	}

	return groves, nil
}

// GetGrovesByMission retrieves all active groves for a mission
func GetGrovesByMission(missionID string) ([]*Grove, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	rows, err := database.Query(
		"SELECT id, mission_id, name, path, repos, status, created_at, updated_at FROM groves WHERE mission_id = ? AND status = 'active' ORDER BY created_at DESC",
		missionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groves []*Grove
	for rows.Next() {
		grove := &Grove{}
		err := rows.Scan(&grove.ID, &grove.MissionID, &grove.Name, &grove.Path, &grove.Repos, &grove.Status, &grove.CreatedAt, &grove.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning grove: %w", err)
		}
		groves = append(groves, grove)
	}

	return groves, nil
}

// RenameGrove updates the name of a grove
func RenameGrove(id, newName string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	// Verify grove exists
	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM groves WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("grove %s not found", id)
	}

	_, err = database.Exec(
		"UPDATE groves SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		newName, id,
	)

	return err
}
