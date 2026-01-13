package main

import (
	"fmt"
	"os"

	"github.com/looneym/orc/internal/cli"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "orc",
		Short: "ORC - Orchestrator for Forest Factory missions",
		Long: `ORC is a CLI tool for managing missions, groves, and work orders.
It coordinates IMPs (Implementation Agents) working in isolated groves (worktrees).`,
	}

	// Add subcommands
	rootCmd.AddCommand(cli.InitCmd())
	rootCmd.AddCommand(cli.MissionCmd())
	rootCmd.AddCommand(cli.WorkOrderCmd())
	rootCmd.AddCommand(cli.GroveCmd())
	rootCmd.AddCommand(cli.HandoffCmd())
	rootCmd.AddCommand(cli.SummaryCmd())
	rootCmd.AddCommand(cli.StatusCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
