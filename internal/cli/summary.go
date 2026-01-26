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
	ctx "github.com/example/orc/internal/context"
	"github.com/example/orc/internal/models"
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

// filterConfig holds all filtering settings for the summary display
type filterConfig struct {
	statusMap      map[string]bool // statuses to hide
	containerTypes map[string]bool // container types to show (empty = all)
	includeTags    map[string]bool // tags to include (empty = all)
	excludeTags    map[string]bool // tags to exclude
	leafTypes      map[string]bool // leaf types to show (empty = all): tasks, notes, questions, plans
}

// getTagColor returns a deterministic color for a tag name
// Uses FNV-1a hash to map tag name to one of 256 colors
func getTagColor(tagName string) *color.Color {
	// Hash the tag name
	h := fnv.New32a()
	h.Write([]byte(tagName))
	hash := h.Sum32()

	// Map to 256-color range (16-231 are the color cube)
	// Avoid 0-15 (basic colors) and 232-255 (grayscale) for better visibility
	colorCode := 16 + (hash % 216) // 216 colors in the color cube

	// Return color with 256-color mode
	return color.New(color.Attribute(38), color.Attribute(5), color.Attribute(colorCode))
}

// colorizeTag wraps a tag name in brackets with deterministic color
func colorizeTag(tagName string) string {
	c := getTagColor(tagName)
	return c.Sprintf("[%s]", tagName)
}

// colorizeWorkbench wraps workbench info with color (cyan for visibility)
func colorizeWorkbench(workbenchName, workbenchID string) string {
	c := color.New(color.FgCyan)
	return c.Sprintf("[Workbench: %s (%s)]", workbenchName, workbenchID)
}

// getIDColor returns a deterministic color for an ID type (TASK, SHIP, COMM)
// Uses FNV-1a hash on the ID prefix to ensure all IDs of same type have same color
func getIDColor(idType string) *color.Color {
	// Hash the ID type (e.g., "TASK", "SHIP", "COMM")
	h := fnv.New32a()
	h.Write([]byte(idType))
	hash := h.Sum32()

	// Map to 256-color range (16-231 are the color cube)
	colorCode := 16 + (hash % 216)

	return color.New(color.Attribute(38), color.Attribute(5), color.Attribute(colorCode))
}

// colorizeID applies deterministic color to an ID based on its prefix
// Example: "TASK-165" → colored "TASK-165" (all TASK-IDs same color)
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

// colorizeStatus formats status as uppercase with semantic color
// Returns empty string for "ready" status (default, not shown)
func colorizeStatus(status string) string {
	if status == "" || status == "ready" || status == "active" {
		return "" // Don't show empty or default statuses
	}
	upper := strings.ToUpper(status)

	// Inline color application (same pattern as [FOCUSED] which works)
	switch status {
	case "needs_design":
		return color.New(color.FgHiYellow).Sprint(upper)
	case "ready_to_implement":
		return color.New(color.FgHiCyan).Sprint(upper)
	case "paused", "draft":
		return color.New(color.FgYellow).Sprint(upper)
	case "blocked":
		return color.New(color.FgRed).Sprint(upper)
	case "design", "open":
		return color.New(color.FgCyan).Sprint(upper)
	case "awaiting_approval", "submitted":
		return color.New(color.FgMagenta).Sprint(upper)
	case "complete", "verified", "merged":
		return color.New(color.FgHiGreen).Sprint(upper)
	case "in_progress", "implement":
		return color.New(color.FgHiBlue).Sprint(upper)
	case "archived", "closed":
		return color.New(color.FgHiBlack).Sprint(upper)
	default:
		return color.New(color.FgWhite).Sprint(upper)
	}
}

