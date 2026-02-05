package cli

import (
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
		Long: `Focus on any commission-level entity.

Focusable entities:
  - COMM-xxx  Commission (watch a work area)
  - SHIP-xxx  Shipment (active implementation)
  - TOME-xxx  Tome (research/exploration)
  - NOTE-xxx  Note (root-level only, not inside shipments)

Smart clear: Running --clear refocuses to the commission of the current focus.
Use --clear --force to fully clear focus with no fallback.

Examples:
  orc focus SHIP-178        # Focus on a shipment
  orc focus TOME-028        # Focus on a tome
  orc focus COMM-001        # Focus on a commission
  orc focus NOTE-322        # Focus on a root-level note
  orc focus --show          # Show current focus
  orc focus --clear         # Smart clear (refocus to commission)
  orc focus --clear --force # Fully clear focus`,
		Args: cobra.MaximumNArgs(1),
		RunE: runFocus,
	}
	cmd.Flags().Bool("show", false, "Show current focus without changing it")
	cmd.Flags().Bool("clear", false, "Clear the current focus")
	cmd.Flags().Bool("force", false, "Fully clear focus (no fallback to commission)")
	return cmd
}

func runFocus(cmd *cobra.Command, args []string) error {
	showOnly, _ := cmd.Flags().GetBool("show")
	clearFlag, _ := cmd.Flags().GetBool("clear")
	forceFlag, _ := cmd.Flags().GetBool("force")

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
		return runIMPFocus(cmd, args, cfg, showOnly, clearFlag, forceFlag)
	case config.PlaceTypeGatehouse:
		// Goblin role - use gatehouse focus
		return runGoblinFocus(cmd, args, cfg, showOnly, clearFlag, forceFlag)
	default:
		return fmt.Errorf("focus requires workbench (IMP) context")
	}
}

// runIMPFocus handles focus for IMP role (workbench context)
// Can focus on any commission-level entity: COMM-xxx, SHIP-xxx, TOME-xxx
func runIMPFocus(_ *cobra.Command, args []string, cfg *config.Config, showOnly, clearFlag, forceFlag bool) error {
	workbenchID := cfg.PlaceID // BENCH-XXX

	if showOnly {
		return showIMPFocus(workbenchID)
	}

	if clearFlag {
		return clearIMPFocus(workbenchID, forceFlag)
	}

	if len(args) == 0 {
		return fmt.Errorf("Usage: orc focus <ID> or orc focus --show or orc focus --clear")
	}

	// Set focus
	containerID := args[0]

	containerType, title, err := validateFocusTarget(containerID)
	if err != nil {
		return err
	}

	return setIMPFocus(workbenchID, containerID, containerType, title)
}

// validateFocusTarget validates the container ID exists and returns its type and title
// Supports COMM-xxx, SHIP-xxx, TOME-xxx (any actor can focus any type)
func validateFocusTarget(id string) (containerType string, title string, err error) {
	ctx := NewContext()
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

	case strings.HasPrefix(id, "TOME-"):
		tome, err := wire.TomeService().GetTome(ctx, id)
		if err != nil {
			return "", "", fmt.Errorf("tome %s not found", id)
		}
		return "Tome", tome.Title, nil

	case strings.HasPrefix(id, "NOTE-"):
		note, err := wire.NoteService().GetNote(ctx, id)
		if err != nil {
			return "", "", fmt.Errorf("note %s not found", id)
		}
		// Only root-level notes can be focused (not inside a shipment)
		if note.ShipmentID != "" {
			return "", "", fmt.Errorf("cannot focus %s: note is inside %s, focus the shipment instead", id, note.ShipmentID)
		}
		return "Note", note.Title, nil

	default:
		return "", "", fmt.Errorf("unknown container type for ID: %s (expected COMM-*, SHIP-*, TOME-*, or NOTE-*)", id)
	}
}

