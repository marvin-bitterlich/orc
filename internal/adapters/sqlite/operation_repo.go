// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// OperationRepository implements secondary.OperationRepository with SQLite.
type OperationRepository struct {
	db *sql.DB
}

// NewOperationRepository creates a new SQLite operation repository.
func NewOperationRepository(db *sql.DB) *OperationRepository {
	return &OperationRepository{db: db}
}

// Create persists a new operation.
func (r *OperationRepository) Create(ctx context.Context, operation *secondary.OperationRecord) error {
	var desc sql.NullString
	if operation.Description != "" {
		desc = sql.NullString{String: operation.Description, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO operations (id, commission_id, title, description, status) VALUES (?, ?, ?, ?, ?)",
		operation.ID, operation.CommissionID, operation.Title, desc, "ready",
	)
	if err != nil {
		return fmt.Errorf("failed to create operation: %w", err)
	}

	return nil
}

// GetByID retrieves an operation by its ID.
func (r *OperationRepository) GetByID(ctx context.Context, id string) (*secondary.OperationRecord, error) {
	var (
		desc        sql.NullString
		createdAt   time.Time
		updatedAt   time.Time
		completedAt sql.NullTime
	)

	record := &secondary.OperationRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, commission_id, title, description, status, created_at, updated_at, completed_at FROM operations WHERE id = ?",
		id,
	).Scan(&record.ID, &record.CommissionID, &record.Title, &desc, &record.Status, &createdAt, &updatedAt, &completedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("operation %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get operation: %w", err)
	}

	record.Description = desc.String
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	if completedAt.Valid {
		record.CompletedAt = completedAt.Time.Format(time.RFC3339)
	}

	return record, nil
}

// List retrieves operations matching the given filters.
func (r *OperationRepository) List(ctx context.Context, filters secondary.OperationFilters) ([]*secondary.OperationRecord, error) {
	query := "SELECT id, commission_id, title, description, status, created_at, updated_at, completed_at FROM operations WHERE 1=1"
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
		return nil, fmt.Errorf("failed to list operations: %w", err)
	}
	defer rows.Close()

	var operations []*secondary.OperationRecord
	for rows.Next() {
		var (
			desc        sql.NullString
			createdAt   time.Time
			updatedAt   time.Time
			completedAt sql.NullTime
		)

		record := &secondary.OperationRecord{}
		err := rows.Scan(&record.ID, &record.CommissionID, &record.Title, &desc, &record.Status, &createdAt, &updatedAt, &completedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan operation: %w", err)
		}

		record.Description = desc.String
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if completedAt.Valid {
			record.CompletedAt = completedAt.Time.Format(time.RFC3339)
		}

		operations = append(operations, record)
	}

	return operations, nil
}

// UpdateStatus updates the status and optionally completed_at timestamp.
func (r *OperationRepository) UpdateStatus(ctx context.Context, id, status string, setCompleted bool) error {
	var query string
	var args []any

	if setCompleted {
		query = "UPDATE operations SET status = ?, completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
		args = []any{status, id}
	} else {
		query = "UPDATE operations SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
		args = []any{status, id}
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update operation status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("operation %s not found", id)
	}

	return nil
}

// GetNextID returns the next available operation ID.
func (r *OperationRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 4) AS INTEGER)), 0) FROM operations",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next operation ID: %w", err)
	}

	return fmt.Sprintf("OP-%03d", maxID+1), nil
}

// CommissionExists checks if a commission exists.
func (r *OperationRepository) CommissionExists(ctx context.Context, commissionID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM commissions WHERE id = ?", commissionID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check commission existence: %w", err)
	}
	return count > 0, nil
}

// Ensure OperationRepository implements the interface
var _ secondary.OperationRepository = (*OperationRepository)(nil)
