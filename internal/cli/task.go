package cli

import (
	"fmt"
	"os"

	"github.com/example/orc/internal/context"
	"github.com/example/orc/internal/models"
	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks (atomic units of work)",
	Long:  "Create, list, claim, complete, and manage tasks in the ORC ledger",
}

var taskCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new task",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		shipmentID, _ := cmd.Flags().GetString("shipment")
		missionID, _ := cmd.Flags().GetString("mission")
		description, _ := cmd.Flags().GetString("description")
		taskType, _ := cmd.Flags().GetString("type")

		// Get mission from context or require explicit flag
		if missionID == "" {
			missionID = context.GetContextMissionID()
			if missionID == "" {
				return fmt.Errorf("no mission context detected\nHint: Use --mission flag or run from a grove/mission directory")
			}
		}

		task, err := models.CreateTask(shipmentID, missionID, title, description, taskType)
		if err != nil {
			return fmt.Errorf("failed to create task: %w", err)
		}

		fmt.Printf("✓ Created task %s: %s\n", task.ID, task.Title)
		if task.ShipmentID.Valid {
			fmt.Printf("  Under shipment: %s\n", task.ShipmentID.String)
		}
		fmt.Printf("  Mission: %s\n", task.MissionID)
		return nil
	},
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		shipmentID, _ := cmd.Flags().GetString("shipment")
		status, _ := cmd.Flags().GetString("status")
		tag, _ := cmd.Flags().GetString("tag")

		var tasks []*models.Task
		var err error

		if tag != "" {
			// Filter by tag
			tasks, err = models.ListTasksByTag(tag)
			if err != nil {
				return fmt.Errorf("failed to list tasks: %w", err)
			}

			// Apply additional filters if specified
			var filteredTasks []*models.Task
			for _, task := range tasks {
				if shipmentID != "" && (!task.ShipmentID.Valid || task.ShipmentID.String != shipmentID) {
					continue
				}
				if status != "" && task.Status != status {
					continue
				}
				filteredTasks = append(filteredTasks, task)
			}
			tasks = filteredTasks
		} else {
			// Use normal list
			tasks, err = models.ListTasks(shipmentID, status)
			if err != nil {
				return fmt.Errorf("failed to list tasks: %w", err)
			}
		}

		if len(tasks) == 0 {
			fmt.Println("No tasks found")
			return nil
		}

		fmt.Printf("Found %d task(s):\n\n", len(tasks))
		for _, task := range tasks {
			statusIcon := getStatusIcon(task.Status)
			pinnedIcon := ""
			if task.Pinned {
				pinnedIcon = " 📌"
			}

			typeStr := ""
			if task.Type.Valid {
				typeStr = fmt.Sprintf(" [%s]", task.Type.String)
			}

			fmt.Printf("%s %s: %s%s [%s]%s\n", statusIcon, task.ID, task.Title, typeStr, task.Status, pinnedIcon)
			if task.ShipmentID.Valid {
				fmt.Printf("   Shipment: %s\n", task.ShipmentID.String)
			}
			if task.AssignedGroveID.Valid {
				fmt.Printf("   Grove: %s\n", task.AssignedGroveID.String)
			}
			fmt.Println()
		}
		return nil
	},
}

