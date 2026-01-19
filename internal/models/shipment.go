package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/db"
)

type Shipment struct {
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

// CreateShipment creates a new shipment
func CreateShipment(missionID, title, description string) (*Shipment, error) {
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

	// Generate shipment ID by finding max existing ID
	var maxID int
	err = database.QueryRow("SELECT COALESCE(MAX(CAST(SUBSTR(id, 6) AS INTEGER)), 0) FROM shipments").Scan(&maxID)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("SHIP-%03d", maxID+1)

	var desc sql.NullString
	if description != "" {
		desc = sql.NullString{String: description, Valid: true}
	}

	_, err = database.Exec(
		"INSERT INTO shipments (id, mission_id, title, description, status) VALUES (?, ?, ?, ?, ?)",
		id, missionID, title, desc, "active",
	)
	if err != nil {
		return nil, err
	}

	return GetShipment(id)
}

// GetShipment retrieves a shipment by ID
func GetShipment(id string) (*Shipment, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	shipment := &Shipment{}
	err = database.QueryRow(
		"SELECT id, mission_id, title, description, status, assigned_grove_id, pinned, created_at, updated_at, completed_at FROM shipments WHERE id = ?",
		id,
	).Scan(&shipment.ID, &shipment.MissionID, &shipment.Title, &shipment.Description, &shipment.Status, &shipment.AssignedGroveID, &shipment.Pinned, &shipment.CreatedAt, &shipment.UpdatedAt, &shipment.CompletedAt)

	if err != nil {
		return nil, err
	}

	return shipment, nil
}

// ListShipments retrieves shipments, optionally filtered by mission and/or status
func ListShipments(missionID, status string) ([]*Shipment, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, mission_id, title, description, status, assigned_grove_id, pinned, created_at, updated_at, completed_at FROM shipments WHERE 1=1"
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

	var shipments []*Shipment
	for rows.Next() {
		shipment := &Shipment{}
		err := rows.Scan(&shipment.ID, &shipment.MissionID, &shipment.Title, &shipment.Description, &shipment.Status, &shipment.AssignedGroveID, &shipment.Pinned, &shipment.CreatedAt, &shipment.UpdatedAt, &shipment.CompletedAt)
		if err != nil {
			return nil, err
		}
		shipments = append(shipments, shipment)
	}

	return shipments, nil
}

// CompleteShipment marks a shipment as complete
func CompleteShipment(id string) error {
	shipment, err := GetShipment(id)
	if err != nil {
		return err
	}

	if shipment.Pinned {
		return fmt.Errorf("cannot complete pinned shipment %s. Unpin first with: orc shipment unpin %s", id, id)
	}

	database, err := db.GetDB()
	if err != nil {
		return err
	}

	_, err = database.Exec(
		"UPDATE shipments SET status = 'complete', completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// UpdateShipment updates the title and/or description of a shipment
func UpdateShipment(id, title, description string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM shipments WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("shipment %s not found", id)
	}

	if title != "" && description != "" {
		_, err = database.Exec(
			"UPDATE shipments SET title = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			title, description, id,
		)
	} else if title != "" {
		_, err = database.Exec(
			"UPDATE shipments SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			title, id,
		)
	} else if description != "" {
		_, err = database.Exec(
			"UPDATE shipments SET description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			description, id,
		)
	}

	return err
}

// PinShipment pins a shipment
func PinShipment(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM shipments WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("shipment %s not found", id)
	}

	_, err = database.Exec(
		"UPDATE shipments SET pinned = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// UnpinShipment unpins a shipment
func UnpinShipment(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM shipments WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("shipment %s not found", id)
	}

	_, err = database.Exec(
		"UPDATE shipments SET pinned = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// AssignShipmentToGrove assigns a shipment to a grove
func AssignShipmentToGrove(shipmentID, groveID string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM shipments WHERE id = ?", shipmentID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("shipment %s not found", shipmentID)
	}

	// Check if grove is already assigned to another shipment
	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM shipments WHERE assigned_grove_id = ? AND id != ?", groveID, shipmentID).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		var assignedShipmentID string
		err = database.QueryRow("SELECT id FROM shipments WHERE assigned_grove_id = ? AND id != ? LIMIT 1", groveID, shipmentID).Scan(&assignedShipmentID)
		if err != nil {
			return err
		}
		return fmt.Errorf("grove already assigned to shipment %s", assignedShipmentID)
	}

	_, err = database.Exec(
		"UPDATE shipments SET assigned_grove_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		groveID, shipmentID,
	)
	if err != nil {
		return err
	}

	// Also assign all child tasks to the same grove
	_, err = database.Exec(
		"UPDATE tasks SET assigned_grove_id = ?, updated_at = CURRENT_TIMESTAMP WHERE shipment_id = ?",
		groveID, shipmentID,
	)

	return err
}

// GetShipmentsByGrove retrieves shipments assigned to a grove
func GetShipmentsByGrove(groveID string) ([]*Shipment, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, mission_id, title, description, status, assigned_grove_id, pinned, created_at, updated_at, completed_at FROM shipments WHERE assigned_grove_id = ?"
	rows, err := database.Query(query, groveID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shipments []*Shipment
	for rows.Next() {
		shipment := &Shipment{}
		err := rows.Scan(&shipment.ID, &shipment.MissionID, &shipment.Title, &shipment.Description, &shipment.Status, &shipment.AssignedGroveID, &shipment.Pinned, &shipment.CreatedAt, &shipment.UpdatedAt, &shipment.CompletedAt)
		if err != nil {
			return nil, err
		}
		shipments = append(shipments, shipment)
	}

	return shipments, nil
}

// DeleteShipment deletes a shipment by ID
func DeleteShipment(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM shipments WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("shipment %s not found", id)
	}

	_, err = database.Exec("DELETE FROM shipments WHERE id = ?", id)
	return err
}

// GetShipmentTasks gets all tasks in a shipment
func GetShipmentTasks(shipmentID string) ([]*Task, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, shipment_id, mission_id, title, description, type, status, priority, assigned_grove_id, pinned, created_at, updated_at, claimed_at, completed_at, conclave_id, promoted_from_id, promoted_from_type FROM tasks WHERE shipment_id = ? ORDER BY created_at ASC"
	rows, err := database.Query(query, shipmentID)
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
