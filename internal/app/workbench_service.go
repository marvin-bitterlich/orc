package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/example/orc/internal/config"
	"github.com/example/orc/internal/core/effects"
	coreworkbench "github.com/example/orc/internal/core/workbench"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// WorkbenchServiceImpl implements the WorkbenchService interface.
type WorkbenchServiceImpl struct {
	workbenchRepo secondary.WorkbenchRepository
	workshopRepo  secondary.WorkshopRepository
	agentProvider secondary.AgentIdentityProvider
	executor      EffectExecutor
	gitService    *GitService
}

// NewWorkbenchService creates a new WorkbenchService with injected dependencies.
func NewWorkbenchService(
	workbenchRepo secondary.WorkbenchRepository,
	workshopRepo secondary.WorkshopRepository,
	agentProvider secondary.AgentIdentityProvider,
	executor EffectExecutor,
) *WorkbenchServiceImpl {
	return &WorkbenchServiceImpl{
		workbenchRepo: workbenchRepo,
		workshopRepo:  workshopRepo,
		agentProvider: agentProvider,
		executor:      executor,
		gitService:    NewGitService(),
	}
}

// CreateWorkbench creates a new workbench.
func (s *WorkbenchServiceImpl) CreateWorkbench(ctx context.Context, req primary.CreateWorkbenchRequest) (*primary.CreateWorkbenchResponse, error) {
	// 1. Get agent identity
	identity, err := s.agentProvider.GetCurrentIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent identity: %w", err)
	}

	// 2. Check if workshop exists
	workshopExists, err := s.workbenchRepo.WorkshopExists(ctx, req.WorkshopID)
	if err != nil {
		return nil, fmt.Errorf("failed to check workshop: %w", err)
	}

	// 3. Guard check
	guardCtx := coreworkbench.CreateWorkbenchContext{
		GuardContext: coreworkbench.GuardContext{
			AgentType:  coreworkbench.AgentType(identity.Type),
			AgentID:    identity.FullID,
			WorkshopID: req.WorkshopID,
		},
		WorkshopExists: workshopExists,
	}
	if result := coreworkbench.CanCreateWorkbench(guardCtx); !result.Allowed {
		return nil, result.Error()
	}

	// 4. Build workbench path (~/wb/<name>)
	basePath := req.BasePath
	if basePath == "" {
		basePath = s.defaultBasePath()
	}
	workbenchPath := filepath.Join(basePath, req.Name)

	// 5. Generate home branch name
	homeBranch := GenerateHomeBranchName(UserInitials, req.Name)

	// 6. Create workbench record in DB
	record := &secondary.WorkbenchRecord{
		Name:          req.Name,
		WorkshopID:    req.WorkshopID,
		RepoID:        req.RepoID,
		WorktreePath:  workbenchPath,
		Status:        "active",
		HomeBranch:    homeBranch,
		CurrentBranch: homeBranch,
	}
	if err := s.workbenchRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create workbench: %w", err)
	}

	// 7. Write .orc/config.json via effects (non-fatal - workbench created even if config fails)
	// Skip if caller wants to write config after worktree setup (avoids race condition)
	if !req.SkipConfigWrite {
		configEffects := s.buildConfigEffects(workbenchPath, record.ID)
		_ = s.executor.Execute(ctx, configEffects)
	}

	return &primary.CreateWorkbenchResponse{
		WorkbenchID: record.ID,
		Workbench:   s.recordToWorkbench(record),
		Path:        workbenchPath,
	}, nil
}

// GetWorkbench retrieves a workbench by ID.
func (s *WorkbenchServiceImpl) GetWorkbench(ctx context.Context, workbenchID string) (*primary.Workbench, error) {
	record, err := s.workbenchRepo.GetByID(ctx, workbenchID)
	if err != nil {
		return nil, fmt.Errorf("workbench not found: %w", err)
	}
	return s.recordToWorkbench(record), nil
}

// GetWorkbenchByPath retrieves a workbench by its filesystem path.
func (s *WorkbenchServiceImpl) GetWorkbenchByPath(ctx context.Context, path string) (*primary.Workbench, error) {
	record, err := s.workbenchRepo.GetByPath(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("workbench not found at path: %w", err)
	}
	return s.recordToWorkbench(record), nil
}

