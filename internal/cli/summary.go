package cli

import (
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/example/orc/internal/config"
	orcctx "github.com/example/orc/internal/context"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

// commissionAliases maps user-friendly aliases to commission IDs
var commissionAliases = map[string]string{
	"test": "COMM-003",
}

// resolveCommissionAlias resolves a commission alias to its ID, or returns the input unchanged
func resolveCommissionAlias(input string) string {
	if resolved, ok := commissionAliases[input]; ok {
		return resolved
	}
	return input
}

// resolveContainerCommission looks up the commission_id for any container type
func resolveContainerCommission(containerID string) string {
	ctx := context.Background()

	// Determine container type from ID prefix
	switch {
	case strings.HasPrefix(containerID, "CON-"):
		if conclave, err := wire.ConclaveService().GetConclave(ctx, containerID); err == nil {
			return conclave.CommissionID
		}
	case strings.HasPrefix(containerID, "SHIP-"):
		if shipment, err := wire.ShipmentService().GetShipment(ctx, containerID); err == nil {
			return shipment.CommissionID
		}
	case strings.HasPrefix(containerID, "TOME-"):
		if tome, err := wire.TomeService().GetTome(ctx, containerID); err == nil {
			return tome.CommissionID
		}
	case strings.HasPrefix(containerID, "COMM-"):
		// Focus is already a commission
		return containerID
	}
	return ""
}

// SummaryCmd returns the summary command
func SummaryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Show summary of commissions with shipments and tomes",
		Long: `Show a summary of commissions with shipments and tomes.

Display modes:
  Default: Show only focused container's commission (if focus is set)
  --all: Show all commissions and containers
  --commission [id]: Show specific commission (or 'current' for focus/context)

Structure:
  Commission
  ├── Shipment (implementation work)
  └── Tome (exploration notes)

Examples:
  orc summary                          # focused container's commission only
  orc summary --all                    # all commissions
  orc summary --commission COMM-001    # specific commission`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get current working directory for config
			cwd, err := os.Getwd()
			if err != nil {
				cwd = ""
			}

			// Get flags
			commissionFilter, _ := cmd.Flags().GetString("commission")
			expandAll, _ := cmd.Flags().GetBool("all")
			debugMode, _ := cmd.Flags().GetBool("debug")

			// Load config for role detection (with Goblin migration if needed)
			cfg, _ := MigrateGoblinConfigIfNeeded(cmd.Context(), cwd)
			role := config.RoleGoblin // Default to Goblin
			workbenchID := ""
			workshopID := ""
			gatehouseID := ""

			if cfg != nil && cfg.PlaceID != "" {
				role = config.GetRoleFromPlaceID(cfg.PlaceID)
				if config.IsWorkbench(cfg.PlaceID) {
					workbenchID = cfg.PlaceID
					// Look up workshop and gatehouse from workbench
					if wb, err := wire.WorkbenchService().GetWorkbench(cmd.Context(), cfg.PlaceID); err == nil {
						workshopID = wb.WorkshopID
						if gh, err := wire.GatehouseService().GetGatehouseByWorkshop(cmd.Context(), wb.WorkshopID); err == nil {
							gatehouseID = gh.ID
						}
					}
				} else if config.IsGatehouse(cfg.PlaceID) {
					gatehouseID = cfg.PlaceID
					// Look up workshop from gatehouse
					gatehouse, err := wire.GatehouseService().GetGatehouse(cmd.Context(), cfg.PlaceID)
					if err == nil {
						workshopID = gatehouse.WorkshopID
					}
				}
			}

			// Get current focus
			focusID := GetCurrentFocus(cfg)

			// Determine which commission to show
			var filterCommissionID string
			if commissionFilter == "current" {
				// First try config in cwd
				commissionID := orcctx.GetContextCommissionID()
				// Fall back to resolving from focus
				if commissionID == "" && focusID != "" {
					commissionID = resolveContainerCommission(focusID)
				}
				if commissionID == "" {
					return fmt.Errorf("--commission current requires a focused container or being in a commission context")
				}
				filterCommissionID = commissionID
			} else if commissionFilter != "" {
				// Resolve aliases first (e.g., "test" -> "COMM-003")
				resolved := resolveCommissionAlias(commissionFilter)

				// Validate commission exists
				if _, err := wire.CommissionService().GetCommission(cmd.Context(), resolved); err != nil {
					return fmt.Errorf("commission %q not found", commissionFilter)
				}
				filterCommissionID = resolved
			} else if focusID != "" && !expandAll {
				// DEFAULT BEHAVIOR: When focused (and not --all), scope to focused commission
				filterCommissionID = resolveContainerCommission(focusID)
			}

			// Get list of commissions to display
			commissions, err := wire.CommissionService().ListCommissions(context.Background(), primary.CommissionFilters{})
			if err != nil {
				return fmt.Errorf("failed to list commissions: %w", err)
			}

			// Filter to open commissions
			var openCommissions []*primary.Commission
			for _, m := range commissions {
				if m.Status == "complete" || m.Status == "archived" {
					continue
				}
				if filterCommissionID != "" && m.ID != filterCommissionID {
					continue
				}
				openCommissions = append(openCommissions, m)
			}

			if len(openCommissions) == 0 {
				if filterCommissionID != "" {
					fmt.Printf("No open containers for %s\n", filterCommissionID)
				} else {
					fmt.Println("No open commissions")
				}
				return nil
			}

			// Render header based on role
			renderHeader(role, workbenchID, workshopID, gatehouseID, focusID, filterCommissionID)

			// Build map of focused containers across all workbenches in this workshop
			workshopFocus := buildWorkshopFocusMap(cmd.Context(), workshopID, workbenchID, gatehouseID)

			// Display each commission
			for i, commission := range openCommissions {
				// Build summary request
				req := primary.SummaryRequest{
					CommissionID: commission.ID,
					WorkbenchID:  workbenchID,
					WorkshopID:   workshopID,
					FocusID:      focusID,
					DebugMode:    debugMode,
				}

				summary, err := wire.SummaryService().GetCommissionSummary(context.Background(), req)
				if err != nil {
					fmt.Printf("Error getting summary for %s: %v\n", commission.ID, err)
					continue
				}

				// Render summary (unified view for all roles)
				renderSummary(summary, focusID, workshopFocus)

				// Render debug info if present
				if summary.DebugInfo != nil && len(summary.DebugInfo.Messages) > 0 {
					fmt.Println()
					renderDebugInfo(summary.DebugInfo)
				}

				if i < len(openCommissions)-1 {
					fmt.Println()
				}
			}

			return nil
		},
	}

	cmd.Flags().StringP("commission", "c", "", "Commission filter: commission ID or 'current' for context commission")
	cmd.Flags().Bool("all", false, "Show all containers (default: only show focused container if set)")
	cmd.Flags().Bool("debug", false, "Show debug info about hidden/filtered content")

	return cmd
}

