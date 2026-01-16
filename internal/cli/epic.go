package cli

import (
	"fmt"
	"os"

	"github.com/example/orc/internal/config"
	"github.com/example/orc/internal/context"
	"github.com/example/orc/internal/models"
	"github.com/spf13/cobra"
)

var epicCmd = &cobra.Command{
	Use:   "epic",
	Short: "Manage epics (top-level work containers)",
	Long:  "Create, list, assign, and manage epics in the ORC ledger",
}

var epicCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new epic",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		missionID, _ := cmd.Flags().GetString("mission")
		description, _ := cmd.Flags().GetString("description")
		contextRef, _ := cmd.Flags().GetString("context-ref")

		// Smart default: use mission context if available
		if missionID == "" {
			if ctxMissionID := context.GetContextMissionID(); ctxMissionID != "" {
				missionID = ctxMissionID
				fmt.Printf("ℹ️  Using mission from context: %s\n", missionID)
			} else {
				missionID = "MISSION-001"
			}
		}

		epic, err := models.CreateEpic(missionID, title, description, contextRef)
		if err != nil {
			return fmt.Errorf("failed to create epic: %w", err)
		}

		fmt.Printf("✓ Created epic %s: %s\n", epic.ID, epic.Title)
		fmt.Printf("  Under mission: %s\n", epic.MissionID)
		if epic.ContextRef.Valid {
			fmt.Printf("  Context: %s\n", epic.ContextRef.String)
		}
		fmt.Println()
		fmt.Println("💡 Next steps:")
		fmt.Printf("   # Add tasks directly: orc task create \"Task title\" --epic %s\n", epic.ID)
		fmt.Printf("   # OR create rabbit holes: orc rabbit-hole create \"RH title\" --epic %s\n", epic.ID)
		return nil
	},
}

var epicListCmd = &cobra.Command{
	Use:   "list",
	Short: "List epics",
	RunE: func(cmd *cobra.Command, args []string) error {
		missionID, _ := cmd.Flags().GetString("mission")
		status, _ := cmd.Flags().GetString("status")

		// Use context mission if available
		if missionID == "" {
			if ctxMissionID := context.GetContextMissionID(); ctxMissionID != "" {
				missionID = ctxMissionID
			}
		}

		epics, err := models.ListEpics(missionID, status)
		if err != nil {
			return fmt.Errorf("failed to list epics: %w", err)
		}

		if len(epics) == 0 {
			fmt.Println("No epics found")
			return nil
		}

		fmt.Printf("Found %d epic(s):\n\n", len(epics))
		for _, epic := range epics {
			statusIcon := getStatusIcon(epic.Status)
			pinnedIcon := ""
			if epic.Pinned {
				pinnedIcon = " 📌"
			}
			fmt.Printf("%s %s: %s [%s]%s\n", statusIcon, epic.ID, epic.Title, epic.Status, pinnedIcon)
			fmt.Printf("   Mission: %s\n", epic.MissionID)
			if epic.AssignedGroveID.Valid {
				fmt.Printf("   Grove: %s\n", epic.AssignedGroveID.String)
			}

			// Show child count
			hasRH, _ := models.HasRabbitHoles(epic.ID)
			if hasRH {
				rhs, _ := models.ListRabbitHoles(epic.ID, "")
				fmt.Printf("   Structure: %d rabbit hole(s)\n", len(rhs))
			} else {
				tasks, _ := models.GetDirectTasks(epic.ID)
				fmt.Printf("   Structure: %d direct task(s)\n", len(tasks))
			}
			fmt.Println()
		}
		return nil
	},
}

