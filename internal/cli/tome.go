package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	orccontext "github.com/example/orc/internal/context"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

var tomeCmd = &cobra.Command{
	Use:   "tome",
	Short: "Manage tomes (knowledge organization containers)",
	Long:  "Create, list, complete, and manage tomes in the ORC ledger",
}

var tomeCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new tome",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		title := args[0]
		missionID, _ := cmd.Flags().GetString("mission")
		description, _ := cmd.Flags().GetString("description")

		// Get mission from context or require explicit flag
		if missionID == "" {
			missionID = orccontext.GetContextMissionID()
			if missionID == "" {
				return fmt.Errorf("no mission context detected\nHint: Use --mission flag or run from a grove/mission directory")
			}
		}

		resp, err := wire.TomeService().CreateTome(ctx, primary.CreateTomeRequest{
			MissionID:   missionID,
			Title:       title,
			Description: description,
		})
		if err != nil {
			return fmt.Errorf("failed to create tome: %w", err)
		}

		tome := resp.Tome
		fmt.Printf("✓ Created tome %s: %s\n", tome.ID, tome.Title)
		fmt.Printf("  Mission: %s\n", tome.MissionID)
		fmt.Printf("  Status: %s\n", tome.Status)
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Printf("   orc note create \"Learning title\" --tome %s --type learning\n", tome.ID)
		return nil
	},
}

var tomeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tomes",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		missionID, _ := cmd.Flags().GetString("mission")
		status, _ := cmd.Flags().GetString("status")

		// Get mission from context if not specified
		if missionID == "" {
			missionID = orccontext.GetContextMissionID()
		}

		tomes, err := wire.TomeService().ListTomes(ctx, primary.TomeFilters{
			MissionID: missionID,
			Status:    status,
		})
		if err != nil {
			return fmt.Errorf("failed to list tomes: %w", err)
		}

		if len(tomes) == 0 {
			fmt.Println("No tomes found")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tMISSION")
		fmt.Fprintln(w, "--\t-----\t------\t-------")
		for _, t := range tomes {
			pinnedMark := ""
			if t.Pinned {
				pinnedMark = " [pinned]"
			}
			statusIcon := "📚"
			if t.Status == "complete" {
				statusIcon = "✅"
			}
			fmt.Fprintf(w, "%s\t%s%s\t%s %s\t%s\n", t.ID, t.Title, pinnedMark, statusIcon, t.Status, t.MissionID)
		}
		w.Flush()
		return nil
	},
}

var tomeShowCmd = &cobra.Command{
	Use:   "show [tome-id]",
	Short: "Show tome details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		tomeID := args[0]

		tome, err := wire.TomeService().GetTome(ctx, tomeID)
		if err != nil {
			return fmt.Errorf("tome not found: %w", err)
		}

		fmt.Printf("Tome: %s\n", tome.ID)
		fmt.Printf("Title: %s\n", tome.Title)
		if tome.Description != "" {
			fmt.Printf("Description: %s\n", tome.Description)
		}
		fmt.Printf("Status: %s\n", tome.Status)
		fmt.Printf("Mission: %s\n", tome.MissionID)
		if tome.Pinned {
			fmt.Printf("Pinned: yes\n")
		}
		fmt.Printf("Created: %s\n", tome.CreatedAt)
		if tome.CompletedAt != "" {
			fmt.Printf("Completed: %s\n", tome.CompletedAt)
		}

		// Show notes in this tome
		notes, err := wire.TomeService().GetTomeNotes(ctx, tomeID)
		if err != nil {
			return fmt.Errorf("failed to get notes: %w", err)
		}

		if len(notes) > 0 {
			fmt.Printf("\nNotes (%d):\n", len(notes))
			for _, note := range notes {
				typeStr := ""
				if note.Type != "" {
					typeStr = fmt.Sprintf(" [%s]", note.Type)
				}
				fmt.Printf("  📝 %s: %s%s\n", note.ID, note.Title, typeStr)
			}
		}

		return nil
	},
}

var tomeCompleteCmd = &cobra.Command{
	Use:   "complete [tome-id]",
	Short: "Mark tome as complete",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		tomeID := args[0]

		err := wire.TomeService().CompleteTome(ctx, tomeID)
		if err != nil {
			return fmt.Errorf("failed to complete tome: %w", err)
		}

		fmt.Printf("✓ Tome %s marked as complete\n", tomeID)
		return nil
	},
}

