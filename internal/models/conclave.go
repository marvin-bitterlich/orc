package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/db"
)

type Conclave struct {
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

// CreateConclave creates a new conclave (ideation session)
func CreateConclave(missionID, title, description string) (*Conclave, error) {
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

	// Generate conclave ID by finding max existing ID
	var maxID int
	err = database.QueryRow("SELECT COALESCE(MAX(CAST(SUBSTR(id, 5) AS INTEGER)), 0) FROM conclaves").Scan(&maxID)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("CON-%03d", maxID+1)

	var desc sql.NullString
	if description != "" {
		desc = sql.NullString{String: description, Valid: true}
	}

	_, err = database.Exec(
		"INSERT INTO conclaves (id, mission_id, title, description, status) VALUES (?, ?, ?, ?, ?)",
		id, missionID, title, desc, "active",
	)
	if err != nil {
		return nil, err
	}

	return GetConclave(id)
}

// GetConclave retrieves a conclave by ID
func GetConclave(id string) (*Conclave, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	c := &Conclave{}
	err = database.QueryRow(
		"SELECT id, mission_id, title, description, status, assigned_grove_id, pinned, created_at, updated_at, completed_at FROM conclaves WHERE id = ?",
		id,
	).Scan(&c.ID, &c.MissionID, &c.Title, &c.Description, &c.Status, &c.AssignedGroveID, &c.Pinned, &c.CreatedAt, &c.UpdatedAt, &c.CompletedAt)

	if err != nil {
		return nil, err
	}

	return c, nil
}

// ListConclaves retrieves conclaves, optionally filtered by mission and/or status
func ListConclaves(missionID, status string) ([]*Conclave, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, mission_id, title, description, status, assigned_grove_id, pinned, created_at, updated_at, completed_at FROM conclaves WHERE 1=1"
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

	var conclaves []*Conclave
	for rows.Next() {
		c := &Conclave{}
		err := rows.Scan(&c.ID, &c.MissionID, &c.Title, &c.Description, &c.Status, &c.AssignedGroveID, &c.Pinned, &c.CreatedAt, &c.UpdatedAt, &c.CompletedAt)
		if err != nil {
			return nil, err
		}
		conclaves = append(conclaves, c)
	}

	return conclaves, nil
}

// CompleteConclave marks a conclave as complete
func CompleteConclave(id string) error {
	c, err := GetConclave(id)
	if err != nil {
		return err
	}

	if c.Pinned {
		return fmt.Errorf("cannot complete pinned conclave %s. Unpin first with: orc conclave unpin %s", id, id)
	}

	database, err := db.GetDB()
	if err != nil {
		return err
	}

	_, err = database.Exec(
		"UPDATE conclaves SET status = 'complete', completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// UpdateConclave updates the title and/or description of a conclave
func UpdateConclave(id, title, description string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM conclaves WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("conclave %s not found", id)
	}

	if title != "" && description != "" {
		_, err = database.Exec(
			"UPDATE conclaves SET title = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			title, description, id,
		)
	} else if title != "" {
		_, err = database.Exec(
			"UPDATE conclaves SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			title, id,
		)
	} else if description != "" {
		_, err = database.Exec(
			"UPDATE conclaves SET description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			description, id,
		)
	}

	return err
}

// PinConclave pins a conclave
func PinConclave(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM conclaves WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("conclave %s not found", id)
	}

	_, err = database.Exec(
		"UPDATE conclaves SET pinned = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// UnpinConclave unpins a conclave
func UnpinConclave(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM conclaves WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("conclave %s not found", id)
	}

	_, err = database.Exec(
		"UPDATE conclaves SET pinned = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// DeleteConclave deletes a conclave by ID
func DeleteConclave(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM conclaves WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("conclave %s not found", id)
	}

	_, err = database.Exec("DELETE FROM conclaves WHERE id = ?", id)
	return err
}

// GetConclavesByGrove returns conclaves assigned to a specific grove
func GetConclavesByGrove(groveID string) ([]*Conclave, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, mission_id, title, description, status, assigned_grove_id, pinned, created_at, updated_at, completed_at FROM conclaves WHERE assigned_grove_id = ?"
	rows, err := database.Query(query, groveID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conclaves []*Conclave
	for rows.Next() {
		c := &Conclave{}
		err := rows.Scan(&c.ID, &c.MissionID, &c.Title, &c.Description, &c.Status, &c.AssignedGroveID, &c.Pinned, &c.CreatedAt, &c.UpdatedAt, &c.CompletedAt)
		if err != nil {
			return nil, err
		}
		conclaves = append(conclaves, c)
	}

	return conclaves, nil
}

// GetConclaveTasks gets all tasks associated with a conclave
func GetConclaveTasks(conclaveID string) ([]*Task, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, shipment_id, mission_id, title, description, type, status, priority, assigned_grove_id, pinned, created_at, updated_at, claimed_at, completed_at, conclave_id, promoted_from_id, promoted_from_type FROM tasks WHERE conclave_id = ? ORDER BY created_at ASC"
	rows, err := database.Query(query, conclaveID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		err := rows.Scan(&task.ID, &task.ShipmentID, &task.MissionID, &task.Title, &task.Description, &task.Type, &task.Status, &task.Priority, &task.AssignedGroveID, &task.Pinned, &task.CreatedAt, &task.UpdatedAt, &task.ClaimedAt, &task.CompletedAt, &task.ConclaveID, &task.PromotedFromID, &task.PromotedFromType)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetConclaveQuestions gets all questions associated with a conclave
func GetConclaveQuestions(conclaveID string) ([]*Question, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, investigation_id, mission_id, title, description, status, answer, pinned, created_at, updated_at, answered_at, conclave_id, promoted_from_id, promoted_from_type FROM questions WHERE conclave_id = ? ORDER BY created_at ASC"
	rows, err := database.Query(query, conclaveID)
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

// GetConclavePlans gets all plans associated with a conclave
func GetConclavePlans(conclaveID string) ([]*Plan, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, shipment_id, mission_id, title, description, status, content, pinned, created_at, updated_at, approved_at, conclave_id, promoted_from_id, promoted_from_type FROM plans WHERE conclave_id = ? ORDER BY created_at ASC"
	rows, err := database.Query(query, conclaveID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*Plan
	for rows.Next() {
		p := &Plan{}
		err := rows.Scan(&p.ID, &p.ShipmentID, &p.MissionID, &p.Title, &p.Description, &p.Status, &p.Content, &p.Pinned, &p.CreatedAt, &p.UpdatedAt, &p.ApprovedAt, &p.ConclaveID, &p.PromotedFromID, &p.PromotedFromType)
		if err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}

	return plans, nil
}
