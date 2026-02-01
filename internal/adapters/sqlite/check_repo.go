package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// CheckRepository implements secondary.CheckRepository using SQLite.
type CheckRepository struct {
	db *sql.DB
}

// NewCheckRepository creates a new CheckRepository.
func NewCheckRepository(db *sql.DB) *CheckRepository {
	return &CheckRepository{db: db}
}

// Create persists a new check.
func (r *CheckRepository) Create(ctx context.Context, check *secondary.CheckRecord) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if check.CapturedAt == "" {
		check.CapturedAt = now
	}
	if check.CreatedAt == "" {
		check.CreatedAt = now
	}

	query := `INSERT INTO checks (id, patrol_id, stuck_id, pane_content, outcome, captured_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	var stuckID interface{}
	if check.StuckID != "" {
		stuckID = check.StuckID
	}

	_, err := r.db.ExecContext(ctx, query,
		check.ID,
		check.PatrolID,
		stuckID,
		check.PaneContent,
		check.Outcome,
		check.CapturedAt,
		check.CreatedAt,
	)
	return err
}

// GetByID retrieves a check by its ID.
func (r *CheckRepository) GetByID(ctx context.Context, id string) (*secondary.CheckRecord, error) {
	query := `SELECT id, patrol_id, stuck_id, pane_content, outcome, captured_at, created_at
		FROM checks WHERE id = ?`

	var check secondary.CheckRecord
	var stuckID, paneContent sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&check.ID,
		&check.PatrolID,
		&stuckID,
		&paneContent,
		&check.Outcome,
		&check.CapturedAt,
		&check.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("check not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	check.StuckID = stuckID.String
	check.PaneContent = paneContent.String

	return &check, nil
}

// GetByPatrol retrieves checks by patrol ID.
func (r *CheckRepository) GetByPatrol(ctx context.Context, patrolID string) ([]*secondary.CheckRecord, error) {
	query := `SELECT id, patrol_id, stuck_id, pane_content, outcome, captured_at, created_at
		FROM checks WHERE patrol_id = ? ORDER BY captured_at DESC`

	rows, err := r.db.QueryContext(ctx, query, patrolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanChecks(rows)
}

// GetLatest retrieves the latest check for a patrol.
func (r *CheckRepository) GetLatest(ctx context.Context, patrolID string) (*secondary.CheckRecord, error) {
	query := `SELECT id, patrol_id, stuck_id, pane_content, outcome, captured_at, created_at
		FROM checks WHERE patrol_id = ? ORDER BY captured_at DESC LIMIT 1`

	var check secondary.CheckRecord
	var stuckID, paneContent sql.NullString

	err := r.db.QueryRowContext(ctx, query, patrolID).Scan(
		&check.ID,
		&check.PatrolID,
		&stuckID,
		&paneContent,
		&check.Outcome,
		&check.CapturedAt,
		&check.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // No checks yet is not an error
	}
	if err != nil {
		return nil, err
	}

	check.StuckID = stuckID.String
	check.PaneContent = paneContent.String

	return &check, nil
}

// List retrieves checks matching the given filters.
func (r *CheckRepository) List(ctx context.Context, filters secondary.CheckFilters) ([]*secondary.CheckRecord, error) {
	query := `SELECT id, patrol_id, stuck_id, pane_content, outcome, captured_at, created_at
		FROM checks WHERE 1=1`
	args := []interface{}{}

	if filters.PatrolID != "" {
		query += " AND patrol_id = ?"
		args = append(args, filters.PatrolID)
	}
	if filters.StuckID != "" {
		query += " AND stuck_id = ?"
		args = append(args, filters.StuckID)
	}
	if filters.Outcome != "" {
		query += " AND outcome = ?"
		args = append(args, filters.Outcome)
	}

	query += " ORDER BY captured_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanChecks(rows)
}

// GetNextID returns the next available check ID.
func (r *CheckRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	query := `SELECT COALESCE(MAX(CAST(SUBSTR(id, 7) AS INTEGER)), 0) FROM checks WHERE id LIKE 'CHECK-%'`
	if err := r.db.QueryRowContext(ctx, query).Scan(&maxID); err != nil {
		return "", err
	}
	return fmt.Sprintf("CHECK-%03d", maxID+1), nil
}

// PatrolExists checks if a patrol exists.
func (r *CheckRepository) PatrolExists(ctx context.Context, patrolID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM patrols WHERE id = ?)`
	if err := r.db.QueryRowContext(ctx, query, patrolID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// scanChecks scans rows into check records.
func (r *CheckRepository) scanChecks(rows *sql.Rows) ([]*secondary.CheckRecord, error) {
	var checks []*secondary.CheckRecord

	for rows.Next() {
		var check secondary.CheckRecord
		var stuckID, paneContent sql.NullString

		if err := rows.Scan(
			&check.ID,
			&check.PatrolID,
			&stuckID,
			&paneContent,
			&check.Outcome,
			&check.CapturedAt,
			&check.CreatedAt,
		); err != nil {
			return nil, err
		}

		check.StuckID = stuckID.String
		check.PaneContent = paneContent.String

		checks = append(checks, &check)
	}

	return checks, rows.Err()
}

// Ensure CheckRepository implements the interface.
var _ secondary.CheckRepository = (*CheckRepository)(nil)
