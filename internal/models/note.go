package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/db"
)

type Note struct {
	ID               string
	MissionID        string
	Title            string
	Content          sql.NullString
	Type             sql.NullString // learning, concern, finding, frq, bug, investigation_report
	ShipmentID       sql.NullString
	InvestigationID  sql.NullString
	ConclaveID       sql.NullString
	TomeID           sql.NullString
	Pinned           bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	PromotedFromID   sql.NullString
	PromotedFromType sql.NullString
}

// Note types
const (
	NoteTypeLearning            = "learning"
	NoteTypeConcern             = "concern"
	NoteTypeFinding             = "finding"
	NoteTypeFRQ                 = "frq"
	NoteTypeBug                 = "bug"
	NoteTypeInvestigationReport = "investigation_report"
)

// CreateNote creates a new note
func CreateNote(missionID, title, content, noteType, containerID, containerType string) (*Note, error) {
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

	// Generate note ID by finding max existing ID
	var maxID int
	err = database.QueryRow("SELECT COALESCE(MAX(CAST(SUBSTR(id, 6) AS INTEGER)), 0) FROM notes").Scan(&maxID)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("NOTE-%03d", maxID+1)

	var cont sql.NullString
	if content != "" {
		cont = sql.NullString{String: content, Valid: true}
	}

	var typeVal sql.NullString
	if noteType != "" {
		typeVal = sql.NullString{String: noteType, Valid: true}
	}

	// Set appropriate container FK based on container type
	var shipmentID, investigationID, conclaveID, tomeID sql.NullString
	if containerID != "" {
		switch containerType {
		case "shipment":
			shipmentID = sql.NullString{String: containerID, Valid: true}
		case "investigation":
			investigationID = sql.NullString{String: containerID, Valid: true}
		case "conclave":
			conclaveID = sql.NullString{String: containerID, Valid: true}
		case "tome":
			tomeID = sql.NullString{String: containerID, Valid: true}
		}
	}

	_, err = database.Exec(
		"INSERT INTO notes (id, mission_id, title, content, type, shipment_id, investigation_id, conclave_id, tome_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		id, missionID, title, cont, typeVal, shipmentID, investigationID, conclaveID, tomeID,
	)
	if err != nil {
		return nil, err
	}

	return GetNote(id)
}

// GetNote retrieves a note by ID
func GetNote(id string) (*Note, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	n := &Note{}
	err = database.QueryRow(
		"SELECT id, mission_id, title, content, type, shipment_id, investigation_id, conclave_id, tome_id, pinned, created_at, updated_at, promoted_from_id, promoted_from_type FROM notes WHERE id = ?",
		id,
	).Scan(&n.ID, &n.MissionID, &n.Title, &n.Content, &n.Type, &n.ShipmentID, &n.InvestigationID, &n.ConclaveID, &n.TomeID, &n.Pinned, &n.CreatedAt, &n.UpdatedAt, &n.PromotedFromID, &n.PromotedFromType)

	if err != nil {
		return nil, err
	}

	return n, nil
}

// ListNotes retrieves notes, optionally filtered by type, container, etc.
func ListNotes(noteType, missionID string) ([]*Note, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, mission_id, title, content, type, shipment_id, investigation_id, conclave_id, tome_id, pinned, created_at, updated_at, promoted_from_id, promoted_from_type FROM notes WHERE 1=1"
	args := []any{}

	if noteType != "" {
		query += " AND type = ?"
		args = append(args, noteType)
	}

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

	var notes []*Note
	for rows.Next() {
		n := &Note{}
		err := rows.Scan(&n.ID, &n.MissionID, &n.Title, &n.Content, &n.Type, &n.ShipmentID, &n.InvestigationID, &n.ConclaveID, &n.TomeID, &n.Pinned, &n.CreatedAt, &n.UpdatedAt, &n.PromotedFromID, &n.PromotedFromType)
		if err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}

	return notes, nil
}

// UpdateNote updates the title and/or content of a note
func UpdateNote(id, title, content string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM notes WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("note %s not found", id)
	}

	if title != "" && content != "" {
		_, err = database.Exec(
			"UPDATE notes SET title = ?, content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			title, content, id,
		)
	} else if title != "" {
		_, err = database.Exec(
			"UPDATE notes SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			title, id,
		)
	} else if content != "" {
		_, err = database.Exec(
			"UPDATE notes SET content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			content, id,
		)
	}

	return err
}

// PinNote pins a note
func PinNote(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM notes WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("note %s not found", id)
	}

	_, err = database.Exec(
		"UPDATE notes SET pinned = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// UnpinNote unpins a note
func UnpinNote(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM notes WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("note %s not found", id)
	}

	_, err = database.Exec(
		"UPDATE notes SET pinned = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// DeleteNote deletes a note by ID
func DeleteNote(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM notes WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("note %s not found", id)
	}

	_, err = database.Exec("DELETE FROM notes WHERE id = ?", id)
	return err
}

// GetNotesByContainer retrieves notes for a specific container
func GetNotesByContainer(containerType, containerID string) ([]*Note, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	var query string
	switch containerType {
	case "shipment":
		query = "SELECT id, mission_id, title, content, type, shipment_id, investigation_id, conclave_id, tome_id, pinned, created_at, updated_at, promoted_from_id, promoted_from_type FROM notes WHERE shipment_id = ? ORDER BY created_at DESC"
	case "investigation":
		query = "SELECT id, mission_id, title, content, type, shipment_id, investigation_id, conclave_id, tome_id, pinned, created_at, updated_at, promoted_from_id, promoted_from_type FROM notes WHERE investigation_id = ? ORDER BY created_at DESC"
	case "conclave":
		query = "SELECT id, mission_id, title, content, type, shipment_id, investigation_id, conclave_id, tome_id, pinned, created_at, updated_at, promoted_from_id, promoted_from_type FROM notes WHERE conclave_id = ? ORDER BY created_at DESC"
	case "tome":
		query = "SELECT id, mission_id, title, content, type, shipment_id, investigation_id, conclave_id, tome_id, pinned, created_at, updated_at, promoted_from_id, promoted_from_type FROM notes WHERE tome_id = ? ORDER BY created_at DESC"
	default:
		return nil, fmt.Errorf("unknown container type: %s", containerType)
	}

	rows, err := database.Query(query, containerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*Note
	for rows.Next() {
		n := &Note{}
		err := rows.Scan(&n.ID, &n.MissionID, &n.Title, &n.Content, &n.Type, &n.ShipmentID, &n.InvestigationID, &n.ConclaveID, &n.TomeID, &n.Pinned, &n.CreatedAt, &n.UpdatedAt, &n.PromotedFromID, &n.PromotedFromType)
		if err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}

	return notes, nil
}
