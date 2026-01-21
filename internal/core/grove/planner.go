package grove

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/example/orc/internal/core/effects"
)

// CreateGrovePlanInput contains pre-fetched data for grove creation.
type CreateGrovePlanInput struct {
	GroveID   string
	GroveName string
	MissionID string
	BasePath  string   // e.g., ~/src/worktrees
	Repos     []string // Repository names
}

// OpenGrovePlanInput contains pre-fetched data for grove opening.
type OpenGrovePlanInput struct {
	GroveID         string
	GroveName       string
	GrovePath       string
	SessionName     string // Current TMux session
	NextWindowIndex int
}

// CreateGrovePlan represents the planned effects for creating a grove.
type CreateGrovePlan struct {
	GroveID       string
	GrovePath     string
	FilesystemOps []effects.FileEffect
	GitOps        []effects.GitEffect
	DatabaseOps   []effects.PersistEffect
}

// Effects returns all effects as a flat slice for execution.
func (p CreateGrovePlan) Effects() []effects.Effect {
	result := make([]effects.Effect, 0, len(p.FilesystemOps)+len(p.GitOps)+len(p.DatabaseOps))
	for _, e := range p.FilesystemOps {
		result = append(result, e)
	}
	for _, e := range p.GitOps {
		result = append(result, e)
	}
	for _, e := range p.DatabaseOps {
		result = append(result, e)
	}
	return result
}

// OpenGrovePlan represents the planned effects for opening a grove.
type OpenGrovePlan struct {
	GroveID   string
	GrovePath string
	TMuxOps   []effects.TMuxEffect
}

// Effects returns all effects as a flat slice for execution.
func (p OpenGrovePlan) Effects() []effects.Effect {
	result := make([]effects.Effect, 0, len(p.TMuxOps))
	for _, e := range p.TMuxOps {
		result = append(result, e)
	}
	return result
}

// GenerateCreateGrovePlan creates a plan for grove creation.
// This is a pure function - all input data must be pre-fetched.
func GenerateCreateGrovePlan(input CreateGrovePlanInput) CreateGrovePlan {
	// Build grove path: {basePath}/{missionID}-{groveName}
	grovePathName := fmt.Sprintf("%s-%s", input.MissionID, input.GroveName)
	grovePath := filepath.Join(input.BasePath, grovePathName)

	plan := CreateGrovePlan{
		GroveID:   input.GroveID,
		GrovePath: grovePath,
	}

	// 1. Create grove directory
	plan.FilesystemOps = append(plan.FilesystemOps, effects.FileEffect{
		Operation: "mkdir",
		Path:      grovePath,
		Mode:      0755,
	})

	// 2. Create .orc directory
	orcDir := filepath.Join(grovePath, ".orc")
	plan.FilesystemOps = append(plan.FilesystemOps, effects.FileEffect{
		Operation: "mkdir",
		Path:      orcDir,
		Mode:      0755,
	})

	// 3. Write .orc/config.json
	configContent := generateGroveConfig(input.GroveID, input.MissionID, input.GroveName, input.Repos)
	plan.FilesystemOps = append(plan.FilesystemOps, effects.FileEffect{
		Operation: "write",
		Path:      filepath.Join(orcDir, "config.json"),
		Content:   configContent,
		Mode:      0644,
	})

	// 4. Git worktree operations for each repo
	for _, repo := range input.Repos {
		plan.GitOps = append(plan.GitOps, effects.GitEffect{
			Operation: "worktree_add",
			RepoPath:  repo, // Will be resolved to ~/src/{repo}
			Args:      []string{grovePath, "-b", input.GroveName},
		})
	}

	// 5. Update database with grove path
	plan.DatabaseOps = append(plan.DatabaseOps, effects.PersistEffect{
		Entity:    "grove",
		Operation: "update",
		Data: map[string]string{
			"id":   input.GroveID,
			"path": grovePath,
		},
	})

	return plan
}

// GenerateOpenGrovePlan creates a plan for opening a grove in TMux.
// This is a pure function - all input data must be pre-fetched.
func GenerateOpenGrovePlan(input OpenGrovePlanInput) OpenGrovePlan {
	plan := OpenGrovePlan{
		GroveID:   input.GroveID,
		GrovePath: input.GrovePath,
	}

	// TMux window with 3-pane IMP layout:
	// +-------------------+-------------------+
	// |                   | claude (IMP)      |
	// | vim               +-------------------+
	// |                   | shell             |
	// +-------------------+-------------------+

	// 1. Create new window
	plan.TMuxOps = append(plan.TMuxOps, effects.TMuxEffect{
		Operation:   "new_window",
		SessionName: input.SessionName,
		WindowName:  input.GroveName,
		Command:     input.GrovePath, // Working directory
	})

	// 2. Split vertical (creates right pane)
	plan.TMuxOps = append(plan.TMuxOps, effects.TMuxEffect{
		Operation:   "split_vertical",
		SessionName: input.SessionName,
		WindowName:  input.GroveName,
		Command:     input.GrovePath,
	})

	// 3. Split horizontal on right pane
	plan.TMuxOps = append(plan.TMuxOps, effects.TMuxEffect{
		Operation:   "split_horizontal",
		SessionName: input.SessionName,
		WindowName:  input.GroveName,
		Command:     input.GrovePath,
	})

	// 4. Launch vim in pane 1 (left)
	plan.TMuxOps = append(plan.TMuxOps, effects.TMuxEffect{
		Operation:   "send_keys",
		SessionName: input.SessionName,
		WindowName:  fmt.Sprintf("%s.1", input.GroveName),
		Command:     "vim",
	})

	// 5. Launch claude IMP in pane 2 (top right)
	plan.TMuxOps = append(plan.TMuxOps, effects.TMuxEffect{
		Operation:   "send_keys",
		SessionName: input.SessionName,
		WindowName:  fmt.Sprintf("%s.2", input.GroveName),
		Command:     `claude "Run the orc prime command to get context"`,
	})

	return plan
}

// groveConfig structure for JSON serialization
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

func generateGroveConfig(groveID, missionID, name string, repos []string) []byte {
	if repos == nil {
		repos = []string{}
	}
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
