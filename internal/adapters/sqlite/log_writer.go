// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"strings"

	"github.com/example/orc/internal/ctxutil"
	"github.com/example/orc/internal/ports/secondary"
)

// LogWriterAdapter implements secondary.LogWriter using WorkshopLogRepository.
type LogWriterAdapter struct {
	logRepo       secondary.WorkshopLogRepository
	workbenchRepo secondary.WorkbenchRepository
}

// NewLogWriterAdapter creates a new LogWriterAdapter.
// workbenchRepo is used to resolve workshop from workbench actors.
func NewLogWriterAdapter(logRepo secondary.WorkshopLogRepository, workbenchRepo secondary.WorkbenchRepository) *LogWriterAdapter {
	return &LogWriterAdapter{
		logRepo:       logRepo,
		workbenchRepo: workbenchRepo,
	}
}

// LogCreate logs a create operation for an entity.
func (w *LogWriterAdapter) LogCreate(ctx context.Context, entityType, entityID string) error {
	return w.writeLog(ctx, entityType, entityID, "create", "", "", "")
}

// LogUpdate logs an update operation for an entity field.
func (w *LogWriterAdapter) LogUpdate(ctx context.Context, entityType, entityID, fieldName, oldValue, newValue string) error {
	return w.writeLog(ctx, entityType, entityID, "update", fieldName, oldValue, newValue)
}

// LogDelete logs a delete operation for an entity.
func (w *LogWriterAdapter) LogDelete(ctx context.Context, entityType, entityID string) error {
	return w.writeLog(ctx, entityType, entityID, "delete", "", "", "")
}

// writeLog writes a log entry with common logic.
func (w *LogWriterAdapter) writeLog(ctx context.Context, entityType, entityID, action, fieldName, oldValue, newValue string) error {
	// Get actor from context
	actorID := ctxutil.ActorFromContext(ctx)

	// Resolve workshop from actor
	workshopID := w.resolveWorkshop(ctx, actorID)
	if workshopID == "" {
		// No workshop context - skip logging
		// This happens for operations outside workshop scope
		return nil
	}

	// Generate log ID
	id, err := w.logRepo.GetNextID(ctx)
	if err != nil {
		return err
	}

	record := &secondary.WorkshopLogRecord{
		ID:         id,
		WorkshopID: workshopID,
		ActorID:    actorID,
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		FieldName:  fieldName,
		OldValue:   oldValue,
		NewValue:   newValue,
	}

	return w.logRepo.Create(ctx, record)
}

// resolveWorkshop resolves the workshop ID from the actor.
// For BENCH-xxx actors, looks up the workbench's workshop.
// For GATE-xxx actors, the gatehouse IS the workshop context.
// Returns empty string if workshop cannot be resolved.
func (w *LogWriterAdapter) resolveWorkshop(ctx context.Context, actorID string) string {
	if actorID == "" {
		return ""
	}

	// Parse actor to find workbench or gatehouse
	// Actor IDs are like "IMP-BENCH-014" or "GOBLIN-GATE-003" or just "GOBLIN"
	if strings.Contains(actorID, "BENCH-") {
		// Extract workbench ID (e.g., "BENCH-014" from "IMP-BENCH-014")
		parts := strings.Split(actorID, "-")
		for i, p := range parts {
			if p == "BENCH" && i+1 < len(parts) {
				workbenchID := "BENCH-" + parts[i+1]
				// Look up workbench to get workshop
				if w.workbenchRepo != nil {
					bench, err := w.workbenchRepo.GetByID(ctx, workbenchID)
					if err == nil && bench != nil {
						return bench.WorkshopID
					}
				}
				return ""
			}
		}
	}

	if strings.Contains(actorID, "GATE-") {
		// For gatehouse actors, we'd need to look up the gatehouse's workshop
		// For now, skip - gatehouses are typically orchestrating, not in workshop scope
		return ""
	}

	return ""
}

// Ensure LogWriterAdapter implements the interface
var _ secondary.LogWriter = (*LogWriterAdapter)(nil)
