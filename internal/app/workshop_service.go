package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/example/orc/internal/config"
	coreworkshop "github.com/example/orc/internal/core/workshop"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// WorkshopServiceImpl implements the WorkshopService interface.
type WorkshopServiceImpl struct {
	workshopRepo     secondary.WorkshopRepository
	workbenchRepo    secondary.WorkbenchRepository
	repoRepo         secondary.RepoRepository
	tmuxAdapter      secondary.TMuxAdapter
	workspaceAdapter secondary.WorkspaceAdapter
}

// NewWorkshopService creates a new WorkshopService with injected dependencies.
func NewWorkshopService(
	workshopRepo secondary.WorkshopRepository,
	workbenchRepo secondary.WorkbenchRepository,
	repoRepo secondary.RepoRepository,
	tmuxAdapter secondary.TMuxAdapter,
	workspaceAdapter secondary.WorkspaceAdapter,
) *WorkshopServiceImpl {
	return &WorkshopServiceImpl{
		workshopRepo:     workshopRepo,
		workbenchRepo:    workbenchRepo,
		repoRepo:         repoRepo,
		tmuxAdapter:      tmuxAdapter,
		workspaceAdapter: workspaceAdapter,
	}
}

// CreateWorkshop creates a new workshop in a factory.
func (s *WorkshopServiceImpl) CreateWorkshop(ctx context.Context, req primary.CreateWorkshopRequest) (*primary.CreateWorkshopResponse, error) {
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

	// 4. Delete
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

	// 2. Check if session already exists
	sessionExists := s.tmuxAdapter.SessionExists(ctx, req.WorkshopID)

	// 3. Compute gatehouse path and check existence
	home, _ := os.UserHomeDir()
	gatehouseDir := coreworkshop.GatehousePath(home, req.WorkshopID, workshop.Name)
	gatehouseDirExists := s.dirExists(gatehouseDir)
	gatehouseConfigExists := s.fileExists(filepath.Join(gatehouseDir, ".orc", "config.json"))

	// 4. Get workbenches and check each one's state
	workbenches, _ := s.workbenchRepo.List(ctx, req.WorkshopID)
	var wbInputs []coreworkshop.WorkbenchPlanInput
	for _, wb := range workbenches {
		repoName := ""
		if wb.RepoID != "" {
			if repo, err := s.repoRepo.GetByID(ctx, wb.RepoID); err == nil {
				repoName = repo.Name
			}
		}
		wbInputs = append(wbInputs, coreworkshop.WorkbenchPlanInput{
			ID:             wb.ID,
			Name:           wb.Name,
			WorktreePath:   wb.WorktreePath,
			RepoName:       repoName,
			HomeBranch:     wb.HomeBranch,
			WorktreeExists: s.dirExists(wb.WorktreePath),
			ConfigExists:   s.fileExists(filepath.Join(wb.WorktreePath, ".orc", "config.json")),
		})
	}

	// 5. Generate plan using pure function
	input := coreworkshop.OpenPlanInput{
		WorkshopID:            req.WorkshopID,
		WorkshopName:          workshop.Name,
		SessionExists:         sessionExists,
		GatehouseDir:          gatehouseDir,
		GatehouseDirExists:    gatehouseDirExists,
		GatehouseConfigExists: gatehouseConfigExists,
		Workbenches:           wbInputs,
	}
	corePlan := coreworkshop.GenerateOpenPlan(input)

	// 6. Convert core plan to primary plan
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

	// 1. Create gatehouse if needed (check Exists flag since GatehouseOp is always set)
	home, _ := os.UserHomeDir()
	gatehouseDir := coreworkshop.GatehousePath(home, plan.WorkshopID, plan.WorkshopName)
	if plan.GatehouseOp != nil && (!plan.GatehouseOp.Exists || !plan.GatehouseOp.ConfigExists) {
		gatehouseDir = s.createGatehouseDir(plan.WorkshopID, plan.WorkshopName)
	}

	// 2. Create workbenches if needed
	workbenches, _ := s.workbenchRepo.List(ctx, plan.WorkshopID)
	for _, wb := range workbenches {
		if err := s.ensureWorktreeExists(ctx, wb); err != nil {
			return nil, fmt.Errorf("failed to create worktree for %s: %w", wb.Name, err)
		}
	}

	// 3. Create tmux session if needed
	if plan.TMuxOp != nil {
		if err := s.tmuxAdapter.CreateSession(ctx, plan.SessionName, gatehouseDir); err != nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}

		// Setup Gatehouse (ORC) window
		if err := s.tmuxAdapter.CreateOrcWindow(ctx, plan.SessionName, gatehouseDir); err != nil {
			return nil, fmt.Errorf("failed to create gatehouse: %w", err)
		}

		// Create tmux windows for each workbench
		for i, wb := range workbenches {
			_ = s.tmuxAdapter.CreateWorkbenchWindow(ctx, plan.SessionName, i+2, wb.Name, wb.WorktreePath)
		}
	}

	return &primary.OpenWorkshopResponse{
		Workshop:           s.recordToWorkshop(workshop),
		SessionName:        plan.SessionName,
		SessionAlreadyOpen: false,
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
func (s *WorkshopServiceImpl) ensureWorktreeExists(ctx context.Context, wb *secondary.WorkbenchRecord) error {
	// Check if worktree already exists
	exists, err := s.workspaceAdapter.WorktreeExists(ctx, wb.WorktreePath)
	if err != nil {
		return err
	}

	if !exists {
		// Need repo to create worktree
		if wb.RepoID == "" {
			// No repo linked - just create the directory
			if err := s.workspaceAdapter.CreateDirectory(ctx, wb.WorktreePath); err != nil {
				return err
			}
		} else {
			// Get repo name
			repo, err := s.repoRepo.GetByID(ctx, wb.RepoID)
			if err != nil {
				return fmt.Errorf("repo %s not found: %w", wb.RepoID, err)
			}

			// Create worktree
			if err := s.workspaceAdapter.CreateWorktree(ctx, repo.Name, wb.HomeBranch, wb.WorktreePath); err != nil {
				return err
			}
		}
	}

	// Ensure IMP config exists
	return s.ensureWorkbenchConfig(wb)
}

// ensureWorkbenchConfig creates the .orc/config.json for a workbench with IMP role.
func (s *WorkshopServiceImpl) ensureWorkbenchConfig(wb *secondary.WorkbenchRecord) error {
	orcDir := filepath.Join(wb.WorktreePath, ".orc")
	configPath := filepath.Join(orcDir, "config.json")

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return nil // Already exists
	}

	// Create .orc directory
	if err := os.MkdirAll(orcDir, 0755); err != nil {
		return fmt.Errorf("failed to create .orc dir: %w", err)
	}

	// Create config with IMP role
	cfg := &config.Config{
		Version:     "1.0",
		Role:        config.RoleIMP,
		WorkbenchID: wb.ID,
	}
	return config.SaveConfig(wb.WorktreePath, cfg)
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

// createGatehouseDir creates the Gatehouse directory for a workshop
// at ~/.orc/ws/{workshop-id}-{slug}/ with a Goblin config
func (s *WorkshopServiceImpl) createGatehouseDir(workshopID, workshopName string) string {
	home, _ := os.UserHomeDir()
	slug := slugify(workshopName)
	dirName := fmt.Sprintf("%s-%s", workshopID, slug)
	dir := filepath.Join(home, ".orc", "ws", dirName)
	_ = os.MkdirAll(dir, 0755)

	// Create .orc subdir for config
	orcDir := filepath.Join(dir, ".orc")
	_ = os.MkdirAll(orcDir, 0755)

	// Create config.json with Goblin role
	cfg := &config.Config{
		Version: "1.0",
		Role:    config.RoleGoblin,
	}
	_ = config.SaveConfig(dir, cfg)

	return dir
}

// slugify converts a name to a URL-friendly slug
func slugify(name string) string {
	// Lowercase and replace spaces with hyphens
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove any characters that aren't alphanumeric or hyphens
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
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
		SessionName:  core.SessionName,
		NothingToDo:  core.NothingToDo,
	}

	if core.GatehouseOp != nil {
		plan.GatehouseOp = &primary.GatehouseOp{
			Path:         core.GatehouseOp.Path,
			Exists:       core.GatehouseOp.Exists,
			ConfigExists: core.GatehouseOp.ConfigExists,
		}
	}

	for _, wb := range core.WorkbenchOps {
		plan.WorkbenchOps = append(plan.WorkbenchOps, primary.WorkbenchOp{
			Name:         wb.Name,
			Path:         wb.Path,
			Exists:       wb.Exists,
			RepoName:     wb.RepoName,
			Branch:       wb.Branch,
			ConfigExists: wb.ConfigExists,
		})
	}

	if core.TMuxOp != nil {
		plan.TMuxOp = &primary.TMuxOp{
			SessionName: core.TMuxOp.SessionName,
			Windows:     core.TMuxOp.Windows,
		}
	}

	return plan
}

// Ensure WorkshopServiceImpl implements the interface
var _ primary.WorkshopService = (*WorkshopServiceImpl)(nil)
