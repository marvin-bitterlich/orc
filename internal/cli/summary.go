package cli

import (
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/example/orc/internal/config"
	orcctx "github.com/example/orc/internal/context"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

// truncate truncates a string to maxLen, replacing newlines with spaces.
func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

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
	ctx := NewContext()

	// Determine container type from ID prefix
	switch {
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
			expandAllCommissions, _ := cmd.Flags().GetBool("expand-all-commissions")

			// Load config for role detection
			cfg, _ := MigrateGoblinConfigIfNeeded(cmd.Context(), cwd)
			role := config.RoleIMP // Default to IMP
			workbenchID := ""
			workshopID := ""

			if cfg != nil && cfg.PlaceID != "" {
				role = config.GetRoleFromPlaceID(cfg.PlaceID)
				if config.IsWorkbench(cfg.PlaceID) {
					workbenchID = cfg.PlaceID
					// Look up workshop from workbench
					if wb, err := wire.WorkbenchService().GetWorkbench(cmd.Context(), cfg.PlaceID); err == nil {
						workshopID = wb.WorkshopID
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
			}

			// DEFAULT BEHAVIOR: When not --all, scope to active commissions derived from focus
			// Active commissions = commissions with focused shipments/tomes/direct focus
			var activeCommissionIDs []string
			if !expandAll && filterCommissionID == "" && workshopID != "" {
				activeCommissionIDs, _ = wire.WorkshopService().GetActiveCommissions(cmd.Context(), workshopID)
			}

			// Get list of commissions to display
			commissions, err := wire.CommissionService().ListCommissions(context.Background(), primary.CommissionFilters{})
			if err != nil {
				return fmt.Errorf("failed to list commissions: %w", err)
			}

			// Build set of active commission IDs for efficient lookup
			activeSet := make(map[string]bool)
			for _, id := range activeCommissionIDs {
				activeSet[id] = true
			}

			// Filter to open commissions
			var openCommissions []*primary.Commission
			for _, m := range commissions {
				if m.Status == "complete" || m.Status == "archived" {
					continue
				}
				// Apply explicit filter if specified
				if filterCommissionID != "" && m.ID != filterCommissionID {
					continue
				}
				// Apply active commissions filter if derived from focus
				if len(activeCommissionIDs) > 0 && !activeSet[m.ID] {
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

			// Determine which commission is "focused" based on focusID
			focusedCommissionID := ""
			if focusID != "" {
				focusedCommissionID = resolveContainerCommission(focusID)
			}

			// Sort commissions: focused commission first, then others by ID
			sort.SliceStable(openCommissions, func(i, j int) bool {
				isFocusedI := openCommissions[i].ID == focusedCommissionID
				isFocusedJ := openCommissions[j].ID == focusedCommissionID
				if isFocusedI != isFocusedJ {
					return isFocusedI // Focused commission first
				}
				return openCommissions[i].ID < openCommissions[j].ID
			})

			// Render header based on role
			renderHeader(role, workbenchID, workshopID, focusID, filterCommissionID)

			// Build map of focused containers across all workbenches in this workshop
			workshopFocus := buildWorkshopFocusMap(cmd.Context(), workshopID, workbenchID)

			// Display each commission
			for i, commission := range openCommissions {
				isFocusedCommission := commission.ID == focusedCommissionID
				shouldExpand := isFocusedCommission || expandAllCommissions

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

				if shouldExpand {
					// Render full summary for focused or expanded commissions
					renderSummary(summary, focusID, workshopFocus)

					// Render debug info if present
					if summary.DebugInfo != nil && len(summary.DebugInfo.Messages) > 0 {
						fmt.Println()
						renderDebugInfo(summary.DebugInfo)
					}
				} else {
					// Render collapsed summary for non-focused commissions
					renderCollapsedCommission(summary)
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
	cmd.Flags().Bool("expand-all-commissions", false, "Expand all commissions (default: only focused commission expanded)")

	return cmd
}

// renderHeader prints the header line based on role
func renderHeader(role, workbenchID, workshopID, _, commissionID string) {
	// Show workshop context
	if workshopID != "" {
		// Fetch workshop name
		workshopName := ""
		if ws, err := wire.WorkshopService().GetWorkshop(NewContext(), workshopID); err == nil {
			workshopName = ws.Name
		}

		fmt.Printf("Workshop %s", workshopID)
		if workshopName != "" {
			fmt.Printf(" (%s)", workshopName)
		}
		fmt.Println()

		// Show workbenches in workshop with tree format
		renderWorkshopBenches(workshopID, workbenchID)
	}

	fmt.Println()
}

// renderWorkshopBenches displays workbenches in the workshop with tree formatting
func renderWorkshopBenches(workshopID, currentWorkbenchID string) {
	ctx := NewContext()

	allWorkbenches, err := wire.WorkbenchService().ListWorkbenches(ctx, primary.WorkbenchFilters{
		WorkshopID: workshopID,
	})
	if err != nil {
		return
	}

	// Filter to active workbenches only
	var workbenches []*primary.Workbench
	for _, wb := range allWorkbenches {
		if wb.Status == "active" {
			workbenches = append(workbenches, wb)
		}
	}

	totalItems := len(workbenches)
	if totalItems == 0 {
		return
	}

	fmt.Println("|")
	itemIdx := 0

	// Render workbenches
	for _, wb := range workbenches {
		isLast := itemIdx == totalItems-1
		prefix := "├── "
		if isLast {
			prefix = "└── "
		}

		// Build workbench line with optional focus indicator
		line := fmt.Sprintf("👹 %s", wb.Name)
		if wb.ID == currentWorkbenchID {
			line = color.New(color.FgHiMagenta).Sprint(line)
		}

		// Add git branch and dirty status (colored branch name)
		if wb.Path != "" {
			branch, dirty, err := getGitBranchStatus(wb.Path)
			if err != nil {
				line += color.New(color.FgHiBlack).Sprint(" [?]")
			} else if dirty {
				line += color.New(color.FgYellow).Sprintf(" [%s]", branch)
			} else {
				line += color.New(color.FgGreen).Sprintf(" [%s]", branch)
			}
		}

		// Add focused shipment inline if workbench has one (skip stale focus)
		focusedID, _ := wire.WorkbenchService().GetFocusedID(ctx, wb.ID)
		if focusedID != "" && !isStaleFocus(ctx, focusedID) {
			line += color.New(color.FgCyan).Sprintf(" → %s", focusedID)
		}

		fmt.Printf("%s%s\n", prefix, line)
		itemIdx++
	}
}

// isStaleFocus returns true if the focus points to a completed/deployed shipment
func isStaleFocus(ctx context.Context, focusedID string) bool {
	if !strings.HasPrefix(focusedID, "SHIP-") {
		return false // Only check shipments
	}
	ship, err := wire.ShipmentService().GetShipment(ctx, focusedID)
	if err != nil {
		return true // Can't find it, consider stale
	}
	// Terminal state: closed
	return ship.Status == "closed"
}

// getGitBranchStatus returns the current git branch and dirty status for a path.
// Returns branch name, whether it's dirty, and any error.
func getGitBranchStatus(path string) (branch string, dirty bool, err error) {
	// Get current branch
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return "", false, err
	}
	branch = strings.TrimSpace(string(out))

	// Check dirty status using diff-index (10x faster than git status on large repos)
	cmd = exec.Command("git", "diff-index", "--quiet", "HEAD", "--")
	cmd.Dir = path
	err = cmd.Run()
	// Exit 0 = clean, Exit 1 = dirty, other = error
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			dirty = true
		} else {
			// Any other error (no commits, not a repo, etc.) - return error
			return branch, false, err
		}
	}

	return branch, dirty, nil
}

// workshopFocusInfo tracks what each workbench in the workshop has focused
type workshopFocusInfo struct {
	containerToWorkbench map[string][]string // containerID -> list of actors focusing it
	myName               string              // current workbench name
	myID                 string              // current workbench ID
}

// buildWorkshopFocusMap fetches focus for all workbenches in the workshop
func buildWorkshopFocusMap(ctx context.Context, workshopID, currentWorkbenchID string) workshopFocusInfo {
	info := workshopFocusInfo{
		containerToWorkbench: make(map[string][]string),
		myID:                 currentWorkbenchID,
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

	// Get each IMP's focus (only active workbenches)
	allWorkbenches, err := wire.WorkbenchService().ListWorkbenches(ctx, primary.WorkbenchFilters{
		WorkshopID: workshopID,
	})
	if err != nil {
		return info
	}

	for _, wb := range allWorkbenches {
		// Skip archived workbenches
		if wb.Status != "active" {
			continue
		}
		// Skip our own workbench - handled separately for [FOCUSED] marker
		if wb.ID == currentWorkbenchID {
			continue
		}

		focusedID, err := wire.WorkbenchService().GetFocusedID(ctx, wb.ID)
		if err != nil || focusedID == "" || isStaleFocus(ctx, focusedID) {
			continue
		}

		info.containerToWorkbench[focusedID] = append(info.containerToWorkbench[focusedID], fmt.Sprintf("%s@%s", wb.Name, wb.ID))
	}

	return info
}

// formatFocusActors formats the focus marker for multi-actor display
// Order: "you" first (if isMeFocused), then others alphabetically
func formatFocusActors(actors []string, isMeFocused bool) string {
	var parts []string

	// "you" first if current actor is focusing
	if isMeFocused {
		parts = append(parts, color.New(color.FgHiMagenta).Sprint("you"))
	}

	// Others (sorted alphabetically, cyan)
	others := append([]string{}, actors...)
	sort.Strings(others)

	for _, o := range others {
		parts = append(parts, color.New(color.FgCyan).Sprint(o))
	}

	if len(parts) == 0 {
		return ""
	}
	return fmt.Sprintf(" [focused by %s]", strings.Join(parts, ", "))
}

// renderCollapsedCommission renders a commission as a single collapsed line with counts
func renderCollapsedCommission(summary *primary.CommissionSummary) {
	// Count items
	shipmentCount := len(summary.Shipments)
	noteCount := len(summary.Notes)
	tomeCount := len(summary.Tomes)

	// Build counts string
	var counts []string
	if shipmentCount > 0 {
		counts = append(counts, pluralize(shipmentCount, "shipment", "shipments"))
	}
	if noteCount > 0 {
		counts = append(counts, pluralize(noteCount, "note", "notes"))
	}
	if tomeCount > 0 {
		counts = append(counts, pluralize(tomeCount, "tome", "tomes"))
	}

	countsStr := ""
	if len(counts) > 0 {
		countsStr = fmt.Sprintf(" (%s)", strings.Join(counts, ", "))
	}

	fmt.Printf("%s - %s%s\n", colorizeID(summary.ID), summary.Title, countsStr)
}

// renderSummary renders the commission with notes, shipments, and tomes in tree format
func renderSummary(summary *primary.CommissionSummary, _ string, workshopFocus workshopFocusInfo) {
	// Commission header with focused marker
	focusedMarker := ""
	if summary.IsFocusedCommission {
		focusedMarker = fmt.Sprintf(" [focused by ✨ %s ✨]", color.New(color.FgHiMagenta).Sprint("you"))
	}
	fmt.Printf("%s%s - %s\n", colorizeID(summary.ID), focusedMarker, summary.Title)

	// Split shipments into focused and non-focused groups
	var focusedShips, otherShips []primary.ShipmentSummary
	for _, ship := range summary.Shipments {
		if ship.IsFocused || len(workshopFocus.containerToWorkbench[ship.ID]) > 0 {
			focusedShips = append(focusedShips, ship)
		} else {
			otherShips = append(otherShips, ship)
		}
	}

	// Sort focused shipments: YOUR focus first, then others
	sort.SliceStable(focusedShips, func(i, j int) bool {
		// IsFocused means "focused by you" - put these first
		if focusedShips[i].IsFocused != focusedShips[j].IsFocused {
			return focusedShips[i].IsFocused
		}
		return false // Keep original order otherwise
	})

	// Calculate total items for tree rendering
	totalItems := len(summary.Notes) + len(focusedShips) + len(otherShips) + len(summary.Tomes)
	if totalItems == 0 {
		return
	}

	fmt.Println("│")
	itemIdx := 0

	// 1. Render focused shipments
	for _, ship := range focusedShips {
		renderShipment(ship, workshopFocus, &itemIdx, totalItems)
	}

	// Visual gap between focused and non-focused shipments
	if len(focusedShips) > 0 && (len(otherShips) > 0 || len(summary.Tomes) > 0 || len(summary.Notes) > 0) {
		fmt.Println("│")
	}

	// 2. Render non-focused shipments
	for _, ship := range otherShips {
		renderShipment(ship, workshopFocus, &itemIdx, totalItems)
	}

	// 3. Render tomes
	for _, tome := range summary.Tomes {
		isLast := itemIdx == totalItems-1 && len(summary.Notes) == 0
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
		focusMark := formatFocusActors(workshopFocus.containerToWorkbench[tome.ID], tome.IsFocused)

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
				fmt.Printf("%s%s %s- %s\n", notePrefix, colorizeID(note.ID), typeMarker, truncate(note.Title, 60))
			}
		}

		itemIdx++
	}

	// Visual gap before commission-level notes
	if len(summary.Notes) > 0 && (len(focusedShips) > 0 || len(otherShips) > 0 || len(summary.Tomes) > 0) {
		fmt.Println("│")
	}

	// 4. Render commission-level notes as tree items (after shipments and tomes)
	for i, note := range summary.Notes {
		isLast := i == len(summary.Notes)-1
		prefix := "├── "
		if isLast {
			prefix = "└── "
		}
		pinnedMark := ""
		if note.Pinned {
			pinnedMark = " 📌"
		}
		typeMarker := ""
		if note.Type != "" {
			typeMarker = color.New(color.FgYellow).Sprintf(" [%s]", note.Type)
		}
		fmt.Printf("%s%s%s%s - %s\n", prefix, colorizeID(note.ID), typeMarker, pinnedMark, truncate(note.Title, 60))
		itemIdx++
	}
}

// renderShipment renders a single shipment with its children if focused
func renderShipment(ship primary.ShipmentSummary, workshopFocus workshopFocusInfo, itemIdx *int, totalItems int) {
	isLast := *itemIdx == totalItems-1
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
	if ship.Status != "" && ship.Status != "closed" {
		statusBadge = " " + colorizeShipmentStatus(ship.Status)
	}
	taskInfo := fmt.Sprintf(" (%d/%d done", ship.TasksDone, ship.TasksTotal)
	if ship.NoteCount > 0 {
		taskInfo += fmt.Sprintf(", %s", pluralize(ship.NoteCount, "note", "notes"))
	}
	taskInfo += ")"
	pinnedMark := ""
	if ship.Pinned {
		pinnedMark = " *"
	}
	focusMark := formatFocusActors(workshopFocus.containerToWorkbench[ship.ID], ship.IsFocused)

	fmt.Printf("%s%s%s%s%s%s - %s%s\n", prefix, colorizeID(ship.ID), statusBadge, benchMarker, focusMark, pinnedMark, ship.Title, taskInfo)

	// Expand children for focused shipment (notes first, then tasks)
	if ship.IsFocused {
		totalChildren := len(ship.Notes) + len(ship.Tasks)
		childIdx := 0

		// Render notes first (context)
		for _, note := range ship.Notes {
			isLastChild := childIdx == totalChildren-1
			nPrefix := taskPrefix + "├── "
			if isLastChild {
				nPrefix = taskPrefix + "└── "
			}
			typeMarker := ""
			if note.Type != "" {
				typeMarker = color.New(color.FgYellow).Sprintf("[%s] ", note.Type)
			}
			fmt.Printf("%s%s %s- %s\n", nPrefix, colorizeID(note.ID), typeMarker, truncate(note.Title, 60))
			childIdx++
		}

		// Render tasks second (work)
		for _, task := range ship.Tasks {
			isLastChild := childIdx == totalChildren-1
			tPrefix := taskPrefix + "├── "
			taskChildPrefix := taskPrefix + "│   "
			if isLastChild {
				tPrefix = taskPrefix + "└── "
				taskChildPrefix = taskPrefix + "    "
			}
			statusMark := ""
			if task.Status != "" && task.Status != "open" {
				statusMark = colorizeStatus(task.Status) + " - "
			}
			fmt.Printf("%s%s - %s%s\n", tPrefix, colorizeID(task.ID), statusMark, task.Title)
			// Render task children (plans, approvals, escalations, receipts)
			renderTaskChildren(task, taskChildPrefix)
			childIdx++
		}
	}

	*itemIdx++
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
	if status == "" || status == "open" {
		return ""
	}
	upper := strings.ToUpper(status)

	switch status {
	case "in-progress":
		return color.New(color.FgHiBlue).Sprint(upper)
	case "blocked":
		return color.New(color.FgRed).Sprint(upper)
	case "closed":
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
	case "ready":
		return color.New(color.FgHiYellow).Sprint("[ready]")
	case "in-progress":
		return color.New(color.FgHiBlue).Sprint("[in-progress]")
	case "closed":
		return color.New(color.FgHiGreen).Sprint("[closed]")
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
	default:
		return upper // draft
	}
}

// renderTaskChildren renders the child entities (plans) under a task
func renderTaskChildren(task primary.TaskSummary, prefix string) {
	totalChildren := len(task.Plans)
	if totalChildren == 0 {
		return
	}

	for i, plan := range task.Plans {
		childPrefix := prefix + "├── "
		if i == totalChildren-1 {
			childPrefix = prefix + "└── "
		}
		fmt.Printf("%s%s %s\n", childPrefix, colorizeID(plan.ID), colorizePlanStatus(plan.Status))
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
