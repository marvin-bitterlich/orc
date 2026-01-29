package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultWorkspacePath(t *testing.T) {
	path, err := DefaultWorkspacePath("COMM-001")
	if err != nil {
		t.Fatalf("DefaultWorkspacePath failed: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "src", "commissions", "COMM-001")

	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestParseWorkshopIDFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "standard workshop path",
			path:     "/Users/test/.orc/ws/WORK-003-myproject/",
			expected: "WORK-003",
		},
		{
			name:     "workshop path without trailing slash",
			path:     "/Users/test/.orc/ws/WORK-001",
			expected: "WORK-001",
		},
		{
			name:     "workshop path with suffix",
			path:     "/home/user/.orc/ws/WORK-042-some-name/subdir",
			expected: "WORK-042",
		},
		{
			name:     "no workshop ID in path",
			path:     "/Users/test/src/worktrees/myproject",
			expected: "",
		},
		{
			name:     "workbench path (not workshop)",
			path:     "/Users/test/src/worktrees/BENCH-014",
			expected: "",
		},
		{
			name:     "empty path",
			path:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseWorkshopIDFromPath(tt.path)
			if result != tt.expected {
				t.Errorf("ParseWorkshopIDFromPath(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestMigrateConfig(t *testing.T) {
	tests := []struct {
		name           string
		initialConfig  map[string]any
		wantOldFocus   string
		wantModified   bool
		wantErr        bool
		checkNewConfig func(t *testing.T, cfg *Config)
	}{
		{
			name: "migrate config with deprecated fields",
			initialConfig: map[string]any{
				"version":       "1.0",
				"role":          "IMP",
				"workbench_id":  "BENCH-014",
				"commission_id": "COMM-001",
				"current_focus": "SHIP-123",
			},
			wantOldFocus: "SHIP-123",
			wantModified: true,
			wantErr:      false,
			checkNewConfig: func(t *testing.T, cfg *Config) {
				if cfg.WorkbenchID != "BENCH-014" {
					t.Errorf("expected workbench_id BENCH-014, got %s", cfg.WorkbenchID)
				}
			},
		},
		{
			name: "already migrated config",
			initialConfig: map[string]any{
				"version":      "1.0",
				"role":         "IMP",
				"workbench_id": "BENCH-014",
			},
			wantOldFocus: "",
			wantModified: false,
			wantErr:      false,
		},
		{
			name: "migrate Goblin config with workshop path",
			initialConfig: map[string]any{
				"version":       "1.0",
				"role":          "GOBLIN",
				"commission_id": "COMM-001",
			},
			wantOldFocus: "",
			wantModified: true,
			wantErr:      false,
			checkNewConfig: func(t *testing.T, cfg *Config) {
				if cfg.Role != "GOBLIN" {
					t.Errorf("expected role GOBLIN, got %s", cfg.Role)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir, err := os.MkdirTemp("", "orc-config-test")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Create .orc directory and config
			orcDir := filepath.Join(tmpDir, ".orc")
			if err := os.MkdirAll(orcDir, 0755); err != nil {
				t.Fatalf("failed to create .orc dir: %v", err)
			}

			configPath := filepath.Join(orcDir, "config.json")
			data, err := json.Marshal(tt.initialConfig)
			if err != nil {
				t.Fatalf("failed to marshal initial config: %v", err)
			}
			if err := os.WriteFile(configPath, data, 0644); err != nil {
				t.Fatalf("failed to write initial config: %v", err)
			}

			// Run migration
			oldFocus, modified, err := MigrateConfig(tmpDir)

			// Check error
			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check old focus
			if oldFocus != tt.wantOldFocus {
				t.Errorf("oldFocus = %q, want %q", oldFocus, tt.wantOldFocus)
			}

			// Check modified
			if modified != tt.wantModified {
				t.Errorf("modified = %v, want %v", modified, tt.wantModified)
			}

			// If we have additional checks
			if tt.checkNewConfig != nil && !tt.wantErr {
				cfg, err := LoadConfig(tmpDir)
				if err != nil {
					t.Fatalf("failed to load config after migration: %v", err)
				}
				tt.checkNewConfig(t, cfg)
			}
		})
	}
}

func TestLoadConfig_BackwardCompatibility(t *testing.T) {
	// Test that loading an old config format works without error
	tmpDir, err := os.MkdirTemp("", "orc-config-compat")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .orc directory and old-format config
	orcDir := filepath.Join(tmpDir, ".orc")
	if err := os.MkdirAll(orcDir, 0755); err != nil {
		t.Fatalf("failed to create .orc dir: %v", err)
	}

	// Old config with deprecated fields (they should be ignored, not cause errors)
	oldConfig := `{"version":"1.0","role":"IMP","workbench_id":"BENCH-001","commission_id":"COMM-001","current_focus":"SHIP-123"}`
	configPath := filepath.Join(orcDir, "config.json")
	if err := os.WriteFile(configPath, []byte(oldConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Should load without error (deprecated fields are ignored)
	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed with old format: %v", err)
	}

	// Core fields should be populated
	if cfg.Version != "1.0" {
		t.Errorf("Version = %q, want 1.0", cfg.Version)
	}
	if cfg.Role != "IMP" {
		t.Errorf("Role = %q, want IMP", cfg.Role)
	}
	if cfg.WorkbenchID != "BENCH-001" {
		t.Errorf("WorkbenchID = %q, want BENCH-001", cfg.WorkbenchID)
	}
}
