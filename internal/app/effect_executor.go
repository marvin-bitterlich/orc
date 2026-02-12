// Package app contains the application layer - service implementations and effect execution.
package app

import (
	"context"
	"fmt"
	"os"

	"github.com/example/orc/internal/core/effects"
	"github.com/example/orc/internal/ports/secondary"
)

// EffectExecutor interprets and executes effects.
// This is the "Imperative Shell" - the only place I/O happens.
type EffectExecutor interface {
	Execute(ctx context.Context, effs []effects.Effect) error
}

// DefaultEffectExecutor implements EffectExecutor with real I/O.
type DefaultEffectExecutor struct {
	commissionRepo   secondary.CommissionRepository
	tmuxAdapter      secondary.TMuxAdapter
	workspaceAdapter secondary.WorkspaceAdapter
}

// NewEffectExecutor creates a new DefaultEffectExecutor with injected dependencies.
func NewEffectExecutor(
	commissionRepo secondary.CommissionRepository,
	tmuxAdapter secondary.TMuxAdapter,
	workspaceAdapter secondary.WorkspaceAdapter,
) *DefaultEffectExecutor {
	return &DefaultEffectExecutor{
		commissionRepo:   commissionRepo,
		tmuxAdapter:      tmuxAdapter,
		workspaceAdapter: workspaceAdapter,
	}
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
	case effects.GitEffect:
		return e.executeGit(ctx, typed)
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
	case "commission":
		return e.executeCommissionOp(ctx, eff)
	default:
		return fmt.Errorf("unknown entity: %s", eff.Entity)
	}
}

func (e *DefaultEffectExecutor) executeCommissionOp(ctx context.Context, eff effects.PersistEffect) error {
	switch eff.Operation {
	case "update_status":
		data, ok := eff.Data.(map[string]string)
		if !ok {
			return fmt.Errorf("invalid commission update data type: %T", eff.Data)
		}
		return e.commissionRepo.Update(ctx, &secondary.CommissionRecord{
			ID:     data["id"],
			Status: data["status"],
		})
	default:
		return fmt.Errorf("unknown commission operation: %s", eff.Operation)
	}
}

func (e *DefaultEffectExecutor) executeTMux(ctx context.Context, eff effects.TMuxEffect) error {
	// TMux lifecycle operations (new_session, new_window) removed - now handled by gotmux
	switch eff.Operation {
	case "send_keys":
		// Send keys would require session lookup
		return nil
	default:
		return fmt.Errorf("unknown tmux operation: %s", eff.Operation)
	}
}

func (e *DefaultEffectExecutor) executeGit(ctx context.Context, eff effects.GitEffect) error {
	switch eff.Operation {
	case "worktree_add":
		// Args[0] = branchName, Args[1] = targetPath
		if len(eff.Args) < 2 {
			return fmt.Errorf("worktree_add requires branchName and targetPath in Args")
		}
		branchName := eff.Args[0]
		targetPath := eff.Args[1]
		return e.workspaceAdapter.CreateWorktree(ctx, eff.RepoPath, branchName, targetPath)
	default:
		return fmt.Errorf("unsupported git operation: %s", eff.Operation)
	}
}
