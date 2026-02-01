package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// StuckRepository implements secondary.StuckRepository using SQLite.
type StuckRepository struct {
	db *sql.DB
}

// NewStuckRepository creates a new StuckRepository.
func NewStuckRepository(db *sql.DB) *StuckRepository {
	return &StuckRepository{db: db}
}

// Create persists a new stuck.
func (r *StuckRepository) Create(ctx context.Context, stuck *secondary.StuckRecord) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if stuck.CreatedAt == "" {
		stuck.CreatedAt = now
	}
	if stuck.UpdatedAt == "" {
		stuck.UpdatedAt = now
	}

	query := `INSERT INTO stucks (id, patrol_id, first_check_id, check_count, status, resolved_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	var firstCheckID, resolvedAt any
	if stuck.FirstCheckID != "" {
		firstCheckID = stuck.FirstCheckID
	}
	if stuck.ResolvedAt != "" {
		resolvedAt = stuck.ResolvedAt
	}

	_, err := r.db.ExecContext(ctx, query,
		stuck.ID,
		stuck.PatrolID,
		firstCheckID,
		stuck.CheckCount,
		stuck.Status,
		resolvedAt,
		stuck.CreatedAt,
		stuck.UpdatedAt,
	)
	return err
}

// GetByID retrieves a stuck by its ID.
func (r *StuckRepository) GetByID(ctx context.Context, id string) (*secondary.StuckRecord, error) {
	query := `SELECT id, patrol_id, first_check_id, check_count, status, resolved_at, created_at, updated_at
		FROM stucks WHERE id = ?`

	var stuck secondary.StuckRecord
	var firstCheckID, resolvedAt sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&stuck.ID,
		&stuck.PatrolID,
		&firstCheckID,
		&stuck.CheckCount,
		&stuck.Status,
		&resolvedAt,
		&stuck.CreatedAt,
		&stuck.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("stuck not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	stuck.FirstCheckID = firstCheckID.String
	stuck.ResolvedAt = resolvedAt.String

	return &stuck, nil
}

// GetByPatrol retrieves stucks by patrol ID.
func (r *StuckRepository) GetByPatrol(ctx context.Context, patrolID string) ([]*secondary.StuckRecord, error) {
	query := `SELECT id, patrol_id, first_check_id, check_count, status, resolved_at, created_at, updated_at
		FROM stucks WHERE patrol_id = ? ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, patrolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanStucks(rows)
}

// GetOpenByPatrol retrieves the open stuck for a patrol (if any).
func (r *StuckRepository) GetOpenByPatrol(ctx context.Context, patrolID string) (*secondary.StuckRecord, error) {
	query := `SELECT id, patrol_id, first_check_id, check_count, status, resolved_at, created_at, updated_at
		FROM stucks WHERE patrol_id = ? AND status = 'open' LIMIT 1`

	var stuck secondary.StuckRecord
	var firstCheckID, resolvedAt sql.NullString

	err := r.db.QueryRowContext(ctx, query, patrolID).Scan(
		&stuck.ID,
		&stuck.PatrolID,
		&firstCheckID,
		&stuck.CheckCount,
		&stuck.Status,
		&resolvedAt,
		&stuck.CreatedAt,
		&stuck.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // No open stuck is not an error
	}
	if err != nil {
		return nil, err
	}

	stuck.FirstCheckID = firstCheckID.String
	stuck.ResolvedAt = resolvedAt.String

	return &stuck, nil
}

// List retrieves stucks matching the given filters.
func (r *StuckRepository) List(ctx context.Context, filters secondary.StuckFilters) ([]*secondary.StuckRecord, error) {
	query := `SELECT id, patrol_id, first_check_id, check_count, status, resolved_at, created_at, updated_at
		FROM stucks WHERE 1=1`
	args := []any{}

	if filters.PatrolID != "" {
		query += " AND patrol_id = ?"
		args = append(args, filters.PatrolID)
	}
	if filters.Status != "" {
		query += " AND status = ?"
		args = append(args, filters.Status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanStucks(rows)
}

// Update updates an existing stuck.
func (r *StuckRepository) Update(ctx context.Context, stuck *secondary.StuckRecord) error {
	stuck.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	query := `UPDATE stucks SET patrol_id = ?, first_check_id = ?, check_count = ?,
		status = ?, resolved_at = ?, updated_at = ? WHERE id = ?`

	var firstCheckID, resolvedAt any
	if stuck.FirstCheckID != "" {
		firstCheckID = stuck.FirstCheckID
	}
	if stuck.ResolvedAt != "" {
		resolvedAt = stuck.ResolvedAt
	}

	result, err := r.db.ExecContext(ctx, query,
		stuck.PatrolID,
		firstCheckID,
		stuck.CheckCount,
		stuck.Status,
		resolvedAt,
		stuck.UpdatedAt,
		stuck.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("stuck not found: %s", stuck.ID)
	}

	return nil
}

// IncrementCount increments the check_count of a stuck.
func (r *StuckRepository) IncrementCount(ctx context.Context, id string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	query := `UPDATE stucks SET check_count = check_count + 1, updated_at = ? WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("stuck not found: %s", id)
	}

	return nil
}

// UpdateStatus updates the status of a stuck.
func (r *StuckRepository) UpdateStatus(ctx context.Context, id, status string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	query := `UPDATE stucks SET status = ?, updated_at = ? WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, status, now, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("stuck not found: %s", id)
	}

	return nil
}

// GetNextID returns the next available stuck ID.
func (r *StuckRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	query := `SELECT COALESCE(MAX(CAST(SUBSTR(id, 7) AS INTEGER)), 0) FROM stucks WHERE id LIKE 'STUCK-%'`
	if err := r.db.QueryRowContext(ctx, query).Scan(&maxID); err != nil {
		return "", err
	}
	return fmt.Sprintf("STUCK-%03d", maxID+1), nil
}

// PatrolExists checks if a patrol exists.
func (r *StuckRepository) PatrolExists(ctx context.Context, patrolID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM patrols WHERE id = ?)`
	if err := r.db.QueryRowContext(ctx, query, patrolID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// scanStucks scans rows into stuck records.
func (r *StuckRepository) scanStucks(rows *sql.Rows) ([]*secondary.StuckRecord, error) {
	var stucks []*secondary.StuckRecord

	for rows.Next() {
		var stuck secondary.StuckRecord
		var firstCheckID, resolvedAt sql.NullString

		if err := rows.Scan(
			&stuck.ID,
			&stuck.PatrolID,
			&firstCheckID,
			&stuck.CheckCount,
			&stuck.Status,
			&resolvedAt,
			&stuck.CreatedAt,
			&stuck.UpdatedAt,
		); err != nil {
			return nil, err
		}

		stuck.FirstCheckID = firstCheckID.String
		stuck.ResolvedAt = resolvedAt.String

		stucks = append(stucks, &stuck)
	}

	return stucks, rows.Err()
}

// Ensure StuckRepository implements the interface.
var _ secondary.StuckRepository = (*StuckRepository)(nil)
