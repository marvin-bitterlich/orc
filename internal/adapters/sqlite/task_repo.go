// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// TaskRepository implements secondary.TaskRepository with SQLite.
type TaskRepository struct {
	db *sql.DB
}

// NewTaskRepository creates a new SQLite task repository.
func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// scanTask scans a task row into a TaskRecord.
func scanTask(scanner interface {
	Scan(dest ...any) error
}) (*secondary.TaskRecord, error) {
	var (
		shipmentID       sql.NullString
		desc             sql.NullString
		taskType         sql.NullString
		priority         sql.NullString
		assignedGroveID  sql.NullString
		pinned           bool
		createdAt        time.Time
		updatedAt        time.Time
		claimedAt        sql.NullTime
		completedAt      sql.NullTime
		conclaveID       sql.NullString
		promotedFromID   sql.NullString
		promotedFromType sql.NullString
	)

	record := &secondary.TaskRecord{}
	err := scanner.Scan(
		&record.ID, &shipmentID, &record.MissionID, &record.Title, &desc,
		&taskType, &record.Status, &priority, &assignedGroveID,
		&pinned, &createdAt, &updatedAt, &claimedAt, &completedAt,
		&conclaveID, &promotedFromID, &promotedFromType,
	)
	if err != nil {
		return nil, err
	}

	record.ShipmentID = shipmentID.String
	record.Description = desc.String
	record.Type = taskType.String
	record.Priority = priority.String
	record.AssignedGroveID = assignedGroveID.String
	record.Pinned = pinned
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	record.ConclaveID = conclaveID.String
	record.PromotedFromID = promotedFromID.String
	record.PromotedFromType = promotedFromType.String

	if claimedAt.Valid {
		record.ClaimedAt = claimedAt.Time.Format(time.RFC3339)
	}
	if completedAt.Valid {
		record.CompletedAt = completedAt.Time.Format(time.RFC3339)
	}

	return record, nil
}

const taskSelectCols = "id, shipment_id, mission_id, title, description, type, status, priority, assigned_grove_id, pinned, created_at, updated_at, claimed_at, completed_at, conclave_id, promoted_from_id, promoted_from_type"

// Create persists a new task.
func (r *TaskRepository) Create(ctx context.Context, task *secondary.TaskRecord) error {
	var shipmentID, desc, taskType sql.NullString

	if task.ShipmentID != "" {
		shipmentID = sql.NullString{String: task.ShipmentID, Valid: true}
	}
	if task.Description != "" {
		desc = sql.NullString{String: task.Description, Valid: true}
	}
	if task.Type != "" {
		taskType = sql.NullString{String: task.Type, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO tasks (id, shipment_id, mission_id, title, description, type, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
		task.ID, shipmentID, task.MissionID, task.Title, desc, taskType, "ready",
	)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	return nil
}

// GetByID retrieves a task by its ID.
func (r *TaskRepository) GetByID(ctx context.Context, id string) (*secondary.TaskRecord, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT "+taskSelectCols+" FROM tasks WHERE id = ?",
		id,
	)

	record, err := scanTask(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return record, nil
}

// List retrieves tasks matching the given filters.
func (r *TaskRepository) List(ctx context.Context, filters secondary.TaskFilters) ([]*secondary.TaskRecord, error) {
	query := "SELECT " + taskSelectCols + " FROM tasks WHERE 1=1"
	args := []any{}

	if filters.ShipmentID != "" {
		query += " AND shipment_id = ?"
		args = append(args, filters.ShipmentID)
	}

	if filters.Status != "" {
		query += " AND status = ?"
		args = append(args, filters.Status)
	}

	if filters.MissionID != "" {
		query += " AND mission_id = ?"
		args = append(args, filters.MissionID)
	}

	query += " ORDER BY created_at ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*secondary.TaskRecord
	for rows.Next() {
		record, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, record)
	}

	return tasks, nil
}

// Update updates an existing task.
func (r *TaskRepository) Update(ctx context.Context, task *secondary.TaskRecord) error {
	query := "UPDATE tasks SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if task.Title != "" {
		query += ", title = ?"
		args = append(args, task.Title)
	}

	if task.Description != "" {
		query += ", description = ?"
		args = append(args, sql.NullString{String: task.Description, Valid: true})
	}

	query += " WHERE id = ?"
	args = append(args, task.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("task %s not found", task.ID)
	}

	return nil
}

// Delete removes a task from persistence.
func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("task %s not found", id)
	}

	return nil
}

// Pin pins a task.
func (r *TaskRepository) Pin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE tasks SET pinned = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to pin task: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("task %s not found", id)
	}

	return nil
}

// Unpin unpins a task.
func (r *TaskRepository) Unpin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE tasks SET pinned = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to unpin task: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("task %s not found", id)
	}

	return nil
}

// GetNextID returns the next available task ID.
func (r *TaskRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 6) AS INTEGER)), 0) FROM tasks",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next task ID: %w", err)
	}

	return fmt.Sprintf("TASK-%03d", maxID+1), nil
}

// GetByGrove retrieves tasks assigned to a grove.
func (r *TaskRepository) GetByGrove(ctx context.Context, groveID string) ([]*secondary.TaskRecord, error) {
	query := "SELECT " + taskSelectCols + " FROM tasks WHERE assigned_grove_id = ?"
	rows, err := r.db.QueryContext(ctx, query, groveID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by grove: %w", err)
	}
	defer rows.Close()

	var tasks []*secondary.TaskRecord
	for rows.Next() {
		record, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, record)
	}

	return tasks, nil
}