var tomePauseCmd = &cobra.Command{
	Use:   "pause [tome-id]",
	Short: "Pause an active tome",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		tomeID := args[0]

		err := wire.TomeService().PauseTome(ctx, tomeID)
		if err != nil {
			return fmt.Errorf("failed to pause tome: %w", err)
		}

		fmt.Printf("✓ Tome %s paused\n", tomeID)
		return nil
	},
}

var tomeResumeCmd = &cobra.Command{
	Use:   "resume [tome-id]",
	Short: "Resume a paused tome",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		tomeID := args[0]

		err := wire.TomeService().ResumeTome(ctx, tomeID)
		if err != nil {
			return fmt.Errorf("failed to resume tome: %w", err)
		}

		fmt.Printf("✓ Tome %s resumed\n", tomeID)
		return nil
	},
}

var tomeUpdateCmd = &cobra.Command{
	Use:   "update [tome-id]",
	Short: "Update tome title and/or description",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		tomeID := args[0]
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")

		if title == "" && description == "" {
			return fmt.Errorf("must specify --title and/or --description")
		}

		err := wire.TomeService().UpdateTome(ctx, primary.UpdateTomeRequest{
			TomeID:      tomeID,
			Title:       title,
			Description: description,
		})
		if err != nil {
			return fmt.Errorf("failed to update tome: %w", err)
		}

		fmt.Printf("✓ Tome %s updated\n", tomeID)
		return nil
	},
}

var tomePinCmd = &cobra.Command{
	Use:   "pin [tome-id]",
	Short: "Pin tome to keep it visible",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		tomeID := args[0]

		err := wire.TomeService().PinTome(ctx, tomeID)
		if err != nil {
			return fmt.Errorf("failed to pin tome: %w", err)
		}

		fmt.Printf("✓ Tome %s pinned 📌\n", tomeID)
		return nil
	},
}

var tomeUnpinCmd = &cobra.Command{
	Use:   "unpin [tome-id]",
	Short: "Unpin tome",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		tomeID := args[0]

		err := wire.TomeService().UnpinTome(ctx, tomeID)
		if err != nil {
			return fmt.Errorf("failed to unpin tome: %w", err)
		}

		fmt.Printf("✓ Tome %s unpinned\n", tomeID)
		return nil
	},
}

var tomeDeleteCmd = &cobra.Command{
	Use:   "delete [tome-id]",
	Short: "Delete a tome",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		tomeID := args[0]

		err := wire.TomeService().DeleteTome(ctx, tomeID)
		if err != nil {
			return fmt.Errorf("failed to delete tome: %w", err)
		}

		fmt.Printf("✓ Tome %s deleted\n", tomeID)
		return nil
	},
}

func init() {
	// tome create flags
	tomeCreateCmd.Flags().StringP("mission", "m", "", "Mission ID (defaults to context)")
	tomeCreateCmd.Flags().StringP("description", "d", "", "Tome description")

	// tome list flags
	tomeListCmd.Flags().StringP("mission", "m", "", "Filter by mission")
	tomeListCmd.Flags().StringP("status", "s", "", "Filter by status (active, complete)")

	// tome update flags
	tomeUpdateCmd.Flags().String("title", "", "New title")
	tomeUpdateCmd.Flags().StringP("description", "d", "", "New description")

	// Register subcommands
	tomeCmd.AddCommand(tomeCreateCmd)
	tomeCmd.AddCommand(tomeListCmd)
	tomeCmd.AddCommand(tomeShowCmd)
	tomeCmd.AddCommand(tomeCompleteCmd)
	tomeCmd.AddCommand(tomePauseCmd)
	tomeCmd.AddCommand(tomeResumeCmd)
	tomeCmd.AddCommand(tomeUpdateCmd)
	tomeCmd.AddCommand(tomePinCmd)
	tomeCmd.AddCommand(tomeUnpinCmd)
	tomeCmd.AddCommand(tomeDeleteCmd)
}

// TomeCmd returns the tome command
func TomeCmd() *cobra.Command {
	return tomeCmd
}
