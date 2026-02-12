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
	coreworkshop "github.com/example/orc/internal/core/workshop"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// WorkshopServiceImpl implements the WorkshopService interface.
type WorkshopServiceImpl struct {
	factoryRepo      secondary.FactoryRepository
	workshopRepo     secondary.WorkshopRepository
	workbenchRepo    secondary.WorkbenchRepository
	repoRepo         secondary.RepoRepository
	tmuxAdapter      secondary.TMuxAdapter
	workspaceAdapter secondary.WorkspaceAdapter
	executor         EffectExecutor
}

// NewWorkshopService creates a new WorkshopService with injected dependencies.
func NewWorkshopService(
	factoryRepo secondary.FactoryRepository,
	workshopRepo secondary.WorkshopRepository,
	workbenchRepo secondary.WorkbenchRepository,
	repoRepo secondary.RepoRepository,
	tmuxAdapter secondary.TMuxAdapter,
	workspaceAdapter secondary.WorkspaceAdapter,
	executor EffectExecutor,
) *WorkshopServiceImpl {
	return &WorkshopServiceImpl{
		factoryRepo:      factoryRepo,
		workshopRepo:     workshopRepo,
		workbenchRepo:    workbenchRepo,
		repoRepo:         repoRepo,
		tmuxAdapter:      tmuxAdapter,
		workspaceAdapter: workspaceAdapter,
		executor:         executor,
	}
}

// CreateWorkshop creates a new workshop in a factory.
func (s *WorkshopServiceImpl) CreateWorkshop(ctx context.Context, req primary.CreateWorkshopRequest) (*primary.CreateWorkshopResponse, error) {
	// If no factory specified, use default
	if req.FactoryID == "" {
		factory, err := s.getOrCreateDefaultFactory(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get default factory: %w", err)
		}
		req.FactoryID = factory.ID
	}

	// 1. Check if factory exists
	factoryExists, err := s.workshopRepo.FactoryExists(ctx, req.FactoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to check factory: %w", err)
	}

	// 2. Guard check
	guardCtx := coreworkshop.CreateWorkshopContext{
		FactoryID:     req.FactoryID,
		FactoryExists: factoryExists,
	}
	if result := coreworkshop.CanCreateWorkshop(guardCtx); !result.Allowed {
		return nil, result.Error()
	}

	// 3. Create workshop record (ID and name generation handled by repo)
	record := &secondary.WorkshopRecord{
		FactoryID: req.FactoryID,
		Name:      req.Name, // May be empty - repo will use name pool
		Status:    "active",
	}
	if err := s.workshopRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create workshop: %w", err)
	}

	return &primary.CreateWorkshopResponse{
		WorkshopID: record.ID,
		Workshop:   s.recordToWorkshop(record),
	}, nil
}

// getOrCreateDefaultFactory returns the "default" factory, creating it if needed.
func (s *WorkshopServiceImpl) getOrCreateDefaultFactory(ctx context.Context) (*secondary.FactoryRecord, error) {
	// Try to get existing "default" factory
	factory, err := s.factoryRepo.GetByName(ctx, "default")
	if err == nil {
		return factory, nil
	}

	// Create default factory
	id, err := s.factoryRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate factory ID: %w", err)
	}

	record := &secondary.FactoryRecord{
		ID:     id,
		Name:   "default",
		Status: "active",
	}
	if err := s.factoryRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create default factory: %w", err)
	}

	return record, nil
}

// GetWorkshop retrieves a workshop by ID.
func (s *WorkshopServiceImpl) GetWorkshop(ctx context.Context, workshopID string) (*primary.Workshop, error) {
	record, err := s.workshopRepo.GetByID(ctx, workshopID)
	if err != nil {
		return nil, fmt.Errorf("workshop not found: %w", err)
	}
	return s.recordToWorkshop(record), nil
}

