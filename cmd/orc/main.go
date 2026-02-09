package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/example/orc/internal/cli"
	"github.com/example/orc/internal/version"
)

func init() {
	// Respect CLICOLOR_FORCE for forcing colors when piped (e.g., in tmux popups)
	if os.Getenv("CLICOLOR_FORCE") == "1" {
		color.NoColor = false
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use:     "orc",
		Short:   "ORC - Orchestrator for Forest Factory commissions",
		Version: version.String(),
		Long: `ORC is a CLI tool for managing commissions, shipments, and tasks.
It coordinates IMPs (Implementation Agents) working in isolated workbenches (worktrees).`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Detect actor identity at CLI startup
			cli.DetectAndStoreActor()
			// Apply global tmux bindings (idempotent, no-op if tmux not running)
			cli.ApplyGlobalBindings()
		},
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
	rootCmd.AddCommand(cli.BootstrapCmd())
	rootCmd.AddCommand(cli.PrimeCmd())
	rootCmd.AddCommand(cli.TestCmd())
	rootCmd.AddCommand(cli.MailCmd())
	rootCmd.AddCommand(cli.NudgeCmd())
	rootCmd.AddCommand(cli.FocusCmd())

	// Entity commands (semantic model)
	rootCmd.AddCommand(cli.NoteCmd())
	rootCmd.AddCommand(cli.PlanCmd())
	rootCmd.AddCommand(cli.TomeCmd())

	// Receipt command
	rootCmd.AddCommand(cli.ReceiptCmd())

	// New entity commands (Stage 2)
	rootCmd.AddCommand(cli.GatehouseCmd())
	rootCmd.AddCommand(cli.KennelCmd())
	rootCmd.AddCommand(cli.PatrolCmd())
	rootCmd.AddCommand(cli.WatchdogCmd())
	rootCmd.AddCommand(cli.ApprovalCmd())
	rootCmd.AddCommand(cli.EscalationCmd())

	// Repository and PR commands
	rootCmd.AddCommand(cli.RepoCmd())
	rootCmd.AddCommand(cli.PRCmd())

	// Infrastructure commands (Factory/Workshop/Workbench hierarchy)
	rootCmd.AddCommand(cli.FactoryCmd())
	rootCmd.AddCommand(cli.WorkshopCmd())
	rootCmd.AddCommand(cli.WorkbenchCmd())
	rootCmd.AddCommand(cli.InfraCmd())
	rootCmd.AddCommand(cli.TmuxCmd())

	// Developer tools
	rootCmd.AddCommand(cli.ScaffoldCmd())
	rootCmd.AddCommand(cli.DebugCmd())
	rootCmd.AddCommand(cli.BackfillCmd())
	rootCmd.AddCommand(cli.LogCmd())

	// Claude Code integration
	rootCmd.AddCommand(cli.HookCmd())

	// Development utilities (orc-dev shim)
	rootCmd.AddCommand(cli.DevCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
