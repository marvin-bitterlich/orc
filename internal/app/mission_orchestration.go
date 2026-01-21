// Package app contains the application services that orchestrate business logic.
package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/example/orc/internal/config"
	"github.com/example/orc/internal/ports/primary"
)

// MissionOrchestrationService handles complex mission infrastructure operations.
// It implements the plan/apply pattern for idempotent infrastructure management.
type MissionOrchestrationService struct {
	missionSvc primary.MissionService
	groveSvc   primary.GroveService
}

// NewMissionOrchestrationService creates a new orchestration service.
func NewMissionOrchestrationService(missionSvc primary.MissionService, groveSvc primary.GroveService) *MissionOrchestrationService {
	return &MissionOrchestrationService{
		missionSvc: missionSvc,
		groveSvc:   groveSvc,
	}
}

// MissionState represents the loaded state of a mission and its groves.
type MissionState struct {
	Mission *primary.Mission
	Groves  []*primary.Grove
}

// InfrastructurePlan describes the changes needed to set up mission infrastructure.
type InfrastructurePlan struct {
	WorkspacePath string
	GrovesDir     string

	// Actions to perform
	CreateWorkspace bool
	CreateGrovesDir bool
	GroveActions    []GroveAction
	ConfigWrites    []ConfigWrite
	Cleanups        []CleanupAction
}

// GroveAction represents an action to take on a grove.
type GroveAction struct {
	GroveID      string
	GroveName    string
	CurrentPath  string
	DesiredPath  string
	Action       string // "exists", "create", "move", "missing"
	PathExists   bool
	UpdateDBPath bool
}

// ConfigWrite represents a config file to write.
type ConfigWrite struct {
	Path    string
	Type    string // "grove", "claude-settings"
	Grove   *primary.Grove
	Content string // For preview only
}

// CleanupAction represents a file to clean up.
type CleanupAction struct {
	Path   string
	Reason string
}

// InfrastructureApplyResult captures the result of applying infrastructure changes.
type InfrastructureApplyResult struct {
	WorkspaceCreated  bool
	GrovesDirCreated  bool
	GrovesProcessed   int
	ConfigsWritten    int
	CleanupsDone      int
	Errors            []string
	GrovesNeedingWork []GroveNeedingWork
}

// GroveNeedingWork describes a grove that needs additional work.
type GroveNeedingWork struct {
	GroveID     string
	GroveName   string
	DesiredPath string
	Message     string
}

// LoadMissionState loads the mission and its groves from the database.
func (s *MissionOrchestrationService) LoadMissionState(ctx context.Context, missionID string) (*MissionState, error) {
	mission, err := s.missionSvc.GetMission(ctx, missionID)
	if err != nil {
		return nil, fmt.Errorf("mission not found: %w", err)
	}

	groves, err := s.groveSvc.ListGroves(ctx, primary.GroveFilters{MissionID: missionID})
	if err != nil {
		return nil, fmt.Errorf("failed to load groves: %w", err)
	}

	return &MissionState{
		Mission: mission,
		Groves:  groves,
	}, nil
}

// AnalyzeInfrastructure generates a plan for setting up mission infrastructure.
func (s *MissionOrchestrationService) AnalyzeInfrastructure(state *MissionState, workspacePath string) *InfrastructurePlan {
	plan := &InfrastructurePlan{
		WorkspacePath: workspacePath,
		GrovesDir:     filepath.Join(workspacePath, "groves"),
	}

	// Check workspace directory
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		plan.CreateWorkspace = true
	}

	// Check groves directory
	if _, err := os.Stat(plan.GrovesDir); os.IsNotExist(err) {
		plan.CreateGrovesDir = true
	}

	// Analyze each grove
	for _, grove := range state.Groves {
		desiredPath := filepath.Join(plan.GrovesDir, grove.Name)
		currentPath := grove.Path

		action := s.analyzeGroveAction(grove, currentPath, desiredPath)
		plan.GroveActions = append(plan.GroveActions, action)

		// Plan config writes for groves that exist or will exist
		groveConfigPath := filepath.Join(desiredPath, ".orc", "config.json")
		if _, err := os.Stat(groveConfigPath); os.IsNotExist(err) {
			plan.ConfigWrites = append(plan.ConfigWrites, ConfigWrite{
				Path:    groveConfigPath,
				Type:    "grove",
				Grove:   grove,
				Content: s.generateGroveConfigPreview(grove),
			})
		}

		// Plan Claude settings write
		claudeSettingsPath := filepath.Join(desiredPath, ".claude", "settings.local.json")
		if _, err := os.Stat(claudeSettingsPath); os.IsNotExist(err) {
			plan.ConfigWrites = append(plan.ConfigWrites, ConfigWrite{
				Path:    claudeSettingsPath,
				Type:    "claude-settings",
				Grove:   grove,
				Content: s.generateClaudeSettingsPreview(),
			})
		}
	}

	// Check for old .orc-mission files to clean up
	oldMissionFile := filepath.Join(workspacePath, ".orc-mission")
	if _, err := os.Stat(oldMissionFile); err == nil {
		plan.Cleanups = append(plan.Cleanups, CleanupAction{
			Path:   oldMissionFile,
			Reason: "legacy .orc-mission file",
		})
	}

	return plan
}

