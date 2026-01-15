package db

import (
	"database/sql"
	"fmt"
)

// schemaVersion tracks the current schema version
const currentSchemaVersion = 11

// Migration represents a database migration
type Migration struct {
	Version int
	Name    string
	Up      func(*sql.DB) error
}

// migrations is the list of all migrations in order
var migrations = []Migration{
	{
		Version: 1,
		Name:    "flatten_hierarchy_and_grove_schema",
		Up:      migrationV1,
	},
	{
		Version: 2,
		Name:    "add_phase_field_to_work_orders",
		Up:      migrationV2,
	},
	{
		Version: 3,
		Name:    "consolidate_status_and_phase_fields",
		Up:      migrationV3,
	},
	{
		Version: 4,
		Name:    "add_pinned_field_to_work_orders",
		Up:      migrationV4,
	},
	{
		Version: 5,
		Name:    "add_messages_table_for_agent_mail",
		Up:      migrationV5,
	},
	{
		Version: 6,
		Name:    "convert_work_orders_to_epics_rabbit_holes_tasks",
		Up:      migrationV6,
	},
	{
		Version: 7,
		Name:    "add_pinned_field_to_missions",
		Up:      migrationV7,
	},
	{
		Version: 8,
		Name:    "add_tags_and_task_tags_tables",
		Up:      migrationV8,
	},
	{
		Version: 9,
		Name:    "add_awaiting_approval_status",
		Up:      migrationV9,
	},
	{
		Version: 10,
		Name:    "add_ready_to_implement_status",
		Up:      migrationV10,
	},
	{
		Version: 11,
		Name:    "remove_implement_status",
		Up:      migrationV11,
	},
}

// RunMigrations executes all pending migrations
func RunMigrations() error {
	db, err := GetDB()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	// Create schema_version table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_version table: %w", err)
	}

	// Get current schema version
	var currentVersion int
	err = db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("failed to get current schema version: %w", err)
	}

	// Run pending migrations
	for _, migration := range migrations {
		if migration.Version <= currentVersion {
			continue
		}

		fmt.Printf("Running migration %d: %s\n", migration.Version, migration.Name)

		// Begin transaction
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %d: %w", migration.Version, err)
		}

		// Run migration
		if err := migration.Up(db); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d failed: %w", migration.Version, err)
		}

		// Record migration
		_, err = tx.Exec("INSERT INTO schema_version (version) VALUES (?)", migration.Version)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
		}

		fmt.Printf("✓ Migration %d completed\n", migration.Version)
	}

	return nil
}

