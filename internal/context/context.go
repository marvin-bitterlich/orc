package context

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/example/orc/internal/config"
)

// MissionContext represents mission context information
type MissionContext struct {
	MissionID     string    `json:"mission_id"`
	WorkspacePath string    `json:"workspace_path"`
	IsMaster      bool      `json:"is_master"`
	CreatedAt     time.Time `json:"created_at"`
}

// GroveContext represents grove context information (IMP territory)
type GroveContext struct {
	GroveID    string    `json:"grove_id"`
	MissionID  string    `json:"mission_id"`
	Name       string    `json:"name"`
	Repos      []string  `json:"repos"`
	CreatedAt  time.Time `json:"created_at"`
	GrovePath  string    `json:"grove_path"`  // Full path to grove directory
	ConfigPath string    `json:"config_path"` // Path to .orc/config.json
}

// DetectGroveContext checks if we're in a grove context (IMP territory)
// by looking for .orc/config.json of type "grove" in current directory or parents
func DetectGroveContext() (*GroveContext, error) {
	// Start from current directory
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Walk up directory tree looking for grove config
	for {
		cfg, err := config.LoadConfigWithFallback(dir)
		if err == nil && cfg.Type == config.TypeGrove {
			// Found grove config - convert to GroveContext
			createdAt, _ := time.Parse(time.RFC3339, cfg.Grove.CreatedAt)
			return &GroveContext{
				GroveID:    cfg.Grove.GroveID,
				MissionID:  cfg.Grove.MissionID,
				Name:       cfg.Grove.Name,
				Repos:      cfg.Grove.Repos,
				CreatedAt:  createdAt,
				GrovePath:  dir,
				ConfigPath: filepath.Join(dir, ".orc", "config.json"),
			}, nil
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding config
			return nil, nil
		}
		dir = parent
	}
}

// DetectMissionContext checks if we're in a mission context
// by looking for .orc/config.json in current directory or parents
func DetectMissionContext() (*MissionContext, error) {
	// Start from current directory
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Walk up directory tree looking for config
	for {
		cfg, err := config.LoadConfigWithFallback(dir)
		if err == nil && cfg.Type == config.TypeMission {
			// Found mission config - convert to MissionContext
			createdAt, _ := time.Parse(time.RFC3339, cfg.Mission.CreatedAt)
			return &MissionContext{
				MissionID:     cfg.Mission.MissionID,
				WorkspacePath: cfg.Mission.WorkspacePath,
				IsMaster:      cfg.Mission.IsMaster,
				CreatedAt:     createdAt,
			}, nil
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding config
			return nil, nil
		}
		dir = parent
	}
}

// WriteMissionContext creates a .orc/config.json file for mission context
func WriteMissionContext(workspacePath, missionID string) error {
	return WriteMissionConfig(workspacePath, missionID, false)
}

// WriteMissionConfig creates a .orc/config.json file with full control over fields
func WriteMissionConfig(workspacePath, missionID string, isMaster bool) error {
	cfg := &config.Config{
		Version: "1.0",
		Type:    config.TypeMission,
		Mission: &config.MissionConfig{
			MissionID:     missionID,
			WorkspacePath: workspacePath,
			IsMaster:      isMaster,
			CreatedAt:     time.Now().Format(time.RFC3339),
		},
	}

	return config.SaveConfig(workspacePath, cfg)
}

// GetContextMissionID returns the mission ID from context, checking grove first, then mission, then global
// Returns empty string if no context found - caller should handle this as an error
func GetContextMissionID() string {
	// Check grove context first (most specific - IMP territory)
	groveCtx, err := DetectGroveContext()
	if err == nil && groveCtx != nil && groveCtx.MissionID != "" {
		fmt.Fprintf(os.Stderr, "(using grove context: %s)\n", groveCtx.MissionID)
		return groveCtx.MissionID
	}

	// Check mission context (ORC territory)
	missionCtx, err := DetectMissionContext()
	if err == nil && missionCtx != nil && missionCtx.MissionID != "" {
		fmt.Fprintf(os.Stderr, "(using mission context: %s)\n", missionCtx.MissionID)
		return missionCtx.MissionID
	}

	// Check global state (~/.orc/config.json)
	homeDir, err := os.UserHomeDir()
	if err == nil {
		cfg, err := config.LoadConfig(homeDir)
		if err == nil && cfg.Type == config.TypeGlobal && cfg.State != nil && cfg.State.ActiveMissionID != "" {
			fmt.Fprintf(os.Stderr, "(using global context: %s)\n", cfg.State.ActiveMissionID)
			return cfg.State.ActiveMissionID
		}
	}

	return "" // No context found - caller must handle
}

// IsMissionContext returns true if we're running in a mission context
func IsMissionContext() bool {
	ctx, _ := DetectMissionContext()
	return ctx != nil
}

// IsOrcSourceDirectory checks if the current directory is the ORC source code directory
// Used to prevent accidental modification of the orchestrator source by IMPs
func IsOrcSourceDirectory() bool {
	// Check for key ORC source files
	markers := []string{"cmd/orc/main.go", "internal/db/schema.go", "go.mod"}

	for _, marker := range markers {
		if _, err := os.Stat(marker); err == nil {
			// Check if go.mod contains ORC module
			if marker == "go.mod" {
				data, err := os.ReadFile(marker)
				if err == nil && len(data) > 0 {
					// Simple check for orc module name
					content := string(data)
					if len(content) > 20 && (content[:20] == "module github.com/lo" || content[:30] == "module github.com/example/orc") {
						return true
					}
				}
			} else {
				return true
			}
		}
	}

	return false
}
