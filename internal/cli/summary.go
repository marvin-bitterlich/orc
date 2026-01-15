package cli

import (
	"fmt"

	"github.com/example/orc/internal/context"
	"github.com/example/orc/internal/models"
	"github.com/spf13/cobra"
)

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
				statusEmoji := getStatusEmoji(mission.Status)
				fmt.Printf("%s %s - %s [%s]\n", statusEmoji, mission.ID, mission.Title, mission.Status)
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
						epicEmoji := getStatusEmoji(epic.Status)
						// Add pin emoji if pinned
						if epic.Pinned {
							epicEmoji = "📌" + epicEmoji
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
						fmt.Printf("%s%s %s - %s [%s]%s\n", prefix, epicEmoji, epic.ID, epic.Title, epic.Status, groveInfo)

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
									rhEmoji := getStatusEmoji(rh.Status)
									if rh.Pinned {
										rhEmoji = "📌" + rhEmoji
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

									fmt.Printf("%s%s %s - %s [%s]\n", rhPrefix, rhEmoji, rh.ID, rh.Title, rh.Status)

									// Display tasks under rabbit hole
									tasks, err := models.GetRabbitHoleTasks(rh.ID)
									if err == nil {
										var activeTasks []*models.Task
										for _, task := range tasks {
											if task.Status != "complete" && !hideMap[task.Status] {
												activeTasks = append(activeTasks, task)
											}
										}

										for t, task := range activeTasks {
											taskEmoji := getStatusEmoji(task.Status)
											if task.Pinned {
												taskEmoji = "📌" + taskEmoji
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

											fmt.Printf("%s%s %s - %s [%s]\n", taskPrefix, taskEmoji, task.ID, task.Title, task.Status)
										}
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
									taskEmoji := getStatusEmoji(task.Status)
									if task.Pinned {
										taskEmoji = "📌" + taskEmoji
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

									fmt.Printf("%s%s %s - %s [%s]\n", taskPrefix, taskEmoji, task.ID, task.Title, task.Status)
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

	return cmd
}

func getStatusEmoji(status string) string {
	switch status {
	case "ready":
		return "📦"
	case "paused":
		return "💤"
	case "design":
		return "📐"
	case "implement":
		return "🔨"
	case "deploy":
		return "🚀"
	case "blocked":
		return "🚫"
	case "complete":
		return "✓"
	default:
		return "📦" // default to ready
	}
}
