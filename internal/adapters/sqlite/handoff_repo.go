// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// HandoffRepository implements secondary.HandoffRepository with SQLite.
// Handoffs are immutable - no Update or Delete operations.
type HandoffRepository struct {
	db *sql.DB
}

// NewHandoffRepository creates a new SQLite handoff repository.
func NewHandoffRepository(db *sql.DB) *HandoffRepository {
	return &HandoffRepository{db: db}
}

// Create persists a new handoff.
func (r *HandoffRepository) Create(ctx context.Context, handoff *secondary.HandoffRecord) error {
	var commissionID, workbenchID, todos sql.NullString

	if handoff.ActiveCommissionID != "" {
		commissionID = sql.NullString{String: handoff.ActiveCommissionID, Valid: true}
	}
	if handoff.ActiveWorkbenchID != "" {
		workbenchID = sql.NullString{String: handoff.ActiveWorkbenchID, Valid: true}
	}
	if handoff.TodosSnapshot != "" {
		todos = sql.NullString{String: handoff.TodosSnapshot, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO handoffs (id, handoff_note, active_commission_id, active_workbench_id, todos_snapshot)
		 VALUES (?, ?, ?, ?, ?)`,
		handoff.ID, handoff.HandoffNote, commissionID, workbenchID, todos,
	)
	if err != nil {
		return fmt.Errorf("failed to create handoff: %w", err)
	}

	return nil
}

// GetByID retrieves a handoff by its ID.
func (r *HandoffRepository) GetByID(ctx context.Context, id string) (*secondary.HandoffRecord, error) {
	var (
		createdAt    time.Time
		commissionID sql.NullString
		workbenchID  sql.NullString
		todos        sql.NullString
	)

	record := &secondary.HandoffRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, created_at, handoff_note, active_commission_id, active_workbench_id, todos_snapshot
		 FROM handoffs WHERE id = ?`,
		id,
	).Scan(&record.ID, &createdAt, &record.HandoffNote, &commissionID, &workbenchID, &todos)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("handoff %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get handoff: %w", err)
	}

	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.ActiveCommissionID = commissionID.String
	record.ActiveWorkbenchID = workbenchID.String
	record.TodosSnapshot = todos.String

	return record, nil
}

// GetLatest retrieves the most recent handoff.
func (r *HandoffRepository) GetLatest(ctx context.Context) (*secondary.HandoffRecord, error) {
	var (
		createdAt    time.Time
		commissionID sql.NullString
		workbenchID  sql.NullString
		todos        sql.NullString
	)

	record := &secondary.HandoffRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, created_at, handoff_note, active_commission_id, active_workbench_id, todos_snapshot
		 FROM handoffs ORDER BY created_at DESC LIMIT 1`,
	).Scan(&record.ID, &createdAt, &record.HandoffNote, &commissionID, &workbenchID, &todos)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no handoffs found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest handoff: %w", err)
	}

	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.ActiveCommissionID = commissionID.String
	record.ActiveWorkbenchID = workbenchID.String
	record.TodosSnapshot = todos.String

	return record, nil
}

// GetLatestForWorkbench retrieves the most recent handoff for a workbench.
func (r *HandoffRepository) GetLatestForWorkbench(ctx context.Context, workbenchID string) (*secondary.HandoffRecord, error) {
	var (
		createdAt    time.Time
		commissionID sql.NullString
		wbID         sql.NullString
		todos        sql.NullString
	)

	record := &secondary.HandoffRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, created_at, handoff_note, active_commission_id, active_workbench_id, todos_snapshot
		 FROM handoffs WHERE active_workbench_id = ? ORDER BY created_at DESC LIMIT 1`,
		workbenchID,
	).Scan(&record.ID, &createdAt, &record.HandoffNote, &commissionID, &wbID, &todos)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no handoffs found for workbench %s", workbenchID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest handoff for workbench: %w", err)
	}

	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.ActiveCommissionID = commissionID.String
	record.ActiveWorkbenchID = wbID.String
	record.TodosSnapshot = todos.String

	return record, nil
}

// List retrieves handoffs with optional limit.
func (r *HandoffRepository) List(ctx context.Context, limit int) ([]*secondary.HandoffRecord, error) {
	query := `SELECT id, created_at, handoff_note, active_commission_id, active_workbench_id, todos_snapshot
	          FROM handoffs ORDER BY created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list handoffs: %w", err)
	}
	defer rows.Close()

	var handoffs []*secondary.HandoffRecord
	for rows.Next() {
		var (
			createdAt    time.Time
			commissionID sql.NullString
			workbenchID  sql.NullString
			todos        sql.NullString
		)

		record := &secondary.HandoffRecord{}
		err := rows.Scan(&record.ID, &createdAt, &record.HandoffNote, &commissionID, &workbenchID, &todos)
		if err != nil {
			return nil, fmt.Errorf("failed to scan handoff: %w", err)
		}

		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.ActiveCommissionID = commissionID.String
		record.ActiveWorkbenchID = workbenchID.String
		record.TodosSnapshot = todos.String

		handoffs = append(handoffs, record)
	}

	return handoffs, nil
}

// GetNextID returns the next available handoff ID.
func (r *HandoffRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 4) AS INTEGER)), 0) FROM handoffs",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next handoff ID: %w", err)
	}

	return fmt.Sprintf("HO-%03d", maxID+1), nil
}

// Ensure HandoffRepository implements the interface
var _ secondary.HandoffRepository = (*HandoffRepository)(nil)