func (s *MissionOrchestrationService) analyzeGroveAction(grove *primary.Grove, currentPath, desiredPath string) GroveAction {
	action := GroveAction{
		GroveID:     grove.ID,
		GroveName:   grove.Name,
		CurrentPath: currentPath,
		DesiredPath: desiredPath,
	}

	// Check if grove exists at current path
	currentExists := false
	if _, err := os.Stat(currentPath); err == nil {
		currentExists = true
	}

	// Check if grove exists at desired path
	desiredExists := false
	if _, err := os.Stat(desiredPath); err == nil {
		desiredExists = true
	}

	action.PathExists = currentExists || desiredExists

	if currentPath != desiredPath {
		if currentExists && !desiredExists {
			action.Action = "move"
			action.UpdateDBPath = true
		} else if !currentExists && !desiredExists {
			action.Action = "missing"
		} else if desiredExists {
			action.Action = "exists"
			if currentPath != desiredPath {
				action.UpdateDBPath = true
			}
		}
	} else {
		if currentExists {
			action.Action = "exists"
		} else {
			action.Action = "missing"
		}
	}

	return action
}

func (s *MissionOrchestrationService) generateGroveConfigPreview(grove *primary.Grove) string {
	reposJSON := "[]"
	if len(grove.Repos) > 0 {
		reposBytes, _ := json.Marshal(grove.Repos)
		reposJSON = string(reposBytes)
	}

	return fmt.Sprintf(`{
  "version": "1.0",
  "type": "grove",
  "grove": {
    "grove_id": "%s",
    "mission_id": "%s",
    "name": "%s",
    "repos": %s,
    "created_at": "<timestamp>"
  }
}`, grove.ID, grove.MissionID, grove.Name, reposJSON)
}

func (s *MissionOrchestrationService) generateClaudeSettingsPreview() string {
	return `{
  "permissions": {
    "defaultMode": "default"
  },
  "enabledPlugins": {
    "developer-tools@intercom-plugins": false
  },
  "hooks": {}
}`
}

// ApplyInfrastructure applies the infrastructure plan.
func (s *MissionOrchestrationService) ApplyInfrastructure(ctx context.Context, plan *InfrastructurePlan) *InfrastructureApplyResult {
	result := &InfrastructureApplyResult{}

	// Create workspace directory
	if err := os.MkdirAll(plan.WorkspacePath, 0755); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to create workspace: %v", err))
	} else {
		result.WorkspaceCreated = plan.CreateWorkspace
	}

	// Create groves directory
	if err := os.MkdirAll(plan.GrovesDir, 0755); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to create groves directory: %v", err))
	} else {
		result.GrovesDirCreated = plan.CreateGrovesDir
	}

	// Process each grove
	for _, action := range plan.GroveActions {
		if action.Action == "move" {
			if err := os.Rename(action.CurrentPath, action.DesiredPath); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("could not move grove %s: %v", action.GroveID, err))
				continue
			}
		}

		if action.Action == "missing" {
			result.GrovesNeedingWork = append(result.GrovesNeedingWork, GroveNeedingWork{
				GroveID:     action.GroveID,
				GroveName:   action.GroveName,
				DesiredPath: action.DesiredPath,
				Message:     "worktree missing, needs materialization",
			})
			continue
		}

		// Create .orc directory in grove
		groveOrcDir := filepath.Join(action.DesiredPath, ".orc")
		if _, err := os.Stat(action.DesiredPath); err == nil {
			os.MkdirAll(groveOrcDir, 0755)
		}

		// Update DB path if needed
		if action.UpdateDBPath {
			if err := s.groveSvc.UpdateGrovePath(ctx, action.GroveID, action.DesiredPath); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to update DB path for grove %s: %v", action.GroveID, err))
			}
		}

		result.GrovesProcessed++
	}

	// Write config files
	for _, configWrite := range plan.ConfigWrites {
		// Only write if grove directory exists
		groveDir := filepath.Dir(filepath.Dir(configWrite.Path))
		if _, err := os.Stat(groveDir); os.IsNotExist(err) {
			continue
		}

		var err error
		if configWrite.Type == "grove" {
			err = s.writeGroveConfig(configWrite.Path, configWrite.Grove)
		} else if configWrite.Type == "claude-settings" {
			err = s.writeClaudeSettings(configWrite.Path)
		}

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to write config %s: %v", configWrite.Path, err))
		} else {
			result.ConfigsWritten++
		}
	}

	// Perform cleanups
	for _, cleanup := range plan.Cleanups {
		if err := os.Remove(cleanup.Path); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("could not remove %s: %v", cleanup.Path, err))
		} else {
			result.CleanupsDone++
		}
	}

	return result
}

