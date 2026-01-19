// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// GroveRepository implements secondary.GroveRepository with SQLite.
type GroveRepository struct {
	db *sql.DB
}

// NewGroveRepository creates a new SQLite grove repository.
func NewGroveRepository(db *sql.DB) *GroveRepository {
	return &GroveRepository{db: db}
}

// Create persists a new grove.
func (r *GroveRepository) Create(ctx context.Context, grove *secondary.GroveRecord) error {
	// Verify mission exists
	var exists int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM missions WHERE id = ?", grove.MissionID,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to verify mission: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("mission %s not found", grove.MissionID)
	}

	// Generate grove ID by finding max existing ID
	var maxID int
	err = r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 7) AS INTEGER)), 0) FROM groves",
	).Scan(&maxID)
	if err != nil {
		return fmt.Errorf("failed to generate grove ID: %w", err)
	}

	id := fmt.Sprintf("GROVE-%03d", maxID+1)

	_, err = r.db.ExecContext(ctx,
		"INSERT INTO groves (id, mission_id, name, path, status) VALUES (?, ?, ?, ?, ?)",
		id, grove.MissionID, grove.Name, grove.WorktreePath, "active",
	)
	if err != nil {
		return fmt.Errorf("failed to create grove: %w", err)
	}

	// Update the record with the generated ID
	grove.ID = id
	return nil
}

// GetByID retrieves a grove by its ID.
func (r *GroveRepository) GetByID(ctx context.Context, id string) (*secondary.GroveRecord, error) {
	var (
		createdAt time.Time
		repos     sql.NullString
	)

	record := &secondary.GroveRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, mission_id, name, path, status, created_at FROM groves WHERE id = ?",
		id,
	).Scan(&record.ID, &record.MissionID, &record.Name, &record.WorktreePath, &record.Status, &createdAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("grove %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get grove: %w", err)
	}

	// Silence unused variable warning
	_ = repos

	record.CreatedAt = createdAt.Format(time.RFC3339)
	return record, nil
}

// GetByPath retrieves a grove by its file path.
func (r *GroveRepository) GetByPath(ctx context.Context, path string) (*secondary.GroveRecord, error) {
	var createdAt time.Time

	record := &secondary.GroveRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, mission_id, name, path, status, created_at FROM groves WHERE path = ?",
		path,
	).Scan(&record.ID, &record.MissionID, &record.Name, &record.WorktreePath, &record.Status, &createdAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("grove with path %s not found", path)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get grove by path: %w", err)
	}

	record.CreatedAt = createdAt.Format(time.RFC3339)
	return record, nil
}

// GetByMission retrieves all groves for a mission.
func (r *GroveRepository) GetByMission(ctx context.Context, missionID string) ([]*secondary.GroveRecord, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, mission_id, name, path, status, created_at FROM groves WHERE mission_id = ? AND status = 'active' ORDER BY created_at DESC",
		missionID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list groves: %w", err)
	}
	defer rows.Close()

	var groves []*secondary.GroveRecord
	for rows.Next() {
		var createdAt time.Time
		record := &secondary.GroveRecord{}

		err := rows.Scan(&record.ID, &record.MissionID, &record.Name, &record.WorktreePath, &record.Status, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan grove: %w", err)
		}

		record.CreatedAt = createdAt.Format(time.RFC3339)
		groves = append(groves, record)
	}

	return groves, nil
}

// List retrieves all groves, optionally filtered by mission.
func (r *GroveRepository) List(ctx context.Context, missionID string) ([]*secondary.GroveRecord, error) {
	query := "SELECT id, mission_id, name, path, status, created_at FROM groves WHERE 1=1"
	args := []any{}

	if missionID != "" {
		query += " AND mission_id = ?"
		args = append(args, missionID)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list groves: %w", err)
	}
	defer rows.Close()

	var groves []*secondary.GroveRecord
	for rows.Next() {
		var createdAt time.Time
		record := &secondary.GroveRecord{}

		err := rows.Scan(&record.ID, &record.MissionID, &record.Name, &record.WorktreePath, &record.Status, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan grove: %w", err)
		}

		record.CreatedAt = createdAt.Format(time.RFC3339)
		groves = append(groves, record)
	}

	return groves, nil
}

// Update updates an existing grove.
func (r *GroveRepository) Update(ctx context.Context, grove *secondary.GroveRecord) error {
	query := "UPDATE groves SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if grove.Name != "" {
		query += ", name = ?"
		args = append(args, grove.Name)
	}

	if grove.WorktreePath != "" {
		query += ", path = ?"
		args = append(args, grove.WorktreePath)
	}

	if grove.Status != "" {
		query += ", status = ?"
		args = append(args, grove.Status)
	}

	query += " WHERE id = ?"
	args = append(args, grove.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update grove: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("grove %s not found", grove.ID)
	}

	return nil
}

// Delete removes a grove from persistence.
func (r *GroveRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM groves WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete grove: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("grove %s not found", id)
	}

	return nil
}

// GetNextID returns the next available grove ID.
func (r *GroveRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 7) AS INTEGER)), 0) FROM groves",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next grove ID: %w", err)
	}

	return fmt.Sprintf("GROVE-%03d", maxID+1), nil
}

// Rename updates the name of a grove.
func (r *GroveRepository) Rename(ctx context.Context, id, newName string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE groves SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		newName, id,
	)
	if err != nil {
		return fmt.Errorf("failed to rename grove: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("grove %s not found", id)
	}

	return nil
}

// UpdatePath updates the path of a grove.
func (r *GroveRepository) UpdatePath(ctx context.Context, id, newPath string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE groves SET path = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		newPath, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update grove path: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("grove %s not found", id)
	}

	return nil
}

// Ensure GroveRepository implements the interface
var _ secondary.GroveRepository = (*GroveRepository)(nil)