// renderHeader prints the header line based on role
func renderHeader(role, workbenchID, workshopID, gatehouseID, focusID, commissionID string) {
	// Show workshop context (for both Goblin and IMP)
	if workshopID != "" {
		fmt.Printf("Workshop %s", workshopID)
		if gatehouseID != "" {
			fmt.Printf(" (Gatehouse: %s)", gatehouseID)
		}
		if config.IsGoblinRole(role) && commissionID != "" {
			fmt.Printf(" - Active: %s", commissionID)
		}
		fmt.Println()

		// Show workbenches in workshop
		renderWorkshopBenches(workshopID, workbenchID)
	}

	// IMP-specific: show current workbench and focus
	if !config.IsGoblinRole(role) && workbenchID != "" {
		if focusID != "" {
			fmt.Printf("Focus: %s\n", focusID)
		}
	}

	fmt.Println()
}

// renderWorkshopBenches displays workbenches in the workshop
func renderWorkshopBenches(workshopID, currentWorkbenchID string) {
	ctx := context.Background()

	workbenches, err := wire.WorkbenchService().ListWorkbenches(ctx, primary.WorkbenchFilters{
		WorkshopID: workshopID,
	})
	if err != nil || len(workbenches) == 0 {
		return
	}

	fmt.Println("Workbenches:")
	for _, wb := range workbenches {
		marker := ""
		if wb.ID == currentWorkbenchID {
			marker = color.New(color.FgHiMagenta).Sprint(" ←")
		}
		fmt.Printf("  - %s (%s)%s\n", wb.ID, wb.Name, marker)
	}
}

// workshopFocusInfo tracks what each workbench in the workshop has focused
type workshopFocusInfo struct {
	containerToWorkbench map[string]string // containerID -> "name@ID" that has it focused
	myName               string            // current workbench name
	myID                 string            // current workbench ID
	goblinID             string            // gatehouse ID
}

