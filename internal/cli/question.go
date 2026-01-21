package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	orcctx "github.com/example/orc/internal/context"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

var questionCmd = &cobra.Command{
	Use:   "question",
	Short: "Manage questions (open queries to be answered)",
	Long:  "Create, list, answer, and manage questions in the ORC ledger",
}

var questionCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new question",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		missionID, _ := cmd.Flags().GetString("mission")
		description, _ := cmd.Flags().GetString("description")
		investigationID, _ := cmd.Flags().GetString("investigation")

		// Get mission from context or require explicit flag
		if missionID == "" {
			missionID = orcctx.GetContextMissionID()
			if missionID == "" {
				return fmt.Errorf("no mission context detected\nHint: Use --mission flag or run from a grove/mission directory")
			}
		}

		ctx := context.Background()
		resp, err := wire.QuestionService().CreateQuestion(ctx, primary.CreateQuestionRequest{
			MissionID:       missionID,
			InvestigationID: investigationID,
			Title:           title,
			Description:     description,
		})
		if err != nil {
			return fmt.Errorf("failed to create question: %w", err)
		}

		question := resp.Question
		fmt.Printf("✓ Created question %s: %s\n", question.ID, question.Title)
		if question.InvestigationID != "" {
			fmt.Printf("  Investigation: %s\n", question.InvestigationID)
		}
		fmt.Printf("  Mission: %s\n", question.MissionID)
		fmt.Printf("  Status: %s\n", question.Status)
		return nil
	},
}

var questionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List questions",
	RunE: func(cmd *cobra.Command, args []string) error {
		missionID, _ := cmd.Flags().GetString("mission")
		investigationID, _ := cmd.Flags().GetString("investigation")
		status, _ := cmd.Flags().GetString("status")

		// Get mission from context if not specified
		if missionID == "" {
			missionID = orcctx.GetContextMissionID()
		}

		ctx := context.Background()
		questions, err := wire.QuestionService().ListQuestions(ctx, primary.QuestionFilters{
			MissionID:       missionID,
			InvestigationID: investigationID,
			Status:          status,
		})
		if err != nil {
			return fmt.Errorf("failed to list questions: %w", err)
		}

		if len(questions) == 0 {
			fmt.Println("No questions found")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tINVESTIGATION")
		fmt.Fprintln(w, "--\t-----\t------\t-------------")
		for _, q := range questions {
			pinnedMark := ""
			if q.Pinned {
				pinnedMark = " [pinned]"
			}
			statusIcon := "❓"
			if q.Status == "answered" {
				statusIcon = "✅"
			}
			inv := "-"
			if q.InvestigationID != "" {
				inv = q.InvestigationID
			}
			fmt.Fprintf(w, "%s\t%s%s\t%s %s\t%s\n", q.ID, q.Title, pinnedMark, statusIcon, q.Status, inv)
		}
		w.Flush()
		return nil
	},
}

var questionShowCmd = &cobra.Command{
	Use:   "show [question-id]",
	Short: "Show question details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		questionID := args[0]

		ctx := context.Background()
		question, err := wire.QuestionService().GetQuestion(ctx, questionID)
		if err != nil {
			return fmt.Errorf("question not found: %w", err)
		}

		fmt.Printf("Question: %s\n", question.ID)
		fmt.Printf("Title: %s\n", question.Title)
		if question.Description != "" {
			fmt.Printf("Description: %s\n", question.Description)
		}
		fmt.Printf("Status: %s\n", question.Status)
		if question.Answer != "" {
			fmt.Printf("Answer: %s\n", question.Answer)
		}
		fmt.Printf("Mission: %s\n", question.MissionID)
		if question.InvestigationID != "" {
			fmt.Printf("Investigation: %s\n", question.InvestigationID)
		}
		if question.ConclaveID != "" {
			fmt.Printf("Conclave: %s\n", question.ConclaveID)
		}
		if question.Pinned {
			fmt.Printf("Pinned: yes\n")
		}
		if question.PromotedFromID != "" {
			fmt.Printf("Promoted from: %s (%s)\n", question.PromotedFromID, question.PromotedFromType)
		}
		fmt.Printf("Created: %s\n", question.CreatedAt)
		if question.AnsweredAt != "" {
			fmt.Printf("Answered: %s\n", question.AnsweredAt)
		}

		return nil
	},
}

