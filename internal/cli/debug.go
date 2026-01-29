package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/config"
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
- Config information (.orc/config.json)
- Workbench context (if IMP)
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

			// Check for config in current directory
			fmt.Printf("Config (.orc/config.json):\n")
			cfg, err := config.LoadConfig(cwd)
			if err == nil {
				configPath := filepath.Join(cwd, ".orc", "config.json")
				fmt.Printf("  Found: %s\n", configPath)
				fmt.Printf("  Role: %s\n", cfg.Role)
				if cfg.WorkbenchID != "" {
					fmt.Printf("  Workbench ID: %s\n", cfg.WorkbenchID)
				}
				if cfg.WorkshopID != "" {
					fmt.Printf("  Workshop ID: %s\n", cfg.WorkshopID)
				}
				fmt.Println()

				// Read and display raw .orc/config.json content
				data, err := os.ReadFile(configPath)
				if err == nil {
					fmt.Printf("  Raw Content:\n")
					var raw map[string]interface{}
					if err := json.Unmarshal(data, &raw); err == nil {
						formatted, _ := json.MarshalIndent(raw, "    ", "  ")
						fmt.Printf("    %s\n\n", string(formatted))
					} else {
						fmt.Printf("    %s\n\n", string(data))
					}
				}
			} else {
				fmt.Printf("  Not found (no .orc/config.json in current directory)\n\n")
			}

			// Check workbench context
			fmt.Printf("Workbench Context:\n")
			workbenchCtx, err := context.DetectWorkbenchContext()
			if err == nil && workbenchCtx != nil {
				fmt.Printf("  Detected IMP context\n")
				fmt.Printf("  Workbench ID: %s\n", workbenchCtx.WorkbenchID)
				fmt.Printf("  Config Path: %s\n\n", workbenchCtx.ConfigPath)
			} else {
				fmt.Printf("  Not in a workbench (Goblin context)\n\n")
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
			if workbenchCtx != nil {
				fmt.Printf("  Context: IMP (workbench-specific)\n")
				fmt.Printf("  Workbench: %s\n", workbenchCtx.WorkbenchID)
			} else {
				fmt.Printf("  Context: Goblin (orchestrator)\n")
				if cfg != nil && cfg.WorkshopID != "" {
					fmt.Printf("  Workshop: %s\n", cfg.WorkshopID)
				}
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
- Role is set (GOBLIN or IMP)
- Workbench ID is present (for IMP)
- Commission ID is present

Useful for debugging workspace setup issues.

Examples:
  orc debug validate-context ~/src/worktrees/test-workbench
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
					fmt.Printf("   OK: Directory exists: %s\n", orcDir)
				} else {
					fmt.Printf("   FAIL: .orc exists but is not a directory\n")
					validationPassed = false
				}
			} else {
				fmt.Printf("   FAIL: Directory not found: %s\n", orcDir)
				validationPassed = false
			}
			fmt.Println()

			// Check 2: .orc/config.json
			fmt.Printf("2. .orc/config.json\n")
			configPath := filepath.Join(orcDir, "config.json")
			if data, err := os.ReadFile(configPath); err == nil {
				fmt.Printf("   OK: File exists: %s\n", configPath)

				// Validate JSON
				var cfg config.Config
				if err := json.Unmarshal(data, &cfg); err == nil {
					fmt.Printf("   OK: Valid JSON\n")

					// Check role
					if cfg.Role != "" {
						fmt.Printf("   OK: Role: %s\n", cfg.Role)
					} else {
						fmt.Printf("   INFO: No role set (defaults to Goblin)\n")
					}

					// Check identity fields based on role
					if cfg.Role == config.RoleIMP {
						if cfg.WorkbenchID != "" {
							fmt.Printf("   OK: Workbench ID: %s\n", cfg.WorkbenchID)
						} else {
							fmt.Printf("   WARN: IMP role but no workbench_id set\n")
						}
					} else if config.IsGoblinRole(cfg.Role) {
						if cfg.WorkshopID != "" {
							fmt.Printf("   OK: Workshop ID: %s\n", cfg.WorkshopID)
						} else {
							fmt.Printf("   INFO: No workshop_id set (global Goblin context)\n")
						}
					}
				} else {
					fmt.Printf("   FAIL: Invalid JSON: %v\n", err)
					validationPassed = false
				}
			} else {
				fmt.Printf("   INFO: File not found: %s (Goblin context assumed)\n", configPath)
			}
			fmt.Println()

			// Overall result
			fmt.Printf("=== Validation Result ===\n")
			if validationPassed {
				fmt.Printf("OK: All critical checks passed\n")
				fmt.Printf("  Context appears to be properly configured\n")
			} else {
				fmt.Printf("FAIL: Some checks failed\n")
				fmt.Printf("  Context may not be properly configured\n")
			}
			fmt.Println()

			return nil
		},
	}
}
