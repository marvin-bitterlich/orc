// Package infra contains pure business logic for infrastructure planning.
package infra

// PlanInput contains pre-fetched data for infrastructure plan generation.
// All values must be gathered by the caller - no I/O in the planner.
type PlanInput struct {
	WorkshopID   string
	WorkshopName string
	FactoryID    string
	FactoryName  string

	// Workshop coordination directory state
	WorkshopDirID           string
	WorkshopDirPath         string
	WorkshopDirPathExists   bool
	WorkshopDirConfigExists bool

	// Workbench state
	Workbenches []WorkbenchPlanInput

	// Orphan state (exist on disk but not in DB)
	OrphanWorkbenches  []WorkbenchPlanInput
	OrphanWorkshopDirs []OrphanDirPlanInput

	// TMux state
	TMuxSessionExists     bool              // Session found by ORC_WORKSHOP_ID
	TMuxActualSessionName string            // Actual session name (may differ after renames)
	TMuxExistingWindows   []string          // Window names currently in session
	TMuxExpectedWindows   []TMuxWindowInput // Windows that should exist (from workbenches)
}

// OrphanDirPlanInput contains pre-fetched data for a single orphan directory.
type OrphanDirPlanInput struct {
	PlaceID string // From config.json
	Path    string
}

// WorkbenchPlanInput contains pre-fetched data for a single workbench.
type WorkbenchPlanInput struct {
	ID             string
	Name           string
	WorktreePath   string
	RepoName       string
	HomeBranch     string
	WorktreeExists bool
	ConfigExists   bool
}

// TMuxWindowInput contains pre-fetched data for an expected tmux window.
type TMuxWindowInput struct {
	Name          string          // Window name (usually workbench name)
	Path          string          // Working directory for the window
	Panes         []TMuxPaneInput // Pane state (if window exists)
	ActualAgent   string          // Current @orc_agent value (empty if not set)
	ExpectedAgent string          // Expected agent (e.g., "IMP-name@BENCH-xxx")
	WorkbenchID   string          // Workbench ID
}

// TMuxPaneInput contains pre-fetched data for a tmux pane.
type TMuxPaneInput struct {
	Index           int    // Pane index (1-based)
	StartPath       string // Initial directory (pane_start_path)
	StartCommand    string // Initial command (pane_start_command, only set via respawn-pane)
	ExpectedPath    string // Expected directory for verification
	ExpectedCommand string // Expected command for verification (empty means no check)
}

// Plan describes infrastructure state for a workshop.
type Plan struct {
	WorkshopID   string
	WorkshopName string
	FactoryID    string
	FactoryName  string

	WorkshopDir *WorkshopDirOp
	Workbenches []WorkbenchOp

	// Orphans (exist on disk but not in DB)
	OrphanWorkbenches  []WorkbenchOp
	OrphanWorkshopDirs []WorkshopDirOp

	// TMux state
	TMuxSession *TMuxSessionOp
}

// WorkshopDirOp describes workshop coordination directory infrastructure state.
type WorkshopDirOp struct {
	ID           string
	Path         string
	Exists       bool
	ConfigExists bool
}

// WorkbenchOp describes workbench infrastructure state.
type WorkbenchOp struct {
	ID           string
	Name         string
	Path         string
	Exists       bool
	ConfigExists bool
	RepoName     string
	Branch       string
}

// TMuxSessionOp describes tmux session infrastructure state.
type TMuxSessionOp struct {
	SessionName   string
	Exists        bool
	Windows       []TMuxWindowOp
	OrphanWindows []TMuxWindowOp // Windows that exist but shouldn't (workbench deleted/archived)
}

// TMuxWindowOp describes tmux window infrastructure state.
type TMuxWindowOp struct {
	Name          string
	Path          string
	Exists        bool
	Panes         []TMuxPaneOp // Pane verification results (only populated if window exists)
	AgentOK       bool         // @orc_agent matches expected
	ActualAgent   string       // Current @orc_agent value
	ExpectedAgent string       // Expected agent (e.g., "IMP-name@BENCH-xxx")
	WorkbenchID   string       // Workbench ID
}

// TMuxPaneOp describes tmux pane verification state.
type TMuxPaneOp struct {
	Index           int    // Pane index (1-based)
	PathOK          bool   // StartPath matches expected
	CommandOK       bool   // StartCommand matches expected (true if no expected command)
	ActualPath      string // Actual pane_start_path
	ActualCommand   string // Actual pane_start_command
	ExpectedPath    string // Expected path
	ExpectedCommand string // Expected command (empty if shell)
}

