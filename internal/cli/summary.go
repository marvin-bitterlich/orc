package cli

import (
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/example/orc/internal/context"
	"github.com/example/orc/internal/models"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

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

// getIDColor returns a deterministic color for an ID type (TASK, EPIC, RH, MISSION)
// Uses FNV-1a hash on the ID prefix to ensure all IDs of same type have same color
func getIDColor(idType string) *color.Color {
	// Hash the ID type (e.g., "TASK", "EPIC", "RH")
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

	prefix := parts[0] // "TASK", "EPIC", "RH", "MISSION"
	c := getIDColor(prefix)
	return c.Sprint(id)
}

// getStatusColor returns a fixed semantic color for a status
// Uses meaningful colors: ready=green, paused=yellow, blocked=red, etc.
func getStatusColor(status string) *color.Color {
	switch status {
	case "ready":
		return color.New(color.FgGreen)
	case "paused":
		return color.New(color.FgYellow)
	case "blocked":
		return color.New(color.FgRed)
	case "implement":
		return color.New(color.FgBlue)
	case "design":
		return color.New(color.FgCyan)
	case "complete":
		return color.New(color.FgHiGreen) // Bright green
	default:
		return color.New(color.FgWhite)
	}
}

// colorizeStatus wraps a status in {status:value} format with semantic color
// Returns empty string for "ready" status (default, not shown)
func colorizeStatus(status string) string {
	if status == "ready" {
		return "" // Don't show default status
	}
	c := getStatusColor(status)
	return c.Sprintf("{status:%s}", status)
}

// SummaryCmd returns the summary command
func SummaryCmd() *cobra.Command {
	var showAll bool

	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Show summary of all open missions, epics, and tasks",
		Long: `Show a hierarchical summary of missions with their epics, rabbit holes, and tasks.

Filtering:
  --mission [id]       Show specific mission (e.g., MISSION-001)
  --mission current    Show current mission (requires mission context)
  --hide paused        Hide paused items
  --hide blocked       Hide blocked items
  --hide paused,blocked  Hide multiple statuses

Examples:
  orc summary                       # all missions, all statuses
  orc summary --mission current     # current mission only
  orc summary --mission MISSION-001 # specific mission
  orc summary --hide paused         # hide paused work
  orc summary --mission current --hide paused,blocked  # focused view`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get mission from flag (default: empty = all missions)
			missionFilter, _ := cmd.Flags().GetString("mission")

			// Determine mission filter
			var filterMissionID string
			if missionFilter == "current" {
				// Get current mission from context
				missionCtx, _ := context.DetectMissionContext()
				if missionCtx == nil || missionCtx.MissionID == "" {
					return fmt.Errorf("--mission current requires being in a mission context (no .orc/config.json found)")
				}
				filterMissionID = missionCtx.MissionID
				fmt.Printf("📊 ORC Summary - %s (Current Mission)\n", filterMissionID)
			} else if missionFilter != "" {
				// Specific mission ID provided
				filterMissionID = missionFilter
				fmt.Printf("📊 ORC Summary - %s\n", filterMissionID)
			} else {
				// No filter - show all missions
				fmt.Println("📊 ORC Summary - Open Work")
			}
			fmt.Println()

			// Get all non-complete missions
			missions, err := models.ListMissions("")
			if err != nil {
				return fmt.Errorf("failed to list missions: %w", err)
			}

			// Get hide statuses from flag
			hideStatuses, _ := cmd.Flags().GetStringSlice("hide")
			hideMap := make(map[string]bool)
			for _, status := range hideStatuses {
				hideMap[status] = true
			}

			// Filter to open missions (not complete or archived)
			var openMissions []*models.Mission
			for _, m := range missions {
				// Always hide complete and archived
				if m.Status == "complete" || m.Status == "archived" {
					continue
				}
				// Hide if in hide list
				if hideMap[m.Status] {
					continue
				}
				// If in deputy context and not showing all, filter to this mission
				if filterMissionID != "" && m.ID != filterMissionID {
					continue
				}
				openMissions = append(openMissions, m)
			}

			if len(openMissions) == 0 {
				if filterMissionID != "" {
					fmt.Printf("No open epics for %s\n", filterMissionID)
				} else {
					fmt.Println("No open missions")
				}
				return nil
			}

			// Display each mission with its epics in tree format
			for i, mission := range openMissions {
				// Display mission
				fmt.Printf("%s - %s\n", colorizeID(mission.ID), mission.Title)
				fmt.Println("│") // Empty line with vertical continuation after mission header

				// Get epics for this mission
				epics, err := models.ListEpics(mission.ID, "")
				if err != nil {
					return fmt.Errorf("failed to list epics for %s: %w", mission.ID, err)
				}

				// Filter to non-complete epics
				var activeEpics []*models.Epic
				for _, epic := range epics {
					// Always hide complete
					if epic.Status == "complete" {
						continue
					}
					// Hide if in hide list
					if hideMap[epic.Status] {
						continue
					}
					activeEpics = append(activeEpics, epic)
				}

				if len(activeEpics) > 0 {
					// Display epics with their children
					for j, epic := range activeEpics {
						pinnedEmoji := ""
						// Add pin emoji if pinned
						if epic.Pinned {
							pinnedEmoji = "📌 "
						}
						groveInfo := ""
						if epic.AssignedGroveID.Valid {
							groveInfo = fmt.Sprintf(" [Grove: %s]", epic.AssignedGroveID.String)
						}

						// Use └── for last epic, ├── for others
						var prefix string
						if j < len(activeEpics)-1 {
							prefix = "├── "
						} else {
							prefix = "└── "
						}
						statusInfo := " " + colorizeStatus(epic.Status)
						fmt.Printf("%s%s%s - %s%s%s\n", prefix, pinnedEmoji, colorizeID(epic.ID), epic.Title, statusInfo, groveInfo)

						// Check if epic has rabbit holes or direct tasks
						hasRH, _ := models.HasRabbitHoles(epic.ID)
						isLastEpic := j == len(activeEpics)-1

						if hasRH {
							// Epic has rabbit holes
							rabbitHoles, err := models.ListRabbitHoles(epic.ID, "")
							if err == nil {
								var activeRHs []*models.RabbitHole
								for _, rh := range rabbitHoles {
									if rh.Status != "complete" && !hideMap[rh.Status] {
										activeRHs = append(activeRHs, rh)
									}
								}

								for k, rh := range activeRHs {
									pinnedEmoji := ""
									if rh.Pinned {
										pinnedEmoji = "📌 "
									}

									var rhPrefix string
									isLastRH := k == len(activeRHs)-1
									if isLastEpic {
										if isLastRH {
											rhPrefix = "    └── "
										} else {
											rhPrefix = "    ├── "
										}
									} else {
										if isLastRH {
											rhPrefix = "│   └── "
										} else {
											rhPrefix = "│   ├── "
										}
									}

									statusInfo := " " + colorizeStatus(rh.Status)
					fmt.Printf("%s%s%s - %s%s\n", rhPrefix, pinnedEmoji, colorizeID(rh.ID), rh.Title, statusInfo)

									// Get expand flag
									expand, _ := cmd.Flags().GetBool("expand")

									if expand {
										// Display all tasks under rabbit hole
										tasks, err := models.GetRabbitHoleTasks(rh.ID)
										if err == nil {
											var activeTasks []*models.Task
											for _, task := range tasks {
												if task.Status != "complete" && !hideMap[task.Status] {
													activeTasks = append(activeTasks, task)
												}
											}

											for t, task := range activeTasks {
												pinnedEmoji := ""
												if task.Pinned {
													pinnedEmoji = "📌 "
												}

												var taskPrefix string
												isLastTask := t == len(activeTasks)-1

												if isLastEpic {
													if isLastRH {
														if isLastTask {
															taskPrefix = "        └── "
														} else {
															taskPrefix = "        ├── "
														}
													} else {
														if isLastTask {
															taskPrefix = "    │   └── "
														} else {
															taskPrefix = "    │   ├── "
														}
													}
												} else {
													if isLastRH {
														if isLastTask {
															taskPrefix = "│       └── "
														} else {
															taskPrefix = "│       ├── "
														}
													} else {
														if isLastTask {
															taskPrefix = "│   │   └── "
														} else {
															taskPrefix = "│   │   ├── "
														}
													}
												}

												tagInfo := ""
											tag, _ := models.GetTaskTag(task.ID)
											if tag != nil {
												tagInfo = " " + colorizeTag(tag.Name)
											}

											statusInfo := " " + colorizeStatus(task.Status)
						fmt.Printf("%s%s%s - %s%s%s\n", taskPrefix, pinnedEmoji, colorizeID(task.ID), task.Title, statusInfo, tagInfo)
											}
										}
									} else {
										// Display summary only
										summary := summarizeRabbitHoleTasks(rh.ID, hideMap)

										// Use indented prefix for summary line
										var summaryPrefix string
										if isLastEpic {
											summaryPrefix = "    "
										} else {
											summaryPrefix = "│   "
										}

										fmt.Printf("%s    [%s]\n", summaryPrefix, summary)
									}
								}
							}
						} else {
							// Epic has direct tasks
							tasks, err := models.GetDirectTasks(epic.ID)
							if err == nil {
								var activeTasks []*models.Task
								for _, task := range tasks {
									if task.Status != "complete" && !hideMap[task.Status] {
										activeTasks = append(activeTasks, task)
									}
								}

								for k, task := range activeTasks {
									pinnedEmoji := ""
									if task.Pinned {
										pinnedEmoji = "📌 "
									}

									var taskPrefix string
									isLastTask := k == len(activeTasks)-1
									if isLastEpic {
										if isLastTask {
											taskPrefix = "    └── "
										} else {
											taskPrefix = "    ├── "
										}
									} else {
										if isLastTask {
											taskPrefix = "│   └── "
										} else {
											taskPrefix = "│   ├── "
										}
									}

									tagInfo := ""
									tag, _ := models.GetTaskTag(task.ID)
									if tag != nil {
										tagInfo = " " + colorizeTag(tag.Name)
									}

									statusInfo := " " + colorizeStatus(task.Status)
						fmt.Printf("%s%s%s - %s%s%s\n", taskPrefix, pinnedEmoji, colorizeID(task.ID), task.Title, statusInfo, tagInfo)
								}
							}
						}

						// Add vertical continuation line between epics (not after last)
						if j < len(activeEpics)-1 {
							fmt.Println("│")
						}
					}
				} else {
					fmt.Println("└── (No active epics)")
				}

				// Add spacing between missions
				if i < len(openMissions)-1 {
					fmt.Println()
				}
			}

			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all missions (override deputy scoping)")
	cmd.Flags().StringP("mission", "m", "", "Mission filter: mission ID or 'current' for context mission")
	cmd.Flags().StringSlice("hide", []string{}, "Hide items with these statuses (comma-separated: paused,blocked)")
	cmd.Flags().Bool("expand", false, "Expand rabbit holes to show all tasks")

	return cmd
}

func summarizeRabbitHoleTasks(rhID string, hideMap map[string]bool) string {
	tasks, err := models.GetRabbitHoleTasks(rhID)
	if err != nil || len(tasks) == 0 {
		return "no tasks"
	}

	// Count tasks by status and collect tags
	statusCounts := make(map[string]int)
	tagCounts := make(map[string]int)
	total := 0
	for _, task := range tasks {
		if task.Status != "complete" && !hideMap[task.Status] {
			statusCounts[task.Status]++
			total++

			// Get tag for this task
			tag, _ := models.GetTaskTag(task.ID)
			if tag != nil {
				tagCounts[tag.Name]++
			}
		}
	}

	if total == 0 {
		return "no active tasks"
	}

	// Build summary string
	parts := []string{}
	if count := statusCounts["ready"]; count > 0 {
		coloredStatus := getStatusColor("ready").Sprint("ready")
		parts = append(parts, fmt.Sprintf("%d %s", count, coloredStatus))
	}
	if count := statusCounts["design"]; count > 0 {
		coloredStatus := getStatusColor("design").Sprint("design")
		parts = append(parts, fmt.Sprintf("%d %s", count, coloredStatus))
	}
	if count := statusCounts["implement"]; count > 0 {
		coloredStatus := getStatusColor("implement").Sprint("implement")
		parts = append(parts, fmt.Sprintf("%d %s", count, coloredStatus))
	}
	if count := statusCounts["blocked"]; count > 0 {
		coloredStatus := getStatusColor("blocked").Sprint("blocked")
		parts = append(parts, fmt.Sprintf("%d %s", count, coloredStatus))
	}
	if count := statusCounts["paused"]; count > 0 {
		coloredStatus := getStatusColor("paused").Sprint("paused")
		parts = append(parts, fmt.Sprintf("%d %s", count, coloredStatus))
	}

	summary := fmt.Sprintf("%d tasks: %s", total, strings.Join(parts, ", "))

	// Add tag counts if any tags present
	if len(tagCounts) > 0 {
		tagParts := []string{}
		for tagName, count := range tagCounts {
			coloredTag := colorizeTag(tagName)
			tagParts = append(tagParts, fmt.Sprintf("%s×%d", coloredTag, count))
		}
		summary += fmt.Sprintf("; tags: %s", strings.Join(tagParts, ", "))
	}

	return summary
}
