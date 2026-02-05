package app

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// UserInitials is the default user initials for branch naming.
// This is configured per-user and defaults to "ml" for El Presidente.
const UserInitials = "ml"

// GitService provides git operations for workbenches.
type GitService struct{}

// NewGitService creates a new GitService.
func NewGitService() *GitService {
	return &GitService{}
}

// StashDanceResult contains the result of a stash dance operation.
type StashDanceResult struct {
	PreviousBranch string
	CurrentBranch  string
	WasStashed     bool
	StashPopped    bool
}

// StashDance performs the stash-checkout-pop workflow for branch switching.
// This safely switches branches even when there are uncommitted changes.
func (s *GitService) StashDance(workbenchPath, targetBranch string) (*StashDanceResult, error) {
	result := &StashDanceResult{}

	// Get current branch
	currentBranch, err := s.GetCurrentBranch(workbenchPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}
	result.PreviousBranch = currentBranch

	// Check if already on target branch
	if currentBranch == targetBranch {
		result.CurrentBranch = targetBranch
		return result, nil
	}

	// Check if dirty
	dirty, err := s.IsDirty(workbenchPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check dirty state: %w", err)
	}

	// Stash if dirty
	if dirty {
		if err := s.runGitCommand(workbenchPath, "stash", "push", "-m", "orc-stash-dance"); err != nil {
			return nil, fmt.Errorf("failed to stash changes: %w", err)
		}
		result.WasStashed = true
	}

	// Checkout target branch
	if err := s.runGitCommand(workbenchPath, "checkout", targetBranch); err != nil {
		// Try to restore if checkout failed and we stashed
		if result.WasStashed {
			_ = s.runGitCommand(workbenchPath, "stash", "pop")
		}
		return nil, fmt.Errorf("failed to checkout %s: %w", targetBranch, err)
	}
	result.CurrentBranch = targetBranch

	// Pop stash if we stashed
	if result.WasStashed {
		if err := s.runGitCommand(workbenchPath, "stash", "pop"); err != nil {
			// Stash pop failed - likely conflicts
			return result, fmt.Errorf("checkout succeeded but stash pop failed (conflicts?): %w", err)
		}
		result.StashPopped = true
	}

	return result, nil
}

// GetCurrentBranch returns the current branch name.
func (s *GitService) GetCurrentBranch(repoPath string) (string, error) {
	output, err := s.runGitCommandOutput(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// IsDirty checks if the working directory has uncommitted changes.
func (s *GitService) IsDirty(repoPath string) (bool, error) {
	output, err := s.runGitCommandOutput(repoPath, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(output) != "", nil
}

// GetDirtyFileCount returns the number of modified/untracked files.
func (s *GitService) GetDirtyFileCount(repoPath string) (int, error) {
	output, err := s.runGitCommandOutput(repoPath, "status", "--porcelain")
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(output) == "" {
		return 0, nil
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	return len(lines), nil
}

// BranchExists checks if a branch exists.
func (s *GitService) BranchExists(repoPath, branchName string) (bool, error) {
	// rev-parse returns error if branch doesn't exist - that's expected, not an error condition
	verifyErr := s.runGitCommand(repoPath, "rev-parse", "--verify", branchName)
	return verifyErr == nil, nil
}

// CreateBranch creates a new branch from a base branch.
func (s *GitService) CreateBranch(repoPath, branchName, baseBranch string) error {
	// First fetch to ensure we have latest refs
	_ = s.runGitCommand(repoPath, "fetch", "origin", baseBranch)

	// Create the branch from the base
	if err := s.runGitCommand(repoPath, "branch", branchName, "origin/"+baseBranch); err != nil {
		// Try without origin prefix (for local base branches)
		if err2 := s.runGitCommand(repoPath, "branch", branchName, baseBranch); err2 != nil {
			return fmt.Errorf("failed to create branch %s: %w", branchName, err)
		}
	}
	return nil
}

// CreateAndCheckoutBranch creates a new branch and checks it out.
func (s *GitService) CreateAndCheckoutBranch(repoPath, branchName, baseBranch string) error {
	// Create branch (if it doesn't exist)
	exists, err := s.BranchExists(repoPath, branchName)
	if err != nil {
		return err
	}
	if !exists {
		if err := s.CreateBranch(repoPath, branchName, baseBranch); err != nil {
			return err
		}
	}

	// Checkout the branch
	return s.runGitCommand(repoPath, "checkout", branchName)
}

// GetAheadBehind returns how many commits the current branch is ahead/behind the remote.
// Returns 0, 0 if there's no tracking branch (not an error condition).
func (s *GitService) GetAheadBehind(repoPath string) (int, int, error) {
	// Get the tracking branch - if command fails, there's no upstream (not an error)
	output, _ := s.runGitCommandOutput(repoPath, "rev-list", "--left-right", "--count", "@{u}...HEAD")
	if output == "" {
		return 0, 0, nil
	}

	parts := strings.Fields(strings.TrimSpace(output))
	if len(parts) != 2 {
		return 0, 0, nil
	}

	behind, _ := strconv.Atoi(parts[0])
	ahead, _ := strconv.Atoi(parts[1])
	return ahead, behind, nil
}

// GetDefaultBranch returns the default branch name for a repo (usually main or master).
func (s *GitService) GetDefaultBranch(repoPath string) (string, error) {
	// Try to get from remote HEAD
	output, err := s.runGitCommandOutput(repoPath, "symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil {
		// Parse refs/remotes/origin/main -> main
		parts := strings.Split(strings.TrimSpace(output), "/")
		if len(parts) > 0 {
			return parts[len(parts)-1], nil
		}
	}

	// Fallback: check if main exists
	exists, _ := s.BranchExists(repoPath, "origin/main")
	if exists {
		return "main", nil
	}

	// Fallback: check if master exists
	exists, _ = s.BranchExists(repoPath, "origin/master")
	if exists {
		return "master", nil
	}

	return "main", nil // Default to main
}

// GenerateShipmentBranchName generates a branch name for a shipment.
// Format: {initials}/SHIP-{id}-{slug}
func GenerateShipmentBranchName(initials, shipmentID, title string) string {
	slug := generateSlug(title, 30)
	return fmt.Sprintf("%s/%s-%s", initials, shipmentID, slug)
}

// GenerateHomeBranchName generates a home branch name for a workbench.
// Format: {initials}/{name}
func GenerateHomeBranchName(initials, workbenchName string) string {
	return fmt.Sprintf("%s/%s", initials, workbenchName)
}

// generateSlug creates a URL-friendly slug from a title.
func generateSlug(title string, maxLen int) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove non-alphanumeric characters (except hyphens)
	re := regexp.MustCompile(`[^a-z0-9-]`)
	slug = re.ReplaceAllString(slug, "")

	// Collapse multiple hyphens
	re = regexp.MustCompile(`-+`)
	slug = re.ReplaceAllString(slug, "-")

	// Trim hyphens from start/end
	slug = strings.Trim(slug, "-")

	// Truncate to max length
	if len(slug) > maxLen {
		slug = slug[:maxLen]
		// Don't end on a hyphen
		slug = strings.TrimRight(slug, "-")
	}

	return slug
}

// runGitCommand executes a git command and returns an error if it fails.
func (s *GitService) runGitCommand(repoPath string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %s", err, stderr.String())
	}
	return nil
}

// runGitCommandOutput executes a git command and returns the stdout.
func (s *GitService) runGitCommandOutput(repoPath string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}
	return stdout.String(), nil
}
