package config

import (
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

func TestLoadConfig_BackwardCompatibility(t *testing.T) {
	// Test that loading an old IMP config format works and migrates to place_id
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

	// Old IMP config with role + workbench_id
	oldConfig := `{"version":"1.0","role":"IMP","workbench_id":"BENCH-001"}`
	configPath := filepath.Join(orcDir, "config.json")
	if err := os.WriteFile(configPath, []byte(oldConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Should load without error and migrate to place_id format
	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed with old format: %v", err)
	}

	// Core fields should be populated
	if cfg.Version != "1.0" {
		t.Errorf("Version = %q, want 1.0", cfg.Version)
	}
	// IMP configs should be migrated to place_id
	if cfg.PlaceID != "BENCH-001" {
		t.Errorf("PlaceID = %q, want BENCH-001", cfg.PlaceID)
	}
}

func TestGetPlaceType(t *testing.T) {
	tests := []struct {
		placeID  string
		expected string
	}{
		{"BENCH-001", PlaceTypeWorkbench},
		{"BENCH-014", PlaceTypeWorkbench},
		{"GATE-001", ""},
		{"", ""},
		{"WORK-001", ""},
		{"COMM-001", ""},
		{"SHIP-001", ""},
		{"BEN", ""},
	}

	for _, tt := range tests {
		t.Run(tt.placeID, func(t *testing.T) {
			result := GetPlaceType(tt.placeID)
			if result != tt.expected {
				t.Errorf("GetPlaceType(%q) = %q, want %q", tt.placeID, result, tt.expected)
			}
		})
	}
}

func TestGetRoleFromPlaceID(t *testing.T) {
	tests := []struct {
		placeID  string
		expected string
	}{
		{"BENCH-001", RoleIMP},
		{"GATE-001", ""},
		{"", ""},
		{"WORK-001", ""},
	}

	for _, tt := range tests {
		t.Run(tt.placeID, func(t *testing.T) {
			result := GetRoleFromPlaceID(tt.placeID)
			if result != tt.expected {
				t.Errorf("GetRoleFromPlaceID(%q) = %q, want %q", tt.placeID, result, tt.expected)
			}
		})
	}
}
