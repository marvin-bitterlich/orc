// Package cli provides CLI commands for the ORC application.
package cli

import (
	gocontext "context"

	"github.com/example/orc/internal/agent"
	orccontext "github.com/example/orc/internal/context"
	"github.com/example/orc/internal/wire"
)

// globalActorID stores the detected actor ID for the current CLI invocation.
// Set once at startup by DetectAndStoreActor().
var globalActorID string

// DetectAndStoreActor detects the current actor identity and stores it globally.
// Should be called once at CLI startup in PersistentPreRun.
func DetectAndStoreActor() {
	identity, err := agent.GetCurrentAgentID()
	if err != nil {
		// Default to GOBLIN on error
		globalActorID = "GOBLIN"
		return
	}
	globalActorID = identity.FullID
}

// GetActorID returns the stored actor ID from CLI startup.
// Returns empty string if DetectAndStoreActor() was not called.
func GetActorID() string {
	return globalActorID
}

// NewContext creates a context.Background() with the current actor ID embedded.
// CLI commands should use this instead of context.Background() directly.
func NewContext() gocontext.Context {
	ctx := gocontext.Background()
	if globalActorID != "" {
		return orccontext.WithActorID(ctx, globalActorID)
	}
	return ctx
}

// ApplyGlobalBindings sets up ORC's global tmux key bindings.
// Safe to call repeatedly (idempotent). Silently ignores errors (tmux may not be running).
// This should be called on every orc command invocation via PersistentPreRun.
func ApplyGlobalBindings() {
	wire.ApplyGlobalTMuxBindings()
}
