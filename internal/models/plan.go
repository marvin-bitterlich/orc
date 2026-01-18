package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/db"
)

type Plan struct {
	ID               string
	ShipmentID       sql.NullString
	MissionID        string
	Title            string
	Description      sql.NullString
	Status           string // draft, approved
	Content          sql.NullString
	Pinned           bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ApprovedAt       sql.NullTime
	ConclaveID       sql.NullString
	PromotedFromID   sql.NullString
	PromotedFromType sql.NullString
}

// CreatePlan creates a new plan under a shipment
func CreatePlan(shipmentID, missionID, title, description, content string) (*Plan, error) {
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

	// If shipment ID specified, verify it exists
	if shipmentID != "" {
		err = database.QueryRow("SELECT COUNT(*) FROM shipments WHERE id = ?", shipmentID).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if exists == 0 {
			return nil, fmt.Errorf("shipment %s not found", shipmentID)
		}

		// Check if shipment already has an active (non-approved) plan
		var activePlans int
		err = database.QueryRow("SELECT COUNT(*) FROM plans WHERE shipment_id = ? AND status = 'draft'", shipmentID).Scan(&activePlans)
		if err != nil {
			return nil, err
		}
		if activePlans > 0 {
			return nil, fmt.Errorf("shipment %s already has an active plan. Approve or delete the existing plan first", shipmentID)
		}
	}

	// Generate plan ID by finding max existing ID
	var maxID int
	err = database.QueryRow("SELECT COALESCE(MAX(CAST(SUBSTR(id, 6) AS INTEGER)), 0) FROM plans").Scan(&maxID)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("PLAN-%03d", maxID+1)

	var desc sql.NullString
	if description != "" {
		desc = sql.NullString{String: description, Valid: true}
	}

	var cont sql.NullString
	if content != "" {
		cont = sql.NullString{String: content, Valid: true}
	}

	var shipIDVal sql.NullString
	if shipmentID != "" {
		shipIDVal = sql.NullString{String: shipmentID, Valid: true}
	}

	_, err = database.Exec(
		"INSERT INTO plans (id, shipment_id, mission_id, title, description, content, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
		id, shipIDVal, missionID, title, desc, cont, "draft",
	)
	if err != nil {
		return nil, err
	}

	return GetPlan(id)
}

// GetPlan retrieves a plan by ID
func GetPlan(id string) (*Plan, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	p := &Plan{}
	err = database.QueryRow(
		"SELECT id, shipment_id, mission_id, title, description, status, content, pinned, created_at, updated_at, approved_at, conclave_id, promoted_from_id, promoted_from_type FROM plans WHERE id = ?",
		id,
	).Scan(&p.ID, &p.ShipmentID, &p.MissionID, &p.Title, &p.Description, &p.Status, &p.Content, &p.Pinned, &p.CreatedAt, &p.UpdatedAt, &p.ApprovedAt, &p.ConclaveID, &p.PromotedFromID, &p.PromotedFromType)

	if err != nil {
		return nil, err
	}

	return p, nil
}

// ListPlans retrieves plans, optionally filtered by shipment/status
func ListPlans(shipmentID, status string) ([]*Plan, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, shipment_id, mission_id, title, description, status, content, pinned, created_at, updated_at, approved_at, conclave_id, promoted_from_id, promoted_from_type FROM plans WHERE 1=1"
	args := []any{}

	if shipmentID != "" {
		query += " AND shipment_id = ?"
		args = append(args, shipmentID)
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

// ApprovePlan marks a plan as approved
func ApprovePlan(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	_, err = database.Exec(
		"UPDATE plans SET status = 'approved', approved_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// UpdatePlan updates the title, description, and/or content of a plan
func UpdatePlan(id, title, description, content string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM plans WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("plan %s not found", id)
	}

	// Build update query dynamically based on what's provided
	updates := []string{}
	args := []any{}

	if title != "" {
		updates = append(updates, "title = ?")
		args = append(args, title)
	}
	if description != "" {
		updates = append(updates, "description = ?")
		args = append(args, description)
	}
	if content != "" {
		updates = append(updates, "content = ?")
		args = append(args, content)
	}

	if len(updates) == 0 {
		return nil // Nothing to update
	}

	updates = append(updates, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, id)

	query := "UPDATE plans SET " + joinStrings(updates, ", ") + " WHERE id = ?"
	_, err = database.Exec(query, args...)

	return err
}

// PinPlan pins a plan
func PinPlan(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM plans WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("plan %s not found", id)
	}

	_, err = database.Exec(
		"UPDATE plans SET pinned = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// UnpinPlan unpins a plan
func UnpinPlan(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM plans WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("plan %s not found", id)
	}

	_, err = database.Exec(
		"UPDATE plans SET pinned = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)

	return err
}

// GetShipmentActivePlan returns the active (draft) plan for a shipment, if any
func GetShipmentActivePlan(shipmentID string) (*Plan, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	p := &Plan{}
	err = database.QueryRow(
		"SELECT id, shipment_id, mission_id, title, description, status, content, pinned, created_at, updated_at, approved_at, conclave_id, promoted_from_id, promoted_from_type FROM plans WHERE shipment_id = ? AND status = 'draft' LIMIT 1",
		shipmentID,
	).Scan(&p.ID, &p.ShipmentID, &p.MissionID, &p.Title, &p.Description, &p.Status, &p.Content, &p.Pinned, &p.CreatedAt, &p.UpdatedAt, &p.ApprovedAt, &p.ConclaveID, &p.PromotedFromID, &p.PromotedFromType)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return p, nil
}

// DeletePlan deletes a plan by ID
func DeletePlan(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM plans WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("plan %s not found", id)
	}

	_, err = database.Exec("DELETE FROM plans WHERE id = ?", id)
	return err
}

// Helper function to join strings
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
