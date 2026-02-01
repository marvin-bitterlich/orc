package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/example/orc/internal/config"
	"github.com/example/orc/internal/core/effects"
	coreinfra "github.com/example/orc/internal/core/infra"
	coreworkshop "github.com/example/orc/internal/core/workshop"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// InfraServiceImpl implements the InfraService interface.
type InfraServiceImpl struct {
	factoryRepo      secondary.FactoryRepository
	workshopRepo     secondary.WorkshopRepository
	workbenchRepo    secondary.WorkbenchRepository
	repoRepo         secondary.RepoRepository
	gatehouseRepo    secondary.GatehouseRepository
	workspaceAdapter secondary.WorkspaceAdapter
	executor         EffectExecutor
}

// NewInfraService creates a new InfraService with injected dependencies.
func NewInfraService(
	factoryRepo secondary.FactoryRepository,
	workshopRepo secondary.WorkshopRepository,
	workbenchRepo secondary.WorkbenchRepository,
	repoRepo secondary.RepoRepository,
	gatehouseRepo secondary.GatehouseRepository,
	workspaceAdapter secondary.WorkspaceAdapter,
	executor EffectExecutor,
) *InfraServiceImpl {
	return &InfraServiceImpl{
		factoryRepo:      factoryRepo,
		workshopRepo:     workshopRepo,
		workbenchRepo:    workbenchRepo,
		repoRepo:         repoRepo,
		gatehouseRepo:    gatehouseRepo,
		workspaceAdapter: workspaceAdapter,
		executor:         executor,
	}
}

// PlanInfra generates a plan showing infrastructure state for a workshop.
func (s *InfraServiceImpl) PlanInfra(ctx context.Context, req primary.InfraPlanRequest) (*primary.InfraPlan, error) {
	// 1. Get workshop
	workshop, err := s.workshopRepo.GetByID(ctx, req.WorkshopID)
	if err != nil {
		return nil, fmt.Errorf("workshop not found: %w", err)
	}

	// 2. Get factory
	factory, err := s.factoryRepo.GetByID(ctx, workshop.FactoryID)
	if err != nil {
		return nil, fmt.Errorf("factory not found: %w", err)
	}

	// 3. Get gatehouse (may not exist)
	var gatehouseID string
	gatehouse, err := s.gatehouseRepo.GetByWorkshop(ctx, req.WorkshopID)
	if err == nil {
		gatehouseID = gatehouse.ID
	}

	// 4. Compute gatehouse path and check existence
	home, _ := os.UserHomeDir()
	gatehousePath := coreworkshop.GatehousePath(home, req.WorkshopID, workshop.Name)
	gatehousePathExists := s.dirExists(gatehousePath)
	gatehouseConfigExists := s.fileExists(filepath.Join(gatehousePath, ".orc", "config.json"))

	// 5. Get workbenches and check each one's state
	workbenches, _ := s.workbenchRepo.List(ctx, req.WorkshopID)
	var wbInputs []coreinfra.WorkbenchPlanInput
	for _, wb := range workbenches {
		repoName := ""
		if wb.RepoID != "" {
			if repo, err := s.repoRepo.GetByID(ctx, wb.RepoID); err == nil {
				repoName = repo.Name
			}
		}
		wbInputs = append(wbInputs, coreinfra.WorkbenchPlanInput{
			ID:             wb.ID,
			Name:           wb.Name,
			WorktreePath:   wb.WorktreePath,
			RepoName:       repoName,
			HomeBranch:     wb.HomeBranch,
			WorktreeExists: s.dirExists(wb.WorktreePath),
			ConfigExists:   s.fileExists(filepath.Join(wb.WorktreePath, ".orc", "config.json")),
		})
	}

	// 6. Generate plan using pure function
	input := coreinfra.PlanInput{
		WorkshopID:            req.WorkshopID,
		WorkshopName:          workshop.Name,
		FactoryID:             factory.ID,
		FactoryName:           factory.Name,
		GatehouseID:           gatehouseID,
		GatehousePath:         gatehousePath,
		GatehousePathExists:   gatehousePathExists,
		GatehouseConfigExists: gatehouseConfigExists,
		Workbenches:           wbInputs,
	}
	corePlan := coreinfra.GeneratePlan(input)

	// 7. Convert core plan to primary plan
	return s.corePlanToPrimary(&corePlan), nil
}

