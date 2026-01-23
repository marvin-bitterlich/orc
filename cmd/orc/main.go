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
		Short:   "ORC - Orchestrator for Forest Factory commissions",
		Version: version.String(),
		Long: `ORC is a CLI tool for managing commissions, shipments, and tasks.
It coordinates IMPs (Implementation Agents) working in isolated workbenches (worktrees).`,
	}

	// Add subcommands
	rootCmd.AddCommand(cli.InitCmd())
	rootCmd.AddCommand(cli.DoctorCmd())
	rootCmd.AddCommand(cli.CommissionCmd())
	rootCmd.AddCommand(cli.ShipmentCmd())
	rootCmd.AddCommand(cli.TaskCmd())
	rootCmd.AddCommand(cli.TagCmd())
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
	rootCmd.AddCommand(cli.PlanCmd())
	rootCmd.AddCommand(cli.TomeCmd())
	rootCmd.AddCommand(cli.InvestigationCmd())
	rootCmd.AddCommand(cli.ConclaveCmd())

	// Spec-Kit execution tracking (Work Orders, Cycles, and Cycle Work Orders)
	rootCmd.AddCommand(cli.WorkOrderCmd())
	rootCmd.AddCommand(cli.CycleCmd())
	rootCmd.AddCommand(cli.CycleWorkOrderCmd())

	// Spec-Kit receipts (Cycle Receipts and Receipts)
	rootCmd.AddCommand(cli.CycleReceiptCmd())
	rootCmd.AddCommand(cli.ReceiptCmd())

	// Repository and PR commands
	rootCmd.AddCommand(cli.RepoCmd())
	rootCmd.AddCommand(cli.PRCmd())

	// Infrastructure commands (Factory/Workshop/Workbench hierarchy)
	rootCmd.AddCommand(cli.FactoryCmd())
	rootCmd.AddCommand(cli.WorkshopCmd())
	rootCmd.AddCommand(cli.WorkbenchCmd())

	// Developer tools
	rootCmd.AddCommand(cli.ScaffoldCmd())
	rootCmd.AddCommand(cli.DebugCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
