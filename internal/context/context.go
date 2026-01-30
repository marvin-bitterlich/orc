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

	// Only return context if this is a workbench (BENCH-XXX place_id)
	if config.IsWorkbench(cfg.PlaceID) {
		return &WorkbenchContext{
			WorkbenchID: cfg.PlaceID,
			Role:        config.RoleIMP,
			ConfigPath:  filepath.Join(dir, ".orc", "config.json"),
		}, nil
	}

	return nil, nil
}

// GetContextWorkbenchID returns the workbench ID (BENCH-xxx) if we're in a workbench context.
// Returns empty string if not in a workbench context.
func GetContextWorkbenchID() string {
	ctx, err := DetectWorkbenchContext()
	if err != nil || ctx == nil {
		return ""
	}
	return ctx.WorkbenchID
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

	// For workbench place, look up commission through workbench chain
	if config.IsWorkbench(cfg.PlaceID) {
		return getCommissionFromWorkbench(cfg.PlaceID)
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

// WriteGatehouseContext creates a .orc/config.json file for a gatehouse (Goblin territory)
func WriteGatehouseContext(gatehousePath, gatehouseID string) error {
	cfg := &config.Config{
		Version: "1.0",
		PlaceID: gatehouseID, // GATE-XXX
	}
	return config.SaveConfig(gatehousePath, cfg)
}

// WriteCommissionContext creates a minimal .orc/config.json for legacy commission workspaces.
// These are not associated with a gatehouse, so place_id is empty.
// This is deprecated - new workflows should use workshops with gatehouses.
func WriteCommissionContext(workspacePath string) error {
	cfg := &config.Config{
		Version: "1.0",
		// No place_id - this is a legacy commission workspace without a gatehouse
	}
	return config.SaveConfig(workspacePath, cfg)
}

// WriteWorkbenchContext creates a .orc/config.json file for a workbench (IMP territory)
func WriteWorkbenchContext(workbenchPath, workbenchID string) error {
	cfg := &config.Config{
		Version: "1.0",
		PlaceID: workbenchID, // BENCH-XXX
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
