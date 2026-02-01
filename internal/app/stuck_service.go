package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/core/detection"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// StuckServiceImpl implements the StuckService interface.
type StuckServiceImpl struct {
	stuckRepo      secondary.StuckRepository
	stuckThreshold int
}

// NewStuckService creates a new StuckService with injected dependencies.
func NewStuckService(stuckRepo secondary.StuckRepository) *StuckServiceImpl {
	return &StuckServiceImpl{
		stuckRepo:      stuckRepo,
		stuckThreshold: detection.DefaultStuckThreshold,
	}
}

// NewStuckServiceWithThreshold creates a StuckService with a custom threshold.
func NewStuckServiceWithThreshold(stuckRepo secondary.StuckRepository, threshold int) *StuckServiceImpl {
	return &StuckServiceImpl{
		stuckRepo:      stuckRepo,
		stuckThreshold: threshold,
	}
}

// ProcessOutcome processes a check outcome and manages stuck state.
// Returns the result indicating what action to take.
func (s *StuckServiceImpl) ProcessOutcome(ctx context.Context, patrolID, checkID, outcome string) (*primary.StuckResult, error) {
	// Get current open stuck for this patrol (if any)
	openStuck, err := s.stuckRepo.GetOpenByPatrol(ctx, patrolID)
	if err != nil {
		return nil, fmt.Errorf("failed to get open stuck: %w", err)
	}

	// Handle based on outcome
	switch outcome {
	case detection.OutcomeError:
		return s.handleFailure(ctx, patrolID, checkID, openStuck)

	case detection.OutcomeWorking:
		// Recovery: if there's an open stuck, resolve it
		if openStuck != nil {
			if err := s.stuckRepo.UpdateStatus(ctx, openStuck.ID, primary.StuckStatusResolved); err != nil {
				return nil, fmt.Errorf("failed to resolve stuck: %w", err)
			}
		}
		return &primary.StuckResult{
			Action:          detection.ActionNone,
			StuckID:         "",
			CheckCount:      0,
			NeedsEscalation: false,
		}, nil

	case detection.OutcomeIdle:
		// Idle triggers a nudge but doesn't start stuck tracking
		return &primary.StuckResult{
			Action:          detection.ActionNudge,
			StuckID:         "",
			CheckCount:      0,
			NeedsEscalation: false,
			Message:         detection.NudgeMessages[detection.OutcomeIdle],
		}, nil

	case detection.OutcomeMenu:
		// Menu triggers BTab, resolve any stuck
		if openStuck != nil {
			if err := s.stuckRepo.UpdateStatus(ctx, openStuck.ID, primary.StuckStatusResolved); err != nil {
				return nil, fmt.Errorf("failed to resolve stuck: %w", err)
			}
		}
		return &primary.StuckResult{
			Action:          detection.ActionBTab,
			StuckID:         "",
			CheckCount:      0,
			NeedsEscalation: false,
		}, nil

	case detection.OutcomeTyped:
		// Typed triggers Enter, resolve any stuck
		if openStuck != nil {
			if err := s.stuckRepo.UpdateStatus(ctx, openStuck.ID, primary.StuckStatusResolved); err != nil {
				return nil, fmt.Errorf("failed to resolve stuck: %w", err)
			}
		}
		return &primary.StuckResult{
			Action:          detection.ActionEnter,
			StuckID:         "",
			CheckCount:      0,
			NeedsEscalation: false,
		}, nil

	default:
		// Unknown outcome, no action
		return &primary.StuckResult{
			Action:          detection.ActionNone,
			StuckID:         "",
			CheckCount:      0,
			NeedsEscalation: false,
		}, nil
	}
}

// handleFailure handles error/failure outcomes.
func (s *StuckServiceImpl) handleFailure(ctx context.Context, patrolID, checkID string, openStuck *secondary.StuckRecord) (*primary.StuckResult, error) {
	var stuckID string
	var checkCount int

	if openStuck == nil {
		// First failure: create new stuck
		id, err := s.stuckRepo.GetNextID(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate stuck ID: %w", err)
		}

		record := &secondary.StuckRecord{
			ID:           id,
			PatrolID:     patrolID,
			FirstCheckID: checkID,
			CheckCount:   1,
			Status:       primary.StuckStatusOpen,
		}
		if err := s.stuckRepo.Create(ctx, record); err != nil {
			return nil, fmt.Errorf("failed to create stuck: %w", err)
		}

		stuckID = id
		checkCount = 1
	} else {
		// Consecutive failure: increment count
		// Capture the new count before calling repo (the struct may be modified in place)
		stuckID = openStuck.ID
		checkCount = openStuck.CheckCount + 1

		if err := s.stuckRepo.IncrementCount(ctx, openStuck.ID); err != nil {
			return nil, fmt.Errorf("failed to increment stuck count: %w", err)
		}
	}

	// Determine action based on count
	if checkCount >= s.stuckThreshold {
		return &primary.StuckResult{
			Action:          detection.ActionEscalate,
			StuckID:         stuckID,
			CheckCount:      checkCount,
			NeedsEscalation: true,
			Message:         fmt.Sprintf("IMP has been stuck on error for %d checks. Needs human intervention.", checkCount),
		}, nil
	}

	return &primary.StuckResult{
		Action:          detection.ActionNudge,
		StuckID:         stuckID,
		CheckCount:      checkCount,
		NeedsEscalation: false,
		Message:         detection.NudgeMessages[detection.OutcomeError],
	}, nil
}

// GetStuck retrieves a stuck by ID.
func (s *StuckServiceImpl) GetStuck(ctx context.Context, stuckID string) (*primary.Stuck, error) {
	record, err := s.stuckRepo.GetByID(ctx, stuckID)
	if err != nil {
		return nil, err
	}
	return s.recordToStuck(record), nil
}

// GetOpenStuck retrieves the open stuck for a patrol (if any).
func (s *StuckServiceImpl) GetOpenStuck(ctx context.Context, patrolID string) (*primary.Stuck, error) {
	record, err := s.stuckRepo.GetOpenByPatrol(ctx, patrolID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}
	return s.recordToStuck(record), nil
}

// ResolveStuck marks a stuck as resolved.
func (s *StuckServiceImpl) ResolveStuck(ctx context.Context, stuckID string) error {
	return s.stuckRepo.UpdateStatus(ctx, stuckID, primary.StuckStatusResolved)
}

// EscalateStuck marks a stuck as escalated.
func (s *StuckServiceImpl) EscalateStuck(ctx context.Context, stuckID string) error {
	return s.stuckRepo.UpdateStatus(ctx, stuckID, primary.StuckStatusEscalated)
}

// Helper methods

func (s *StuckServiceImpl) recordToStuck(r *secondary.StuckRecord) *primary.Stuck {
	return &primary.Stuck{
		ID:           r.ID,
		PatrolID:     r.PatrolID,
		FirstCheckID: r.FirstCheckID,
		CheckCount:   r.CheckCount,
		Status:       r.Status,
		ResolvedAt:   r.ResolvedAt,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}

// Ensure StuckServiceImpl implements the interface
var _ primary.StuckService = (*StuckServiceImpl)(nil)
