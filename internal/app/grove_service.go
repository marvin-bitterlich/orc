package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	coregrove "github.com/example/orc/internal/core/grove"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// GroveServiceImpl implements the GroveService interface.
type GroveServiceImpl struct {
	groveRepo     secondary.GroveRepository
	missionRepo   secondary.MissionRepository
	agentProvider secondary.AgentIdentityProvider
	executor      EffectExecutor
}

// NewGroveService creates a new GroveService with injected dependencies.
func NewGroveService(
	groveRepo secondary.GroveRepository,
	missionRepo secondary.MissionRepository,
	agentProvider secondary.AgentIdentityProvider,
	executor EffectExecutor,
) *GroveServiceImpl {
	return &GroveServiceImpl{
		groveRepo:     groveRepo,
		missionRepo:   missionRepo,
		agentProvider: agentProvider,
		executor:      executor,
	}
}

// CreateGrove creates a new grove.
func (s *GroveServiceImpl) CreateGrove(ctx context.Context, req primary.CreateGroveRequest) (*primary.CreateGroveResponse, error) {
	// 1. Get agent identity
	identity, err := s.agentProvider.GetCurrentIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent identity: %w", err)
	}

	// 2. Check if mission exists
	_, err = s.missionRepo.GetByID(ctx, req.MissionID)
	missionExists := err == nil

	// 3. Guard check
	guardCtx := coregrove.CreateGroveContext{
		GuardContext: coregrove.GuardContext{
			AgentType: coregrove.AgentType(identity.Type),
			AgentID:   identity.FullID,
			MissionID: req.MissionID,
		},
		MissionExists: missionExists,
	}
	if result := coregrove.CanCreateGrove(guardCtx); !result.Allowed {
		return nil, result.Error()
	}

	// 4. Create grove record in DB first to get ID
	record := &secondary.GroveRecord{
		Name:      req.Name,
		MissionID: req.MissionID,
		Status:    "active",
	}
	if err := s.groveRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create grove: %w", err)
	}

	// 5. Generate plan
	basePath := req.BasePath
	if basePath == "" {
		basePath = s.defaultBasePath()
	}

	planInput := coregrove.CreateGrovePlanInput{
		GroveID:   record.ID,
		GroveName: req.Name,
		MissionID: req.MissionID,
		BasePath:  basePath,
		Repos:     req.Repos,
	}
	plan := coregrove.GenerateCreateGrovePlan(planInput)

	// 6. Execute effects
	if err := s.executor.Execute(ctx, plan.Effects()); err != nil {
		return nil, fmt.Errorf("failed to execute create plan: %w", err)
	}

	// 7. Update grove path in DB
	record.WorktreePath = plan.GrovePath
	if err := s.groveRepo.Update(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to update grove path: %w", err)
	}

	return &primary.CreateGroveResponse{
		GroveID: record.ID,
		Grove:   s.recordToGrove(record),
		Path:    plan.GrovePath,
	}, nil
}

// OpenGrove opens a grove in TMux.
func (s *GroveServiceImpl) OpenGrove(ctx context.Context, req primary.OpenGroveRequest) (*primary.OpenGroveResponse, error) {
	// 1. Fetch grove
	grove, err := s.groveRepo.GetByID(ctx, req.GroveID)
	if err != nil {
		return nil, fmt.Errorf("grove not found: %w", err)
	}

	// 2. Check path exists
	pathExists := s.pathExists(grove.WorktreePath)

	// 3. Check TMux session (via environment)
	inTMux := os.Getenv("TMUX") != ""

	// 4. Guard check
	guardCtx := coregrove.OpenGroveContext{
		GroveID:       req.GroveID,
		GroveExists:   true,
		PathExists:    pathExists,
		InTMuxSession: inTMux,
	}
	if result := coregrove.CanOpenGrove(guardCtx); !result.Allowed {
		return nil, result.Error()
	}

	// 5. Get TMux session info
	sessionName := s.getTMuxSession()

	// 6. Generate plan
	planInput := coregrove.OpenGrovePlanInput{
		GroveID:         grove.ID,
		GroveName:       grove.Name,
		GrovePath:       grove.WorktreePath,
		SessionName:     sessionName,
		NextWindowIndex: 0, // Will be determined by executor
	}
	plan := coregrove.GenerateOpenGrovePlan(planInput)

	// 7. Execute effects
	if err := s.executor.Execute(ctx, plan.Effects()); err != nil {
		return nil, fmt.Errorf("failed to open grove: %w", err)
	}

	return &primary.OpenGroveResponse{
		Grove:       s.recordToGrove(grove),
		SessionName: sessionName,
		WindowName:  grove.Name,
	}, nil
}

