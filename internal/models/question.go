package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/db"
)

type Question struct {
	ID               string
	InvestigationID  sql.NullString
	MissionID        string
	Title            string
	Description      sql.NullString
	Status           string // open, answered
	Answer           sql.NullString
	Pinned           bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	AnsweredAt       sql.NullTime
	ConclaveID       sql.NullString
	PromotedFromID   sql.NullString
	PromotedFromType sql.NullString
}

// CreateQuestion creates a new question under an investigation
func CreateQuestion(investigationID, missionID, title, description string) (*Question, error) {
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

	// If investigation ID specified, verify it exists
	if investigationID != "" {
		err = database.QueryRow("SELECT COUNT(*) FROM investigations WHERE id = ?", investigationID).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if exists == 0 {
			return nil, fmt.Errorf("investigation %s not found", investigationID)
		}
	}

	// Generate question ID by finding max existing ID
	var maxID int
	err = database.QueryRow("SELECT COALESCE(MAX(CAST(SUBSTR(id, 3) AS INTEGER)), 0) FROM questions").Scan(&maxID)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("Q-%03d", maxID+1)

	var desc sql.NullString
	if description != "" {
		desc = sql.NullString{String: description, Valid: true}
	}

	var invIDVal sql.NullString
	if investigationID != "" {
		invIDVal = sql.NullString{String: investigationID, Valid: true}
	}

	_, err = database.Exec(
		"INSERT INTO questions (id, investigation_id, mission_id, title, description, status) VALUES (?, ?, ?, ?, ?, ?)",
		id, invIDVal, missionID, title, desc, "open",
	)
	if err != nil {
		return nil, err
	}

	return GetQuestion(id)
}

// GetQuestion retrieves a question by ID
func GetQuestion(id string) (*Question, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	q := &Question{}
	err = database.QueryRow(
		"SELECT id, investigation_id, mission_id, title, description, status, answer, pinned, created_at, updated_at, answered_at, conclave_id, promoted_from_id, promoted_from_type FROM questions WHERE id = ?",
		id,
	).Scan(&q.ID, &q.InvestigationID, &q.MissionID, &q.Title, &q.Description, &q.Status, &q.Answer, &q.Pinned, &q.CreatedAt, &q.UpdatedAt, &q.AnsweredAt, &q.ConclaveID, &q.PromotedFromID, &q.PromotedFromType)

	if err != nil {
		return nil, err
	}

	return q, nil
}

// ListQuestions retrieves questions, optionally filtered by investigation/status
func ListQuestions(investigationID, status string) ([]*Question, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, investigation_id, mission_id, title, description, status, answer, pinned, created_at, updated_at, answered_at, conclave_id, promoted_from_id, promoted_from_type FROM questions WHERE 1=1"
	args := []any{}

	if investigationID != "" {
		query += " AND investigation_id = ?"
		args = append(args, investigationID)
	}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	query += " ORDER BY created_at ASC"

	rows, err := database.Query(query, args...)
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

// AnswerQuestion marks a question as answered with the given answer
func AnswerQuestion(id, answer string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	_, err = database.Exec(
		"UPDATE questions SET status = 'answered', answer = ?, answered_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		answer, id,
	)

	return err
}

// UpdateQuestion updates the title and/or description of a question
func UpdateQuestion(id, title, description string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM questions WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("question %s not found", id)
	}

	if title != "" && description != "" {
		_, err = database.Exec(
			"UPDATE questions SET title = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			title, description, id,
		)
	} else if title != "" {
		_, err = database.Exec(
			"UPDATE questions SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			title, id,
		)
	} else if description != "" {
		_, err = database.Exec(
			"UPDATE questions SET description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			description, id,
		)
	}

	return err
}

// PinQuestion pins a question
func PinQuestion(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM questions WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("question %s not found", id)
	}

	_, err = database.Exec(
		"UPDATE questions SET pinned = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// UnpinQuestion unpins a question
func UnpinQuestion(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM questions WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("question %s not found", id)
	}

	_, err = database.Exec(
		"UPDATE questions SET pinned = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// DeleteQuestion deletes a question by ID
func DeleteQuestion(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM questions WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("question %s not found", id)
	}

	_, err = database.Exec("DELETE FROM questions WHERE id = ?", id)
	return err
}
