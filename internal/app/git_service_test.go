package app

import (
	"testing"
)

// GitService tests are intentionally minimal because GitService calls os/exec directly
// without dependency injection, making it difficult to mock git operations.
// Future work: Either refactor GitService to use an injectable executor or
// create integration tests with temporary git repos.

func TestGitService_GenerateShipmentBranchName(t *testing.T) {
	// Test the pure function that generates branch names
	name := GenerateShipmentBranchName("ml", "SHIP-001", "test feature")

	// Should have format: ml/SHIP-001-test-feature
	if name == "" {
		t.Error("expected non-empty branch name")
	}
}

func TestGitService_GenerateHomeBranchName(t *testing.T) {
	// Test the pure function that generates home branch names
	name := GenerateHomeBranchName("ml", "orc-014")

	// Should have format: ml/{name} (e.g., ml/orc-014)
	expected := "ml/orc-014"
	if name != expected {
		t.Errorf("expected branch name '%s', got '%s'", expected, name)
	}
}

func TestGitService_NewGitService(t *testing.T) {
	// Verify we can create a GitService
	service := NewGitService()
	if service == nil {
		t.Error("expected non-nil service")
	}
}
