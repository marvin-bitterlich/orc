package primary

import "context"

// LogService defines the primary port for workshop activity log operations.
type LogService interface {
	// ListLogs retrieves log entries matching the given filters.
	ListLogs(ctx context.Context, filters LogFilters) ([]*LogEntry, error)

	// GetLog retrieves a single log entry by ID.
	GetLog(ctx context.Context, id string) (*LogEntry, error)

	// PruneLogs deletes log entries older than the specified number of days.
	PruneLogs(ctx context.Context, olderThanDays int) (int, error)
}

// LogEntry represents a workshop activity log entry at the port boundary.
type LogEntry struct {
	ID         string
	WorkshopID string
	Timestamp  string
	ActorID    string
	EntityType string
	EntityID   string
	Action     string // 'create', 'update', 'delete'
	FieldName  string // For updates only
	OldValue   string
	NewValue   string
	CreatedAt  string
}

// LogFilters contains filter options for querying logs.
type LogFilters struct {
	WorkshopID string
	EntityType string
	EntityID   string
	ActorID    string
	Action     string
	Limit      int
}
