package context

import (
	gocontext "context"

	"github.com/example/orc/internal/ctxutil"
)

// WithActorID returns a context with the actor ID embedded.
// Actor ID should be the place ID (BENCH-xxx) or "ORC" for orchestrator.
// This is a convenience wrapper around ctxutil.WithActorID.
func WithActorID(ctx gocontext.Context, actorID string) gocontext.Context {
	return ctxutil.WithActorID(ctx, actorID)
}

// ActorFromContext returns the actor ID from context, or empty string if not set.
// This is a convenience wrapper around ctxutil.ActorFromContext.
func ActorFromContext(ctx gocontext.Context) string {
	return ctxutil.ActorFromContext(ctx)
}
