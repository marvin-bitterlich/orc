package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/config"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

// FocusCmd returns the focus command
func FocusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "focus [container-id]",
		Short: "Set or show the currently focused container",
		Long: `Focus on a specific container (Shipment or Tome).

For IMP (workbench context):
  - Can focus on Shipments (SHIP-xxx) or Tomes (TOME-xxx)
  - Focus stored in workbenches.focused_id

The focused container appears in 'orc prime' output and can be used as default
for other commands.

Examples:
  orc focus SHIP-178    # Focus on a shipment
  orc focus TOME-028    # Focus on a tome
  orc focus --show      # Show current focus
  orc focus --clear     # Clear the current focus`,
		Args: cobra.MaximumNArgs(1),
		RunE: runFocus,
	}
	cmd.Flags().Bool("show", false, "Show current focus without changing it")
	cmd.Flags().Bool("clear", false, "Clear the current focus")
	return cmd
}

func runFocus(cmd *cobra.Command, args []string) error {
	showOnly, _ := cmd.Flags().GetBool("show")
	clearFlag, _ := cmd.Flags().GetBool("clear")

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Load config from cwd only (with Goblin migration if needed)
	cfg, err := MigrateGoblinConfigIfNeeded(cmd.Context(), cwd)
	if err != nil {
		return fmt.Errorf("no ORC config found in current directory")
	}

	// Route based on place_id type
	placeType := config.GetPlaceType(cfg.PlaceID)
	switch placeType {
	case config.PlaceTypeWorkbench:
		// IMP role - use workbench focus
		return runIMPFocus(cmd, args, cfg, showOnly, clearFlag)
	case config.PlaceTypeGatehouse:
		// Goblin role - focus not supported (conclaves removed)
		return fmt.Errorf("focus not supported in gatehouse context (use workbench context)")
	default:
		return fmt.Errorf("focus requires workbench (IMP) context")
	}
}

// runIMPFocus handles focus for IMP role (workbench context)
// IMP can focus on SHIP-xxx or TOME-xxx
func runIMPFocus(_ *cobra.Command, args []string, cfg *config.Config, showOnly, clearFlag bool) error {
	workbenchID := cfg.PlaceID // BENCH-XXX

	if showOnly {
		return showIMPFocus(workbenchID)
	}

	if clearFlag {
		return clearIMPFocus(workbenchID)
	}

	if len(args) == 0 {
		return fmt.Errorf("Usage: orc focus <ID> or orc focus --show or orc focus --clear")
	}

	// Set focus
	containerID := args[0]

	// IMP can focus on SHIP-xxx or TOME-xxx
	if !strings.HasPrefix(containerID, "SHIP-") && !strings.HasPrefix(containerID, "TOME-") {
		return fmt.Errorf("can only focus on Shipments (SHIP-xxx) or Tomes (TOME-xxx), got: %s", containerID)
	}

	containerType, title, err := validateAndGetInfo(containerID)
	if err != nil {
		return err
	}

	return setIMPFocus(workbenchID, containerID, containerType, title)
}

// validateAndGetInfo validates the container ID exists and returns its type and title
func validateAndGetInfo(id string) (containerType string, title string, err error) {
	ctx := context.Background()
	switch {
	case strings.HasPrefix(id, "SHIP-"):
		ship, err := wire.ShipmentService().GetShipment(ctx, id)
		if err != nil {
			return "", "", fmt.Errorf("shipment %s not found", id)
		}
		return "Shipment", ship.Title, nil

	case strings.HasPrefix(id, "TOME-"):
		tome, err := wire.TomeService().GetTome(ctx, id)
		if err != nil {
			return "", "", fmt.Errorf("tome %s not found", id)
		}
		return "Tome", tome.Title, nil

	default:
		return "", "", fmt.Errorf("unknown container type for ID: %s (expected SHIP-* or TOME-*)", id)
	}
}

// showIMPFocus displays the current IMP focus from DB
func showIMPFocus(workbenchID string) error {
	ctx := context.Background()

	focusID, err := wire.WorkbenchService().GetFocusedID(ctx, workbenchID)
	if err != nil {
		return fmt.Errorf("failed to get focus: %w", err)
	}

	if focusID == "" {
		fmt.Println("No focus set")
		fmt.Println("\nSet focus with: orc focus <SHIP-xxx> or orc focus <TOME-xxx>")
		return nil
	}

	containerType, title, err := validateAndGetInfo(focusID)
	if err != nil {
		// Focus is set but container no longer exists - graceful degradation
		fmt.Printf("Focus: %s (container not found - may have been deleted)\n", focusID)
		return nil //nolint:nilerr // intentional: show info even if container deleted
	}

	fmt.Printf("Focus: %s\n", focusID)
	fmt.Printf("  %s: %s\n", containerType, title)
	return nil
}

