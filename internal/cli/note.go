package cli

import (
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
	Long: `Create a new note in the ledger.

Notes can be attached to a container (shipment, conclave, or tome) or exist
directly under the commission. If no container flag is provided, the note
is created at the commission level.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		title := args[0]
		commissionID, _ := cmd.Flags().GetString("commission")
		content, _ := cmd.Flags().GetString("content")
		noteType, _ := cmd.Flags().GetString("type")
		shipmentID, _ := cmd.Flags().GetString("shipment")
		conclaveID, _ := cmd.Flags().GetString("conclave")
		tomeID, _ := cmd.Flags().GetString("tome")

		// Validate entity IDs
		if err := validateEntityID(shipmentID, "shipment"); err != nil {
			return err
		}
		if err := validateEntityID(conclaveID, "conclave"); err != nil {
			return err
		}
		if err := validateEntityID(tomeID, "tome"); err != nil {
			return err
		}

		// Get commission from context or require explicit flag
		if commissionID == "" {
			commissionID = orccontext.GetContextCommissionID()
			if commissionID == "" {
				return fmt.Errorf("no commission context detected\nHint: Use --commission flag or run from a workbench directory")
			}
		}

		// Validate note type if specified
		validTypes := map[string]bool{
			"learning": true,
			"concern":  true,
			"finding":  true,
			"frq":      true,
			"bug":      true,
			"spec":     true,
			"roadmap":  true,
			"decision": true,
			"question": true,
			"vision":   true,
			"idea":     true,
			"exorcism": true,
		}
		if noteType != "" && !validTypes[noteType] {
			return fmt.Errorf("invalid note type: %s\nValid types: learning, concern, finding, frq, bug, spec, roadmap, decision, question, vision, idea, exorcism", noteType)
		}

		// Determine container
		containerID := ""
		containerType := ""
		if shipmentID != "" {
			containerID = shipmentID
			containerType = "shipment"
		} else if conclaveID != "" {
			containerID = conclaveID
			containerType = "conclave"
		} else if tomeID != "" {
			containerID = tomeID
			containerType = "tome"
		}

		resp, err := wire.NoteService().CreateNote(ctx, primary.CreateNoteRequest{
			CommissionID:  commissionID,
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
		} else {
			fmt.Printf("  Container: (commission-level)\n")
		}
		fmt.Printf("  Commission: %s\n", note.CommissionID)
		return nil
	},
}

var noteListCmd = &cobra.Command{
	Use:   "list",
	Short: "List notes",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		commissionID, _ := cmd.Flags().GetString("commission")
		noteType, _ := cmd.Flags().GetString("type")
		shipmentID, _ := cmd.Flags().GetString("shipment")
		tomeID, _ := cmd.Flags().GetString("tome")
		commissionOnly, _ := cmd.Flags().GetBool("commission-only")

		// Validate entity IDs
		if err := validateEntityID(shipmentID, "shipment"); err != nil {
			return err
		}
		if err := validateEntityID(tomeID, "tome"); err != nil {
			return err
		}

		// Get commission from context if not specified
		if commissionID == "" {
			commissionID = orccontext.GetContextCommissionID()
		}

		var notes []*primary.Note
		var err error

		// If specific container specified, use container query
		if shipmentID != "" {
			notes, err = wire.NoteService().GetNotesByContainer(ctx, "shipment", shipmentID)
		} else if tomeID != "" {
			notes, err = wire.NoteService().GetNotesByContainer(ctx, "tome", tomeID)
		} else if commissionOnly {
			// List only commission-level notes (not in any container)
			if commissionID == "" {
				return fmt.Errorf("--commission-only requires a commission context or --commission flag")
			}
			notes, err = wire.NoteService().GetNotesByContainer(ctx, "commission", commissionID)
		} else {
			notes, err = wire.NoteService().ListNotes(ctx, primary.NoteFilters{
				Type:         noteType,
				CommissionID: commissionID,
			})
		}

		if err != nil {
			return fmt.Errorf("failed to list notes: %w", err)
		}

		if len(notes) == 0 {
			fmt.Println("No notes found.")
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
			} else if n.TomeID != "" {
				container = n.TomeID
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
		ctx := NewContext()
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
		fmt.Printf("Commission: %s\n", note.CommissionID)
		if note.ShipmentID != "" {
			fmt.Printf("Shipment: %s\n", note.ShipmentID)
		}
		if note.TomeID != "" {
			fmt.Printf("Tome: %s\n", note.TomeID)
		}
		if note.Pinned {
			fmt.Printf("Pinned: yes\n")
		}
		if note.PromotedFromID != "" {
			fmt.Printf("Promoted from: %s (%s)\n", note.PromotedFromID, note.PromotedFromType)
		}
		if note.CloseReason != "" {
			fmt.Printf("Close reason: %s\n", note.CloseReason)
		}
		if note.ClosedByNoteID != "" {
			fmt.Printf("Closed by: %s\n", note.ClosedByNoteID)
		}
		fmt.Printf("Created: %s\n", note.CreatedAt)
		fmt.Printf("Updated: %s\n", note.UpdatedAt)
		if note.ClosedAt != "" {
			fmt.Printf("Closed: %s\n", note.ClosedAt)
		}

		return nil
	},
}

var noteUpdateCmd = &cobra.Command{
	Use:   "update [note-id]",
	Short: "Update note title, content, and/or type",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		noteID := args[0]
		title, _ := cmd.Flags().GetString("title")
		content, _ := cmd.Flags().GetString("content")
		noteType, _ := cmd.Flags().GetString("type")

		if title == "" && content == "" && noteType == "" {
			return fmt.Errorf("must specify --title, --content, and/or --type")
		}

		// Validate note type if specified
		if noteType != "" {
			validTypes := map[string]bool{
				"learning": true,
				"concern":  true,
				"finding":  true,
				"frq":      true,
				"bug":      true,
				"spec":     true,
				"roadmap":  true,
				"decision": true,
				"question": true,
				"vision":   true,
				"idea":     true,
				"exorcism": true,
			}
			if !validTypes[noteType] {
				return fmt.Errorf("invalid note type: %s\nValid types: learning, concern, finding, frq, bug, spec, roadmap, decision, question, vision, idea, exorcism", noteType)
			}
		}

		err := wire.NoteService().UpdateNote(ctx, primary.UpdateNoteRequest{
			NoteID:  noteID,
			Title:   title,
			Content: content,
			Type:    noteType,
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
		ctx := NewContext()
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
		ctx := NewContext()
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
		ctx := NewContext()
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
	Short: "Close a note with a reason",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		noteID := args[0]
		reason, _ := cmd.Flags().GetString("reason")
		byNoteID, _ := cmd.Flags().GetString("by")

		if reason == "" {
			return fmt.Errorf("--reason is required\nValid reasons: superseded, synthesized, resolved, deferred, duplicate, stale")
		}

		err := wire.NoteService().CloseNote(ctx, primary.CloseNoteRequest{
			NoteID:   noteID,
			Reason:   reason,
			ByNoteID: byNoteID,
		})
		if err != nil {
			return fmt.Errorf("failed to close note: %w", err)
		}

		msg := fmt.Sprintf("✓ Note %s closed (reason: %s)", noteID, reason)
		if byNoteID != "" {
			msg += fmt.Sprintf(" by %s", byNoteID)
		}
		fmt.Println(msg)
		return nil
	},
}

var noteReopenCmd = &cobra.Command{
	Use:   "reopen [note-id]",
	Short: "Reopen a closed note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		noteID := args[0]

		err := wire.NoteService().ReopenNote(ctx, noteID)
		if err != nil {
			return fmt.Errorf("failed to reopen note: %w", err)
		}

		fmt.Printf("✓ Note %s reopened\n", noteID)
		return nil
	},
}

var noteMoveCmd = &cobra.Command{
	Use:   "move [note-id]",
	Short: "Move a note to a different container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		noteID := args[0]
		toTome, _ := cmd.Flags().GetString("to-tome")
		toShipment, _ := cmd.Flags().GetString("to-shipment")
		toCommission, _ := cmd.Flags().GetString("to-commission")

		// Validate exactly one target specified
		targetCount := 0
		if toTome != "" {
			targetCount++
		}
		if toShipment != "" {
			targetCount++
		}
		if toCommission != "" {
			targetCount++
		}

		if targetCount == 0 {
			return fmt.Errorf("must specify exactly one target: --to-tome, --to-shipment, or --to-commission")
		}
		if targetCount > 1 {
			return fmt.Errorf("cannot specify multiple targets")
		}

		err := wire.NoteService().MoveNote(ctx, primary.MoveNoteRequest{
			NoteID:         noteID,
			ToTomeID:       toTome,
			ToShipmentID:   toShipment,
			ToCommissionID: toCommission,
		})
		if err != nil {
			return fmt.Errorf("failed to move note: %w", err)
		}

		target := ""
		if toTome != "" {
			target = toTome
		} else if toShipment != "" {
			target = toShipment
		} else {
			target = toCommission + " (commission level)"
		}

		fmt.Printf("✓ Note %s moved to %s\n", noteID, target)
		return nil
	},
}

var noteMergeCmd = &cobra.Command{
	Use:   "merge [source-id] [target-id]",
	Short: "Merge source note into target note",
	Long:  "Merges content from source note into target note, then closes source with a merge reference",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		sourceID := args[0]
		targetID := args[1]

		err := wire.NoteService().MergeNotes(ctx, primary.MergeNoteRequest{
			SourceNoteID: sourceID,
			TargetNoteID: targetID,
		})
		if err != nil {
			return fmt.Errorf("failed to merge notes: %w", err)
		}

		fmt.Printf("✓ Merged %s into %s\n", sourceID, targetID)
		fmt.Printf("  Source %s is now closed\n", sourceID)
		return nil
	},
}

func init() {
	// note create flags
	noteCreateCmd.Flags().StringP("commission", "c", "", "Commission ID (defaults to context)")
	noteCreateCmd.Flags().String("content", "", "Note content")
	noteCreateCmd.Flags().StringP("type", "t", "", "Note type (learning, concern, finding, frq, bug, spec, roadmap, decision, question, vision, idea, exorcism)")
	noteCreateCmd.Flags().String("shipment", "", "Shipment ID to attach note to")
	noteCreateCmd.Flags().String("conclave", "", "Conclave ID to attach note to")
	noteCreateCmd.Flags().String("tome", "", "Tome ID to attach note to")

	// note list flags
	noteListCmd.Flags().StringP("commission", "c", "", "Filter by commission")
	noteListCmd.Flags().StringP("type", "t", "", "Filter by type")
	noteListCmd.Flags().String("shipment", "", "Filter by shipment")
	noteListCmd.Flags().String("tome", "", "Filter by tome")
	noteListCmd.Flags().Bool("commission-only", false, "List only commission-level notes (not in any container)")

	// note update flags
	noteUpdateCmd.Flags().String("title", "", "New title")
	noteUpdateCmd.Flags().String("content", "", "New content")
	noteUpdateCmd.Flags().String("type", "", "Note type (learning, concern, finding, frq, bug, spec, roadmap, decision, question, vision, idea, exorcism)")

	// note move flags
	noteMoveCmd.Flags().String("to-tome", "", "Move to tome")
	noteMoveCmd.Flags().String("to-shipment", "", "Move to shipment")
	noteMoveCmd.Flags().String("to-conclave", "", "Move to conclave")
	noteMoveCmd.Flags().String("to-commission", "", "Promote to commission level (clears container associations)")

	// note close flags
	noteCloseCmd.Flags().StringP("reason", "r", "", "Close reason (required): superseded, synthesized, resolved, deferred, duplicate, stale")
	noteCloseCmd.Flags().String("by", "", "Reference to another note (optional)")

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
	noteCmd.AddCommand(noteMoveCmd)
	noteCmd.AddCommand(noteMergeCmd)
}

// NoteCmd returns the note command
func NoteCmd() *cobra.Command {
	return noteCmd
}
