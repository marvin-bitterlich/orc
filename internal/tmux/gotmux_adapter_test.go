package tmux

import (
	"testing"
)

func TestNewGotmuxAdapter(t *testing.T) {
	adapter, err := NewGotmuxAdapter()
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	if adapter == nil {
		t.Fatal("adapter should not be nil")
	}

	if adapter.tmux == nil {
		t.Fatal("adapter.tmux should not be nil")
	}
}

func TestSessionExists(t *testing.T) {
	adapter, err := NewGotmuxAdapter()
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	// Test with non-existent session
	exists := adapter.SessionExists("nonexistent-session-test-12345")
	if exists {
		t.Error("SessionExists should return false for non-existent session")
	}
}

func TestAttachInstructions(t *testing.T) {
	adapter, err := NewGotmuxAdapter()
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	instructions := adapter.AttachInstructions("test-session")
	if instructions == "" {
		t.Error("AttachInstructions should return non-empty string")
	}

	// Check that it contains the session name
	if !contains(instructions, "test-session") {
		t.Error("AttachInstructions should contain the session name")
	}

	// Check that it contains basic tmux instructions
	if !contains(instructions, "tmux attach") {
		t.Error("AttachInstructions should contain 'tmux attach'")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