func (s *InfraServiceImpl) corePlanToPrimary(core *coreinfra.Plan) *primary.InfraPlan {
	plan := &primary.InfraPlan{
		WorkshopID:   core.WorkshopID,
		WorkshopName: core.WorkshopName,
		FactoryID:    core.FactoryID,
		FactoryName:  core.FactoryName,
	}

	// Map gatehouse
	if core.Gatehouse != nil {
		plan.Gatehouse = &primary.InfraGatehouseOp{
			ID:           core.Gatehouse.ID,
			Path:         core.Gatehouse.Path,
			Status:       boolToOpStatus(core.Gatehouse.Exists),
			ConfigStatus: boolToOpStatus(core.Gatehouse.ConfigExists),
		}
	}

	// Map workbenches
	for _, wb := range core.Workbenches {
		status := boolToOpStatus(wb.Exists)
		// If workbench is in DB but not on filesystem, it's MISSING
		if !wb.Exists && wb.ID != "" {
			status = primary.OpMissing
		}
		plan.Workbenches = append(plan.Workbenches, primary.InfraWorkbenchOp{
			ID:           wb.ID,
			Name:         wb.Name,
			Path:         wb.Path,
			Status:       status,
			ConfigStatus: boolToOpStatus(wb.ConfigExists),
			RepoName:     wb.RepoName,
			Branch:       wb.Branch,
		})
	}

	return plan
}

func (s *InfraServiceImpl) dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func (s *InfraServiceImpl) fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// ApplyInfra executes the infrastructure plan, creating directories, worktrees, and configs.
func (s *InfraServiceImpl) ApplyInfra(ctx context.Context, plan *primary.InfraPlan) (*primary.InfraApplyResponse, error) {
	response := &primary.InfraApplyResponse{
		WorkshopID:   plan.WorkshopID,
		WorkshopName: plan.WorkshopName,
	}

	// Check if nothing to do
	nothingToDo := true
	if plan.Gatehouse != nil && (plan.Gatehouse.Status == primary.OpCreate || plan.Gatehouse.ConfigStatus == primary.OpCreate) {
		nothingToDo = false
	}
	for _, wb := range plan.Workbenches {
		if wb.Status == primary.OpCreate || wb.Status == primary.OpMissing || wb.ConfigStatus == primary.OpCreate {
			nothingToDo = false
			break
		}
	}
	if nothingToDo {
		response.NothingToDo = true
		return response, nil
	}

	// 1. Create gatehouse directory and config if needed
	if plan.Gatehouse != nil {
		if plan.Gatehouse.Status == primary.OpCreate || plan.Gatehouse.ConfigStatus == primary.OpCreate {
			gatehouseID := plan.Gatehouse.ID
			// If no gatehouse record exists, create one
			if gatehouseID == "" {
				gatehouse, err := s.ensureGatehouseExists(ctx, plan.WorkshopID)
				if err != nil {
					return nil, fmt.Errorf("failed to ensure gatehouse: %w", err)
				}
				gatehouseID = gatehouse.ID
			}

			if err := s.createGatehouseDir(plan.Gatehouse.Path, gatehouseID); err != nil {
				return nil, fmt.Errorf("failed to create gatehouse directory: %w", err)
			}
			response.GatehouseCreated = true
			response.ConfigsCreated++
		}
	}

	// 2. Create workbench worktrees and configs
	workbenches, _ := s.workbenchRepo.List(ctx, plan.WorkshopID)
	for _, wb := range workbenches {
		var planWb *primary.InfraWorkbenchOp
		for i := range plan.Workbenches {
			if plan.Workbenches[i].ID == wb.ID {
				planWb = &plan.Workbenches[i]
				break
			}
		}
		if planWb == nil {
			continue
		}

		// Create worktree if needed
		if planWb.Status == primary.OpCreate || planWb.Status == primary.OpMissing {
			if err := s.ensureWorktreeExists(ctx, wb); err != nil {
				return nil, fmt.Errorf("failed to create worktree for %s: %w", wb.Name, err)
			}
			response.WorkbenchesCreated++
		} else if planWb.ConfigStatus == primary.OpCreate {
			// Only config needs to be created
			if err := s.ensureConfigExists(ctx, wb); err != nil {
				return nil, fmt.Errorf("failed to create config for %s: %w", wb.Name, err)
			}
		}

		if planWb.Status == primary.OpCreate || planWb.Status == primary.OpMissing || planWb.ConfigStatus == primary.OpCreate {
			response.ConfigsCreated++
		}
	}

	return response, nil
}

