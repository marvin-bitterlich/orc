// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// WorkshopLogRepository implements secondary.WorkshopLogRepository with SQLite.
type WorkshopLogRepository struct {
	db *sql.DB
}

// NewWorkshopLogRepository creates a new SQLite workshop log repository.
func NewWorkshopLogRepository(db *sql.DB) *WorkshopLogRepository {
	return &WorkshopLogRepository{db: db}
}

// Create persists a new workshop log entry.
func (r *WorkshopLogRepository) Create(ctx context.Context, log *secondary.WorkshopLogRecord) error {
	var actorID, fieldName, oldValue, newValue sql.NullString
	if log.ActorID != "" {
		actorID = sql.NullString{String: log.ActorID, Valid: true}
	}
	if log.FieldName != "" {
		fieldName = sql.NullString{String: log.FieldName, Valid: true}
	}
	if log.OldValue != "" {
		oldValue = sql.NullString{String: log.OldValue, Valid: true}
	}
	if log.NewValue != "" {
		newValue = sql.NullString{String: log.NewValue, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO workshop_logs (id, workshop_id, actor_id, entity_type, entity_id, action, field_name, old_value, new_value) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		log.ID,
		log.WorkshopID,
		actorID,
		log.EntityType,
		log.EntityID,
		log.Action,
		fieldName,
		oldValue,
		newValue,
	)
	if err != nil {
		return fmt.Errorf("failed to create workshop log: %w", err)
	}

	return nil
}

// GetByID retrieves a log entry by its ID.
func (r *WorkshopLogRepository) GetByID(ctx context.Context, id string) (*secondary.WorkshopLogRecord, error) {
	var (
		actorID   sql.NullString
		fieldName sql.NullString
		oldValue  sql.NullString
		newValue  sql.NullString
		timestamp time.Time
		createdAt time.Time
	)

	record := &secondary.WorkshopLogRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, workshop_id, timestamp, actor_id, entity_type, entity_id, action, field_name, old_value, new_value, created_at FROM workshop_logs WHERE id = ?`,
		id,
	).Scan(&record.ID,
		&record.WorkshopID,
		&timestamp,
		&actorID,
		&record.EntityType,
		&record.EntityID,
		&record.Action,
		&fieldName,
		&oldValue,
		&newValue,
		&createdAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("workshop log %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workshop log: %w", err)
	}
	record.Timestamp = timestamp.Format(time.RFC3339)
	record.ActorID = actorID.String
	record.FieldName = fieldName.String
	record.OldValue = oldValue.String
	record.NewValue = newValue.String
	record.CreatedAt = createdAt.Format(time.RFC3339)

	return record, nil
}

// List retrieves log entries matching the given filters.
func (r *WorkshopLogRepository) List(ctx context.Context, filters secondary.WorkshopLogFilters) ([]*secondary.WorkshopLogRecord, error) {
	query := `SELECT id, workshop_id, timestamp, actor_id, entity_type, entity_id, action, field_name, old_value, new_value, created_at FROM workshop_logs WHERE 1=1`
	args := []any{}

	if filters.WorkshopID != "" {
		query += " AND workshop_id = ?"
		args = append(args, filters.WorkshopID)
	}

	if filters.EntityType != "" {
		query += " AND entity_type = ?"
		args = append(args, filters.EntityType)
	}

	if filters.EntityID != "" {
		query += " AND entity_id = ?"
		args = append(args, filters.EntityID)
	}

	if filters.ActorID != "" {
		query += " AND actor_id = ?"
		args = append(args, filters.ActorID)
	}

	if filters.Action != "" {
		query += " AND action = ?"
		args = append(args, filters.Action)
	}

	query += " ORDER BY timestamp DESC"

	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list workshop logs: %w", err)
	}
	defer rows.Close()

	var logs []*secondary.WorkshopLogRecord
	for rows.Next() {
		var (
			actorID   sql.NullString
			fieldName sql.NullString
			oldValue  sql.NullString
			newValue  sql.NullString
			timestamp time.Time
			createdAt time.Time
		)

		record := &secondary.WorkshopLogRecord{}
		err := rows.Scan(&record.ID,
			&record.WorkshopID,
			&timestamp,
			&actorID,
			&record.EntityType,
			&record.EntityID,
			&record.Action,
			&fieldName,
			&oldValue,
			&newValue,
			&createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workshop log: %w", err)
		}
		record.Timestamp = timestamp.Format(time.RFC3339)
		record.ActorID = actorID.String
		record.FieldName = fieldName.String
		record.OldValue = oldValue.String
		record.NewValue = newValue.String
		record.CreatedAt = createdAt.Format(time.RFC3339)

		logs = append(logs, record)
	}

	return logs, nil
}

// GetNextID returns the next available log ID.
func (r *WorkshopLogRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	prefixLen := len("WL-") + 1
	err := r.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT COALESCE(MAX(CAST(SUBSTR(id, %d) AS INTEGER)), 0) FROM workshop_logs", prefixLen),
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next workshop log ID: %w", err)
	}

	return fmt.Sprintf("WL-%04d", maxID+1), nil
}

// WorkshopExists checks if a workshop exists (for validation).
func (r *WorkshopLogRepository) WorkshopExists(ctx context.Context, workshopID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM workshops WHERE id = ?", workshopID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check workshop existence: %w", err)
	}
	return count > 0, nil
}

// PruneOlderThan deletes log entries older than the given number of days.
func (r *WorkshopLogRepository) PruneOlderThan(ctx context.Context, days int) (int, error) {
	result, err := r.db.ExecContext(ctx,
		"DELETE FROM workshop_logs WHERE timestamp < datetime('now', ?)",
		fmt.Sprintf("-%d days", days),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to prune workshop logs: %w", err)
	}

	count, _ := result.RowsAffected()
	return int(count), nil
}

// Ensure WorkshopLogRepository implements the interface
var _ secondary.WorkshopLogRepository = (*WorkshopLogRepository)(nil)