// setIMPFocus sets the IMP focus in the DB
func setIMPFocus(workbenchID, containerID, containerType, title string) error {
	ctx := context.Background()

	// Update focus in DB
	if err := wire.WorkbenchService().UpdateFocusedID(ctx, workbenchID, containerID); err != nil {
		return fmt.Errorf("failed to set focus: %w", err)
	}

	fmt.Printf("Focused on %s: %s\n", containerType, containerID)
	fmt.Printf("  %s\n", title)

	// Auto-checkout branch and auto-transition status for shipments
	if strings.HasPrefix(containerID, "SHIP-") {
		// Auto-transition shipment status: draft → exploring
		ship, err := wire.ShipmentService().GetShipment(ctx, containerID)
		if err == nil {
			newStatus, err := wire.ShipmentService().TriggerAutoTransition(ctx, containerID, "focus")
			if err == nil && newStatus != "" {
				fmt.Printf("  ✓ Status: %s → %s\n", ship.Status, newStatus)
			}
		}

		if err := autoCheckoutShipmentBranch(workbenchID, containerID); err != nil {
			fmt.Printf("  (branch checkout skipped: %v)\n", err)
		} else {
			fmt.Println("  ✓ Branch checked out")
		}
	}

	fmt.Println("\nRun 'orc prime' to see updated context.")
	return nil
}

// autoCheckoutShipmentBranch checks out the shipment's branch in the workbench
func autoCheckoutShipmentBranch(workbenchID, shipmentID string) error {
	ctx := context.Background()

	// Get shipment to find its branch
	ship, err := wire.ShipmentService().GetShipment(ctx, shipmentID)
	if err != nil {
		return err
	}

	// Shipments should have a branch field
	if ship.Branch == "" {
		return fmt.Errorf("shipment has no branch assigned")
	}

	// Checkout via workbench service (uses stash dance)
	_, err = wire.WorkbenchService().CheckoutBranch(ctx, primary.CheckoutBranchRequest{
		WorkbenchID:  workbenchID,
		TargetBranch: ship.Branch,
	})
	return err
}

// clearIMPFocus clears the IMP focus in DB
func clearIMPFocus(workbenchID string) error {
	ctx := context.Background()

	if err := wire.WorkbenchService().UpdateFocusedID(ctx, workbenchID, ""); err != nil {
		return fmt.Errorf("failed to clear focus: %w", err)
	}

	fmt.Println("Focus cleared")
	return nil
}

// GetCurrentFocus is exported for use by other commands (e.g., prime)
// Returns the focused ID from DB based on place_id context
func GetCurrentFocus(cfg *config.Config) string {
	if cfg == nil || cfg.PlaceID == "" {
		return ""
	}

	ctx := context.Background()

	placeType := config.GetPlaceType(cfg.PlaceID)
	switch placeType {
	case config.PlaceTypeWorkbench:
		// IMP context - use workbench focus
		focusID, err := wire.WorkbenchService().GetFocusedID(ctx, cfg.PlaceID)
		if err != nil {
			return ""
		}
		return focusID
	case config.PlaceTypeGatehouse:
		// Gatehouse context - no focus support (conclaves removed)
		return ""
	}

	return ""
}

// GetFocusInfo returns the type and title for a focus ID, or empty strings if invalid
func GetFocusInfo(focusID string) (containerType, title, status string) {
	if focusID == "" {
		return "", "", ""
	}

	ctx := context.Background()
	switch {
	case strings.HasPrefix(focusID, "SHIP-"):
		if ship, err := wire.ShipmentService().GetShipment(ctx, focusID); err == nil {
			return "Shipment", ship.Title, ship.Status
		}
	case strings.HasPrefix(focusID, "TOME-"):
		if tome, err := wire.TomeService().GetTome(ctx, focusID); err == nil {
			return "Tome", tome.Title, tome.Status
		}
	}
	return "", "", ""
}
