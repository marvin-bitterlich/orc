// +build ignore

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// Task represents a task record from the database
type Task struct {
	ID          string
	ShipmentID  sql.NullString
	MissionID   string
	Title       string
	Description sql.NullString
	Type        sql.NullString
	Status      string
}

// NoteType mapping based on title prefix
var prefixToNoteType = map[string]string{
	"Learning:":   "learning",
	"Rule:":       "learning",
	"Concern:":    "concern",
	"Reflection:": "learning",
	"Finding:":    "finding",
	"FRQ:":        "frq",
	"Bug:":        "bug",
}

// Note: FRQ = Feature Request - these are ideas/wishes, not actionable tasks

func main() {
	dryRun := flag.Bool("dry-run", false, "Preview migration without executing")
	flag.Parse()

	// Find ORC database
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home dir: %v\n", err)
		os.Exit(1)
	}
	dbPath := filepath.Join(homeDir, ".orc", "orc.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Find tasks to migrate (across ALL shipments)
	tasks, err := findTasksToMigrate(db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding tasks: %v\n", err)
		os.Exit(1)
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found to migrate")
		return
	}

	fmt.Printf("Found %d task(s) to migrate:\n\n", len(tasks))

	for _, task := range tasks {
		noteType := determineNoteType(task.Title)
		cleanTitle := cleanTitle(task.Title)

		fmt.Printf("  %s: %s\n", task.ID, task.Title)
		fmt.Printf("    -> Note type: %s\n", noteType)
		fmt.Printf("    -> Clean title: %s\n", cleanTitle)
		fmt.Println()
	}

	if *dryRun {
		fmt.Println("=== DRY RUN - No changes made ===")
		return
	}

	fmt.Println("=== Executing migration ===")
	fmt.Println()

	migrated := 0
	for _, task := range tasks {
		noteType := determineNoteType(task.Title)
		cleanTitle := cleanTitle(task.Title)

		err := migrateTask(db, task, cleanTitle, noteType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error migrating %s: %v\n", task.ID, err)
			continue
		}

		fmt.Printf("✓ Migrated %s -> NOTE\n", task.ID)
		migrated++
	}

	fmt.Printf("\n=== Migration complete: %d/%d tasks migrated ===\n", migrated, len(tasks))
}

func findTasksToMigrate(db *sql.DB) ([]Task, error) {
	// Build regex pattern for titles that should be migrated
	patterns := []string{}
	for prefix := range prefixToNoteType {
		patterns = append(patterns, regexp.QuoteMeta(prefix))
	}
	pattern := "^(" + strings.Join(patterns, "|") + ")"
	re := regexp.MustCompile(pattern)

	query := `
		SELECT id, shipment_id, mission_id, title, description, type, status
		FROM tasks
		ORDER BY shipment_id, created_at ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.ShipmentID, &t.MissionID, &t.Title, &t.Description, &t.Type, &t.Status)
		if err != nil {
			return nil, err
		}

		// Only include tasks that match the pattern
		if re.MatchString(t.Title) {
			tasks = append(tasks, t)
		}
	}

	return tasks, nil
}

func determineNoteType(title string) string {
	for prefix, noteType := range prefixToNoteType {
		if strings.HasPrefix(title, prefix) {
			return noteType
		}
	}
	return "learning" // default
}

func cleanTitle(title string) string {
	for prefix := range prefixToNoteType {
		if strings.HasPrefix(title, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(title, prefix))
		}
	}
	return title
}

func migrateTask(db *sql.DB, task Task, cleanTitle, noteType string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Generate note ID
	var maxID int
	err = tx.QueryRow("SELECT COALESCE(MAX(CAST(SUBSTR(id, 6) AS INTEGER)), 0) FROM notes").Scan(&maxID)
	if err != nil {
		return err
	}
	noteID := fmt.Sprintf("NOTE-%03d", maxID+1)

	// Create note
	_, err = tx.Exec(`
		INSERT INTO notes (id, mission_id, title, content, type, shipment_id, promoted_from_id, promoted_from_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, noteID, task.MissionID, cleanTitle, task.Description, noteType, task.ShipmentID, task.ID, "task")
	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}

	// Delete original task
	_, err = tx.Exec("DELETE FROM tasks WHERE id = ?", task.ID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return tx.Commit()
}