// UpdateWorkbenchPath updates the filesystem path of a workbench.
func (s *WorkbenchServiceImpl) UpdateWorkbenchPath(ctx context.Context, workbenchID, newPath string) error {
	// Verify workbench exists
	_, err := s.workbenchRepo.GetByID(ctx, workbenchID)
	if err != nil {
		return fmt.Errorf("workbench not found: %w", err)
	}

	return s.workbenchRepo.UpdatePath(ctx, workbenchID, newPath)
}

// ListWorkbenches lists workbenches with optional filters.
func (s *WorkbenchServiceImpl) ListWorkbenches(ctx context.Context, filters primary.WorkbenchFilters) ([]*primary.Workbench, error) {
	records, err := s.workbenchRepo.List(ctx, filters.WorkshopID)
	if err != nil {
		return nil, fmt.Errorf("failed to list workbenches: %w", err)
	}

	workbenches := make([]*primary.Workbench, len(records))
	for i, r := range records {
		workbenches[i] = s.recordToWorkbench(r)
	}
	return workbenches, nil
}

// RenameWorkbench renames a workbench.
func (s *WorkbenchServiceImpl) RenameWorkbench(ctx context.Context, req primary.RenameWorkbenchRequest) error {
	// 1. Check workbench exists
	_, err := s.workbenchRepo.GetByID(ctx, req.WorkbenchID)
	workbenchExists := err == nil

	// 2. Guard check
	if result := coreworkbench.CanRenameWorkbench(workbenchExists, req.WorkbenchID); !result.Allowed {
		return result.Error()
	}

	// 3. Update name
	return s.workbenchRepo.Rename(ctx, req.WorkbenchID, req.NewName)
}

// DeleteWorkbench deletes a workbench.
func (s *WorkbenchServiceImpl) DeleteWorkbench(ctx context.Context, req primary.DeleteWorkbenchRequest) error {
	// 1. Fetch workbench
	_, err := s.workbenchRepo.GetByID(ctx, req.WorkbenchID)
	if err != nil {
		return fmt.Errorf("workbench not found: %w", err)
	}

	// 2. Count active work (simplified - could add task repo)
	activeTaskCount := 0

	// 3. Guard check
	guardCtx := coreworkbench.DeleteWorkbenchContext{
		WorkbenchID:     req.WorkbenchID,
		ActiveTaskCount: activeTaskCount,
		ForceDelete:     req.Force,
	}
	if result := coreworkbench.CanDeleteWorkbench(guardCtx); !result.Allowed {
		return result.Error()
	}

	// 4. Delete from database
	return s.workbenchRepo.Delete(ctx, req.WorkbenchID)
}

// Helper methods

func (s *WorkbenchServiceImpl) recordToWorkbench(r *secondary.WorkbenchRecord) *primary.Workbench {
	return &primary.Workbench{
		ID:            r.ID,
		Name:          r.Name,
		WorkshopID:    r.WorkshopID,
		RepoID:        r.RepoID,
		Path:          r.WorktreePath,
		Status:        r.Status,
		HomeBranch:    r.HomeBranch,
		CurrentBranch: r.CurrentBranch,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

func (s *WorkbenchServiceImpl) defaultBasePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "wb")
}

func (s *WorkbenchServiceImpl) pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// buildConfigEffects generates FileEffects for writing .orc/config.json
func (s *WorkbenchServiceImpl) buildConfigEffects(workbenchPath, workbenchID string) []effects.Effect {
	orcDir := filepath.Join(workbenchPath, ".orc")
	configPath := filepath.Join(orcDir, "config.json")

	cfg := &config.Config{
		Version: "1.0",
		PlaceID: workbenchID, // BENCH-XXX
	}
	configJSON, _ := json.MarshalIndent(cfg, "", "  ")

	return []effects.Effect{
		effects.FileEffect{Operation: "mkdir", Path: orcDir, Mode: 0755},
		effects.FileEffect{Operation: "write", Path: configPath, Content: configJSON, Mode: 0644},
	}
}

