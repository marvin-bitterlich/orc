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

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO tomes (id, mission_id, title, description, status) VALUES (?, ?, ?, ?, ?)",
		tome.ID, tome.MissionID, tome.Title, desc, "active",
	)
	if err != nil {
		return fmt.Errorf("failed to create tome: %w", err)
	}

	return nil
}

// GetByID retrieves a tome by its ID.
func (r *TomeRepository) GetByID(ctx context.Context, id string) (*secondary.TomeRecord, error) {
	var (
		desc            sql.NullString
		assignedGroveID sql.NullString
		pinned          bool
		createdAt       time.Time
		updatedAt       time.Time
		completedAt     sql.NullTime
	)

	record := &secondary.TomeRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, mission_id, title, description, status, assigned_grove_id, pinned, created_at, updated_at, completed_at FROM tomes WHERE id = ?",
		id,
	).Scan(&record.ID, &record.MissionID, &record.Title, &desc, &record.Status, &assignedGroveID, &pinned, &createdAt, &updatedAt, &completedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tome %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tome: %w", err)
	}

	record.Description = desc.String
	record.AssignedGroveID = assignedGroveID.String
	record.Pinned = pinned
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	if completedAt.Valid {
		record.CompletedAt = completedAt.Time.Format(time.RFC3339)
	}

	return record, nil
}

// List retrieves tomes matching the given filters.
func (r *TomeRepository) List(ctx context.Context, filters secondary.TomeFilters) ([]*secondary.TomeRecord, error) {
	query := "SELECT id, mission_id, title, description, status, assigned_grove_id, pinned, created_at, updated_at, completed_at FROM tomes WHERE 1=1"
	args := []any{}

	if filters.MissionID != "" {
		query += " AND mission_id = ?"
		args = append(args, filters.MissionID)
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
			desc            sql.NullString
			assignedGroveID sql.NullString
			pinned          bool
			createdAt       time.Time
			updatedAt       time.Time
			completedAt     sql.NullTime
		)

		record := &secondary.TomeRecord{}
		err := rows.Scan(&record.ID, &record.MissionID, &record.Title, &desc, &record.Status, &assignedGroveID, &pinned, &createdAt, &updatedAt, &completedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tome: %w", err)
		}

		record.Description = desc.String
		record.AssignedGroveID = assignedGroveID.String
		record.Pinned = pinned
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if completedAt.Valid {
			record.CompletedAt = completedAt.Time.Format(time.RFC3339)
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

// UpdateStatus updates the status and optionally completed_at timestamp.
func (r *TomeRepository) UpdateStatus(ctx context.Context, id, status string, setCompleted bool) error {
	var query string
	var args []any

	if setCompleted {
		query = "UPDATE tomes SET status = ?, completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
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

// GetByGrove retrieves tomes assigned to a grove.
func (r *TomeRepository) GetByGrove(ctx context.Context, groveID string) ([]*secondary.TomeRecord, error) {
	query := "SELECT id, mission_id, title, description, status, assigned_grove_id, pinned, created_at, updated_at, completed_at FROM tomes WHERE assigned_grove_id = ?"
	rows, err := r.db.QueryContext(ctx, query, groveID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tomes by grove: %w", err)
	}
	defer rows.Close()

	var tomes []*secondary.TomeRecord
	for rows.Next() {
		var (
			desc            sql.NullString
			assignedGroveID sql.NullString
			pinned          bool
			createdAt       time.Time
			updatedAt       time.Time
			completedAt     sql.NullTime
		)

		record := &secondary.TomeRecord{}
		err := rows.Scan(&record.ID, &record.MissionID, &record.Title, &desc, &record.Status, &assignedGroveID, &pinned, &createdAt, &updatedAt, &completedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tome: %w", err)
		}

		record.Description = desc.String
		record.AssignedGroveID = assignedGroveID.String
		record.Pinned = pinned
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if completedAt.Valid {
			record.CompletedAt = completedAt.Time.Format(time.RFC3339)
		}

		tomes = append(tomes, record)
	}

	return tomes, nil
}

// AssignGrove assigns a tome to a grove.
func (r *TomeRepository) AssignGrove(ctx context.Context, tomeID, groveID string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE tomes SET assigned_grove_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		groveID, tomeID,
	)
	if err != nil {
		return fmt.Errorf("failed to assign grove to tome: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("tome %s not found", tomeID)
	}

	return nil
}

// MissionExists checks if a mission exists.
func (r *TomeRepository) MissionExists(ctx context.Context, missionID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM missions WHERE id = ?", missionID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check mission existence: %w", err)
	}
	return count > 0, nil
}

// Ensure TomeRepository implements the interface
var _ secondary.TomeRepository = (*TomeRepository)(nil)
