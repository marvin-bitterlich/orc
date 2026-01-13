package db

const schemaSQL = `
-- Missions (Strategic work streams)
CREATE TABLE IF NOT EXISTS missions (
	id TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	description TEXT,
	status TEXT NOT NULL CHECK(status IN ('active', 'paused', 'complete', 'archived')) DEFAULT 'active',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	completed_at DATETIME
);

-- Work Orders (Individual tasks) - Flat hierarchy under missions
CREATE TABLE IF NOT EXISTS work_orders (
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
	FOREIGN KEY (parent_id) REFERENCES work_orders(id),
	FOREIGN KEY (assigned_grove_id) REFERENCES groves(id)
);

-- Groves (Physical workspaces) - Mission-level worktrees
CREATE TABLE IF NOT EXISTS groves (
	id TEXT PRIMARY KEY,
	mission_id TEXT NOT NULL,
	name TEXT NOT NULL,
	path TEXT NOT NULL UNIQUE,
	repos TEXT,
	status TEXT NOT NULL CHECK(status IN ('active', 'archived')) DEFAULT 'active',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (mission_id) REFERENCES missions(id)
);

-- Handoffs (Claude-to-Claude context transfer)
CREATE TABLE IF NOT EXISTS handoffs (
	id TEXT PRIMARY KEY,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	handoff_note TEXT NOT NULL,
	active_mission_id TEXT,
	active_work_orders TEXT,
	active_grove_id TEXT,
	todos_snapshot TEXT,
	graphiti_episode_uuid TEXT,
	FOREIGN KEY (active_mission_id) REFERENCES missions(id),
	FOREIGN KEY (active_grove_id) REFERENCES groves(id)
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_missions_status ON missions(status);
CREATE INDEX IF NOT EXISTS idx_work_orders_mission ON work_orders(mission_id);
CREATE INDEX IF NOT EXISTS idx_work_orders_status ON work_orders(status);
CREATE INDEX IF NOT EXISTS idx_work_orders_parent ON work_orders(parent_id);
CREATE INDEX IF NOT EXISTS idx_groves_mission ON groves(mission_id);
CREATE INDEX IF NOT EXISTS idx_groves_status ON groves(status);
CREATE INDEX IF NOT EXISTS idx_handoffs_created ON handoffs(created_at DESC);
`

// InitSchema creates the database schema
func InitSchema() error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	// Check if schema_version table exists to determine if this is a fresh install
	var tableCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_version'").Scan(&tableCount)
	if err != nil {
		return err
	}

	if tableCount == 0 {
		// Fresh install - check if we have old schema tables
		var oldTableCount int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('operations', 'expeditions')").Scan(&oldTableCount)
		if err != nil {
			return err
		}

		if oldTableCount > 0 {
			// Old schema exists - run migrations
			return RunMigrations()
		} else {
			// Completely fresh install - create new schema directly
			_, err = db.Exec(schemaSQL)
			return err
		}
	}

	// schema_version table exists - run any pending migrations
	return RunMigrations()
}
