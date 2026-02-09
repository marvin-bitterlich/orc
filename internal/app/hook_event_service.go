package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// HookEventServiceImpl implements the HookEventService interface.
type HookEventServiceImpl struct {
	hookEventRepo secondary.HookEventRepository
}

// NewHookEventService creates a new HookEventService with injected dependencies.
func NewHookEventService(hookEventRepo secondary.HookEventRepository) *HookEventServiceImpl {
	return &HookEventServiceImpl{
		hookEventRepo: hookEventRepo,
	}
}

// LogHookEvent logs a new hook invocation event.
func (s *HookEventServiceImpl) LogHookEvent(ctx context.Context, req primary.LogHookEventRequest) (*primary.LogHookEventResponse, error) {
	// Get next ID
	nextID, err := s.hookEventRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hook event ID: %w", err)
	}

	// Create record
	record := &secondary.HookEventRecord{
		ID:                  nextID,
		WorkbenchID:         req.WorkbenchID,
		HookType:            req.HookType,
		PayloadJSON:         req.PayloadJSON,
		Cwd:                 req.Cwd,
		SessionID:           req.SessionID,
		ShipmentID:          req.ShipmentID,
		ShipmentStatus:      req.ShipmentStatus,
		TaskCountIncomplete: req.TaskCountIncomplete,
		Decision:            req.Decision,
		Reason:              req.Reason,
		DurationMs:          req.DurationMs,
		Error:               req.Error,
	}

	if err := s.hookEventRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to log hook event: %w", err)
	}

	// Fetch created event
	created, err := s.hookEventRepo.GetByID(ctx, nextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created hook event: %w", err)
	}

	return &primary.LogHookEventResponse{
		EventID: created.ID,
		Event:   s.recordToHookEvent(created),
	}, nil
}

// GetHookEvent retrieves a hook event by ID.
func (s *HookEventServiceImpl) GetHookEvent(ctx context.Context, eventID string) (*primary.HookEvent, error) {
	record, err := s.hookEventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	return s.recordToHookEvent(record), nil
}

// ListHookEvents retrieves hook events matching the given filters.
func (s *HookEventServiceImpl) ListHookEvents(ctx context.Context, filters primary.HookEventFilters) ([]*primary.HookEvent, error) {
	records, err := s.hookEventRepo.List(ctx, secondary.HookEventFilters{
		WorkbenchID: filters.WorkbenchID,
		HookType:    filters.HookType,
		Limit:       filters.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list hook events: %w", err)
	}

	events := make([]*primary.HookEvent, len(records))
	for i, r := range records {
		events[i] = s.recordToHookEvent(r)
	}
	return events, nil
}

// Helper methods

func (s *HookEventServiceImpl) recordToHookEvent(r *secondary.HookEventRecord) *primary.HookEvent {
	return &primary.HookEvent{
		ID:                  r.ID,
		WorkbenchID:         r.WorkbenchID,
		HookType:            r.HookType,
		Timestamp:           r.Timestamp,
		PayloadJSON:         r.PayloadJSON,
		Cwd:                 r.Cwd,
		SessionID:           r.SessionID,
		ShipmentID:          r.ShipmentID,
		ShipmentStatus:      r.ShipmentStatus,
		TaskCountIncomplete: r.TaskCountIncomplete,
		Decision:            r.Decision,
		Reason:              r.Reason,
		DurationMs:          r.DurationMs,
		Error:               r.Error,
		CreatedAt:           r.CreatedAt,
	}
}

// Ensure HookEventServiceImpl implements the interface.
var _ primary.HookEventService = (*HookEventServiceImpl)(nil)
