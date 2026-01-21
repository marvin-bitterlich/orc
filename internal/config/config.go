package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ConfigType string

const (
	TypeMission ConfigType = "mission"
	TypeGrove   ConfigType = "grove"
	TypeGlobal  ConfigType = "global"
)

// Role constants
const (
	RoleORC = "ORC" // Orchestrator agent
	RoleIMP = "IMP" // Implementation agent
)

type Config struct {
	Version string     `json:"version"`
	Type    ConfigType `json:"type"`

	Mission *MissionConfig `json:"mission,omitempty"`
	Grove   *GroveConfig   `json:"grove,omitempty"`
	State   *StateConfig   `json:"state,omitempty"`
}

type MissionConfig struct {
	MissionID     string `json:"mission_id"`
	WorkspacePath string `json:"workspace_path"`
	IsMaster      bool   `json:"is_master,omitempty"`
	Role          string `json:"role,omitempty"`          // "ORC", "IMP", or empty
	CurrentFocus  string `json:"current_focus,omitempty"` // Focused container ID (SHIP-*, CON-*, INV-*, TOME-*)
	CreatedAt     string `json:"created_at"`
}

type GroveConfig struct {
	GroveID      string   `json:"grove_id"`
	MissionID    string   `json:"mission_id"`
	Name         string   `json:"name"`
	Repos        []string `json:"repos"`
	Role         string   `json:"role,omitempty"`          // "IMP" typically, or empty
	CurrentFocus string   `json:"current_focus,omitempty"` // Focused container ID (SHIP-*, CON-*, INV-*, TOME-*)
	CreatedAt    string   `json:"created_at"`
}

type StateConfig struct {
	ActiveMissionID  string `json:"active_mission_id"`
	CurrentHandoffID string `json:"current_handoff_id"`
	Role             string `json:"role,omitempty"`          // "ORC" for global orchestrator
	CurrentFocus     string `json:"current_focus,omitempty"` // Focused container ID (SHIP-*, CON-*, INV-*, TOME-*)
	LastUpdated      string `json:"last_updated"`
}

// LoadConfig reads config.json from directory
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

	// Validate type matches populated fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// SaveConfig writes config.json to directory
func SaveConfig(dir string, cfg *Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

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

// Validate ensures config structure matches type
func (c *Config) Validate() error {
	switch c.Type {
	case TypeMission:
		if c.Mission == nil {
			return fmt.Errorf("type 'mission' requires mission field")
		}
		if c.Grove != nil || c.State != nil {
			return fmt.Errorf("type 'mission' should only have mission field")
		}
	case TypeGrove:
		if c.Grove == nil {
			return fmt.Errorf("type 'grove' requires grove field")
		}
		if c.Mission != nil || c.State != nil {
			return fmt.Errorf("type 'grove' should only have grove field")
		}
	case TypeGlobal:
		if c.State == nil {
			return fmt.Errorf("type 'global' requires state field")
		}
		if c.Mission != nil || c.Grove != nil {
			return fmt.Errorf("type 'global' should only have state field")
		}
	default:
		return fmt.Errorf("invalid type: %s", c.Type)
	}
	return nil
}

// LoadConfigWithFallback loads config from .orc/config.json
// Name kept for compatibility with callers, but no longer has fallback behavior
func LoadConfigWithFallback(dir string) (*Config, error) {
	return LoadConfig(dir)
}