// resolveToCommission resolves any focusable entity to its commission ID
func resolveToCommission(id string) string {
	ctx := NewContext()
	switch {
	case strings.HasPrefix(id, "COMM-"):
		return id
	case strings.HasPrefix(id, "SHIP-"):
		if ship, err := wire.ShipmentService().GetShipment(ctx, id); err == nil {
			return ship.CommissionID
		}
	case strings.HasPrefix(id, "TOME-"):
		if tome, err := wire.TomeService().GetTome(ctx, id); err == nil {
			return tome.CommissionID
		}
	case strings.HasPrefix(id, "NOTE-"):
		if note, err := wire.NoteService().GetNote(ctx, id); err == nil {
			return note.CommissionID
		}
	}
	return ""
}

// showIMPFocus displays the current IMP focus from DB
func showIMPFocus(workbenchID string) error {
	ctx := NewContext()

	focusID, err := wire.WorkbenchService().GetFocusedID(ctx, workbenchID)
	if err != nil {
		return fmt.Errorf("failed to get focus: %w", err)
	}

	if focusID == "" {
		fmt.Println("No focus set")
		fmt.Println("\nSet focus with: orc focus <SHIP-xxx> or orc focus <TOME-xxx>")
		return nil
	}

	containerType, title, err := validateFocusTarget(focusID)
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
	ctx := NewContext()

	// Check for focus exclusivity - another IMP cannot focus the same container
	// Notes are exempt from exclusivity (multiple actors can focus the same note)
	if !strings.HasPrefix(containerID, "NOTE-") {
		otherWorkbenches, err := wire.WorkbenchService().GetWorkbenchesByFocusedID(ctx, containerID)
		if err != nil {
			return fmt.Errorf("failed to check focus exclusivity: %w", err)
		}
		for _, wb := range otherWorkbenches {
			if wb.ID != workbenchID {
				return fmt.Errorf("%s is already focused by %s (%s)", containerID, wb.ID, wb.Name)
			}
		}
	}

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

	// Update tmux ORC_CONTEXT for session browser
	updateWorkshopContext(workbenchID)

	// Update tmux @orc_focus window option for session picker
	updateFocusWindowOption(containerID, title)

	fmt.Println("\nRun 'orc prime' to see updated context.")
	return nil
}

