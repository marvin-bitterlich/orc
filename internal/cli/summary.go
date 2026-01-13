package cli

import (
	"fmt"

	"github.com/looneym/orc/internal/models"
	"github.com/spf13/cobra"
)

// SummaryCmd returns the summary command
func SummaryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "summary",
		Short: "Show summary of all open missions and work orders",
		Long: `Display a high-level overview of all work in progress:
- Open missions with their work orders

This provides a global view of all work across ORC.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get all non-complete missions
			missions, err := models.ListMissions("")
			if err != nil {
				return fmt.Errorf("failed to list missions: %w", err)
			}

			// Filter to open missions (not complete or archived)
			var openMissions []*models.Mission
			for _, m := range missions {
				if m.Status != "complete" && m.Status != "archived" {
					openMissions = append(openMissions, m)
				}
			}

			if len(openMissions) == 0 {
				fmt.Println("No open missions")
				return nil
			}

			fmt.Println("📊 ORC Summary - Open Work")
			fmt.Println()

			// Display each mission with its work orders in tree format
			for i, mission := range openMissions {
				// Display mission
				statusEmoji := getStatusEmoji(mission.Status)
				fmt.Printf("%s %s - %s [%s]\n", statusEmoji, mission.ID, mission.Title, mission.Status)

				// Get work orders for this mission
				workOrders, err := models.ListWorkOrders(mission.ID, "")
				if err != nil {
					return fmt.Errorf("failed to list work orders for %s: %w", mission.ID, err)
				}

				// Filter to non-complete work orders
				var activeWOs []*models.WorkOrder
				for _, wo := range workOrders {
					if wo.Status != "complete" {
						activeWOs = append(activeWOs, wo)
					}
				}

				if len(activeWOs) > 0 {
					for j, wo := range activeWOs {
						woEmoji := getStatusEmoji(wo.Status)

						// Determine tree characters
						var prefix string
						if j < len(activeWOs)-1 {
							prefix = "├── "
						} else {
							prefix = "└── "
						}

						groveInfo := ""
						if wo.AssignedGroveID.Valid {
							groveInfo = fmt.Sprintf(" [Grove: %s]", wo.AssignedGroveID.String)
						}

						fmt.Printf("%s%s %s - %s [%s]%s\n", prefix, woEmoji, wo.ID, wo.Title, wo.Status, groveInfo)
					}
				} else {
					fmt.Println("└── (No active work orders)")
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
}

func getStatusEmoji(status string) string {
	switch status {
	case "active", "in_progress":
		return "🚀"
	case "planning":
		return "📋"
	case "paused":
		return "⏸️"
	case "backlog":
		return "📦"
	case "next":
		return "⏭️"
	case "complete":
		return "✅"
	case "cancelled", "archived":
		return "🗄️"
	default:
		return "•"
	}
}