// migrationV1 flattens the hierarchy (removes operations/expeditions) and updates grove schema
func migrationV1(db *sql.DB) error {
	// Step 1: Create new work_orders table with updated schema
	_, err := db.Exec(`
		CREATE TABLE work_orders_new (
			id TEXT PRIMARY KEY,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			type TEXT CHECK(type IN ('research', 'implementation', 'fix', 'documentation', 'maintenance')),
			status TEXT NOT NULL CHECK(status IN ('backlog', 'next', 'in_progress', 'complete')) DEFAULT 'backlog',
			priority TEXT CHECK(priority IN ('low', 'medium', 'high')),
			parent_id TEXT,
			assigned_grove_id TEXT,
			context_ref TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			claimed_at DATETIME,
			completed_at DATETIME,
			FOREIGN KEY (mission_id) REFERENCES missions(id),
			FOREIGN KEY (parent_id) REFERENCES work_orders_new(id),
			FOREIGN KEY (assigned_grove_id) REFERENCES groves(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create work_orders_new: %w", err)
	}

	// Step 2: Migrate work orders data (operation_id -> mission_id via operations table)
	// All existing operations belong to MISSION-001, so we can hardcode it
	_, err = db.Exec(`
		INSERT INTO work_orders_new (
			id, mission_id, title, description, status, context_ref,
			created_at, updated_at, claimed_at, completed_at
		)
		SELECT
			wo.id,
			'MISSION-001' as mission_id,
			wo.title,
			wo.description,
			wo.status,
			wo.context_ref,
			wo.created_at,
			wo.updated_at,
			wo.claimed_at,
			wo.completed_at
		FROM work_orders wo
	`)
	if err != nil {
		return fmt.Errorf("failed to migrate work orders: %w", err)
	}

	// Step 3: Create new groves table with updated schema
	_, err = db.Exec(`
		CREATE TABLE groves_new (
			id TEXT PRIMARY KEY,
			mission_id TEXT NOT NULL,
			name TEXT NOT NULL,
			path TEXT NOT NULL UNIQUE,
			repos TEXT,
			status TEXT NOT NULL CHECK(status IN ('active', 'archived')) DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (mission_id) REFERENCES missions(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create groves_new: %w", err)
	}

	// Step 4: Migrate groves data (no existing data, but keep table structure)
	_, err = db.Exec(`
		INSERT INTO groves_new (id, mission_id, name, path, repos, status, created_at, updated_at)
		SELECT
			id,
			'MISSION-001' as mission_id,
			'grove-' || substr(id, 7) as name,
			path,
			repos,
			CASE WHEN status = 'idle' THEN 'active' ELSE status END,
			created_at,
			updated_at
		FROM groves
		WHERE status != 'idle' OR status = 'idle'
	`)
	if err != nil {
		return fmt.Errorf("failed to migrate groves: %w", err)
	}

	// Step 5: Update handoffs table (remove expedition_id, operation_id references)
	_, err = db.Exec(`
		CREATE TABLE handoffs_new (
			id TEXT PRIMARY KEY,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			handoff_note TEXT NOT NULL,
			active_mission_id TEXT,
			active_work_orders TEXT,
			active_grove_id TEXT,
			todos_snapshot TEXT,
			graphiti_episode_uuid TEXT,
			FOREIGN KEY (active_mission_id) REFERENCES missions(id),
			FOREIGN KEY (active_grove_id) REFERENCES groves_new(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create handoffs_new: %w", err)
	}

	// Migrate handoffs (convert active_work_order_id to JSON array)
	_, err = db.Exec(`
		INSERT INTO handoffs_new (
			id, created_at, handoff_note, active_mission_id, active_work_orders,
			active_grove_id, todos_snapshot, graphiti_episode_uuid
		)
		SELECT
			id,
			created_at,
			handoff_note,
			active_mission_id,
			CASE
				WHEN active_work_order_id IS NOT NULL
				THEN '["' || active_work_order_id || '"]'
				ELSE NULL
			END as active_work_orders,
			NULL as active_grove_id,
			todos_snapshot,
			graphiti_episode_uuid
		FROM handoffs
	`)
	if err != nil {
		return fmt.Errorf("failed to migrate handoffs: %w", err)
	}

	// Step 6: Drop old tables
	_, err = db.Exec("DROP TABLE IF EXISTS work_orders")
	if err != nil {
		return fmt.Errorf("failed to drop old work_orders: %w", err)
	}

	_, err = db.Exec("DROP TABLE IF EXISTS groves")
	if err != nil {
		return fmt.Errorf("failed to drop old groves: %w", err)
	}

	_, err = db.Exec("DROP TABLE IF EXISTS handoffs")
	if err != nil {
		return fmt.Errorf("failed to drop old handoffs: %w", err)
	}

	// Drop obsolete tables
	_, err = db.Exec("DROP TABLE IF EXISTS plans")
	if err != nil {
		return fmt.Errorf("failed to drop plans: %w", err)
	}

	_, err = db.Exec("DROP TABLE IF EXISTS dependencies")
	if err != nil {
		return fmt.Errorf("failed to drop dependencies: %w", err)
	}

	_, err = db.Exec("DROP TABLE IF EXISTS expeditions")
	if err != nil {
		return fmt.Errorf("failed to drop expeditions: %w", err)
	}

	_, err = db.Exec("DROP TABLE IF EXISTS operations")
	if err != nil {
		return fmt.Errorf("failed to drop operations: %w", err)
	}

	// Step 7: Rename new tables to original names
	_, err = db.Exec("ALTER TABLE work_orders_new RENAME TO work_orders")
	if err != nil {
		return fmt.Errorf("failed to rename work_orders_new: %w", err)
	}

	_, err = db.Exec("ALTER TABLE groves_new RENAME TO groves")
	if err != nil {
		return fmt.Errorf("failed to rename groves_new: %w", err)
	}

	_, err = db.Exec("ALTER TABLE handoffs_new RENAME TO handoffs")
	if err != nil {
		return fmt.Errorf("failed to rename handoffs_new: %w", err)
	}

	// Step 8: Create indexes
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_work_orders_mission ON work_orders(mission_id);
		CREATE INDEX IF NOT EXISTS idx_work_orders_status ON work_orders(status);
		CREATE INDEX IF NOT EXISTS idx_work_orders_parent ON work_orders(parent_id);
		CREATE INDEX IF NOT EXISTS idx_groves_mission ON groves(mission_id);
		CREATE INDEX IF NOT EXISTS idx_groves_status ON groves(status);
		CREATE INDEX IF NOT EXISTS idx_handoffs_created ON handoffs(created_at DESC);
	`)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// migrationV2 adds phase field to work_orders
func migrationV2(db *sql.DB) error {
	// Add phase column to work_orders table
	// Options: ready (default), paused, design, implement, deploy, blocked
	_, err := db.Exec(`
		ALTER TABLE work_orders ADD COLUMN phase TEXT
		CHECK(phase IN ('ready', 'paused', 'design', 'implement', 'deploy', 'blocked'))
		DEFAULT 'ready'
	`)
	if err != nil {
		return fmt.Errorf("failed to add phase column: %w", err)
	}

	// Set existing work orders to appropriate phase based on status
	// backlog/next -> ready
	// in_progress -> implement
	// complete -> deploy
	_, err = db.Exec(`
		UPDATE work_orders SET phase = CASE
			WHEN status = 'in_progress' THEN 'implement'
			WHEN status = 'complete' THEN 'deploy'
			ELSE 'ready'
		END
	`)
	if err != nil {
		return fmt.Errorf("failed to set initial phase values: %w", err)
	}

	return nil
}

// migrationV3 consolidates status and phase into single status field
func migrationV3(db *sql.DB) error {
	// Create new work_orders table with consolidated status field
	// Status values: ready, design, implement, deploy, blocked, paused, complete
	_, err := db.Exec(`
		CREATE TABLE work_orders_new (
			id TEXT PRIMARY KEY,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			type TEXT CHECK(type IN ('research', 'implementation', 'fix', 'documentation', 'maintenance')),
			status TEXT NOT NULL CHECK(status IN ('ready', 'design', 'implement', 'deploy', 'blocked', 'paused', 'complete')) DEFAULT 'ready',
			priority TEXT CHECK(priority IN ('low', 'medium', 'high')),
			parent_id TEXT,
			assigned_grove_id TEXT,
			context_ref TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			claimed_at DATETIME,
			completed_at DATETIME,
			FOREIGN KEY (mission_id) REFERENCES missions(id),
			FOREIGN KEY (parent_id) REFERENCES work_orders_new(id),
			FOREIGN KEY (assigned_grove_id) REFERENCES groves(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create work_orders_new: %w", err)
	}

	// Migrate data: for complete work orders keep 'complete', otherwise use phase value
	_, err = db.Exec(`
		INSERT INTO work_orders_new (
			id, mission_id, title, description, type, status, priority,
			parent_id, assigned_grove_id, context_ref,
			created_at, updated_at, claimed_at, completed_at
		)
		SELECT
			id, mission_id, title, description, type,
			CASE
				WHEN status = 'complete' THEN 'complete'
				ELSE phase
			END as status,
			priority, parent_id, assigned_grove_id, context_ref,
			created_at, updated_at, claimed_at, completed_at
		FROM work_orders
	`)
	if err != nil {
		return fmt.Errorf("failed to migrate work_orders data: %w", err)
	}

	// Drop old table
	_, err = db.Exec(`DROP TABLE work_orders`)
	if err != nil {
		return fmt.Errorf("failed to drop old work_orders table: %w", err)
	}

	// Rename new table
	_, err = db.Exec(`ALTER TABLE work_orders_new RENAME TO work_orders`)
	if err != nil {
		return fmt.Errorf("failed to rename work_orders_new: %w", err)
	}

	return nil
}

// migrationV4 adds pinned field to work_orders
func migrationV4(db *sql.DB) error {
	// Add pinned column to work_orders table
	// Default FALSE - work orders are not pinned by default
	_, err := db.Exec(`
		ALTER TABLE work_orders ADD COLUMN pinned INTEGER DEFAULT 0
	`)
	if err != nil {
		return fmt.Errorf("failed to add pinned column: %w", err)
	}

	return nil
}

// migrationV5 adds messages table for agent mail system
func migrationV5(db *sql.DB) error {
	// Create messages table for async agent communication
	// Agents: DEPUTY-{MISSION-ID} and IMP-{GROVE-ID}
	_, err := db.Exec(`
		CREATE TABLE messages (
			id TEXT PRIMARY KEY,
			sender TEXT NOT NULL,
			recipient TEXT NOT NULL,
			subject TEXT NOT NULL,
			body TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			read INTEGER DEFAULT 0,
			mission_id TEXT NOT NULL,
			FOREIGN KEY (mission_id) REFERENCES missions(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create messages table: %w", err)
	}

	// Create indexes for performance
	_, err = db.Exec(`
		CREATE INDEX idx_messages_recipient ON messages(recipient, read);
		CREATE INDEX idx_messages_mission ON messages(mission_id);
		CREATE INDEX idx_messages_timestamp ON messages(timestamp DESC);
	`)
	if err != nil {
		return fmt.Errorf("failed to create messages indexes: %w", err)
	}

	return nil
}

// migrationV6 converts work_orders table to epics, rabbit_holes, and tasks
func migrationV6(db *sql.DB) error {
	// Step 1: Create epics table
	_, err := db.Exec(`
		CREATE TABLE epics (
			id TEXT PRIMARY KEY,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL DEFAULT 'ready' CHECK(status IN ('ready', 'design', 'implement', 'deploy', 'blocked', 'paused', 'complete')),
			priority TEXT CHECK(priority IN ('low', 'medium', 'high')),
			assigned_grove_id TEXT,
			context_ref TEXT,
			pinned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			FOREIGN KEY (mission_id) REFERENCES missions(id),
			FOREIGN KEY (assigned_grove_id) REFERENCES groves(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create epics table: %w", err)
	}

	// Step 2: Create rabbit_holes table
	_, err = db.Exec(`
		CREATE TABLE rabbit_holes (
			id TEXT PRIMARY KEY,
			epic_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL DEFAULT 'ready' CHECK(status IN ('ready', 'design', 'implement', 'deploy', 'blocked', 'paused', 'complete')),
			priority TEXT CHECK(priority IN ('low', 'medium', 'high')),
			pinned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			FOREIGN KEY (epic_id) REFERENCES epics(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create rabbit_holes table: %w", err)
	}

	// Step 3: Create tasks table
	_, err = db.Exec(`
		CREATE TABLE tasks (
			id TEXT PRIMARY KEY,
			epic_id TEXT,
			rabbit_hole_id TEXT,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			type TEXT CHECK(type IN ('research', 'implementation', 'fix', 'documentation', 'maintenance')),
			status TEXT NOT NULL DEFAULT 'ready' CHECK(status IN ('ready', 'design', 'implement', 'deploy', 'blocked', 'paused', 'complete')),
			priority TEXT CHECK(priority IN ('low', 'medium', 'high')),
			assigned_grove_id TEXT,
			context_ref TEXT,
			pinned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			claimed_at DATETIME,
			completed_at DATETIME,
			FOREIGN KEY (epic_id) REFERENCES epics(id) ON DELETE CASCADE,
			FOREIGN KEY (rabbit_hole_id) REFERENCES rabbit_holes(id) ON DELETE CASCADE,
			FOREIGN KEY (mission_id) REFERENCES missions(id),
			FOREIGN KEY (assigned_grove_id) REFERENCES groves(id),
			CHECK ((epic_id IS NOT NULL AND rabbit_hole_id IS NULL) OR
			       (epic_id IS NULL AND rabbit_hole_id IS NOT NULL))
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tasks table: %w", err)
	}

	// Step 4: Migrate top-level work orders (parent_id IS NULL) → epics
	// Convert WO-001 → EPIC-001
	_, err = db.Exec(`
		INSERT INTO epics (
			id, mission_id, title, description, status, priority,
			assigned_grove_id, context_ref, pinned, created_at,
			updated_at, completed_at
		)
		SELECT
			REPLACE(id, 'WO-', 'EPIC-'),
			mission_id,
			title,
			description,
			status,
			priority,
			assigned_grove_id,
			context_ref,
			pinned,
			created_at,
			updated_at,
			completed_at
		FROM work_orders
		WHERE parent_id IS NULL
	`)
	if err != nil {
		return fmt.Errorf("failed to migrate top-level work orders to epics: %w", err)
	}

	// Step 5: Migrate child work orders → tasks
	// Convert WO-010 → TASK-010, parent WO-001 → EPIC-001
	_, err = db.Exec(`
		INSERT INTO tasks (
			id, epic_id, rabbit_hole_id, mission_id, title, description,
			type, status, priority, assigned_grove_id, context_ref,
			pinned, created_at, updated_at, claimed_at, completed_at
		)
		SELECT
			REPLACE(id, 'WO-', 'TASK-'),
			REPLACE(parent_id, 'WO-', 'EPIC-'),
			NULL,
			mission_id,
			title,
			description,
			type,
			status,
			priority,
			assigned_grove_id,
			context_ref,
			pinned,
			created_at,
			updated_at,
			claimed_at,
			completed_at
		FROM work_orders
		WHERE parent_id IS NOT NULL
	`)
	if err != nil {
		return fmt.Errorf("failed to migrate child work orders to tasks: %w", err)
	}

	// Step 6: Create indexes for performance
	_, err = db.Exec(`
		CREATE INDEX idx_epics_mission ON epics(mission_id);
		CREATE INDEX idx_epics_status ON epics(status);
		CREATE INDEX idx_epics_grove ON epics(assigned_grove_id);

		CREATE INDEX idx_rabbit_holes_epic ON rabbit_holes(epic_id);
		CREATE INDEX idx_rabbit_holes_status ON rabbit_holes(status);

		CREATE INDEX idx_tasks_epic ON tasks(epic_id);
		CREATE INDEX idx_tasks_rabbit_hole ON tasks(rabbit_hole_id);
		CREATE INDEX idx_tasks_mission ON tasks(mission_id);
		CREATE INDEX idx_tasks_status ON tasks(status);
		CREATE INDEX idx_tasks_grove ON tasks(assigned_grove_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	// Step 7: Drop old work_orders indexes
	_, err = db.Exec(`
		DROP INDEX IF EXISTS idx_work_orders_mission;
		DROP INDEX IF EXISTS idx_work_orders_status;
		DROP INDEX IF EXISTS idx_work_orders_parent;
	`)
	if err != nil {
		return fmt.Errorf("failed to drop old indexes: %w", err)
	}

	// Step 8: Drop work_orders table
	_, err = db.Exec(`DROP TABLE work_orders`)
	if err != nil {
		return fmt.Errorf("failed to drop work_orders table: %w", err)
	}

	return nil
}

// migrationV7 adds pinned field to missions table
func migrationV7(db *sql.DB) error {
	// Add pinned column to missions table
	_, err := db.Exec(`
		ALTER TABLE missions ADD COLUMN pinned INTEGER DEFAULT 0
	`)
	if err != nil {
		return fmt.Errorf("failed to add pinned column to missions: %w", err)
	}

	return nil
}

// migrationV8 adds tags and task_tags tables for tag system
func migrationV8(db *sql.DB) error {
	// Step 1: Create tags table
	_, err := db.Exec(`
		CREATE TABLE tags (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tags table: %w", err)
	}

	// Step 2: Create task_tags junction table
	_, err = db.Exec(`
		CREATE TABLE task_tags (
			id TEXT PRIMARY KEY,
			task_id TEXT NOT NULL,
			tag_id TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE,
			UNIQUE(task_id, tag_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create task_tags table: %w", err)
	}

	// Step 3: Create indexes for performance
	_, err = db.Exec(`
		CREATE INDEX idx_tags_name ON tags(name);
		CREATE INDEX idx_task_tags_task ON task_tags(task_id);
		CREATE INDEX idx_task_tags_tag ON task_tags(tag_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	// Step 4: Seed 10 glossary tags from TASK-192
	glossaryTags := []struct {
		id          string
		name        string
		description string
	}{
		{"TAG-001", "graphiti", "Graphiti memory system integration and functionality"},
		{"TAG-002", "handoff", "Agent handoff coordination and context transfer"},
		{"TAG-003", "orc-prime", "ORC Prime orchestrator agent capabilities"},
		{"TAG-004", "mission-infra", "Mission workspace and infrastructure setup"},
		{"TAG-005", "desired-state", "Desired state tracking and reconciliation"},
		{"TAG-006", "tech-plans", "Technical planning and design documentation"},
		{"TAG-007", "semantic-epic-system", "9-epic semantic knowledge management"},
		{"TAG-008", "database-schema", "Database migrations and schema changes"},
		{"TAG-009", "orc-summary", "Summary command display and formatting"},
		{"TAG-010", "testing", "Test infrastructure and validation"},
	}

	for _, tag := range glossaryTags {
		_, err = db.Exec(
			"INSERT INTO tags (id, name, description) VALUES (?, ?, ?)",
			tag.id, tag.name, tag.description,
		)
		if err != nil {
			return fmt.Errorf("failed to seed tag %s: %w", tag.name, err)
		}
	}

	return nil
}

// migrationV9 adds 'awaiting_approval' status to tasks
func migrationV9(db *sql.DB) error {
	// SQLite doesn't support ALTER TABLE to modify CHECK constraints
	// Must recreate table with new constraint

	// Step 1: Create new tasks table with updated status CHECK constraint
	_, err := db.Exec(`
		CREATE TABLE tasks_new (
			id TEXT PRIMARY KEY,
			epic_id TEXT,
			rabbit_hole_id TEXT,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			type TEXT CHECK(type IN ('research', 'implementation', 'fix', 'documentation', 'maintenance')),
			status TEXT NOT NULL DEFAULT 'ready' CHECK(status IN ('ready', 'needs_design', 'design', 'implement', 'deploy', 'blocked', 'paused', 'awaiting_approval', 'complete')),
			priority TEXT CHECK(priority IN ('low', 'medium', 'high')),
			assigned_grove_id TEXT,
			context_ref TEXT,
			pinned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			claimed_at DATETIME,
			completed_at DATETIME,
			FOREIGN KEY (epic_id) REFERENCES epics(id) ON DELETE CASCADE,
			FOREIGN KEY (rabbit_hole_id) REFERENCES rabbit_holes(id) ON DELETE CASCADE,
			FOREIGN KEY (mission_id) REFERENCES missions(id),
			FOREIGN KEY (assigned_grove_id) REFERENCES groves(id),
			CHECK ((epic_id IS NOT NULL AND rabbit_hole_id IS NULL) OR
			       (epic_id IS NULL AND rabbit_hole_id IS NOT NULL))
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tasks_new table: %w", err)
	}

	// Step 2: Copy all data from old tasks table
	_, err = db.Exec(`
		INSERT INTO tasks_new SELECT * FROM tasks
	`)
	if err != nil {
		return fmt.Errorf("failed to copy tasks data: %w", err)
	}

	// Step 3: Drop old tasks table
	_, err = db.Exec(`DROP TABLE tasks`)
	if err != nil {
		return fmt.Errorf("failed to drop old tasks table: %w", err)
	}

	// Step 4: Rename new table to tasks
	_, err = db.Exec(`ALTER TABLE tasks_new RENAME TO tasks`)
	if err != nil {
		return fmt.Errorf("failed to rename tasks_new to tasks: %w", err)
	}

	// Step 5: Recreate indexes
	_, err = db.Exec(`
		CREATE INDEX idx_tasks_epic ON tasks(epic_id);
		CREATE INDEX idx_tasks_rabbit_hole ON tasks(rabbit_hole_id);
		CREATE INDEX idx_tasks_mission ON tasks(mission_id);
		CREATE INDEX idx_tasks_status ON tasks(status);
		CREATE INDEX idx_tasks_grove ON tasks(assigned_grove_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to recreate task indexes: %w", err)
	}

	return nil
}

func migrationV10(db *sql.DB) error {
	// SQLite doesn't support ALTER TABLE to modify CHECK constraints
	// Must recreate table with new constraint

	// Step 1: Create new tasks table with updated status CHECK constraint
	_, err := db.Exec(`
		CREATE TABLE tasks_new (
			id TEXT PRIMARY KEY,
			epic_id TEXT,
			rabbit_hole_id TEXT,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			type TEXT CHECK(type IN ('research', 'implementation', 'fix', 'documentation', 'maintenance')),
			status TEXT NOT NULL DEFAULT 'ready' CHECK(status IN ('ready', 'needs_design', 'ready_to_implement', 'design', 'implement', 'deploy', 'blocked', 'paused', 'awaiting_approval', 'complete')),
			priority TEXT CHECK(priority IN ('low', 'medium', 'high')),
			assigned_grove_id TEXT,
			context_ref TEXT,
			pinned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			claimed_at DATETIME,
			completed_at DATETIME,
			FOREIGN KEY (epic_id) REFERENCES epics(id) ON DELETE CASCADE,
			FOREIGN KEY (rabbit_hole_id) REFERENCES rabbit_holes(id) ON DELETE CASCADE,
			FOREIGN KEY (mission_id) REFERENCES missions(id),
			FOREIGN KEY (assigned_grove_id) REFERENCES groves(id),
			CHECK ((epic_id IS NOT NULL AND rabbit_hole_id IS NULL) OR
			       (epic_id IS NULL AND rabbit_hole_id IS NOT NULL))
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tasks_new table: %w", err)
	}

	// Step 2: Copy all data from old tasks table
	_, err = db.Exec(`
		INSERT INTO tasks_new SELECT * FROM tasks
	`)
	if err != nil {
		return fmt.Errorf("failed to copy tasks data: %w", err)
	}

	// Step 3: Drop old tasks table
	_, err = db.Exec(`DROP TABLE tasks`)
	if err != nil {
		return fmt.Errorf("failed to drop old tasks table: %w", err)
	}

	// Step 4: Rename new table to tasks
	_, err = db.Exec(`ALTER TABLE tasks_new RENAME TO tasks`)
	if err != nil {
		return fmt.Errorf("failed to rename tasks_new to tasks: %w", err)
	}

	// Step 5: Recreate indexes
	_, err = db.Exec(`
		CREATE INDEX idx_tasks_epic ON tasks(epic_id);
		CREATE INDEX idx_tasks_rabbit_hole ON tasks(rabbit_hole_id);
		CREATE INDEX idx_tasks_mission ON tasks(mission_id);
		CREATE INDEX idx_tasks_status ON tasks(status);
		CREATE INDEX idx_tasks_grove ON tasks(assigned_grove_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to recreate task indexes: %w", err)
	}

	return nil
}

func migrationV11(db *sql.DB) error {
	// SQLite doesn't support ALTER TABLE to modify CHECK constraints
	// Must recreate table with new constraint
	// Removes 'implement' status from valid statuses

	// Step 1: Create new tasks table with updated status CHECK constraint (no 'implement')
	_, err := db.Exec(`
		CREATE TABLE tasks_new (
			id TEXT PRIMARY KEY,
			epic_id TEXT,
			rabbit_hole_id TEXT,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			type TEXT CHECK(type IN ('research', 'implementation', 'fix', 'documentation', 'maintenance')),
			status TEXT NOT NULL DEFAULT 'ready' CHECK(status IN ('ready', 'needs_design', 'ready_to_implement', 'design', 'deploy', 'blocked', 'paused', 'awaiting_approval', 'complete')),
			priority TEXT CHECK(priority IN ('low', 'medium', 'high')),
			assigned_grove_id TEXT,
			context_ref TEXT,
			pinned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			claimed_at DATETIME,
			completed_at DATETIME,
			FOREIGN KEY (epic_id) REFERENCES epics(id) ON DELETE CASCADE,
			FOREIGN KEY (rabbit_hole_id) REFERENCES rabbit_holes(id) ON DELETE CASCADE,
			FOREIGN KEY (mission_id) REFERENCES missions(id),
			FOREIGN KEY (assigned_grove_id) REFERENCES groves(id),
			CHECK ((epic_id IS NOT NULL AND rabbit_hole_id IS NULL) OR
			       (epic_id IS NULL AND rabbit_hole_id IS NOT NULL))
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tasks_new table: %w", err)
	}

	// Step 2: Copy all data from old tasks table
	_, err = db.Exec(`
		INSERT INTO tasks_new SELECT * FROM tasks
	`)
	if err != nil {
		return fmt.Errorf("failed to copy tasks data: %w", err)
	}

	// Step 3: Drop old tasks table
	_, err = db.Exec(`DROP TABLE tasks`)
	if err != nil {
		return fmt.Errorf("failed to drop old tasks table: %w", err)
	}

	// Step 4: Rename new table to tasks
	_, err = db.Exec(`ALTER TABLE tasks_new RENAME TO tasks`)
	if err != nil {
		return fmt.Errorf("failed to rename tasks_new to tasks: %w", err)
	}

	// Step 5: Recreate indexes
	_, err = db.Exec(`
		CREATE INDEX idx_tasks_epic ON tasks(epic_id);
		CREATE INDEX idx_tasks_rabbit_hole ON tasks(rabbit_hole_id);
		CREATE INDEX idx_tasks_mission ON tasks(mission_id);
		CREATE INDEX idx_tasks_status ON tasks(status);
		CREATE INDEX idx_tasks_grove ON tasks(assigned_grove_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to recreate task indexes: %w", err)
	}

	return nil
}