func (s *MissionOrchestrationService) writeGroveConfig(path string, grove *primary.Grove) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	cfg := &config.Config{
		Version: "1.0",
		Type:    config.TypeGrove,
		Grove: &config.GroveConfig{
			GroveID:   grove.ID,
			MissionID: grove.MissionID,
			Name:      grove.Name,
			Repos:     grove.Repos,
			CreatedAt: grove.CreatedAt,
		},
	}

	// Use config package's save function if available, otherwise manual write
	grovePath := filepath.Dir(dir) // Go up from .orc to grove root
	return config.SaveConfig(grovePath, cfg)
}

func (s *MissionOrchestrationService) writeClaudeSettings(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	settings := map[string]interface{}{
		"permissions": map[string]interface{}{
			"defaultMode": "default",
		},
		"enabledPlugins": map[string]bool{
			"developer-tools@intercom-plugins": false,
		},
		"hooks": map[string]interface{}{},
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings: %w", err)
	}

	return nil
}

// TmuxSessionPlan describes the TMux session to create or update.
type TmuxSessionPlan struct {
	SessionName   string
	WorkingDir    string
	SessionExists bool
	WindowPlans   []WindowPlan
}

// WindowPlan describes a TMux window to create or update.
type WindowPlan struct {
	Index       int
	Name        string
	GroveID     string
	GrovePath   string
	Action      string // "create", "exists", "update", "skip"
	PaneCount   int
	NeedsUpdate bool
}

// TmuxSessionResult captures the result of applying TMux session changes.
type TmuxSessionResult struct {
	SessionCreated bool
	WindowsCreated int
	WindowsUpdated int
	Errors         []string
}

// PlanTmuxSession generates a plan for the TMux session.
func (s *MissionOrchestrationService) PlanTmuxSession(state *MissionState, workspacePath, sessionName string, sessionExists bool, windowChecker TmuxWindowChecker) *TmuxSessionPlan {
	plan := &TmuxSessionPlan{
		SessionName:   sessionName,
		WorkingDir:    workspacePath,
		SessionExists: sessionExists,
	}

	grovesDir := filepath.Join(workspacePath, "groves")

	for i, grove := range state.Groves {
		windowIndex := i + 1
		grovePath := filepath.Join(grovesDir, grove.Name)

		windowPlan := WindowPlan{
			Index:     windowIndex,
			Name:      grove.Name,
			GroveID:   grove.ID,
			GrovePath: grovePath,
		}

		// Check if grove path exists
		if _, err := os.Stat(grovePath); os.IsNotExist(err) {
			windowPlan.Action = "skip"
			plan.WindowPlans = append(plan.WindowPlans, windowPlan)
			continue
		}

		if sessionExists && windowChecker != nil {
			if windowChecker.WindowExists(sessionName, grove.Name) {
				paneCount := windowChecker.GetPaneCount(sessionName, grove.Name)
				pane2Cmd := windowChecker.GetPaneCommand(sessionName, grove.Name, 2)

				windowPlan.PaneCount = paneCount
				if paneCount == 3 && pane2Cmd == "orc" {
					windowPlan.Action = "exists"
				} else if paneCount == 3 {
					windowPlan.Action = "update"
					windowPlan.NeedsUpdate = true
				} else {
					windowPlan.Action = "update"
					windowPlan.NeedsUpdate = true
				}
			} else {
				windowPlan.Action = "create"
			}
		} else {
			windowPlan.Action = "create"
		}

		plan.WindowPlans = append(plan.WindowPlans, windowPlan)
	}

	return plan
}

// TmuxWindowChecker is an interface for checking TMux window state.
// This allows dependency injection for testing.
type TmuxWindowChecker interface {
	WindowExists(session, window string) bool
	GetPaneCount(session, window string) int
	GetPaneCommand(session, window string, pane int) string
}

// DefaultWorkspacePath returns the default workspace path for a mission.
func DefaultWorkspacePath(missionID string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, "src", "missions", missionID), nil
}

// FormatTimestamp formats a time for display.
func FormatTimestamp(t time.Time) string {
	return t.Format(time.RFC3339)
}
