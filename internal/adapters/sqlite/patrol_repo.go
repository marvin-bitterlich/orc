package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// PatrolRepository implements secondary.PatrolRepository using SQLite.
type PatrolRepository struct {
	db *sql.DB
}

// NewPatrolRepository creates a new PatrolRepository.
func NewPatrolRepository(db *sql.DB) *PatrolRepository {
	return &PatrolRepository{db: db}
}

// Create persists a new patrol.
func (r *PatrolRepository) Create(ctx context.Context, patrol *secondary.PatrolRecord) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if patrol.CreatedAt == "" {
		patrol.CreatedAt = now
	}
	if patrol.UpdatedAt == "" {
		patrol.UpdatedAt = now
	}

	query := `INSERT INTO patrols (id, kennel_id, target, status, config, started_at, ended_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	var startedAt, endedAt, config interface{}
	if patrol.StartedAt != "" {
		startedAt = patrol.StartedAt
	}
	if patrol.EndedAt != "" {
		endedAt = patrol.EndedAt
	}
	if patrol.Config != "" {
		config = patrol.Config
	}

	_, err := r.db.ExecContext(ctx, query,
		patrol.ID,
		patrol.KennelID,
		patrol.Target,
		patrol.Status,
		config,
		startedAt,
		endedAt,
		patrol.CreatedAt,
		patrol.UpdatedAt,
	)
	return err
}

// GetByID retrieves a patrol by its ID.
func (r *PatrolRepository) GetByID(ctx context.Context, id string) (*secondary.PatrolRecord, error) {
	query := `SELECT id, kennel_id, target, status, config, started_at, ended_at, created_at, updated_at
		FROM patrols WHERE id = ?`

	var patrol secondary.PatrolRecord
	var config, startedAt, endedAt sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&patrol.ID,
		&patrol.KennelID,
		&patrol.Target,
		&patrol.Status,
		&config,
		&startedAt,
		&endedAt,
		&patrol.CreatedAt,
		&patrol.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("patrol not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	patrol.Config = config.String
	patrol.StartedAt = startedAt.String
	patrol.EndedAt = endedAt.String

	return &patrol, nil
}

// GetByKennel retrieves patrols by kennel ID.
func (r *PatrolRepository) GetByKennel(ctx context.Context, kennelID string) ([]*secondary.PatrolRecord, error) {
	query := `SELECT id, kennel_id, target, status, config, started_at, ended_at, created_at, updated_at
		FROM patrols WHERE kennel_id = ? ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, kennelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPatrols(rows)
}

// GetActiveByKennel retrieves the active patrol for a kennel.
func (r *PatrolRepository) GetActiveByKennel(ctx context.Context, kennelID string) (*secondary.PatrolRecord, error) {
	query := `SELECT id, kennel_id, target, status, config, started_at, ended_at, created_at, updated_at
		FROM patrols WHERE kennel_id = ? AND status = 'active' LIMIT 1`

	var patrol secondary.PatrolRecord
	var config, startedAt, endedAt sql.NullString

	err := r.db.QueryRowContext(ctx, query, kennelID).Scan(
		&patrol.ID,
		&patrol.KennelID,
		&patrol.Target,
		&patrol.Status,
		&config,
		&startedAt,
		&endedAt,
		&patrol.CreatedAt,
		&patrol.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // No active patrol is not an error
	}
	if err != nil {
		return nil, err
	}

	patrol.Config = config.String
	patrol.StartedAt = startedAt.String
	patrol.EndedAt = endedAt.String

	return &patrol, nil
}

// List retrieves patrols matching the given filters.
func (r *PatrolRepository) List(ctx context.Context, filters secondary.PatrolFilters) ([]*secondary.PatrolRecord, error) {
	query := `SELECT id, kennel_id, target, status, config, started_at, ended_at, created_at, updated_at
		FROM patrols WHERE 1=1`
	args := []interface{}{}

	if filters.KennelID != "" {
		query += " AND kennel_id = ?"
		args = append(args, filters.KennelID)
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

	return r.scanPatrols(rows)
}

// Update updates an existing patrol.
func (r *PatrolRepository) Update(ctx context.Context, patrol *secondary.PatrolRecord) error {
	patrol.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	query := `UPDATE patrols SET kennel_id = ?, target = ?, status = ?, config = ?,
		started_at = ?, ended_at = ?, updated_at = ? WHERE id = ?`

	var startedAt, endedAt, config interface{}
	if patrol.StartedAt != "" {
		startedAt = patrol.StartedAt
	}
	if patrol.EndedAt != "" {
		endedAt = patrol.EndedAt
	}
	if patrol.Config != "" {
		config = patrol.Config
	}

	result, err := r.db.ExecContext(ctx, query,
		patrol.KennelID,
		patrol.Target,
		patrol.Status,
		config,
		startedAt,
		endedAt,
		patrol.UpdatedAt,
		patrol.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("patrol not found: %s", patrol.ID)
	}

	return nil
}

// UpdateStatus updates the status of a patrol.
func (r *PatrolRepository) UpdateStatus(ctx context.Context, id, status string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	query := `UPDATE patrols SET status = ?, updated_at = ? WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, status, now, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("patrol not found: %s", id)
	}

	return nil
}

// GetNextID returns the next available patrol ID.
func (r *PatrolRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	query := `SELECT COALESCE(MAX(CAST(SUBSTR(id, 8) AS INTEGER)), 0) FROM patrols WHERE id LIKE 'PATROL-%'`
	if err := r.db.QueryRowContext(ctx, query).Scan(&maxID); err != nil {
		return "", err
	}
	return fmt.Sprintf("PATROL-%03d", maxID+1), nil
}

// KennelExists checks if a kennel exists.
func (r *PatrolRepository) KennelExists(ctx context.Context, kennelID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM kennels WHERE id = ?)`
	if err := r.db.QueryRowContext(ctx, query, kennelID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// scanPatrols scans rows into patrol records.
func (r *PatrolRepository) scanPatrols(rows *sql.Rows) ([]*secondary.PatrolRecord, error) {
	var patrols []*secondary.PatrolRecord

	for rows.Next() {
		var patrol secondary.PatrolRecord
		var config, startedAt, endedAt sql.NullString

		if err := rows.Scan(
			&patrol.ID,
			&patrol.KennelID,
			&patrol.Target,
			&patrol.Status,
			&config,
			&startedAt,
			&endedAt,
			&patrol.CreatedAt,
			&patrol.UpdatedAt,
		); err != nil {
			return nil, err
		}

		patrol.Config = config.String
		patrol.StartedAt = startedAt.String
		patrol.EndedAt = endedAt.String

		patrols = append(patrols, &patrol)
	}

	return patrols, rows.Err()
}

// Ensure PatrolRepository implements the interface.
var _ secondary.PatrolRepository = (*PatrolRepository)(nil)
