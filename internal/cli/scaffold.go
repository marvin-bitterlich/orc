package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/scaffold"
)

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold",
	Short: "Generate boilerplate code",
	Long:  "Generate entities, migrations, and other boilerplate code for ORC",
}

var scaffoldEntityCmd = &cobra.Command{
	Use:   "entity [name]",
	Short: "Generate a complete 7-layer entity stack",
	Long: `Generate all files for a new entity following ORC's hexagonal architecture:
  - Model (internal/models/)
  - Primary port (internal/ports/primary/)
  - Secondary port snippet (internal/ports/secondary/persistence.go)
  - Service implementation (internal/app/)
  - SQLite repository (internal/adapters/sqlite/)
  - CLI commands (internal/cli/)
  - Wire integration snippet (internal/wire/wire.go)

Field types: string, int, bool, time (add ? suffix for nullable)

Examples:
  orc scaffold entity widget --fields "name:string,value:int"
  orc scaffold entity alert --fields "title:string,severity:string" --status "active,acknowledged,resolved"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		entityName := args[0]
		fieldsStr, _ := cmd.Flags().GetString("fields")
		statusStr, _ := cmd.Flags().GetString("status")
		idPrefix, _ := cmd.Flags().GetString("id-prefix")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		// Build entity spec
		spec, err := scaffold.BuildEntitySpec(entityName, fieldsStr, statusStr, idPrefix)
		if err != nil {
			return err
		}

		// Generate files
		gen := scaffold.NewGenerator()
		result, err := gen.GenerateEntity(spec)
		if err != nil {
			return fmt.Errorf("failed to generate entity: %w", err)
		}

		// Display what will be created
		fmt.Printf("Generating entity '%s'", spec.Name)
		if len(spec.Fields) > 0 {
			fieldStrs := make([]string, len(spec.Fields))
			for i, f := range spec.Fields {
				nullable := ""
				if f.Nullable {
					nullable = "?"
				}
				fieldStrs[i] = fmt.Sprintf("%s(%s%s)", f.NameSnake, f.Type, nullable)
			}
			fmt.Printf(" with fields: %s", strings.Join(fieldStrs, ", "))
		}
		fmt.Println()
		fmt.Println()

		// Show files to create
		fmt.Println("Files to create:")
		for _, f := range result.Files {
			if f.Operation == "create" {
				fmt.Printf("  %s\n", f.Path)
			}
		}
		fmt.Println()

		// Show files to modify
		fmt.Println("Files to modify:")
		for _, f := range result.Files {
			if f.Operation != "create" {
				fmt.Printf("  %s\n", f.Path)
			}
		}
		fmt.Println()

		if dryRun {
			fmt.Println("(dry-run mode - no files written)")
			fmt.Println()
			// Show generated content for inspection
			for _, f := range result.Files {
				if f.Operation == "create" {
					fmt.Printf("--- %s ---\n", f.Path)
					fmt.Println(f.Content)
					fmt.Println()
				}
			}
			return nil
		}

		// Confirm
		fmt.Print("Proceed? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Aborted.")
			return nil
		}

		// Write files
		for _, f := range result.Files {
			if err := writeGeneratedFile(f); err != nil {
				return fmt.Errorf("failed to write %s: %w", f.Path, err)
			}
			if f.Operation == "create" {
				fmt.Printf("✓ Created %s\n", f.Path)
			} else {
				fmt.Printf("✓ Modified %s\n", f.Path)
			}
		}

		// Show next steps
		fmt.Println()
		fmt.Println("Next steps:")
		for i, step := range result.NextSteps {
			fmt.Printf("  %d. %s\n", i+1, step)
		}

		return nil
	},
}

var scaffoldMigrationCmd = &cobra.Command{
	Use:   "migration [name]",
	Short: "Generate a migration boilerplate",
	Long: `Generate a new migration function in internal/db/migrations.go.

The migration will be assigned the next available version number automatically.

Examples:
  orc scaffold migration create_widgets_table
  orc scaffold migration add_priority_to_widgets`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		migrationName := args[0]
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		forEntity, _ := cmd.Flags().GetString("for-entity")

		// Detect next version
		nextVersion, err := detectNextMigrationVersion()
		if err != nil {
			return fmt.Errorf("failed to detect migration version: %w", err)
		}

		// Build migration spec
		migSpec := &scaffold.MigrationSpec{
			Version:   nextVersion,
			Name:      scaffold.ToPascalCase(migrationName),
			NameSnake: scaffold.ToSnakeCase(migrationName),
		}

		// If creating for an entity, parse entity spec
		var entitySpec *scaffold.EntitySpec
		if forEntity != "" {
			// Parse the entity spec (fields/status should be provided)
			fieldsStr, _ := cmd.Flags().GetString("fields")
			statusStr, _ := cmd.Flags().GetString("status")
			idPrefix, _ := cmd.Flags().GetString("id-prefix")

			entitySpec, err = scaffold.BuildEntitySpec(forEntity, fieldsStr, statusStr, idPrefix)
			if err != nil {
				return err
			}
		}

		// Generate migration
		gen := scaffold.NewGenerator()
		result, err := gen.GenerateMigration(migSpec, entitySpec)
		if err != nil {
			return fmt.Errorf("failed to generate migration: %w", err)
		}

		fmt.Printf("Generating migration V%d: %s\n", migSpec.Version, migSpec.NameSnake)
		fmt.Println()
		fmt.Println("File to modify:")
		for _, f := range result.Files {
			fmt.Printf("  %s\n", f.Path)
		}
		fmt.Println()

		if dryRun {
			fmt.Println("(dry-run mode - no files written)")
			fmt.Println()
			for _, f := range result.Files {
				fmt.Printf("--- Snippet for %s ---\n", f.Path)
				fmt.Println(f.Snippet)
			}
			return nil
		}

		// Confirm
		fmt.Print("Proceed? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Aborted.")
			return nil
		}

		// Write files (append snippets)
		for _, f := range result.Files {
			if err := appendToMigrationsFile(f.Snippet, migSpec.Version, migSpec.NameSnake); err != nil {
				return fmt.Errorf("failed to modify %s: %w", f.Path, err)
			}
			fmt.Printf("✓ Modified %s\n", f.Path)
		}

		// Show next steps
		fmt.Println()
		fmt.Println("Next steps:")
		for i, step := range result.NextSteps {
			fmt.Printf("  %d. %s\n", i+1, step)
		}

		return nil
	},
}

func init() {
	// Entity flags
	scaffoldEntityCmd.Flags().String("fields", "", "Field specifications (e.g., 'name:string,count:int,active:bool')")
	scaffoldEntityCmd.Flags().String("status", "", "Status values for FSM (e.g., 'draft,active,completed')")
	scaffoldEntityCmd.Flags().String("id-prefix", "", "ID prefix (defaults to uppercase entity name)")
	scaffoldEntityCmd.Flags().Bool("dry-run", false, "Preview without writing files")

	// Migration flags
	scaffoldMigrationCmd.Flags().Bool("dry-run", false, "Preview without writing files")
	scaffoldMigrationCmd.Flags().String("for-entity", "", "Generate table creation for entity")
	scaffoldMigrationCmd.Flags().String("fields", "", "Field specifications (for --for-entity)")
	scaffoldMigrationCmd.Flags().String("status", "", "Status values (for --for-entity)")
	scaffoldMigrationCmd.Flags().String("id-prefix", "", "ID prefix (for --for-entity)")

	scaffoldCmd.AddCommand(scaffoldEntityCmd)
	scaffoldCmd.AddCommand(scaffoldMigrationCmd)
}

// ScaffoldCmd returns the scaffold command
func ScaffoldCmd() *cobra.Command {
	return scaffoldCmd
}

// writeGeneratedFile writes a generated file to disk.
func writeGeneratedFile(f scaffold.GeneratedFile) error {
	if f.Operation == "create" {
		// Ensure directory exists
		dir := filepath.Dir(f.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		return os.WriteFile(f.Path, []byte(f.Content), 0644)
	}

	// For append/insert operations, we'd need to modify existing files
	// This is more complex and would require finding the insertion point
	fmt.Printf("  NOTE: %s requires manual integration - see snippet below:\n", f.Path)
	fmt.Println("  ---")
	fmt.Println(f.Snippet)
	fmt.Println("  ---")
	return nil
}

// detectNextMigrationVersion reads migrations.go and finds the next version number.
func detectNextMigrationVersion() (int, error) {
	migrationsPath := "internal/db/migrations.go"

	content, err := os.ReadFile(migrationsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil // If file doesn't exist, start at 1
		}
		return 0, err
	}

	// Find all Version: N patterns
	re := regexp.MustCompile(`Version:\s*(\d+)`)
	matches := re.FindAllStringSubmatch(string(content), -1)

	maxVersion := 0
	for _, match := range matches {
		if len(match) > 1 {
			v, _ := strconv.Atoi(match[1])
			if v > maxVersion {
				maxVersion = v
			}
		}
	}

	return maxVersion + 1, nil
}

// appendToMigrationsFile appends migration to the migrations.go file.
func appendToMigrationsFile(snippet string, version int, name string) error {
	migrationsPath := "internal/db/migrations.go"

	content, err := os.ReadFile(migrationsPath)
	if err != nil {
		return err
	}

	// Find the migrations slice closing bracket and insert before it
	// Also need to add the migration function at the end
	contentStr := string(content)

	// Add to migrations slice (find the closing of the slice)
	migrationsEntry := fmt.Sprintf(`	{
		Version: %d,
		Name:    "%s",
		Up:      migrationV%d,
	},
`, version, name, version)

	// Find the last migration entry and add after it
	lastVersionRe := regexp.MustCompile(`(\s*\{\s*Version:\s*\d+[^}]+\},)\s*\}`)
	contentStr = lastVersionRe.ReplaceAllString(contentStr, "${1}\n"+migrationsEntry+"}")

	// Add the migration function at the end of the file
	contentStr += "\n" + snippet

	return os.WriteFile(migrationsPath, []byte(contentStr), 0644)
}
