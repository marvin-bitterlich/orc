package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/example/orc/internal/config"
	"github.com/example/orc/internal/core/effects"
	coreinfra "github.com/example/orc/internal/core/infra"
	coreworkbench "github.com/example/orc/internal/core/workbench"
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
	tmuxAdapter      secondary.TMuxAdapter
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
	tmuxAdapter secondary.TMuxAdapter,
	executor EffectExecutor,
) *InfraServiceImpl {
	return &InfraServiceImpl{
		factoryRepo:      factoryRepo,
		workshopRepo:     workshopRepo,
		workbenchRepo:    workbenchRepo,
		repoRepo:         repoRepo,
		gatehouseRepo:    gatehouseRepo,
		workspaceAdapter: workspaceAdapter,
		tmuxAdapter:      tmuxAdapter,
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
	workshopArchived := workshop.Status == "archived"

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
	allWorkbenches, _ := s.workbenchRepo.List(ctx, req.WorkshopID)
	// Filter to active workbenches only (archived workbenches are excluded from planning)
	var workbenches []*secondary.WorkbenchRecord
	for _, wb := range allWorkbenches {
		if wb.Status != "archived" {
			workbenches = append(workbenches, wb)
		}
	}
	var wbInputs []coreinfra.WorkbenchPlanInput
	for _, wb := range workbenches {
		repoName := ""
		if wb.RepoID != "" {
			if repo, err := s.repoRepo.GetByID(ctx, wb.RepoID); err == nil {
				repoName = repo.Name
			}
		}
		wbPath := coreworkbench.ComputePath(wb.Name)
		wbInputs = append(wbInputs, coreinfra.WorkbenchPlanInput{
			ID:             wb.ID,
			Name:           wb.Name,
			WorktreePath:   wbPath,
			RepoName:       repoName,
			HomeBranch:     wb.HomeBranch,
			WorktreeExists: s.dirExists(wbPath),
			ConfigExists:   s.fileExists(filepath.Join(wbPath, ".orc", "config.json")),
		})
	}

	// 6. Scan for orphaned configs on disk (true orphans only - no DB record)
	// Note: Archived workbenches are NOT added to orphan list - they have DB records
	// and should not have their directories deleted by infra apply.
	orphanWbs, orphanGhs := s.scanForOrphans(ctx, workbenches, gatehouse)

	// 7. Fetch TMux session state
	tmuxSessionName := ""
	tmuxSessionExists := false
	var tmuxExistingWindows []string
	var tmuxExpectedWindows []coreinfra.TMuxWindowInput

	if s.tmuxAdapter != nil {
		tmuxSessionName = s.tmuxAdapter.FindSessionByWorkshopID(ctx, req.WorkshopID)
		tmuxSessionExists = tmuxSessionName != ""
		if tmuxSessionExists {
			tmuxExistingWindows, _ = s.tmuxAdapter.ListWindows(ctx, tmuxSessionName)
		}
		// Build expected windows - first goblin window, then workbenches
		// (skip if workshop is archived - entire session should be deleted)
		if !workshopArchived {
			// Goblin window name: goblin-NNN (derived from GATE-NNN)
			goblinWindowName := "goblin-" + strings.TrimPrefix(gatehouseID, "GATE-")
			goblinExpectedAgent := fmt.Sprintf("GOBLIN@%s", gatehouseID)
			goblinActualAgent := ""
			if tmuxSessionExists {
				goblinActualAgent = s.tmuxAdapter.GetWindowOption(ctx, tmuxSessionName+":"+goblinWindowName, "@orc_agent")
			}
			tmuxExpectedWindows = append(tmuxExpectedWindows, coreinfra.TMuxWindowInput{
				Name:          goblinWindowName,
				Path:          gatehousePath,
				ExpectedAgent: goblinExpectedAgent,
				ActualAgent:   goblinActualAgent,
			})
			// Workbench windows (IMPs)
			for _, wb := range workbenches {
				wbPath := coreworkbench.ComputePath(wb.Name)
				expectedAgent := fmt.Sprintf("IMP-%s@%s", wb.Name, wb.ID)
				actualAgent := ""
				if tmuxSessionExists && s.tmuxAdapter.WindowExists(ctx, tmuxSessionName, wb.Name) {
					actualAgent = s.tmuxAdapter.GetWindowOption(ctx, tmuxSessionName+":"+wb.Name, "@orc_agent")
				}
				windowInput := coreinfra.TMuxWindowInput{
					Name:          wb.Name,
					Path:          wbPath,
					ExpectedAgent: expectedAgent,
					ActualAgent:   actualAgent,
				}
				// Fetch pane data if window exists
				if tmuxSessionExists && s.tmuxAdapter.WindowExists(ctx, tmuxSessionName, wb.Name) {
					windowInput.Panes = s.fetchPaneData(ctx, tmuxSessionName, wb.Name, wbPath)
				}
				tmuxExpectedWindows = append(tmuxExpectedWindows, windowInput)
			}
		}
	}

	// 8. Generate plan using pure function (all I/O already done above)
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
		OrphanWorkbenches:     orphanWbs,
		OrphanGatehouses:      orphanGhs,
		// TMux state
		TMuxSessionExists:     tmuxSessionExists,
		TMuxActualSessionName: tmuxSessionName,
		TMuxExistingWindows:   tmuxExistingWindows,
		TMuxExpectedWindows:   tmuxExpectedWindows,
	}
	corePlan := coreinfra.GeneratePlan(input)

	// 9. Convert core plan to primary plan
	result := s.corePlanToPrimary(&corePlan)
	result.Force = req.Force
	result.NoDelete = req.NoDelete
	return result, nil
}

