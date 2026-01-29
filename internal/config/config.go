package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// Role constants
const (
	RoleGoblin = "GOBLIN" // Orchestrator agent (formerly ORC)
	RoleIMP    = "IMP"    // Implementation agent
)

// Config represents the flat ORC configuration (identity only)
type Config struct {
	Version     string `json:"version"`
	Role        string `json:"role"`                   // "GOBLIN" or "IMP"
	WorkbenchID string `json:"workbench_id,omitempty"` // BENCH-XXX (for IMP)
	WorkshopID  string `json:"workshop_id,omitempty"`  // WORK-XXX (for GOBLIN)
}

// legacyConfig is used for reading old config format during migration
type legacyConfig struct {
	Version      string `json:"version"`
	Role         string `json:"role"`
	WorkbenchID  string `json:"workbench_id,omitempty"`
	CommissionID string `json:"commission_id,omitempty"` // deprecated
	CurrentFocus string `json:"current_focus,omitempty"` // deprecated
}

// LoadConfig reads .orc/config.json from the specified directory.
// Resolution order: cwd only (no home fallback).
// Returns error if no config found - caller should handle accordingly.
func LoadConfig(dir string) (*Config, error) {
	path := filepath.Join(dir, ".orc", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
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
func MigrateConfig(dir string) (string, bool, error) {
	path := filepath.Join(dir, ".orc", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false, fmt.Errorf("failed to read config: %w", err)
	}

	// Read as legacy format to capture deprecated fields
	var legacy legacyConfig
	if err := json.Unmarshal(data, &legacy); err != nil {
		return "", false, fmt.Errorf("failed to parse config: %w", err)
	}

	// Extract old focus for potential DB migration
	oldFocus := legacy.CurrentFocus

	// Check if migration is needed
	hasDeprecatedFields := legacy.CommissionID != "" || legacy.CurrentFocus != ""
	if !hasDeprecatedFields {
		return "", false, nil // Already migrated
	}

	// Build new config (identity only)
	newCfg := &Config{
		Version:     legacy.Version,
		Role:        legacy.Role,
		WorkbenchID: legacy.WorkbenchID,
	}

	// For Goblin role, try to parse workshop ID from directory path
	if IsGoblinRole(legacy.Role) {
		workshopID := ParseWorkshopIDFromPath(dir)
		if workshopID != "" {
			newCfg.WorkshopID = workshopID
		}
	}

	// Save updated config
	if err := SaveConfig(dir, newCfg); err != nil {
		return oldFocus, false, fmt.Errorf("failed to save migrated config: %w", err)
	}

	return oldFocus, true, nil
}