var taskShowCmd = &cobra.Command{
	Use:   "show [task-id]",
	Short: "Show task details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		task, err := models.GetTask(taskID)
		if err != nil {
			return fmt.Errorf("task not found: %w", err)
		}

		// Display task details
		fmt.Printf("Task: %s\n", task.ID)
		fmt.Printf("Title: %s\n", task.Title)
		if task.Description.Valid {
			fmt.Printf("Description: %s\n", task.Description.String)
		}
		fmt.Printf("Status: %s\n", task.Status)
		if task.Type.Valid {
			fmt.Printf("Type: %s\n", task.Type.String)
		}
		fmt.Printf("Mission: %s\n", task.MissionID)
		if task.ShipmentID.Valid {
			fmt.Printf("Shipment: %s\n", task.ShipmentID.String)
		}
		if task.AssignedGroveID.Valid {
			fmt.Printf("Assigned Grove: %s\n", task.AssignedGroveID.String)
		}
		if task.Priority.Valid {
			fmt.Printf("Priority: %s\n", task.Priority.String)
		}
		if task.Pinned {
			fmt.Printf("Pinned: yes\n")
		}
		if task.ConclaveID.Valid {
			fmt.Printf("Conclave: %s\n", task.ConclaveID.String)
		}
		if task.PromotedFromID.Valid {
			fmt.Printf("Promoted from: %s (%s)\n", task.PromotedFromID.String, task.PromotedFromType.String)
		}
		fmt.Printf("Created: %s\n", task.CreatedAt.Format("2006-01-02 15:04"))
		if task.ClaimedAt.Valid {
			fmt.Printf("Claimed: %s\n", task.ClaimedAt.Time.Format("2006-01-02 15:04"))
		}
		if task.CompletedAt.Valid {
			fmt.Printf("Completed: %s\n", task.CompletedAt.Time.Format("2006-01-02 15:04"))
		}

		return nil
	},
}

var taskClaimCmd = &cobra.Command{
	Use:   "claim [task-id]",
	Short: "Claim a task (mark as implement)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		// Try to get grove from current directory
		cwd, _ := os.Getwd()
		grove, _ := models.GetGroveByPath(cwd)

		groveID := ""
		if grove != nil {
			groveID = grove.ID
		}

		err := models.ClaimTask(taskID, groveID)
		if err != nil {
			return fmt.Errorf("failed to claim task: %w", err)
		}

		fmt.Printf("✓ Task %s claimed\n", taskID)
		if groveID != "" {
			fmt.Printf("  Assigned to grove: %s\n", groveID)
		}
		fmt.Println()
		fmt.Println("💡 Next steps:")
		fmt.Println("   # Do the work...")
		fmt.Printf("   orc task complete %s\n", taskID)
		return nil
	},
}

var taskCompleteCmd = &cobra.Command{
	Use:   "complete [task-id]",
	Short: "Mark task as complete",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		err := models.CompleteTask(taskID)
		if err != nil {
			return fmt.Errorf("failed to complete task: %w", err)
		}

		fmt.Printf("✓ Task %s marked as complete\n", taskID)
		fmt.Println()
		fmt.Println("💡 Check for next task:")
		fmt.Println("   orc shipment check-assignment  # See progress")
		fmt.Println("   orc task list --status ready  # Find next task")
		return nil
	},
}

var taskUpdateCmd = &cobra.Command{
	Use:   "update [task-id]",
	Short: "Update task title and/or description",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")

		if title == "" && description == "" {
			return fmt.Errorf("must specify --title and/or --description")
		}

		err := models.UpdateTask(taskID, title, description)
		if err != nil {
			return fmt.Errorf("failed to update task: %w", err)
		}

		fmt.Printf("✓ Task %s updated\n", taskID)
		return nil
	},
}

var taskPinCmd = &cobra.Command{
	Use:   "pin [task-id]",
	Short: "Pin task to keep it visible",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		err := models.PinTask(taskID)
		if err != nil {
			return fmt.Errorf("failed to pin task: %w", err)
		}

		fmt.Printf("✓ Task %s pinned 📌\n", taskID)
		return nil
	},
}

var taskUnpinCmd = &cobra.Command{
	Use:   "unpin [task-id]",
	Short: "Unpin task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		err := models.UnpinTask(taskID)
		if err != nil {
			return fmt.Errorf("failed to unpin task: %w", err)
		}

		fmt.Printf("✓ Task %s unpinned\n", taskID)
		return nil
	},
}

var taskDiscoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Find and optionally claim ready tasks",
	Long:  "Discover ready tasks assigned to the current grove",
	RunE: func(cmd *cobra.Command, args []string) error {
		autoClaim, _ := cmd.Flags().GetBool("auto-claim")

		// Get current grove
		cwd, _ := os.Getwd()
		grove, err := models.GetGroveByPath(cwd)
		if err != nil {
			return fmt.Errorf("not in a grove directory: %w", err)
		}

		// Get tasks assigned to this grove with ready status
		tasks, err := models.GetTasksByGrove(grove.ID)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}

		// Filter to ready tasks
		var readyTasks []*models.Task
		for _, task := range tasks {
			if task.Status == "ready" {
				readyTasks = append(readyTasks, task)
			}
		}

		if len(readyTasks) == 0 {
			fmt.Println("✓ No ready tasks found")
			fmt.Println()
			fmt.Println("💡 Check assignment status:")
			fmt.Println("   orc shipment check-assignment")
			return nil
		}

		fmt.Printf("Found %d ready task(s):\n\n", len(readyTasks))
		for _, task := range readyTasks {
			fmt.Printf("📦 %s: %s\n", task.ID, task.Title)
			if task.Description.Valid {
				fmt.Printf("   %s\n", task.Description.String)
			}
		}

		if autoClaim && len(readyTasks) > 0 {
			// Claim first ready task
			task := readyTasks[0]
			err := models.ClaimTask(task.ID, grove.ID)
			if err != nil {
				return fmt.Errorf("failed to claim task: %w", err)
			}

			fmt.Println()
			fmt.Printf("✓ Auto-claimed: %s\n", task.ID)
			fmt.Println()
			fmt.Println("💡 Get started:")
			fmt.Printf("   orc task show %s  # See details\n", task.ID)
			fmt.Println("   # Do the work...")
			fmt.Printf("   orc task complete %s\n", task.ID)
		} else {
			fmt.Println()
			fmt.Println("💡 To claim a task:")
			fmt.Printf("   orc task claim %s\n", readyTasks[0].ID)
		}

		return nil
	},
}

var taskTagCmd = &cobra.Command{
	Use:   "tag [task-id] [tag-name]",
	Short: "Add a tag to a task",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]
		tagName := args[1]

		err := models.AddTagToTask(taskID, tagName)
		if err != nil {
			return fmt.Errorf("failed to tag task: %w", err)
		}

		fmt.Printf("✓ Task %s tagged with '%s'\n", taskID, tagName)
		return nil
	},
}

var taskUntagCmd = &cobra.Command{
	Use:   "untag [task-id]",
	Short: "Remove tag from a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		err := models.RemoveTagFromTask(taskID)
		if err != nil {
			return fmt.Errorf("failed to untag task: %w", err)
		}

		fmt.Printf("✓ Task %s untagged\n", taskID)
		return nil
	},
}

func init() {
	// task create flags
	taskCreateCmd.Flags().String("shipment", "", "Shipment ID")
	taskCreateCmd.Flags().StringP("mission", "m", "", "Mission ID (defaults to context)")
	taskCreateCmd.Flags().StringP("description", "d", "", "Task description")
	taskCreateCmd.Flags().String("type", "", "Task type (research, implementation, fix, documentation, maintenance)")

	// task list flags
	taskListCmd.Flags().String("shipment", "", "Filter by shipment")
	taskListCmd.Flags().StringP("status", "s", "", "Filter by status")
	taskListCmd.Flags().String("tag", "", "Filter by tag")

	// task update flags
	taskUpdateCmd.Flags().String("title", "", "New title")
	taskUpdateCmd.Flags().StringP("description", "d", "", "New description")

	// task discover flags
	taskDiscoverCmd.Flags().Bool("auto-claim", false, "Automatically claim the first ready task")

	// Register subcommands
	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskShowCmd)
	taskCmd.AddCommand(taskClaimCmd)
	taskCmd.AddCommand(taskCompleteCmd)
	taskCmd.AddCommand(taskUpdateCmd)
	taskCmd.AddCommand(taskPinCmd)
	taskCmd.AddCommand(taskUnpinCmd)
	taskCmd.AddCommand(taskDiscoverCmd)
	taskCmd.AddCommand(taskTagCmd)
	taskCmd.AddCommand(taskUntagCmd)
}

// TaskCmd returns the task command
func TaskCmd() *cobra.Command {
	return taskCmd
}

// getStatusIcon returns an emoji icon for a task status
func getStatusIcon(status string) string {
	switch status {
	case "ready":
		return "📦"
	case "implement", "in_progress":
		return "🔧"
	case "complete":
		return "✅"
	case "blocked":
		return "🚫"
	case "paused":
		return "⏸️"
	default:
		return "📋"
	}
}
