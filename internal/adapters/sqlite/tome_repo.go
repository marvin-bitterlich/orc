// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// TomeRepository implements secondary.TomeRepository with SQLite.
type TomeRepository struct {
	db *sql.DB
}

// NewTomeRepository creates a new SQLite tome repository.
func NewTomeRepository(db *sql.DB) *TomeRepository {
	return &TomeRepository{db: db}
}

// Create persists a new tome.
func (r *TomeRepository) Create(ctx context.Context, tome *secondary.TomeRecord) error {
	var desc sql.NullString
	if tome.Description != "" {
		desc = sql.NullString{String: tome.Description, Valid: true}
	}

	var conclaveID sql.NullString
	if tome.ConclaveID != "" {
		conclaveID = sql.NullString{String: tome.ConclaveID, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO tomes (id, commission_id, conclave_id, title, description, status) VALUES (?, ?, ?, ?, ?, ?)",
		tome.ID, tome.CommissionID, conclaveID, tome.Title, desc, "open",
	)
	if err != nil {
		return fmt.Errorf("failed to create tome: %w", err)
	}

	return nil
}

// GetByID retrieves a tome by its ID.
func (r *TomeRepository) GetByID(ctx context.Context, id string) (*secondary.TomeRecord, error) {
	var (
		desc                sql.NullString
		conclaveID          sql.NullString
		assignedWorkbenchID sql.NullString
		pinned              bool
		createdAt           time.Time
		updatedAt           time.Time
		closedAt            sql.NullTime
	)

	record := &secondary.TomeRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, commission_id, conclave_id, title, description, status, assigned_workbench_id, pinned, created_at, updated_at, closed_at FROM tomes WHERE id = ?",
		id,
	).Scan(&record.ID, &record.CommissionID, &conclaveID, &record.Title, &desc, &record.Status, &assignedWorkbenchID, &pinned, &createdAt, &updatedAt, &closedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tome %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tome: %w", err)
	}

	record.Description = desc.String
	record.ConclaveID = conclaveID.String
	record.AssignedWorkbenchID = assignedWorkbenchID.String
	record.Pinned = pinned
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	if closedAt.Valid {
		record.ClosedAt = closedAt.Time.Format(time.RFC3339)
	}

	return record, nil
}

// List retrieves tomes matching the given filters.
func (r *TomeRepository) List(ctx context.Context, filters secondary.TomeFilters) ([]*secondary.TomeRecord, error) {
	query := "SELECT id, commission_id, conclave_id, title, description, status, assigned_workbench_id, pinned, created_at, updated_at, closed_at FROM tomes WHERE 1=1"
	args := []any{}

	if filters.CommissionID != "" {
		query += " AND commission_id = ?"
		args = append(args, filters.CommissionID)
	}

	if filters.ConclaveID != "" {
		query += " AND conclave_id = ?"
		args = append(args, filters.ConclaveID)
	}

	if filters.Status != "" {
		query += " AND status = ?"
		args = append(args, filters.Status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tomes: %w", err)
	}
	defer rows.Close()

	var tomes []*secondary.TomeRecord
	for rows.Next() {
		var (
			desc                sql.NullString
			conclaveID          sql.NullString
			assignedWorkbenchID sql.NullString
			pinned              bool
			createdAt           time.Time
			updatedAt           time.Time
			closedAt            sql.NullTime
		)

		record := &secondary.TomeRecord{}
		err := rows.Scan(&record.ID, &record.CommissionID, &conclaveID, &record.Title, &desc, &record.Status, &assignedWorkbenchID, &pinned, &createdAt, &updatedAt, &closedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tome: %w", err)
		}

		record.Description = desc.String
		record.ConclaveID = conclaveID.String
		record.AssignedWorkbenchID = assignedWorkbenchID.String
		record.Pinned = pinned
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if closedAt.Valid {
			record.ClosedAt = closedAt.Time.Format(time.RFC3339)
		}

		tomes = append(tomes, record)
	}

	return tomes, nil
}

// Update updates an existing tome.
func (r *TomeRepository) Update(ctx context.Context, tome *secondary.TomeRecord) error {
	query := "UPDATE tomes SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if tome.Title != "" {
		query += ", title = ?"
		args = append(args, tome.Title)
	}

	if tome.Description != "" {
		query += ", description = ?"
		args = append(args, sql.NullString{String: tome.Description, Valid: true})
	}

	query += " WHERE id = ?"
	args = append(args, tome.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update tome: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("tome %s not found", tome.ID)
	}

	return nil
}

// Delete removes a tome from persistence.
func (r *TomeRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM tomes WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete tome: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("tome %s not found", id)
	}

	return nil
}

