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

-- Operations (Tactical groupings of work)
CREATE TABLE IF NOT EXISTS operations (
	id TEXT PRIMARY KEY,
	mission_id TEXT NOT NULL,
	title TEXT NOT NULL,
	description TEXT,
	status TEXT NOT NULL CHECK(status IN ('backlog', 'active', 'complete', 'cancelled')) DEFAULT 'backlog',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	completed_at DATETIME,
	FOREIGN KEY (mission_id) REFERENCES missions(id)
);

-- Work Orders (Individual tasks)
CREATE TABLE IF NOT EXISTS work_orders (
	id TEXT PRIMARY KEY,
	operation_id TEXT NOT NULL,
	title TEXT NOT NULL,
	description TEXT,
	status TEXT NOT NULL CHECK(status IN ('backlog', 'next', 'in_progress', 'complete')) DEFAULT 'backlog',
	assigned_imp TEXT,
	context_ref TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	claimed_at DATETIME,
	completed_at DATETIME,
	FOREIGN KEY (operation_id) REFERENCES operations(id)
);

-- Expeditions (WHO/HOW coordination)
CREATE TABLE IF NOT EXISTS expeditions (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	work_order_id TEXT,
	assigned_imp TEXT,
	status TEXT NOT NULL CHECK(status IN ('planning', 'active', 'paused', 'complete')) DEFAULT 'planning',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (work_order_id) REFERENCES work_orders(id)
);

-- Groves (Physical workspaces)
CREATE TABLE IF NOT EXISTS groves (
	id TEXT PRIMARY KEY,
	path TEXT NOT NULL UNIQUE,
	repos TEXT,
	expedition_id TEXT,
	status TEXT NOT NULL CHECK(status IN ('active', 'idle', 'archived')) DEFAULT 'active',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (expedition_id) REFERENCES expeditions(id)
);

-- Dependencies (Blocking relationships)
CREATE TABLE IF NOT EXISTS dependencies (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	source_expedition_id TEXT NOT NULL,
	blocks_expedition_id TEXT NOT NULL,
	reason TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (source_expedition_id) REFERENCES expeditions(id),
	FOREIGN KEY (blocks_expedition_id) REFERENCES expeditions(id)
);

-- Plans (Integration with Graphiti)
CREATE TABLE IF NOT EXISTS plans (
	id TEXT PRIMARY KEY,
	expedition_id TEXT NOT NULL,
	title TEXT NOT NULL,
	status TEXT NOT NULL CHECK(status IN ('draft', 'approved', 'rejected', 'implemented')) DEFAULT 'draft',
	graphiti_episode_uuid TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	approved_at DATETIME,
	FOREIGN KEY (expedition_id) REFERENCES expeditions(id)
);

-- Handoffs (Claude-to-Claude context transfer)
CREATE TABLE IF NOT EXISTS handoffs (
	id TEXT PRIMARY KEY,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	handoff_note TEXT NOT NULL,
	active_mission_id TEXT,
	active_operation_id TEXT,
	active_work_order_id TEXT,
	active_expedition_id TEXT,
	todos_snapshot TEXT,
	graphiti_episode_uuid TEXT,
	FOREIGN KEY (active_mission_id) REFERENCES missions(id),
	FOREIGN KEY (active_operation_id) REFERENCES operations(id),
	FOREIGN KEY (active_work_order_id) REFERENCES work_orders(id),
	FOREIGN KEY (active_expedition_id) REFERENCES expeditions(id)
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_missions_status ON missions(status);
CREATE INDEX IF NOT EXISTS idx_operations_mission ON operations(mission_id);
CREATE INDEX IF NOT EXISTS idx_operations_status ON operations(status);
CREATE INDEX IF NOT EXISTS idx_work_orders_operation ON work_orders(operation_id);
CREATE INDEX IF NOT EXISTS idx_work_orders_status ON work_orders(status);
CREATE INDEX IF NOT EXISTS idx_expeditions_status ON expeditions(status);
CREATE INDEX IF NOT EXISTS idx_groves_status ON groves(status);
CREATE INDEX IF NOT EXISTS idx_groves_expedition ON groves(expedition_id);
CREATE INDEX IF NOT EXISTS idx_plans_expedition ON plans(expedition_id);
CREATE INDEX IF NOT EXISTS idx_handoffs_created ON handoffs(created_at DESC);
`

// InitSchema creates the database schema
func InitSchema() error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	_, err = db.Exec(schemaSQL)
	return err
}