// ListWorkshops lists workshops with optional filters.
func (s *WorkshopServiceImpl) ListWorkshops(ctx context.Context, filters primary.WorkshopFilters) ([]*primary.Workshop, error) {
	records, err := s.workshopRepo.List(ctx, secondary.WorkshopFilters{
		FactoryID: filters.FactoryID,
		Status:    filters.Status,
		Limit:     filters.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list workshops: %w", err)
	}

	workshops := make([]*primary.Workshop, len(records))
	for i, r := range records {
		workshops[i] = s.recordToWorkshop(r)
	}
	return workshops, nil
}

// UpdateWorkshop updates workshop name.
func (s *WorkshopServiceImpl) UpdateWorkshop(ctx context.Context, req primary.UpdateWorkshopRequest) error {
	// 1. Check workshop exists
	_, err := s.workshopRepo.GetByID(ctx, req.WorkshopID)
	if err != nil {
		return fmt.Errorf("workshop not found: %w", err)
	}

	// 2. Update
	record := &secondary.WorkshopRecord{
		ID:   req.WorkshopID,
		Name: req.Name,
	}
	if err := s.workshopRepo.Update(ctx, record); err != nil {
		return fmt.Errorf("failed to update workshop: %w", err)
	}

	return nil
}

// DeleteWorkshop deletes a workshop.
func (s *WorkshopServiceImpl) DeleteWorkshop(ctx context.Context, req primary.DeleteWorkshopRequest) error {
	// 1. Check workshop exists
	_, err := s.workshopRepo.GetByID(ctx, req.WorkshopID)
	workshopExists := err == nil

	// 2. Count workbenches
	workbenchCount, _ := s.workshopRepo.CountWorkbenches(ctx, req.WorkshopID)

	// 3. Guard check
	guardCtx := coreworkshop.DeleteWorkshopContext{
		WorkshopID:     req.WorkshopID,
		WorkshopExists: workshopExists,
		WorkbenchCount: workbenchCount,
		ForceDelete:    req.Force,
	}
	if result := coreworkshop.CanDeleteWorkshop(guardCtx); !result.Allowed {
		return result.Error()
	}

	// 4. Delete from database (infrastructure cleanup handled by orc tmux apply)
	return s.workshopRepo.Delete(ctx, req.WorkshopID)
}

// PlanOpenWorkshop generates a plan for opening a workshop without executing it.
// This gathers current state and computes what operations would be needed.
func (s *WorkshopServiceImpl) PlanOpenWorkshop(ctx context.Context, req primary.OpenWorkshopRequest) (*primary.OpenWorkshopPlan, error) {
	// 1. Get workshop
	workshop, err := s.workshopRepo.GetByID(ctx, req.WorkshopID)
	if err != nil {
		return nil, fmt.Errorf("workshop not found: %w", err)
	}

	// 2. Get factory info
	factory, err := s.factoryRepo.GetByID(ctx, workshop.FactoryID)
	if err != nil {
		return nil, fmt.Errorf("factory not found: %w", err)
	}

	// 3. Find existing session by workshop ID env var (survives session renames)
	actualSessionName := s.tmuxAdapter.FindSessionByWorkshopID(ctx, req.WorkshopID)
	sessionExists := actualSessionName != ""

	// 4. Get existing windows if session exists
	var existingWindows []string
	if sessionExists {
		existingWindows, _ = s.tmuxAdapter.ListWindows(ctx, actualSessionName)
	}

	// 5. Get workbenches and check each one's state
	workbenches, _ := s.workbenchRepo.List(ctx, req.WorkshopID)
	var wbInputs []coreworkshop.WorkbenchPlanInput
	for _, wb := range workbenches {
		repoName := ""
		if wb.RepoID != "" {
			if repo, err := s.repoRepo.GetByID(ctx, wb.RepoID); err == nil {
				repoName = repo.Name
			}
		}
		wbPath := coreworkbench.ComputePath(wb.Name)
		wbInputs = append(wbInputs, coreworkshop.WorkbenchPlanInput{
			ID:             wb.ID,
			Name:           wb.Name,
			WorktreePath:   wbPath,
			RepoName:       repoName,
			HomeBranch:     wb.HomeBranch,
			WorktreeExists: s.dirExists(wbPath),
			ConfigExists:   s.fileExists(filepath.Join(wbPath, ".orc", "config.json")),
			Status:         wb.Status,
		})
	}

	// 7. Generate plan using pure function
	input := coreworkshop.OpenPlanInput{
		WorkshopID:        req.WorkshopID,
		WorkshopName:      workshop.Name,
		FactoryID:         factory.ID,
		FactoryName:       factory.Name,
		SessionExists:     sessionExists,
		ActualSessionName: actualSessionName,
		ExistingWindows:   existingWindows,
		Workbenches:       wbInputs,
	}
	corePlan := coreworkshop.GenerateOpenPlan(input)

	// 8. Convert core plan to primary plan
	return s.corePlanToPrimary(&corePlan), nil
}

// ApplyOpenWorkshop executes a previously generated open plan.
func (s *WorkshopServiceImpl) ApplyOpenWorkshop(ctx context.Context, plan *primary.OpenWorkshopPlan) (*primary.OpenWorkshopResponse, error) {
	// Get workshop for response
	workshop, err := s.workshopRepo.GetByID(ctx, plan.WorkshopID)
	if err != nil {
		return nil, fmt.Errorf("workshop not found: %w", err)
	}

	// If nothing to do, just return attach instructions
	if plan.NothingToDo {
		return &primary.OpenWorkshopResponse{
			Workshop:           s.recordToWorkshop(workshop),
			SessionName:        plan.SessionName,
			SessionAlreadyOpen: true,
			AttachInstructions: s.tmuxAdapter.AttachInstructions(plan.SessionName),
		}, nil
	}

	// 1. Create workbenches if needed
	workbenches, _ := s.workbenchRepo.List(ctx, plan.WorkshopID)
	for _, wb := range workbenches {
		if err := s.ensureWorktreeExists(ctx, wb); err != nil {
			return nil, fmt.Errorf("failed to create worktree for %s: %w", wb.Name, err)
		}
	}

	// 3. TMux lifecycle removed - now handled by gotmux via `orc tmux apply`
	sessionAlreadyOpen := false
	if plan.TMuxOp != nil && plan.TMuxOp.AddToExisting {
		// Session detection only (lifecycle creation removed)
		sessionAlreadyOpen = true
	}

	return &primary.OpenWorkshopResponse{
		Workshop:           s.recordToWorkshop(workshop),
		SessionName:        plan.SessionName,
		SessionAlreadyOpen: sessionAlreadyOpen,
		AttachInstructions: s.tmuxAdapter.AttachInstructions(plan.SessionName),
	}, nil
}

// OpenWorkshop launches a TMux session for the workshop (plan + apply in one call).
// This is kept for backward compatibility.
func (s *WorkshopServiceImpl) OpenWorkshop(ctx context.Context, req primary.OpenWorkshopRequest) (*primary.OpenWorkshopResponse, error) {
	plan, err := s.PlanOpenWorkshop(ctx, req)
	if err != nil {
		return nil, err
	}
	return s.ApplyOpenWorkshop(ctx, plan)
}

// ensureWorktreeExists creates a worktree and IMP config if they don't exist.
// All effects are batched into a single execute call for atomicity.
func (s *WorkshopServiceImpl) ensureWorktreeExists(ctx context.Context, wb *secondary.WorkbenchRecord) error {
	// Compute path from name
	wbPath := coreworkbench.ComputePath(wb.Name)

	// Check if worktree already exists
	exists, err := s.workspaceAdapter.WorktreeExists(ctx, wbPath)
	if err != nil {
		return err
	}

	var effs []effects.Effect

	// 1. Build worktree effect (only if needed)
	if !exists {
		if wb.RepoID == "" {
			// No repo linked - just create the directory via FileEffect
			effs = append(effs, effects.FileEffect{
				Operation: "mkdir",
				Path:      wbPath,
				Mode:      0755,
			})
		} else {
			// Get repo name
			repo, err := s.repoRepo.GetByID(ctx, wb.RepoID)
			if err != nil {
				return fmt.Errorf("repo %s not found: %w", wb.RepoID, err)
			}

			// Create worktree via GitEffect
			effs = append(effs, effects.GitEffect{
				Operation: "worktree_add",
				RepoPath:  repo.LocalPath,
				Args:      []string{wb.HomeBranch, wbPath},
			})
		}
	}

	// 2. Build config effects (always - idempotent)
	orcDir := filepath.Join(wbPath, ".orc")
	configPath := filepath.Join(orcDir, "config.json")
	cfg := &config.Config{
		Version: "1.0",
		PlaceID: wb.ID, // BENCH-XXX
	}
	configJSON, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	effs = append(effs,
		effects.FileEffect{Operation: "mkdir", Path: orcDir, Mode: 0755},
		effects.FileEffect{Operation: "write", Path: configPath, Content: configJSON, Mode: 0644},
	)

	// 3. Execute ALL effects in one batch
	if len(effs) > 0 {
		return s.executor.Execute(ctx, effs)
	}
	return nil
}

// CloseWorkshop kills the workshop's TMux session.
func (s *WorkshopServiceImpl) CloseWorkshop(ctx context.Context, workshopID string) error {
	if !s.tmuxAdapter.SessionExists(ctx, workshopID) {
		return nil // Session not running, nothing to do
	}

	if err := s.tmuxAdapter.KillSession(ctx, workshopID); err != nil {
		return fmt.Errorf("failed to close session: %w", err)
	}

	return nil
}

// Helper methods

func (s *WorkshopServiceImpl) recordToWorkshop(r *secondary.WorkshopRecord) *primary.Workshop {
	return &primary.Workshop{
		ID:        r.ID,
		FactoryID: r.FactoryID,
		Name:      r.Name,
		Status:    r.Status,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

func (s *WorkshopServiceImpl) dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func (s *WorkshopServiceImpl) fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func (s *WorkshopServiceImpl) corePlanToPrimary(core *coreworkshop.OpenWorkshopPlan) *primary.OpenWorkshopPlan {
	plan := &primary.OpenWorkshopPlan{
		WorkshopID:   core.WorkshopID,
		WorkshopName: core.WorkshopName,
		FactoryID:    core.FactoryID,
		FactoryName:  core.FactoryName,
		SessionName:  core.SessionName,
		NothingToDo:  core.NothingToDo,
	}

	// Map DB state
	if len(core.Workbenches) > 0 {
		plan.DBState = &primary.DBStatePlan{
			WorkbenchCount: len(core.Workbenches),
		}
		for _, wb := range core.Workbenches {
			plan.DBState.Workbenches = append(plan.DBState.Workbenches, primary.WorkbenchDBState{
				ID:     wb.ID,
				Name:   wb.Name,
				Path:   wb.Path,
				Status: wb.Status,
			})
		}
	}

	// Map gatehouse op with derived status
	// Map workbench ops with derived status
	for _, wb := range core.WorkbenchOps {
		plan.WorkbenchOps = append(plan.WorkbenchOps, primary.WorkbenchOp{
			ID:           wb.ID,
			Name:         wb.Name,
			Path:         wb.Path,
			Exists:       wb.Exists,
			RepoName:     wb.RepoName,
			Branch:       wb.Branch,
			ConfigExists: wb.ConfigExists,
			Status:       boolToOpStatus(wb.Exists),
			ConfigStatus: boolToOpStatus(wb.ConfigExists),
		})
	}

	// Map tmux op with derived status
	if core.TMuxOp != nil {
		var windows []primary.TMuxWindowOp
		for _, w := range core.TMuxOp.Windows {
			windows = append(windows, primary.TMuxWindowOp{
				Index:  w.Index,
				Name:   w.Name,
				Path:   w.Path,
				Status: primary.OpCreate, // New windows are always CREATE
			})
		}

		// Session status depends on whether we're creating or adding to existing
		sessionStatus := primary.OpCreate
		if core.TMuxOp.AddToExisting {
			sessionStatus = primary.OpExists
		}

		plan.TMuxOp = &primary.TMuxOp{
			SessionName:   core.TMuxOp.SessionName,
			SessionStatus: sessionStatus,
			Windows:       windows,
			AddToExisting: core.TMuxOp.AddToExisting,
		}
	}

	return plan
}

// boolToOpStatus converts an existence flag to an OpStatus.
func boolToOpStatus(exists bool) primary.OpStatus {
	if exists {
		return primary.OpExists
	}
	return primary.OpCreate
}

// SetActiveCommission sets the active commission for a workshop (Goblin context).
// Pass empty string to clear.
func (s *WorkshopServiceImpl) SetActiveCommission(ctx context.Context, workshopID, commissionID string) error {
	return s.workshopRepo.SetActiveCommissionID(ctx, workshopID, commissionID)
}

// GetActiveCommission returns the active commission ID for a workshop.
func (s *WorkshopServiceImpl) GetActiveCommission(ctx context.Context, workshopID string) (string, error) {
	record, err := s.workshopRepo.GetByID(ctx, workshopID)
	if err != nil {
		return "", fmt.Errorf("workshop not found: %w", err)
	}
	return record.ActiveCommissionID, nil
}

// GetActiveCommissions returns commission IDs derived from focus:
// - Gatehouse focused_id (resolved to commission)
// - All workbench focused_ids in workshop (resolved to commission)
// Returns deduplicated commission IDs.
func (s *WorkshopServiceImpl) GetActiveCommissions(ctx context.Context, workshopID string) ([]string, error) {
	return s.workshopRepo.GetActiveCommissions(ctx, workshopID)
}

// ArchiveWorkshop soft-deletes a workshop by setting status to 'archived'.
func (s *WorkshopServiceImpl) ArchiveWorkshop(ctx context.Context, workshopID string) error {
	record, err := s.workshopRepo.GetByID(ctx, workshopID)
	if err != nil {
		return fmt.Errorf("workshop not found: %w", err)
	}
	if record.Status == "archived" {
		return fmt.Errorf("workshop %s is already archived", workshopID)
	}

	// Check for active (non-archived) workbenches
	workbenches, err := s.workbenchRepo.GetByWorkshop(ctx, workshopID)
	if err != nil {
		return fmt.Errorf("failed to check workbenches: %w", err)
	}
	activeCount := 0
	for _, wb := range workbenches {
		if wb.Status != "archived" {
			activeCount++
		}
	}
	if activeCount > 0 {
		return fmt.Errorf("cannot archive workshop: %d active workbench(es) remaining. Archive them first with 'orc workbench archive'", activeCount)
	}

	record.Status = "archived"
	return s.workshopRepo.Update(ctx, record)
}

// Ensure WorkshopServiceImpl implements the interface
var _ primary.WorkshopService = (*WorkshopServiceImpl)(nil)