// buildWorkshopFocusMap fetches focus for all workbenches and the goblin in the workshop
func buildWorkshopFocusMap(ctx context.Context, workshopID, currentWorkbenchID, gatehouseID string) workshopFocusInfo {
	info := workshopFocusInfo{
		containerToWorkbench: make(map[string]string),
		myID:                 currentWorkbenchID,
		goblinID:             gatehouseID,
	}

	if workshopID == "" {
		return info
	}

	// Get current workbench name
	if currentWorkbenchID != "" {
		if wb, err := wire.WorkbenchService().GetWorkbench(ctx, currentWorkbenchID); err == nil {
			info.myName = wb.Name
		}
	}

	// If no workbench but gatehouse exists, we're running as Goblin
	if info.myName == "" && gatehouseID != "" {
		info.myName = "goblin"
		info.myID = gatehouseID
	}

	// Get Goblin's focus (workshop-level focus)
	goblinFocusID, err := wire.WorkshopService().GetFocusedConclaveID(ctx, workshopID)
	if err == nil && goblinFocusID != "" {
		info.containerToWorkbench[goblinFocusID] = fmt.Sprintf("goblin@%s", gatehouseID)
	}

	// Get each IMP's focus
	workbenches, err := wire.WorkbenchService().ListWorkbenches(ctx, primary.WorkbenchFilters{
		WorkshopID: workshopID,
	})
	if err != nil {
		return info
	}

	for _, wb := range workbenches {
		// Skip our own workbench - handled separately for [FOCUSED] marker
		if wb.ID == currentWorkbenchID {
			continue
		}

		focusedID, err := wire.WorkbenchService().GetFocusedID(ctx, wb.ID)
		if err != nil || focusedID == "" {
			continue
		}

		info.containerToWorkbench[focusedID] = fmt.Sprintf("%s@%s", wb.Name, wb.ID)
	}

	return info
}

// renderSummary renders the commission with flat lists of shipments and tomes
func renderSummary(summary *primary.CommissionSummary, _ string, workshopFocus workshopFocusInfo) {
	// Commission header with focused marker
	focusedMarker := ""
	if summary.IsFocusedCommission {
		focusedMarker = color.New(color.FgHiMagenta).Sprint(" [focused]")
	}
	fmt.Printf("%s%s - %s\n", colorizeID(summary.ID), focusedMarker, summary.Title)
	fmt.Println("│")

	totalItems := len(summary.Shipments) + len(summary.Tomes)
	itemIdx := 0

	// Render shipments
	for _, ship := range summary.Shipments {
		isLast := itemIdx == totalItems-1
		prefix := "├── "
		taskPrefix := "│   "
		if isLast {
			prefix = "└── "
			taskPrefix = "    "
		}

		benchMarker := ""
		if ship.BenchID != "" {
			if ship.BenchName != "" {
				benchMarker = color.New(color.FgCyan).Sprintf(" [assigned to %s@%s]", ship.BenchName, ship.BenchID)
			} else {
				benchMarker = color.New(color.FgCyan).Sprintf(" [assigned to %s]", ship.BenchID)
			}
		}
		statusBadge := ""
		if ship.Status != "" && ship.Status != "complete" {
			statusBadge = " " + colorizeShipmentStatus(ship.Status)
		}
		taskInfo := fmt.Sprintf(" (%d/%d done)", ship.TasksDone, ship.TasksTotal)
		pinnedMark := ""
		if ship.Pinned {
			pinnedMark = " *"
		}
		focusMark := ""
		if ship.IsFocused {
			focusMark = color.New(color.FgHiMagenta).Sprint(" [focused by you]")
		} else if who := workshopFocus.containerToWorkbench[ship.ID]; who != "" {
			focusMark = color.New(color.FgCyan).Sprintf(" [focused by %s]", who)
		}

		fmt.Printf("%s%s%s%s%s%s - %s%s\n", prefix, colorizeID(ship.ID), statusBadge, benchMarker, focusMark, pinnedMark, ship.Title, taskInfo)

		// Expand tasks for focused shipment
		if ship.IsFocused && len(ship.Tasks) > 0 {
			for j, task := range ship.Tasks {
				isLastTask := j == len(ship.Tasks)-1
				tPrefix := taskPrefix + "├── "
				taskChildPrefix := taskPrefix + "│   "
				if isLastTask {
					tPrefix = taskPrefix + "└── "
					taskChildPrefix = taskPrefix + "    "
				}
				statusMark := ""
				if task.Status != "" && task.Status != "ready" {
					statusMark = colorizeStatus(task.Status) + " - "
				}
				fmt.Printf("%s%s - %s%s\n", tPrefix, colorizeID(task.ID), statusMark, task.Title)
				// Render task children (plans, approvals, escalations, receipts)
				renderTaskChildren(task, taskChildPrefix)
			}
		}

		itemIdx++
	}

	// Render tomes
	for _, tome := range summary.Tomes {
		isLast := itemIdx == totalItems-1
		tomePrefix := "├── "
		tomeChildPrefix := "│   "
		if isLast {
			tomePrefix = "└── "
			tomeChildPrefix = "    "
		}

		noteInfo := ""
		if tome.NoteCount > 0 && len(tome.Notes) == 0 {
			noteInfo = fmt.Sprintf(" (%s)", pluralize(tome.NoteCount, "note", "notes"))
		}
		pinnedMark := ""
		if tome.Pinned {
			pinnedMark = " *"
		}
		focusMark := ""
		if tome.IsFocused {
			focusMark = color.New(color.FgHiMagenta).Sprint(" [focused by you]")
		} else if who := workshopFocus.containerToWorkbench[tome.ID]; who != "" {
			focusMark = color.New(color.FgCyan).Sprintf(" [focused by %s]", who)
		}

		fmt.Printf("%s%s%s%s - %s%s\n", tomePrefix, colorizeID(tome.ID), focusMark, pinnedMark, tome.Title, noteInfo)

		// Expand notes for focused tome
		if len(tome.Notes) > 0 {
			for j, note := range tome.Notes {
				isLastNote := j == len(tome.Notes)-1
				notePrefix := tomeChildPrefix + "├── "
				if isLastNote {
					notePrefix = tomeChildPrefix + "└── "
				}
				typeMarker := ""
				if note.Type != "" {
					typeMarker = color.New(color.FgYellow).Sprintf("[%s] ", note.Type)
				}
				fmt.Printf("%s%s %s- %s\n", notePrefix, colorizeID(note.ID), typeMarker, note.Title)
			}
		}

		itemIdx++
	}
}