// Pin pins a tome.
func (r *TomeRepository) Pin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE tomes SET pinned = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to pin tome: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("tome %s not found", id)
	}

	return nil
}

// Unpin unpins a tome.
func (r *TomeRepository) Unpin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE tomes SET pinned = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to unpin tome: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("tome %s not found", id)
	}

	return nil
}

// GetNextID returns the next available tome ID.
func (r *TomeRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 6) AS INTEGER)), 0) FROM tomes",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next tome ID: %w", err)
	}

	return fmt.Sprintf("TOME-%03d", maxID+1), nil
}

// UpdateStatus updates the status and optionally closed_at timestamp.
func (r *TomeRepository) UpdateStatus(ctx context.Context, id, status string, setCompleted bool) error {
	var query string
	var args []any

	if setCompleted {
		query = "UPDATE tomes SET status = ?, closed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
		args = []any{status, id}
	} else {
		query = "UPDATE tomes SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
		args = []any{status, id}
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update tome status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("tome %s not found", id)
	}

	return nil
}

// GetByWorkbench retrieves tomes assigned to a workbench.
func (r *TomeRepository) GetByWorkbench(ctx context.Context, workbenchID string) ([]*secondary.TomeRecord, error) {
	query := "SELECT id, commission_id, conclave_id, title, description, status, assigned_workbench_id, pinned, created_at, updated_at, closed_at FROM tomes WHERE assigned_workbench_id = ?"
	rows, err := r.db.QueryContext(ctx, query, workbenchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tomes by workbench: %w", err)
	}
	defer rows.Close()

	var tomes []*secondary.TomeRecord
	for rows.Next() {
		var (
			desc                sql.NullString
			conclaveID          sql.NullString
			assignedWorkbenchID sql.NullString
			pinned              bool
			createdAt           time.Time
			updatedAt           time.Time
			closedAt            sql.NullTime
		)

		record := &secondary.TomeRecord{}
		err := rows.Scan(&record.ID, &record.CommissionID, &conclaveID, &record.Title, &desc, &record.Status, &assignedWorkbenchID, &pinned, &createdAt, &updatedAt, &closedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tome: %w", err)
		}

		record.Description = desc.String
		record.ConclaveID = conclaveID.String
		record.AssignedWorkbenchID = assignedWorkbenchID.String
		record.Pinned = pinned
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if closedAt.Valid {
			record.ClosedAt = closedAt.Time.Format(time.RFC3339)
		}

		tomes = append(tomes, record)
	}

	return tomes, nil
}

// GetByConclave retrieves tomes belonging to a conclave.
func (r *TomeRepository) GetByConclave(ctx context.Context, conclaveID string) ([]*secondary.TomeRecord, error) {
	query := "SELECT id, commission_id, conclave_id, title, description, status, assigned_workbench_id, pinned, created_at, updated_at, closed_at FROM tomes WHERE conclave_id = ? ORDER BY created_at DESC"
	rows, err := r.db.QueryContext(ctx, query, conclaveID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tomes by conclave: %w", err)
	}
	defer rows.Close()

	var tomes []*secondary.TomeRecord
	for rows.Next() {
		var (
			desc                sql.NullString
			conclaveIDVal       sql.NullString
			assignedWorkbenchID sql.NullString
			pinned              bool
			createdAt           time.Time
			updatedAt           time.Time
			closedAt            sql.NullTime
		)

		record := &secondary.TomeRecord{}
		err := rows.Scan(&record.ID, &record.CommissionID, &conclaveIDVal, &record.Title, &desc, &record.Status, &assignedWorkbenchID, &pinned, &createdAt, &updatedAt, &closedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tome: %w", err)
		}

		record.Description = desc.String
		record.ConclaveID = conclaveIDVal.String
		record.AssignedWorkbenchID = assignedWorkbenchID.String
		record.Pinned = pinned
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if closedAt.Valid {
			record.ClosedAt = closedAt.Time.Format(time.RFC3339)
		}

		tomes = append(tomes, record)
	}

	return tomes, nil
}

// AssignWorkbench assigns a tome to a workbench.
func (r *TomeRepository) AssignWorkbench(ctx context.Context, tomeID, workbenchID string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE tomes SET assigned_workbench_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		workbenchID, tomeID,
	)
	if err != nil {
		return fmt.Errorf("failed to assign workbench to tome: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("tome %s not found", tomeID)
	}

	return nil
}

// CommissionExists checks if a commission exists.
func (r *TomeRepository) CommissionExists(ctx context.Context, commissionID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM commissions WHERE id = ?", commissionID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check commission existence: %w", err)
	}
	return count > 0, nil
}

// Ensure TomeRepository implements the interface
var _ secondary.TomeRepository = (*TomeRepository)(nil)
