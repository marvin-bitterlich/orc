package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// LogServiceImpl implements the LogService interface.
type LogServiceImpl struct {
	logRepo secondary.WorkshopLogRepository
}

// NewLogService creates a new LogService with injected dependencies.
func NewLogService(logRepo secondary.WorkshopLogRepository) *LogServiceImpl {
	return &LogServiceImpl{
		logRepo: logRepo,
	}
}

// ListLogs retrieves log entries matching the given filters.
func (s *LogServiceImpl) ListLogs(ctx context.Context, filters primary.LogFilters) ([]*primary.LogEntry, error) {
	records, err := s.logRepo.List(ctx, secondary.WorkshopLogFilters{
		WorkshopID: filters.WorkshopID,
		EntityType: filters.EntityType,
		EntityID:   filters.EntityID,
		ActorID:    filters.ActorID,
		Action:     filters.Action,
		Limit:      filters.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list logs: %w", err)
	}

	entries := make([]*primary.LogEntry, len(records))
	for i, r := range records {
		entries[i] = s.recordToLogEntry(r)
	}
	return entries, nil
}

// GetLog retrieves a single log entry by ID.
func (s *LogServiceImpl) GetLog(ctx context.Context, id string) (*primary.LogEntry, error) {
	record, err := s.logRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.recordToLogEntry(record), nil
}

// PruneLogs deletes log entries older than the specified number of days.
func (s *LogServiceImpl) PruneLogs(ctx context.Context, olderThanDays int) (int, error) {
	return s.logRepo.PruneOlderThan(ctx, olderThanDays)
}

// Helper methods

func (s *LogServiceImpl) recordToLogEntry(r *secondary.WorkshopLogRecord) *primary.LogEntry {
	return &primary.LogEntry{
		ID:         r.ID,
		WorkshopID: r.WorkshopID,
		Timestamp:  r.Timestamp,
		ActorID:    r.ActorID,
		EntityType: r.EntityType,
		EntityID:   r.EntityID,
		Action:     r.Action,
		FieldName:  r.FieldName,
		OldValue:   r.OldValue,
		NewValue:   r.NewValue,
		CreatedAt:  r.CreatedAt,
	}
}

// Ensure LogServiceImpl implements the interface
var _ primary.LogService = (*LogServiceImpl)(nil)
