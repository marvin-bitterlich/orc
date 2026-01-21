// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// PlanRepository implements secondary.PlanRepository with SQLite.
type PlanRepository struct {
	db *sql.DB
}

// NewPlanRepository creates a new SQLite plan repository.
func NewPlanRepository(db *sql.DB) *PlanRepository {
	return &PlanRepository{db: db}
}

// Create persists a new plan.
func (r *PlanRepository) Create(ctx context.Context, plan *secondary.PlanRecord) error {
	var desc sql.NullString
	if plan.Description != "" {
		desc = sql.NullString{String: plan.Description, Valid: true}
	}

	var content sql.NullString
	if plan.Content != "" {
		content = sql.NullString{String: plan.Content, Valid: true}
	}

	var shipmentID sql.NullString
	if plan.ShipmentID != "" {
		shipmentID = sql.NullString{String: plan.ShipmentID, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO plans (id, shipment_id, mission_id, title, description, content, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
		plan.ID, shipmentID, plan.MissionID, plan.Title, desc, content, "draft",
	)
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}

	return nil
}

// GetByID retrieves a plan by its ID.
func (r *PlanRepository) GetByID(ctx context.Context, id string) (*secondary.PlanRecord, error) {
	var (
		shipmentID       sql.NullString
		desc             sql.NullString
		content          sql.NullString
		pinned           bool
		createdAt        time.Time
		updatedAt        time.Time
		approvedAt       sql.NullTime
		conclaveID       sql.NullString
		promotedFromID   sql.NullString
		promotedFromType sql.NullString
	)

	record := &secondary.PlanRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, shipment_id, mission_id, title, description, status, content, pinned,
			created_at, updated_at, approved_at, conclave_id, promoted_from_id, promoted_from_type
		FROM plans WHERE id = ?`,
		id,
	).Scan(&record.ID, &shipmentID, &record.MissionID, &record.Title, &desc, &record.Status, &content, &pinned,
		&createdAt, &updatedAt, &approvedAt, &conclaveID, &promotedFromID, &promotedFromType)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("plan %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	record.ShipmentID = shipmentID.String
	record.Description = desc.String
	record.Content = content.String
	record.Pinned = pinned
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	if approvedAt.Valid {
		record.ApprovedAt = approvedAt.Time.Format(time.RFC3339)
	}
	record.ConclaveID = conclaveID.String
	record.PromotedFromID = promotedFromID.String
	record.PromotedFromType = promotedFromType.String

	return record, nil
}

// List retrieves plans matching the given filters.
func (r *PlanRepository) List(ctx context.Context, filters secondary.PlanFilters) ([]*secondary.PlanRecord, error) {
	query := `SELECT id, shipment_id, mission_id, title, description, status, content, pinned,
		created_at, updated_at, approved_at, conclave_id, promoted_from_id, promoted_from_type
		FROM plans WHERE 1=1`
	args := []any{}

	if filters.ShipmentID != "" {
		query += " AND shipment_id = ?"
		args = append(args, filters.ShipmentID)
	}

	if filters.MissionID != "" {
		query += " AND mission_id = ?"
		args = append(args, filters.MissionID)
	}

	if filters.Status != "" {
		query += " AND status = ?"
		args = append(args, filters.Status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list plans: %w", err)
	}
	defer rows.Close()

	var plans []*secondary.PlanRecord
	for rows.Next() {
		var (
			shipmentID       sql.NullString
			desc             sql.NullString
			content          sql.NullString
			pinned           bool
			createdAt        time.Time
			updatedAt        time.Time
			approvedAt       sql.NullTime
			conclaveID       sql.NullString
			promotedFromID   sql.NullString
			promotedFromType sql.NullString
		)

		record := &secondary.PlanRecord{}
		err := rows.Scan(&record.ID, &shipmentID, &record.MissionID, &record.Title, &desc, &record.Status, &content, &pinned,
			&createdAt, &updatedAt, &approvedAt, &conclaveID, &promotedFromID, &promotedFromType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plan: %w", err)
		}

		record.ShipmentID = shipmentID.String
		record.Description = desc.String
		record.Content = content.String
		record.Pinned = pinned
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if approvedAt.Valid {
			record.ApprovedAt = approvedAt.Time.Format(time.RFC3339)
		}
		record.ConclaveID = conclaveID.String
		record.PromotedFromID = promotedFromID.String
		record.PromotedFromType = promotedFromType.String

		plans = append(plans, record)
	}

	return plans, nil
}

// Update updates an existing plan.
func (r *PlanRepository) Update(ctx context.Context, plan *secondary.PlanRecord) error {
	query := "UPDATE plans SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if plan.Title != "" {
		query += ", title = ?"
		args = append(args, plan.Title)
	}

	if plan.Description != "" {
		query += ", description = ?"
		args = append(args, sql.NullString{String: plan.Description, Valid: true})
	}

	if plan.Content != "" {
		query += ", content = ?"
		args = append(args, sql.NullString{String: plan.Content, Valid: true})
	}

	query += " WHERE id = ?"
	args = append(args, plan.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update plan: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("plan %s not found", plan.ID)
	}

	return nil
}

// Delete removes a plan from persistence.
func (r *PlanRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM plans WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete plan: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("plan %s not found", id)
	}

	return nil
}

// Pin pins a plan.
func (r *PlanRepository) Pin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE plans SET pinned = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to pin plan: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("plan %s not found", id)
	}

	return nil
}

// Unpin unpins a plan.
func (r *PlanRepository) Unpin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE plans SET pinned = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to unpin plan: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("plan %s not found", id)
	}

	return nil
}

// GetNextID returns the next available plan ID.
func (r *PlanRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 6) AS INTEGER)), 0) FROM plans",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next plan ID: %w", err)
	}

	return fmt.Sprintf("PLAN-%03d", maxID+1), nil
}

// Approve approves a plan and sets the approved_at timestamp.
func (r *PlanRepository) Approve(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE plans SET status = 'approved', approved_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to approve plan: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("plan %s not found", id)
	}

	return nil
}

// GetActivePlanForShipment retrieves the active (draft) plan for a shipment.
func (r *PlanRepository) GetActivePlanForShipment(ctx context.Context, shipmentID string) (*secondary.PlanRecord, error) {
	var (
		shipmentIDCol    sql.NullString
		desc             sql.NullString
		content          sql.NullString
		pinned           bool
		createdAt        time.Time
		updatedAt        time.Time
		approvedAt       sql.NullTime
		conclaveID       sql.NullString
		promotedFromID   sql.NullString
		promotedFromType sql.NullString
	)

	record := &secondary.PlanRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, shipment_id, mission_id, title, description, status, content, pinned,
			created_at, updated_at, approved_at, conclave_id, promoted_from_id, promoted_from_type
		FROM plans WHERE shipment_id = ? AND status = 'draft' LIMIT 1`,
		shipmentID,
	).Scan(&record.ID, &shipmentIDCol, &record.MissionID, &record.Title, &desc, &record.Status, &content, &pinned,
		&createdAt, &updatedAt, &approvedAt, &conclaveID, &promotedFromID, &promotedFromType)

	if err == sql.ErrNoRows {
		return nil, nil // No active plan is not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active plan for shipment: %w", err)
	}

	record.ShipmentID = shipmentIDCol.String
	record.Description = desc.String
	record.Content = content.String
	record.Pinned = pinned
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	if approvedAt.Valid {
		record.ApprovedAt = approvedAt.Time.Format(time.RFC3339)
	}
	record.ConclaveID = conclaveID.String
	record.PromotedFromID = promotedFromID.String
	record.PromotedFromType = promotedFromType.String

	return record, nil
}

// HasActivePlanForShipment checks if a shipment has an active (draft) plan.
func (r *PlanRepository) HasActivePlanForShipment(ctx context.Context, shipmentID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM plans WHERE shipment_id = ? AND status = 'draft'",
		shipmentID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check active plan for shipment: %w", err)
	}
	return count > 0, nil
}

// MissionExists checks if a mission exists.
func (r *PlanRepository) MissionExists(ctx context.Context, missionID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM missions WHERE id = ?", missionID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check mission existence: %w", err)
	}
	return count > 0, nil
}

// ShipmentExists checks if a shipment exists.
func (r *PlanRepository) ShipmentExists(ctx context.Context, shipmentID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM shipments WHERE id = ?", shipmentID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check shipment existence: %w", err)
	}
	return count > 0, nil
}

// Ensure PlanRepository implements the interface
var _ secondary.PlanRepository = (*PlanRepository)(nil)