// pluralize returns "N singular" or "N plural" based on count
func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}
	return fmt.Sprintf("%d %s", count, plural)
}

// colorizeID applies deterministic color to an ID based on its prefix
func colorizeID(id string) string {
	// Extract prefix (everything before first hyphen)
	parts := strings.Split(id, "-")
	if len(parts) < 2 {
		return id // No prefix, return plain
	}

	prefix := parts[0] // "TASK", "SHIP", "COMM"
	c := getIDColor(prefix)
	return c.Sprint(id)
}

// getIDColor returns a deterministic color for an ID type (TASK, SHIP, COMM, etc.)
// Uses FNV-1a hash on the ID prefix to ensure all IDs of same type have same color
func getIDColor(idType string) *color.Color {
	h := fnv.New32a()
	h.Write([]byte(idType))
	hash := h.Sum32()

	// Map to 256-color range (16-231 are the color cube)
	colorCode := 16 + (hash % 216)

	return color.New(color.Attribute(38), color.Attribute(5), color.Attribute(colorCode))
}

// colorizeStatus formats status with semantic color
func colorizeStatus(status string) string {
	if status == "" || status == "ready" {
		return ""
	}
	upper := strings.ToUpper(status)

	switch status {
	case "in_progress":
		return color.New(color.FgHiBlue).Sprint(upper)
	case "paused":
		return color.New(color.FgYellow).Sprint(upper)
	case "blocked":
		return color.New(color.FgRed).Sprint(upper)
	case "complete":
		return color.New(color.FgHiGreen).Sprint(upper)
	default:
		return color.New(color.FgWhite).Sprint(upper)
	}
}

// colorizeShipmentStatus formats shipment status badge with semantic color
func colorizeShipmentStatus(status string) string {
	switch status {
	case "draft":
		return color.New(color.FgHiBlack).Sprint("[draft]")
	case "exploring":
		return color.New(color.FgHiCyan).Sprint("[exploring]")
	case "specced":
		return color.New(color.FgHiMagenta).Sprint("[specced]")
	case "tasked":
		return color.New(color.FgHiYellow).Sprint("[tasked]")
	case "in_progress":
		return color.New(color.FgHiBlue).Sprint("[in_progress]")
	case "paused":
		return color.New(color.FgYellow).Sprint("[paused]")
	case "complete":
		return color.New(color.FgHiGreen).Sprint("[complete]")
	default:
		return fmt.Sprintf("[%s]", status)
	}
}

