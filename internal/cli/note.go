package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/example/orc/internal/context"
	"github.com/example/orc/internal/models"
	"github.com/spf13/cobra"
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
			missionID = context.GetContextMissionID()
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

		note, err := models.CreateNote(missionID, title, content, noteType, containerID, containerType)
		if err != nil {
			return fmt.Errorf("failed to create note: %w", err)
		}

		fmt.Printf("✓ Created note %s: %s\n", note.ID, note.Title)
		if note.Type.Valid {
			fmt.Printf("  Type: %s\n", note.Type.String)
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
		missionID, _ := cmd.Flags().GetString("mission")
		noteType, _ := cmd.Flags().GetString("type")
		shipmentID, _ := cmd.Flags().GetString("shipment")
		investigationID, _ := cmd.Flags().GetString("investigation")
		tomeID, _ := cmd.Flags().GetString("tome")

		// Get mission from context if not specified
		if missionID == "" {
			missionID = context.GetContextMissionID()
		}

		var notes []*models.Note
		var err error

		// If specific container specified, use container query
		if shipmentID != "" {
			notes, err = models.GetNotesByContainer("shipment", shipmentID)
		} else if investigationID != "" {
			notes, err = models.GetNotesByContainer("investigation", investigationID)
		} else if tomeID != "" {
			notes, err = models.GetNotesByContainer("tome", tomeID)
		} else {
			notes, err = models.ListNotes(noteType, missionID)
		}

		if err != nil {
			return fmt.Errorf("failed to list notes: %w", err)
		}

		if len(notes) == 0 {
			fmt.Println("No notes found")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tTYPE\tCONTAINER")
		fmt.Fprintln(w, "--\t-----\t----\t---------")
		for _, n := range notes {
			pinnedMark := ""
			if n.Pinned {
				pinnedMark = " [pinned]"
			}
			typeStr := "-"
			if n.Type.Valid {
				typeStr = n.Type.String
			}
			container := "-"
			if n.ShipmentID.Valid {
				container = n.ShipmentID.String
			} else if n.InvestigationID.Valid {
				container = n.InvestigationID.String
			} else if n.TomeID.Valid {
				container = n.TomeID.String
			} else if n.ConclaveID.Valid {
				container = n.ConclaveID.String
			}
			fmt.Fprintf(w, "%s\t%s%s\t%s\t%s\n", n.ID, n.Title, pinnedMark, typeStr, container)
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
		noteID := args[0]

		note, err := models.GetNote(noteID)
		if err != nil {
			return fmt.Errorf("note not found: %w", err)
		}

		fmt.Printf("Note: %s\n", note.ID)
		fmt.Printf("Title: %s\n", note.Title)
		if note.Content.Valid {
			fmt.Printf("Content: %s\n", note.Content.String)
		}
		if note.Type.Valid {
			fmt.Printf("Type: %s\n", note.Type.String)
		}
		fmt.Printf("Mission: %s\n", note.MissionID)
		if note.ShipmentID.Valid {
			fmt.Printf("Shipment: %s\n", note.ShipmentID.String)
		}
		if note.InvestigationID.Valid {
			fmt.Printf("Investigation: %s\n", note.InvestigationID.String)
		}
		if note.TomeID.Valid {
			fmt.Printf("Tome: %s\n", note.TomeID.String)
		}
		if note.ConclaveID.Valid {
			fmt.Printf("Conclave: %s\n", note.ConclaveID.String)
		}
		if note.Pinned {
			fmt.Printf("Pinned: yes\n")
		}
		if note.PromotedFromID.Valid {
			fmt.Printf("Promoted from: %s (%s)\n", note.PromotedFromID.String, note.PromotedFromType.String)
		}
		fmt.Printf("Created: %s\n", note.CreatedAt.Format("2006-01-02 15:04"))
		fmt.Printf("Updated: %s\n", note.UpdatedAt.Format("2006-01-02 15:04"))

		return nil
	},
}

var noteUpdateCmd = &cobra.Command{
	Use:   "update [note-id]",
	Short: "Update note title and/or content",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID := args[0]
		title, _ := cmd.Flags().GetString("title")
		content, _ := cmd.Flags().GetString("content")

		if title == "" && content == "" {
			return fmt.Errorf("must specify --title and/or --content")
		}

		err := models.UpdateNote(noteID, title, content)
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
		noteID := args[0]

		err := models.PinNote(noteID)
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
		noteID := args[0]

		err := models.UnpinNote(noteID)
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
		noteID := args[0]

		err := models.DeleteNote(noteID)
		if err != nil {
			return fmt.Errorf("failed to delete note: %w", err)
		}

		fmt.Printf("✓ Note %s deleted\n", noteID)
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
}

// NoteCmd returns the note command
func NoteCmd() *cobra.Command {
	return noteCmd
}
