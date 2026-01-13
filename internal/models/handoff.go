package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/looneym/orc/internal/db"
)

type Handoff struct {
	ID                 string
	CreatedAt          time.Time
	HandoffNote        string
	ActiveMissionID    sql.NullString
	ActiveOperationID  sql.NullString
	ActiveWorkOrderID  sql.NullString
	ActiveExpeditionID sql.NullString
	TodosSnapshot      sql.NullString
	GraphitiEpisodeUUID sql.NullString
}

// CreateHandoff creates a new handoff with a narrative note
func CreateHandoff(note string, activeMissionID, activeOperationID, activeWorkOrderID, activeExpeditionID, todosJSON, graphitiUUID string) (*Handoff, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	// Generate handoff ID
	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM handoffs").Scan(&count)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("HO-%03d", count+1)

	// Handle nullable strings
	var missionID, operationID, workOrderID, expeditionID, todos, graphiti sql.NullString

	if activeMissionID != "" {
		missionID = sql.NullString{String: activeMissionID, Valid: true}
	}
	if activeOperationID != "" {
		operationID = sql.NullString{String: activeOperationID, Valid: true}
	}
	if activeWorkOrderID != "" {
		workOrderID = sql.NullString{String: activeWorkOrderID, Valid: true}
	}
	if activeExpeditionID != "" {
		expeditionID = sql.NullString{String: activeExpeditionID, Valid: true}
	}
	if todosJSON != "" {
		todos = sql.NullString{String: todosJSON, Valid: true}
	}
	if graphitiUUID != "" {
		graphiti = sql.NullString{String: graphitiUUID, Valid: true}
	}

	_, err = database.Exec(
		`INSERT INTO handoffs (id, handoff_note, active_mission_id, active_operation_id, active_work_order_id, active_expedition_id, todos_snapshot, graphiti_episode_uuid)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, note, missionID, operationID, workOrderID, expeditionID, todos, graphiti,
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
		`SELECT id, created_at, handoff_note, active_mission_id, active_operation_id, active_work_order_id, active_expedition_id, todos_snapshot, graphiti_episode_uuid
		 FROM handoffs WHERE id = ?`,
		id,
	).Scan(&h.ID, &h.CreatedAt, &h.HandoffNote, &h.ActiveMissionID, &h.ActiveOperationID, &h.ActiveWorkOrderID, &h.ActiveExpeditionID, &h.TodosSnapshot, &h.GraphitiEpisodeUUID)

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
		`SELECT id, created_at, handoff_note, active_mission_id, active_operation_id, active_work_order_id, active_expedition_id, todos_snapshot, graphiti_episode_uuid
		 FROM handoffs ORDER BY created_at DESC LIMIT 1`,
	).Scan(&h.ID, &h.CreatedAt, &h.HandoffNote, &h.ActiveMissionID, &h.ActiveOperationID, &h.ActiveWorkOrderID, &h.ActiveExpeditionID, &h.TodosSnapshot, &h.GraphitiEpisodeUUID)

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

	query := `SELECT id, created_at, handoff_note, active_mission_id, active_operation_id, active_work_order_id, active_expedition_id, todos_snapshot, graphiti_episode_uuid
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
		err := rows.Scan(&h.ID, &h.CreatedAt, &h.HandoffNote, &h.ActiveMissionID, &h.ActiveOperationID, &h.ActiveWorkOrderID, &h.ActiveExpeditionID, &h.TodosSnapshot, &h.GraphitiEpisodeUUID)
		if err != nil {
			return nil, err
		}
		handoffs = append(handoffs, h)
	}

	return handoffs, nil
}