// ensureGatehouseExists returns the gatehouse for a workshop, creating it if needed.
func (s *InfraServiceImpl) ensureGatehouseExists(ctx context.Context, workshopID string) (*secondary.GatehouseRecord, error) {
	existing, err := s.gatehouseRepo.GetByWorkshop(ctx, workshopID)
	if err == nil {
		return existing, nil
	}

	id, err := s.gatehouseRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate gatehouse ID: %w", err)
	}

	gatehouse := &secondary.GatehouseRecord{
		ID:         id,
		WorkshopID: workshopID,
		Status:     "active",
	}
	if err := s.gatehouseRepo.Create(ctx, gatehouse); err != nil {
		return nil, fmt.Errorf("failed to create gatehouse: %w", err)
	}

	return gatehouse, nil
}

// createGatehouseDir creates the gatehouse directory with config.
func (s *InfraServiceImpl) createGatehouseDir(path, gatehouseID string) error {
	var effs []effects.Effect

	// Create directory
	effs = append(effs, effects.FileEffect{
		Operation: "mkdir",
		Path:      path,
		Mode:      0755,
	})

	// Create .orc subdir
	orcDir := filepath.Join(path, ".orc")
	effs = append(effs, effects.FileEffect{
		Operation: "mkdir",
		Path:      orcDir,
		Mode:      0755,
	})

	// Create config
	cfg := &config.Config{
		Version: "1.0",
		PlaceID: gatehouseID,
	}
	configJSON, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	configPath := filepath.Join(orcDir, "config.json")
	effs = append(effs, effects.FileEffect{
		Operation: "write",
		Path:      configPath,
		Content:   configJSON,
		Mode:      0644,
	})

	return s.executor.Execute(context.Background(), effs)
}

// ensureWorktreeExists creates a worktree and IMP config if they don't exist.
func (s *InfraServiceImpl) ensureWorktreeExists(ctx context.Context, wb *secondary.WorkbenchRecord) error {
	exists, err := s.workspaceAdapter.WorktreeExists(ctx, wb.WorktreePath)
	if err != nil {
		return err
	}

	var effs []effects.Effect

	if !exists {
		if wb.RepoID == "" {
			effs = append(effs, effects.FileEffect{
				Operation: "mkdir",
				Path:      wb.WorktreePath,
				Mode:      0755,
			})
		} else {
			repo, err := s.repoRepo.GetByID(ctx, wb.RepoID)
			if err != nil {
				return fmt.Errorf("repo %s not found: %w", wb.RepoID, err)
			}
			effs = append(effs, effects.GitEffect{
				Operation: "worktree_add",
				RepoPath:  repo.LocalPath,
				Args:      []string{wb.HomeBranch, wb.WorktreePath},
			})
		}
	}

	// Create config
	orcDir := filepath.Join(wb.WorktreePath, ".orc")
	configPath := filepath.Join(orcDir, "config.json")
	cfg := &config.Config{
		Version: "1.0",
		PlaceID: wb.ID,
	}
	configJSON, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	effs = append(effs,
		effects.FileEffect{Operation: "mkdir", Path: orcDir, Mode: 0755},
		effects.FileEffect{Operation: "write", Path: configPath, Content: configJSON, Mode: 0644},
	)

	if len(effs) > 0 {
		return s.executor.Execute(ctx, effs)
	}
	return nil
}

// ensureConfigExists creates only the config for an existing worktree.
func (s *InfraServiceImpl) ensureConfigExists(ctx context.Context, wb *secondary.WorkbenchRecord) error {
	orcDir := filepath.Join(wb.WorktreePath, ".orc")
	configPath := filepath.Join(orcDir, "config.json")
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

// Ensure InfraServiceImpl implements the interface
var _ primary.InfraService = (*InfraServiceImpl)(nil)