// getEntityType returns the entity type string for tag lookups based on ID prefix
func getEntityType(id string) string {
	parts := strings.Split(id, "-")
	if len(parts) < 1 {
		return ""
	}
	switch parts[0] {
	case "TASK":
		return "task"
	case "PLAN":
		return "plan"
	case "NOTE":
		return "note"
	case "WO":
		return "work_order"
	case "CWO":
		return "cycle_work_order"
	case "CREC":
		return "cycle_receipt"
	case "REC":
		return "receipt"
	default:
		return ""
	}
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
	case strings.HasPrefix(containerID, "INV-"):
		if inv, err := wire.InvestigationService().GetInvestigation(ctx, containerID); err == nil {
			return inv.CommissionID
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

// shouldShowLeaf checks if a leaf item should be shown based on tag filters
// Returns (show bool, tagName string)
func shouldShowLeaf(entityID string, filters *filterConfig) (bool, string) {
	entityType := getEntityType(entityID)
	if entityType == "" {
		return true, "" // Unknown type, show it
	}

	// Check leaf type filter (hide specified types)
	if len(filters.leafTypes) > 0 {
		// Support both singular (task) and plural (tasks) forms
		if filters.leafTypes[entityType] || filters.leafTypes[entityType+"s"] {
			return false, ""
		}
	}

	tag, _ := wire.TagService().GetEntityTag(context.Background(), entityID, entityType)
	tagName := ""
	if tag != nil {
		tagName = tag.Name
	}

	// If include tags specified, must match one
	if len(filters.includeTags) > 0 {
		if tagName == "" || !filters.includeTags[tagName] {
			return false, tagName
		}
	}

	// If exclude tags specified, must not match any
	if filters.excludeTags[tagName] && tagName != "" {
		return false, tagName
	}

	return true, tagName
}

// displayShipmentChildren shows Spec-Kit artifacts, tasks, and notes under a shipment
// Returns count of visible items
func displayShipmentChildren(shipmentID, prefix string, filters *filterConfig) int {
	ctx := context.Background()
	visibleCount := 0

	// Collect all items to display (to calculate isLast correctly)
	type displayItem struct {
		render func(isLast bool)
	}
	var items []displayItem

	// 1. WorkOrder (1:1 with shipment)
	wo, _ := wire.WorkOrderService().GetWorkOrderByShipment(ctx, shipmentID)
	if wo != nil {
		items = append(items, displayItem{
			render: func(isLast bool) {
				childPrefix := prefix + "├── "
				if isLast {
					childPrefix = prefix + "└── "
				}
				outcome := truncate(wo.Outcome, 50)
				statusInfo := colorizeStatus(wo.Status)
				if statusInfo != "" {
					fmt.Printf("%s%s - %s - %s\n", childPrefix, colorizeID(wo.ID), statusInfo, outcome)
				} else {
					fmt.Printf("%s%s - %s\n", childPrefix, colorizeID(wo.ID), outcome)
				}
			},
		})
	}

	// 2. Receipt (1:1 with shipment)
	rec, _ := wire.ReceiptService().GetReceiptByShipment(ctx, shipmentID)
	if rec != nil {
		items = append(items, displayItem{
			render: func(isLast bool) {
				childPrefix := prefix + "├── "
				if isLast {
					childPrefix = prefix + "└── "
				}
				outcome := truncate(rec.DeliveredOutcome, 50)
				statusInfo := colorizeStatus(rec.Status)
				if statusInfo != "" {
					fmt.Printf("%s%s - %s - %s\n", childPrefix, colorizeID(rec.ID), statusInfo, outcome)
				} else {
					fmt.Printf("%s%s - %s\n", childPrefix, colorizeID(rec.ID), outcome)
				}
			},
		})
	}

	// 3. Cycles (1:many with shipment) - with nested CWO and CREC
	cycles, _ := wire.CycleService().ListCycles(ctx, primary.CycleFilters{ShipmentID: shipmentID})
	for _, cycle := range cycles {
		// Capture cycle in closure
		c := cycle
		items = append(items, displayItem{
			render: func(isLast bool) {
				childPrefix := prefix + "├── "
				nestedPrefix := prefix + "│   "
				if isLast {
					childPrefix = prefix + "└── "
					nestedPrefix = prefix + "    "
				}
				// Format cycle ID as SHIP-XXX-C# for display
				cycleDisplayID := fmt.Sprintf("%s-C%d", shipmentID, c.SequenceNumber)
				statusInfo := colorizeStatus(c.Status)
				if statusInfo != "" {
					fmt.Printf("%s%s - %s\n", childPrefix, colorizeID(cycleDisplayID), statusInfo)
				} else {
					fmt.Printf("%s%s\n", childPrefix, colorizeID(cycleDisplayID))
				}

				// Get cycle children: Plans, CWO, CREC
				cyclePlans, _ := wire.PlanService().ListPlans(ctx, primary.PlanFilters{CycleID: c.ID})
				cwo, _ := wire.CycleWorkOrderService().GetCycleWorkOrderByCycle(ctx, c.ID)
				var crec *primary.CycleReceipt
				if cwo != nil {
					crec, _ = wire.CycleReceiptService().GetCycleReceiptByCWO(ctx, cwo.ID)
				}

				// Determine what children exist for prefix logic
				hasCWO := cwo != nil
				hasCREC := crec != nil

				// Nested: Plans (linked to cycle)
				for i, plan := range cyclePlans {
					planTitle := truncate(plan.Title, 40)
					planStatus := colorizeStatus(plan.Status)
					isLastPlan := i == len(cyclePlans)-1 && !hasCWO
					planPrefix := nestedPrefix + "├── "
					if isLastPlan {
						planPrefix = nestedPrefix + "└── "
					}
					if planStatus != "" {
						fmt.Printf("%s%s - %s - %s\n", planPrefix, colorizeID(plan.ID), planStatus, planTitle)
					} else {
						fmt.Printf("%s%s - %s\n", planPrefix, colorizeID(plan.ID), planTitle)
					}
				}

				// Nested: CWO (1:1 with cycle)
				if cwo != nil {
					cwoOutcome := truncate(cwo.Outcome, 40)
					cwoStatus := colorizeStatus(cwo.Status)
					cwoPrefix := nestedPrefix + "├── "
					if !hasCREC {
						cwoPrefix = nestedPrefix + "└── "
					}
					if cwoStatus != "" {
						fmt.Printf("%s%s - %s - %s\n", cwoPrefix, colorizeID(cwo.ID), cwoStatus, cwoOutcome)
					} else {
						fmt.Printf("%s%s - %s\n", cwoPrefix, colorizeID(cwo.ID), cwoOutcome)
					}

					// Nested: CREC (1:1 with CWO)
					if crec != nil {
						crecOutcome := truncate(crec.DeliveredOutcome, 40)
						crecStatus := colorizeStatus(crec.Status)
						crecPrefix := nestedPrefix + "└── "
						if crecStatus != "" {
							fmt.Printf("%s%s - %s - %s\n", crecPrefix, colorizeID(crec.ID), crecStatus, crecOutcome)
						} else {
							fmt.Printf("%s%s - %s\n", crecPrefix, colorizeID(crec.ID), crecOutcome)
						}
					}
				}
			},
		})
	}

	// 4. Tasks
	shipmentTasks, _ := wire.ShipmentService().GetShipmentTasks(ctx, shipmentID)
	for _, t := range shipmentTasks {
		if t.Status == "complete" || filters.statusMap[t.Status] {
			continue
		}
		show, tagName := shouldShowLeaf(t.ID, filters)
		if !show {
			continue
		}
		// Capture in closure
		task := t
		tag := tagName
		items = append(items, displayItem{
			render: func(isLast bool) {
				childPrefix := prefix + "├── "
				if isLast {
					childPrefix = prefix + "└── "
				}
				tagInfo := ""
				if tag != "" {
					tagInfo = " " + colorizeTag(tag)
				}
				statusInfo := colorizeStatus(task.Status)
				if statusInfo != "" {
					fmt.Printf("%s%s - %s - %s%s\n", childPrefix, colorizeID(task.ID), statusInfo, task.Title, tagInfo)
				} else {
					fmt.Printf("%s%s - %s%s\n", childPrefix, colorizeID(task.ID), task.Title, tagInfo)
				}
			},
		})
	}

	// 5. Notes (filter out closed)
	serviceNotes, _ := wire.NoteService().GetNotesByContainer(ctx, "shipment", shipmentID)
	for _, n := range serviceNotes {
		if n.Status == "closed" {
			continue
		}
		show, tagName := shouldShowLeaf(n.ID, filters)
		if !show {
			continue
		}
		// Capture in closure
		note := n
		tag := tagName
		items = append(items, displayItem{
			render: func(isLast bool) {
				pinnedEmoji := ""
				if note.Pinned {
					pinnedEmoji = "📌 "
				}
				childPrefix := prefix + "├── "
				if isLast {
					childPrefix = prefix + "└── "
				}
				tagInfo := ""
				if tag != "" {
					tagInfo = " " + colorizeTag(tag)
				}
				fmt.Printf("%s%s%s - %s%s\n", childPrefix, pinnedEmoji, colorizeID(note.ID), note.Title, tagInfo)
			},
		})
	}

	// Render all items
	for i, item := range items {
		isLast := i == len(items)-1
		item.render(isLast)
		visibleCount++
	}

	return visibleCount
}

// displayConclaveChildren shows tomes (with notes), tasks, plans, and unfiled notes under a conclave
// Returns count of visible items
func displayConclaveChildren(conclaveID, prefix string, filters *filterConfig) int {
	ctx := context.Background()
	visibleCount := 0

	// 1. Get tomes belonging to this conclave
	tomes, _ := wire.TomeService().ListTomes(ctx, primary.TomeFilters{ConclaveID: conclaveID})
	var openTomes []*primary.Tome
	for _, t := range tomes {
		if t.Status != "closed" && !filters.statusMap[t.Status] {
			openTomes = append(openTomes, t)
		}
	}

	// 2. Get tasks via ConclaveService
	serviceTasks, _ := wire.ConclaveService().GetConclaveTasks(ctx, conclaveID)
	var tasks []*models.Task
	for _, t := range serviceTasks {
		tasks = append(tasks, &models.Task{
			ID:     t.ID,
			Title:  t.Title,
			Status: t.Status,
			Pinned: t.Pinned,
		})
	}

	// 3. Get plans via ConclaveService
	servicePlans, _ := wire.ConclaveService().GetConclavePlans(ctx, conclaveID)
	var plans []*models.Plan
	for _, p := range servicePlans {
		plans = append(plans, &models.Plan{
			ID:     p.ID,
			Title:  p.Title,
			Status: p.Status,
			Pinned: p.Pinned,
		})
	}

	// 4. Get unfiled notes (notes with conclave_id but no tome_id - the repo already filters these)
	serviceNotes, _ := wire.NoteService().GetNotesByContainer(ctx, "conclave", conclaveID)
	var notes []*models.Note
	for _, n := range serviceNotes {
		if n.Status == "closed" {
			continue
		}
		notes = append(notes, &models.Note{
			ID:     n.ID,
			Title:  n.Title,
			Pinned: n.Pinned,
		})
	}

	// Collect all items to display
	type displayItem struct {
		render func(isLast bool)
	}
	var items []displayItem

	// Add tomes (with their nested notes)
	for _, tome := range openTomes {
		t := tome // capture for closure
		items = append(items, displayItem{
			render: func(isLast bool) {
				pinnedEmoji := ""
				if t.Pinned {
					pinnedEmoji = "📌 "
				}
				childPrefix := prefix + "├── "
				nestedPrefix := prefix + "│   "
				if isLast {
					childPrefix = prefix + "└── "
					nestedPrefix = prefix + "    "
				}
				statusInfo := colorizeStatus(t.Status)
				if statusInfo != "" {
					fmt.Printf("%s%s%s - %s - %s\n", childPrefix, pinnedEmoji, colorizeID(t.ID), statusInfo, t.Title)
				} else {
					fmt.Printf("%s%s%s - %s\n", childPrefix, pinnedEmoji, colorizeID(t.ID), t.Title)
				}
				// Display tome's notes
				displayTomeChildren(t.ID, nestedPrefix, filters)
			},
		})
	}

	// Collect regular children (tasks, plans, unfiled notes) with tag filtering
	type childItem struct {
		id      string
		title   string
		status  string
		pinned  bool
		tagName string
	}
	var regularChildren []childItem
	hiddenCount := 0

	for _, t := range tasks {
		if t.Status == "complete" || filters.statusMap[t.Status] {
			continue
		}
		show, tagName := shouldShowLeaf(t.ID, filters)
		if show {
			regularChildren = append(regularChildren, childItem{t.ID, t.Title, t.Status, t.Pinned, tagName})
		} else {
			hiddenCount++
		}
	}
	for _, p := range plans {
		if p.Status == "approved" || filters.statusMap[p.Status] {
			continue
		}
		show, tagName := shouldShowLeaf(p.ID, filters)
		if show {
			regularChildren = append(regularChildren, childItem{p.ID, p.Title, p.Status, p.Pinned, tagName})
		} else {
			hiddenCount++
		}
	}
	for _, n := range notes {
		show, tagName := shouldShowLeaf(n.ID, filters)
		if show {
			regularChildren = append(regularChildren, childItem{n.ID, n.Title, "", n.Pinned, tagName})
		} else {
			hiddenCount++
		}
	}

	// Add regular children as display items
	for i, child := range regularChildren {
		c := child
		idx := i
		items = append(items, displayItem{
			render: func(isLast bool) {
				pinnedEmoji := ""
				if c.pinned {
					pinnedEmoji = "📌 "
				}
				// Calculate actual isLast considering remaining items
				actualIsLast := isLast && idx == len(regularChildren)-1 && hiddenCount == 0
				childPrefix := prefix + "├── "
				if actualIsLast {
					childPrefix = prefix + "└── "
				}
				tagInfo := ""
				if c.tagName != "" {
					tagInfo = " " + colorizeTag(c.tagName)
				}
				statusInfo := colorizeStatus(c.status)
				if statusInfo != "" {
					fmt.Printf("%s%s%s - %s - %s%s\n", childPrefix, pinnedEmoji, colorizeID(c.id), statusInfo, c.title, tagInfo)
				} else {
					fmt.Printf("%s%s%s - %s%s\n", childPrefix, pinnedEmoji, colorizeID(c.id), c.title, tagInfo)
				}
			},
		})
	}

	// Render all items
	for i, item := range items {
		isLast := i == len(items)-1 && hiddenCount == 0
		item.render(isLast)
		visibleCount++
	}

	// Show hidden count if any
	if hiddenCount > 0 {
		fmt.Printf("%s└── (%d other items)\n", prefix, hiddenCount)
	}

	return visibleCount
}

// displayInvestigationChildren shows notes under an investigation with tag filtering
// Returns count of visible items
func displayInvestigationChildren(investigationID, prefix string, filters *filterConfig) int {
	serviceNotes, _ := wire.NoteService().GetNotesByContainer(context.Background(), "investigation", investigationID)
	// Convert to models.Note for the rest of the function (filter out closed notes)
	var notes []*models.Note
	for _, n := range serviceNotes {
		if n.Status == "closed" {
			continue
		}
		notes = append(notes, &models.Note{
			ID:     n.ID,
			Title:  n.Title,
			Pinned: n.Pinned,
		})
	}

	// Collect all children with tag filtering
	type childItem struct {
		id      string
		title   string
		status  string
		pinned  bool
		tagName string
	}
	var visible []childItem
	hiddenCount := 0

	for _, n := range notes {
		show, tagName := shouldShowLeaf(n.ID, filters)
		if show {
			visible = append(visible, childItem{n.ID, n.Title, "", n.Pinned, tagName})
		} else {
			hiddenCount++
		}
	}

	for k, child := range visible {
		pinnedEmoji := ""
		if child.pinned {
			pinnedEmoji = "📌 "
		}
		isLast := k == len(visible)-1 && hiddenCount == 0
		childPrefix := prefix + "├── "
		if isLast {
			childPrefix = prefix + "└── "
		}
		tagInfo := ""
		if child.tagName != "" {
			tagInfo = " " + colorizeTag(child.tagName)
		}
		statusInfo := colorizeStatus(child.status)
		if statusInfo != "" {
			fmt.Printf("%s%s%s - %s - %s%s\n", childPrefix, pinnedEmoji, colorizeID(child.id), statusInfo, child.title, tagInfo)
		} else {
			fmt.Printf("%s%s%s - %s%s\n", childPrefix, pinnedEmoji, colorizeID(child.id), child.title, tagInfo)
		}
	}

	// Show hidden count if any
	if hiddenCount > 0 {
		fmt.Printf("%s└── (%d other items)\n", prefix, hiddenCount)
	}

	return len(visible)
}

// displayTomeChildren shows notes under a tome with tag filtering
// Returns count of visible notes
func displayTomeChildren(tomeID, prefix string, filters *filterConfig) int {
	serviceNotes, err := wire.TomeService().GetTomeNotes(context.Background(), tomeID)
	if err != nil || len(serviceNotes) == 0 {
		return 0
	}
	// Convert to models.Note for the rest of the function (filter out closed notes)
	var notes []*models.Note
	for _, n := range serviceNotes {
		if n.Status == "closed" {
			continue
		}
		notes = append(notes, &models.Note{
			ID:     n.ID,
			Title:  n.Title,
			Pinned: n.Pinned,
		})
	}

	// Collect all children with tag filtering
	type childItem struct {
		id      string
		title   string
		pinned  bool
		tagName string
	}
	var visible []childItem
	hiddenCount := 0

	for _, n := range notes {
		show, tagName := shouldShowLeaf(n.ID, filters)
		if show {
			visible = append(visible, childItem{n.ID, n.Title, n.Pinned, tagName})
		} else {
			hiddenCount++
		}
	}

	for k, child := range visible {
		pinnedEmoji := ""
		if child.pinned {
			pinnedEmoji = "📌 "
		}
		isLast := k == len(visible)-1 && hiddenCount == 0
		childPrefix := prefix + "├── "
		if isLast {
			childPrefix = prefix + "└── "
		}
		tagInfo := ""
		if child.tagName != "" {
			tagInfo = " " + colorizeTag(child.tagName)
		}
		fmt.Printf("%s%s%s - %s%s\n", childPrefix, pinnedEmoji, colorizeID(child.id), child.title, tagInfo)
	}

	// Show hidden count if any
	if hiddenCount > 0 {
		fmt.Printf("%s└── (%d other items)\n", prefix, hiddenCount)
	}

	return len(visible)
}

// containerInfo holds container display information
type containerInfo struct {
	id            string
	title         string
	status        string
	pinned        bool
	workbenchID   string
	containerType string // "shipment", "conclave", "investigation", "tome"
}

// SummaryCmd returns the summary command
func SummaryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Show summary of all open commissions and containers",
		Long: `Show a hierarchical summary of commissions with all container types.

Display modes:
  Default: Show only focused container's commission (if focus is set)
  --all: Show all commissions and containers
  --commission [id]: Show specific commission (or 'current' for focus/context)

Containers shown:
  - Shipments (SHIP-*) with Tasks
  - Conclaves (CON-*) with Tasks/Questions/Plans
  - Investigations (INV-*) with Questions
  - Tomes (TOME-*) with Notes

Filtering:
  --commission [id]           Show specific commission (or 'current')
  --filter-statuses paused    Hide items with these statuses
  --filter-containers SHIP    Show only these container types
  --tags research             Show only leaves with these tags
  --not-tags blocked          Hide leaves with these tags

Examples:
  orc summary                          # focused container's commission only
  orc summary --all                    # all commissions
  orc summary --commission current     # explicit current commission
  orc summary --filter-containers SHIP,CON --all
  orc summary --tags research --all
  orc summary --filter-statuses paused,blocked`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get current working directory for config
			cwd, err := os.Getwd()
			if err != nil {
				cwd = ""
			}

			// Get flags
			commissionFilter, _ := cmd.Flags().GetString("commission")
			expandAll, _ := cmd.Flags().GetBool("all")
			filterStatuses, _ := cmd.Flags().GetStringSlice("filter-statuses")
			filterContainers, _ := cmd.Flags().GetStringSlice("filter-containers")
			filterLeaves, _ := cmd.Flags().GetStringSlice("filter-leaves")
			includeTags, _ := cmd.Flags().GetStringSlice("tags")
			excludeTags, _ := cmd.Flags().GetStringSlice("not-tags")

			// Build filter config
			filters := &filterConfig{
				statusMap:      make(map[string]bool),
				containerTypes: make(map[string]bool),
				includeTags:    make(map[string]bool),
				excludeTags:    make(map[string]bool),
				leafTypes:      make(map[string]bool),
			}
			for _, s := range filterStatuses {
				filters.statusMap[s] = true
			}
			for _, c := range filterContainers {
				filters.containerTypes[strings.ToUpper(c)] = true
			}
			for _, lt := range filterLeaves {
				filters.leafTypes[strings.ToLower(strings.TrimSpace(lt))] = true
			}
			for _, t := range includeTags {
				filters.includeTags[t] = true
			}
			for _, t := range excludeTags {
				filters.excludeTags[t] = true
			}

			// Get current focus from cwd config only (no home fallback)
			cfg, _ := config.LoadConfig(cwd)
			focusID := GetCurrentFocus(cfg)

			// Determine commission filter
			var filterCommissionID string
			if commissionFilter == "current" {
				// First try config in cwd
				commissionID := ctx.GetContextCommissionID()
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

			// Build header with filter info
			headerParts := []string{"📊 ORC Summary"}
			if filterCommissionID != "" {
				headerParts = append(headerParts, filterCommissionID)
			}
			if len(includeTags) > 0 {
				headerParts = append(headerParts, fmt.Sprintf("tags=%s", strings.Join(includeTags, ",")))
			}
			fmt.Println(strings.Join(headerParts, " - "))
			fmt.Println()

			// Get all non-complete commissions via service
			commissions, err := wire.CommissionService().ListCommissions(context.Background(), primary.CommissionFilters{})
			if err != nil {
				return fmt.Errorf("failed to list commissions: %w", err)
			}

			// Filter to open commissions (not complete or archived)
			var openCommissions []*primary.Commission
			for _, m := range commissions {
				if m.Status == "complete" || m.Status == "archived" {
					continue
				}
				if filters.statusMap[m.Status] {
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

			// Display each commission
			for i, commission := range openCommissions {
				// Collect all active containers for this commission
				var allContainers []containerInfo
				var focusedContainer *containerInfo

				// Collect shipments
				if len(filters.containerTypes) == 0 || filters.containerTypes["SHIP"] {
					shipments, _ := wire.ShipmentService().ListShipments(context.Background(), primary.ShipmentFilters{CommissionID: commission.ID})
					for _, s := range shipments {
						if s.Status == "complete" || filters.statusMap[s.Status] {
							continue
						}
						c := containerInfo{
							id: s.ID, title: s.Title, status: s.Status,
							pinned: s.Pinned, workbenchID: s.AssignedWorkbenchID, containerType: "shipment",
						}
						if s.ID == focusID {
							focusedContainer = &c
						}
						allContainers = append(allContainers, c)
					}
				}

				// Collect conclaves
				if len(filters.containerTypes) == 0 || filters.containerTypes["CON"] {
					conclaves, _ := wire.ConclaveService().ListConclaves(context.Background(), primary.ConclaveFilters{CommissionID: commission.ID})
					for _, c := range conclaves {
						if c.Status == "closed" || filters.statusMap[c.Status] {
							continue
						}
						cont := containerInfo{
							id: c.ID, title: c.Title, status: c.Status,
							pinned: c.Pinned, workbenchID: "", containerType: "conclave",
						}
						if c.ID == focusID {
							focusedContainer = &cont
						}
						allContainers = append(allContainers, cont)
					}
				}

				// Collect investigations
				if len(filters.containerTypes) == 0 || filters.containerTypes["INV"] {
					investigations, _ := wire.InvestigationService().ListInvestigations(context.Background(), primary.InvestigationFilters{CommissionID: commission.ID})
					for _, inv := range investigations {
						if inv.Status == "complete" || filters.statusMap[inv.Status] {
							continue
						}
						c := containerInfo{
							id: inv.ID, title: inv.Title, status: inv.Status,
							pinned: inv.Pinned, workbenchID: "", containerType: "investigation",
						}
						if inv.ID == focusID {
							focusedContainer = &c
						}
						allContainers = append(allContainers, c)
					}
				}

				// Collect tomes (only those without a conclave_id - conclaved tomes appear nested under conclaves)
				if len(filters.containerTypes) == 0 || filters.containerTypes["TOME"] {
					tomes, _ := wire.TomeService().ListTomes(context.Background(), primary.TomeFilters{CommissionID: commission.ID})
					for _, t := range tomes {
						if t.Status == "complete" || filters.statusMap[t.Status] {
							continue
						}
						// Skip tomes that have a conclave_id - they appear nested under their conclave
						if t.ConclaveID != "" {
							continue
						}
						c := containerInfo{
							id: t.ID, title: t.Title, status: t.Status,
							pinned: t.Pinned, containerType: "tome",
						}
						if t.ID == focusID {
							focusedContainer = &c
						}
						allContainers = append(allContainers, c)
					}
				}

				// Decide what to show
				var containersToShow []containerInfo
				otherContainerCount := 0

				if focusedContainer != nil && !expandAll {
					// Show only focused container
					containersToShow = []containerInfo{*focusedContainer}
					otherContainerCount = len(allContainers) - 1
				} else {
					// Show all containers (if tag filtering, only show containers with matching leaves)
					if len(filters.includeTags) > 0 {
						for _, c := range allContainers {
							hasMatchingLeaves := containerHasMatchingLeaves(c, filters)
							if hasMatchingLeaves {
								containersToShow = append(containersToShow, c)
							} else {
								otherContainerCount++
							}
						}
					} else {
						containersToShow = allContainers
					}
				}

				if len(containersToShow) == 0 && otherContainerCount == 0 {
					continue // Skip this commission entirely
				}

				// Display commission header
				fmt.Printf("%s - %s\n", colorizeID(commission.ID), commission.Title)
				fmt.Println("│")

				if len(containersToShow) == 0 {
					fmt.Println("└── (No containers with matching items)")
				} else {
					for j, container := range containersToShow {
						isFocused := container.id == focusID
						pinnedEmoji := ""
						if container.pinned {
							pinnedEmoji = "📌 "
						}
						workbenchInfo := ""
						if container.workbenchID != "" {
							workbench, err := wire.WorkbenchService().GetWorkbench(context.Background(), container.workbenchID)
							if err == nil {
								workbenchInfo = " " + colorizeWorkbench(workbench.Name, workbench.ID)
							}
						}

						isLast := j == len(containersToShow)-1 && otherContainerCount == 0
						prefix := "├── "
						if isLast {
							prefix = "└── "
						}

						// Add [FOCUSED] marker
						focusMarker := ""
						if isFocused {
							focusMarker = color.New(color.FgHiMagenta).Sprint("[FOCUSED] ")
						}

						statusInfo := colorizeStatus(container.status)
						if statusInfo != "" {
							fmt.Printf("%s%s%s%s - %s - %s%s\n", prefix, focusMarker, pinnedEmoji, colorizeID(container.id), statusInfo, container.title, workbenchInfo)
						} else {
							fmt.Printf("%s%s%s%s - %s%s\n", prefix, focusMarker, pinnedEmoji, colorizeID(container.id), container.title, workbenchInfo)
						}

						// Display children based on container type
						childPrefix := "│   "
						if isLast {
							childPrefix = "    "
						}

						switch container.containerType {
						case "shipment":
							displayShipmentChildren(container.id, childPrefix, filters)
						case "conclave":
							displayConclaveChildren(container.id, childPrefix, filters)
						case "investigation":
							displayInvestigationChildren(container.id, childPrefix, filters)
						case "tome":
							displayTomeChildren(container.id, childPrefix, filters)
						}

						if !isLast || otherContainerCount > 0 {
							fmt.Println("│")
						}
					}
				}

				// Show other containers count
				if otherContainerCount > 0 {
					msg := fmt.Sprintf("(%d other containers", otherContainerCount)
					if focusedContainer != nil && !expandAll {
						msg += " - use --all to show"
					}
					msg += ")"
					fmt.Printf("└── %s\n", msg)
				}

				if i < len(openCommissions)-1 {
					fmt.Println()
				}
			}

			fmt.Println()

			return nil
		},
	}

	cmd.Flags().StringP("commission", "c", "", "Commission filter: commission ID or 'current' for context commission")
	cmd.Flags().Bool("all", false, "Show all containers (default: only show focused container if set)")
	cmd.Flags().StringSlice("filter-statuses", []string{}, "Hide items with these statuses (comma-separated: paused,blocked)")
	cmd.Flags().StringSlice("filter-containers", []string{}, "Show only these container types (comma-separated: SHIP,CON,INV,TOME)")
	cmd.Flags().StringSlice("filter-leaves", []string{}, "Hide these leaf types (comma-separated: tasks,notes,questions,plans)")
	cmd.Flags().StringSlice("tags", []string{}, "Show only leaves with these tags")
	cmd.Flags().StringSlice("not-tags", []string{}, "Hide leaves with these tags")

	return cmd
}

// containerHasMatchingLeaves checks if a container has any leaves matching the tag filter
func containerHasMatchingLeaves(c containerInfo, filters *filterConfig) bool {
	switch c.containerType {
	case "shipment":
		shipmentTasks, _ := wire.ShipmentService().GetShipmentTasks(context.Background(), c.id)
		// Convert to models.Task for tag checking
		var tasks []*models.Task
		for _, t := range shipmentTasks {
			tasks = append(tasks, &models.Task{
				ID:     t.ID,
				Title:  t.Title,
				Status: t.Status,
				Pinned: t.Pinned,
			})
		}
		serviceNotes, _ := wire.NoteService().GetNotesByContainer(context.Background(), "shipment", c.id)
		var notes []*models.Note
		for _, n := range serviceNotes {
			if n.Status == "closed" {
				continue
			}
			notes = append(notes, &models.Note{ID: n.ID, Title: n.Title, Pinned: n.Pinned})
		}
		for _, t := range tasks {
			if t.Status == "complete" || filters.statusMap[t.Status] {
				continue
			}
			show, _ := shouldShowLeaf(t.ID, filters)
			if show {
				return true
			}
		}
		for _, n := range notes {
			show, _ := shouldShowLeaf(n.ID, filters)
			if show {
				return true
			}
		}
	case "conclave":
		tasks, _ := wire.ConclaveService().GetConclaveTasks(context.Background(), c.id)
		plans, _ := wire.ConclaveService().GetConclavePlans(context.Background(), c.id)
		serviceNotes, _ := wire.NoteService().GetNotesByContainer(context.Background(), "conclave", c.id)
		var notes []*models.Note
		for _, n := range serviceNotes {
			if n.Status == "closed" {
				continue
			}
			notes = append(notes, &models.Note{ID: n.ID, Title: n.Title, Pinned: n.Pinned})
		}
		for _, t := range tasks {
			if t.Status == "complete" || filters.statusMap[t.Status] {
				continue
			}
			show, _ := shouldShowLeaf(t.ID, filters)
			if show {
				return true
			}
		}
		for _, p := range plans {
			if p.Status == "approved" || filters.statusMap[p.Status] {
				continue
			}
			show, _ := shouldShowLeaf(p.ID, filters)
			if show {
				return true
			}
		}
		for _, n := range notes {
			show, _ := shouldShowLeaf(n.ID, filters)
			if show {
				return true
			}
		}
	case "investigation":
		serviceNotes, _ := wire.NoteService().GetNotesByContainer(context.Background(), "investigation", c.id)
		var notes []*models.Note
		for _, n := range serviceNotes {
			if n.Status == "closed" {
				continue
			}
			notes = append(notes, &models.Note{ID: n.ID, Title: n.Title, Pinned: n.Pinned})
		}
		for _, n := range notes {
			show, _ := shouldShowLeaf(n.ID, filters)
			if show {
				return true
			}
		}
	case "tome":
		serviceNotes, _ := wire.TomeService().GetTomeNotes(context.Background(), c.id)
		for _, n := range serviceNotes {
			if n.Status == "closed" {
				continue
			}
			show, _ := shouldShowLeaf(n.ID, filters)
			if show {
				return true
			}
		}
	}
	return false
}