// GetByShipment retrieves tasks for a shipment.
func (r *TaskRepository) GetByShipment(ctx context.Context, shipmentID string) ([]*secondary.TaskRecord, error) {
	query := "SELECT " + taskSelectCols + " FROM tasks WHERE shipment_id = ? ORDER BY created_at ASC"
	rows, err := r.db.QueryContext(ctx, query, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by shipment: %w", err)
	}
	defer rows.Close()

	var tasks []*secondary.TaskRecord
	for rows.Next() {
		record, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, record)
	}

	return tasks, nil
}

// UpdateStatus updates the status with optional timestamps.
func (r *TaskRepository) UpdateStatus(ctx context.Context, id, status string, setClaimed, setCompleted bool) error {
	query := "UPDATE tasks SET status = ?, updated_at = CURRENT_TIMESTAMP"
	args := []any{status}

	if setClaimed {
		query += ", claimed_at = CURRENT_TIMESTAMP"
	}
	if setCompleted {
		query += ", completed_at = CURRENT_TIMESTAMP"
	}

	query += " WHERE id = ?"
	args = append(args, id)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("task %s not found", id)
	}

	return nil
}

// Claim claims a task for a grove.
func (r *TaskRepository) Claim(ctx context.Context, id, groveID string) error {
	var groveIDNullable sql.NullString
	if groveID != "" {
		groveIDNullable = sql.NullString{String: groveID, Valid: true}
	}

	result, err := r.db.ExecContext(ctx,
		"UPDATE tasks SET status = 'in_progress', assigned_grove_id = ?, claimed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		groveIDNullable, id,
	)
	if err != nil {
		return fmt.Errorf("failed to claim task: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("task %s not found", id)
	}

	return nil
}

// AssignGroveByShipment assigns all tasks of a shipment to a grove.
func (r *TaskRepository) AssignGroveByShipment(ctx context.Context, shipmentID, groveID string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE tasks SET assigned_grove_id = ?, updated_at = CURRENT_TIMESTAMP WHERE shipment_id = ?",
		groveID, shipmentID,
	)
	if err != nil {
		return fmt.Errorf("failed to assign grove to shipment tasks: %w", err)
	}

	return nil
}

// MissionExists checks if a mission exists.
func (r *TaskRepository) MissionExists(ctx context.Context, missionID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM missions WHERE id = ?", missionID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check mission existence: %w", err)
	}
	return count > 0, nil
}

// ShipmentExists checks if a shipment exists.
func (r *TaskRepository) ShipmentExists(ctx context.Context, shipmentID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM shipments WHERE id = ?", shipmentID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check shipment existence: %w", err)
	}
	return count > 0, nil
}

// GetTag retrieves the tag for a task (nil if none).
func (r *TaskRepository) GetTag(ctx context.Context, taskID string) (*secondary.TagRecord, error) {
	var tagID, tagName string
	err := r.db.QueryRowContext(ctx,
		"SELECT t.id, t.name FROM tags t INNER JOIN entity_tags et ON t.id = et.tag_id WHERE et.entity_id = ? AND et.entity_type = 'task'",
		taskID,
	).Scan(&tagID, &tagName)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task tag: %w", err)
	}

	return &secondary.TagRecord{ID: tagID, Name: tagName}, nil
}

// AddTag adds a tag to a task.
func (r *TaskRepository) AddTag(ctx context.Context, taskID, tagID string) error {
	// Get next entity tag ID
	nextID, err := r.GetNextEntityTagID(ctx)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx,
		"INSERT INTO entity_tags (id, entity_id, entity_type, tag_id) VALUES (?, ?, 'task', ?)",
		nextID, taskID, tagID,
	)
	if err != nil {
		return fmt.Errorf("failed to add tag to task: %w", err)
	}

	return nil
}

// RemoveTag removes the tag from a task.
func (r *TaskRepository) RemoveTag(ctx context.Context, taskID string) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM entity_tags WHERE entity_id = ? AND entity_type = 'task'",
		taskID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove tag from task: %w", err)
	}

	return nil
}

// ListByTag retrieves tasks with a specific tag.
func (r *TaskRepository) ListByTag(ctx context.Context, tagID string) ([]*secondary.TaskRecord, error) {
	query := `
		SELECT t.id, t.shipment_id, t.mission_id, t.title, t.description,
		       t.type, t.status, t.priority, t.assigned_grove_id,
		       t.pinned, t.created_at, t.updated_at, t.claimed_at, t.completed_at,
		       t.conclave_id, t.promoted_from_id, t.promoted_from_type
		FROM tasks t
		INNER JOIN entity_tags et ON t.id = et.entity_id AND et.entity_type = 'task'
		WHERE et.tag_id = ?
		ORDER BY t.created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tagID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks by tag: %w", err)
	}
	defer rows.Close()

	var tasks []*secondary.TaskRecord
	for rows.Next() {
		record, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, record)
	}

	return tasks, nil
}

// GetNextEntityTagID returns the next available entity tag ID.
func (r *TaskRepository) GetNextEntityTagID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 4) AS INTEGER)), 0) FROM entity_tags",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next entity tag ID: %w", err)
	}

	return fmt.Sprintf("ET-%03d", maxID+1), nil
}

// Ensure TaskRepository implements the interface
var _ secondary.TaskRepository = (*TaskRepository)(nil)
