package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/looneym/orc/internal/db"
)

type WorkOrder struct {
	ID              string
	MissionID       string
	Title           string
	Description     sql.NullString
	Type            sql.NullString
	Status          string
	Priority        sql.NullString
	ParentID        sql.NullString
	AssignedGroveID sql.NullString
	ContextRef      sql.NullString
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ClaimedAt       sql.NullTime
	CompletedAt     sql.NullTime
}

// CreateWorkOrder creates a new work order
func CreateWorkOrder(missionID, title, description, contextRef string) (*WorkOrder, error) {
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

	// Generate work order ID
	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM work_orders").Scan(&count)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("WO-%03d", count+1)

	var desc sql.NullString
	if description != "" {
		desc = sql.NullString{String: description, Valid: true}
	}

	var ctxRef sql.NullString
	if contextRef != "" {
		ctxRef = sql.NullString{String: contextRef, Valid: true}
	}

	_, err = database.Exec(
		"INSERT INTO work_orders (id, mission_id, title, description, context_ref, status) VALUES (?, ?, ?, ?, ?, ?)",
		id, missionID, title, desc, ctxRef, "backlog",
	)
	if err != nil {
		return nil, err
	}

	return GetWorkOrder(id)
}

// GetWorkOrder retrieves a work order by ID
func GetWorkOrder(id string) (*WorkOrder, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	wo := &WorkOrder{}
	err = database.QueryRow(
		"SELECT id, mission_id, title, description, type, status, priority, parent_id, assigned_grove_id, context_ref, created_at, updated_at, claimed_at, completed_at FROM work_orders WHERE id = ?",
		id,
	).Scan(&wo.ID, &wo.MissionID, &wo.Title, &wo.Description, &wo.Type, &wo.Status, &wo.Priority, &wo.ParentID, &wo.AssignedGroveID, &wo.ContextRef, &wo.CreatedAt, &wo.UpdatedAt, &wo.ClaimedAt, &wo.CompletedAt)

	if err != nil {
		return nil, err
	}

	return wo, nil
}

// ListWorkOrders retrieves work orders, optionally filtered by mission and/or status
func ListWorkOrders(missionID, status string) ([]*WorkOrder, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, mission_id, title, description, type, status, priority, parent_id, assigned_grove_id, context_ref, created_at, updated_at, claimed_at, completed_at FROM work_orders WHERE 1=1"
	args := []interface{}{}

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

	var orders []*WorkOrder
	for rows.Next() {
		wo := &WorkOrder{}
		err := rows.Scan(&wo.ID, &wo.MissionID, &wo.Title, &wo.Description, &wo.Type, &wo.Status, &wo.Priority, &wo.ParentID, &wo.AssignedGroveID, &wo.ContextRef, &wo.CreatedAt, &wo.UpdatedAt, &wo.ClaimedAt, &wo.CompletedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, wo)
	}

	return orders, nil
}

// ClaimWorkOrder assigns a work order to a grove and marks it as in_progress
func ClaimWorkOrder(id, groveID string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var groveIDNullable sql.NullString
	if groveID != "" {
		groveIDNullable = sql.NullString{String: groveID, Valid: true}
	}

	_, err = database.Exec(
		"UPDATE work_orders SET status = 'in_progress', assigned_grove_id = ?, claimed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		groveIDNullable, id,
	)

	return err
}

// CompleteWorkOrder marks a work order as complete
func CompleteWorkOrder(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	_, err = database.Exec(
		"UPDATE work_orders SET status = 'complete', completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}
