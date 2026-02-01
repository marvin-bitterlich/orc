// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// KennelRepository implements secondary.KennelRepository with SQLite.
type KennelRepository struct {
	db *sql.DB
}

// NewKennelRepository creates a new SQLite kennel repository.
func NewKennelRepository(db *sql.DB) *KennelRepository {
	return &KennelRepository{db: db}
}

// Create persists a new kennel.
func (r *KennelRepository) Create(ctx context.Context, kennel *secondary.KennelRecord) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO kennels (id, workbench_id, status) VALUES (?, ?, ?)`,
		kennel.ID,
		kennel.WorkbenchID,
		kennel.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to create kennel: %w", err)
	}

	return nil
}

// GetByID retrieves a kennel by its ID.
func (r *KennelRepository) GetByID(ctx context.Context, id string) (*secondary.KennelRecord, error) {
	var (
		createdAt time.Time
		updatedAt time.Time
	)

	record := &secondary.KennelRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, workbench_id, status, created_at, updated_at FROM kennels WHERE id = ?`,
		id,
	).Scan(&record.ID, &record.WorkbenchID, &record.Status, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("kennel %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get kennel: %w", err)
	}
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)

	return record, nil
}

// GetByWorkbench retrieves a kennel by workbench ID.
func (r *KennelRepository) GetByWorkbench(ctx context.Context, workbenchID string) (*secondary.KennelRecord, error) {
	var (
		createdAt time.Time
		updatedAt time.Time
	)

	record := &secondary.KennelRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, workbench_id, status, created_at, updated_at FROM kennels WHERE workbench_id = ?`,
		workbenchID,
	).Scan(&record.ID, &record.WorkbenchID, &record.Status, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("kennel for workbench %s not found", workbenchID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get kennel: %w", err)
	}
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)

	return record, nil
}

// List retrieves kennels matching the given filters.
func (r *KennelRepository) List(ctx context.Context, filters secondary.KennelFilters) ([]*secondary.KennelRecord, error) {
	query := `SELECT id, workbench_id, status, created_at, updated_at FROM kennels WHERE 1=1`
	args := []any{}

	if filters.WorkbenchID != "" {
		query += " AND workbench_id = ?"
		args = append(args, filters.WorkbenchID)
	}

	if filters.Status != "" {
		query += " AND status = ?"
		args = append(args, filters.Status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list kennels: %w", err)
	}
	defer rows.Close()

	var kennels []*secondary.KennelRecord
	for rows.Next() {
		var (
			createdAt time.Time
			updatedAt time.Time
		)

		record := &secondary.KennelRecord{}
		err := rows.Scan(&record.ID, &record.WorkbenchID, &record.Status, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan kennel: %w", err)
		}
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)

		kennels = append(kennels, record)
	}

	return kennels, nil
}

// Update updates an existing kennel.
func (r *KennelRepository) Update(ctx context.Context, kennel *secondary.KennelRecord) error {
	query := "UPDATE kennels SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if kennel.Status != "" {
		query += ", status = ?"
		args = append(args, kennel.Status)
	}

	query += " WHERE id = ?"
	args = append(args, kennel.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update kennel: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("kennel %s not found", kennel.ID)
	}

	return nil
}

// Delete removes a kennel from persistence.
func (r *KennelRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM kennels WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete kennel: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("kennel %s not found", id)
	}

	return nil
}

// GetNextID returns the next available kennel ID.
func (r *KennelRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	prefixLen := len("KENNEL-") + 1
	err := r.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT COALESCE(MAX(CAST(SUBSTR(id, %d) AS INTEGER)), 0) FROM kennels", prefixLen),
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next kennel ID: %w", err)
	}

	return fmt.Sprintf("KENNEL-%03d", maxID+1), nil
}

// UpdateStatus updates the status of a kennel.
func (r *KennelRepository) UpdateStatus(ctx context.Context, id, status string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE kennels SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		status, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update kennel status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("kennel %s not found", id)
	}

	return nil
}

// WorkbenchExists checks if a workbench exists.
func (r *KennelRepository) WorkbenchExists(ctx context.Context, workbenchID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM workbenches WHERE id = ?", workbenchID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check workbench existence: %w", err)
	}
	return count > 0, nil
}

// WorkbenchHasKennel checks if a workbench already has a kennel (for 1:1 constraint).
func (r *KennelRepository) WorkbenchHasKennel(ctx context.Context, workbenchID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM kennels WHERE workbench_id = ?", workbenchID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check existing kennel: %w", err)
	}
	return count > 0, nil
}

// Ensure KennelRepository implements the interface
var _ secondary.KennelRepository = (*KennelRepository)(nil)