// scanForOrphans scans filesystem for config.json files with place_ids not in DB.
// Scans ~/wb/*/.orc/config.json for workbenches and ~/.orc/ws/WORK-*/.orc/config.json for gatehouses.
func (s *InfraServiceImpl) scanForOrphans(ctx context.Context, knownWorkbenches []*secondary.WorkbenchRecord, knownGatehouse *secondary.GatehouseRecord) ([]coreinfra.WorkbenchPlanInput, []coreinfra.GatehousePlanInput) {
	var orphanWbs []coreinfra.WorkbenchPlanInput
	var orphanGhs []coreinfra.GatehousePlanInput

	home, err := os.UserHomeDir()
	if err != nil {
		return orphanWbs, orphanGhs
	}

	// Build set of known IDs for quick lookup
	knownWbIDs := make(map[string]bool)
	for _, wb := range knownWorkbenches {
		knownWbIDs[wb.ID] = true
	}
	knownGhID := ""
	if knownGatehouse != nil {
		knownGhID = knownGatehouse.ID
	}

	// Scan ~/wb/*/.orc/config.json for workbenches
	wbPattern := filepath.Join(home, "wb", "*", ".orc", "config.json")
	wbConfigs, _ := filepath.Glob(wbPattern)
	for _, configPath := range wbConfigs {
		placeID, err := s.readPlaceID(configPath)
		if err != nil || placeID == "" {
			continue
		}
		// Only consider BENCH-* place IDs
		if !strings.HasPrefix(placeID, "BENCH-") {
			continue
		}
		// Check if this place_id exists in DB
		if knownWbIDs[placeID] {
			continue // Known workbench, not an orphan
		}
		// Verify it's truly not in DB (cross-check with repo)
		_, err = s.workbenchRepo.GetByID(ctx, placeID)
		if err == nil {
			continue // Found in DB, not an orphan
		}
		// This is an orphan
		wbPath := filepath.Dir(filepath.Dir(configPath)) // Go up from .orc/config.json
		orphanWbs = append(orphanWbs, coreinfra.WorkbenchPlanInput{
			ID:           placeID,
			Name:         filepath.Base(wbPath),
			WorktreePath: wbPath,
		})
	}

	// Scan ~/.orc/ws/WORK-*/.orc/config.json for gatehouses
	ghPattern := filepath.Join(home, ".orc", "ws", "WORK-*", ".orc", "config.json")
	ghConfigs, _ := filepath.Glob(ghPattern)
	for _, configPath := range ghConfigs {
		placeID, err := s.readPlaceID(configPath)
		if err != nil || placeID == "" {
			continue
		}
		// Only consider GATE-* place IDs
		if !strings.HasPrefix(placeID, "GATE-") {
			continue
		}
		// Check if this is the known gatehouse
		if placeID == knownGhID {
			continue // Known gatehouse, not an orphan
		}
		// Verify it's truly not in DB
		_, err = s.gatehouseRepo.GetByID(ctx, placeID)
		if err == nil {
			continue // Found in DB, not an orphan
		}
		// This is an orphan
		ghPath := filepath.Dir(filepath.Dir(configPath))
		orphanGhs = append(orphanGhs, coreinfra.GatehousePlanInput{
			PlaceID: placeID,
			Path:    ghPath,
		})
	}

	return orphanWbs, orphanGhs
}

// readPlaceID reads the place_id from a config.json file.
func (s *InfraServiceImpl) readPlaceID(configPath string) (string, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}
	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", err
	}
	return cfg.PlaceID, nil
}