var epicShowCmd = &cobra.Command{
	Use:   "show [epic-id]",
	Short: "Show epic details with children",
	Long: `Show epic details with children (tasks or rabbit holes).

If no epic-id is provided, shows the currently focused epic from config.json.
Use 'orc epic focus <epic-id>' to set the focused epic.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var epicID string

		if len(args) > 0 {
			epicID = args[0]
		} else {
			// Try to get current_epic from config.json
			epicID = getCurrentEpic()
			if epicID == "" {
				return fmt.Errorf("no epic-id provided and no focused epic set\nHint: Use 'orc epic focus <epic-id>' to set a focused epic")
			}
			fmt.Printf("(using focused epic from config)\n\n")
		}

		epic, err := models.GetEpic(epicID)
		if err != nil {
			return fmt.Errorf("epic not found: %w", err)
		}

		// Display epic details
		fmt.Printf("Epic: %s\n", epic.ID)
		fmt.Printf("Title: %s\n", epic.Title)
		if epic.Description.Valid {
			fmt.Printf("Description: %s\n", epic.Description.String)
		}
		fmt.Printf("Status: %s\n", epic.Status)
		fmt.Printf("Mission: %s\n", epic.MissionID)
		if epic.AssignedGroveID.Valid {
			fmt.Printf("Assigned Grove: %s\n", epic.AssignedGroveID.String)
		}
		if epic.Priority.Valid {
			fmt.Printf("Priority: %s\n", epic.Priority.String)
		}
		if epic.Pinned {
			fmt.Printf("Pinned: yes\n")
		}
		fmt.Printf("Created: %s\n", epic.CreatedAt.Format("2006-01-02 15:04"))
		if epic.CompletedAt.Valid {
			fmt.Printf("Completed: %s\n", epic.CompletedAt.Time.Format("2006-01-02 15:04"))
		}
		fmt.Println()

		// Display children
		hasRH, _ := models.HasRabbitHoles(epic.ID)
		if hasRH {
			// Epic has rabbit holes
			rhs, err := models.ListRabbitHoles(epic.ID, "")
			if err != nil {
				return fmt.Errorf("failed to get rabbit holes: %w", err)
			}

			fmt.Printf("Rabbit Holes (%d):\n", len(rhs))
			for _, rh := range rhs {
				statusIcon := getStatusIcon(rh.Status)
				fmt.Printf("  %s %s: %s [%s]\n", statusIcon, rh.ID, rh.Title, rh.Status)

				// Show tasks under this rabbit hole
				tasks, _ := models.GetRabbitHoleTasks(rh.ID)
				for _, task := range tasks {
					taskIcon := getStatusIcon(task.Status)
					fmt.Printf("    %s %s: %s [%s]\n", taskIcon, task.ID, task.Title, task.Status)
				}
			}
		} else {
			// Epic has direct tasks
			tasks, err := models.GetDirectTasks(epic.ID)
			if err != nil {
				return fmt.Errorf("failed to get tasks: %w", err)
			}

			fmt.Printf("Tasks (%d):\n", len(tasks))
			for _, task := range tasks {
				statusIcon := getStatusIcon(task.Status)
				fmt.Printf("  %s %s: %s [%s]\n", statusIcon, task.ID, task.Title, task.Status)
			}
		}

		return nil
	},
}

var epicCompleteCmd = &cobra.Command{
	Use:   "complete [epic-id]",
	Short: "Mark epic as complete",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		epicID := args[0]

		err := models.CompleteEpic(epicID)
		if err != nil {
			return fmt.Errorf("failed to complete epic: %w", err)
		}

		fmt.Printf("✓ Epic %s marked as complete\n", epicID)
		return nil
	},
}

var epicUpdateCmd = &cobra.Command{
	Use:   "update [epic-id]",
	Short: "Update epic title and/or description",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		epicID := args[0]
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")

		if title == "" && description == "" {
			return fmt.Errorf("must specify --title and/or --description")
		}

		err := models.UpdateEpic(epicID, title, description)
		if err != nil {
			return fmt.Errorf("failed to update epic: %w", err)
		}

		fmt.Printf("✓ Epic %s updated\n", epicID)
		return nil
	},
}

var epicPinCmd = &cobra.Command{
	Use:   "pin [epic-id]",
	Short: "Pin epic to keep it visible",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		epicID := args[0]

		err := models.PinEpic(epicID)
		if err != nil {
			return fmt.Errorf("failed to pin epic: %w", err)
		}

		fmt.Printf("✓ Epic %s pinned 📌\n", epicID)
		return nil
	},
}

var epicUnpinCmd = &cobra.Command{
	Use:   "unpin [epic-id]",
	Short: "Unpin epic",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		epicID := args[0]

		err := models.UnpinEpic(epicID)
		if err != nil {
			return fmt.Errorf("failed to unpin epic: %w", err)
		}

		fmt.Printf("✓ Epic %s unpinned\n", epicID)
		return nil
	},
}

var epicAssignCmd = &cobra.Command{
	Use:   "assign [epic-id]",
	Short: "Assign epic to grove (assigns entire epic + all children)",
	Long: `Assign an EPIC (with optional children) to a grove.

The entire epic (parent + all children - rabbit holes + tasks) is assigned to the grove.
1:1:1 relationship: one grove can only work on one epic.

Examples:
  orc epic assign EPIC-001 --grove GROVE-009`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		epicID := args[0]
		groveID, _ := cmd.Flags().GetString("grove")

		if groveID == "" {
			return fmt.Errorf("--grove flag is required")
		}

		// Get epic
		epic, err := models.GetEpic(epicID)
		if err != nil {
			return fmt.Errorf("epic not found: %w", err)
		}

		// Get grove
		grove, err := models.GetGrove(groveID)
		if err != nil {
			return fmt.Errorf("grove not found: %w", err)
		}

		// Assign epic + all children to grove
		err = models.AssignEpicToGrove(epicID, groveID)
		if err != nil {
			return fmt.Errorf("failed to assign epic: %w", err)
		}

		// Count children for output
		hasRH, _ := models.HasRabbitHoles(epicID)
		var childCount int
		if hasRH {
			rhs, _ := models.ListRabbitHoles(epicID, "")
			childCount = len(rhs)
			// Count tasks under rabbit holes
			var taskCount int
			for _, rh := range rhs {
				tasks, _ := models.GetRabbitHoleTasks(rh.ID)
				taskCount += len(tasks)
			}
			fmt.Printf("✓ Assigned epic %s to %s\n", epicID, groveID)
			fmt.Printf("  Epic: %s\n", epic.Title)
			fmt.Printf("  Rabbit Holes: %d\n", childCount)
			fmt.Printf("  Tasks: %d\n", taskCount)
		} else {
			tasks, _ := models.GetDirectTasks(epicID)
			childCount = len(tasks)
			fmt.Printf("✓ Assigned epic %s to %s\n", epicID, groveID)
			fmt.Printf("  Epic: %s\n", epic.Title)
			fmt.Printf("  Tasks: %d\n", childCount)
		}

		// Write assignment file
		err = writeEpicAssignment(grove.Path, epic)
		if err != nil {
			return fmt.Errorf("failed to write assignment file: %w", err)
		}

		fmt.Printf("  Assignment written to: %s/.orc/assigned-work.json\n", grove.Path)
		fmt.Printf("  Epic status: ready → implement\n")
		fmt.Println()
		fmt.Printf("💡 IMP can check assignment:\n")
		fmt.Printf("   cd %s\n", grove.Path)
		fmt.Printf("   orc epic check-assignment\n")
		return nil
	},
}

var epicCheckAssignmentCmd = &cobra.Command{
	Use:   "check-assignment",
	Short: "Check if this grove has an assigned epic",
	Long:  "Run from within a grove directory to see the assigned epic and tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Try to read epic assignment
		epicAssignment, err := models.ReadEpicAssignment(cwd)
		if err != nil {
			fmt.Println("✓ No work assignment found")
			fmt.Println()
			fmt.Println("💡 This grove has no assigned epic yet")
			return nil
		}

		// Display epic assignment
		fmt.Printf("🌳 Epic Assignment: %s\n", epicAssignment.EpicID)
		fmt.Printf("   Title: %s\n", epicAssignment.EpicTitle)
		if epicAssignment.EpicDescription != "" {
			fmt.Printf("   Description: %s\n", epicAssignment.EpicDescription)
		}
		fmt.Printf("   Mission: %s\n", epicAssignment.MissionID)
		fmt.Printf("   Assigned: %s\n", epicAssignment.AssignedAt[:10])
		fmt.Printf("   By: %s\n", epicAssignment.AssignedBy)
		fmt.Printf("   Status: %s\n", epicAssignment.Status)
		fmt.Println()

		// Display progress and children
		p := epicAssignment.Progress
		if epicAssignment.Structure == "rabbit_holes" {
			// Epic with rabbit holes
			if p.TotalRabbitHoles > 0 {
				percentage := 0.0
				if p.TotalTasks > 0 {
					percentage = (float64(p.CompletedTasks) / float64(p.TotalTasks)) * 100
				}
				fmt.Printf("📊 Progress: %d/%d tasks complete (%.0f%%)\n\n", p.CompletedTasks, p.TotalTasks, percentage)

				fmt.Println("Rabbit Holes & Tasks:")
				for _, rh := range epicAssignment.RabbitHoles {
					rhIcon := getStatusIcon(rh.Status)
					fmt.Printf("  %s %s: %s [%s]\n", rhIcon, rh.RabbitHoleID, rh.Title, rh.Status)
					for _, task := range rh.Tasks {
						taskIcon := getStatusIcon(task.Status)
						fmt.Printf("    %s %s: %s [%s]\n", taskIcon, task.TaskID, task.Title, task.Status)
					}
				}
			}
		} else {
			// Epic with direct tasks
			if p.TotalTasks > 0 {
				percentage := (float64(p.CompletedTasks) / float64(p.TotalTasks)) * 100
				fmt.Printf("📊 Progress: %d/%d tasks complete (%.0f%%)\n\n", p.CompletedTasks, p.TotalTasks, percentage)

				fmt.Println("Tasks:")
				for _, task := range epicAssignment.Tasks {
					taskIcon := getStatusIcon(task.Status)
					fmt.Printf("  %s %s: %s [%s]\n", taskIcon, task.TaskID, task.Title, task.Status)
				}
			}
		}

		// Show next steps
		fmt.Println()
		fmt.Println("💡 To work on next task:")
		fmt.Println("   orc task list --status ready  # Find ready tasks")
		fmt.Println("   orc task claim TASK-XXX       # Claim a task")
		fmt.Println("   # Do the work...")
		fmt.Println("   orc task complete TASK-XXX    # Mark complete")

		return nil
	},
}

var epicUpdateAssignmentCmd = &cobra.Command{
	Use:   "update-assignment",
	Short: "Update the status of the current grove's epic assignment",
	RunE: func(cmd *cobra.Command, args []string) error {
		status, _ := cmd.Flags().GetString("status")

		if status == "" {
			return fmt.Errorf("--status flag is required")
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Read epic assignment
		epicAssignment, err := models.ReadEpicAssignment(cwd)
		if err != nil {
			return fmt.Errorf("no epic assignment found in current directory: %w", err)
		}

		// Update epic assignment file
		err = models.UpdateEpicAssignmentStatus(cwd, status)
		if err != nil {
			return fmt.Errorf("failed to update assignment status: %w", err)
		}

		fmt.Printf("✓ Epic assignment status updated: %s → %s\n", epicAssignment.Status, status)
		fmt.Printf("  Epic: %s\n", epicAssignment.EpicID)
		return nil
	},
}

// Helper function to write epic assignment file
func writeEpicAssignment(groveDir string, epic *models.Epic) error {
	// Determine structure (rabbit holes or direct tasks)
	hasRH, _ := models.HasRabbitHoles(epic.ID)

	if hasRH {
		// Get rabbit holes and their tasks
		rhs, err := models.ListRabbitHoles(epic.ID, "")
		if err != nil {
			return err
		}

		var rabbitHoles []*models.RabbitHole
		for _, rh := range rhs {
			rabbitHoles = append(rabbitHoles, rh)
		}

		return models.WriteEpicAssignmentWithRabbitHoles(groveDir, epic, rabbitHoles, "ORC")
	} else {
		// Get direct tasks
		tasks, err := models.GetDirectTasks(epic.ID)
		if err != nil {
			return err
		}

		return models.WriteEpicAssignmentWithTasks(groveDir, epic, tasks, "ORC")
	}
}

var epicFocusCmd = &cobra.Command{
	Use:   "focus [epic-id]",
	Short: "Set the current epic focus in config",
	Long: `Set the currently focused epic in the .orc/config.json file.

This tells orc prime which epic you're working on, so Claude knows the context.
If no epic is specified, clears the current focus.

Examples:
  orc epic focus EPIC-123    # Set focus to EPIC-123
  orc epic focus             # Clear current focus`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get current directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Load existing config
		cfg, err := config.LoadConfig(cwd)
		if err != nil {
			return fmt.Errorf("no .orc/config.json found in %s: %w", cwd, err)
		}

		var epicID string
		if len(args) > 0 {
			epicID = args[0]

			// Validate epic exists
			epic, err := models.GetEpic(epicID)
			if err != nil {
				return fmt.Errorf("epic %s not found: %w", epicID, err)
			}

			fmt.Printf("✓ Setting focus to: %s - %s\n", epic.ID, epic.Title)
		} else {
			fmt.Println("✓ Clearing epic focus")
		}

		// Update config based on type
		switch cfg.Type {
		case config.TypeGrove:
			cfg.Grove.CurrentEpic = epicID
		case config.TypeMission:
			cfg.Mission.CurrentEpic = epicID
		case config.TypeGlobal:
			cfg.State.CurrentEpic = epicID
		default:
			return fmt.Errorf("unknown config type: %s", cfg.Type)
		}

		// Save config
		if err := config.SaveConfig(cwd, cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		if epicID != "" {
			fmt.Println()
			fmt.Println("💡 Run 'orc prime' to see updated context")
		}

		return nil
	},
}

