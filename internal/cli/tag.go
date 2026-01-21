package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Manage tags (classification labels for tasks)",
	Long:  "Create, list, show, and delete tags in the ORC ledger",
}

var tagCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new tag",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		name := args[0]
		description, _ := cmd.Flags().GetString("description")

		resp, err := wire.TagService().CreateTag(ctx, primary.CreateTagRequest{
			Name:        name,
			Description: description,
		})
		if err != nil {
			return fmt.Errorf("failed to create tag: %w", err)
		}

		tag := resp.Tag
		fmt.Printf("✓ Created tag %s: %s\n", tag.ID, tag.Name)
		if tag.Description != "" {
			fmt.Printf("  Description: %s\n", tag.Description)
		}
		return nil
	},
}

var tagListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		tags, err := wire.TagService().ListTags(ctx)
		if err != nil {
			return fmt.Errorf("failed to list tags: %w", err)
		}

		if len(tags) == 0 {
			fmt.Println("No tags found")
			return nil
		}

		fmt.Printf("Found %d tag(s):\n\n", len(tags))
		for _, tag := range tags {
			fmt.Printf("%-10s %s", tag.ID, tag.Name)
			if tag.Description != "" {
				fmt.Printf(" - %s", tag.Description)
			}
			fmt.Println()
		}
		return nil
	},
}

var tagShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show tag details and associated tasks",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		name := args[0]

		tag, err := wire.TagService().GetTagByName(ctx, name)
		if err != nil {
			return fmt.Errorf("tag not found: %w", err)
		}

		// Display tag details
		fmt.Printf("Tag: %s (%s)\n", tag.Name, tag.ID)
		if tag.Description != "" {
			fmt.Printf("Description: %s\n", tag.Description)
		}
		fmt.Printf("Created: %s\n", tag.CreatedAt)
		fmt.Println()

		// Display tasks with this tag
		tasks, err := wire.TaskService().ListTasksByTag(ctx, name)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}

		if len(tasks) == 0 {
			fmt.Println("No tasks tagged with this tag")
		} else {
			fmt.Printf("Tasks (%d):\n", len(tasks))
			for _, task := range tasks {
				statusIcon := getStatusIcon(task.Status)
				fmt.Printf("  %s %s: %s [%s]\n", statusIcon, task.ID, task.Title, task.Status)
			}
		}

		return nil
	},
}

var tagDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a tag (removes from all tasks)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		name := args[0]

		// Get tag by name
		tag, err := wire.TagService().GetTagByName(ctx, name)
		if err != nil {
			return fmt.Errorf("tag not found: %w", err)
		}

		// Delete the tag (cascade removes entity_tags)
		err = wire.TagService().DeleteTag(ctx, tag.ID)
		if err != nil {
			return fmt.Errorf("failed to delete tag: %w", err)
		}

		fmt.Printf("✓ Deleted tag: %s\n", name)
		return nil
	},
}

func init() {
	// tag create flags
	tagCreateCmd.Flags().StringP("description", "d", "", "Tag description")

	// Register subcommands
	tagCmd.AddCommand(tagCreateCmd)
	tagCmd.AddCommand(tagListCmd)
	tagCmd.AddCommand(tagShowCmd)
	tagCmd.AddCommand(tagDeleteCmd)
}

// TagCmd returns the tag command
func TagCmd() *cobra.Command {
	return tagCmd
}
