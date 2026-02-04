package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "View workshop activity logs",
	Long:  "View, search, and manage workshop activity logs (audit trail)",
}

var logTailCmd = &cobra.Command{
	Use:   "tail",
	Short: "Show recent activity",
	Long:  "Show recent activity log entries (default 50)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		limit, _ := cmd.Flags().GetInt("limit")
		workshopID, _ := cmd.Flags().GetString("workshop")
		actorID, _ := cmd.Flags().GetString("actor")
		entityType, _ := cmd.Flags().GetString("type")
		follow, _ := cmd.Flags().GetBool("follow")

		if limit <= 0 {
			limit = 50
		}

		filters := primary.LogFilters{
			WorkshopID: workshopID,
			ActorID:    actorID,
			EntityType: entityType,
			Limit:      limit,
		}

		// Initial fetch
		entries, err := wire.LogService().ListLogs(ctx, filters)
		if err != nil {
			return fmt.Errorf("failed to fetch logs: %w", err)
		}

		printLogEntries(entries)

		// If --follow, poll for new entries
		if follow {
			var lastTimestamp string
			if len(entries) > 0 {
				lastTimestamp = entries[0].Timestamp
			}

			for {
				time.Sleep(1 * time.Second)

				newEntries, err := wire.LogService().ListLogs(ctx, filters)
				if err != nil {
					fmt.Printf("Error fetching logs: %v\n", err)
					continue
				}

				// Print only entries newer than lastTimestamp
				for i := len(newEntries) - 1; i >= 0; i-- {
					entry := newEntries[i]
					if lastTimestamp == "" || entry.Timestamp > lastTimestamp {
						printLogEntry(entry)
						if entry.Timestamp > lastTimestamp {
							lastTimestamp = entry.Timestamp
						}
					}
				}
			}
		}

		return nil
	},
}

var logShowCmd = &cobra.Command{
	Use:   "show [entity-id]",
	Short: "Show activity for a specific entity",
	Long:  "Show activity history for a specific entity (e.g., SHIP-243, TASK-001)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		actorID, _ := cmd.Flags().GetString("actor")
		limit, _ := cmd.Flags().GetInt("limit")

		filters := primary.LogFilters{
			ActorID: actorID,
			Limit:   limit,
		}

		// If entity ID provided, filter by it
		if len(args) > 0 {
			filters.EntityID = args[0]
		}

		entries, err := wire.LogService().ListLogs(ctx, filters)
		if err != nil {
			return fmt.Errorf("failed to fetch logs: %w", err)
		}

		if len(entries) == 0 {
			fmt.Println("No log entries found.")
			return nil
		}

		printLogEntries(entries)
		return nil
	},
}

var logPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Delete old log entries",
	Long:  "Delete log entries older than the specified number of days (default 30)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		days, _ := cmd.Flags().GetInt("days")

		if days <= 0 {
			days = 30
		}

		count, err := wire.LogService().PruneLogs(ctx, days)
		if err != nil {
			return fmt.Errorf("failed to prune logs: %w", err)
		}

		if count == 0 {
			fmt.Printf("No log entries older than %d days found.\n", days)
		} else {
			fmt.Printf("Pruned %d log entries older than %d days.\n", count, days)
		}
		return nil
	},
}

func printLogEntries(entries []*primary.LogEntry) {
	if len(entries) == 0 {
		fmt.Println("No log entries found.")
		return
	}

	fmt.Printf("Found %d log entries:\n\n", len(entries))

	// Print in reverse order (oldest first) for tail view
	for i := len(entries) - 1; i >= 0; i-- {
		printLogEntry(entries[i])
	}
}

func printLogEntry(entry *primary.LogEntry) {
	// Format: timestamp | actor | action | entity_type/entity_id | field changes
	actorStr := entry.ActorID
	if actorStr == "" {
		actorStr = "-"
	}

	actionIcon := getActionIcon(entry.Action)

	// Base line
	fmt.Printf("%s | %-12s | %s %s | %s/%s",
		formatTimestamp(entry.Timestamp),
		actorStr,
		actionIcon,
		entry.Action,
		entry.EntityType,
		entry.EntityID,
	)

	// Field changes for updates
	if entry.Action == "update" && entry.FieldName != "" {
		fmt.Printf(" | %s: %s -> %s", entry.FieldName, entry.OldValue, entry.NewValue)
	}

	fmt.Println()
}

func getActionIcon(action string) string {
	switch action {
	case "create":
		return "+"
	case "update":
		return "~"
	case "delete":
		return "-"
	default:
		return "?"
	}
}

func formatTimestamp(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}
	return t.Format("2006-01-02 15:04:05")
}

// LogCmd returns the log command with all subcommands attached.
func LogCmd() *cobra.Command {
	// log tail
	logTailCmd.Flags().IntP("limit", "n", 50, "Number of entries to show")
	logTailCmd.Flags().String("workshop", "", "Filter by workshop ID")
	logTailCmd.Flags().String("actor", "", "Filter by actor ID")
	logTailCmd.Flags().String("type", "", "Filter by entity type")
	logTailCmd.Flags().BoolP("follow", "f", false, "Follow mode: poll for new entries")

	// log show
	logShowCmd.Flags().String("actor", "", "Filter by actor ID")
	logShowCmd.Flags().IntP("limit", "n", 100, "Maximum entries to show")

	// log prune
	logPruneCmd.Flags().Int("days", 30, "Delete entries older than N days")

	logCmd.AddCommand(logTailCmd)
	logCmd.AddCommand(logShowCmd)
	logCmd.AddCommand(logPruneCmd)

	return logCmd
}
