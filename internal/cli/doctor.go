package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/version"
)

// DoctorCmd returns the doctor command for environment validation
func DoctorCmd() *cobra.Command {
	var quiet bool

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Validate ORC environment and Claude Code configuration",
		Long: `Comprehensive environment health check for ORC.

Validates:
- Claude Code workspace trust configuration (CRITICAL)
- Directory structure (worktrees, missions)
- Database existence and integrity
- Binary installation and PATH

Provides actionable error messages with fix instructions for any issues found.

Examples:
  orc doctor              # Run full health check
  orc doctor --quiet      # Exit code only (0=healthy, 1=issues)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !quiet {
				fmt.Print("\n=== ORC Environment Health Check ===\n\n")
			}

			issuesFound := false

			// Check 1: Claude Code Settings (CRITICAL)
			if !quiet {
				fmt.Println("1. Claude Code Settings (CRITICAL)")
			}
			if err := checkClaudeSettings(quiet); err != nil {
				issuesFound = true
				if !quiet {
					fmt.Printf("   %s\n\n", err)
				}
			} else if !quiet {
				fmt.Println("   ✓ ~/.claude/settings.json exists")
				fmt.Println("   ✓ Valid JSON structure")
				fmt.Println("   ✓ permissions.additionalDirectories configured")
				fmt.Println("   ✓ ~/src/worktrees in trusted directories")
				fmt.Println("   ✓ ~/src/missions in trusted directories")
				fmt.Println()
			}

			// Check 2: Directory Structure
			if !quiet {
				fmt.Println("2. Directory Structure")
			}
			if err := checkDirectories(quiet); err != nil {
				issuesFound = true
				if !quiet {
					fmt.Printf("   %s\n\n", err)
				}
			}

			// Check 3: Database
			if !quiet {
				fmt.Println("3. Database")
			}
			if err := checkDatabase(quiet); err != nil {
				issuesFound = true
				if !quiet {
					fmt.Printf("   %s\n\n", err)
				}
			}

			// Check 4: Binary Installation
			if !quiet {
				fmt.Println("4. Binary Installation")
			}
			if err := checkBinary(quiet); err != nil {
				issuesFound = true
				if !quiet {
					fmt.Printf("   %s\n\n", err)
				}
			}

			// Overall status
			if !quiet {
				if issuesFound {
					fmt.Println("=== Overall Status: CRITICAL ISSUES FOUND ===")
					fmt.Println("Fix the above errors before using ORC.")
				} else {
					fmt.Println("=== Overall Status: HEALTHY ===")
					fmt.Println("All critical checks passed. ORC is ready to use.")
				}
				fmt.Println()
			}

			if issuesFound {
				return fmt.Errorf("environment validation failed")
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode - exit code only")

	return cmd
}

func checkClaudeSettings(quiet bool) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("✗ Failed to get home directory: %w", err)
	}

	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")

	// Check existence
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return fmt.Errorf(`✗ ~/.claude/settings.json NOT FOUND

   ERROR: Claude Code workspace trust not configured

   This is REQUIRED for ORC to function. Without it, Claude instances
   in groves and missions will fail with permission errors.

   FIX: Create ~/.claude/settings.json with:

   cat > ~/.claude/settings.json <<'EOF'
   {
     "permissions": {
       "additionalDirectories": [
         "~/src/worktrees",
         "~/src/missions"
       ]
     }
   }
   EOF`)
	}

	// Read and validate JSON
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf("✗ Failed to read ~/.claude/settings.json: %w", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("✗ ~/.claude/settings.json is not valid JSON: %w", err)
	}

	// Check permissions.additionalDirectories
	permissions, ok := settings["permissions"].(map[string]interface{})
	if !ok {
		return fmt.Errorf(`✗ permissions.additionalDirectories NOT configured

   FIX: Add to ~/.claude/settings.json:

   {
     "permissions": {
       "additionalDirectories": [
         "~/src/worktrees",
         "~/src/missions"
       ]
     }
   }`)
	}

	additionalDirs, ok := permissions["additionalDirectories"].([]interface{})
	if !ok {
		return fmt.Errorf(`✗ permissions.additionalDirectories NOT configured

   FIX: Add to permissions object in ~/.claude/settings.json:

   "additionalDirectories": [
     "~/src/worktrees",
     "~/src/missions"
   ]`)
	}

	// Verify required directories
	foundDirs := make(map[string]bool)
	for _, dir := range additionalDirs {
		if dirStr, ok := dir.(string); ok {
			foundDirs[dirStr] = true
		}
	}

	missingWorktrees := !foundDirs["~/src/worktrees"]
	missingMissions := !foundDirs["~/src/missions"]

	if missingWorktrees || missingMissions {
		msg := "✗ Missing required directories:\n"
		if missingWorktrees {
			msg += "     - ~/src/worktrees\n"
		}
		if missingMissions {
			msg += "     - ~/src/missions\n"
		}
		msg += "\n   FIX: Add missing directories to additionalDirectories array"
		return errors.New(msg)
	}

	return nil
}

func checkDirectories(quiet bool) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("✗ Failed to get home directory: %w", err)
	}

	worktreesPath := filepath.Join(homeDir, "src", "worktrees")
	missionsPath := filepath.Join(homeDir, "src", "missions")

	worktreesExists := true
	if _, err := os.Stat(worktreesPath); os.IsNotExist(err) {
		worktreesExists = false
	}

	missionsExists := true
	if _, err := os.Stat(missionsPath); os.IsNotExist(err) {
		missionsExists = false
	}

	if !quiet {
		if worktreesExists {
			// Count groves
			entries, _ := os.ReadDir(worktreesPath)
			groveCount := len(entries)
			fmt.Printf("   ✓ ~/src/worktrees exists (%d groves)\n", groveCount)
		} else {
			fmt.Println("   ⚠️  ~/src/worktrees does not exist (will be created on first grove)")
		}

		if missionsExists {
			// Count missions
			entries, _ := os.ReadDir(missionsPath)
			missionCount := len(entries)
			fmt.Printf("   ✓ ~/src/missions exists (%d missions)\n", missionCount)
		} else {
			fmt.Println("   ⚠️  ~/src/missions does not exist (will be created on first mission)")
		}
		fmt.Println()
	}

	// These are warnings, not errors - directories will be created on demand
	return nil
}

func checkDatabase(quiet bool) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("✗ Failed to get home directory: %w", err)
	}

	dbPath := filepath.Join(homeDir, ".orc", "orc.db")

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf(`✗ ~/.orc/orc.db NOT FOUND

   ERROR: ORC database not initialized

   FIX: Run 'orc init' to initialize the database`)
	}

	if !quiet {
		fmt.Println("   ✓ ~/.orc/orc.db exists")

		// Get file size
		info, _ := os.Stat(dbPath)
		sizeKB := info.Size() / 1024
		fmt.Printf("   ✓ Database size: %d KB\n", sizeKB)
		fmt.Println()
	}

	return nil
}

func checkBinary(quiet bool) error {
	// Check if orc is in PATH
	orcPath, err := exec.LookPath("orc")
	if err != nil {
		return fmt.Errorf(`✗ 'orc' binary not found in PATH

   ERROR: ORC is not installed or not in your PATH

   FIX: Ensure 'go install' completed and ~/go/bin is in PATH`)
	}

	if !quiet {
		fmt.Printf("   ✓ orc binary: %s\n", orcPath)
		fmt.Println("   ✓ In PATH: yes")
		fmt.Printf("   ✓ Version: %s\n", version.String())
	}

	// Check for stale local binary if we're in the ORC repo
	if isInOrcRepo() {
		if !quiet {
			fmt.Println()
			fmt.Println("   Checking local development binary...")
		}
		if err := checkLocalBinaryFreshness(quiet); err != nil {
			if !quiet {
				fmt.Printf("   %s\n", err)
			}
			// This is a warning, not an error
		}
	}

	if !quiet {
		fmt.Println()
	}

	return nil
}

// isInOrcRepo checks if we're in the ORC repository
func isInOrcRepo() bool {
	// Check for go.mod with the ORC module
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "module github.com/example/orc")
}

// checkLocalBinaryFreshness warns if ./orc is stale compared to source
func checkLocalBinaryFreshness(quiet bool) error {
	// Check if ./orc exists
	localBinary := "./orc"
	info, err := os.Stat(localBinary)
	if os.IsNotExist(err) {
		if !quiet {
			fmt.Println("   ⚠️  No local ./orc binary found")
			fmt.Println("      Run 'make dev' to build for development")
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("⚠️  Error checking ./orc: %w", err)
	}

	// Get the version from the local binary
	cmd := exec.Command(localBinary, "--version")
	output, err := cmd.Output()
	if err != nil {
		if !quiet {
			fmt.Println("   ⚠️  Local ./orc exists but failed to get version")
			fmt.Println("      May be corrupted - run 'make dev' to rebuild")
		}
		return nil
	}

	localVersion := strings.TrimSpace(string(output))

	// Get the current git commit
	gitCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	gitOutput, err := gitCmd.Output()
	if err != nil {
		// Can't check git, skip freshness check
		if !quiet {
			fmt.Printf("   ✓ Local ./orc: %s\n", localVersion)
		}
		return nil
	}
	currentCommit := strings.TrimSpace(string(gitOutput))

	// Check if the local binary was built from the current commit
	if strings.Contains(localVersion, currentCommit) {
		if !quiet {
			fmt.Printf("   ✓ Local ./orc is fresh (commit: %s)\n", currentCommit)
			fmt.Printf("      Built: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
		}
	} else {
		// Extract commit from version string if possible
		if !quiet {
			fmt.Printf("   ⚠️  Local ./orc may be STALE\n")
			fmt.Printf("      Binary:  %s\n", localVersion)
			fmt.Printf("      Current: commit %s\n", currentCommit)
			fmt.Println()
			fmt.Println("      FIX: Run 'make dev' to rebuild from current source")
		}
	}

	return nil
}
