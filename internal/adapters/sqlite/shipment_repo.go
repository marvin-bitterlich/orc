// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// ShipmentRepository implements secondary.ShipmentRepository with SQLite.
type ShipmentRepository struct {
	db *sql.DB
}

// NewShipmentRepository creates a new SQLite shipment repository.
func NewShipmentRepository(db *sql.DB) *ShipmentRepository {
	return &ShipmentRepository{db: db}
}

// Create persists a new shipment.
func (r *ShipmentRepository) Create(ctx context.Context, shipment *secondary.ShipmentRecord) error {
	var desc sql.NullString
	if shipment.Description != "" {
		desc = sql.NullString{String: shipment.Description, Valid: true}
	}

	var repoID, branch sql.NullString
	if shipment.RepoID != "" {
		repoID = sql.NullString{String: shipment.RepoID, Valid: true}
	}
	if shipment.Branch != "" {
		branch = sql.NullString{String: shipment.Branch, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO shipments (id, commission_id, title, description, status, repo_id, branch) VALUES (?, ?, ?, ?, ?, ?, ?)",
		shipment.ID, shipment.CommissionID, shipment.Title, desc, "active", repoID, branch,
	)
	if err != nil {
		return fmt.Errorf("failed to create shipment: %w", err)
	}

	return nil
}

// GetByID retrieves a shipment by its ID.
func (r *ShipmentRepository) GetByID(ctx context.Context, id string) (*secondary.ShipmentRecord, error) {
	var (
		desc                sql.NullString
		assignedWorkbenchID sql.NullString
		repoID              sql.NullString
		branch              sql.NullString
		pinned              bool
		createdAt           time.Time
		updatedAt           time.Time
		completedAt         sql.NullTime
	)

	record := &secondary.ShipmentRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, commission_id, title, description, status, assigned_workbench_id, repo_id, branch, pinned, created_at, updated_at, completed_at FROM shipments WHERE id = ?",
		id,
	).Scan(&record.ID, &record.CommissionID, &record.Title, &desc, &record.Status, &assignedWorkbenchID, &repoID, &branch, &pinned, &createdAt, &updatedAt, &completedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("shipment %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get shipment: %w", err)
	}

	record.Description = desc.String
	record.AssignedWorkbenchID = assignedWorkbenchID.String
	if repoID.Valid {
		record.RepoID = repoID.String
	}
	if branch.Valid {
		record.Branch = branch.String
	}
	record.Pinned = pinned
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	if completedAt.Valid {
		record.CompletedAt = completedAt.Time.Format(time.RFC3339)
	}

	return record, nil
}

// List retrieves shipments matching the given filters.
func (r *ShipmentRepository) List(ctx context.Context, filters secondary.ShipmentFilters) ([]*secondary.ShipmentRecord, error) {
	query := "SELECT id, commission_id, title, description, status, assigned_workbench_id, repo_id, branch, pinned, created_at, updated_at, completed_at FROM shipments WHERE 1=1"
	args := []any{}

	if filters.CommissionID != "" {
		query += " AND commission_id = ?"
		args = append(args, filters.CommissionID)
	}

	if filters.Status != "" {
		query += " AND status = ?"
		args = append(args, filters.Status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list shipments: %w", err)
	}
	defer rows.Close()

	var shipments []*secondary.ShipmentRecord
	for rows.Next() {
		var (
			desc                sql.NullString
			assignedWorkbenchID sql.NullString
			repoID              sql.NullString
			branch              sql.NullString
			pinned              bool
			createdAt           time.Time
			updatedAt           time.Time
			completedAt         sql.NullTime
		)

		record := &secondary.ShipmentRecord{}
		err := rows.Scan(&record.ID, &record.CommissionID, &record.Title, &desc, &record.Status, &assignedWorkbenchID, &repoID, &branch, &pinned, &createdAt, &updatedAt, &completedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shipment: %w", err)
		}

		record.Description = desc.String
		record.AssignedWorkbenchID = assignedWorkbenchID.String
		if repoID.Valid {
			record.RepoID = repoID.String
		}
		if branch.Valid {
			record.Branch = branch.String
		}
		record.Pinned = pinned
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if completedAt.Valid {
			record.CompletedAt = completedAt.Time.Format(time.RFC3339)
		}

		shipments = append(shipments, record)
	}

	return shipments, nil
}

// Update updates an existing shipment.
func (r *ShipmentRepository) Update(ctx context.Context, shipment *secondary.ShipmentRecord) error {
	query := "UPDATE shipments SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if shipment.Title != "" {
		query += ", title = ?"
		args = append(args, shipment.Title)
	}

	if shipment.Description != "" {
		query += ", description = ?"
		args = append(args, sql.NullString{String: shipment.Description, Valid: true})
	}

	query += " WHERE id = ?"
	args = append(args, shipment.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update shipment: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("shipment %s not found", shipment.ID)
	}

	return nil
}

