package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	orccontext "github.com/example/orc/internal/context"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
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
		ctx := context.Background()
		title := args[0]
		shipmentID, _ := cmd.Flags().GetString("shipment")
		investigationID, _ := cmd.Flags().GetString("investigation")
		commissionID, _ := cmd.Flags().GetString("commission")
		description, _ := cmd.Flags().GetString("description")
		taskType, _ := cmd.Flags().GetString("type")

		// Get commission from context or require explicit flag
		if commissionID == "" {
			commissionID = orccontext.GetContextCommissionID()
			if commissionID == "" {
				return fmt.Errorf("no commission context detected\nHint: Use --commission flag or run from a workbench directory")
			}
		}

		resp, err := wire.TaskService().CreateTask(ctx, primary.CreateTaskRequest{
			ShipmentID:      shipmentID,
			InvestigationID: investigationID,
			CommissionID:    commissionID,
			Title:           title,
			Description:     description,
			Type:            taskType,
		})
		if err != nil {
			return fmt.Errorf("failed to create task: %w", err)
		}

		task := resp.Task
		fmt.Printf("✓ Created task %s: %s\n", task.ID, task.Title)
		if task.ShipmentID != "" {
			fmt.Printf("  Under shipment: %s\n", task.ShipmentID)
		}
		if task.InvestigationID != "" {
			fmt.Printf("  Under investigation: %s\n", task.InvestigationID)
		}
		fmt.Printf("  Commission: %s\n", task.CommissionID)
		return nil
	},
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		shipmentID, _ := cmd.Flags().GetString("shipment")
		investigationID, _ := cmd.Flags().GetString("investigation")
		status, _ := cmd.Flags().GetString("status")
		tag, _ := cmd.Flags().GetString("tag")

		var tasks []*primary.Task
		var err error

		if tag != "" {
			// Filter by tag
			tasks, err = wire.TaskService().ListTasksByTag(ctx, tag)
			if err != nil {
				return fmt.Errorf("failed to list tasks: %w", err)
			}

			// Apply additional filters if specified
			var filteredTasks []*primary.Task
			for _, task := range tasks {
				if shipmentID != "" && task.ShipmentID != shipmentID {
					continue
				}
				if investigationID != "" && task.InvestigationID != investigationID {
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
			tasks, err = wire.TaskService().ListTasks(ctx, primary.TaskFilters{
				ShipmentID:      shipmentID,
				InvestigationID: investigationID,
				Status:          status,
			})
			if err != nil {
				return fmt.Errorf("failed to list tasks: %w", err)
			}
		}

		if len(tasks) == 0 {
			fmt.Println("No tasks found.")
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
			if task.Type != "" {
				typeStr = fmt.Sprintf(" [%s]", task.Type)
			}

			fmt.Printf("%s %s: %s%s [%s]%s\n", statusIcon, task.ID, task.Title, typeStr, task.Status, pinnedIcon)
			if task.ShipmentID != "" {
				fmt.Printf("   Shipment: %s\n", task.ShipmentID)
			}
			if task.InvestigationID != "" {
				fmt.Printf("   Investigation: %s\n", task.InvestigationID)
			}
			if task.AssignedWorkbenchID != "" {
				fmt.Printf("   Workbench: %s\n", task.AssignedWorkbenchID)
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
		ctx := context.Background()
		taskID := args[0]

		task, err := wire.TaskService().GetTask(ctx, taskID)
		if err != nil {
			return fmt.Errorf("task not found: %w", err)
		}

		// Display task details
		fmt.Printf("Task: %s\n", task.ID)
		fmt.Printf("Title: %s\n", task.Title)
		if task.Description != "" {
			fmt.Printf("Description: %s\n", task.Description)
		}
		fmt.Printf("Status: %s\n", task.Status)
		if task.Type != "" {
			fmt.Printf("Type: %s\n", task.Type)
		}
		fmt.Printf("Commission: %s\n", task.CommissionID)
		if task.ShipmentID != "" {
			fmt.Printf("Shipment: %s\n", task.ShipmentID)
		}
		if task.InvestigationID != "" {
			fmt.Printf("Investigation: %s\n", task.InvestigationID)
		}
		if task.TomeID != "" {
			fmt.Printf("Tome: %s\n", task.TomeID)
		}
		if task.ConclaveID != "" {
			fmt.Printf("Conclave: %s\n", task.ConclaveID)
		}
		if task.AssignedWorkbenchID != "" {
			fmt.Printf("Assigned Workbench: %s\n", task.AssignedWorkbenchID)
		}
		if task.Priority != "" {
			fmt.Printf("Priority: %s\n", task.Priority)
		}
		if task.Pinned {
			fmt.Printf("Pinned: yes\n")
		}
		fmt.Printf("Created: %s\n", task.CreatedAt)
		if task.ClaimedAt != "" {
			fmt.Printf("Claimed: %s\n", task.ClaimedAt)
		}
		if task.CompletedAt != "" {
			fmt.Printf("Completed: %s\n", task.CompletedAt)
		}
		if task.Tag != nil {
			fmt.Printf("Tag: %s\n", task.Tag.Name)
		}

		return nil
	},
}

var taskClaimCmd = &cobra.Command{
	Use:   "claim [task-id]",
	Short: "Claim a task (mark as implement)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		taskID := args[0]

		// Try to get workbench from current directory
		cwd, _ := os.Getwd()
		workbench, _ := wire.WorkbenchService().GetWorkbenchByPath(ctx, cwd)

		workbenchID := ""
		if workbench != nil {
			workbenchID = workbench.ID
		}

		err := wire.TaskService().ClaimTask(ctx, primary.ClaimTaskRequest{
			TaskID:      taskID,
			WorkbenchID: workbenchID,
		})
		if err != nil {
			return fmt.Errorf("failed to claim task: %w", err)
		}

		fmt.Printf("✓ Task %s claimed\n", taskID)
		if workbenchID != "" {
			fmt.Printf("  Assigned to workbench: %s\n", workbenchID)
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
		ctx := context.Background()
		taskID := args[0]

		err := wire.TaskService().CompleteTask(ctx, taskID)
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

var taskPauseCmd = &cobra.Command{
	Use:   "pause [task-id]",
	Short: "Pause an in-progress task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		taskID := args[0]

		err := wire.TaskService().PauseTask(ctx, taskID)
		if err != nil {
			return fmt.Errorf("failed to pause task: %w", err)
		}

		fmt.Printf("✓ Task %s paused\n", taskID)
		return nil
	},
}

var taskResumeCmd = &cobra.Command{
	Use:   "resume [task-id]",
	Short: "Resume a paused task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		taskID := args[0]

		err := wire.TaskService().ResumeTask(ctx, taskID)
		if err != nil {
			return fmt.Errorf("failed to resume task: %w", err)
		}

		fmt.Printf("✓ Task %s resumed\n", taskID)
		return nil
	},
}

var taskUpdateCmd = &cobra.Command{
	Use:   "update [task-id]",
	Short: "Update task title and/or description",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		taskID := args[0]
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")

		if title == "" && description == "" {
			return fmt.Errorf("must specify --title and/or --description")
		}

		err := wire.TaskService().UpdateTask(ctx, primary.UpdateTaskRequest{
			TaskID:      taskID,
			Title:       title,
			Description: description,
		})
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
		ctx := context.Background()
		taskID := args[0]

		err := wire.TaskService().PinTask(ctx, taskID)
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
		ctx := context.Background()
		taskID := args[0]

		err := wire.TaskService().UnpinTask(ctx, taskID)
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
	Long:  "Discover ready tasks assigned to the current workbench",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		autoClaim, _ := cmd.Flags().GetBool("auto-claim")

		// Get current workbench
		cwd, _ := os.Getwd()
		workbench, err := wire.WorkbenchService().GetWorkbenchByPath(ctx, cwd)
		if err != nil {
			return fmt.Errorf("not in a workbench directory: %w", err)
		}

		// Get tasks assigned to this workbench with ready status
		readyTasks, err := wire.TaskService().DiscoverTasks(ctx, workbench.ID)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
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
			if task.Description != "" {
				fmt.Printf("   %s\n", task.Description)
			}
		}

		if autoClaim && len(readyTasks) > 0 {
			// Claim first ready task
			task := readyTasks[0]
			err := wire.TaskService().ClaimTask(ctx, primary.ClaimTaskRequest{
				TaskID:      task.ID,
				WorkbenchID: workbench.ID,
			})
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
		ctx := context.Background()
		taskID := args[0]
		tagName := args[1]

		err := wire.TaskService().TagTask(ctx, taskID, tagName)
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
		ctx := context.Background()
		taskID := args[0]

		err := wire.TaskService().UntagTask(ctx, taskID)
		if err != nil {
			return fmt.Errorf("failed to untag task: %w", err)
		}

		fmt.Printf("✓ Task %s untagged\n", taskID)
		return nil
	},
}

var taskMoveCmd = &cobra.Command{
	Use:   "move [task-id]",
	Short: "Move a task to a different container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		taskID := args[0]
		toShipment, _ := cmd.Flags().GetString("to-shipment")
		toTome, _ := cmd.Flags().GetString("to-tome")
		toConclave, _ := cmd.Flags().GetString("to-conclave")

		// Validate exactly one target specified
		targetCount := 0
		if toShipment != "" {
			targetCount++
		}
		if toTome != "" {
			targetCount++
		}
		if toConclave != "" {
			targetCount++
		}

		if targetCount == 0 {
			return fmt.Errorf("must specify exactly one target: --to-shipment, --to-tome, or --to-conclave")
		}
		if targetCount > 1 {
			return fmt.Errorf("cannot specify multiple targets")
		}

		err := wire.TaskService().MoveTask(ctx, primary.MoveTaskRequest{
			TaskID:       taskID,
			ToShipmentID: toShipment,
			ToTomeID:     toTome,
			ToConclaveID: toConclave,
		})
		if err != nil {
			return fmt.Errorf("failed to move task: %w", err)
		}

		target := ""
		if toShipment != "" {
			target = toShipment
		} else if toTome != "" {
			target = toTome
		} else {
			target = toConclave
		}

		fmt.Printf("✓ Task %s moved to %s\n", taskID, target)
		return nil
	},
}