// autoCheckoutShipmentBranch checks out the shipment's branch in the workbench
func autoCheckoutShipmentBranch(workbenchID, shipmentID string) error {
	ctx := NewContext()

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

// updateFocusWindowOption sets the @orc_focus tmux window option for the session picker.
// Format: "{ID}: {title}" where ID is SHIP-xxx, TOME-xxx, COMM-xxx, etc.
func updateFocusWindowOption(containerID, title string) {
	ctx := NewContext()

	// Get current tmux session and window
	sessionName := wire.TMuxAdapter().GetCurrentSessionName(ctx)
	if sessionName == "" {
		return // Not in tmux
	}

	// Build focus string
	focusStr := ""
	if containerID != "" && title != "" {
		focusStr = fmt.Sprintf("%s: %s", containerID, title)
	}

	// Set window option on current window (session:)
	target := sessionName + ":"
	_ = wire.TMuxAdapter().SetWindowOption(ctx, target, "@orc_focus", focusStr)
}

// updateWorkshopContext computes active commissions for the workshop and sets ORC_CONTEXT tmux env var.
// This enriches the tmux session browser (prefix+s) with commission info.
func updateWorkshopContext(workbenchID string) {
	ctx := NewContext()

	// Get workbench to find workshop ID
	wb, err := wire.WorkbenchService().GetWorkbench(ctx, workbenchID)
	if err != nil || wb.WorkshopID == "" {
		return // Silently fail - this is a nice-to-have
	}

	// Get current tmux session name
	sessionName := wire.TMuxAdapter().GetCurrentSessionName(ctx)
	if sessionName == "" {
		return // Not in tmux
	}

	// Get active commissions from workshop (derives from all focused entities)
	commissionIDs, err := wire.WorkshopService().GetActiveCommissions(ctx, wb.WorkshopID)
	if err != nil {
		return
	}

	// Build context string: "Title [ID], Title [ID], ..."
	var parts []string
	for _, commID := range commissionIDs {
		comm, err := wire.CommissionService().GetCommission(ctx, commID)
		if err != nil {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s [%s]", comm.Title, commID))
	}

	context := strings.Join(parts, ", ")
	if context == "" {
		context = "(idle)"
	}

	// Set tmux environment variable
	_ = wire.TMuxAdapter().SetEnvironment(ctx, sessionName, "ORC_CONTEXT", context)
}

// clearIMPFocus clears the IMP focus in DB
// If force=false, smart clear: refocus to the commission of the current focus
// If force=true, fully clear focus (no fallback)
func clearIMPFocus(workbenchID string, force bool) error {
	ctx := NewContext()

	// Get current focus
	currentFocusID, _ := wire.WorkbenchService().GetFocusedID(ctx, workbenchID)

	if currentFocusID == "" {
		fmt.Println("No focus set")
		return nil
	}

	// If --force, fully clear
	if force {
		if err := wire.WorkbenchService().UpdateFocusedID(ctx, workbenchID, ""); err != nil {
			return fmt.Errorf("failed to clear focus: %w", err)
		}
		updateFocusWindowOption("", "") // Clear @orc_focus
		fmt.Println("Focus fully cleared")
		return nil
	}

	// Already at commission level?
	if strings.HasPrefix(currentFocusID, "COMM-") {
		fmt.Println("Already at commission level")
		return nil
	}

	// Smart clear: resolve to commission and refocus
	commissionID := resolveToCommission(currentFocusID)
	if commissionID == "" {
		// Fallback: if we can't resolve, just clear
		if err := wire.WorkbenchService().UpdateFocusedID(ctx, workbenchID, ""); err != nil {
			return fmt.Errorf("failed to clear focus: %w", err)
		}
		updateFocusWindowOption("", "") // Clear @orc_focus
		fmt.Println("Focus cleared")
		return nil
	}

	// Refocus to commission - get commission title for @orc_focus
	comm, _ := wire.CommissionService().GetCommission(ctx, commissionID)
	commTitle := ""
	if comm != nil {
		commTitle = comm.Title
	}
	if err := wire.WorkbenchService().UpdateFocusedID(ctx, workbenchID, commissionID); err != nil {
		return fmt.Errorf("failed to refocus to commission: %w", err)
	}
	updateFocusWindowOption(commissionID, commTitle)
	fmt.Printf("Focus cleared from %s → %s (commission)\n", currentFocusID, commissionID)
	return nil
}

// runGoblinFocus handles focus for Goblin role (gatehouse context)
// Can focus on any commission-level entity: COMM-xxx, SHIP-xxx, TOME-xxx
func runGoblinFocus(_ *cobra.Command, args []string, cfg *config.Config, showOnly, clearFlag, forceFlag bool) error {
	gatehouseID := cfg.PlaceID // GATE-XXX

	if showOnly {
		return showGoblinFocus(gatehouseID)
	}

	if clearFlag {
		return clearGoblinFocus(gatehouseID, forceFlag)
	}

	if len(args) == 0 {
		return fmt.Errorf("Usage: orc focus <ID> or orc focus --show or orc focus --clear")
	}

	// Set focus
	containerID := args[0]

	containerType, title, err := validateFocusTarget(containerID)
	if err != nil {
		return err
	}

	return setGoblinFocus(gatehouseID, containerID, containerType, title)
}

// showGoblinFocus displays the current Goblin focus from DB
func showGoblinFocus(gatehouseID string) error {
	ctx := NewContext()

	focusID, err := wire.GatehouseService().GetFocusedID(ctx, gatehouseID)
	if err != nil {
		return fmt.Errorf("failed to get focus: %w", err)
	}

	if focusID == "" {
		fmt.Println("No focus set")
		fmt.Println("\nSet focus with: orc focus <COMM-xxx>, <SHIP-xxx>, or <TOME-xxx>")
		return nil
	}

	containerType, title, err := validateFocusTarget(focusID)
	if err != nil {
		// Focus is set but container no longer exists - graceful degradation
		fmt.Printf("Focus: %s (container not found - may have been deleted)\n", focusID)
		return nil //nolint:nilerr // intentional: show info even if container deleted
	}

	fmt.Printf("Focus: %s\n", focusID)
	fmt.Printf("  %s: %s\n", containerType, title)
	return nil
}

// setGoblinFocus sets the Goblin focus in the DB
func setGoblinFocus(gatehouseID, containerID, containerType, title string) error {
	ctx := NewContext()

	// Update focus in DB
	if err := wire.GatehouseService().UpdateFocusedID(ctx, gatehouseID, containerID); err != nil {
		return fmt.Errorf("failed to set focus: %w", err)
	}

	// Update tmux @orc_focus window option for session picker
	updateFocusWindowOption(containerID, title)

	fmt.Printf("Focused on %s: %s\n", containerType, containerID)
	fmt.Printf("  %s\n", title)
	fmt.Println("\nRun 'orc prime' to see updated context.")
	return nil
}

// clearGoblinFocus clears the Goblin focus in DB
// If force=false, smart clear: refocus to the commission of the current focus
// If force=true, fully clear focus (no fallback)
func clearGoblinFocus(gatehouseID string, force bool) error {
	ctx := NewContext()

	// Get current focus
	currentFocusID, _ := wire.GatehouseService().GetFocusedID(ctx, gatehouseID)

	if currentFocusID == "" {
		fmt.Println("No focus set")
		return nil
	}

	// If --force, fully clear
	if force {
		if err := wire.GatehouseService().UpdateFocusedID(ctx, gatehouseID, ""); err != nil {
			return fmt.Errorf("failed to clear focus: %w", err)
		}
		updateFocusWindowOption("", "") // Clear @orc_focus
		fmt.Println("Focus fully cleared")
		return nil
	}

	// Already at commission level?
	if strings.HasPrefix(currentFocusID, "COMM-") {
		fmt.Println("Already at commission level")
		return nil
	}

	// Smart clear: resolve to commission and refocus
	commissionID := resolveToCommission(currentFocusID)
	if commissionID == "" {
		// Fallback: if we can't resolve, just clear
		if err := wire.GatehouseService().UpdateFocusedID(ctx, gatehouseID, ""); err != nil {
			return fmt.Errorf("failed to clear focus: %w", err)
		}
		updateFocusWindowOption("", "") // Clear @orc_focus
		fmt.Println("Focus cleared")
		return nil
	}

	// Refocus to commission - get commission title for @orc_focus
	comm, _ := wire.CommissionService().GetCommission(ctx, commissionID)
	commTitle := ""
	if comm != nil {
		commTitle = comm.Title
	}
	if err := wire.GatehouseService().UpdateFocusedID(ctx, gatehouseID, commissionID); err != nil {
		return fmt.Errorf("failed to refocus to commission: %w", err)
	}
	updateFocusWindowOption(commissionID, commTitle)
	fmt.Printf("Focus cleared from %s → %s (commission)\n", currentFocusID, commissionID)
	return nil
}

// GetCurrentFocus is exported for use by other commands (e.g., prime)
// Returns the focused ID from DB based on place_id context
func GetCurrentFocus(cfg *config.Config) string {
	if cfg == nil || cfg.PlaceID == "" {
		return ""
	}

	ctx := NewContext()

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
		// Gatehouse context - use gatehouse focus
		focusID, err := wire.GatehouseService().GetFocusedID(ctx, cfg.PlaceID)
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

	ctx := NewContext()
	switch {
	case strings.HasPrefix(focusID, "COMM-"):
		if comm, err := wire.CommissionService().GetCommission(ctx, focusID); err == nil {
			return "Commission", comm.Title, comm.Status
		}
	case strings.HasPrefix(focusID, "SHIP-"):
		if ship, err := wire.ShipmentService().GetShipment(ctx, focusID); err == nil {
			return "Shipment", ship.Title, ship.Status
		}
	case strings.HasPrefix(focusID, "TOME-"):
		if tome, err := wire.TomeService().GetTome(ctx, focusID); err == nil {
			return "Tome", tome.Title, tome.Status
		}
	case strings.HasPrefix(focusID, "NOTE-"):
		if note, err := wire.NoteService().GetNote(ctx, focusID); err == nil {
			return "Note", note.Title, note.Status
		}
	}
	return "", "", ""
}
