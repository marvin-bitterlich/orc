package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/db"
)

type Handoff struct {
	ID                  string
	CreatedAt           time.Time
	HandoffNote         string
	ActiveMissionID     sql.NullString
	ActiveWorkOrders    sql.NullString // JSON array of work order IDs
	ActiveGroveID       sql.NullString
	TodosSnapshot       sql.NullString
	GraphitiEpisodeUUID sql.NullString
}

// CreateHandoff creates a new handoff with a narrative note
func CreateHandoff(note string, activeMissionID string, activeWorkOrders []string, todosJSON, graphitiUUID, activeGroveID string) (*Handoff, error) {
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
	var missionID, workOrders, todos, graphiti, groveID sql.NullString

	if activeMissionID != "" {
		missionID = sql.NullString{String: activeMissionID, Valid: true}
	}

	// Convert work orders array to JSON
	if len(activeWorkOrders) > 0 {
		workOrdersStr := "["
		for i, wo := range activeWorkOrders {
			if i > 0 {
				workOrdersStr += ","
			}
			workOrdersStr += fmt.Sprintf(`"%s"`, wo)
		}
		workOrdersStr += "]"
		workOrders = sql.NullString{String: workOrdersStr, Valid: true}
	}

	if todosJSON != "" {
		todos = sql.NullString{String: todosJSON, Valid: true}
	}
	if graphitiUUID != "" {
		graphiti = sql.NullString{String: graphitiUUID, Valid: true}
	}
	if activeGroveID != "" {
		groveID = sql.NullString{String: activeGroveID, Valid: true}
	}

	_, err = database.Exec(
		`INSERT INTO handoffs (id, handoff_note, active_mission_id, active_work_orders, active_grove_id, todos_snapshot, graphiti_episode_uuid)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, note, missionID, workOrders, groveID, todos, graphiti,
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
		`SELECT id, created_at, handoff_note, active_mission_id, active_work_orders, active_grove_id, todos_snapshot, graphiti_episode_uuid
		 FROM handoffs WHERE id = ?`,
		id,
	).Scan(&h.ID, &h.CreatedAt, &h.HandoffNote, &h.ActiveMissionID, &h.ActiveWorkOrders, &h.ActiveGroveID, &h.TodosSnapshot, &h.GraphitiEpisodeUUID)

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
		`SELECT id, created_at, handoff_note, active_mission_id, active_work_orders, active_grove_id, todos_snapshot, graphiti_episode_uuid
		 FROM handoffs ORDER BY created_at DESC LIMIT 1`,
	).Scan(&h.ID, &h.CreatedAt, &h.HandoffNote, &h.ActiveMissionID, &h.ActiveWorkOrders, &h.ActiveGroveID, &h.TodosSnapshot, &h.GraphitiEpisodeUUID)

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
		`SELECT id, created_at, handoff_note, active_mission_id, active_work_orders, active_grove_id, todos_snapshot, graphiti_episode_uuid
		 FROM handoffs WHERE active_grove_id = ? ORDER BY created_at DESC LIMIT 1`,
		groveID,
	).Scan(&h.ID, &h.CreatedAt, &h.HandoffNote, &h.ActiveMissionID, &h.ActiveWorkOrders, &h.ActiveGroveID, &h.TodosSnapshot, &h.GraphitiEpisodeUUID)

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

	query := `SELECT id, created_at, handoff_note, active_mission_id, active_work_orders, active_grove_id, todos_snapshot, graphiti_episode_uuid
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
		err := rows.Scan(&h.ID, &h.CreatedAt, &h.HandoffNote, &h.ActiveMissionID, &h.ActiveWorkOrders, &h.ActiveGroveID, &h.TodosSnapshot, &h.GraphitiEpisodeUUID)
		if err != nil {
			return nil, err
		}
		handoffs = append(handoffs, h)
	}

	return handoffs, nil
}
