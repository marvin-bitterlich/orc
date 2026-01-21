package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/cli"
	"github.com/example/orc/internal/version"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "orc",
		Short:   "ORC - Orchestrator for Forest Factory missions",
		Version: version.String(),
		Long: `ORC is a CLI tool for managing missions, groves, shipments, and tasks.
It coordinates IMPs (Implementation Agents) working in isolated groves (worktrees).`,
	}

	// Add subcommands
	rootCmd.AddCommand(cli.InitCmd())
	rootCmd.AddCommand(cli.DoctorCmd())
	rootCmd.AddCommand(cli.MissionCmd())
	rootCmd.AddCommand(cli.ShipmentCmd())
	rootCmd.AddCommand(cli.TaskCmd())
	rootCmd.AddCommand(cli.TagCmd())
	rootCmd.AddCommand(cli.GroveCmd())
	rootCmd.AddCommand(cli.HandoffCmd())
	rootCmd.AddCommand(cli.SummaryCmd())
	rootCmd.AddCommand(cli.StatusCmd())
	rootCmd.AddCommand(cli.AttachCmd())
	rootCmd.AddCommand(cli.ConnectCmd())
	rootCmd.AddCommand(cli.PrimeCmd())
	rootCmd.AddCommand(cli.TestCmd())
	rootCmd.AddCommand(cli.MailCmd())
	rootCmd.AddCommand(cli.NudgeCmd())
	rootCmd.AddCommand(cli.FocusCmd())

	// Entity commands (semantic model)
	rootCmd.AddCommand(cli.NoteCmd())
	rootCmd.AddCommand(cli.QuestionCmd())
	rootCmd.AddCommand(cli.PlanCmd())
	rootCmd.AddCommand(cli.TomeCmd())
	rootCmd.AddCommand(cli.InvestigationCmd())
	rootCmd.AddCommand(cli.ConclaveCmd())

	// Developer tools
	rootCmd.AddCommand(cli.ScaffoldCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