var questionAnswerCmd = &cobra.Command{
	Use:   "answer [question-id]",
	Short: "Provide an answer to a question",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		questionID := args[0]
		answer, _ := cmd.Flags().GetString("answer")

		if answer == "" {
			return fmt.Errorf("must specify --answer")
		}

		ctx := context.Background()
		err := wire.QuestionService().AnswerQuestion(ctx, questionID, answer)
		if err != nil {
			return fmt.Errorf("failed to answer question: %w", err)
		}

		fmt.Printf("✓ Question %s answered\n", questionID)
		return nil
	},
}

var questionUpdateCmd = &cobra.Command{
	Use:   "update [question-id]",
	Short: "Update question title and/or description",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		questionID := args[0]
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")

		if title == "" && description == "" {
			return fmt.Errorf("must specify --title and/or --description")
		}

		ctx := context.Background()
		err := wire.QuestionService().UpdateQuestion(ctx, primary.UpdateQuestionRequest{
			QuestionID:  questionID,
			Title:       title,
			Description: description,
		})
		if err != nil {
			return fmt.Errorf("failed to update question: %w", err)
		}

		fmt.Printf("✓ Question %s updated\n", questionID)
		return nil
	},
}

var questionPinCmd = &cobra.Command{
	Use:   "pin [question-id]",
	Short: "Pin question to keep it visible",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		questionID := args[0]

		ctx := context.Background()
		err := wire.QuestionService().PinQuestion(ctx, questionID)
		if err != nil {
			return fmt.Errorf("failed to pin question: %w", err)
		}

		fmt.Printf("✓ Question %s pinned 📌\n", questionID)
		return nil
	},
}

var questionUnpinCmd = &cobra.Command{
	Use:   "unpin [question-id]",
	Short: "Unpin question",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		questionID := args[0]

		ctx := context.Background()
		err := wire.QuestionService().UnpinQuestion(ctx, questionID)
		if err != nil {
			return fmt.Errorf("failed to unpin question: %w", err)
		}

		fmt.Printf("✓ Question %s unpinned\n", questionID)
		return nil
	},
}

var questionDeleteCmd = &cobra.Command{
	Use:   "delete [question-id]",
	Short: "Delete a question",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		questionID := args[0]

		ctx := context.Background()
		err := wire.QuestionService().DeleteQuestion(ctx, questionID)
		if err != nil {
			return fmt.Errorf("failed to delete question: %w", err)
		}

		fmt.Printf("✓ Question %s deleted\n", questionID)
		return nil
	},
}

func init() {
	// question create flags
	questionCreateCmd.Flags().StringP("mission", "m", "", "Mission ID (defaults to context)")
	questionCreateCmd.Flags().StringP("description", "d", "", "Question description")
	questionCreateCmd.Flags().String("investigation", "", "Investigation ID to attach question to")

	// question list flags
	questionListCmd.Flags().StringP("mission", "m", "", "Filter by mission")
	questionListCmd.Flags().String("investigation", "", "Filter by investigation")
	questionListCmd.Flags().StringP("status", "s", "", "Filter by status (open, answered)")

	// question answer flags
	questionAnswerCmd.Flags().StringP("answer", "a", "", "The answer to the question")

	// question update flags
	questionUpdateCmd.Flags().String("title", "", "New title")
	questionUpdateCmd.Flags().StringP("description", "d", "", "New description")

	// Register subcommands
	questionCmd.AddCommand(questionCreateCmd)
	questionCmd.AddCommand(questionListCmd)
	questionCmd.AddCommand(questionShowCmd)
	questionCmd.AddCommand(questionAnswerCmd)
	questionCmd.AddCommand(questionUpdateCmd)
	questionCmd.AddCommand(questionPinCmd)
	questionCmd.AddCommand(questionUnpinCmd)
	questionCmd.AddCommand(questionDeleteCmd)
}

// QuestionCmd returns the question command
func QuestionCmd() *cobra.Command {
	return questionCmd
}
