package db

import (
	"database/sql"
	"fmt"
)

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
	{
		Version: 12,
		Name:    "drop_graphiti_episode_uuid_column",
		Up:      migrationV12,
	},
	{
		Version: 13,
		Name:    "create_container_tables",
		Up:      migrationV13,
	},
	{
		Version: 14,
		Name:    "create_leaf_tables",
		Up:      migrationV14,
	},
	{
		Version: 15,
		Name:    "add_conclave_fks_to_leaves",
		Up:      migrationV15,
	},
	{
		Version: 16,
		Name:    "generalize_tags_to_entity_tags",
		Up:      migrationV16,
	},
	{
		Version: 17,
		Name:    "migrate_epics_to_shipments",
		Up:      migrationV17,
	},
	{
		Version: 18,
		Name:    "add_missing_columns_for_entity_cli",
		Up:      migrationV18,
	},
	{
		Version: 19,
		Name:    "add_paused_status_to_containers",
		Up:      migrationV19,
	},
	{
		Version: 20,
		Name:    "add_paused_status_to_tasks",
		Up:      migrationV20,
	},
	{
		Version: 21,
		Name:    "add_status_to_notes",
		Up:      migrationV21,
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
	// Agents: ORC and IMP-{GROVE-ID}
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

// migrationV12 drops the graphiti_episode_uuid column from handoffs table
// Graphiti integration is being removed to simplify the system
func migrationV12(db *sql.DB) error {
	// SQLite doesn't support DROP COLUMN in older versions
	// Must recreate table without the column

	// Step 1: Create new handoffs table without graphiti_episode_uuid
	_, err := db.Exec(`
		CREATE TABLE handoffs_new (
			id TEXT PRIMARY KEY,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			handoff_note TEXT NOT NULL,
			active_mission_id TEXT,
			active_work_orders TEXT,
			active_grove_id TEXT,
			todos_snapshot TEXT,
			FOREIGN KEY (active_mission_id) REFERENCES missions(id),
			FOREIGN KEY (active_grove_id) REFERENCES groves(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create handoffs_new table: %w", err)
	}

	// Step 2: Copy data (excluding graphiti_episode_uuid)
	_, err = db.Exec(`
		INSERT INTO handoffs_new (
			id, created_at, handoff_note, active_mission_id,
			active_work_orders, active_grove_id, todos_snapshot
		)
		SELECT
			id, created_at, handoff_note, active_mission_id,
			active_work_orders, active_grove_id, todos_snapshot
		FROM handoffs
	`)
	if err != nil {
		return fmt.Errorf("failed to copy handoffs data: %w", err)
	}

	// Step 3: Drop old handoffs table
	_, err = db.Exec(`DROP TABLE handoffs`)
	if err != nil {
		return fmt.Errorf("failed to drop old handoffs table: %w", err)
	}

	// Step 4: Rename new table to handoffs
	_, err = db.Exec(`ALTER TABLE handoffs_new RENAME TO handoffs`)
	if err != nil {
		return fmt.Errorf("failed to rename handoffs_new to handoffs: %w", err)
	}

	// Step 5: Recreate index
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_handoffs_created ON handoffs(created_at DESC)`)
	if err != nil {
		return fmt.Errorf("failed to recreate handoffs index: %w", err)
	}

	return nil
}

// migrationV13 creates container tables: shipments, investigations, conclaves, tomes
func migrationV13(db *sql.DB) error {
	// Step 1: Create shipments table (replaces epics for execution work)
	_, err := db.Exec(`
		CREATE TABLE shipments (
			id TEXT PRIMARY KEY,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL CHECK(status IN ('active', 'complete')) DEFAULT 'active',
			assigned_grove_id TEXT,
			pinned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			FOREIGN KEY (mission_id) REFERENCES missions(id),
			FOREIGN KEY (assigned_grove_id) REFERENCES groves(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create shipments table: %w", err)
	}

	// Step 2: Create investigations table (research mode)
	_, err = db.Exec(`
		CREATE TABLE investigations (
			id TEXT PRIMARY KEY,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL CHECK(status IN ('active', 'complete')) DEFAULT 'active',
			pinned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			FOREIGN KEY (mission_id) REFERENCES missions(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create investigations table: %w", err)
	}

	// Step 3: Create conclaves table (ideation mode, cross-cutting)
	_, err = db.Exec(`
		CREATE TABLE conclaves (
			id TEXT PRIMARY KEY,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL CHECK(status IN ('active', 'complete')) DEFAULT 'active',
			pinned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			FOREIGN KEY (mission_id) REFERENCES missions(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create conclaves table: %w", err)
	}

	// Step 4: Create tomes table (note organization)
	_, err = db.Exec(`
		CREATE TABLE tomes (
			id TEXT PRIMARY KEY,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL CHECK(status IN ('active', 'complete')) DEFAULT 'active',
			pinned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			FOREIGN KEY (mission_id) REFERENCES missions(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tomes table: %w", err)
	}

	// Step 5: Create indexes
	_, err = db.Exec(`
		CREATE INDEX idx_shipments_mission ON shipments(mission_id);
		CREATE INDEX idx_shipments_status ON shipments(status);
		CREATE INDEX idx_shipments_grove ON shipments(assigned_grove_id);
		CREATE INDEX idx_investigations_mission ON investigations(mission_id);
		CREATE INDEX idx_investigations_status ON investigations(status);
		CREATE INDEX idx_conclaves_mission ON conclaves(mission_id);
		CREATE INDEX idx_conclaves_status ON conclaves(status);
		CREATE INDEX idx_tomes_mission ON tomes(mission_id);
		CREATE INDEX idx_tomes_status ON tomes(status);
	`)
	if err != nil {
		return fmt.Errorf("failed to create container indexes: %w", err)
	}

	return nil
}

// migrationV14 creates leaf tables: questions, plans, notes
func migrationV14(db *sql.DB) error {
	// Step 1: Create questions table (lives in investigations)
	_, err := db.Exec(`
		CREATE TABLE questions (
			id TEXT PRIMARY KEY,
			investigation_id TEXT NOT NULL,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL CHECK(status IN ('open', 'answered')) DEFAULT 'open',
			answer TEXT,
			pinned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			answered_at DATETIME,
			promoted_from_id TEXT,
			promoted_from_type TEXT,
			FOREIGN KEY (investigation_id) REFERENCES investigations(id) ON DELETE CASCADE,
			FOREIGN KEY (mission_id) REFERENCES missions(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create questions table: %w", err)
	}

	// Step 2: Create plans table (lives in shipments, one active per shipment)
	_, err = db.Exec(`
		CREATE TABLE plans (
			id TEXT PRIMARY KEY,
			shipment_id TEXT NOT NULL,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			content TEXT,
			status TEXT NOT NULL CHECK(status IN ('draft', 'approved')) DEFAULT 'draft',
			pinned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			approved_at DATETIME,
			FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE,
			FOREIGN KEY (mission_id) REFERENCES missions(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create plans table: %w", err)
	}

	// Create unique constraint for active plan per shipment
	_, err = db.Exec(`
		CREATE UNIQUE INDEX idx_plans_active_shipment ON plans(shipment_id) WHERE status = 'draft'
	`)
	if err != nil {
		return fmt.Errorf("failed to create active plan constraint: %w", err)
	}

	// Step 3: Create notes table (typed, can float or attach to containers)
	_, err = db.Exec(`
		CREATE TABLE notes (
			id TEXT PRIMARY KEY,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			content TEXT,
			type TEXT NOT NULL CHECK(type IN ('learning', 'concern', 'finding', 'frq', 'bug', 'investigation_report')) DEFAULT 'learning',
			status TEXT NOT NULL CHECK(status IN ('open', 'closed')) DEFAULT 'open',
			shipment_id TEXT,
			investigation_id TEXT,
			conclave_id TEXT,
			tome_id TEXT,
			pinned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			closed_at DATETIME,
			promoted_from_id TEXT,
			promoted_from_type TEXT,
			FOREIGN KEY (mission_id) REFERENCES missions(id),
			FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE SET NULL,
			FOREIGN KEY (investigation_id) REFERENCES investigations(id) ON DELETE SET NULL,
			FOREIGN KEY (conclave_id) REFERENCES conclaves(id) ON DELETE SET NULL,
			FOREIGN KEY (tome_id) REFERENCES tomes(id) ON DELETE SET NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create notes table: %w", err)
	}

	// Step 4: Create indexes
	_, err = db.Exec(`
		CREATE INDEX idx_questions_investigation ON questions(investigation_id);
		CREATE INDEX idx_questions_mission ON questions(mission_id);
		CREATE INDEX idx_questions_status ON questions(status);
		CREATE INDEX idx_plans_shipment ON plans(shipment_id);
		CREATE INDEX idx_plans_mission ON plans(mission_id);
		CREATE INDEX idx_plans_status ON plans(status);
		CREATE INDEX idx_notes_mission ON notes(mission_id);
		CREATE INDEX idx_notes_type ON notes(type);
		CREATE INDEX idx_notes_status ON notes(status);
		CREATE INDEX idx_notes_shipment ON notes(shipment_id);
		CREATE INDEX idx_notes_investigation ON notes(investigation_id);
		CREATE INDEX idx_notes_conclave ON notes(conclave_id);
		CREATE INDEX idx_notes_tome ON notes(tome_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to create leaf indexes: %w", err)
	}

	return nil
}

// migrationV15 adds conclave FK columns to tasks, questions, plans for cross-cutting behavior
func migrationV15(db *sql.DB) error {
	// Add conclave_id to tasks (conclave can hold any leaf type)
	_, err := db.Exec(`ALTER TABLE tasks ADD COLUMN conclave_id TEXT REFERENCES conclaves(id) ON DELETE SET NULL`)
	if err != nil {
		return fmt.Errorf("failed to add conclave_id to tasks: %w", err)
	}

	// Add conclave_id to questions
	_, err = db.Exec(`ALTER TABLE questions ADD COLUMN conclave_id TEXT REFERENCES conclaves(id) ON DELETE SET NULL`)
	if err != nil {
		return fmt.Errorf("failed to add conclave_id to questions: %w", err)
	}

	// Add conclave_id to plans
	_, err = db.Exec(`ALTER TABLE plans ADD COLUMN conclave_id TEXT REFERENCES conclaves(id) ON DELETE SET NULL`)
	if err != nil {
		return fmt.Errorf("failed to add conclave_id to plans: %w", err)
	}

	// Create indexes for conclave lookups
	_, err = db.Exec(`
		CREATE INDEX idx_tasks_conclave ON tasks(conclave_id);
		CREATE INDEX idx_questions_conclave ON questions(conclave_id);
		CREATE INDEX idx_plans_conclave ON plans(conclave_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to create conclave indexes: %w", err)
	}

	return nil
}

// migrationV16 generalizes tags from task-only to all entity types
func migrationV16(db *sql.DB) error {
	// Step 1: Create entity_tags table
	_, err := db.Exec(`
		CREATE TABLE entity_tags (
			id TEXT PRIMARY KEY,
			entity_id TEXT NOT NULL,
			entity_type TEXT NOT NULL CHECK(entity_type IN ('task', 'question', 'plan', 'note', 'shipment', 'investigation', 'conclave', 'tome')),
			tag_id TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE,
			UNIQUE(entity_id, entity_type, tag_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create entity_tags table: %w", err)
	}

	// Step 2: Migrate existing task_tags to entity_tags
	_, err = db.Exec(`
		INSERT INTO entity_tags (id, entity_id, entity_type, tag_id, created_at)
		SELECT id, task_id, 'task', tag_id, created_at FROM task_tags
	`)
	if err != nil {
		return fmt.Errorf("failed to migrate task_tags to entity_tags: %w", err)
	}

	// Step 3: Create indexes
	_, err = db.Exec(`
		CREATE INDEX idx_entity_tags_entity ON entity_tags(entity_id, entity_type);
		CREATE INDEX idx_entity_tags_tag ON entity_tags(tag_id);
		CREATE INDEX idx_entity_tags_type ON entity_tags(entity_type);
	`)
	if err != nil {
		return fmt.Errorf("failed to create entity_tags indexes: %w", err)
	}

	// Step 4: Drop old task_tags table
	_, err = db.Exec(`DROP TABLE task_tags`)
	if err != nil {
		return fmt.Errorf("failed to drop task_tags table: %w", err)
	}

	return nil
}

// migrationV17 migrates epics to shipments and removes rabbit_holes
func migrationV17(db *sql.DB) error {
	// Step 1: Copy all epics to shipments (EPIC-NNN → SHIP-NNN)
	_, err := db.Exec(`
		INSERT INTO shipments (id, mission_id, title, description, status, assigned_grove_id, pinned, created_at, updated_at, completed_at)
		SELECT
			REPLACE(id, 'EPIC-', 'SHIP-'),
			mission_id,
			title,
			description,
			CASE WHEN status = 'complete' THEN 'complete' ELSE 'active' END,
			assigned_grove_id,
			pinned,
			created_at,
			updated_at,
			completed_at
		FROM epics
	`)
	if err != nil {
		return fmt.Errorf("failed to copy epics to shipments: %w", err)
	}

	// Step 2: Move tasks from rabbit_holes to their parent epic (now shipment)
	// Need to update both epic_id and clear rabbit_hole_id atomically to satisfy CHECK constraint
	_, err = db.Exec(`
		UPDATE tasks
		SET
			epic_id = (SELECT rh.epic_id FROM rabbit_holes rh WHERE rh.id = tasks.rabbit_hole_id),
			rabbit_hole_id = NULL
		WHERE rabbit_hole_id IS NOT NULL
	`)
	if err != nil {
		return fmt.Errorf("failed to move tasks from rabbit_holes to epics: %w", err)
	}

	// Step 3: Create new tasks table with shipment_id instead of epic_id
	_, err = db.Exec(`
		CREATE TABLE tasks_new (
			id TEXT PRIMARY KEY,
			shipment_id TEXT,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			type TEXT CHECK(type IN ('research', 'implementation', 'fix', 'documentation', 'maintenance')),
			status TEXT NOT NULL CHECK(status IN ('ready', 'in_progress', 'complete')) DEFAULT 'ready',
			priority TEXT CHECK(priority IN ('low', 'medium', 'high')),
			assigned_grove_id TEXT,
			pinned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			claimed_at DATETIME,
			completed_at DATETIME,
			conclave_id TEXT,
			promoted_from_id TEXT,
			promoted_from_type TEXT,
			FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE,
			FOREIGN KEY (mission_id) REFERENCES missions(id),
			FOREIGN KEY (assigned_grove_id) REFERENCES groves(id),
			FOREIGN KEY (conclave_id) REFERENCES conclaves(id) ON DELETE SET NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tasks_new table: %w", err)
	}

	// Step 4: Copy tasks to new table, converting epic_id to shipment_id
	_, err = db.Exec(`
		INSERT INTO tasks_new (id, shipment_id, mission_id, title, description, type, status, priority, assigned_grove_id, pinned, created_at, updated_at, claimed_at, completed_at, conclave_id)
		SELECT
			id,
			REPLACE(epic_id, 'EPIC-', 'SHIP-'),
			mission_id,
			title,
			description,
			type,
			CASE
				WHEN status = 'complete' THEN 'complete'
				WHEN status IN ('implement', 'in_progress', 'design', 'deploy') THEN 'in_progress'
				ELSE 'ready'
			END,
			priority,
			assigned_grove_id,
			pinned,
			created_at,
			updated_at,
			claimed_at,
			completed_at,
			conclave_id
		FROM tasks
	`)
	if err != nil {
		return fmt.Errorf("failed to copy tasks to tasks_new: %w", err)
	}

	// Step 5: Drop old tasks table and rename new one
	_, err = db.Exec(`DROP TABLE tasks`)
	if err != nil {
		return fmt.Errorf("failed to drop old tasks table: %w", err)
	}

	_, err = db.Exec(`ALTER TABLE tasks_new RENAME TO tasks`)
	if err != nil {
		return fmt.Errorf("failed to rename tasks_new to tasks: %w", err)
	}

	// Step 6: Recreate tasks indexes
	_, err = db.Exec(`
		CREATE INDEX idx_tasks_shipment ON tasks(shipment_id);
		CREATE INDEX idx_tasks_mission ON tasks(mission_id);
		CREATE INDEX idx_tasks_status ON tasks(status);
		CREATE INDEX idx_tasks_grove ON tasks(assigned_grove_id);
		CREATE INDEX idx_tasks_conclave ON tasks(conclave_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to recreate tasks indexes: %w", err)
	}

	// Step 7: Drop rabbit_holes table
	_, err = db.Exec(`DROP TABLE rabbit_holes`)
	if err != nil {
		return fmt.Errorf("failed to drop rabbit_holes table: %w", err)
	}

	// Step 8: Drop epics table
	_, err = db.Exec(`DROP TABLE epics`)
	if err != nil {
		return fmt.Errorf("failed to drop epics table: %w", err)
	}

	// Step 9: Drop old epic indexes (if they still exist)
	_, err = db.Exec(`
		DROP INDEX IF EXISTS idx_epics_mission;
		DROP INDEX IF EXISTS idx_epics_status;
		DROP INDEX IF EXISTS idx_epics_grove;
		DROP INDEX IF EXISTS idx_rabbit_holes_epic;
		DROP INDEX IF EXISTS idx_rabbit_holes_status;
		DROP INDEX IF EXISTS idx_tasks_epic;
		DROP INDEX IF EXISTS idx_tasks_rabbit_hole;
	`)
	if err != nil {
		return fmt.Errorf("failed to drop old indexes: %w", err)
	}

	return nil
}

// migrationV18 adds missing columns for the entity CLI commands
func migrationV18(db *sql.DB) error {
	// Add assigned_grove_id to investigations table
	_, err := db.Exec(`ALTER TABLE investigations ADD COLUMN assigned_grove_id TEXT REFERENCES groves(id)`)
	if err != nil {
		return fmt.Errorf("failed to add assigned_grove_id to investigations: %w", err)
	}

	// Add assigned_grove_id to conclaves table
	_, err = db.Exec(`ALTER TABLE conclaves ADD COLUMN assigned_grove_id TEXT REFERENCES groves(id)`)
	if err != nil {
		return fmt.Errorf("failed to add assigned_grove_id to conclaves: %w", err)
	}

	// Add description column to plans table
	_, err = db.Exec(`ALTER TABLE plans ADD COLUMN description TEXT`)
	if err != nil {
		return fmt.Errorf("failed to add description to plans: %w", err)
	}

	// Add promoted_from columns to plans table
	_, err = db.Exec(`ALTER TABLE plans ADD COLUMN promoted_from_id TEXT`)
	if err != nil {
		return fmt.Errorf("failed to add promoted_from_id to plans: %w", err)
	}
	_, err = db.Exec(`ALTER TABLE plans ADD COLUMN promoted_from_type TEXT`)
	if err != nil {
		return fmt.Errorf("failed to add promoted_from_type to plans: %w", err)
	}

	// For questions table, we need to make investigation_id nullable
	// SQLite doesn't support ALTER COLUMN, so we recreate the table
	_, err = db.Exec(`
		CREATE TABLE questions_new (
			id TEXT PRIMARY KEY,
			investigation_id TEXT,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL CHECK(status IN ('open', 'answered')) DEFAULT 'open',
			answer TEXT,
			pinned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			answered_at DATETIME,
			conclave_id TEXT,
			promoted_from_id TEXT,
			promoted_from_type TEXT,
			FOREIGN KEY (investigation_id) REFERENCES investigations(id) ON DELETE CASCADE,
			FOREIGN KEY (mission_id) REFERENCES missions(id),
			FOREIGN KEY (conclave_id) REFERENCES conclaves(id) ON DELETE SET NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create questions_new table: %w", err)
	}

	// Copy existing data
	_, err = db.Exec(`
		INSERT INTO questions_new (id, investigation_id, mission_id, title, description, status, answer, pinned, created_at, updated_at, answered_at, conclave_id, promoted_from_id, promoted_from_type)
		SELECT id, investigation_id, mission_id, title, description, status, answer, pinned, created_at, updated_at, answered_at, conclave_id, promoted_from_id, promoted_from_type FROM questions
	`)
	if err != nil {
		return fmt.Errorf("failed to copy questions data: %w", err)
	}

	// Drop old table
	_, err = db.Exec(`DROP TABLE questions`)
	if err != nil {
		return fmt.Errorf("failed to drop old questions table: %w", err)
	}

	// Rename new table
	_, err = db.Exec(`ALTER TABLE questions_new RENAME TO questions`)
	if err != nil {
		return fmt.Errorf("failed to rename questions_new to questions: %w", err)
	}

	// Recreate indexes
	_, err = db.Exec(`
		CREATE INDEX idx_questions_investigation ON questions(investigation_id);
		CREATE INDEX idx_questions_mission ON questions(mission_id);
		CREATE INDEX idx_questions_status ON questions(status);
		CREATE INDEX idx_questions_conclave ON questions(conclave_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to recreate questions indexes: %w", err)
	}

	// For notes table, make type column nullable (or have a default)
	// Looking at notes migration, type has NOT NULL but we want it optional
	_, err = db.Exec(`
		CREATE TABLE notes_new (
			id TEXT PRIMARY KEY,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			content TEXT,
			type TEXT CHECK(type IN ('learning', 'concern', 'finding', 'frq', 'bug', 'investigation_report')),
			shipment_id TEXT,
			investigation_id TEXT,
			conclave_id TEXT,
			tome_id TEXT,
			pinned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			promoted_from_id TEXT,
			promoted_from_type TEXT,
			FOREIGN KEY (mission_id) REFERENCES missions(id),
			FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE SET NULL,
			FOREIGN KEY (investigation_id) REFERENCES investigations(id) ON DELETE SET NULL,
			FOREIGN KEY (conclave_id) REFERENCES conclaves(id) ON DELETE SET NULL,
			FOREIGN KEY (tome_id) REFERENCES tomes(id) ON DELETE SET NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create notes_new table: %w", err)
	}

	// Copy existing data (dropping status and closed_at as we're simplifying)
	_, err = db.Exec(`
		INSERT INTO notes_new (id, mission_id, title, content, type, shipment_id, investigation_id, conclave_id, tome_id, pinned, created_at, updated_at, promoted_from_id, promoted_from_type)
		SELECT id, mission_id, title, content, type, shipment_id, investigation_id, conclave_id, tome_id, pinned, created_at, updated_at, promoted_from_id, promoted_from_type FROM notes
	`)
	if err != nil {
		return fmt.Errorf("failed to copy notes data: %w", err)
	}

	// Drop old table
	_, err = db.Exec(`DROP TABLE notes`)
	if err != nil {
		return fmt.Errorf("failed to drop old notes table: %w", err)
	}

	// Rename new table
	_, err = db.Exec(`ALTER TABLE notes_new RENAME TO notes`)
	if err != nil {
		return fmt.Errorf("failed to rename notes_new to notes: %w", err)
	}

	// Recreate indexes
	_, err = db.Exec(`
		CREATE INDEX idx_notes_mission ON notes(mission_id);
		CREATE INDEX idx_notes_type ON notes(type);
		CREATE INDEX idx_notes_shipment ON notes(shipment_id);
		CREATE INDEX idx_notes_investigation ON notes(investigation_id);
		CREATE INDEX idx_notes_conclave ON notes(conclave_id);
		CREATE INDEX idx_notes_tome ON notes(tome_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to recreate notes indexes: %w", err)
	}

	// For plans table, we also need to make shipment_id nullable
	_, err = db.Exec(`
		CREATE TABLE plans_new (
			id TEXT PRIMARY KEY,
			shipment_id TEXT,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			content TEXT,
			status TEXT NOT NULL CHECK(status IN ('draft', 'approved')) DEFAULT 'draft',
			pinned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			approved_at DATETIME,
			conclave_id TEXT,
			promoted_from_id TEXT,
			promoted_from_type TEXT,
			FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE,
			FOREIGN KEY (mission_id) REFERENCES missions(id),
			FOREIGN KEY (conclave_id) REFERENCES conclaves(id) ON DELETE SET NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create plans_new table: %w", err)
	}

	// Copy existing data
	_, err = db.Exec(`
		INSERT INTO plans_new (id, shipment_id, mission_id, title, content, status, pinned, created_at, updated_at, approved_at, conclave_id)
		SELECT id, shipment_id, mission_id, title, content, status, pinned, created_at, updated_at, approved_at, conclave_id FROM plans
	`)
	if err != nil {
		return fmt.Errorf("failed to copy plans data: %w", err)
	}

	// Drop old unique index first if it exists
	_, _ = db.Exec(`DROP INDEX IF EXISTS idx_plans_active_shipment`)

	// Drop old table
	_, err = db.Exec(`DROP TABLE plans`)
	if err != nil {
		return fmt.Errorf("failed to drop old plans table: %w", err)
	}

	// Rename new table
	_, err = db.Exec(`ALTER TABLE plans_new RENAME TO plans`)
	if err != nil {
		return fmt.Errorf("failed to rename plans_new to plans: %w", err)
	}

	// Recreate indexes
	_, err = db.Exec(`
		CREATE INDEX idx_plans_shipment ON plans(shipment_id);
		CREATE INDEX idx_plans_mission ON plans(mission_id);
		CREATE INDEX idx_plans_status ON plans(status);
		CREATE INDEX idx_plans_conclave ON plans(conclave_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to recreate plans indexes: %w", err)
	}

	// Create unique constraint for active plan per shipment (if shipment_id is not null)
	_, err = db.Exec(`
		CREATE UNIQUE INDEX idx_plans_active_shipment ON plans(shipment_id) WHERE status = 'draft' AND shipment_id IS NOT NULL
	`)
	if err != nil {
		return fmt.Errorf("failed to create active plan constraint: %w", err)
	}

	return nil
}

// migrationV19 adds 'paused' status to container tables (shipments, investigations, conclaves, tomes)
func migrationV19(db *sql.DB) error {
	// Shipments table - recreate with updated CHECK constraint
	_, err := db.Exec(`
		CREATE TABLE shipments_new (
			id TEXT PRIMARY KEY,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL CHECK(status IN ('active', 'paused', 'complete')) DEFAULT 'active',
			assigned_grove_id TEXT,
			pinned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			FOREIGN KEY (mission_id) REFERENCES missions(id),
			FOREIGN KEY (assigned_grove_id) REFERENCES groves(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create shipments_new table: %w", err)
	}

	_, err = db.Exec(`
		INSERT INTO shipments_new (id, mission_id, title, description, status, assigned_grove_id, pinned, created_at, updated_at, completed_at)
		SELECT id, mission_id, title, description, status, assigned_grove_id, pinned, created_at, updated_at, completed_at FROM shipments
	`)
	if err != nil {
		return fmt.Errorf("failed to copy shipments data: %w", err)
	}

	_, err = db.Exec(`DROP TABLE shipments`)
	if err != nil {
		return fmt.Errorf("failed to drop old shipments table: %w", err)
	}

	_, err = db.Exec(`ALTER TABLE shipments_new RENAME TO shipments`)
	if err != nil {
		return fmt.Errorf("failed to rename shipments_new: %w", err)
	}

	_, err = db.Exec(`
		CREATE INDEX idx_shipments_mission ON shipments(mission_id);
		CREATE INDEX idx_shipments_status ON shipments(status);
		CREATE INDEX idx_shipments_grove ON shipments(assigned_grove_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to recreate shipments indexes: %w", err)
	}

	// Investigations table - recreate with updated CHECK constraint
	_, err = db.Exec(`
		CREATE TABLE investigations_new (
			id TEXT PRIMARY KEY,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL CHECK(status IN ('active', 'paused', 'complete')) DEFAULT 'active',
			pinned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			assigned_grove_id TEXT REFERENCES groves(id),
			FOREIGN KEY (mission_id) REFERENCES missions(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create investigations_new table: %w", err)
	}

	_, err = db.Exec(`
		INSERT INTO investigations_new (id, mission_id, title, description, status, pinned, created_at, updated_at, completed_at, assigned_grove_id)
		SELECT id, mission_id, title, description, status, pinned, created_at, updated_at, completed_at, assigned_grove_id FROM investigations
	`)
	if err != nil {
		return fmt.Errorf("failed to copy investigations data: %w", err)
	}

	_, err = db.Exec(`DROP TABLE investigations`)
	if err != nil {
		return fmt.Errorf("failed to drop old investigations table: %w", err)
	}

	_, err = db.Exec(`ALTER TABLE investigations_new RENAME TO investigations`)
	if err != nil {
		return fmt.Errorf("failed to rename investigations_new: %w", err)
	}

	_, err = db.Exec(`
		CREATE INDEX idx_investigations_mission ON investigations(mission_id);
		CREATE INDEX idx_investigations_status ON investigations(status);
	`)
	if err != nil {
		return fmt.Errorf("failed to recreate investigations indexes: %w", err)
	}

	// Conclaves table - recreate with updated CHECK constraint
	_, err = db.Exec(`
		CREATE TABLE conclaves_new (
			id TEXT PRIMARY KEY,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL CHECK(status IN ('active', 'paused', 'complete')) DEFAULT 'active',
			pinned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			assigned_grove_id TEXT REFERENCES groves(id),
			FOREIGN KEY (mission_id) REFERENCES missions(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create conclaves_new table: %w", err)
	}

	_, err = db.Exec(`
		INSERT INTO conclaves_new (id, mission_id, title, description, status, pinned, created_at, updated_at, completed_at, assigned_grove_id)
		SELECT id, mission_id, title, description, status, pinned, created_at, updated_at, completed_at, assigned_grove_id FROM conclaves
	`)
	if err != nil {
		return fmt.Errorf("failed to copy conclaves data: %w", err)
	}

	_, err = db.Exec(`DROP TABLE conclaves`)
	if err != nil {
		return fmt.Errorf("failed to drop old conclaves table: %w", err)
	}

	_, err = db.Exec(`ALTER TABLE conclaves_new RENAME TO conclaves`)
	if err != nil {
		return fmt.Errorf("failed to rename conclaves_new: %w", err)
	}

	_, err = db.Exec(`
		CREATE INDEX idx_conclaves_mission ON conclaves(mission_id);
		CREATE INDEX idx_conclaves_status ON conclaves(status);
	`)
	if err != nil {
		return fmt.Errorf("failed to recreate conclaves indexes: %w", err)
	}

	// Tomes table - recreate with updated CHECK constraint
	_, err = db.Exec(`
		CREATE TABLE tomes_new (
			id TEXT PRIMARY KEY,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL CHECK(status IN ('active', 'paused', 'complete')) DEFAULT 'active',
			pinned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			FOREIGN KEY (mission_id) REFERENCES missions(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tomes_new table: %w", err)
	}

	_, err = db.Exec(`
		INSERT INTO tomes_new (id, mission_id, title, description, status, pinned, created_at, updated_at, completed_at)
		SELECT id, mission_id, title, description, status, pinned, created_at, updated_at, completed_at FROM tomes
	`)
	if err != nil {
		return fmt.Errorf("failed to copy tomes data: %w", err)
	}

	_, err = db.Exec(`DROP TABLE tomes`)
	if err != nil {
		return fmt.Errorf("failed to drop old tomes table: %w", err)
	}

	_, err = db.Exec(`ALTER TABLE tomes_new RENAME TO tomes`)
	if err != nil {
		return fmt.Errorf("failed to rename tomes_new: %w", err)
	}

	_, err = db.Exec(`
		CREATE INDEX idx_tomes_mission ON tomes(mission_id);
		CREATE INDEX idx_tomes_status ON tomes(status);
	`)
	if err != nil {
		return fmt.Errorf("failed to recreate tomes indexes: %w", err)
	}

	return nil
}

// migrationV20 adds 'paused' status to tasks
func migrationV20(db *sql.DB) error {
	// Tasks table - recreate with updated CHECK constraint
	_, err := db.Exec(`
		CREATE TABLE tasks_new (
			id TEXT PRIMARY KEY,
			shipment_id TEXT,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			type TEXT CHECK(type IN ('research', 'implementation', 'fix', 'documentation', 'maintenance')),
			status TEXT NOT NULL CHECK(status IN ('ready', 'in_progress', 'paused', 'complete')) DEFAULT 'ready',
			priority TEXT CHECK(priority IN ('low', 'medium', 'high')),
			assigned_grove_id TEXT,
			pinned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			claimed_at DATETIME,
			completed_at DATETIME,
			conclave_id TEXT,
			promoted_from_id TEXT,
			promoted_from_type TEXT,
			FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE,
			FOREIGN KEY (mission_id) REFERENCES missions(id),
			FOREIGN KEY (assigned_grove_id) REFERENCES groves(id),
			FOREIGN KEY (conclave_id) REFERENCES conclaves(id) ON DELETE SET NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tasks_new table: %w", err)
	}

	_, err = db.Exec(`
		INSERT INTO tasks_new (id, shipment_id, mission_id, title, description, type, status, priority, assigned_grove_id, pinned, created_at, updated_at, claimed_at, completed_at, conclave_id, promoted_from_id, promoted_from_type)
		SELECT id, shipment_id, mission_id, title, description, type, status, priority, assigned_grove_id, pinned, created_at, updated_at, claimed_at, completed_at, conclave_id, promoted_from_id, promoted_from_type FROM tasks
	`)
	if err != nil {
		return fmt.Errorf("failed to copy tasks data: %w", err)
	}

	_, err = db.Exec(`DROP TABLE tasks`)
	if err != nil {
		return fmt.Errorf("failed to drop old tasks table: %w", err)
	}

	_, err = db.Exec(`ALTER TABLE tasks_new RENAME TO tasks`)
	if err != nil {
		return fmt.Errorf("failed to rename tasks_new: %w", err)
	}

	_, err = db.Exec(`
		CREATE INDEX idx_tasks_shipment ON tasks(shipment_id);
		CREATE INDEX idx_tasks_mission ON tasks(mission_id);
		CREATE INDEX idx_tasks_status ON tasks(status);
		CREATE INDEX idx_tasks_grove ON tasks(assigned_grove_id);
		CREATE INDEX idx_tasks_conclave ON tasks(conclave_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to recreate tasks indexes: %w", err)
	}

	return nil
}

// migrationV21 adds status field to notes (open/closed)
func migrationV21(db *sql.DB) error {
	// Notes table - recreate with status column
	_, err := db.Exec(`
		CREATE TABLE notes_new (
			id TEXT PRIMARY KEY,
			mission_id TEXT NOT NULL,
			title TEXT NOT NULL,
			content TEXT,
			type TEXT CHECK(type IN ('learning', 'concern', 'finding', 'frq', 'bug', 'investigation_report')),
			status TEXT NOT NULL CHECK(status IN ('open', 'closed')) DEFAULT 'open',
			shipment_id TEXT,
			investigation_id TEXT,
			conclave_id TEXT,
			tome_id TEXT,
			pinned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			closed_at DATETIME,
			promoted_from_id TEXT,
			promoted_from_type TEXT,
			FOREIGN KEY (mission_id) REFERENCES missions(id),
			FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE SET NULL,
			FOREIGN KEY (investigation_id) REFERENCES investigations(id) ON DELETE SET NULL,
			FOREIGN KEY (conclave_id) REFERENCES conclaves(id) ON DELETE SET NULL,
			FOREIGN KEY (tome_id) REFERENCES tomes(id) ON DELETE SET NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create notes_new table: %w", err)
	}

	// Copy existing data (all existing notes default to 'open' status)
	_, err = db.Exec(`
		INSERT INTO notes_new (id, mission_id, title, content, type, status, shipment_id, investigation_id, conclave_id, tome_id, pinned, created_at, updated_at, promoted_from_id, promoted_from_type)
		SELECT id, mission_id, title, content, type, 'open', shipment_id, investigation_id, conclave_id, tome_id, pinned, created_at, updated_at, promoted_from_id, promoted_from_type FROM notes
	`)
	if err != nil {
		return fmt.Errorf("failed to copy notes data: %w", err)
	}

	// Drop old table
	_, err = db.Exec(`DROP TABLE notes`)
	if err != nil {
		return fmt.Errorf("failed to drop old notes table: %w", err)
	}

	// Rename new table
	_, err = db.Exec(`ALTER TABLE notes_new RENAME TO notes`)
	if err != nil {
		return fmt.Errorf("failed to rename notes_new to notes: %w", err)
	}

	// Recreate indexes
	_, err = db.Exec(`
		CREATE INDEX idx_notes_mission ON notes(mission_id);
		CREATE INDEX idx_notes_type ON notes(type);
		CREATE INDEX idx_notes_status ON notes(status);
		CREATE INDEX idx_notes_shipment ON notes(shipment_id);
		CREATE INDEX idx_notes_investigation ON notes(investigation_id);
		CREATE INDEX idx_notes_conclave ON notes(conclave_id);
		CREATE INDEX idx_notes_tome ON notes(tome_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to recreate notes indexes: %w", err)
	}

	return nil
}
