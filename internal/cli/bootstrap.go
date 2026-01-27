// Package cli provides CLI commands for the ORC application.
package cli

import (
	"github.com/example/orc/internal/wire"
)

// ApplyGlobalBindings sets up ORC's global tmux key bindings.
// Safe to call repeatedly (idempotent). Silently ignores errors (tmux may not be running).
// This should be called on every orc command invocation via PersistentPreRun.
func ApplyGlobalBindings() {
	wire.ApplyGlobalTMuxBindings()
}
