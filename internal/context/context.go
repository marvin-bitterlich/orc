package context

import (
	"os"
	"path/filepath"

	"github.com/example/orc/internal/config"
)

// WorkbenchContext represents workbench context information (IMP territory)
type WorkbenchContext struct {
	WorkbenchID string `json:"workbench_id"`
	Role        string `json:"role"`
	ConfigPath  string `json:"config_path"` // Path to .orc/config.json
}

// DetectWorkbenchContext checks if we're in a workbench context (IMP territory)
// by looking for .orc/config.json in current directory only (no tree walking)
func DetectWorkbenchContext() (*WorkbenchContext, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	cfg, err := config.LoadConfig(dir)
	if err != nil {
		return nil, nil //nolint:nilerr // No config found is not an error, just no context
	}

	// Only return context if this is an IMP config with workbench
	if cfg.Role == config.RoleIMP && cfg.WorkbenchID != "" {
		return &WorkbenchContext{
			WorkbenchID: cfg.WorkbenchID,
			Role:        cfg.Role,
			ConfigPath:  filepath.Join(dir, ".orc", "config.json"),
		}, nil
	}

	return nil, nil
}

// GetContextCommissionID returns the commission ID from workbench context.
// For IMP contexts, looks up commission via workbench → workshop → factory → commission chain.
// Returns empty string if no context found - caller should handle this.
func GetContextCommissionID() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	cfg, err := config.LoadConfig(dir)
	if err != nil {
		return ""
	}

	// For IMP role, look up commission through workbench chain
	if cfg.Role == config.RoleIMP && cfg.WorkbenchID != "" {
		return getCommissionFromWorkbench(cfg.WorkbenchID)
	}

	return ""
}

// getCommissionFromWorkbench looks up commission ID via workbench → workshop → factory chain.
// This is a placeholder that returns empty - requires DB access which is wired elsewhere.
func getCommissionFromWorkbench(_ string) string {
	// Note: Full implementation requires DB lookup through wire.
	// For now, commission context comes from focused containers.
	return ""
}

// WriteCommissionContext creates a .orc/config.json file for a commission workspace.
// Uses Goblin role by default since commission workspaces are for orchestration.
// If the path matches a workshop pattern, sets WorkshopID.
func WriteCommissionContext(workspacePath string) error {
	cfg := &config.Config{
		Version:    "1.0",
		Role:       config.RoleGoblin,
		WorkshopID: config.ParseWorkshopIDFromPath(workspacePath),
	}
	return config.SaveConfig(workspacePath, cfg)
}

// WriteWorkbenchContext creates a .orc/config.json file for a workbench (IMP territory)
func WriteWorkbenchContext(workbenchPath, workbenchID string) error {
	cfg := &config.Config{
		Version:     "1.0",
		Role:        config.RoleIMP,
		WorkbenchID: workbenchID,
	}
	return config.SaveConfig(workbenchPath, cfg)
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