func init() {
	// task create flags
	taskCreateCmd.Flags().String("shipment", "", "Shipment ID")
	taskCreateCmd.Flags().String("investigation", "", "Investigation ID")
	taskCreateCmd.Flags().StringP("commission", "c", "", "Commission ID (defaults to context)")
	taskCreateCmd.Flags().StringP("description", "d", "", "Task description")
	taskCreateCmd.Flags().String("type", "", "Task type (research, implementation, fix, documentation, maintenance)")

	// task list flags
	taskListCmd.Flags().String("shipment", "", "Filter by shipment")
	taskListCmd.Flags().String("investigation", "", "Filter by investigation")
	taskListCmd.Flags().StringP("status", "s", "", "Filter by status")
	taskListCmd.Flags().String("tag", "", "Filter by tag")

	// task update flags
	taskUpdateCmd.Flags().String("title", "", "New title")
	taskUpdateCmd.Flags().StringP("description", "d", "", "New description")

	// task discover flags
	taskDiscoverCmd.Flags().Bool("auto-claim", false, "Automatically claim the first ready task")

	// task move flags
	taskMoveCmd.Flags().String("to-shipment", "", "Move to shipment")
	taskMoveCmd.Flags().String("to-tome", "", "Move to tome")
	taskMoveCmd.Flags().String("to-conclave", "", "Move to conclave")

	// Register subcommands
	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskShowCmd)
	taskCmd.AddCommand(taskClaimCmd)
	taskCmd.AddCommand(taskCompleteCmd)
	taskCmd.AddCommand(taskPauseCmd)
	taskCmd.AddCommand(taskResumeCmd)
	taskCmd.AddCommand(taskUpdateCmd)
	taskCmd.AddCommand(taskPinCmd)
	taskCmd.AddCommand(taskUnpinCmd)
	taskCmd.AddCommand(taskDiscoverCmd)
	taskCmd.AddCommand(taskTagCmd)
	taskCmd.AddCommand(taskUntagCmd)
	taskCmd.AddCommand(taskMoveCmd)
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