// Helper function for status icons
func getStatusIcon(status string) string {
	switch status {
	case "complete":
		return "✅"
	case "implement", "in_progress":
		return "🔄"
	case "ready":
		return "📦"
	case "blocked":
		return "🚫"
	case "paused":
		return "⏸️"
	default:
		return "•"
	}
}

// getCurrentEpic returns the current_epic from config.json if set
func getCurrentEpic() string {
	// Try mission context first
	missionCtx, _ := context.DetectMissionContext()
	if missionCtx != nil {
		cfg, err := config.LoadConfig(missionCtx.WorkspacePath)
		if err == nil {
			switch cfg.Type {
			case config.TypeGrove:
				if cfg.Grove != nil && cfg.Grove.CurrentEpic != "" {
					return cfg.Grove.CurrentEpic
				}
			case config.TypeMission:
				if cfg.Mission != nil && cfg.Mission.CurrentEpic != "" {
					return cfg.Mission.CurrentEpic
				}
			}
		}

		// Also try current directory for grove config
		cwd, _ := os.Getwd()
		cfg, err = config.LoadConfig(cwd)
		if err == nil && cfg.Type == config.TypeGrove && cfg.Grove != nil && cfg.Grove.CurrentEpic != "" {
			return cfg.Grove.CurrentEpic
		}
	}

	// Try global config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	cfg, err := config.LoadConfig(homeDir)
	if err != nil {
		return ""
	}
	if cfg.State != nil && cfg.State.CurrentEpic != "" {
		return cfg.State.CurrentEpic
	}

	return ""
}

