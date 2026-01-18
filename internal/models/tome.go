package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/db"
)

type Tome struct {
	ID          string
	MissionID   string
	Title       string
	Description sql.NullString
	Status      string
	Pinned      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt sql.NullTime
}

// CreateTome creates a new tome (organization container)
func CreateTome(missionID, title, description string) (*Tome, error) {
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

	// Generate tome ID by finding max existing ID
	var maxID int
	err = database.QueryRow("SELECT COALESCE(MAX(CAST(SUBSTR(id, 6) AS INTEGER)), 0) FROM tomes").Scan(&maxID)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("TOME-%03d", maxID+1)

	var desc sql.NullString
	if description != "" {
		desc = sql.NullString{String: description, Valid: true}
	}

	_, err = database.Exec(
		"INSERT INTO tomes (id, mission_id, title, description, status) VALUES (?, ?, ?, ?, ?)",
		id, missionID, title, desc, "active",
	)
	if err != nil {
		return nil, err
	}

	return GetTome(id)
}

// GetTome retrieves a tome by ID
func GetTome(id string) (*Tome, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	t := &Tome{}
	err = database.QueryRow(
		"SELECT id, mission_id, title, description, status, pinned, created_at, updated_at, completed_at FROM tomes WHERE id = ?",
		id,
	).Scan(&t.ID, &t.MissionID, &t.Title, &t.Description, &t.Status, &t.Pinned, &t.CreatedAt, &t.UpdatedAt, &t.CompletedAt)

	if err != nil {
		return nil, err
	}

	return t, nil
}

// ListTomes retrieves tomes, optionally filtered by mission and/or status
func ListTomes(missionID, status string) ([]*Tome, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, mission_id, title, description, status, pinned, created_at, updated_at, completed_at FROM tomes WHERE 1=1"
	args := []any{}

	if missionID != "" {
		query += " AND mission_id = ?"
		args = append(args, missionID)
	}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := database.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tomes []*Tome
	for rows.Next() {
		t := &Tome{}
		err := rows.Scan(&t.ID, &t.MissionID, &t.Title, &t.Description, &t.Status, &t.Pinned, &t.CreatedAt, &t.UpdatedAt, &t.CompletedAt)
		if err != nil {
			return nil, err
		}
		tomes = append(tomes, t)
	}

	return tomes, nil
}

// CompleteTome marks a tome as complete
func CompleteTome(id string) error {
	t, err := GetTome(id)
	if err != nil {
		return err
	}

	if t.Pinned {
		return fmt.Errorf("cannot complete pinned tome %s. Unpin first with: orc tome unpin %s", id, id)
	}

	database, err := db.GetDB()
	if err != nil {
		return err
	}

	_, err = database.Exec(
		"UPDATE tomes SET status = 'complete', completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// UpdateTome updates the title and/or description of a tome
func UpdateTome(id, title, description string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM tomes WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("tome %s not found", id)
	}

	if title != "" && description != "" {
		_, err = database.Exec(
			"UPDATE tomes SET title = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			title, description, id,
		)
	} else if title != "" {
		_, err = database.Exec(
			"UPDATE tomes SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			title, id,
		)
	} else if description != "" {
		_, err = database.Exec(
			"UPDATE tomes SET description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			description, id,
		)
	}

	return err
}

// PinTome pins a tome
func PinTome(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM tomes WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("tome %s not found", id)
	}

	_, err = database.Exec(
		"UPDATE tomes SET pinned = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// UnpinTome unpins a tome
func UnpinTome(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM tomes WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("tome %s not found", id)
	}

	_, err = database.Exec(
		"UPDATE tomes SET pinned = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// DeleteTome deletes a tome by ID
func DeleteTome(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM tomes WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("tome %s not found", id)
	}

	_, err = database.Exec("DELETE FROM tomes WHERE id = ?", id)
	return err
}

// GetTomeNotes gets all notes in a tome
func GetTomeNotes(tomeID string) ([]*Note, error) {
	return GetNotesByContainer("tome", tomeID)
}
