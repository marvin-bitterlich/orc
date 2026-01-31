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
		Long: `Focus on a specific container (Conclave, Shipment, or Tome).

For Goblin (workshop context):
  - Can only focus on Conclaves (CON-xxx)
  - Focus stored in workshops.focused_conclave_id

For IMP (workbench context):
  - Can focus on Conclaves (CON-xxx), Shipments (SHIP-xxx), or Tomes (TOME-xxx)
  - Focus stored in workbenches.focused_id

The focused container appears in 'orc prime' output and can be used as default
for other commands.

Examples:
  orc focus CON-007     # Focus on a conclave (works for both roles)
  orc focus SHIP-178    # Focus on a shipment (IMP only)
  orc focus TOME-028    # Focus on a tome (IMP only)
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
		// Goblin role - need to look up workshop from gatehouse
		return runGoblinFocus(cmd, args, cfg, showOnly, clearFlag)
	default:
		return fmt.Errorf("focus requires workbench (IMP) or gatehouse (Goblin) context")
	}
}

// runIMPFocus handles focus for IMP role (workbench context)
// IMP can focus on CON-xxx or SHIP-xxx
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

	// IMP can focus on CON-xxx, SHIP-xxx, or TOME-xxx
	if !strings.HasPrefix(containerID, "CON-") && !strings.HasPrefix(containerID, "SHIP-") && !strings.HasPrefix(containerID, "TOME-") {
		return fmt.Errorf("IMP can only focus on Conclaves (CON-xxx), Shipments (SHIP-xxx), or Tomes (TOME-xxx), got: %s", containerID)
	}

	containerType, title, err := validateAndGetInfo(containerID)
	if err != nil {
		return err
	}

	return setIMPFocus(workbenchID, containerID, containerType, title)
}

// runGoblinFocus handles focus for Goblin role (gatehouse context)
// Goblin can ONLY focus on CON-xxx
// Gatehouse place_id is GATE-XXX, need to look up workshop
func runGoblinFocus(_ *cobra.Command, args []string, cfg *config.Config, showOnly, clearFlag bool) error {
	gatehouseID := cfg.PlaceID // GATE-XXX

	// Look up workshop from gatehouse
	ctx := context.Background()
	gatehouse, err := wire.GatehouseService().GetGatehouse(ctx, gatehouseID)
	if err != nil {
		return fmt.Errorf("failed to get gatehouse: %w", err)
	}
	workshopID := gatehouse.WorkshopID

	if showOnly {
		return showGoblinFocus(workshopID)
	}

	if clearFlag {
		return clearGoblinFocus(workshopID)
	}

	if len(args) == 0 {
		return fmt.Errorf("Usage: orc focus <CON-xxx> or orc focus --show or orc focus --clear")
	}

	containerID := args[0]

	// Goblin can ONLY focus on CON-xxx
	if !strings.HasPrefix(containerID, "CON-") {
		return fmt.Errorf("Goblin can only focus on Conclaves (CON-xxx), got: %s", containerID)
	}

	containerType, title, err := validateAndGetInfo(containerID)
	if err != nil {
		return err
	}

	return setGoblinFocus(workshopID, containerID, containerType, title)
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

	case strings.HasPrefix(id, "CON-"):
		con, err := wire.ConclaveService().GetConclave(ctx, id)
		if err != nil {
			return "", "", fmt.Errorf("conclave %s not found", id)
		}
		return "Conclave", con.Title, nil

	case strings.HasPrefix(id, "TOME-"):
		tome, err := wire.TomeService().GetTome(ctx, id)
		if err != nil {
			return "", "", fmt.Errorf("tome %s not found", id)
		}
		return "Tome", tome.Title, nil

	default:
		return "", "", fmt.Errorf("unknown container type for ID: %s (expected CON-*, SHIP-*, or TOME-*)", id)
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
		fmt.Println("\nSet focus with: orc focus <CON-xxx>, orc focus <SHIP-xxx>, or orc focus <TOME-xxx>")
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

// showGoblinFocus displays the current Goblin focus from DB
func showGoblinFocus(workshopID string) error {
	ctx := context.Background()

	focusID, err := wire.WorkshopService().GetFocusedConclaveID(ctx, workshopID)
	if err != nil {
		return fmt.Errorf("failed to get focus: %w", err)
	}

	if focusID == "" {
		fmt.Println("No focus set")
		fmt.Println("\nSet focus with: orc focus <CON-xxx>")
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

	// Auto-checkout branch for shipments when in a workbench
	if strings.HasPrefix(containerID, "SHIP-") {
		if err := autoCheckoutShipmentBranch(workbenchID, containerID); err != nil {
			fmt.Printf("  (branch checkout skipped: %v)\n", err)
		} else {
			fmt.Println("  ✓ Branch checked out")
		}
	}

	fmt.Println("\nRun 'orc prime' to see updated context.")
	return nil
}

// setGoblinFocus sets the Goblin focus in the DB
func setGoblinFocus(workshopID, containerID, containerType, title string) error {
	ctx := context.Background()

	// Update focus in DB
	if err := wire.WorkshopService().UpdateFocusedConclaveID(ctx, workshopID, containerID); err != nil {
		return fmt.Errorf("failed to set focus: %w", err)
	}

	fmt.Printf("Focused on %s: %s\n", containerType, containerID)
	fmt.Printf("  %s\n", title)

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

// clearGoblinFocus clears the Goblin focus in DB
func clearGoblinFocus(workshopID string) error {
	ctx := context.Background()

	if err := wire.WorkshopService().UpdateFocusedConclaveID(ctx, workshopID, ""); err != nil {
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
		// Goblin context - look up workshop from gatehouse
		gatehouse, err := wire.GatehouseService().GetGatehouse(ctx, cfg.PlaceID)
		if err != nil {
			return ""
		}
		focusID, err := wire.WorkshopService().GetFocusedConclaveID(ctx, gatehouse.WorkshopID)
		if err != nil {
			return ""
		}
		return focusID
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
	case strings.HasPrefix(focusID, "CON-"):
		if con, err := wire.ConclaveService().GetConclave(ctx, focusID); err == nil {
			return "Conclave", con.Title, con.Status
		}
	case strings.HasPrefix(focusID, "TOME-"):
		if tome, err := wire.TomeService().GetTome(ctx, focusID); err == nil {
			return "Tome", tome.Title, tome.Status
		}
	}
	return "", "", ""
}
