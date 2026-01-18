package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/db"
)

type Investigation struct {
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

// CreateInvestigation creates a new investigation
func CreateInvestigation(missionID, title, description string) (*Investigation, error) {
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

	// Generate investigation ID by finding max existing ID
	var maxID int
	err = database.QueryRow("SELECT COALESCE(MAX(CAST(SUBSTR(id, 5) AS INTEGER)), 0) FROM investigations").Scan(&maxID)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("INV-%03d", maxID+1)

	var desc sql.NullString
	if description != "" {
		desc = sql.NullString{String: description, Valid: true}
	}

	_, err = database.Exec(
		"INSERT INTO investigations (id, mission_id, title, description, status) VALUES (?, ?, ?, ?, ?)",
		id, missionID, title, desc, "active",
	)
	if err != nil {
		return nil, err
	}

	return GetInvestigation(id)
}

// GetInvestigation retrieves an investigation by ID
func GetInvestigation(id string) (*Investigation, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	inv := &Investigation{}
	err = database.QueryRow(
		"SELECT id, mission_id, title, description, status, assigned_grove_id, pinned, created_at, updated_at, completed_at FROM investigations WHERE id = ?",
		id,
	).Scan(&inv.ID, &inv.MissionID, &inv.Title, &inv.Description, &inv.Status, &inv.AssignedGroveID, &inv.Pinned, &inv.CreatedAt, &inv.UpdatedAt, &inv.CompletedAt)

	if err != nil {
		return nil, err
	}

	return inv, nil
}

// ListInvestigations retrieves investigations, optionally filtered by mission and/or status
func ListInvestigations(missionID, status string) ([]*Investigation, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, mission_id, title, description, status, assigned_grove_id, pinned, created_at, updated_at, completed_at FROM investigations WHERE 1=1"
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

	var investigations []*Investigation
	for rows.Next() {
		inv := &Investigation{}
		err := rows.Scan(&inv.ID, &inv.MissionID, &inv.Title, &inv.Description, &inv.Status, &inv.AssignedGroveID, &inv.Pinned, &inv.CreatedAt, &inv.UpdatedAt, &inv.CompletedAt)
		if err != nil {
			return nil, err
		}
		investigations = append(investigations, inv)
	}

	return investigations, nil
}

// CompleteInvestigation marks an investigation as complete
func CompleteInvestigation(id string) error {
	inv, err := GetInvestigation(id)
	if err != nil {
		return err
	}

	if inv.Pinned {
		return fmt.Errorf("cannot complete pinned investigation %s. Unpin first with: orc investigation unpin %s", id, id)
	}

	database, err := db.GetDB()
	if err != nil {
		return err
	}

	_, err = database.Exec(
		"UPDATE investigations SET status = 'complete', completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// UpdateInvestigation updates the title and/or description of an investigation
func UpdateInvestigation(id, title, description string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM investigations WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("investigation %s not found", id)
	}

	if title != "" && description != "" {
		_, err = database.Exec(
			"UPDATE investigations SET title = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			title, description, id,
		)
	} else if title != "" {
		_, err = database.Exec(
			"UPDATE investigations SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			title, id,
		)
	} else if description != "" {
		_, err = database.Exec(
			"UPDATE investigations SET description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			description, id,
		)
	}

	return err
}

// PinInvestigation pins an investigation
func PinInvestigation(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM investigations WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("investigation %s not found", id)
	}

	_, err = database.Exec(
		"UPDATE investigations SET pinned = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// UnpinInvestigation unpins an investigation
func UnpinInvestigation(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM investigations WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("investigation %s not found", id)
	}

	_, err = database.Exec(
		"UPDATE investigations SET pinned = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// AssignInvestigationToGrove assigns an investigation to a grove
func AssignInvestigationToGrove(investigationID, groveID string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM investigations WHERE id = ?", investigationID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("investigation %s not found", investigationID)
	}

	_, err = database.Exec(
		"UPDATE investigations SET assigned_grove_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		groveID, investigationID,
	)

	return err
}

// DeleteInvestigation deletes an investigation by ID
func DeleteInvestigation(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM investigations WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("investigation %s not found", id)
	}

	_, err = database.Exec("DELETE FROM investigations WHERE id = ?", id)
	return err
}

// GetInvestigationsByGrove returns investigations assigned to a specific grove
func GetInvestigationsByGrove(groveID string) ([]*Investigation, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, mission_id, title, description, status, assigned_grove_id, pinned, created_at, updated_at, completed_at FROM investigations WHERE assigned_grove_id = ?"
	rows, err := database.Query(query, groveID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var investigations []*Investigation
	for rows.Next() {
		inv := &Investigation{}
		err := rows.Scan(&inv.ID, &inv.MissionID, &inv.Title, &inv.Description, &inv.Status, &inv.AssignedGroveID, &inv.Pinned, &inv.CreatedAt, &inv.UpdatedAt, &inv.CompletedAt)
		if err != nil {
			return nil, err
		}
		investigations = append(investigations, inv)
	}

	return investigations, nil
}

// GetInvestigationQuestions gets all questions in an investigation
func GetInvestigationQuestions(investigationID string) ([]*Question, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, investigation_id, mission_id, title, description, status, answer, pinned, created_at, updated_at, answered_at, conclave_id, promoted_from_id, promoted_from_type FROM questions WHERE investigation_id = ? ORDER BY created_at ASC"
	rows, err := database.Query(query, investigationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []*Question
	for rows.Next() {
		q := &Question{}
		err := rows.Scan(&q.ID, &q.InvestigationID, &q.MissionID, &q.Title, &q.Description, &q.Status, &q.Answer, &q.Pinned, &q.CreatedAt, &q.UpdatedAt, &q.AnsweredAt, &q.ConclaveID, &q.PromotedFromID, &q.PromotedFromType)
		if err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}

	return questions, nil
}
