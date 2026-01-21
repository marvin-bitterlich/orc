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
	var missionID, groveID, todos sql.NullString

	if handoff.ActiveMissionID != "" {
		missionID = sql.NullString{String: handoff.ActiveMissionID, Valid: true}
	}
	if handoff.ActiveGroveID != "" {
		groveID = sql.NullString{String: handoff.ActiveGroveID, Valid: true}
	}
	if handoff.TodosSnapshot != "" {
		todos = sql.NullString{String: handoff.TodosSnapshot, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO handoffs (id, handoff_note, active_mission_id, active_grove_id, todos_snapshot)
		 VALUES (?, ?, ?, ?, ?)`,
		handoff.ID, handoff.HandoffNote, missionID, groveID, todos,
	)
	if err != nil {
		return fmt.Errorf("failed to create handoff: %w", err)
	}

	return nil
}

// GetByID retrieves a handoff by its ID.
func (r *HandoffRepository) GetByID(ctx context.Context, id string) (*secondary.HandoffRecord, error) {
	var (
		createdAt time.Time
		missionID sql.NullString
		groveID   sql.NullString
		todos     sql.NullString
	)

	record := &secondary.HandoffRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, created_at, handoff_note, active_mission_id, active_grove_id, todos_snapshot
		 FROM handoffs WHERE id = ?`,
		id,
	).Scan(&record.ID, &createdAt, &record.HandoffNote, &missionID, &groveID, &todos)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("handoff %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get handoff: %w", err)
	}

	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.ActiveMissionID = missionID.String
	record.ActiveGroveID = groveID.String
	record.TodosSnapshot = todos.String

	return record, nil
}

// GetLatest retrieves the most recent handoff.
func (r *HandoffRepository) GetLatest(ctx context.Context) (*secondary.HandoffRecord, error) {
	var (
		createdAt time.Time
		missionID sql.NullString
		groveID   sql.NullString
		todos     sql.NullString
	)

	record := &secondary.HandoffRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, created_at, handoff_note, active_mission_id, active_grove_id, todos_snapshot
		 FROM handoffs ORDER BY created_at DESC LIMIT 1`,
	).Scan(&record.ID, &createdAt, &record.HandoffNote, &missionID, &groveID, &todos)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no handoffs found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest handoff: %w", err)
	}

	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.ActiveMissionID = missionID.String
	record.ActiveGroveID = groveID.String
	record.TodosSnapshot = todos.String

	return record, nil
}

// GetLatestForGrove retrieves the most recent handoff for a grove.
func (r *HandoffRepository) GetLatestForGrove(ctx context.Context, groveID string) (*secondary.HandoffRecord, error) {
	var (
		createdAt time.Time
		missionID sql.NullString
		grove     sql.NullString
		todos     sql.NullString
	)

	record := &secondary.HandoffRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, created_at, handoff_note, active_mission_id, active_grove_id, todos_snapshot
		 FROM handoffs WHERE active_grove_id = ? ORDER BY created_at DESC LIMIT 1`,
		groveID,
	).Scan(&record.ID, &createdAt, &record.HandoffNote, &missionID, &grove, &todos)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no handoffs found for grove %s", groveID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest handoff for grove: %w", err)
	}

	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.ActiveMissionID = missionID.String
	record.ActiveGroveID = grove.String
	record.TodosSnapshot = todos.String

	return record, nil
}

// List retrieves handoffs with optional limit.
func (r *HandoffRepository) List(ctx context.Context, limit int) ([]*secondary.HandoffRecord, error) {
	query := `SELECT id, created_at, handoff_note, active_mission_id, active_grove_id, todos_snapshot
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
			createdAt time.Time
			missionID sql.NullString
			groveID   sql.NullString
			todos     sql.NullString
		)

		record := &secondary.HandoffRecord{}
		err := rows.Scan(&record.ID, &createdAt, &record.HandoffNote, &missionID, &groveID, &todos)
		if err != nil {
			return nil, fmt.Errorf("failed to scan handoff: %w", err)
		}

		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.ActiveMissionID = missionID.String
		record.ActiveGroveID = groveID.String
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
