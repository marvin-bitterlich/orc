package cli

import (
	"context"

	"github.com/example/orc/internal/config"
	"github.com/example/orc/internal/wire"
)

// MigrateGoblinConfigIfNeeded checks for old Goblin configs and migrates them.
// Old Goblin configs have {role: "GOBLIN", workshop_id: "WORK-xxx"} but no place_id.
// This function looks up the gatehouse ID for the workshop and saves the migrated config.
// Returns the (potentially updated) config.
func MigrateGoblinConfigIfNeeded(ctx context.Context, dir string) (*config.Config, error) {
	cfg, err := config.LoadConfig(dir)
	if err != nil {
		return nil, err
	}

	// Already migrated - has place_id
	if cfg.PlaceID != "" {
		return cfg, nil
	}

	// Check if this is an old Goblin config with workshop_id
	legacy, err := config.LoadLegacyConfig(dir)
	if err != nil {
		return cfg, nil //nolint:nilerr // Can't read legacy format, return as-is (graceful degradation)
	}

	// Only migrate Goblin configs with workshop_id
	if !config.IsGoblinRole(legacy.Role) || legacy.WorkshopID == "" {
		return cfg, nil
	}

	// Look up gatehouse for this workshop
	gatehouse, err := wire.GatehouseService().GetGatehouseByWorkshop(ctx, legacy.WorkshopID)
	if err != nil {
		return cfg, nil //nolint:nilerr // Gatehouse not found, return as-is (graceful degradation)
	}

	// Migrate config
	cfg.PlaceID = gatehouse.ID
	if err := config.SaveConfig(dir, cfg); err != nil {
		return cfg, nil //nolint:nilerr // Save failed, return unmigrated (graceful degradation)
	}

	return cfg, nil
}
