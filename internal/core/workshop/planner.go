// Package workshop contains pure business logic for workshop operations.
package workshop

import "path/filepath"

// OpenPlanInput contains pre-fetched data for plan generation.
// All values must be gathered by the caller - no I/O in the planner.
type OpenPlanInput struct {
	WorkshopID            string
	WorkshopName          string
	FactoryID             string
	FactoryName           string
	SessionExists         bool
	GatehouseDir          string
	GatehouseDirExists    bool
	GatehouseConfigExists bool
	Workbenches           []WorkbenchPlanInput
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
	Status         string // DB status: active, archived, etc.
}

// OpenWorkshopPlan describes what will be created when opening a workshop.
type OpenWorkshopPlan struct {
	WorkshopID   string
	WorkshopName string
	FactoryID    string
	FactoryName  string
	SessionName  string
	Workbenches  []WorkbenchDBState // For DB state display
	GatehouseOp  *GatehouseOp
	WorkbenchOps []WorkbenchOp
	TMuxOp       *TMuxOp
	NothingToDo  bool
}

// WorkbenchDBState describes a workbench as stored in the database.
type WorkbenchDBState struct {
	ID     string
	Name   string
	Path   string
	Status string
}

// GatehouseOp describes the gatehouse directory operation.
type GatehouseOp struct {
	Path         string
	Exists       bool
	ConfigExists bool
}

// WorkbenchOp describes a workbench worktree operation.
type WorkbenchOp struct {
	ID           string
	Name         string
	Path         string
	Exists       bool
	RepoName     string
	Branch       string
	ConfigExists bool
}

// TMuxOp describes the tmux session operation.
type TMuxOp struct {
	SessionName string
	Windows     []TMuxWindowOp
}

// TMuxWindowOp describes a tmux window operation.
type TMuxWindowOp struct {
	Index int
	Name  string
	Path  string
}

// GenerateOpenPlan creates a plan for opening workshop infrastructure.
// This is a pure function - all input data must be pre-fetched.
// The plan includes ALL items (existing and new) so the display can show both.
func GenerateOpenPlan(input OpenPlanInput) OpenWorkshopPlan {
	plan := OpenWorkshopPlan{
		WorkshopID:   input.WorkshopID,
		WorkshopName: input.WorkshopName,
		FactoryID:    input.FactoryID,
		FactoryName:  input.FactoryName,
		SessionName:  input.WorkshopID,
	}

	// DB State - workbenches from database
	for _, wb := range input.Workbenches {
		plan.Workbenches = append(plan.Workbenches, WorkbenchDBState{
			ID:     wb.ID,
			Name:   wb.Name,
			Path:   wb.WorktreePath,
			Status: wb.Status,
		})
	}

	// Gatehouse - always include so we can display existing vs new
	plan.GatehouseOp = &GatehouseOp{
		Path:         input.GatehouseDir,
		Exists:       input.GatehouseDirExists,
		ConfigExists: input.GatehouseConfigExists,
	}

	// Workbenches - always include all
	for _, wb := range input.Workbenches {
		plan.WorkbenchOps = append(plan.WorkbenchOps, WorkbenchOp{
			ID:           wb.ID,
			Name:         wb.Name,
			Path:         wb.WorktreePath,
			Exists:       wb.WorktreeExists,
			RepoName:     wb.RepoName,
			Branch:       wb.HomeBranch,
			ConfigExists: wb.ConfigExists,
		})
	}

	// TMux - include if session doesn't exist
	if !input.SessionExists {
		windows := []TMuxWindowOp{
			{Index: 0, Name: "orc", Path: input.GatehouseDir},
		}
		for i, wb := range input.Workbenches {
			windows = append(windows, TMuxWindowOp{
				Index: i + 1,
				Name:  wb.Name,
				Path:  wb.WorktreePath,
			})
		}
		plan.TMuxOp = &TMuxOp{
			SessionName: input.WorkshopID,
			Windows:     windows,
		}
	}

	// Check if nothing to do - all infrastructure exists
	gatehouseReady := input.GatehouseDirExists && input.GatehouseConfigExists
	workbenchesReady := true
	for _, wb := range input.Workbenches {
		if !wb.WorktreeExists || !wb.ConfigExists {
			workbenchesReady = false
			break
		}
	}
	sessionReady := input.SessionExists

	plan.NothingToDo = gatehouseReady && workbenchesReady && sessionReady

	return plan
}

// slugify converts a name to a URL-friendly slug.
func Slugify(name string) string {
	var result []byte
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			result = append(result, byte(r))
		case r >= 'A' && r <= 'Z':
			result = append(result, byte(r+32)) // lowercase
		case r >= '0' && r <= '9':
			result = append(result, byte(r))
		case r == ' ' || r == '-' || r == '_':
			result = append(result, '-')
		}
	}
	return string(result)
}

// GatehousePath returns the path for a workshop's gatehouse directory.
func GatehousePath(homeDir, workshopID, workshopName string) string {
	slug := Slugify(workshopName)
	dirName := workshopID + "-" + slug
	return filepath.Join(homeDir, ".orc", "ws", dirName)
}
