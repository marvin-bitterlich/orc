package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// Role constants (for backwards compatibility during migration)
const (
	RoleGoblin = "GOBLIN" // Orchestrator agent (formerly ORC)
	RoleIMP    = "IMP"    // Implementation agent
)

// Place type constants
const (
	PlaceTypeWorkbench = "workbench" // BENCH-XXX
	PlaceTypeGatehouse = "gatehouse" // GATE-XXX
	PlaceTypeKennel    = "kennel"    // KENNEL-XXX
)

// Config represents the flat ORC configuration (identity only)
// New format uses place_id; legacy role-based format is migrated on load.
type Config struct {
	Version string `json:"version"`
	PlaceID string `json:"place_id"` // BENCH-XXX or GATE-XXX
}

// legacyConfigV2 is used for reading v2 config format (role-based) during migration
type legacyConfigV2 struct {
	Version     string `json:"version"`
	Role        string `json:"role,omitempty"`
	WorkbenchID string `json:"workbench_id,omitempty"` // BENCH-XXX (for IMP)
	WorkshopID  string `json:"workshop_id,omitempty"`  // WORK-XXX (for GOBLIN)
}

// legacyConfigV1 is used for reading v1 config format during migration
type legacyConfigV1 struct {
	Version      string `json:"version"`
	Role         string `json:"role"`
	WorkbenchID  string `json:"workbench_id,omitempty"`
	CommissionID string `json:"commission_id,omitempty"` // deprecated
	CurrentFocus string `json:"current_focus,omitempty"` // deprecated
}

