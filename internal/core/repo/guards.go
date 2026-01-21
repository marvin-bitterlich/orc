// Package repo contains the pure business logic for repository operations.
// Guards are pure functions that evaluate preconditions without side effects.
package repo

import (
	"fmt"
	"strings"
)

// GuardResult represents the outcome of a guard evaluation.
type GuardResult struct {
	Allowed bool
	Reason  string
}

// Error converts the guard result to an error if not allowed.
func (r GuardResult) Error() error {
	if r.Allowed {
		return nil
	}
	return fmt.Errorf("%s", r.Reason)
}

// CreateRepoContext provides context for repository creation guards.
type CreateRepoContext struct {
	Name       string
	NameExists bool // true if a repo with this name already exists
}

// ArchiveRepoContext provides context for repository archive guards.
type ArchiveRepoContext struct {
	RepoID string
	Status string
}

// RestoreRepoContext provides context for repository restore guards.
type RestoreRepoContext struct {
	RepoID string
	Status string
}

// DeleteRepoContext provides context for repository deletion guards.
type DeleteRepoContext struct {
	RepoID       string
	HasActivePRs bool
}

// CanCreateRepo evaluates whether a repository can be created.
// Rules:
// - Name must not be empty
// - Name must be unique
func CanCreateRepo(ctx CreateRepoContext) GuardResult {
	// Rule 1: Name must not be empty
	if strings.TrimSpace(ctx.Name) == "" {
		return GuardResult{
			Allowed: false,
			Reason:  "repository name cannot be empty",
		}
	}

	// Rule 2: Name must be unique
	if ctx.NameExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("repository with name %q already exists", ctx.Name),
		}
	}

	return GuardResult{Allowed: true}
}

// CanArchiveRepo evaluates whether a repository can be archived.
// Rules:
// - Status must be "active"
func CanArchiveRepo(ctx ArchiveRepoContext) GuardResult {
	if ctx.Status != "active" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only archive active repositories (current status: %s)", ctx.Status),
		}
	}

	return GuardResult{Allowed: true}
}

// CanRestoreRepo evaluates whether a repository can be restored.
// Rules:
// - Status must be "archived"
func CanRestoreRepo(ctx RestoreRepoContext) GuardResult {
	if ctx.Status != "archived" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only restore archived repositories (current status: %s)", ctx.Status),
		}
	}

	return GuardResult{Allowed: true}
}

// CanDeleteRepo evaluates whether a repository can be deleted.
// Rules:
// - No active PRs can reference this repository
func CanDeleteRepo(ctx DeleteRepoContext) GuardResult {
	if ctx.HasActivePRs {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("cannot delete repository %s with active pull requests", ctx.RepoID),
		}
	}

	return GuardResult{Allowed: true}
}
