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
		Long: `Focus on a specific container (Commission, Shipment, Conclave, or Tome).

The focused container appears in 'orc prime' output and can be used as default
for other commands.

Container types are auto-detected from ID prefix:
  COMM-*  → Commission (work package)
  SHIP-*  → Shipment (execution work)
  CON-*   → Conclave (ideation session)
  TOME-*  → Tome (knowledge collection)

When focusing on a Shipment while in a workbench context, the shipment's branch
will be automatically checked out using the stash dance.

Examples:
  orc focus COMM-001    # Focus on a commission
  orc focus SHIP-178    # Focus on a shipment (auto-checkouts branch)
  orc focus             # Clear focus
  orc focus --show      # Show current focus`,
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

	// Load config from cwd only
	cfg, err := config.LoadConfig(cwd)
	if err != nil {
		return fmt.Errorf("no ORC config found in current directory")
	}

	// Focus is only supported for IMP contexts with a workbench
	if cfg.Role != config.RoleIMP || cfg.WorkbenchID == "" {
		return fmt.Errorf("focus is only available in workbench contexts (IMP role)")
	}

	if showOnly {
		return showCurrentFocus(cfg.WorkbenchID)
	}

	if clearFlag || len(args) == 0 {
		// Clear focus
		return clearFocus(cfg.WorkbenchID)
	}

	// Set focus
	containerID := args[0]
	containerType, title, err := validateAndGetInfo(containerID)
	if err != nil {
		return err
	}

	return setFocus(cfg, containerID, containerType, title)
}

// validateAndGetInfo validates the container ID exists and returns its type and title
func validateAndGetInfo(id string) (containerType string, title string, err error) {
	ctx := context.Background()
	switch {
	case strings.HasPrefix(id, "COMM-"):
		comm, err := wire.CommissionService().GetCommission(ctx, id)
		if err != nil {
			return "", "", fmt.Errorf("commission %s not found", id)
		}
		return "Commission", comm.Title, nil

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
		return "", "", fmt.Errorf("unknown container type for ID: %s (expected COMM-*, SHIP-*, CON-*, or TOME-*)", id)
	}
}

// showCurrentFocus displays the current focus from DB
func showCurrentFocus(workbenchID string) error {
	ctx := context.Background()

	focusID, err := wire.WorkbenchService().GetFocusedID(ctx, workbenchID)
	if err != nil {
		return fmt.Errorf("failed to get focus: %w", err)
	}

	if focusID == "" {
		fmt.Println("No focus set")
		fmt.Println("\nSet focus with: orc focus <container-id>")
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

// setFocus sets the focus in the DB
func setFocus(cfg *config.Config, containerID, containerType, title string) error {
	ctx := context.Background()

	// Update focus in DB
	if err := wire.WorkbenchService().UpdateFocusedID(ctx, cfg.WorkbenchID, containerID); err != nil {
		return fmt.Errorf("failed to set focus: %w", err)
	}

	fmt.Printf("Focused on %s: %s\n", containerType, containerID)
	fmt.Printf("  %s\n", title)

	// Auto-checkout branch for shipments when in a workbench
	if strings.HasPrefix(containerID, "SHIP-") && cfg.WorkbenchID != "" {
		if err := autoCheckoutShipmentBranch(cfg.WorkbenchID, containerID); err != nil {
			fmt.Printf("  (branch checkout skipped: %v)\n", err)
		} else {
			fmt.Println("  ✓ Branch checked out")
		}
	}

	// Auto-rename tmux session for any focused container
	if os.Getenv("TMUX") != "" {
		if err := autoRenameTmuxSession(cfg, containerID, title); err != nil {
			fmt.Printf("  (tmux session rename skipped: %v)\n", err)
		}
	}

	fmt.Println("\nRun 'orc prime' to see updated context.")
	return nil
}

// autoRenameTmuxSession renames the tmux session to reflect the focused container.
// Format: "Workshop Name - COMM-XXX/SHIP-XXX - Title" (includes commission context)
func autoRenameTmuxSession(cfg *config.Config, containerID, title string) error {
	ctx := context.Background()

	// Get current session name
	currentSession := wire.TMuxAdapter().GetCurrentSessionName(ctx)
	if currentSession == "" {
		return fmt.Errorf("could not determine current session")
	}

	// Get workshop name from workbench context if available
	workshopName := "Workshop"
	if cfg.WorkbenchID != "" {
		wb, err := wire.WorkbenchService().GetWorkbench(ctx, cfg.WorkbenchID)
		if err == nil && wb.WorkshopID != "" {
			ws, err := wire.WorkshopService().GetWorkshop(ctx, wb.WorkshopID)
			if err == nil {
				workshopName = ws.Name
			}
		}
	}

	// Resolve commission context for non-commission containers
	contextPart := containerID
	switch {
	case strings.HasPrefix(containerID, "COMM-"):
		// Commission is the top level, use as-is
		contextPart = containerID
	case strings.HasPrefix(containerID, "SHIP-"):
		if ship, err := wire.ShipmentService().GetShipment(ctx, containerID); err == nil && ship.CommissionID != "" {
			contextPart = ship.CommissionID + "/" + containerID
		}
	case strings.HasPrefix(containerID, "CON-"):
		if con, err := wire.ConclaveService().GetConclave(ctx, containerID); err == nil && con.CommissionID != "" {
			contextPart = con.CommissionID + "/" + containerID
		}
	case strings.HasPrefix(containerID, "TOME-"):
		if tome, err := wire.TomeService().GetTome(ctx, containerID); err == nil && tome.CommissionID != "" {
			contextPart = tome.CommissionID + "/" + containerID
		}
	}

	// Build new name: "Workshop Name - COMM-XXX/SHIP-XXX - Title"
	newName := fmt.Sprintf("%s - %s - %s", workshopName, contextPart, title)

	return wire.TMuxAdapter().RenameSession(ctx, currentSession, newName)
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

// clearFocus clears the current focus in DB
func clearFocus(workbenchID string) error {
	ctx := context.Background()

	if err := wire.WorkbenchService().UpdateFocusedID(ctx, workbenchID, ""); err != nil {
		return fmt.Errorf("failed to clear focus: %w", err)
	}

	fmt.Println("Focus cleared")
	return nil
}

// GetCurrentFocus is exported for use by other commands (e.g., prime)
// Returns the focused ID from DB if in workbench context, empty string otherwise
func GetCurrentFocus(cfg *config.Config) string {
	if cfg == nil || cfg.WorkbenchID == "" {
		return ""
	}

	ctx := context.Background()
	focusID, err := wire.WorkbenchService().GetFocusedID(ctx, cfg.WorkbenchID)
	if err != nil {
		return ""
	}
	return focusID
}

// GetFocusInfo returns the type and title for a focus ID, or empty strings if invalid
func GetFocusInfo(focusID string) (containerType, title, status string) {
	if focusID == "" {
		return "", "", ""
	}

	ctx := context.Background()
	switch {
	case strings.HasPrefix(focusID, "COMM-"):
		if comm, err := wire.CommissionService().GetCommission(ctx, focusID); err == nil {
			return "Commission", comm.Title, comm.Status
		}
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
