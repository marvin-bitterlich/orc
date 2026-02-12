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
)

// Config represents the flat ORC configuration (identity only)
// New format uses place_id; legacy role-based format is migrated on load.
type Config struct {
	Version string `json:"version"`
	PlaceID string `json:"place_id"` // BENCH-XXX
}

// legacyIMPConfig is used for reading old IMP config format during migration
type legacyIMPConfig struct {
	Version     string `json:"version"`
	Role        string `json:"role,omitempty"`
	WorkbenchID string `json:"workbench_id,omitempty"` // BENCH-XXX (for IMP)
}

// LoadConfig reads .orc/config.json from the specified directory.
// Resolution order: cwd only (no home fallback).
// Returns error if no config found - caller should handle accordingly.
//
// Handles automatic migration from legacy IMP format:
// - Legacy: {role, workbench_id}
// - Current: {place_id}
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

	// Try parsing as legacy IMP format
	var legacy legacyIMPConfig
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

// GetPlaceType returns the place type for a given place ID.
// Returns "workbench" for BENCH-XXX, or "" for unknown.
func GetPlaceType(placeID string) string {
	if len(placeID) < 5 {
		return ""
	}
	if placeID[:5] == "BENCH" {
		return PlaceTypeWorkbench
	}
	return ""
}

// IsWorkbench returns true if the place ID is a workbench (BENCH-XXX)
func IsWorkbench(placeID string) bool {
	return GetPlaceType(placeID) == PlaceTypeWorkbench
}

// GetRoleFromPlaceID derives the agent role from a place ID.
// BENCH-XXX → IMP
func GetRoleFromPlaceID(placeID string) string {
	if GetPlaceType(placeID) == PlaceTypeWorkbench {
		return RoleIMP
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