func init() {
	// epic create flags
	epicCreateCmd.Flags().StringP("mission", "m", "", "Mission ID (defaults to context or MISSION-001)")
	epicCreateCmd.Flags().StringP("description", "d", "", "Epic description")
	epicCreateCmd.Flags().String("context-ref", "", "External context reference (optional)")

	// epic list flags
	epicListCmd.Flags().StringP("mission", "m", "", "Filter by mission")
	epicListCmd.Flags().StringP("status", "s", "", "Filter by status")

	// epic update flags
	epicUpdateCmd.Flags().String("title", "", "New title")
	epicUpdateCmd.Flags().StringP("description", "d", "", "New description")

	// epic assign flags
	epicAssignCmd.Flags().String("grove", "", "Grove ID (required)")
	epicAssignCmd.MarkFlagRequired("grove")

	// epic update-assignment flags
	epicUpdateAssignmentCmd.Flags().String("status", "", "New status (assigned, in_progress, complete)")
	epicUpdateAssignmentCmd.MarkFlagRequired("status")

	// Register subcommands
	epicCmd.AddCommand(epicCreateCmd)
	epicCmd.AddCommand(epicListCmd)
	epicCmd.AddCommand(epicShowCmd)
	epicCmd.AddCommand(epicCompleteCmd)
	epicCmd.AddCommand(epicUpdateCmd)
	epicCmd.AddCommand(epicPinCmd)
	epicCmd.AddCommand(epicUnpinCmd)
	epicCmd.AddCommand(epicAssignCmd)
	epicCmd.AddCommand(epicCheckAssignmentCmd)
	epicCmd.AddCommand(epicUpdateAssignmentCmd)
	epicCmd.AddCommand(epicFocusCmd)
}

// EpicCmd returns the epic command
func EpicCmd() *cobra.Command {
	return epicCmd
}
