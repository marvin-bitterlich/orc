// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// HookEventRepository implements secondary.HookEventRepository with SQLite.
type HookEventRepository struct {
	db *sql.DB
}

// NewHookEventRepository creates a new SQLite hook event repository.
func NewHookEventRepository(db *sql.DB) *HookEventRepository {
	return &HookEventRepository{db: db}
}

// Create persists a new hook event.
func (r *HookEventRepository) Create(ctx context.Context, event *secondary.HookEventRecord) error {
	var payloadJSON, cwd, sessionID, shipmentID, shipmentStatus, reason, errStr sql.NullString
	var taskCountIncomplete, durationMs sql.NullInt64

	if event.PayloadJSON != "" {
		payloadJSON = sql.NullString{String: event.PayloadJSON, Valid: true}
	}
	if event.Cwd != "" {
		cwd = sql.NullString{String: event.Cwd, Valid: true}
	}
	if event.SessionID != "" {
		sessionID = sql.NullString{String: event.SessionID, Valid: true}
	}
	if event.ShipmentID != "" {
		shipmentID = sql.NullString{String: event.ShipmentID, Valid: true}
	}
	if event.ShipmentStatus != "" {
		shipmentStatus = sql.NullString{String: event.ShipmentStatus, Valid: true}
	}
	if event.TaskCountIncomplete >= 0 {
		taskCountIncomplete = sql.NullInt64{Int64: int64(event.TaskCountIncomplete), Valid: true}
	}
	if event.Reason != "" {
		reason = sql.NullString{String: event.Reason, Valid: true}
	}
	if event.DurationMs >= 0 {
		durationMs = sql.NullInt64{Int64: int64(event.DurationMs), Valid: true}
	}
	if event.Error != "" {
		errStr = sql.NullString{String: event.Error, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO hook_events (id, workbench_id, hook_type, payload_json, cwd, session_id, shipment_id, shipment_status, task_count_incomplete, decision, reason, duration_ms, error) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		event.ID,
		event.WorkbenchID,
		event.HookType,
		payloadJSON,
		cwd,
		sessionID,
		shipmentID,
		shipmentStatus,
		taskCountIncomplete,
		event.Decision,
		reason,
		durationMs,
		errStr,
	)
	if err != nil {
		return fmt.Errorf("failed to create hook event: %w", err)
	}

	return nil
}

// GetByID retrieves a hook event by its ID.
func (r *HookEventRepository) GetByID(ctx context.Context, id string) (*secondary.HookEventRecord, error) {
	var (
		payloadJSON         sql.NullString
		cwd                 sql.NullString
		sessionID           sql.NullString
		shipmentID          sql.NullString
		shipmentStatus      sql.NullString
		taskCountIncomplete sql.NullInt64
		reason              sql.NullString
		durationMs          sql.NullInt64
		errStr              sql.NullString
		timestamp           time.Time
		createdAt           time.Time
	)

	record := &secondary.HookEventRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, workbench_id, hook_type, timestamp, payload_json, cwd, session_id, shipment_id, shipment_status, task_count_incomplete, decision, reason, duration_ms, error, created_at FROM hook_events WHERE id = ?`,
		id,
	).Scan(&record.ID,
		&record.WorkbenchID,
		&record.HookType,
		&timestamp,
		&payloadJSON,
		&cwd,
		&sessionID,
		&shipmentID,
		&shipmentStatus,
		&taskCountIncomplete,
		&record.Decision,
		&reason,
		&durationMs,
		&errStr,
		&createdAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("hook event %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get hook event: %w", err)
	}

	record.Timestamp = timestamp.Format(time.RFC3339)
	record.PayloadJSON = payloadJSON.String
	record.Cwd = cwd.String
	record.SessionID = sessionID.String
	record.ShipmentID = shipmentID.String
	record.ShipmentStatus = shipmentStatus.String
	if taskCountIncomplete.Valid {
		record.TaskCountIncomplete = int(taskCountIncomplete.Int64)
	} else {
		record.TaskCountIncomplete = -1
	}
	record.Reason = reason.String
	if durationMs.Valid {
		record.DurationMs = int(durationMs.Int64)
	} else {
		record.DurationMs = -1
	}
	record.Error = errStr.String
	record.CreatedAt = createdAt.Format(time.RFC3339)

	return record, nil
}

// List retrieves hook events matching the given filters.
func (r *HookEventRepository) List(ctx context.Context, filters secondary.HookEventFilters) ([]*secondary.HookEventRecord, error) {
	query := `SELECT id, workbench_id, hook_type, timestamp, payload_json, cwd, session_id, shipment_id, shipment_status, task_count_incomplete, decision, reason, duration_ms, error, created_at FROM hook_events WHERE 1=1`
	args := []any{}

	if filters.WorkbenchID != "" {
		query += " AND workbench_id = ?"
		args = append(args, filters.WorkbenchID)
	}

	if filters.HookType != "" {
		query += " AND hook_type = ?"
		args = append(args, filters.HookType)
	}

	query += " ORDER BY timestamp DESC"

	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list hook events: %w", err)
	}
	defer rows.Close()

	var events []*secondary.HookEventRecord
	for rows.Next() {
		var (
			payloadJSON         sql.NullString
			cwd                 sql.NullString
			sessionID           sql.NullString
			shipmentID          sql.NullString
			shipmentStatus      sql.NullString
			taskCountIncomplete sql.NullInt64
			reason              sql.NullString
			durationMs          sql.NullInt64
			errStr              sql.NullString
			timestamp           time.Time
			createdAt           time.Time
		)

		record := &secondary.HookEventRecord{}
		err := rows.Scan(&record.ID,
			&record.WorkbenchID,
			&record.HookType,
			&timestamp,
			&payloadJSON,
			&cwd,
			&sessionID,
			&shipmentID,
			&shipmentStatus,
			&taskCountIncomplete,
			&record.Decision,
			&reason,
			&durationMs,
			&errStr,
			&createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan hook event: %w", err)
		}

		record.Timestamp = timestamp.Format(time.RFC3339)
		record.PayloadJSON = payloadJSON.String
		record.Cwd = cwd.String
		record.SessionID = sessionID.String
		record.ShipmentID = shipmentID.String
		record.ShipmentStatus = shipmentStatus.String
		if taskCountIncomplete.Valid {
			record.TaskCountIncomplete = int(taskCountIncomplete.Int64)
		} else {
			record.TaskCountIncomplete = -1
		}
		record.Reason = reason.String
		if durationMs.Valid {
			record.DurationMs = int(durationMs.Int64)
		} else {
			record.DurationMs = -1
		}
		record.Error = errStr.String
		record.CreatedAt = createdAt.Format(time.RFC3339)

		events = append(events, record)
	}

	return events, nil
}

// GetNextID returns the next available hook event ID.
func (r *HookEventRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	prefixLen := len("HEV-") + 1
	err := r.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT COALESCE(MAX(CAST(SUBSTR(id, %d) AS INTEGER)), 0) FROM hook_events", prefixLen),
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next hook event ID: %w", err)
	}

	return fmt.Sprintf("HEV-%04d", maxID+1), nil
}

// Ensure HookEventRepository implements the interface
var _ secondary.HookEventRepository = (*HookEventRepository)(nil)
