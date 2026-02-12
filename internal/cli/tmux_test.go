package cli

import (
	"context"
	"fmt"
	"testing"
)

// TestTmuxArchiveWorkbenchCmdStructure verifies the archive-workbench subcommand
// is registered under the tmux command with the correct metadata.
func TestTmuxArchiveWorkbenchCmdStructure(t *testing.T) {
	tmux := TmuxCmd()

	// Find archive-workbench subcommand
	var found bool
	for _, sub := range tmux.Commands() {
		if sub.Use == "archive-workbench" {
			found = true
			if sub.Short == "" {
				t.Error("archive-workbench command should have a Short description")
			}
			break
		}
	}
	if !found {
		t.Fatal("archive-workbench subcommand not registered under tmux")
	}
}

// TestArchiveWorkbenchRunE_GetwdError verifies that a getwd failure
// is surfaced with a clear error message.
func TestArchiveWorkbenchRunE_GetwdError(t *testing.T) {
	failGetwd := func() (string, error) {
		return "", fmt.Errorf("simulated getwd failure")
	}

	err := archiveWorkbenchRunE(context.Background(), failGetwd, nil)
	if err == nil {
		t.Fatal("expected error when getwd fails")
	}
	want := "failed to get working directory"
	if err.Error()[:len(want)] != want {
		t.Errorf("error = %q, want prefix %q", err.Error(), want)
	}
}
