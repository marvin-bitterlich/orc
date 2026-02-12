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
	workbenchRepo    secondary.WorkbenchRepository
	workshopRepo     secondary.WorkshopRepository
	repoRepo         secondary.RepoRepository
	agentProvider    secondary.AgentIdentityProvider
	executor         EffectExecutor
	gitService       *GitService
	workspaceAdapter secondary.WorkspaceAdapter
}

// NewWorkbenchService creates a new WorkbenchService with injected dependencies.
func NewWorkbenchService(
	workbenchRepo secondary.WorkbenchRepository,
	workshopRepo secondary.WorkshopRepository,
	repoRepo secondary.RepoRepository,
	agentProvider secondary.AgentIdentityProvider,
	executor EffectExecutor,
	workspaceAdapter secondary.WorkspaceAdapter,
) *WorkbenchServiceImpl {
	return &WorkbenchServiceImpl{
		workbenchRepo:    workbenchRepo,
		workshopRepo:     workshopRepo,
		repoRepo:         repoRepo,
		agentProvider:    agentProvider,
		executor:         executor,
		gitService:       NewGitService(),
		workspaceAdapter: workspaceAdapter,
	}
}

// CreateWorkbench creates a new workbench.
func (s *WorkbenchServiceImpl) CreateWorkbench(ctx context.Context, req primary.CreateWorkbenchRequest) (*primary.CreateWorkbenchResponse, error) {
	// 1. Check if workshop exists
	workshopExists, err := s.workbenchRepo.WorkshopExists(ctx, req.WorkshopID)
	if err != nil {
		return nil, fmt.Errorf("failed to check workshop: %w", err)
	}

	// 2. Guard check
	guardCtx := coreworkbench.CreateWorkbenchContext{
		WorkshopID:     req.WorkshopID,
		WorkshopExists: workshopExists,
	}
	if result := coreworkbench.CanCreateWorkbench(guardCtx); !result.Allowed {
		return nil, result.Error()
	}

	// 3. Get next workbench ID to extract number for auto-generated name
	nextID, err := s.workbenchRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get next workbench ID: %w", err)
	}
	benchNumber := coreworkbench.ParseWorkbenchNumber(nextID)

	// 5. Auto-generate name if not provided (requires RepoID)
	name := req.Name
	if name == "" {
		if req.RepoID == "" {
			return nil, fmt.Errorf("name is required when repo_id is not provided")
		}
		repo, err := s.repoRepo.GetByID(ctx, req.RepoID)
		if err != nil {
			return nil, fmt.Errorf("failed to get repo for name generation: %w", err)
		}
		name = fmt.Sprintf("%s-%03d", repo.Name, benchNumber)
	}

	// 6. Compute workbench path (deterministic: ~/wb/<name>)
	workbenchPath := coreworkbench.ComputePath(name)

	// 7. Generate home branch name
	homeBranch := GenerateHomeBranchName(UserInitials, name)

	// 8. Create workbench record in DB
	record := &secondary.WorkbenchRecord{
		Name:          name,
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

	// 9. Create worktree/directory and config immediately (not deferred to infra apply)
	if err := s.ensureWorktreeExists(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create worktree: %w", err)
	}
	if err := s.ensureConfigExists(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
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
	// 1. Check workbench exists
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

	// 4. Delete from database (infrastructure cleanup handled by orc tmux apply)
	return s.workbenchRepo.Delete(ctx, req.WorkbenchID)
}

// Helper methods

func (s *WorkbenchServiceImpl) recordToWorkbench(r *secondary.WorkbenchRecord) *primary.Workbench {
	return &primary.Workbench{
		ID:            r.ID,
		Name:          r.Name,
		WorkshopID:    r.WorkshopID,
		RepoID:        r.RepoID,
		Path:          coreworkbench.ComputePath(r.Name),
		Status:        r.Status,
		HomeBranch:    r.HomeBranch,
		CurrentBranch: r.CurrentBranch,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

func (s *WorkbenchServiceImpl) pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ensureWorktreeExists creates a worktree (or directory if no repo) if it doesn't already exist.
func (s *WorkbenchServiceImpl) ensureWorktreeExists(ctx context.Context, wb *secondary.WorkbenchRecord) error {
	wbPath := coreworkbench.ComputePath(wb.Name)

	// Check if worktree already exists (idempotent)
	exists, err := s.workspaceAdapter.WorktreeExists(ctx, wbPath)
	if err != nil {
		return err
	}
	if exists {
		return nil // Already exists, nothing to do
	}

	var effs []effects.Effect

	// If no repo linked, just create a directory
	if wb.RepoID == "" {
		effs = append(effs, effects.FileEffect{
			Operation: "mkdir",
			Path:      wbPath,
			Mode:      0755,
		})
	} else {
		// Linked to repo - create git worktree
		repo, err := s.repoRepo.GetByID(ctx, wb.RepoID)
		if err != nil {
			return fmt.Errorf("repo %s not found: %w", wb.RepoID, err)
		}
		effs = append(effs, effects.GitEffect{
			Operation: "worktree_add",
			RepoPath:  repo.LocalPath,
			Args:      []string{wb.HomeBranch, wbPath},
		})
	}

	if len(effs) > 0 {
		return s.executor.Execute(ctx, effs)
	}
	return nil
}

// ensureConfigExists creates the .orc/config.json file if it doesn't already exist.
func (s *WorkbenchServiceImpl) ensureConfigExists(ctx context.Context, wb *secondary.WorkbenchRecord) error {
	wbPath := coreworkbench.ComputePath(wb.Name)
	orcDir := filepath.Join(wbPath, ".orc")
	configPath := filepath.Join(orcDir, "config.json")

	// Check if config already exists (idempotent)
	if _, err := os.Stat(configPath); err == nil {
		return nil // Already exists, nothing to do
	}

	cfg := &config.Config{
		Version: "1.0",
		PlaceID: wb.ID,
	}
	configJSON, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	effs := []effects.Effect{
		effects.FileEffect{Operation: "mkdir", Path: orcDir, Mode: 0755},
		effects.FileEffect{Operation: "write", Path: configPath, Content: configJSON, Mode: 0644},
	}

	return s.executor.Execute(ctx, effs)
}

// CheckoutBranch switches to a target branch using stash dance.
func (s *WorkbenchServiceImpl) CheckoutBranch(ctx context.Context, req primary.CheckoutBranchRequest) (*primary.CheckoutBranchResponse, error) {
	// 1. Get workbench
	workbench, err := s.workbenchRepo.GetByID(ctx, req.WorkbenchID)
	if err != nil {
		return nil, fmt.Errorf("workbench not found: %w", err)
	}

	// 2. Compute path and check it exists
	wbPath := coreworkbench.ComputePath(workbench.Name)
	if !s.pathExists(wbPath) {
		return nil, fmt.Errorf("workbench path does not exist: %s", wbPath)
	}

	// 3. Perform stash dance
	result, err := s.gitService.StashDance(wbPath, req.TargetBranch)
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

	// Compute path
	wbPath := coreworkbench.ComputePath(workbench.Name)

	status := &primary.WorkbenchGitStatus{
		WorkbenchID: workbenchID,
		HomeBranch:  workbench.HomeBranch,
	}

	// 2. Check if path exists
	if !s.pathExists(wbPath) {
		status.CurrentBranch = workbench.CurrentBranch // Use stored value
		return status, nil
	}

	// 3. Get current branch from git
	currentBranch, err := s.gitService.GetCurrentBranch(wbPath)
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
	dirty, err := s.gitService.IsDirty(wbPath)
	if err == nil {
		status.IsDirty = dirty
	}

	// 5. Get dirty file count
	count, err := s.gitService.GetDirtyFileCount(wbPath)
	if err == nil {
		status.DirtyFiles = count
	}

	// 6. Get ahead/behind
	ahead, behind, err := s.gitService.GetAheadBehind(wbPath)
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

// GetWorkbenchesByFocusedID returns all active workbenches focused on a given container.
// Used for focus exclusivity checks (IMP cannot focus on container already focused by another IMP).
func (s *WorkbenchServiceImpl) GetWorkbenchesByFocusedID(ctx context.Context, focusedID string) ([]*primary.Workbench, error) {
	records, err := s.workbenchRepo.GetByFocusedID(ctx, focusedID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workbenches by focused ID: %w", err)
	}

	workbenches := make([]*primary.Workbench, len(records))
	for i, r := range records {
		workbenches[i] = s.recordToWorkbench(r)
	}
	return workbenches, nil
}

// ArchiveWorkbench soft-deletes a workbench by setting status to 'archived'.
func (s *WorkbenchServiceImpl) ArchiveWorkbench(ctx context.Context, workbenchID string) error {
	record, err := s.workbenchRepo.GetByID(ctx, workbenchID)
	if err != nil {
		return fmt.Errorf("workbench not found: %w", err)
	}
	if record.Status == "archived" {
		return fmt.Errorf("workbench %s is already archived", workbenchID)
	}
	record.Status = "archived"
	return s.workbenchRepo.Update(ctx, record)
}

// Ensure WorkbenchServiceImpl implements the interface
var _ primary.WorkbenchService = (*WorkbenchServiceImpl)(nil)