// Delete removes a shipment from persistence.
func (r *ShipmentRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM shipments WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete shipment: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("shipment %s not found", id)
	}

	return nil
}

// Pin pins a shipment.
func (r *ShipmentRepository) Pin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE shipments SET pinned = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to pin shipment: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("shipment %s not found", id)
	}

	return nil
}

// Unpin unpins a shipment.
func (r *ShipmentRepository) Unpin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE shipments SET pinned = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to unpin shipment: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("shipment %s not found", id)
	}

	return nil
}

// GetNextID returns the next available shipment ID.
func (r *ShipmentRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 6) AS INTEGER)), 0) FROM shipments",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next shipment ID: %w", err)
	}

	return fmt.Sprintf("SHIP-%03d", maxID+1), nil
}

// GetByWorkbench retrieves shipments assigned to a workbench.
func (r *ShipmentRepository) GetByWorkbench(ctx context.Context, workbenchID string) ([]*secondary.ShipmentRecord, error) {
	query := "SELECT id, commission_id, title, description, status, assigned_workbench_id, repo_id, branch, pinned, created_at, updated_at, completed_at FROM shipments WHERE assigned_workbench_id = ?"
	rows, err := r.db.QueryContext(ctx, query, workbenchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shipments by workbench: %w", err)
	}
	defer rows.Close()

	var shipments []*secondary.ShipmentRecord
	for rows.Next() {
		var (
			desc                sql.NullString
			assignedWorkbenchID sql.NullString
			repoID              sql.NullString
			branch              sql.NullString
			pinned              bool
			createdAt           time.Time
			updatedAt           time.Time
			completedAt         sql.NullTime
		)

		record := &secondary.ShipmentRecord{}
		err := rows.Scan(&record.ID, &record.CommissionID, &record.Title, &desc, &record.Status, &assignedWorkbenchID, &repoID, &branch, &pinned, &createdAt, &updatedAt, &completedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shipment: %w", err)
		}

		record.Description = desc.String
		record.AssignedWorkbenchID = assignedWorkbenchID.String
		if repoID.Valid {
			record.RepoID = repoID.String
		}
		if branch.Valid {
			record.Branch = branch.String
		}
		record.Pinned = pinned
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if completedAt.Valid {
			record.CompletedAt = completedAt.Time.Format(time.RFC3339)
		}

		shipments = append(shipments, record)
	}

	return shipments, nil
}

// AssignWorkbench assigns a shipment to a workbench.
func (r *ShipmentRepository) AssignWorkbench(ctx context.Context, shipmentID, workbenchID string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE shipments SET assigned_workbench_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		workbenchID, shipmentID,
	)
	if err != nil {
		return fmt.Errorf("failed to assign workbench to shipment: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("shipment %s not found", shipmentID)
	}

	return nil
}

// UpdateStatus updates the status and optionally completed_at timestamp.
func (r *ShipmentRepository) UpdateStatus(ctx context.Context, id, status string, setCompleted bool) error {
	var query string
	var args []any

	if setCompleted {
		query = "UPDATE shipments SET status = ?, completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
		args = []any{status, id}
	} else {
		query = "UPDATE shipments SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
		args = []any{status, id}
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update shipment status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("shipment %s not found", id)
	}

	return nil
}

// CommissionExists checks if a commission exists.
func (r *ShipmentRepository) CommissionExists(ctx context.Context, commissionID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM commissions WHERE id = ?", commissionID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check commission existence: %w", err)
	}
	return count > 0, nil
}

// WorkbenchAssignedToOther checks if workbench is assigned to another shipment.
// Returns the shipment ID if assigned to another, empty string if not.
func (r *ShipmentRepository) WorkbenchAssignedToOther(ctx context.Context, workbenchID, excludeShipmentID string) (string, error) {
	var shipmentID string
	err := r.db.QueryRowContext(ctx,
		"SELECT id FROM shipments WHERE assigned_workbench_id = ? AND id != ? LIMIT 1",
		workbenchID, excludeShipmentID,
	).Scan(&shipmentID)

	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to check workbench assignment: %w", err)
	}

	return shipmentID, nil
}

// Ensure ShipmentRepository implements the interface
var _ secondary.ShipmentRepository = (*ShipmentRepository)(nil)
