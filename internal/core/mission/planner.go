// Package mission contains the pure business logic for mission operations.
// This file contains pure planner functions that generate effects.
package mission

import (
	"encoding/json"
	"path/filepath"
	"time"

	"github.com/example/orc/internal/core/effects"
)

// GrovePlanInput represents a grove for planning purposes.
type GrovePlanInput struct {
	ID          string
	Name        string
	CurrentPath string   // Current path in DB (may differ from desired)
	Repos       []string // List of repo URLs
	PathExists  bool     // Does the worktree exist on disk?
}

// LaunchPlanInput contains the inputs needed to generate a launch plan.
// All values are pre-fetched by the caller - no I/O in the planner.
type LaunchPlanInput struct {
	MissionID     string
	MissionTitle  string
	WorkspacePath string
	CreateTMux    bool
	Groves        []GrovePlanInput
}

// LaunchPlan represents the planned effects for launching a mission.
type LaunchPlan struct {
	MissionID     string
	WorkspacePath string
	FilesystemOps []effects.FileEffect
	DatabaseOps   []effects.PersistEffect
	TMuxOps       []effects.TMuxEffect
}

// Effects returns all effects as a flat slice for execution.
func (p LaunchPlan) Effects() []effects.Effect {
	result := make([]effects.Effect, 0, len(p.FilesystemOps)+len(p.DatabaseOps)+len(p.TMuxOps))
	for _, e := range p.FilesystemOps {
		result = append(result, e)
	}
	for _, e := range p.DatabaseOps {
		result = append(result, e)
	}
	for _, e := range p.TMuxOps {
		result = append(result, e)
	}
	return result
}

// GenerateLaunchPlan creates a plan for launching mission infrastructure.
// This is a pure function - all input data must be pre-fetched.
func GenerateLaunchPlan(input LaunchPlanInput) LaunchPlan {
	plan := LaunchPlan{
		MissionID:     input.MissionID,
		WorkspacePath: input.WorkspacePath,
	}

	grovesDir := filepath.Join(input.WorkspacePath, "groves")

	// 1. Create workspace directory
	plan.FilesystemOps = append(plan.FilesystemOps, effects.FileEffect{
		Operation: "mkdir",
		Path:      input.WorkspacePath,
		Mode:      0755,
	})

	// 2. Create groves directory
	plan.FilesystemOps = append(plan.FilesystemOps, effects.FileEffect{
		Operation: "mkdir",
		Path:      grovesDir,
		Mode:      0755,
	})

	// 3. Process each grove
	for _, grove := range input.Groves {
		desiredPath := filepath.Join(grovesDir, grove.Name)

		// Create .orc directory for grove config
		plan.FilesystemOps = append(plan.FilesystemOps, effects.FileEffect{
			Operation: "mkdir",
			Path:      filepath.Join(desiredPath, ".orc"),
			Mode:      0755,
		})

		// Generate and write grove config
		configContent := generateGroveConfig(grove.ID, input.MissionID, grove.Name, grove.Repos)
		plan.FilesystemOps = append(plan.FilesystemOps, effects.FileEffect{
			Operation: "write",
			Path:      filepath.Join(desiredPath, ".orc", "config.json"),
			Content:   configContent,
			Mode:      0644,
		})

		// Update DB path if different from desired
		if grove.CurrentPath != desiredPath {
			plan.DatabaseOps = append(plan.DatabaseOps, effects.PersistEffect{
				Entity:    "grove",
				Operation: "update",
				Data: map[string]string{
					"id":   grove.ID,
					"path": desiredPath,
				},
			})
		}
	}

	// 4. TMux operations (optional)
	if input.CreateTMux {
		sessionName := "orc-" + input.MissionID

		plan.TMuxOps = append(plan.TMuxOps, effects.TMuxEffect{
			Operation:   "new_session",
			SessionName: sessionName,
		})

		for _, grove := range input.Groves {
			if grove.PathExists {
				grovePath := filepath.Join(grovesDir, grove.Name)
				plan.TMuxOps = append(plan.TMuxOps, effects.TMuxEffect{
					Operation:   "new_window",
					SessionName: sessionName,
					WindowName:  grove.Name,
					Command:     grovePath, // Path as working directory
				})
			}
		}
	}

	return plan
}

// StartPlanInput contains the inputs needed to generate a start plan.
type StartPlanInput struct {
	MissionID     string
	WorkspacePath string
	Groves        []GrovePlanInput
}

// StartPlan represents the planned effects for starting a mission.
type StartPlan struct {
	MissionID string
	TMuxOps   []effects.TMuxEffect
}

// Effects returns all effects as a flat slice for execution.
func (p StartPlan) Effects() []effects.Effect {
	result := make([]effects.Effect, 0, len(p.TMuxOps))
	for _, e := range p.TMuxOps {
		result = append(result, e)
	}
	return result
}

// GenerateStartPlan creates a plan for starting a mission's tmux session.
// This is a simpler version of launch that only handles tmux setup.
func GenerateStartPlan(input StartPlanInput) StartPlan {
	plan := StartPlan{
		MissionID: input.MissionID,
	}

	sessionName := "orc-" + input.MissionID
	grovesDir := filepath.Join(input.WorkspacePath, "groves")

	// Create new session
	plan.TMuxOps = append(plan.TMuxOps, effects.TMuxEffect{
		Operation:   "new_session",
		SessionName: sessionName,
	})

	// Create window for each grove
	for _, grove := range input.Groves {
		if grove.PathExists {
			grovePath := filepath.Join(grovesDir, grove.Name)
			plan.TMuxOps = append(plan.TMuxOps, effects.TMuxEffect{
				Operation:   "new_window",
				SessionName: sessionName,
				WindowName:  grove.Name,
				Command:     grovePath,
			})
		}
	}

	return plan
}

// groveConfig represents the structure of a grove config file.
type groveConfig struct {
	Version string           `json:"version"`
	Type    string           `json:"type"`
	Grove   groveConfigInner `json:"grove"`
}

type groveConfigInner struct {
	GroveID   string   `json:"grove_id"`
	MissionID string   `json:"mission_id"`
	Name      string   `json:"name"`
	Repos     []string `json:"repos"`
	CreatedAt string   `json:"created_at"`
}

// generateGroveConfig creates the JSON config content for a grove.
func generateGroveConfig(groveID, missionID, name string, repos []string) []byte {
	config := groveConfig{
		Version: "1.0",
		Type:    "grove",
		Grove: groveConfigInner{
			GroveID:   groveID,
			MissionID: missionID,
			Name:      name,
			Repos:     repos,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}

	data, _ := json.MarshalIndent(config, "", "  ")
	return data
}
