package context

import (
	gocontext "context"
	"os"
	"path/filepath"

	"github.com/example/orc/internal/config"
	"github.com/example/orc/internal/wire"
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
// For IMP contexts, looks up commission via workbench's focused shipment or tome.
// Returns empty string if no context found.
func GetContextCommissionID() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	cfg, err := config.LoadConfig(dir)
	if err != nil {
		return ""
	}

	// For workbench place, look up commission through focused container
	if config.IsWorkbench(cfg.PlaceID) {
		return getCommissionFromWorkbench(cfg.PlaceID)
	}

	return ""
}

// getCommissionFromWorkbench looks up commission ID via workbench's focused container.
// Returns the commission that the workbench's focused shipment/tome belongs to.
func getCommissionFromWorkbench(workbenchID string) string {
	ctx := gocontext.Background()

	// Get workbench's focused ID
	focusedID, err := wire.WorkbenchService().GetFocusedID(ctx, workbenchID)
	if err != nil || focusedID == "" {
		return ""
	}

	// Look up commission from focused container
	switch {
	case len(focusedID) > 5 && focusedID[:5] == "SHIP-":
		ship, err := wire.ShipmentService().GetShipment(ctx, focusedID)
		if err != nil {
			return ""
		}
		return ship.CommissionID
	case len(focusedID) > 5 && focusedID[:5] == "TOME-":
		tome, err := wire.TomeService().GetTome(ctx, focusedID)
		if err != nil {
			return ""
		}
		return tome.CommissionID
	}

	return ""
}

// GetContextFactoryID returns the factory ID from workbench context.
// For IMP contexts, looks up factory via workbench → workshop → factory.
// Returns empty string if no context found.
func GetContextFactoryID() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	cfg, err := config.LoadConfig(dir)
	if err != nil {
		return ""
	}

	// For workbench place, look up factory through workbench chain
	if config.IsWorkbench(cfg.PlaceID) {
		return getFactoryFromWorkbench(cfg.PlaceID)
	}

	return ""
}

// getFactoryFromWorkbench looks up factory ID via workbench → workshop → factory chain.
func getFactoryFromWorkbench(workbenchID string) string {
	ctx := gocontext.Background()

	// 1. Get workbench to find workshop
	bench, err := wire.WorkbenchService().GetWorkbench(ctx, workbenchID)
	if err != nil || bench.WorkshopID == "" {
		return ""
	}

	// 2. Get workshop to find factory
	workshop, err := wire.WorkshopService().GetWorkshop(ctx, bench.WorkshopID)
	if err != nil {
		return ""
	}

	return workshop.FactoryID
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
