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

// colorizeGrove wraps grove info with color (cyan for visibility)
func colorizeGrove(groveName, groveID string) string {
	c := color.New(color.FgCyan)
	return c.Sprintf("[Grove: %s (%s)]", groveName, groveID)
}

// getIDColor returns a deterministic color for an ID type (TASK, SHIP, MISSION)
// Uses FNV-1a hash on the ID prefix to ensure all IDs of same type have same color
func getIDColor(idType string) *color.Color {
	// Hash the ID type (e.g., "TASK", "SHIP", "MISSION")
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

	prefix := parts[0] // "TASK", "SHIP", "MISSION"
	c := getIDColor(prefix)
	return c.Sprint(id)
}

// getStatusColor returns a fixed semantic color for a status
// Uses meaningful colors: ready=green, paused=yellow, blocked=red, etc.
func getStatusColor(status string) *color.Color {
	switch status {
	case "ready":
		return color.New(color.FgGreen)
	case "needs_design":
		return color.New(color.FgHiYellow) // Bright yellow for needs design
	case "ready_to_implement":
		return color.New(color.FgHiCyan) // Bright cyan for ready to implement
	case "paused":
		return color.New(color.FgYellow)
	case "blocked":
		return color.New(color.FgRed)
	case "design":
		return color.New(color.FgCyan)
	case "awaiting_approval":
		return color.New(color.FgMagenta) // Magenta for awaiting approval
	case "complete":
		return color.New(color.FgHiGreen) // Bright green
	default:
		return color.New(color.FgWhite)
	}
}

// colorizeStatus formats status as uppercase with semantic color
// Returns empty string for "ready" status (default, not shown)
func colorizeStatus(status string) string {
	if status == "" || status == "ready" || status == "active" {
		return "" // Don't show empty or default statuses
	}
	c := getStatusColor(status)
	return c.Sprintf("%s", strings.ToUpper(status))
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
	case "Q":
		return "question"
	case "PLAN":
		return "plan"
	case "NOTE":
		return "note"
	default:
		return ""
	}
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

