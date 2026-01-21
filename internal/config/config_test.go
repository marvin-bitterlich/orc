package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultWorkspacePath(t *testing.T) {
	path, err := DefaultWorkspacePath("MISSION-001")
	if err != nil {
		t.Fatalf("DefaultWorkspacePath failed: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "src", "missions", "MISSION-001")

	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}
