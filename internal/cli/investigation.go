package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/example/orc/internal/context"
	"github.com/example/orc/internal/models"
	"github.com/spf13/cobra"
)

var investigationCmd = &cobra.Command{
	Use:     "investigation",
	Aliases: []string{"inv"},
	Short:   "Manage investigations (research containers)",
	Long:    "Create, list, complete, and manage investigations in the ORC ledger",
}

var investigationCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new investigation",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		missionID, _ := cmd.Flags().GetString("mission")
		description, _ := cmd.Flags().GetString("description")

		// Get mission from context or require explicit flag
		if missionID == "" {
			missionID = context.GetContextMissionID()
			if missionID == "" {
				return fmt.Errorf("no mission context detected\nHint: Use --mission flag or run from a grove/mission directory")
			}
		}

		investigation, err := models.CreateInvestigation(missionID, title, description)
		if err != nil {
			return fmt.Errorf("failed to create investigation: %w", err)
		}

		fmt.Printf("✓ Created investigation %s: %s\n", investigation.ID, investigation.Title)
		fmt.Printf("  Mission: %s\n", investigation.MissionID)
		fmt.Printf("  Status: %s\n", investigation.Status)
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Printf("   orc question create \"Question title\" --investigation %s\n", investigation.ID)
		return nil
	},
}

var investigationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List investigations",
	RunE: func(cmd *cobra.Command, args []string) error {
		missionID, _ := cmd.Flags().GetString("mission")
		status, _ := cmd.Flags().GetString("status")

		// Get mission from context if not specified
		if missionID == "" {
			missionID = context.GetContextMissionID()
		}

		investigations, err := models.ListInvestigations(missionID, status)
		if err != nil {
			return fmt.Errorf("failed to list investigations: %w", err)
		}

		if len(investigations) == 0 {
			fmt.Println("No investigations found")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tMISSION")
		fmt.Fprintln(w, "--\t-----\t------\t-------")
		for _, inv := range investigations {
			pinnedMark := ""
			if inv.Pinned {
				pinnedMark = " [pinned]"
			}
			statusIcon := "🔍"
			if inv.Status == "complete" {
				statusIcon = "✅"
			}
			fmt.Fprintf(w, "%s\t%s%s\t%s %s\t%s\n", inv.ID, inv.Title, pinnedMark, statusIcon, inv.Status, inv.MissionID)
		}
		w.Flush()
		return nil
	},
}

var investigationShowCmd = &cobra.Command{
	Use:   "show [investigation-id]",
	Short: "Show investigation details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		investigationID := args[0]

		investigation, err := models.GetInvestigation(investigationID)
		if err != nil {
			return fmt.Errorf("investigation not found: %w", err)
		}

		fmt.Printf("Investigation: %s\n", investigation.ID)
		fmt.Printf("Title: %s\n", investigation.Title)
		if investigation.Description.Valid {
			fmt.Printf("Description: %s\n", investigation.Description.String)
		}
		fmt.Printf("Status: %s\n", investigation.Status)
		fmt.Printf("Mission: %s\n", investigation.MissionID)
		if investigation.AssignedGroveID.Valid {
			fmt.Printf("Assigned Grove: %s\n", investigation.AssignedGroveID.String)
		}
		if investigation.Pinned {
			fmt.Printf("Pinned: yes\n")
		}
		fmt.Printf("Created: %s\n", investigation.CreatedAt.Format("2006-01-02 15:04"))
		if investigation.CompletedAt.Valid {
			fmt.Printf("Completed: %s\n", investigation.CompletedAt.Time.Format("2006-01-02 15:04"))
		}

		// Show questions in this investigation
		questions, err := models.GetInvestigationQuestions(investigationID)
		if err != nil {
			return fmt.Errorf("failed to get questions: %w", err)
		}

		if len(questions) > 0 {
			fmt.Printf("\nQuestions (%d):\n", len(questions))
			for _, q := range questions {
				statusIcon := "❓"
				if q.Status == "answered" {
					statusIcon = "✅"
				}
				fmt.Printf("  %s %s: %s [%s]\n", statusIcon, q.ID, q.Title, q.Status)
			}
		}

		return nil
	},
}

