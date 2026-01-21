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

var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "Manage notes (learnings, concerns, findings)",
	Long:  "Create, list, update, and manage notes in the ORC ledger",
}

var noteCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new note",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		title := args[0]
		missionID, _ := cmd.Flags().GetString("mission")
		content, _ := cmd.Flags().GetString("content")
		noteType, _ := cmd.Flags().GetString("type")
		shipmentID, _ := cmd.Flags().GetString("shipment")
		investigationID, _ := cmd.Flags().GetString("investigation")
		conclaveID, _ := cmd.Flags().GetString("conclave")
		tomeID, _ := cmd.Flags().GetString("tome")

		// Get mission from context or require explicit flag
		if missionID == "" {
			missionID = orccontext.GetContextMissionID()
			if missionID == "" {
				return fmt.Errorf("no mission context detected\nHint: Use --mission flag or run from a grove/mission directory")
			}
		}

		// Validate note type if specified
		validTypes := map[string]bool{
			"learning":             true,
			"concern":              true,
			"finding":              true,
			"frq":                  true,
			"bug":                  true,
			"investigation_report": true,
		}
		if noteType != "" && !validTypes[noteType] {
			return fmt.Errorf("invalid note type: %s\nValid types: learning, concern, finding, frq, bug, investigation_report", noteType)
		}

		// Determine container
		containerID := ""
		containerType := ""
		if shipmentID != "" {
			containerID = shipmentID
			containerType = "shipment"
		} else if investigationID != "" {
			containerID = investigationID
			containerType = "investigation"
		} else if conclaveID != "" {
			containerID = conclaveID
			containerType = "conclave"
		} else if tomeID != "" {
			containerID = tomeID
			containerType = "tome"
		}

		resp, err := wire.NoteService().CreateNote(ctx, primary.CreateNoteRequest{
			MissionID:     missionID,
			Title:         title,
			Content:       content,
			Type:          noteType,
			ContainerID:   containerID,
			ContainerType: containerType,
		})
		if err != nil {
			return fmt.Errorf("failed to create note: %w", err)
		}

		note := resp.Note
		fmt.Printf("✓ Created note %s: %s\n", note.ID, note.Title)
		if note.Type != "" {
			fmt.Printf("  Type: %s\n", note.Type)
		}
		if containerID != "" {
			fmt.Printf("  Container: %s (%s)\n", containerID, containerType)
		}
		fmt.Printf("  Mission: %s\n", note.MissionID)
		return nil
	},
}

var noteListCmd = &cobra.Command{
	Use:   "list",
	Short: "List notes",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		missionID, _ := cmd.Flags().GetString("mission")
		noteType, _ := cmd.Flags().GetString("type")
		shipmentID, _ := cmd.Flags().GetString("shipment")
		investigationID, _ := cmd.Flags().GetString("investigation")
		tomeID, _ := cmd.Flags().GetString("tome")

		// Get mission from context if not specified
		if missionID == "" {
			missionID = orccontext.GetContextMissionID()
		}

		var notes []*primary.Note
		var err error

		// If specific container specified, use container query
		if shipmentID != "" {
			notes, err = wire.NoteService().GetNotesByContainer(ctx, "shipment", shipmentID)
		} else if investigationID != "" {
			notes, err = wire.NoteService().GetNotesByContainer(ctx, "investigation", investigationID)
		} else if tomeID != "" {
			notes, err = wire.NoteService().GetNotesByContainer(ctx, "tome", tomeID)
		} else {
			notes, err = wire.NoteService().ListNotes(ctx, primary.NoteFilters{
				Type:      noteType,
				MissionID: missionID,
			})
		}

		if err != nil {
			return fmt.Errorf("failed to list notes: %w", err)
		}

		if len(notes) == 0 {
			fmt.Println("No notes found")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tTYPE\tSTATUS\tCONTAINER")
		fmt.Fprintln(w, "--\t-----\t----\t------\t---------")
		for _, n := range notes {
			pinnedMark := ""
			if n.Pinned {
				pinnedMark = " [pinned]"
			}
			typeStr := "-"
			if n.Type != "" {
				typeStr = n.Type
			}
			statusStr := n.Status
			if statusStr == "" {
				statusStr = "open"
			}
			container := "-"
			if n.ShipmentID != "" {
				container = n.ShipmentID
			} else if n.InvestigationID != "" {
				container = n.InvestigationID
			} else if n.TomeID != "" {
				container = n.TomeID
			} else if n.ConclaveID != "" {
				container = n.ConclaveID
			}
			fmt.Fprintf(w, "%s\t%s%s\t%s\t%s\t%s\n", n.ID, n.Title, pinnedMark, typeStr, statusStr, container)
		}
		w.Flush()
		return nil
	},
}