// LoadConfig reads .orc/config.json from the specified directory.
// Resolution order: cwd only (no home fallback).
// Returns error if no config found - caller should handle accordingly.
//
// This function handles automatic migration from legacy formats:
// - v1: {role, workbench_id, commission_id, current_focus}
// - v2: {role, workbench_id, workshop_id}
// - v3 (current): {place_id}
//
// For IMP configs (workbench_id present), migration is automatic.
// For Goblin configs (workshop_id), migration requires DB lookup for gatehouse ID.
// Use LoadConfigWithGatehouseLookup for full migration support.
func LoadConfig(dir string) (*Config, error) {
	path := filepath.Join(dir, ".orc", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Try parsing as new format first
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// If place_id is set, it's already new format
	if cfg.PlaceID != "" {
		return &cfg, nil
	}

	// Try parsing as legacy v2 format
	var legacy legacyConfigV2
	if err := json.Unmarshal(data, &legacy); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Migrate IMP configs (workbench_id → place_id)
	if legacy.Role == RoleIMP && legacy.WorkbenchID != "" {
		cfg.Version = legacy.Version
		cfg.PlaceID = legacy.WorkbenchID
		// Save migrated config
		_ = SaveConfig(dir, &cfg)
		return &cfg, nil
	}

	// For Goblin configs, we can't migrate without DB lookup
	// Return a config with empty PlaceID - caller should use LoadConfigWithGatehouseLookup
	if IsGoblinRole(legacy.Role) {
		cfg.Version = legacy.Version
		// PlaceID will be empty - caller needs to resolve gatehouse
		return &cfg, nil
	}

	return &cfg, nil
}

// SaveConfig writes config.json to directory
func SaveConfig(dir string, cfg *Config) error {
	orcDir := filepath.Join(dir, ".orc")
	if err := os.MkdirAll(orcDir, 0755); err != nil {
		return fmt.Errorf("failed to create .orc dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	path := filepath.Join(orcDir, "config.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// IsGoblinRole returns true if the role is Goblin (including backwards-compat "ORC")
func IsGoblinRole(role string) bool {
	return role == RoleGoblin || role == "ORC"
}

// GetPlaceType returns the place type for a given place ID.
// Returns "workbench" for BENCH-XXX, "gatehouse" for GATE-XXX, "kennel" for KENNEL-XXX, or "" for unknown.
func GetPlaceType(placeID string) string {
	if len(placeID) < 5 {
		return ""
	}
	// Check 6-char prefix first for KENNEL
	if len(placeID) >= 6 && placeID[:6] == "KENNEL" {
		return PlaceTypeKennel
	}
	prefix := placeID[:5]
	switch prefix {
	case "BENCH":
		return PlaceTypeWorkbench
	case "GATE-":
		return PlaceTypeGatehouse
	}
	return ""
}

// IsKennel returns true if the place ID is a kennel (KENNEL-XXX)
func IsKennel(placeID string) bool {
	return GetPlaceType(placeID) == PlaceTypeKennel
}

// IsWorkbench returns true if the place ID is a workbench (BENCH-XXX)
func IsWorkbench(placeID string) bool {
	return GetPlaceType(placeID) == PlaceTypeWorkbench
}

// IsGatehouse returns true if the place ID is a gatehouse (GATE-XXX)
func IsGatehouse(placeID string) bool {
	return GetPlaceType(placeID) == PlaceTypeGatehouse
}

// GetRoleFromPlaceID derives the agent role from a place ID.
// BENCH-XXX → IMP, GATE-XXX → GOBLIN
func GetRoleFromPlaceID(placeID string) string {
	switch GetPlaceType(placeID) {
	case PlaceTypeWorkbench:
		return RoleIMP
	case PlaceTypeGatehouse:
		return RoleGoblin
	}
	return ""
}

// DefaultWorkspacePath returns the default workspace path for a commission.
func DefaultWorkspacePath(commissionID string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, "src", "commissions", commissionID), nil
}

// workshopIDPattern matches WORK-XXX format in path segments
var workshopIDPattern = regexp.MustCompile(`WORK-\d{3}`)

// ParseWorkshopIDFromPath extracts WORK-xxx from paths like ~/.orc/ws/WORK-003-name/
func ParseWorkshopIDFromPath(path string) string {
	match := workshopIDPattern.FindString(path)
	return match
}

// MigrateConfig updates old config format to new identity-only format.
// Returns (oldFocusID, wasModified, error) - caller can use oldFocusID for DB migration.
// This is only for IMP configs; Goblin configs require DB lookup for gatehouse ID.
func MigrateConfig(dir string) (string, bool, error) {
	path := filepath.Join(dir, ".orc", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false, fmt.Errorf("failed to read config: %w", err)
	}

	// Read as legacy v1 format to capture deprecated fields
	var legacyV1 legacyConfigV1
	if err := json.Unmarshal(data, &legacyV1); err != nil {
		return "", false, fmt.Errorf("failed to parse config: %w", err)
	}

	// Extract old focus for potential DB migration
	oldFocus := legacyV1.CurrentFocus

	// Check if migration is needed
	hasDeprecatedFields := legacyV1.CommissionID != "" || legacyV1.CurrentFocus != ""
	if !hasDeprecatedFields {
		return "", false, nil // Already migrated
	}

	// Build new config using place_id
	newCfg := &Config{
		Version: legacyV1.Version,
	}

	// For IMP role, migrate workbench_id to place_id
	if legacyV1.Role == RoleIMP && legacyV1.WorkbenchID != "" {
		newCfg.PlaceID = legacyV1.WorkbenchID
	}
	// Note: Goblin configs need DB lookup for gatehouse ID - handled elsewhere

	// Save updated config
	if err := SaveConfig(dir, newCfg); err != nil {
		return oldFocus, false, fmt.Errorf("failed to save migrated config: %w", err)
	}

	return oldFocus, true, nil
}

// LoadLegacyConfig reads config and returns the legacy format for callers that need
// to inspect workshop_id for gatehouse lookup.
func LoadLegacyConfig(dir string) (*legacyConfigV2, error) {
	path := filepath.Join(dir, ".orc", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var legacy legacyConfigV2
	if err := json.Unmarshal(data, &legacy); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &legacy, nil
}
