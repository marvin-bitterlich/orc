package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/example/orc/internal/context"
	"github.com/example/orc/internal/models"
	"github.com/spf13/cobra"
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
			missionID = context.GetContextMissionID()
			if missionID == "" {
				return fmt.Errorf("no mission context detected\nHint: Use --mission flag or run from a grove/mission directory")
			}
		}

		question, err := models.CreateQuestion(investigationID, missionID, title, description)
		if err != nil {
			return fmt.Errorf("failed to create question: %w", err)
		}

		fmt.Printf("✓ Created question %s: %s\n", question.ID, question.Title)
		if question.InvestigationID.Valid {
			fmt.Printf("  Investigation: %s\n", question.InvestigationID.String)
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
			missionID = context.GetContextMissionID()
		}

		var questions []*models.Question
		var err error

		questions, err = models.ListQuestions(investigationID, status)
		if err != nil {
			return fmt.Errorf("failed to list questions: %w", err)
		}

		// Filter by mission if specified
		if missionID != "" {
			var filtered []*models.Question
			for _, q := range questions {
				if q.MissionID == missionID {
					filtered = append(filtered, q)
				}
			}
			questions = filtered
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
			if q.InvestigationID.Valid {
				inv = q.InvestigationID.String
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

		question, err := models.GetQuestion(questionID)
		if err != nil {
			return fmt.Errorf("question not found: %w", err)
		}

		fmt.Printf("Question: %s\n", question.ID)
		fmt.Printf("Title: %s\n", question.Title)
		if question.Description.Valid {
			fmt.Printf("Description: %s\n", question.Description.String)
		}
		fmt.Printf("Status: %s\n", question.Status)
		if question.Answer.Valid {
			fmt.Printf("Answer: %s\n", question.Answer.String)
		}
		fmt.Printf("Mission: %s\n", question.MissionID)
		if question.InvestigationID.Valid {
			fmt.Printf("Investigation: %s\n", question.InvestigationID.String)
		}
		if question.ConclaveID.Valid {
			fmt.Printf("Conclave: %s\n", question.ConclaveID.String)
		}
		if question.Pinned {
			fmt.Printf("Pinned: yes\n")
		}
		if question.PromotedFromID.Valid {
			fmt.Printf("Promoted from: %s (%s)\n", question.PromotedFromID.String, question.PromotedFromType.String)
		}
		fmt.Printf("Created: %s\n", question.CreatedAt.Format("2006-01-02 15:04"))
		if question.AnsweredAt.Valid {
			fmt.Printf("Answered: %s\n", question.AnsweredAt.Time.Format("2006-01-02 15:04"))
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

		err := models.AnswerQuestion(questionID, answer)
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

		err := models.UpdateQuestion(questionID, title, description)
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

		err := models.PinQuestion(questionID)
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

		err := models.UnpinQuestion(questionID)
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

		err := models.DeleteQuestion(questionID)
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