var investigationCompleteCmd = &cobra.Command{
	Use:   "complete [investigation-id]",
	Short: "Mark investigation as complete",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		investigationID := args[0]

		err := models.CompleteInvestigation(investigationID)
		if err != nil {
			return fmt.Errorf("failed to complete investigation: %w", err)
		}

		fmt.Printf("✓ Investigation %s marked as complete\n", investigationID)
		return nil
	},
}

var investigationUpdateCmd = &cobra.Command{
	Use:   "update [investigation-id]",
	Short: "Update investigation title and/or description",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		investigationID := args[0]
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")

		if title == "" && description == "" {
			return fmt.Errorf("must specify --title and/or --description")
		}

		err := models.UpdateInvestigation(investigationID, title, description)
		if err != nil {
			return fmt.Errorf("failed to update investigation: %w", err)
		}

		fmt.Printf("✓ Investigation %s updated\n", investigationID)
		return nil
	},
}

var investigationPinCmd = &cobra.Command{
	Use:   "pin [investigation-id]",
	Short: "Pin investigation to keep it visible",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		investigationID := args[0]

		err := models.PinInvestigation(investigationID)
		if err != nil {
			return fmt.Errorf("failed to pin investigation: %w", err)
		}

		fmt.Printf("✓ Investigation %s pinned 📌\n", investigationID)
		return nil
	},
}

var investigationUnpinCmd = &cobra.Command{
	Use:   "unpin [investigation-id]",
	Short: "Unpin investigation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		investigationID := args[0]

		err := models.UnpinInvestigation(investigationID)
		if err != nil {
			return fmt.Errorf("failed to unpin investigation: %w", err)
		}

		fmt.Printf("✓ Investigation %s unpinned\n", investigationID)
		return nil
	},
}

var investigationDeleteCmd = &cobra.Command{
	Use:   "delete [investigation-id]",
	Short: "Delete an investigation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		investigationID := args[0]

		err := models.DeleteInvestigation(investigationID)
		if err != nil {
			return fmt.Errorf("failed to delete investigation: %w", err)
		}

		fmt.Printf("✓ Investigation %s deleted\n", investigationID)
		return nil
	},
}

var investigationAssignCmd = &cobra.Command{
	Use:   "assign [investigation-id] [grove-id]",
	Short: "Assign investigation to a grove",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		investigationID := args[0]
		groveID := args[1]

		err := models.AssignInvestigationToGrove(investigationID, groveID)
		if err != nil {
			return fmt.Errorf("failed to assign investigation: %w", err)
		}

		fmt.Printf("✓ Investigation %s assigned to grove %s\n", investigationID, groveID)
		return nil
	},
}

func init() {
	// investigation create flags
	investigationCreateCmd.Flags().StringP("mission", "m", "", "Mission ID (defaults to context)")
	investigationCreateCmd.Flags().StringP("description", "d", "", "Investigation description")

	// investigation list flags
	investigationListCmd.Flags().StringP("mission", "m", "", "Filter by mission")
	investigationListCmd.Flags().StringP("status", "s", "", "Filter by status (active, complete)")

	// investigation update flags
	investigationUpdateCmd.Flags().String("title", "", "New title")
	investigationUpdateCmd.Flags().StringP("description", "d", "", "New description")

	// Register subcommands
	investigationCmd.AddCommand(investigationCreateCmd)
	investigationCmd.AddCommand(investigationListCmd)
	investigationCmd.AddCommand(investigationShowCmd)
	investigationCmd.AddCommand(investigationCompleteCmd)
	investigationCmd.AddCommand(investigationUpdateCmd)
	investigationCmd.AddCommand(investigationPinCmd)
	investigationCmd.AddCommand(investigationUnpinCmd)
	investigationCmd.AddCommand(investigationDeleteCmd)
	investigationCmd.AddCommand(investigationAssignCmd)
}

// InvestigationCmd returns the investigation command
func InvestigationCmd() *cobra.Command {
	return investigationCmd
}
