package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Role constants
const (
	RoleGoblin = "GOBLIN" // Orchestrator agent (formerly ORC)
	RoleIMP    = "IMP"    // Implementation agent
)

// Config represents the flat ORC configuration
type Config struct {
	Version      string `json:"version"`
	Role         string `json:"role"`                    // "GOBLIN" or "IMP"
	WorkbenchID  string `json:"workbench_id,omitempty"`  // BENCH-XXX (for IMP)
	CommissionID string `json:"commission_id,omitempty"` // COMM-XXX
	CurrentFocus string `json:"current_focus,omitempty"` // SHIP-*, CON-*, etc.
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
