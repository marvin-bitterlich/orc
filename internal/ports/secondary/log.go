package secondary

import "context"

// LogWriter defines the interface for writing audit log entries.
// Implementations extract actor from context and handle workshop resolution.
type LogWriter interface {
	// LogCreate logs a create operation for an entity.
	LogCreate(ctx context.Context, entityType, entityID string) error

	// LogUpdate logs an update operation for an entity field.
	// fieldName, oldValue, newValue describe what changed.
	LogUpdate(ctx context.Context, entityType, entityID, fieldName, oldValue, newValue string) error

	// LogDelete logs a delete operation for an entity.
	LogDelete(ctx context.Context, entityType, entityID string) error
}
