// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	corecommission "github.com/example/orc/internal/core/commission"
	"github.com/example/orc/internal/ports/secondary"
)

// CommissionRepository implements secondary.CommissionRepository with SQLite.
type CommissionRepository struct {
	db *sql.DB
}

// NewCommissionRepository creates a new SQLite commission repository.
func NewCommissionRepository(db *sql.DB) *CommissionRepository {
	return &CommissionRepository{db: db}
}

// Create persists a new commission.
// The commission record must have ID and Status pre-populated by the service layer.
func (r *CommissionRepository) Create(ctx context.Context, commission *secondary.CommissionRecord) error {
	if commission.ID == "" {
		return fmt.Errorf("commission ID must be pre-populated by service layer")
	}
	if commission.Status == "" {
		return fmt.Errorf("commission Status must be pre-populated by service layer")
	}

	var desc sql.NullString
	if commission.Description != "" {
		desc = sql.NullString{String: commission.Description, Valid: true}
	}

	var workshopID sql.NullString
	if commission.WorkshopID != "" {
		workshopID = sql.NullString{String: commission.WorkshopID, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO commissions (id, workshop_id, title, description, status) VALUES (?, ?, ?, ?, ?)",
		commission.ID, workshopID, commission.Title, desc, commission.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to create commission: %w", err)
	}

	return nil
}

// GetByID retrieves a commission by its ID.
func (r *CommissionRepository) GetByID(ctx context.Context, id string) (*secondary.CommissionRecord, error) {
	var (
		workshopID  sql.NullString
		desc        sql.NullString
		pinned      bool
		createdAt   time.Time
		updatedAt   sql.NullTime
		completedAt sql.NullTime
	)

	record := &secondary.CommissionRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, workshop_id, title, description, status, pinned, created_at, updated_at, completed_at FROM commissions WHERE id = ?",
		id,
	).Scan(&record.ID, &workshopID, &record.Title, &desc, &record.Status, &pinned, &createdAt, &updatedAt, &completedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("commission %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get commission: %w", err)
	}

	record.WorkshopID = workshopID.String
	record.Description = desc.String
	record.Pinned = pinned
	record.CreatedAt = createdAt.Format(time.RFC3339)
	if updatedAt.Valid {
		record.UpdatedAt = updatedAt.Time.Format(time.RFC3339)
	}
	if completedAt.Valid {
		record.CompletedAt = completedAt.Time.Format(time.RFC3339)
	}

	return record, nil
}

// List retrieves commissions matching the given filters.
func (r *CommissionRepository) List(ctx context.Context, filters secondary.CommissionFilters) ([]*secondary.CommissionRecord, error) {
	query := "SELECT id, workshop_id, title, description, status, pinned, created_at, updated_at, completed_at FROM commissions"
	args := []any{}

	if filters.Status != "" {
		query += " WHERE status = ?"
		args = append(args, filters.Status)
	}

	query += " ORDER BY created_at DESC"

	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list commissions: %w", err)
	}
	defer rows.Close()

	var commissions []*secondary.CommissionRecord
	for rows.Next() {
		var (
			workshopID  sql.NullString
			desc        sql.NullString
			pinned      bool
			createdAt   time.Time
			updatedAt   sql.NullTime
			completedAt sql.NullTime
		)

		record := &secondary.CommissionRecord{}
		err := rows.Scan(&record.ID, &workshopID, &record.Title, &desc, &record.Status, &pinned, &createdAt, &updatedAt, &completedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan commission: %w", err)
		}

		record.WorkshopID = workshopID.String
		record.Description = desc.String
		record.Pinned = pinned
		record.CreatedAt = createdAt.Format(time.RFC3339)
		if updatedAt.Valid {
			record.UpdatedAt = updatedAt.Time.Format(time.RFC3339)
		}
		if completedAt.Valid {
			record.CompletedAt = completedAt.Time.Format(time.RFC3339)
		}

		commissions = append(commissions, record)
	}

	return commissions, nil
}

// Update updates an existing commission.
// The service layer is responsible for setting CompletedAt when status changes to complete.
func (r *CommissionRepository) Update(ctx context.Context, commission *secondary.CommissionRecord) error {
	// Build dynamic query based on what's being updated
	query := "UPDATE commissions SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if commission.Title != "" {
		query += ", title = ?"
		args = append(args, commission.Title)
	}

	if commission.Description != "" {
		query += ", description = ?"
		args = append(args, sql.NullString{String: commission.Description, Valid: true})
	}

	if commission.Status != "" {
		query += ", status = ?"
		args = append(args, commission.Status)
	}

	// CompletedAt is set by service layer when transitioning to complete status
	if commission.CompletedAt != "" {
		completedTime, err := time.Parse(time.RFC3339, commission.CompletedAt)
		if err == nil {
			query += ", completed_at = ?"
			args = append(args, sql.NullTime{Time: completedTime, Valid: true})
		}
	}

	query += " WHERE id = ?"
	args = append(args, commission.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update commission: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("commission %s not found", commission.ID)
	}

	return nil
}

// Delete removes a commission from persistence.
func (r *CommissionRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM commissions WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete commission: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("commission %s not found", id)
	}

	return nil
}

// GetNextID returns the next available commission ID.
// Uses core function for ID format to keep business logic in the functional core.
// COMM-XXX format where XXX is extracted from position 6 (COMM- is 5 chars + dash)
func (r *CommissionRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 6) AS INTEGER)), 0) FROM commissions",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next commission ID: %w", err)
	}

	return corecommission.GenerateCommissionID(maxID), nil
}

// CountShipments returns the number of shipments for a commission.
func (r *CommissionRepository) CountShipments(ctx context.Context, commissionID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM shipments WHERE commission_id = ?",
		commissionID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count shipments: %w", err)
	}

	return count, nil
}

// Pin pins a commission to keep it visible.
func (r *CommissionRepository) Pin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE commissions SET pinned = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to pin commission: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("commission %s not found", id)
	}

	return nil
}

// Unpin unpins a commission.
func (r *CommissionRepository) Unpin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE commissions SET pinned = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to unpin commission: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("commission %s not found", id)
	}

	return nil
}

// Ensure CommissionRepository implements the interface
var _ secondary.CommissionRepository = (*CommissionRepository)(nil)