// GetGrove retrieves a grove by ID.
func (s *GroveServiceImpl) GetGrove(ctx context.Context, groveID string) (*primary.Grove, error) {
	record, err := s.groveRepo.GetByID(ctx, groveID)
	if err != nil {
		return nil, fmt.Errorf("grove not found: %w", err)
	}
	return s.recordToGrove(record), nil
}

// GetGroveByPath retrieves a grove by its filesystem path.
func (s *GroveServiceImpl) GetGroveByPath(ctx context.Context, path string) (*primary.Grove, error) {
	record, err := s.groveRepo.GetByPath(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("grove not found at path: %w", err)
	}
	return s.recordToGrove(record), nil
}

// UpdateGrovePath updates the filesystem path of a grove.
func (s *GroveServiceImpl) UpdateGrovePath(ctx context.Context, groveID, newPath string) error {
	// Verify grove exists
	_, err := s.groveRepo.GetByID(ctx, groveID)
	if err != nil {
		return fmt.Errorf("grove not found: %w", err)
	}

	return s.groveRepo.UpdatePath(ctx, groveID, newPath)
}

// ListGroves lists groves with optional filters.
func (s *GroveServiceImpl) ListGroves(ctx context.Context, filters primary.GroveFilters) ([]*primary.Grove, error) {
	// Use List which handles empty missionID (lists all groves)
	records, err := s.groveRepo.List(ctx, filters.MissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list groves: %w", err)
	}

	groves := make([]*primary.Grove, len(records))
	for i, r := range records {
		groves[i] = s.recordToGrove(r)
	}
	return groves, nil
}

// RenameGrove renames a grove.
func (s *GroveServiceImpl) RenameGrove(ctx context.Context, req primary.RenameGroveRequest) error {
	// 1. Check grove exists
	grove, err := s.groveRepo.GetByID(ctx, req.GroveID)
	if err != nil {
		return fmt.Errorf("grove not found: %w", err)
	}

	// 2. Guard check
	if result := coregrove.CanRenameGrove(true, req.GroveID); !result.Allowed {
		return result.Error()
	}

	// 3. Update name
	grove.Name = req.NewName
	if err := s.groveRepo.Update(ctx, grove); err != nil {
		return fmt.Errorf("failed to rename grove: %w", err)
	}

	return nil
}

// DeleteGrove deletes a grove.
func (s *GroveServiceImpl) DeleteGrove(ctx context.Context, req primary.DeleteGroveRequest) error {
	// 1. Fetch grove
	_, err := s.groveRepo.GetByID(ctx, req.GroveID)
	if err != nil {
		return fmt.Errorf("grove not found: %w", err)
	}

	// 2. Count active work (simplified - could add task repo)
	activeTaskCount := 0

	// 3. Guard check
	guardCtx := coregrove.DeleteGroveContext{
		GroveID:         req.GroveID,
		ActiveTaskCount: activeTaskCount,
		ForceDelete:     req.Force,
	}
	if result := coregrove.CanDeleteGrove(guardCtx); !result.Allowed {
		return result.Error()
	}

	// 4. Delete from database
	return s.groveRepo.Delete(ctx, req.GroveID)
}

// Helper methods

func (s *GroveServiceImpl) recordToGrove(r *secondary.GroveRecord) *primary.Grove {
	return &primary.Grove{
		ID:        r.ID,
		Name:      r.Name,
		MissionID: r.MissionID,
		Path:      r.WorktreePath,
		Status:    r.Status,
		CreatedAt: r.CreatedAt,
	}
}

func (s *GroveServiceImpl) defaultBasePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "src", "worktrees")
}

func (s *GroveServiceImpl) pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (s *GroveServiceImpl) getTMuxSession() string {
	// In production, would parse TMUX env var or run tmux display-message
	// For now, return a default
	return "orc"
}

// Ensure GroveServiceImpl implements the interface
var _ primary.GroveService = (*GroveServiceImpl)(nil)