// CheckoutBranch switches to a target branch using stash dance.
func (s *WorkbenchServiceImpl) CheckoutBranch(ctx context.Context, req primary.CheckoutBranchRequest) (*primary.CheckoutBranchResponse, error) {
	// 1. Get workbench
	workbench, err := s.workbenchRepo.GetByID(ctx, req.WorkbenchID)
	if err != nil {
		return nil, fmt.Errorf("workbench not found: %w", err)
	}

	// 2. Check path exists
	if !s.pathExists(workbench.WorktreePath) {
		return nil, fmt.Errorf("workbench path does not exist: %s", workbench.WorktreePath)
	}

	// 3. Perform stash dance
	result, err := s.gitService.StashDance(workbench.WorktreePath, req.TargetBranch)
	if err != nil {
		return nil, fmt.Errorf("stash dance failed: %w", err)
	}

	// 4. Update current branch in database
	workbench.CurrentBranch = result.CurrentBranch
	if err := s.workbenchRepo.Update(ctx, workbench); err != nil {
		// Log but don't fail - the git operation succeeded
		fmt.Printf("Warning: failed to update current branch in database: %v\n", err)
	}

	return &primary.CheckoutBranchResponse{
		PreviousBranch: result.PreviousBranch,
		CurrentBranch:  result.CurrentBranch,
		StashApplied:   result.WasStashed && result.StashPopped,
	}, nil
}

// GetWorkbenchStatus returns the current git status of a workbench.
func (s *WorkbenchServiceImpl) GetWorkbenchStatus(ctx context.Context, workbenchID string) (*primary.WorkbenchGitStatus, error) {
	// 1. Get workbench
	workbench, err := s.workbenchRepo.GetByID(ctx, workbenchID)
	if err != nil {
		return nil, fmt.Errorf("workbench not found: %w", err)
	}

	status := &primary.WorkbenchGitStatus{
		WorkbenchID: workbenchID,
		HomeBranch:  workbench.HomeBranch,
	}

	// 2. Check if path exists
	if !s.pathExists(workbench.WorktreePath) {
		status.CurrentBranch = workbench.CurrentBranch // Use stored value
		return status, nil
	}

	// 3. Get current branch from git
	currentBranch, err := s.gitService.GetCurrentBranch(workbench.WorktreePath)
	if err == nil {
		status.CurrentBranch = currentBranch
		// Update database if different
		if currentBranch != workbench.CurrentBranch {
			workbench.CurrentBranch = currentBranch
			_ = s.workbenchRepo.Update(ctx, workbench)
		}
	} else {
		status.CurrentBranch = workbench.CurrentBranch
	}

	// 4. Get dirty state
	dirty, err := s.gitService.IsDirty(workbench.WorktreePath)
	if err == nil {
		status.IsDirty = dirty
	}

	// 5. Get dirty file count
	count, err := s.gitService.GetDirtyFileCount(workbench.WorktreePath)
	if err == nil {
		status.DirtyFiles = count
	}

	// 6. Get ahead/behind
	ahead, behind, err := s.gitService.GetAheadBehind(workbench.WorktreePath)
	if err == nil {
		status.AheadBy = ahead
		status.BehindBy = behind
	}

	return status, nil
}

// UpdateFocusedID sets or clears the focused container ID for a workbench.
func (s *WorkbenchServiceImpl) UpdateFocusedID(ctx context.Context, workbenchID, focusedID string) error {
	return s.workbenchRepo.UpdateFocusedID(ctx, workbenchID, focusedID)
}

// GetFocusedID returns the currently focused container ID for a workbench.
func (s *WorkbenchServiceImpl) GetFocusedID(ctx context.Context, workbenchID string) (string, error) {
	record, err := s.workbenchRepo.GetByID(ctx, workbenchID)
	if err != nil {
		return "", fmt.Errorf("workbench not found: %w", err)
	}
	return record.FocusedID, nil
}

// Ensure WorkbenchServiceImpl implements the interface
var _ primary.WorkbenchService = (*WorkbenchServiceImpl)(nil)
