package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

// FactoryCmd returns the factory command
func FactoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "factory",
		Short: "Manage factories (TMux sessions)",
		Long:  `Create and manage factories - persistent TMux runtime environments.`,
	}

	cmd.AddCommand(factoryCreateCmd())
	cmd.AddCommand(factoryListCmd())
	cmd.AddCommand(factoryShowCmd())
	cmd.AddCommand(factoryDeleteCmd())

	return cmd
}

func factoryCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new factory",
		Long: `Create a new factory (TMux session environment).

A Factory is the top-level runtime environment, typically corresponding
to a TMux session. Factories contain Workshops, which contain Workbenches.

Examples:
  orc factory create phoenix-dev
  orc factory create staging-env`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			name := args[0]

			resp, err := wire.FactoryService().CreateFactory(ctx, primary.CreateFactoryRequest{
				Name: name,
			})
			if err != nil {
				return fmt.Errorf("failed to create factory: %w", err)
			}

			fmt.Printf("✓ Created factory %s: %s\n", resp.FactoryID, resp.Factory.Name)
			return nil
		},
	}

	return cmd
}

func factoryListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all factories",
		Long:  `List all factories with their current status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			factories, err := wire.FactoryService().ListFactories(ctx, primary.FactoryFilters{})
			if err != nil {
				return fmt.Errorf("failed to list factories: %w", err)
			}

			if len(factories) == 0 {
				fmt.Println("No factories found.")
				fmt.Println()
				fmt.Println("Create your first factory:")
				fmt.Println("  orc factory create my-factory")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tSTATUS\tCREATED")
			fmt.Fprintln(w, "--\t----\t------\t-------")

			for _, f := range factories {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					f.ID,
					f.Name,
					f.Status,
					f.CreatedAt,
				)
			}

			w.Flush()
			return nil
		},
	}

	return cmd
}

func factoryShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [factory-id]",
		Short: "Show factory details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			factory, err := wire.FactoryService().GetFactory(ctx, args[0])
			if err != nil {
				return fmt.Errorf("factory not found: %w", err)
			}

			fmt.Printf("Factory: %s\n", factory.ID)
			fmt.Printf("Name: %s\n", factory.Name)
			fmt.Printf("Status: %s\n", factory.Status)
			fmt.Printf("Created: %s\n", factory.CreatedAt)

			return nil
		},
	}
}

func factoryDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [factory-id]",
		Short: "Delete a factory",
		Long: `Delete a factory from the database.

WARNING: This is a destructive operation. Factories with workshops
or commissions require the --force flag.

Examples:
  orc factory delete FACT-001
  orc factory delete FACT-001 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			factoryID := args[0]

			err := wire.FactoryService().DeleteFactory(ctx, primary.DeleteFactoryRequest{
				FactoryID: factoryID,
				Force:     force,
			})
			if err != nil {
				return err
			}

			fmt.Printf("✓ Factory %s deleted\n", factoryID)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force delete even with workshops/commissions")

	return cmd
}