// colorizePlanStatus formats plan status with semantic color and marker
func colorizePlanStatus(status string) string {
	upper := strings.ToUpper(status)
	switch status {
	case "approved":
		return color.New(color.FgHiGreen).Sprintf("✓ %s", upper)
	case "escalated":
		return color.New(color.FgYellow).Sprintf("⚠ %s", upper)
	case "pending_review":
		return color.New(color.FgCyan).Sprint(upper)
	default:
		return upper // draft, superseded
	}
}

// colorizeApprovalOutcome formats approval outcome with semantic color and marker
func colorizeApprovalOutcome(outcome string) string {
	upper := strings.ToUpper(outcome)
	switch outcome {
	case "approved":
		return color.New(color.FgHiGreen).Sprintf("✓ %s", upper)
	case "escalated":
		return color.New(color.FgYellow).Sprintf("⚠ %s", upper)
	default:
		return upper
	}
}

// colorizeEscalationStatus formats escalation status with semantic color and marker
func colorizeEscalationStatus(status string, targetActorID string) string {
	upper := strings.ToUpper(status)
	targetInfo := ""
	if targetActorID != "" {
		targetInfo = fmt.Sprintf(" (%s)", targetActorID)
	}
	switch status {
	case "pending":
		return color.New(color.FgYellow).Sprintf("⚠ %s%s", upper, targetInfo)
	case "resolved":
		return color.New(color.FgHiGreen).Sprintf("✓ %s%s", upper, targetInfo)
	default:
		return fmt.Sprintf("%s%s", upper, targetInfo)
	}
}

// colorizeReceiptStatus formats receipt status with semantic color and marker
func colorizeReceiptStatus(status string) string {
	upper := strings.ToUpper(status)
	switch status {
	case "verified":
		return color.New(color.FgHiGreen).Sprintf("✓ %s", upper)
	case "submitted":
		return color.New(color.FgCyan).Sprint(upper)
	default:
		return upper // draft
	}
}

// renderTaskChildren renders the child entities (plans, approvals, escalations, receipts) under a task
func renderTaskChildren(task primary.TaskSummary, prefix string) {
	// Count total children
	totalChildren := len(task.Plans) + len(task.Approvals) + len(task.Escalations) + len(task.Receipts)
	if totalChildren == 0 {
		return
	}

	childIdx := 0

	// Render plans
	for _, plan := range task.Plans {
		isLast := childIdx == totalChildren-1
		childPrefix := prefix + "├── "
		if isLast {
			childPrefix = prefix + "└── "
		}
		fmt.Printf("%s%s %s\n", childPrefix, colorizeID(plan.ID), colorizePlanStatus(plan.Status))
		childIdx++
	}

	// Render approvals
	for _, approval := range task.Approvals {
		isLast := childIdx == totalChildren-1
		childPrefix := prefix + "├── "
		if isLast {
			childPrefix = prefix + "└── "
		}
		fmt.Printf("%s%s %s\n", childPrefix, colorizeID(approval.ID), colorizeApprovalOutcome(approval.Outcome))
		childIdx++
	}

	// Render escalations
	for _, esc := range task.Escalations {
		isLast := childIdx == totalChildren-1
		childPrefix := prefix + "├── "
		if isLast {
			childPrefix = prefix + "└── "
		}
		fmt.Printf("%s%s %s\n", childPrefix, colorizeID(esc.ID), colorizeEscalationStatus(esc.Status, esc.TargetActorID))
		childIdx++
	}

	// Render receipts
	for _, receipt := range task.Receipts {
		isLast := childIdx == totalChildren-1
		childPrefix := prefix + "├── "
		if isLast {
			childPrefix = prefix + "└── "
		}
		fmt.Printf("%s%s %s\n", childPrefix, colorizeID(receipt.ID), colorizeReceiptStatus(receipt.Status))
		childIdx++
	}
}

// renderDebugInfo renders the debug information section
func renderDebugInfo(info *primary.DebugInfo) {
	debugColor := color.New(color.FgHiBlack)
	debugColor.Println("─── Debug Info ───")
	for _, msg := range info.Messages {
		debugColor.Printf("  %s\n", msg)
	}
}
