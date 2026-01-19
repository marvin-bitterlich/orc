// Package app contains the application layer - service implementations and effect execution.
package app

import (
	"context"
	"fmt"
	"os"

	"github.com/example/orc/internal/core/effects"
	"github.com/example/orc/internal/models"
	"github.com/example/orc/internal/tmux"
)

// EffectExecutor interprets and executes effects.
// This is the "Imperative Shell" - the only place I/O happens.
type EffectExecutor interface {
	Execute(ctx context.Context, effs []effects.Effect) error
}

// DefaultEffectExecutor implements EffectExecutor with real I/O.
type DefaultEffectExecutor struct{}

// NewEffectExecutor creates a new DefaultEffectExecutor.
func NewEffectExecutor() *DefaultEffectExecutor {
	return &DefaultEffectExecutor{}
}

// Execute processes a slice of effects, executing each in sequence.
func (e *DefaultEffectExecutor) Execute(ctx context.Context, effs []effects.Effect) error {
	for _, eff := range effs {
		if err := e.executeOne(ctx, eff); err != nil {
			return fmt.Errorf("failed to execute %s effect: %w", eff.EffectType(), err)
		}
	}
	return nil
}

func (e *DefaultEffectExecutor) executeOne(ctx context.Context, eff effects.Effect) error {
	switch typed := eff.(type) {
	case effects.FileEffect:
		return e.executeFile(ctx, typed)
	case effects.PersistEffect:
		return e.executePersist(ctx, typed)
	case effects.TMuxEffect:
		return e.executeTMux(ctx, typed)
	case effects.CompositeEffect:
		return e.Execute(ctx, typed.Effects)
	case effects.NoEffect:
		return nil
	case effects.LogEffect:
		fmt.Printf("[%s] %s\n", typed.Level, typed.Message)
		return nil
	default:
		return fmt.Errorf("unknown effect type: %T", eff)
	}
}

func (e *DefaultEffectExecutor) executeFile(ctx context.Context, eff effects.FileEffect) error {
	switch eff.Operation {
	case "mkdir":
		return os.MkdirAll(eff.Path, os.FileMode(eff.Mode))
	case "write":
		return os.WriteFile(eff.Path, eff.Content, os.FileMode(eff.Mode))
	case "read":
		// Read operations are typically used in planning, not execution
		return nil
	case "exists":
		// Existence checks are typically used in planning, not execution
		return nil
	default:
		return fmt.Errorf("unknown file operation: %s", eff.Operation)
	}
}

func (e *DefaultEffectExecutor) executePersist(ctx context.Context, eff effects.PersistEffect) error {
	switch eff.Entity {
	case "grove":
		return e.executeGroveOp(ctx, eff)
	case "mission":
		return e.executeMissionOp(ctx, eff)
	default:
		return fmt.Errorf("unknown entity: %s", eff.Entity)
	}
}

func (e *DefaultEffectExecutor) executeGroveOp(ctx context.Context, eff effects.PersistEffect) error {
	switch eff.Operation {
	case "update":
		data, ok := eff.Data.(map[string]string)
		if !ok {
			return fmt.Errorf("invalid grove update data type: %T", eff.Data)
		}
		if path, exists := data["path"]; exists {
			return models.UpdateGrovePath(data["id"], path)
		}
		return nil
	default:
		return fmt.Errorf("unknown grove operation: %s", eff.Operation)
	}
}

func (e *DefaultEffectExecutor) executeMissionOp(ctx context.Context, eff effects.PersistEffect) error {
	switch eff.Operation {
	case "update_status":
		data, ok := eff.Data.(map[string]string)
		if !ok {
			return fmt.Errorf("invalid mission update data type: %T", eff.Data)
		}
		return models.UpdateMissionStatus(data["id"], data["status"])
	default:
		return fmt.Errorf("unknown mission operation: %s", eff.Operation)
	}
}

func (e *DefaultEffectExecutor) executeTMux(ctx context.Context, eff effects.TMuxEffect) error {
	switch eff.Operation {
	case "new_session":
		workingDir := eff.Command
		if workingDir == "" {
			workingDir = "."
		}
		_, err := tmux.NewSession(eff.SessionName, workingDir)
		return err
	case "new_window":
		// TMux window creation requires an existing session
		// This is handled through the session object in the current tmux package
		// For now, we'll skip standalone window creation
		return nil
	case "send_keys":
		// Send keys would require session lookup
		return nil
	default:
		return fmt.Errorf("unknown tmux operation: %s", eff.Operation)
	}
}