// GeneratePlan creates an infrastructure plan.
// This is a pure function - all input data must be pre-fetched.
func GeneratePlan(input PlanInput) Plan {
	plan := Plan{
		WorkshopID:   input.WorkshopID,
		WorkshopName: input.WorkshopName,
		FactoryID:    input.FactoryID,
		FactoryName:  input.FactoryName,
	}

	// Workshop coordination directory
	plan.WorkshopDir = &WorkshopDirOp{
		ID:           input.WorkshopDirID,
		Path:         input.WorkshopDirPath,
		Exists:       input.WorkshopDirPathExists,
		ConfigExists: input.WorkshopDirConfigExists,
	}

	// Workbenches
	for _, wb := range input.Workbenches {
		plan.Workbenches = append(plan.Workbenches, WorkbenchOp{
			ID:           wb.ID,
			Name:         wb.Name,
			Path:         wb.WorktreePath,
			Exists:       wb.WorktreeExists,
			ConfigExists: wb.ConfigExists,
			RepoName:     wb.RepoName,
			Branch:       wb.HomeBranch,
		})
	}

	// Orphan workbenches (exist on disk but not in DB)
	for _, wb := range input.OrphanWorkbenches {
		plan.OrphanWorkbenches = append(plan.OrphanWorkbenches, WorkbenchOp{
			ID:           wb.ID,
			Name:         wb.Name,
			Path:         wb.WorktreePath,
			Exists:       true, // By definition, orphans exist on disk
			ConfigExists: true,
		})
	}

	// Orphan workshop directories
	for _, od := range input.OrphanWorkshopDirs {
		plan.OrphanWorkshopDirs = append(plan.OrphanWorkshopDirs, WorkshopDirOp{
			ID:           od.PlaceID,
			Path:         od.Path,
			Exists:       true,
			ConfigExists: true,
		})
	}

	// TMux session state
	plan.TMuxSession = buildTMuxSessionOp(input)

	return plan
}

// buildTMuxSessionOp creates the TMux session operation plan.
func buildTMuxSessionOp(input PlanInput) *TMuxSessionOp {
	sessionOp := &TMuxSessionOp{
		SessionName: input.TMuxActualSessionName,
		Exists:      input.TMuxSessionExists,
	}

	// Build set of existing windows for O(1) lookup
	existingSet := make(map[string]bool)
	for _, w := range input.TMuxExistingWindows {
		existingSet[w] = true
	}

	// Build set of expected window names for orphan detection
	expectedSet := make(map[string]bool)
	for _, w := range input.TMuxExpectedWindows {
		expectedSet[w.Name] = true
	}

	// Check each expected window
	for _, expected := range input.TMuxExpectedWindows {
		windowOp := TMuxWindowOp{
			Name:          expected.Name,
			Path:          expected.Path,
			Exists:        existingSet[expected.Name],
			ActualAgent:   expected.ActualAgent,
			ExpectedAgent: expected.ExpectedAgent,
			AgentOK:       expected.ActualAgent == expected.ExpectedAgent,
			WorkbenchID:   expected.WorkbenchID,
		}

		// Add pane verification if window exists and has pane data
		if windowOp.Exists && len(expected.Panes) > 0 {
			for _, pane := range expected.Panes {
				paneOp := TMuxPaneOp{
					Index:           pane.Index,
					ActualPath:      pane.StartPath,
					ActualCommand:   pane.StartCommand,
					ExpectedPath:    pane.ExpectedPath,
					ExpectedCommand: pane.ExpectedCommand,
				}
				// Check path match
				paneOp.PathOK = pane.StartPath == pane.ExpectedPath
				// Check command match (OK if no expected command)
				if pane.ExpectedCommand == "" {
					paneOp.CommandOK = true
				} else {
					paneOp.CommandOK = pane.StartCommand == pane.ExpectedCommand
				}
				windowOp.Panes = append(windowOp.Panes, paneOp)
			}
		}

		sessionOp.Windows = append(sessionOp.Windows, windowOp)
	}

	// Find orphan windows (exist but not expected)
	for _, windowName := range input.TMuxExistingWindows {
		if !expectedSet[windowName] {
			sessionOp.OrphanWindows = append(sessionOp.OrphanWindows, TMuxWindowOp{
				Name:   windowName,
				Exists: true,
			})
		}
	}

	return sessionOp
}
