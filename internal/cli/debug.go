package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/context"
)

// DebugCmd returns the debug command
func DebugCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Debug and diagnostic commands",
		Long:  `Tools for debugging ORC context detection and environment setup.`,
	}

	cmd.AddCommand(debugSessionInfoCmd())
	cmd.AddCommand(debugValidateContextCmd())

	return cmd
}

func debugSessionInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "session-info",
		Short: "Show current session context information",
		Long: `Display detailed information about the current ORC context detection.

Shows:
- Current working directory
- Mission context (.orc/config.json)
- Grove context (.orc/config.json with grove config)
- Environment variables (TMUX, etc.)

Useful for debugging context detection issues.

Examples:
  orc debug session-info`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}

			fmt.Printf("\n=== ORC Session Context ===\n\n")

			// Current directory
			fmt.Printf("Current Directory:\n")
			fmt.Printf("  %s\n\n", cwd)

			// Check for mission config
			fmt.Printf("Mission Context (.orc/config.json):\n")
			missionCtx, err := context.DetectMissionContext()
			if err == nil && missionCtx != nil {
				configPath := filepath.Join(missionCtx.WorkspacePath, ".orc", "config.json")
				fmt.Printf("  ✓ Found: %s\n", configPath)
				fmt.Printf("  Mission ID: %s\n", missionCtx.MissionID)
				fmt.Printf("  Workspace: %s\n\n", missionCtx.WorkspacePath)

				// Read and display .orc/config.json content
				data, err := os.ReadFile(configPath)
				if err == nil {
					fmt.Printf("  Content:\n")
					var cfg map[string]interface{}
					if err := json.Unmarshal(data, &cfg); err == nil {
						formatted, _ := json.MarshalIndent(cfg, "    ", "  ")
						fmt.Printf("    %s\n\n", string(formatted))
					} else {
						fmt.Printf("    %s\n\n", string(data))
					}
				}
			} else {
				fmt.Printf("  ✗ Not found (not in a mission context)\n\n")
			}

			// Check for workspace config
			fmt.Printf("Workspace Config (.orc/config.json):\n")
			if missionCtx != nil {
				configPath := filepath.Join(missionCtx.WorkspacePath, ".orc", "config.json")
				if data, err := os.ReadFile(configPath); err == nil {
					fmt.Printf("  ✓ Found: %s\n", configPath)

					var config map[string]interface{}
					if err := json.Unmarshal(data, &config); err == nil {
						fmt.Printf("  Content:\n")
						formatted, _ := json.MarshalIndent(config, "    ", "  ")
						fmt.Printf("    %s\n\n", string(formatted))
					}
				} else {
					fmt.Printf("  ✗ Not found\n\n")
				}
			} else {
				fmt.Printf("  N/A (no mission context)\n\n")
			}

			// Check for grove config
			fmt.Printf("Grove Config (.orc/config.json in current dir):\n")
			localConfigPath := filepath.Join(cwd, ".orc", "config.json")
			if data, err := os.ReadFile(localConfigPath); err == nil {
				fmt.Printf("  ✓ Found: %s\n", localConfigPath)

				var config map[string]interface{}
				if err := json.Unmarshal(data, &config); err == nil {
					fmt.Printf("  Content:\n")
					formatted, _ := json.MarshalIndent(config, "    ", "  ")
					fmt.Printf("    %s\n\n", string(formatted))
				}
			} else {
				fmt.Printf("  ✗ Not found (not in a grove)\n\n")
			}

			// Environment variables
			fmt.Printf("Environment:\n")
			if tmux := os.Getenv("TMUX"); tmux != "" {
				fmt.Printf("  TMUX: %s\n", tmux)
			} else {
				fmt.Printf("  TMUX: (not set - not in TMux session)\n")
			}

			// Context detection result
			fmt.Printf("\nContext Detection Result:\n")
			if missionCtx != nil {
				fmt.Printf("  Context: Mission (mission-specific)\n")
				fmt.Printf("  Mission: %s\n", missionCtx.MissionID)
			} else {
				fmt.Printf("  Context: Master (global orchestrator)\n")
			}

			fmt.Println()

			return nil
		},
	}
}

func debugValidateContextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate-context [directory]",
		Short: "Validate ORC context setup for a directory",
		Long: `Validate that a directory has proper ORC context markers and config.

Checks:
- .orc/config.json exists and is valid JSON
- Mission ID or grove ID is present
- Directory structure is correct

Useful for debugging mission workspace setup and grove creation issues.

Examples:
  orc debug validate-context ~/src/missions/MISSION-001
  orc debug validate-context ~/src/worktrees/test-grove
  orc debug validate-context .`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := args[0]

			// Resolve to absolute path
			absDir, err := filepath.Abs(dir)
			if err != nil {
				return fmt.Errorf("failed to resolve directory path: %w", err)
			}

			fmt.Printf("\n=== Validating ORC Context: %s ===\n\n", absDir)

			// Check if directory exists
			if _, err := os.Stat(absDir); os.IsNotExist(err) {
				return fmt.Errorf("directory does not exist: %s", absDir)
			}

			validationPassed := true

			// Check 1: .orc directory
			fmt.Printf("1. .orc directory\n")
			orcDir := filepath.Join(absDir, ".orc")
			if info, err := os.Stat(orcDir); err == nil {
				if info.IsDir() {
					fmt.Printf("   ✓ Directory exists: %s\n", orcDir)
				} else {
					fmt.Printf("   ✗ .orc exists but is not a directory\n")
					validationPassed = false
				}
			} else {
				fmt.Printf("   ✗ Directory not found: %s\n", orcDir)
				validationPassed = false
			}
			fmt.Println()

			// Check 2: .orc/config.json
			fmt.Printf("2. .orc/config.json\n")
			configPath := filepath.Join(orcDir, "config.json")
			if data, err := os.ReadFile(configPath); err == nil {
				fmt.Printf("   ✓ File exists: %s\n", configPath)

				// Validate JSON
				var cfg map[string]interface{}
				if err := json.Unmarshal(data, &cfg); err == nil {
					fmt.Printf("   ✓ Valid JSON\n")

					// Check for type and relevant ID fields
					if cfgType, ok := cfg["type"].(string); ok {
						fmt.Printf("   ✓ Config type: %s\n", cfgType)
					}
					// Check nested objects for mission_id
					if grove, ok := cfg["grove"].(map[string]interface{}); ok {
						if missionID, ok := grove["mission_id"].(string); ok && missionID != "" {
							fmt.Printf("   ✓ grove.mission_id present: %s\n", missionID)
						}
					} else if mission, ok := cfg["mission"].(map[string]interface{}); ok {
						if missionID, ok := mission["mission_id"].(string); ok && missionID != "" {
							fmt.Printf("   ✓ mission.mission_id present: %s\n", missionID)
						}
					} else if state, ok := cfg["state"].(map[string]interface{}); ok {
						if activeMissionID, ok := state["active_mission_id"].(string); ok && activeMissionID != "" {
							fmt.Printf("   ✓ state.active_mission_id present: %s\n", activeMissionID)
						}
					}
				} else {
					fmt.Printf("   ✗ Invalid JSON: %v\n", err)
					validationPassed = false
				}
			} else {
				fmt.Printf("   ⚠️  File not found: %s (optional for some contexts)\n", configPath)
			}
			fmt.Println()

			// Overall result
			fmt.Printf("=== Validation Result ===\n")
			if validationPassed {
				fmt.Printf("✓ All critical checks passed\n")
				fmt.Printf("  Context appears to be properly configured\n")
			} else {
				fmt.Printf("✗ Some checks failed\n")
				fmt.Printf("  Context may not be properly configured\n")
			}
			fmt.Println()

			return nil
		},
	}
}
