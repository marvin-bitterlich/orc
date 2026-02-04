// Package ctxutil provides context utilities that can be safely imported anywhere.
// This package has no internal dependencies to avoid import cycles.
package ctxutil

import "context"

// ActorKey is the context key for actor ID.
// Exported so it can be used consistently across packages.
type ActorKey struct{}

// WithActorID returns a context with the actor ID embedded.
func WithActorID(ctx context.Context, actorID string) context.Context {
	return context.WithValue(ctx, ActorKey{}, actorID)
}

// ActorFromContext returns the actor ID from context, or empty string if not set.
func ActorFromContext(ctx context.Context) string {
	if v := ctx.Value(ActorKey{}); v != nil {
		return v.(string)
	}
	return ""
}
