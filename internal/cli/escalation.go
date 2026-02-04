package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	orcctx "github.com/example/orc/internal/context"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

var escalationCmd = &cobra.Command{
	Use:   "escalation",
	Short: "Manage escalations",
	Long:  "List and view escalations in the ORC ledger",
}

var escalationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List escalations",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		status, _ := cmd.Flags().GetString("status")

		escalations, err := wire.EscalationService().ListEscalations(ctx, primary.EscalationFilters{
			Status: status,
		})
		if err != nil {
			return fmt.Errorf("failed to list escalations: %w", err)
		}

		if len(escalations) == 0 {
			fmt.Println("No escalations found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tPLAN\tTASK\tSTATUS\tORIGIN\tTARGET\tCREATED")
		fmt.Fprintln(w, "--\t----\t----\t------\t------\t------\t-------")
		for _, item := range escalations {
			target := item.TargetActorID
			if target == "" {
				target = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				item.ID,
				item.PlanID,
				item.TaskID,
				item.Status,
				item.OriginActorID,
				target,
				item.CreatedAt,
			)
		}
		w.Flush()
		return nil
	},
}

var escalationShowCmd = &cobra.Command{
	Use:   "show [escalation-id]",
	Short: "Show escalation details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		escalationID := args[0]

		escalation, err := wire.EscalationService().GetEscalation(ctx, escalationID)
		if err != nil {
			return fmt.Errorf("escalation not found: %w", err)
		}

		fmt.Printf("Escalation: %s\n", escalation.ID)
		fmt.Printf("Plan: %s\n", escalation.PlanID)
		fmt.Printf("Task: %s\n", escalation.TaskID)
		if escalation.ApprovalID != "" {
			fmt.Printf("Approval: %s\n", escalation.ApprovalID)
		}
		fmt.Printf("Reason: %s\n", escalation.Reason)
		fmt.Printf("Status: %s\n", escalation.Status)
		fmt.Printf("Routing Rule: %s\n", escalation.RoutingRule)
		fmt.Printf("Origin Actor: %s\n", escalation.OriginActorID)
		if escalation.TargetActorID != "" {
			fmt.Printf("Target Actor: %s\n", escalation.TargetActorID)
		}
		if escalation.Resolution != "" {
			fmt.Printf("Resolution: %s\n", escalation.Resolution)
		}
		if escalation.ResolvedBy != "" {
			fmt.Printf("Resolved By: %s\n", escalation.ResolvedBy)
		}
		fmt.Printf("Created: %s\n", escalation.CreatedAt)
		if escalation.ResolvedAt != "" {
			fmt.Printf("Resolved: %s\n", escalation.ResolvedAt)
		}

		return nil
	},
}

var escalationResolveCmd = &cobra.Command{
	Use:   "resolve [escalation-id]",
	Short: "Resolve an escalation",
	Long:  "Resolve an escalation with an outcome (approved or rejected) and optional feedback",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		escalationID := args[0]
		outcome, _ := cmd.Flags().GetString("outcome")
		resolution, _ := cmd.Flags().GetString("resolution")

		if outcome == "" {
			return fmt.Errorf("--outcome is required (approved or rejected)")
		}

		if outcome != "approved" && outcome != "rejected" {
			return fmt.Errorf("--outcome must be 'approved' or 'rejected'")
		}

		// Get resolver actor from gatehouse context
		resolvedBy := orcctx.GetContextWorkbenchID()
		// Note: Gatehouses use GATE-xxx IDs, but they're detected same way as workbenches
		// If no context, allow resolution without resolver ID (for admin use)

		ctx := NewContext()
		err := wire.EscalationService().ResolveEscalation(ctx, primary.ResolveEscalationRequest{
			EscalationID: escalationID,
			Outcome:      outcome,
			Resolution:   resolution,
			ResolvedBy:   resolvedBy,
		})
		if err != nil {
			return fmt.Errorf("failed to resolve escalation: %w", err)
		}

		status := "resolved"
		if outcome == "rejected" {
			status = "dismissed"
		}
		fmt.Printf("✓ Escalation %s %s\n", escalationID, status)
		if resolution != "" {
			fmt.Printf("  Resolution: %s\n", resolution)
		}
		return nil
	},
}

func init() {
	// escalation list flags
	escalationListCmd.Flags().StringP("status", "s", "", "Filter by status (pending|resolved|dismissed)")

	// escalation resolve flags
	escalationResolveCmd.Flags().StringP("outcome", "o", "", "Resolution outcome: approved or rejected (required)")
	escalationResolveCmd.Flags().StringP("resolution", "r", "", "Optional resolution feedback/explanation")

	// Register subcommands
	escalationCmd.AddCommand(escalationListCmd)
	escalationCmd.AddCommand(escalationShowCmd)
	escalationCmd.AddCommand(escalationResolveCmd)
}

// EscalationCmd returns the escalation command
func EscalationCmd() *cobra.Command {
	return escalationCmd
}