// displayShipmentChildren shows tasks and notes under a shipment with tag filtering
// Returns count of visible items
func displayShipmentChildren(shipmentID, prefix string, filters *filterConfig) int {
	shipmentTasks, _ := wire.ShipmentService().GetShipmentTasks(context.Background(), shipmentID)
	// Convert to models.Task for the rest of the function
	var tasks []*models.Task
	for _, t := range shipmentTasks {
		tasks = append(tasks, &models.Task{
			ID:     t.ID,
			Title:  t.Title,
			Status: t.Status,
			Pinned: t.Pinned,
		})
	}
	serviceNotes, _ := wire.NoteService().GetNotesByContainer(context.Background(), "shipment", shipmentID)
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

	for _, t := range tasks {
		if t.Status == "complete" || filters.statusMap[t.Status] {
			continue
		}
		show, tagName := shouldShowLeaf(t.ID, filters)
		if show {
			visible = append(visible, childItem{t.ID, t.Title, t.Status, t.Pinned, tagName})
		} else {
			hiddenCount++
		}
	}
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

// displayConclaveChildren shows tasks/questions/plans/notes under a conclave with tag filtering
// Returns count of visible items
func displayConclaveChildren(conclaveID, prefix string, filters *filterConfig) int {
	ctx := context.Background()
	// Get tasks via ConclaveService
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
	// Get questions via ConclaveService
	serviceQuestions, _ := wire.ConclaveService().GetConclaveQuestions(ctx, conclaveID)
	var questions []*models.Question
	for _, q := range serviceQuestions {
		questions = append(questions, &models.Question{
			ID:     q.ID,
			Title:  q.Title,
			Status: q.Status,
			Pinned: q.Pinned,
		})
	}
	// Get plans via ConclaveService
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
	// Get notes (filter out closed notes)
	serviceNotes, _ := wire.NoteService().GetNotesByContainer(ctx, "conclave", conclaveID)
	// Convert to models.Note for the rest of the function
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

	for _, t := range tasks {
		if t.Status == "complete" || filters.statusMap[t.Status] {
			continue
		}
		show, tagName := shouldShowLeaf(t.ID, filters)
		if show {
			visible = append(visible, childItem{t.ID, t.Title, t.Status, t.Pinned, tagName})
		} else {
			hiddenCount++
		}
	}
	for _, q := range questions {
		if q.Status == "answered" || filters.statusMap[q.Status] {
			continue
		}
		show, tagName := shouldShowLeaf(q.ID, filters)
		if show {
			visible = append(visible, childItem{q.ID, q.Title, q.Status, q.Pinned, tagName})
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
			visible = append(visible, childItem{p.ID, p.Title, p.Status, p.Pinned, tagName})
		} else {
			hiddenCount++
		}
	}
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

// displayInvestigationChildren shows questions and notes under an investigation with tag filtering
// Returns count of visible items
func displayInvestigationChildren(investigationID, prefix string, filters *filterConfig) int {
	invQuestions, _ := wire.InvestigationService().GetInvestigationQuestions(context.Background(), investigationID)
	// Convert to models.Question for the rest of the function
	var questions []*models.Question
	for _, q := range invQuestions {
		questions = append(questions, &models.Question{
			ID:     q.ID,
			Title:  q.Title,
			Status: q.Status,
			Pinned: q.Pinned,
		})
	}
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

	for _, q := range questions {
		if q.Status == "answered" || filters.statusMap[q.Status] {
			continue
		}
		show, tagName := shouldShowLeaf(q.ID, filters)
		if show {
			visible = append(visible, childItem{q.ID, q.Title, q.Status, q.Pinned, tagName})
		} else {
			hiddenCount++
		}
	}
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
	groveID       string
	containerType string // "shipment", "conclave", "investigation", "tome"
}

// SummaryCmd returns the summary command
func SummaryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Show summary of all open missions and containers",
		Long: `Show a hierarchical summary of missions with all container types.

Display modes:
  Default: Show only focused container (if focus is set)
  --all: Show all containers

Containers shown:
  - Shipments (SHIP-*) with Tasks
  - Conclaves (CON-*) with Tasks/Questions/Plans
  - Investigations (INV-*) with Questions
  - Tomes (TOME-*) with Notes

Filtering:
  --mission [id]              Show specific mission (or 'current')
  --filter-statuses paused    Hide items with these statuses
  --filter-containers SHIP    Show only these container types
  --tags research             Show only leaves with these tags
  --not-tags blocked          Hide leaves with these tags

Examples:
  orc summary                          # focused container only (if set)
  orc summary --all                    # all containers
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
			missionFilter, _ := cmd.Flags().GetString("mission")
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

			// Get current focus
			cfg, _ := config.LoadConfig(cwd)
			focusID := GetCurrentFocus(cfg)

			// Determine mission filter
			var filterMissionID string
			if missionFilter == "current" {
				// Get current mission from context
				missionCtx, _ := ctx.DetectMissionContext()
				if missionCtx == nil || missionCtx.MissionID == "" {
					return fmt.Errorf("--mission current requires being in a mission context (no .orc/config.json found)")
				}
				filterMissionID = missionCtx.MissionID
			} else if missionFilter != "" {
				filterMissionID = missionFilter
			}

			// Build header with filter info
			headerParts := []string{"📊 ORC Summary"}
			if filterMissionID != "" {
				headerParts = append(headerParts, filterMissionID)
			}
			if len(includeTags) > 0 {
				headerParts = append(headerParts, fmt.Sprintf("tags=%s", strings.Join(includeTags, ",")))
			}
			fmt.Println(strings.Join(headerParts, " - "))
			fmt.Println()

			// Get all non-complete missions via service
			missions, err := wire.MissionService().ListMissions(context.Background(), primary.MissionFilters{})
			if err != nil {
				return fmt.Errorf("failed to list missions: %w", err)
			}

			// Filter to open missions (not complete or archived)
			var openMissions []*primary.Mission
			for _, m := range missions {
				if m.Status == "complete" || m.Status == "archived" {
					continue
				}
				if filters.statusMap[m.Status] {
					continue
				}
				if filterMissionID != "" && m.ID != filterMissionID {
					continue
				}
				openMissions = append(openMissions, m)
			}

			if len(openMissions) == 0 {
				if filterMissionID != "" {
					fmt.Printf("No open containers for %s\n", filterMissionID)
				} else {
					fmt.Println("No open missions")
				}
				return nil
			}

			// Display each mission
			for i, mission := range openMissions {
				// Collect all active containers for this mission
				var allContainers []containerInfo
				var focusedContainer *containerInfo

				// Collect shipments
				if len(filters.containerTypes) == 0 || filters.containerTypes["SHIP"] {
					shipments, _ := wire.ShipmentService().ListShipments(context.Background(), primary.ShipmentFilters{MissionID: mission.ID})
					for _, s := range shipments {
						if s.Status == "complete" || filters.statusMap[s.Status] {
							continue
						}
						c := containerInfo{
							id: s.ID, title: s.Title, status: s.Status,
							pinned: s.Pinned, groveID: s.AssignedGroveID, containerType: "shipment",
						}
						if s.ID == focusID {
							focusedContainer = &c
						}
						allContainers = append(allContainers, c)
					}
				}

				// Collect conclaves
				if len(filters.containerTypes) == 0 || filters.containerTypes["CON"] {
					conclaves, _ := wire.ConclaveService().ListConclaves(context.Background(), primary.ConclaveFilters{MissionID: mission.ID})
					for _, c := range conclaves {
						if c.Status == "complete" || filters.statusMap[c.Status] {
							continue
						}
						cont := containerInfo{
							id: c.ID, title: c.Title, status: c.Status,
							pinned: c.Pinned, groveID: c.AssignedGroveID, containerType: "conclave",
						}
						if c.ID == focusID {
							focusedContainer = &cont
						}
						allContainers = append(allContainers, cont)
					}
				}

				// Collect investigations
				if len(filters.containerTypes) == 0 || filters.containerTypes["INV"] {
					investigations, _ := wire.InvestigationService().ListInvestigations(context.Background(), primary.InvestigationFilters{MissionID: mission.ID})
					for _, inv := range investigations {
						if inv.Status == "complete" || filters.statusMap[inv.Status] {
							continue
						}
						c := containerInfo{
							id: inv.ID, title: inv.Title, status: inv.Status,
							pinned: inv.Pinned, groveID: inv.AssignedGroveID, containerType: "investigation",
						}
						if inv.ID == focusID {
							focusedContainer = &c
						}
						allContainers = append(allContainers, c)
					}
				}

				// Collect tomes
				if len(filters.containerTypes) == 0 || filters.containerTypes["TOME"] {
					tomes, _ := wire.TomeService().ListTomes(context.Background(), primary.TomeFilters{MissionID: mission.ID})
					for _, t := range tomes {
						if t.Status == "complete" || filters.statusMap[t.Status] {
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
					continue // Skip this mission entirely
				}

				// Display mission header
				fmt.Printf("%s - %s\n", colorizeID(mission.ID), mission.Title)
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
						groveInfo := ""
						if container.groveID != "" {
							grove, err := wire.GroveService().GetGrove(context.Background(), container.groveID)
							if err == nil {
								groveInfo = " " + colorizeGrove(grove.Name, grove.ID)
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
							fmt.Printf("%s%s%s%s - %s - %s%s\n", prefix, focusMarker, pinnedEmoji, colorizeID(container.id), statusInfo, container.title, groveInfo)
						} else {
							fmt.Printf("%s%s%s%s - %s%s\n", prefix, focusMarker, pinnedEmoji, colorizeID(container.id), container.title, groveInfo)
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

				if i < len(openMissions)-1 {
					fmt.Println()
				}
			}

			fmt.Println()

			return nil
		},
	}

	cmd.Flags().StringP("mission", "m", "", "Mission filter: mission ID or 'current' for context mission")
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
		questions, _ := wire.ConclaveService().GetConclaveQuestions(context.Background(), c.id)
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
		for _, q := range questions {
			if q.Status == "answered" || filters.statusMap[q.Status] {
				continue
			}
			show, _ := shouldShowLeaf(q.ID, filters)
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
		invQuestions, _ := wire.InvestigationService().GetInvestigationQuestions(context.Background(), c.id)
		var questions []*models.Question
		for _, q := range invQuestions {
			questions = append(questions, &models.Question{ID: q.ID, Title: q.Title, Status: q.Status, Pinned: q.Pinned})
		}
		serviceNotes, _ := wire.NoteService().GetNotesByContainer(context.Background(), "investigation", c.id)
		var notes []*models.Note
		for _, n := range serviceNotes {
			if n.Status == "closed" {
				continue
			}
			notes = append(notes, &models.Note{ID: n.ID, Title: n.Title, Pinned: n.Pinned})
		}
		for _, q := range questions {
			if q.Status == "answered" || filters.statusMap[q.Status] {
				continue
			}
			show, _ := shouldShowLeaf(q.ID, filters)
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
