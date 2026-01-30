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
		Short: "Show hierarchical summary of commissions and containers",
		Long: `Show a hierarchical summary of commissions with conclaves, tomes, and shipments.

Display modes:
  Default: Show only focused container's commission (if focus is set)
  --all: Show all commissions and containers
  --commission [id]: Show specific commission (or 'current' for focus/context)

Role-based views:
  Goblin (Workshop): Full tree with [BENCH-xxx] markers on assigned shipments
  IMP (Workbench): Filtered tree with [FOCUSED] marker, hides other workbenches' shipments

Structure:
  Commission
  ├── Conclave (design discussions)
  │   ├── Tome (notes)
  │   └── Shipment (tasks)
  ├── LIBRARY (parked tomes)
  └── SHIPYARD (parked shipments)

Examples:
  orc summary                          # focused container's commission only
  orc summary --all                    # all commissions
  orc summary --commission COMM-001    # specific commission
  orc summary --all-shipments          # IMP: show shipments from other workbenches`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get current working directory for config
			cwd, err := os.Getwd()
			if err != nil {
				cwd = ""
			}

			// Get flags
			commissionFilter, _ := cmd.Flags().GetString("commission")
			expandAll, _ := cmd.Flags().GetBool("all")
			allShipments, _ := cmd.Flags().GetBool("all-shipments")
			expandLibrary, _ := cmd.Flags().GetBool("expand")

			// Load config for role detection (with Goblin migration if needed)
			cfg, _ := MigrateGoblinConfigIfNeeded(cmd.Context(), cwd)
			role := config.RoleGoblin // Default to Goblin
			workbenchID := ""
			workshopID := ""

			if cfg != nil && cfg.PlaceID != "" {
				role = config.GetRoleFromPlaceID(cfg.PlaceID)
				if config.IsWorkbench(cfg.PlaceID) {
					workbenchID = cfg.PlaceID
				} else if config.IsGatehouse(cfg.PlaceID) {
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
			renderHeader(role, workbenchID, workshopID, focusID, filterCommissionID)

			// Display each commission
			for i, commission := range openCommissions {
				// Build summary request
				req := primary.SummaryRequest{
					CommissionID:     commission.ID,
					Role:             role,
					WorkbenchID:      workbenchID,
					WorkshopID:       workshopID,
					FocusID:          focusID,
					ShowAllShipments: allShipments,
					ExpandLibrary:    expandLibrary,
				}

				summary, err := wire.SummaryService().GetCommissionSummary(context.Background(), req)
				if err != nil {
					fmt.Printf("Error getting summary for %s: %v\n", commission.ID, err)
					continue
				}

				// Render based on role
				if config.IsGoblinRole(role) {
					renderGoblinSummary(summary, focusID)
				} else {
					renderIMPSummary(summary, focusID, allShipments)
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
	cmd.Flags().Bool("all-shipments", false, "Show all shipments including those assigned to other workbenches (IMP only)")
	cmd.Flags().Bool("expand", false, "Expand LIBRARY and SHIPYARD to show individual contents")

	return cmd
}

// renderHeader prints the header line based on role
func renderHeader(role, workbenchID, workshopID, focusID, commissionID string) {
	if config.IsGoblinRole(role) {
		if workshopID != "" {
			fmt.Printf("Workshop %s", workshopID)
			if commissionID != "" {
				fmt.Printf(" - Active: %s", commissionID)
			}
			fmt.Println()
		}
	} else {
		if workbenchID != "" {
			fmt.Printf("Workbench %s", workbenchID)
			if focusID != "" {
				fmt.Printf(" - Focus: %s", focusID)
			}
			fmt.Println()
		}
	}
	fmt.Println()
}

// renderGoblinSummary renders the full tree view for Goblin role
func renderGoblinSummary(summary *primary.CommissionSummary, _ string) {
	// Commission header
	fmt.Printf("%s - %s\n", colorizeID(summary.ID), summary.Title)
	fmt.Println("│")

	// Conclaves
	for i, con := range summary.Conclaves {
		isLastConclave := i == len(summary.Conclaves)-1

		prefix := "├── "
		childPrefix := "│   "
		if isLastConclave && summary.Library.TomeCount == 0 && summary.Shipyard.ShipmentCount == 0 {
			prefix = "└── "
			childPrefix = "    "
		}

		focusMarker := ""
		if con.IsFocused {
			focusMarker = color.New(color.FgHiMagenta).Sprint(" [FOCUSED]")
		}
		pinnedMarker := ""
		if con.Pinned {
			pinnedMarker = " *"
		}

		// Tomes under conclave
		totalChildren := len(con.Tomes) + len(con.Shipments)

		// Build counts suffix for conclave
		countParts := []string{}
		if len(con.Tomes) > 0 {
			countParts = append(countParts, pluralize(len(con.Tomes), "tome", "tomes"))
		}
		if len(con.Shipments) > 0 {
			countParts = append(countParts, pluralize(len(con.Shipments), "shipment", "shipments"))
		}
		countSuffix := ""
		if len(countParts) > 0 {
			countSuffix = fmt.Sprintf(" (%s)", strings.Join(countParts, ", "))
		}

		fmt.Printf("%s%s%s%s - %s%s\n", prefix, colorizeID(con.ID), focusMarker, pinnedMarker, con.Title, countSuffix)
		childIdx := 0

		for _, tome := range con.Tomes {
			isLast := childIdx == totalChildren-1
			tomePrefix := childPrefix + "├── "
			tomeChildPrefix := childPrefix + "│   "
			if isLast {
				tomePrefix = childPrefix + "└── "
				tomeChildPrefix = childPrefix + "    "
			}

			noteInfo := ""
			if tome.NoteCount > 0 && len(tome.Notes) == 0 {
				// Show count only if notes aren't expanded
				noteInfo = fmt.Sprintf(" (%s)", pluralize(tome.NoteCount, "note", "notes"))
			}
			pinnedMark := ""
			if tome.Pinned {
				pinnedMark = " *"
			}
			focusMark := ""
			if tome.IsFocused {
				focusMark = color.New(color.FgHiMagenta).Sprint(" [FOCUSED]")
			}

			fmt.Printf("%s%s%s%s - %s%s\n", tomePrefix, colorizeID(tome.ID), focusMark, pinnedMark, tome.Title, noteInfo)

			// Expand notes for focused tome/conclave
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

			childIdx++
		}

		// Shipments under conclave
		for _, ship := range con.Shipments {
			isLast := childIdx == totalChildren-1
			shipPrefix := childPrefix + "├── "
			taskPrefix := childPrefix + "│   "
			if isLast {
				shipPrefix = childPrefix + "└── "
				taskPrefix = childPrefix + "    "
			}

			benchMarker := ""
			if ship.BenchID != "" {
				benchMarker = color.New(color.FgCyan).Sprintf(" [%s]", ship.BenchID)
			}
			taskInfo := fmt.Sprintf(" (%d/%d done)", ship.TasksDone, ship.TasksTotal)
			pinnedMark := ""
			if ship.Pinned {
				pinnedMark = " *"
			}
			focusMark := ""
			if ship.IsFocused {
				focusMark = color.New(color.FgHiMagenta).Sprint(" [FOCUSED]")
			}

			fmt.Printf("%s%s%s%s%s - %s%s\n", shipPrefix, colorizeID(ship.ID), benchMarker, focusMark, pinnedMark, ship.Title, taskInfo)

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

			childIdx++
		}

		// Add spacing between conclaves
		if i < len(summary.Conclaves)-1 || summary.Library.TomeCount > 0 || summary.Shipyard.ShipmentCount > 0 {
			fmt.Println("│")
		}
	}

	// Library (always shown)
	libPrefix := "├── "
	libChildPrefix := "│   "
	if summary.Shipyard.ShipmentCount == 0 {
		libPrefix = "└── "
		libChildPrefix = "    "
	}
	fmt.Printf("%s%s (%s)\n", libPrefix, colorizeLabel("LIBRARY"), pluralize(summary.Library.TomeCount, "tome", "tomes"))

	// Expanded library tomes
	if len(summary.Library.Tomes) > 0 {
		for i, tome := range summary.Library.Tomes {
			isLast := i == len(summary.Library.Tomes)-1
			tomePrefix := libChildPrefix + "├── "
			if isLast {
				tomePrefix = libChildPrefix + "└── "
			}
			noteInfo := ""
			if tome.NoteCount > 0 {
				noteInfo = fmt.Sprintf(" (%d notes)", tome.NoteCount)
			}
			pinnedMark := ""
			if tome.Pinned {
				pinnedMark = " *"
			}
			focusMark := ""
			if tome.IsFocused {
				focusMark = color.New(color.FgHiMagenta).Sprint(" [FOCUSED]")
			}
			fmt.Printf("%s%s%s%s - %s%s\n", tomePrefix, colorizeID(tome.ID), focusMark, pinnedMark, tome.Title, noteInfo)
		}
	}

	// Shipyard (always shown)
	fmt.Printf("└── %s (%s)\n", colorizeLabel("SHIPYARD"), pluralize(summary.Shipyard.ShipmentCount, "shipment", "shipments"))

	// Expanded shipyard shipments
	if len(summary.Shipyard.Shipments) > 0 {
		for i, ship := range summary.Shipyard.Shipments {
			isLast := i == len(summary.Shipyard.Shipments)-1
			shipPrefix := "    ├── "
			if isLast {
				shipPrefix = "    └── "
			}
			benchMarker := ""
			if ship.BenchID != "" {
				benchMarker = color.New(color.FgCyan).Sprintf(" [%s]", ship.BenchID)
			}
			taskInfo := fmt.Sprintf(" (%d/%d done)", ship.TasksDone, ship.TasksTotal)
			pinnedMark := ""
			if ship.Pinned {
				pinnedMark = " *"
			}
			focusMark := ""
			if ship.IsFocused {
				focusMark = color.New(color.FgHiMagenta).Sprint(" [FOCUSED]")
			}
			fmt.Printf("%s%s%s%s%s - %s%s\n", shipPrefix, colorizeID(ship.ID), benchMarker, focusMark, pinnedMark, ship.Title, taskInfo)
		}
	}
}

// renderIMPSummary renders the filtered tree view for IMP role
func renderIMPSummary(summary *primary.CommissionSummary, _ string, showAll bool) {
	// Commission header
	fmt.Printf("%s - %s\n", colorizeID(summary.ID), summary.Title)
	fmt.Println("│")

	// Conclaves
	for i, con := range summary.Conclaves {
		isLastConclave := i == len(summary.Conclaves)-1

		prefix := "├── "
		childPrefix := "│   "
		if isLastConclave && summary.Library.TomeCount == 0 && summary.Shipyard.ShipmentCount == 0 && summary.HiddenShipmentCount == 0 {
			prefix = "└── "
			childPrefix = "    "
		}

		focusMarker := ""
		if con.IsFocused {
			focusMarker = color.New(color.FgHiMagenta).Sprint(" [FOCUSED]")
		}
		pinnedMarker := ""
		if con.Pinned {
			pinnedMarker = " *"
		}

		// Tomes under conclave (no note counts for IMP - simpler view)
		totalChildren := len(con.Tomes) + len(con.Shipments)

		// Build counts suffix for conclave
		countParts := []string{}
		if len(con.Tomes) > 0 {
			countParts = append(countParts, pluralize(len(con.Tomes), "tome", "tomes"))
		}
		if len(con.Shipments) > 0 {
			countParts = append(countParts, pluralize(len(con.Shipments), "shipment", "shipments"))
		}
		countSuffix := ""
		if len(countParts) > 0 {
			countSuffix = fmt.Sprintf(" (%s)", strings.Join(countParts, ", "))
		}

		fmt.Printf("%s%s%s%s - %s%s\n", prefix, colorizeID(con.ID), focusMarker, pinnedMarker, con.Title, countSuffix)
		childIdx := 0

		for _, tome := range con.Tomes {
			isLast := childIdx == totalChildren-1
			tomePrefix := childPrefix + "├── "
			tomeChildPrefix := childPrefix + "│   "
			if isLast {
				tomePrefix = childPrefix + "└── "
				tomeChildPrefix = childPrefix + "    "
			}

			noteInfo := ""
			if tome.NoteCount > 0 && len(tome.Notes) == 0 {
				// Show count only if notes aren't expanded
				noteInfo = fmt.Sprintf(" (%s)", pluralize(tome.NoteCount, "note", "notes"))
			}
			pinnedMark := ""
			if tome.Pinned {
				pinnedMark = " *"
			}
			focusMark := ""
			if tome.IsFocused {
				focusMark = color.New(color.FgHiMagenta).Sprint(" [FOCUSED]")
			}

			fmt.Printf("%s%s%s%s - %s%s\n", tomePrefix, colorizeID(tome.ID), focusMark, pinnedMark, tome.Title, noteInfo)

			// Expand notes for focused tome/conclave
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

			childIdx++
		}

		// Shipments under conclave
		for _, ship := range con.Shipments {
			isLast := childIdx == totalChildren-1
			shipPrefix := childPrefix + "├── "
			taskPrefix := childPrefix + "│   "
			if isLast {
				shipPrefix = childPrefix + "└── "
				taskPrefix = childPrefix + "    "
			}

			// IMP sees [FOCUSED] for their focused shipment instead of bench ID
			focusMark := ""
			if ship.IsFocused {
				focusMark = color.New(color.FgHiMagenta).Sprint(" [FOCUSED]")
			}
			taskInfo := fmt.Sprintf(" (%d/%d done)", ship.TasksDone, ship.TasksTotal)
			pinnedMark := ""
			if ship.Pinned {
				pinnedMark = " *"
			}

			fmt.Printf("%s%s%s%s - %s%s\n", shipPrefix, colorizeID(ship.ID), focusMark, pinnedMark, ship.Title, taskInfo)

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

			childIdx++
		}

		// Add spacing between conclaves
		if i < len(summary.Conclaves)-1 || summary.Library.TomeCount > 0 || summary.Shipyard.ShipmentCount > 0 || summary.HiddenShipmentCount > 0 {
			fmt.Println("│")
		}
	}

	// Library (always shown)
	libPrefix := "├── "
	libChildPrefix := "│   "
	if summary.Shipyard.ShipmentCount == 0 && summary.HiddenShipmentCount == 0 {
		libPrefix = "└── "
		libChildPrefix = "    "
	}
	fmt.Printf("%s%s (%s)\n", libPrefix, colorizeLabel("LIBRARY"), pluralize(summary.Library.TomeCount, "tome", "tomes"))

	// Expanded library tomes
	if len(summary.Library.Tomes) > 0 {
		for i, tome := range summary.Library.Tomes {
			isLast := i == len(summary.Library.Tomes)-1
			tomePrefix := libChildPrefix + "├── "
			if isLast {
				tomePrefix = libChildPrefix + "└── "
			}
			pinnedMark := ""
			if tome.Pinned {
				pinnedMark = " *"
			}
			focusMark := ""
			if tome.IsFocused {
				focusMark = color.New(color.FgHiMagenta).Sprint(" [FOCUSED]")
			}
			fmt.Printf("%s%s%s%s - %s\n", tomePrefix, colorizeID(tome.ID), focusMark, pinnedMark, tome.Title)
		}
	}

	// Shipyard (always shown)
	shipyardPrefix := "└── "
	shipyardChildPrefix := "    "
	if summary.HiddenShipmentCount > 0 {
		shipyardPrefix = "├── "
		shipyardChildPrefix = "│   "
	}
	fmt.Printf("%s%s (%s)\n", shipyardPrefix, colorizeLabel("SHIPYARD"), pluralize(summary.Shipyard.ShipmentCount, "shipment", "shipments"))

	// Expanded shipyard shipments
	if len(summary.Shipyard.Shipments) > 0 {
		for i, ship := range summary.Shipyard.Shipments {
			isLast := i == len(summary.Shipyard.Shipments)-1 && summary.HiddenShipmentCount == 0
			shipPrefix := shipyardChildPrefix + "├── "
			if isLast {
				shipPrefix = shipyardChildPrefix + "└── "
			}
			focusMark := ""
			if ship.IsFocused {
				focusMark = color.New(color.FgHiMagenta).Sprint(" [FOCUSED]")
			}
			taskInfo := fmt.Sprintf(" (%d/%d done)", ship.TasksDone, ship.TasksTotal)
			pinnedMark := ""
			if ship.Pinned {
				pinnedMark = " *"
			}
			fmt.Printf("%s%s%s%s - %s%s\n", shipPrefix, colorizeID(ship.ID), focusMark, pinnedMark, ship.Title, taskInfo)
		}
	}

	// Hidden shipments message for IMP
	if summary.HiddenShipmentCount > 0 && !showAll {
		fmt.Printf("└── (%d other shipments hidden - use --all-shipments to show)\n", summary.HiddenShipmentCount)
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

// colorizeLabel applies deterministic color to a label string (e.g., "LIBRARY", "SHIPYARD")
func colorizeLabel(label string) string {
	h := fnv.New32a()
	h.Write([]byte(label))
	hash := h.Sum32()
	colorCode := 16 + (hash % 216)
	c := color.New(color.Attribute(38), color.Attribute(5), color.Attribute(colorCode))
	return c.Sprint(label)
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
