package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/config"
	"github.com/example/orc/internal/wire"
)

// StatusCmd returns the status command
func StatusCmd() *cobra.Command {
	var showHandoff bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current work context from config.json",
		Long: `Display the current work context based on .orc/config.json in current directory:
- Active commission, shipments, and tasks
- Current focus (if any)
- Role (GOBLIN or IMP)

This provides a focused view of "where am I right now?"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config from current directory
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			cfg, cfgErr := MigrateGoblinConfigIfNeeded(cmd.Context(), cwd)
			if cfgErr != nil {
				// No config - show minimal status
				fmt.Println("❓ ORC Status - No Context")
				fmt.Println()
				fmt.Println("No .orc/config.json found in current directory.")
				fmt.Println("This is a Goblin context (no workbench configured).")
				fmt.Println()
				fmt.Println("Run `orc commission list` to see available commissions.")
				return nil //nolint:nilerr // Missing config is intentionally not an error
			}

			// Determine role from place_id
			role := config.GetRoleFromPlaceID(cfg.PlaceID)
			if role == "" {
				role = config.RoleGoblin // default
			}

			// Show status based on role
			if role == config.RoleGoblin {
				fmt.Println("👺 ORC Status - Goblin Context")
			} else if role == config.RoleIMP {
				fmt.Println("👹 ORC Status - IMP Context")
				if config.IsWorkbench(cfg.PlaceID) {
					fmt.Printf("  🔧 Workbench: %s\n", cfg.PlaceID)
				}
			}
			fmt.Println()

			// Display current focus if set (read from DB for IMP context)
			focusID := GetCurrentFocus(cfg)
			if focusID != "" {
				containerType, title, status := GetFocusInfo(focusID)
				if containerType != "" {
					fmt.Printf("🎯 Focus: %s - %s [%s]\n", focusID, title, status)
					fmt.Printf("   (%s)\n", containerType)
				} else {
					fmt.Printf("🎯 Focus: %s (container not found)\n", focusID)
				}
				fmt.Println()
			}

			// If IMP, show workbench-specific info
			if role == config.RoleIMP && config.IsWorkbench(cfg.PlaceID) {
				// Show shipments assigned to this workbench
				shipments, err := wire.ShipmentService().GetShipmentsByWorkbench(context.Background(), cfg.PlaceID)
				if err == nil && len(shipments) > 0 {
					fmt.Println("📦 Assigned Shipments:")
					for _, s := range shipments {
						fmt.Printf("  - %s: %s [%s]\n", s.ID, s.Title, s.Status)
					}
					fmt.Println()
				}
			}

			// Show latest handoff if requested
			if showHandoff {
				handoffs, err := wire.HandoffService().ListHandoffs(context.Background(), 1)
				if err == nil && len(handoffs) > 0 {
					h := handoffs[0]
					fmt.Printf("🤝 Latest Handoff: %s\n", h.ID)
					fmt.Printf("   Created: %s\n", h.CreatedAt)
					fmt.Println()
					fmt.Println("--- HANDOFF NOTE ---")
					fmt.Println(h.HandoffNote)
					fmt.Println("--- END HANDOFF ---")
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&showHandoff, "handoff", "n", false, "Show latest handoff note")

	return cmd
}