var noteShowCmd = &cobra.Command{
	Use:   "show [note-id]",
	Short: "Show note details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		noteID := args[0]

		note, err := wire.NoteService().GetNote(ctx, noteID)
		if err != nil {
			return fmt.Errorf("note not found: %w", err)
		}

		fmt.Printf("Note: %s\n", note.ID)
		fmt.Printf("Title: %s\n", note.Title)
		if note.Content != "" {
			fmt.Printf("Content: %s\n", note.Content)
		}
		if note.Type != "" {
			fmt.Printf("Type: %s\n", note.Type)
		}
		status := note.Status
		if status == "" {
			status = "open"
		}
		fmt.Printf("Status: %s\n", status)
		fmt.Printf("Mission: %s\n", note.MissionID)
		if note.ShipmentID != "" {
			fmt.Printf("Shipment: %s\n", note.ShipmentID)
		}
		if note.InvestigationID != "" {
			fmt.Printf("Investigation: %s\n", note.InvestigationID)
		}
		if note.TomeID != "" {
			fmt.Printf("Tome: %s\n", note.TomeID)
		}
		if note.ConclaveID != "" {
			fmt.Printf("Conclave: %s\n", note.ConclaveID)
		}
		if note.Pinned {
			fmt.Printf("Pinned: yes\n")
		}
		if note.PromotedFromID != "" {
			fmt.Printf("Promoted from: %s (%s)\n", note.PromotedFromID, note.PromotedFromType)
		}
		fmt.Printf("Created: %s\n", note.CreatedAt)
		fmt.Printf("Updated: %s\n", note.UpdatedAt)

		return nil
	},
}

var noteUpdateCmd = &cobra.Command{
	Use:   "update [note-id]",
	Short: "Update note title and/or content",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		noteID := args[0]
		title, _ := cmd.Flags().GetString("title")
		content, _ := cmd.Flags().GetString("content")

		if title == "" && content == "" {
			return fmt.Errorf("must specify --title and/or --content")
		}

		err := wire.NoteService().UpdateNote(ctx, primary.UpdateNoteRequest{
			NoteID:  noteID,
			Title:   title,
			Content: content,
		})
		if err != nil {
			return fmt.Errorf("failed to update note: %w", err)
		}

		fmt.Printf("✓ Note %s updated\n", noteID)
		return nil
	},
}

var notePinCmd = &cobra.Command{
	Use:   "pin [note-id]",
	Short: "Pin note to keep it visible",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		noteID := args[0]

		err := wire.NoteService().PinNote(ctx, noteID)
		if err != nil {
			return fmt.Errorf("failed to pin note: %w", err)
		}

		fmt.Printf("✓ Note %s pinned 📌\n", noteID)
		return nil
	},
}

var noteUnpinCmd = &cobra.Command{
	Use:   "unpin [note-id]",
	Short: "Unpin note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		noteID := args[0]

		err := wire.NoteService().UnpinNote(ctx, noteID)
		if err != nil {
			return fmt.Errorf("failed to unpin note: %w", err)
		}

		fmt.Printf("✓ Note %s unpinned\n", noteID)
		return nil
	},
}

var noteDeleteCmd = &cobra.Command{
	Use:   "delete [note-id]",
	Short: "Delete a note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		noteID := args[0]

		err := wire.NoteService().DeleteNote(ctx, noteID)
		if err != nil {
			return fmt.Errorf("failed to delete note: %w", err)
		}

		fmt.Printf("✓ Note %s deleted\n", noteID)
		return nil
	},
}

var noteCloseCmd = &cobra.Command{
	Use:   "close [note-id]",
	Short: "Close a note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		noteID := args[0]

		err := wire.NoteService().CloseNote(ctx, noteID)
		if err != nil {
			return fmt.Errorf("failed to close note: %w", err)
		}

		fmt.Printf("✓ Note %s closed\n", noteID)
		return nil
	},
}

var noteReopenCmd = &cobra.Command{
	Use:   "reopen [note-id]",
	Short: "Reopen a closed note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		noteID := args[0]

		err := wire.NoteService().ReopenNote(ctx, noteID)
		if err != nil {
			return fmt.Errorf("failed to reopen note: %w", err)
		}

		fmt.Printf("✓ Note %s reopened\n", noteID)
		return nil
	},
}

func init() {
	// note create flags
	noteCreateCmd.Flags().StringP("mission", "m", "", "Mission ID (defaults to context)")
	noteCreateCmd.Flags().StringP("content", "c", "", "Note content")
	noteCreateCmd.Flags().StringP("type", "t", "", "Note type (learning, concern, finding, frq, bug, investigation_report)")
	noteCreateCmd.Flags().String("shipment", "", "Shipment ID to attach note to")
	noteCreateCmd.Flags().String("investigation", "", "Investigation ID to attach note to")
	noteCreateCmd.Flags().String("conclave", "", "Conclave ID to attach note to")
	noteCreateCmd.Flags().String("tome", "", "Tome ID to attach note to")

	// note list flags
	noteListCmd.Flags().StringP("mission", "m", "", "Filter by mission")
	noteListCmd.Flags().StringP("type", "t", "", "Filter by type")
	noteListCmd.Flags().String("shipment", "", "Filter by shipment")
	noteListCmd.Flags().String("investigation", "", "Filter by investigation")
	noteListCmd.Flags().String("tome", "", "Filter by tome")

	// note update flags
	noteUpdateCmd.Flags().String("title", "", "New title")
	noteUpdateCmd.Flags().StringP("content", "c", "", "New content")

	// Register subcommands
	noteCmd.AddCommand(noteCreateCmd)
	noteCmd.AddCommand(noteListCmd)
	noteCmd.AddCommand(noteShowCmd)
	noteCmd.AddCommand(noteUpdateCmd)
	noteCmd.AddCommand(notePinCmd)
	noteCmd.AddCommand(noteUnpinCmd)
	noteCmd.AddCommand(noteDeleteCmd)
	noteCmd.AddCommand(noteCloseCmd)
	noteCmd.AddCommand(noteReopenCmd)
}

// NoteCmd returns the note command
func NoteCmd() *cobra.Command {
	return noteCmd
}
