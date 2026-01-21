// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	coremission "github.com/example/orc/internal/core/mission"
	"github.com/example/orc/internal/ports/secondary"
)

// MissionRepository implements secondary.MissionRepository with SQLite.
type MissionRepository struct {
	db *sql.DB
}

// NewMissionRepository creates a new SQLite mission repository.
func NewMissionRepository(db *sql.DB) *MissionRepository {
	return &MissionRepository{db: db}
}

// Create persists a new mission.
// The mission record must have ID and Status pre-populated by the service layer.
func (r *MissionRepository) Create(ctx context.Context, mission *secondary.MissionRecord) error {
	if mission.ID == "" {
		return fmt.Errorf("mission ID must be pre-populated by service layer")
	}
	if mission.Status == "" {
		return fmt.Errorf("mission Status must be pre-populated by service layer")
	}

	var desc sql.NullString
	if mission.Description != "" {
		desc = sql.NullString{String: mission.Description, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO missions (id, title, description, status) VALUES (?, ?, ?, ?)",
		mission.ID, mission.Title, desc, mission.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to create mission: %w", err)
	}

	return nil
}

// GetByID retrieves a mission by its ID.
func (r *MissionRepository) GetByID(ctx context.Context, id string) (*secondary.MissionRecord, error) {
	var (
		desc        sql.NullString
		pinned      bool
		createdAt   time.Time
		updatedAt   time.Time
		completedAt sql.NullTime
	)

	record := &secondary.MissionRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, title, description, status, pinned, created_at, updated_at, completed_at FROM missions WHERE id = ?",
		id,
	).Scan(&record.ID, &record.Title, &desc, &record.Status, &pinned, &createdAt, &updatedAt, &completedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("mission %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get mission: %w", err)
	}

	record.Description = desc.String
	record.Pinned = pinned
	record.CreatedAt = createdAt.Format(time.RFC3339)
	if completedAt.Valid {
		record.CompletedAt = completedAt.Time.Format(time.RFC3339)
	}

	return record, nil
}

// List retrieves missions matching the given filters.
func (r *MissionRepository) List(ctx context.Context, filters secondary.MissionFilters) ([]*secondary.MissionRecord, error) {
	query := "SELECT id, title, description, status, pinned, created_at, updated_at, completed_at FROM missions"
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
		return nil, fmt.Errorf("failed to list missions: %w", err)
	}
	defer rows.Close()

	var missions []*secondary.MissionRecord
	for rows.Next() {
		var (
			desc        sql.NullString
			pinned      bool
			createdAt   time.Time
			updatedAt   time.Time
			completedAt sql.NullTime
		)

		record := &secondary.MissionRecord{}
		err := rows.Scan(&record.ID, &record.Title, &desc, &record.Status, &pinned, &createdAt, &updatedAt, &completedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan mission: %w", err)
		}

		record.Description = desc.String
		record.Pinned = pinned
		record.CreatedAt = createdAt.Format(time.RFC3339)
		if completedAt.Valid {
			record.CompletedAt = completedAt.Time.Format(time.RFC3339)
		}

		missions = append(missions, record)
	}

	return missions, nil
}

// Update updates an existing mission.
// The service layer is responsible for setting CompletedAt when status changes to complete.
func (r *MissionRepository) Update(ctx context.Context, mission *secondary.MissionRecord) error {
	// Build dynamic query based on what's being updated
	query := "UPDATE missions SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if mission.Title != "" {
		query += ", title = ?"
		args = append(args, mission.Title)
	}

	if mission.Description != "" {
		query += ", description = ?"
		args = append(args, sql.NullString{String: mission.Description, Valid: true})
	}

	if mission.Status != "" {
		query += ", status = ?"
		args = append(args, mission.Status)
	}

	// CompletedAt is set by service layer when transitioning to complete status
	if mission.CompletedAt != "" {
		completedTime, err := time.Parse(time.RFC3339, mission.CompletedAt)
		if err == nil {
			query += ", completed_at = ?"
			args = append(args, sql.NullTime{Time: completedTime, Valid: true})
		}
	}

	query += " WHERE id = ?"
	args = append(args, mission.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update mission: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("mission %s not found", mission.ID)
	}

	return nil
}

// Delete removes a mission from persistence.
func (r *MissionRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM missions WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete mission: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("mission %s not found", id)
	}

	return nil
}

// GetNextID returns the next available mission ID.
// Uses core function for ID format to keep business logic in the functional core.
func (r *MissionRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 9) AS INTEGER)), 0) FROM missions",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next mission ID: %w", err)
	}

	return coremission.GenerateMissionID(maxID), nil
}

// CountShipments returns the number of shipments for a mission.
func (r *MissionRepository) CountShipments(ctx context.Context, missionID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM shipments WHERE mission_id = ?",
		missionID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count shipments: %w", err)
	}

	return count, nil
}

// Pin pins a mission to keep it visible.
func (r *MissionRepository) Pin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE missions SET pinned = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to pin mission: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("mission %s not found", id)
	}

	return nil
}

// Unpin unpins a mission.
func (r *MissionRepository) Unpin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE missions SET pinned = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to unpin mission: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("mission %s not found", id)
	}

	return nil
}

// Ensure MissionRepository implements the interface
var _ secondary.MissionRepository = (*MissionRepository)(nil)
