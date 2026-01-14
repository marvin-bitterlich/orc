package db

import (
	"database/sql"
	"fmt"
)

// schemaVersion tracks the current schema version
const currentSchemaVersion = 3

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