// isWorktreeDirty checks if a git worktree has uncommitted changes.
// Returns (dirty, modified_count, untracked_count, error).
func (s *InfraServiceImpl) isWorktreeDirty(path string) (bool, int, int, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		// Not a git repo or git error - treat as not dirty
		return false, 0, 0, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return false, 0, 0, nil
	}

	modified := 0
	untracked := 0
	for _, line := range lines {
		if len(line) < 2 {
			continue
		}
		// Status codes: M = modified, A = added, D = deleted, ?? = untracked
		if strings.HasPrefix(line, "??") {
			untracked++
		} else {
			modified++
		}
	}

	return modified > 0 || untracked > 0, modified, untracked, nil
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

	// Map orphan workbenches (exist on disk but not in DB)
	for _, wb := range core.OrphanWorkbenches {
		plan.OrphanWorkbenches = append(plan.OrphanWorkbenches, primary.InfraWorkbenchOp{
			ID:           wb.ID,
			Name:         wb.Name,
			Path:         wb.Path,
			Status:       primary.OpDelete,
			ConfigStatus: primary.OpDelete,
		})
	}

	// Map orphan gatehouses
	for _, gh := range core.OrphanGatehouses {
		plan.OrphanGatehouses = append(plan.OrphanGatehouses, primary.InfraGatehouseOp{
			ID:           gh.ID,
			Path:         gh.Path,
			Status:       primary.OpDelete,
			ConfigStatus: primary.OpDelete,
		})
	}

	// Map TMux session
	if core.TMuxSession != nil {
		plan.TMuxSession = &primary.InfraTMuxSessionOp{
			SessionName: core.TMuxSession.SessionName,
			Status:      boolToOpStatus(core.TMuxSession.Exists),
		}
		// Map expected windows
		for _, w := range core.TMuxSession.Windows {
			windowOp := primary.InfraTMuxWindowOp{
				Name:          w.Name,
				Path:          w.Path,
				Status:        boolToOpStatus(w.Exists),
				AgentOK:       w.AgentOK,
				ActualAgent:   w.ActualAgent,
				ExpectedAgent: w.ExpectedAgent,
			}
			// Map pane verification data
			for _, p := range w.Panes {
				windowOp.Panes = append(windowOp.Panes, primary.InfraTMuxPaneOp{
					Index:           p.Index,
					PathOK:          p.PathOK,
					CommandOK:       p.CommandOK,
					ActualPath:      p.ActualPath,
					ActualCommand:   p.ActualCommand,
					ExpectedPath:    p.ExpectedPath,
					ExpectedCommand: p.ExpectedCommand,
				})
			}
			plan.TMuxSession.Windows = append(plan.TMuxSession.Windows, windowOp)
		}
		// Map orphan windows (exist but shouldn't)
		for _, w := range core.TMuxSession.OrphanWindows {
			plan.TMuxSession.OrphanWindows = append(plan.TMuxSession.OrphanWindows, primary.InfraTMuxWindowOp{
				Name:   w.Name,
				Path:   w.Path,
				Status: primary.OpDelete,
			})
		}
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

// fetchPaneData retrieves pane verification data for an existing window.
// Expected layout: Pane 1 (vim), Pane 2 (orc connect), Pane 3 (shell).
func (s *InfraServiceImpl) fetchPaneData(ctx context.Context, sessionName, windowName, expectedPath string) []coreinfra.TMuxPaneInput {
	var panes []coreinfra.TMuxPaneInput

	// Pane 1: vim (expected command = "vim")
	panes = append(panes, coreinfra.TMuxPaneInput{
		Index:           1,
		StartPath:       s.tmuxAdapter.GetPaneStartPath(ctx, sessionName, windowName, 1),
		StartCommand:    s.tmuxAdapter.GetPaneStartCommand(ctx, sessionName, windowName, 1),
		ExpectedPath:    expectedPath,
		ExpectedCommand: "vim",
	})

	// Pane 2: IMP (expected command = "orc connect")
	panes = append(panes, coreinfra.TMuxPaneInput{
		Index:           2,
		StartPath:       s.tmuxAdapter.GetPaneStartPath(ctx, sessionName, windowName, 2),
		StartCommand:    s.tmuxAdapter.GetPaneStartCommand(ctx, sessionName, windowName, 2),
		ExpectedPath:    expectedPath,
		ExpectedCommand: "orc connect",
	})

	// Pane 3: shell (no expected command, just verify path)
	panes = append(panes, coreinfra.TMuxPaneInput{
		Index:           3,
		StartPath:       s.tmuxAdapter.GetPaneStartPath(ctx, sessionName, windowName, 3),
		StartCommand:    s.tmuxAdapter.GetPaneStartCommand(ctx, sessionName, windowName, 3),
		ExpectedPath:    expectedPath,
		ExpectedCommand: "", // Shell - no command expected
	})

	return panes
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
	// Check for orphan deletions
	if len(plan.OrphanWorkbenches) > 0 || len(plan.OrphanGatehouses) > 0 {
		nothingToDo = false
	}
	// Check for TMux operations
	if plan.TMuxSession != nil {
		if plan.TMuxSession.Status == primary.OpCreate {
			nothingToDo = false
		}
		for _, w := range plan.TMuxSession.Windows {
			if w.Status == primary.OpCreate {
				nothingToDo = false
				break
			}
		}
		if len(plan.TMuxSession.OrphanWindows) > 0 {
			nothingToDo = false
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

	// 3. Handle TMux session and windows
	// Note: Filesystem deletion of orphan workbenches/gatehouses is intentionally
	// removed from infra apply. Use `orc infra cleanup` for orphan deletion.
	if s.tmuxAdapter != nil && plan.TMuxSession != nil {
		sessionName := plan.TMuxSession.SessionName
		if sessionName == "" {
			// Generate session name from workshop name if not set
			sessionName = plan.WorkshopName
		}

		// Create session if needed
		if plan.TMuxSession.Status == primary.OpCreate {
			// Use gatehouse path as working directory
			workingDir := ""
			if plan.Gatehouse != nil {
				workingDir = plan.Gatehouse.Path
			}
			if err := s.tmuxAdapter.CreateSession(ctx, sessionName, workingDir); err != nil {
				return nil, fmt.Errorf("failed to create tmux session: %w", err)
			}
			// Set ORC_WORKSHOP_ID environment variable
			if err := s.tmuxAdapter.SetEnvironment(ctx, sessionName, "ORC_WORKSHOP_ID", plan.WorkshopID); err != nil {
				return nil, fmt.Errorf("failed to set ORC_WORKSHOP_ID: %w", err)
			}
		}

		// Create windows for workbenches
		// When session is newly created, first rename __init__ window to the first planned window name
		firstWindow := true
		windowIndex := 1 // Start at 1 since window 0 is typically the default window
		for _, w := range plan.TMuxSession.Windows {
			if w.Status == primary.OpCreate {
				if firstWindow && plan.TMuxSession.Status == primary.OpCreate {
					// Rename the __init__ placeholder window instead of creating new
					initTarget := sessionName + ":__init__"
					if err := s.tmuxAdapter.RenameWindow(ctx, initTarget, w.Name); err != nil {
						return nil, fmt.Errorf("failed to rename init window to %s: %w", w.Name, err)
					}
					// Launch orc connect --role goblin in goblin window
					if strings.HasPrefix(w.Name, "goblin-") {
						if err := s.tmuxAdapter.SetupGoblinPane(ctx, sessionName, w.Name); err != nil {
							return nil, fmt.Errorf("failed to setup goblin pane: %w", err)
						}
					}
				} else {
					if err := s.tmuxAdapter.CreateWorkbenchWindowShell(ctx, sessionName, windowIndex, w.Name, w.Path); err != nil {
						return nil, fmt.Errorf("failed to create tmux window %s: %w", w.Name, err)
					}
				}
			}
			firstWindow = false
			// Set @orc_agent if not matching expected
			if !w.AgentOK && w.ExpectedAgent != "" {
				target := sessionName + ":" + w.Name
				_ = s.tmuxAdapter.SetWindowOption(ctx, target, "@orc_agent", w.ExpectedAgent)
			}
			windowIndex++
		}

		// Kill orphan windows (unless --no-delete)
		if !plan.NoDelete {
			for _, w := range plan.TMuxSession.OrphanWindows {
				if w.Status == primary.OpDelete {
					if err := s.tmuxAdapter.KillWindow(ctx, sessionName, w.Name); err != nil {
						// Log but don't fail - window may already be gone
						continue
					}
				}
			}
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
	wbPath := coreworkbench.ComputePath(wb.Name)
	exists, err := s.workspaceAdapter.WorktreeExists(ctx, wbPath)
	if err != nil {
		return err
	}

	var effs []effects.Effect

	if !exists {
		if wb.RepoID == "" {
			effs = append(effs, effects.FileEffect{
				Operation: "mkdir",
				Path:      wbPath,
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
				Args:      []string{wb.HomeBranch, wbPath},
			})
		}
	}

	// Create config
	orcDir := filepath.Join(wbPath, ".orc")
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
	wbPath := coreworkbench.ComputePath(wb.Name)
	orcDir := filepath.Join(wbPath, ".orc")
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

// CleanupWorkbench removes the worktree directory for a workbench.
func (s *InfraServiceImpl) CleanupWorkbench(ctx context.Context, req primary.CleanupWorkbenchRequest) error {
	// Get workbench to compute path and get workshop info
	wb, err := s.workbenchRepo.GetByID(ctx, req.WorkbenchID)
	if err != nil {
		return fmt.Errorf("workbench not found: %w", err)
	}

	wbPath := coreworkbench.ComputePath(wb.Name)

	// Check for dirty worktree if not forced
	if !req.Force {
		dirty, modified, untracked, err := s.isWorktreeDirty(wbPath)
		if err == nil && dirty {
			return fmt.Errorf("cannot delete %s: worktree has uncommitted changes (%d modified, %d untracked). Use --force to override", req.WorkbenchID, modified, untracked)
		}
	}

	// Remove the worktree directory
	if err := os.RemoveAll(wbPath); err != nil {
		return fmt.Errorf("failed to remove worktree %s: %w", wbPath, err)
	}

	// Kill tmux window if exists (best effort - don't fail if not found)
	if s.tmuxAdapter != nil {
		sessionName := s.tmuxAdapter.FindSessionByWorkshopID(ctx, wb.WorkshopID)
		if sessionName != "" {
			_ = s.tmuxAdapter.KillWindow(ctx, sessionName, wb.Name)
		}
	}

	return nil
}

// CleanupWorkshop removes all infrastructure for a workshop.
func (s *InfraServiceImpl) CleanupWorkshop(ctx context.Context, req primary.CleanupWorkshopRequest) error {
	// Get workshop info before deletion
	workshop, err := s.workshopRepo.GetByID(ctx, req.WorkshopID)
	if err != nil {
		return fmt.Errorf("workshop not found: %w", err)
	}

	// Get all workbenches for this workshop
	workbenches, _ := s.workbenchRepo.List(ctx, req.WorkshopID)

	// Cleanup each workbench
	for _, wb := range workbenches {
		if err := s.CleanupWorkbench(ctx, primary.CleanupWorkbenchRequest{
			WorkbenchID: wb.ID,
			Force:       req.Force,
		}); err != nil {
			return fmt.Errorf("failed to cleanup workbench %s: %w", wb.ID, err)
		}
	}

	// Remove gatehouse directory
	home, _ := os.UserHomeDir()
	gatehousePath := coreworkshop.GatehousePath(home, req.WorkshopID, workshop.Name)
	if err := os.RemoveAll(gatehousePath); err != nil {
		return fmt.Errorf("failed to remove gatehouse directory: %w", err)
	}

	// Kill tmux session if exists
	if s.tmuxAdapter != nil {
		sessionName := s.tmuxAdapter.FindSessionByWorkshopID(ctx, req.WorkshopID)
		if sessionName != "" {
			_ = s.tmuxAdapter.KillSession(ctx, sessionName)
		}
	}

	return nil
}

// CleanupOrphans scans for and removes orphaned infrastructure.
func (s *InfraServiceImpl) CleanupOrphans(ctx context.Context, req primary.CleanupOrphansRequest) (*primary.CleanupOrphansResponse, error) {
	response := &primary.CleanupOrphansResponse{}

	// Scan for orphans (passing empty known lists to find all orphans)
	orphanWbs, orphanGhs := s.scanForOrphans(ctx, nil, nil)

	// Delete orphan workbenches
	for _, wb := range orphanWbs {
		// Check for dirty worktree if not forced
		if !req.Force {
			dirty, modified, untracked, err := s.isWorktreeDirty(wb.WorktreePath)
			if err == nil && dirty {
				return nil, fmt.Errorf("cannot delete %s: worktree has uncommitted changes (%d modified, %d untracked). Use --force to override", wb.ID, modified, untracked)
			}
		}
		if err := os.RemoveAll(wb.WorktreePath); err != nil {
			return nil, fmt.Errorf("failed to delete orphan workbench %s: %w", wb.ID, err)
		}
		response.WorkbenchesDeleted++
	}

	// Delete orphan gatehouses
	for _, gh := range orphanGhs {
		if err := os.RemoveAll(gh.Path); err != nil {
			return nil, fmt.Errorf("failed to delete orphan gatehouse %s: %w", gh.PlaceID, err)
		}
		response.GatehousesDeleted++
	}

	return response, nil
}

// Ensure InfraServiceImpl implements the interface
var _ primary.InfraService = (*InfraServiceImpl)(nil)
