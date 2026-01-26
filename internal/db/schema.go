package db

// SchemaSQL is the complete modern schema for fresh ORC installs.
// This schema reflects the current state after all migrations.
//
// # Schema Drift Protection
//
// This is the SINGLE SOURCE OF TRUTH for the database schema. All tests use
// this schema via GetSchemaSQL(), which provides two layers of protection:
//
//  1. No hardcoded schemas: `make schema-check` fails if any test file contains
//     CREATE TABLE statements. Tests must use db.GetSchemaSQL() instead.
//
//  2. Immediate failure on drift: If repository code references a column that
//     doesn't exist in this schema, tests fail immediately with "no such column".
//     This catches drift at development time, not production.
//
// # Keeping Schema in Sync
//
// IMPORTANT: Keep this in sync with migrations. Use Atlas to verify:
//
//	atlas schema diff --from "sqlite:///$HOME/.orc/orc.db" --to "sqlite:///tmp/fresh.db" --dev-url "sqlite://dev?mode=memory"
//
// When adding new columns or tables:
//  1. Add migration in internal/db/migrations/
//  2. Update SchemaSQL here
//  3. Run `make test` to verify alignment
const SchemaSQL = `
-- Tags (generic tagging system)
CREATE TABLE IF NOT EXISTS tags (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	description TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS entity_tags (
	id TEXT PRIMARY KEY,
	entity_id TEXT NOT NULL,
	entity_type TEXT NOT NULL CHECK(entity_type IN ('task', 'plan', 'note', 'shipment', 'investigation', 'conclave', 'tome')),
	tag_id TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE,
	UNIQUE(entity_id, entity_type, tag_id)
);

-- Repos (Repository configurations)
CREATE TABLE IF NOT EXISTS repos (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	url TEXT,
	local_path TEXT,
	default_branch TEXT DEFAULT 'main',
	status TEXT NOT NULL CHECK(status IN ('active', 'archived')) DEFAULT 'active',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Factories (TMux sessions - runtime environments)
CREATE TABLE IF NOT EXISTS factories (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	status TEXT NOT NULL CHECK(status IN ('active', 'archived')) DEFAULT 'active',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Workshops (TMux sessions - runtime environments within a factory)
CREATE TABLE IF NOT EXISTS workshops (
	id TEXT PRIMARY KEY,
	factory_id TEXT NOT NULL,
	name TEXT NOT NULL,
	status TEXT NOT NULL CHECK(status IN ('active', 'archived')) DEFAULT 'active',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (factory_id) REFERENCES factories(id)
);

-- Workbenches (Git worktrees within a workshop)
CREATE TABLE IF NOT EXISTS workbenches (
	id TEXT PRIMARY KEY,
	workshop_id TEXT NOT NULL,
	name TEXT NOT NULL,
	path TEXT NOT NULL UNIQUE,
	repo_id TEXT,
	status TEXT NOT NULL CHECK(status IN ('active', 'archived')) DEFAULT 'active',
	home_branch TEXT,
	current_branch TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (workshop_id) REFERENCES workshops(id),
	FOREIGN KEY (repo_id) REFERENCES repos(id)
);

-- Commissions (Tracks of work - what you're working on)
CREATE TABLE IF NOT EXISTS commissions (
	id TEXT PRIMARY KEY,
	factory_id TEXT,
	title TEXT NOT NULL,
	description TEXT,
	status TEXT NOT NULL CHECK(status IN ('initial', 'active', 'paused', 'complete', 'archived', 'deleted')) DEFAULT 'initial',
	pinned INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	started_at DATETIME,
	completed_at DATETIME,
	updated_at DATETIME,
	FOREIGN KEY (factory_id) REFERENCES factories(id)
);

-- Messages (Agent mail system)
CREATE TABLE IF NOT EXISTS messages (
	id TEXT PRIMARY KEY,
	sender TEXT NOT NULL,
	recipient TEXT NOT NULL,
	subject TEXT NOT NULL,
	body TEXT NOT NULL,
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	read INTEGER DEFAULT 0,
	commission_id TEXT NOT NULL,
	FOREIGN KEY (commission_id) REFERENCES commissions(id)
);

-- Shipments (Work containers)
CREATE TABLE IF NOT EXISTS shipments (
	id TEXT PRIMARY KEY,
	commission_id TEXT NOT NULL,
	title TEXT NOT NULL,
	description TEXT,
	status TEXT NOT NULL CHECK(status IN ('active', 'in_progress', 'paused', 'complete')) DEFAULT 'active',
	assigned_workbench_id TEXT,
	repo_id TEXT,
	branch TEXT,
	pinned INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	completed_at DATETIME,
	FOREIGN KEY (commission_id) REFERENCES commissions(id),
	FOREIGN KEY (assigned_workbench_id) REFERENCES workbenches(id),
	FOREIGN KEY (repo_id) REFERENCES repos(id)
);

-- Investigations (Research containers)
CREATE TABLE IF NOT EXISTS investigations (
	id TEXT PRIMARY KEY,
	commission_id TEXT NOT NULL,
	conclave_id TEXT,
	title TEXT NOT NULL,
	description TEXT,
	status TEXT NOT NULL CHECK(status IN ('active', 'in_progress', 'resolved', 'complete')) DEFAULT 'active',
	assigned_workbench_id TEXT,
	pinned INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	completed_at DATETIME,
	FOREIGN KEY (commission_id) REFERENCES commissions(id),
	FOREIGN KEY (conclave_id) REFERENCES conclaves(id) ON DELETE SET NULL,
	FOREIGN KEY (assigned_workbench_id) REFERENCES workbenches(id)
);

-- Tomes (Knowledge containers)
CREATE TABLE IF NOT EXISTS tomes (
	id TEXT PRIMARY KEY,
	commission_id TEXT NOT NULL,
	conclave_id TEXT,
	title TEXT NOT NULL,
	description TEXT,
	status TEXT NOT NULL CHECK(status IN ('open', 'closed')) DEFAULT 'open',
	assigned_workbench_id TEXT,
	pinned INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	closed_at DATETIME,
	FOREIGN KEY (commission_id) REFERENCES commissions(id),
	FOREIGN KEY (conclave_id) REFERENCES conclaves(id) ON DELETE SET NULL,
	FOREIGN KEY (assigned_workbench_id) REFERENCES workbenches(id)
);

-- Tasks (Atomic units of work)
CREATE TABLE IF NOT EXISTS tasks (
	id TEXT PRIMARY KEY,
	shipment_id TEXT,
	commission_id TEXT NOT NULL,
	investigation_id TEXT,
	tome_id TEXT,
	conclave_id TEXT,
	title TEXT NOT NULL,
	description TEXT,
	type TEXT CHECK(type IN ('research', 'implementation', 'fix', 'documentation', 'maintenance')),
	status TEXT NOT NULL CHECK(status IN ('ready', 'in_progress', 'blocked', 'paused', 'complete')) DEFAULT 'ready',
	priority TEXT CHECK(priority IN ('low', 'medium', 'high')),
	assigned_workbench_id TEXT,
	context_ref TEXT,
	pinned INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	claimed_at DATETIME,
	completed_at DATETIME,
	FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE,
	FOREIGN KEY (commission_id) REFERENCES commissions(id),
	FOREIGN KEY (investigation_id) REFERENCES investigations(id) ON DELETE SET NULL,
	FOREIGN KEY (tome_id) REFERENCES tomes(id) ON DELETE SET NULL,
	FOREIGN KEY (conclave_id) REFERENCES conclaves(id) ON DELETE SET NULL,
	FOREIGN KEY (assigned_workbench_id) REFERENCES workbenches(id)
);

-- Operations (Legacy work units - migrated from missions era)
CREATE TABLE IF NOT EXISTS operations (
	id TEXT PRIMARY KEY,
	commission_id TEXT NOT NULL,
	title TEXT NOT NULL,
	description TEXT,
	status TEXT NOT NULL CHECK(status IN ('ready', 'in_progress', 'complete')) DEFAULT 'ready',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	completed_at DATETIME,
	FOREIGN KEY (commission_id) REFERENCES commissions(id)
);

-- Handoffs (Claude-to-Claude context transfer)
CREATE TABLE IF NOT EXISTS handoffs (
	id TEXT PRIMARY KEY,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	handoff_note TEXT NOT NULL,
	active_commission_id TEXT,
	active_workbench_id TEXT,
	todos_snapshot TEXT,
	FOREIGN KEY (active_commission_id) REFERENCES commissions(id)
);

-- PRs (Pull requests)
CREATE TABLE IF NOT EXISTS prs (
	id TEXT PRIMARY KEY,
	shipment_id TEXT NOT NULL UNIQUE,
	repo_id TEXT NOT NULL,
	commission_id TEXT NOT NULL,
	number INTEGER,
	title TEXT NOT NULL,
	description TEXT,
	branch TEXT NOT NULL,
	target_branch TEXT,
	url TEXT,
	status TEXT NOT NULL CHECK(status IN ('draft', 'open', 'approved', 'merged', 'closed')) DEFAULT 'open',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	merged_at DATETIME,
	closed_at DATETIME,
	FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE,
	FOREIGN KEY (repo_id) REFERENCES repos(id),
	FOREIGN KEY (commission_id) REFERENCES commissions(id)
);

-- Plans (Implementation plans)
CREATE TABLE IF NOT EXISTS plans (
	id TEXT PRIMARY KEY,
	commission_id TEXT NOT NULL,
	shipment_id TEXT,
	cycle_id TEXT,
	title TEXT NOT NULL,
	description TEXT,
	content TEXT,
	status TEXT NOT NULL CHECK(status IN ('draft', 'approved', 'superseded')) DEFAULT 'draft',
	pinned INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	approved_at DATETIME,
	conclave_id TEXT,
	promoted_from_id TEXT,
	promoted_from_type TEXT,
	FOREIGN KEY (commission_id) REFERENCES commissions(id),
	FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE SET NULL,
	FOREIGN KEY (cycle_id) REFERENCES cycles(id) ON DELETE SET NULL,
	FOREIGN KEY (conclave_id) REFERENCES conclaves(id) ON DELETE SET NULL
);

-- Conclaves (Decision containers)
CREATE TABLE IF NOT EXISTS conclaves (
	id TEXT PRIMARY KEY,
	commission_id TEXT NOT NULL,
	shipment_id TEXT,
	title TEXT NOT NULL,
	description TEXT,
	status TEXT NOT NULL CHECK(status IN ('open', 'paused', 'closed')) DEFAULT 'open',
	decision TEXT,
	pinned INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	decided_at DATETIME,
	FOREIGN KEY (commission_id) REFERENCES commissions(id),
	FOREIGN KEY (shipment_id) REFERENCES shipments(id)
);

-- Notes (Observations and learnings)
CREATE TABLE IF NOT EXISTS notes (
	id TEXT PRIMARY KEY,
	commission_id TEXT NOT NULL,
	shipment_id TEXT,
	investigation_id TEXT,
	conclave_id TEXT,
	tome_id TEXT,
	title TEXT NOT NULL,
	content TEXT,
	type TEXT,
	status TEXT NOT NULL CHECK(status IN ('open', 'resolved', 'closed')) DEFAULT 'open',
	pinned INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	closed_at DATETIME,
	promoted_from_id TEXT,
	promoted_from_type TEXT,
	FOREIGN KEY (commission_id) REFERENCES commissions(id),
	FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE SET NULL,
	FOREIGN KEY (investigation_id) REFERENCES investigations(id) ON DELETE SET NULL,
	FOREIGN KEY (conclave_id) REFERENCES conclaves(id) ON DELETE SET NULL,
	FOREIGN KEY (tome_id) REFERENCES tomes(id) ON DELETE SET NULL
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
CREATE INDEX IF NOT EXISTS idx_entity_tags_entity ON entity_tags(entity_id, entity_type);
CREATE INDEX IF NOT EXISTS idx_entity_tags_tag ON entity_tags(tag_id);
CREATE INDEX IF NOT EXISTS idx_entity_tags_type ON entity_tags(entity_type);
CREATE INDEX IF NOT EXISTS idx_repos_name ON repos(name);
CREATE INDEX IF NOT EXISTS idx_repos_status ON repos(status);
CREATE INDEX IF NOT EXISTS idx_factories_name ON factories(name);
CREATE INDEX IF NOT EXISTS idx_factories_status ON factories(status);
CREATE INDEX IF NOT EXISTS idx_workshops_factory ON workshops(factory_id);
CREATE INDEX IF NOT EXISTS idx_workshops_status ON workshops(status);
CREATE INDEX IF NOT EXISTS idx_workbenches_workshop ON workbenches(workshop_id);
CREATE INDEX IF NOT EXISTS idx_workbenches_status ON workbenches(status);
CREATE INDEX IF NOT EXISTS idx_workbenches_repo ON workbenches(repo_id);
CREATE INDEX IF NOT EXISTS idx_commissions_factory ON commissions(factory_id);
CREATE INDEX IF NOT EXISTS idx_commissions_status ON commissions(status);
CREATE INDEX IF NOT EXISTS idx_messages_recipient ON messages(recipient, read);
CREATE INDEX IF NOT EXISTS idx_messages_commission ON messages(commission_id);
CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_shipments_commission ON shipments(commission_id);
CREATE INDEX IF NOT EXISTS idx_shipments_status ON shipments(status);
CREATE INDEX IF NOT EXISTS idx_shipments_workbench ON shipments(assigned_workbench_id);
CREATE INDEX IF NOT EXISTS idx_investigations_commission ON investigations(commission_id);
CREATE INDEX IF NOT EXISTS idx_investigations_status ON investigations(status);
CREATE INDEX IF NOT EXISTS idx_investigations_conclave ON investigations(conclave_id);
CREATE INDEX IF NOT EXISTS idx_tomes_commission ON tomes(commission_id);
CREATE INDEX IF NOT EXISTS idx_tomes_conclave ON tomes(conclave_id);
CREATE INDEX IF NOT EXISTS idx_tasks_shipment ON tasks(shipment_id);
CREATE INDEX IF NOT EXISTS idx_tasks_commission ON tasks(commission_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_workbench ON tasks(assigned_workbench_id);
CREATE INDEX IF NOT EXISTS idx_tasks_investigation ON tasks(investigation_id);
CREATE INDEX IF NOT EXISTS idx_tasks_tome ON tasks(tome_id);
CREATE INDEX IF NOT EXISTS idx_tasks_conclave ON tasks(conclave_id);
CREATE INDEX IF NOT EXISTS idx_operations_commission ON operations(commission_id);
CREATE INDEX IF NOT EXISTS idx_handoffs_created ON handoffs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_prs_shipment ON prs(shipment_id);
CREATE INDEX IF NOT EXISTS idx_prs_repo ON prs(repo_id);
CREATE INDEX IF NOT EXISTS idx_prs_commission ON prs(commission_id);
CREATE INDEX IF NOT EXISTS idx_prs_status ON prs(status);
CREATE INDEX IF NOT EXISTS idx_plans_commission ON plans(commission_id);
CREATE INDEX IF NOT EXISTS idx_plans_status ON plans(status);
CREATE INDEX IF NOT EXISTS idx_plans_cycle ON plans(cycle_id);
CREATE INDEX IF NOT EXISTS idx_conclaves_commission ON conclaves(commission_id);
CREATE INDEX IF NOT EXISTS idx_conclaves_status ON conclaves(status);
CREATE INDEX IF NOT EXISTS idx_notes_commission ON notes(commission_id);
CREATE INDEX IF NOT EXISTS idx_notes_shipment ON notes(shipment_id);

-- Work Orders (1:1 with Shipment)
CREATE TABLE IF NOT EXISTS work_orders (
	id TEXT PRIMARY KEY,
	shipment_id TEXT NOT NULL UNIQUE,
	outcome TEXT NOT NULL,
	acceptance_criteria TEXT,  -- JSON array of criteria
	status TEXT NOT NULL CHECK(status IN ('draft', 'active', 'complete')) DEFAULT 'draft',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE
);

-- Cycles (n:1 with Shipment, ordered)
CREATE TABLE IF NOT EXISTS cycles (
	id TEXT PRIMARY KEY,
	shipment_id TEXT NOT NULL,
	sequence_number INTEGER NOT NULL,
	status TEXT NOT NULL CHECK(status IN ('draft', 'approved', 'implementing', 'review', 'complete', 'blocked', 'closed', 'failed')) DEFAULT 'draft',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	started_at DATETIME,
	completed_at DATETIME,
	FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE,
	UNIQUE(shipment_id, sequence_number)
);

CREATE INDEX IF NOT EXISTS idx_work_orders_shipment ON work_orders(shipment_id);
CREATE INDEX IF NOT EXISTS idx_work_orders_status ON work_orders(status);
CREATE INDEX IF NOT EXISTS idx_cycles_shipment ON cycles(shipment_id);
CREATE INDEX IF NOT EXISTS idx_cycles_status ON cycles(status);
CREATE INDEX IF NOT EXISTS idx_cycles_sequence ON cycles(shipment_id, sequence_number);

-- Cycle Work Orders (1:1 with Cycle)
CREATE TABLE IF NOT EXISTS cycle_work_orders (
	id TEXT PRIMARY KEY,
	cycle_id TEXT NOT NULL UNIQUE,
	shipment_id TEXT NOT NULL,
	outcome TEXT NOT NULL,
	acceptance_criteria TEXT,  -- JSON array of criteria
	status TEXT NOT NULL CHECK(status IN ('draft', 'active', 'complete')) DEFAULT 'draft',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (cycle_id) REFERENCES cycles(id) ON DELETE CASCADE,
	FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_cycle_work_orders_cycle ON cycle_work_orders(cycle_id);
CREATE INDEX IF NOT EXISTS idx_cycle_work_orders_shipment ON cycle_work_orders(shipment_id);
CREATE INDEX IF NOT EXISTS idx_cycle_work_orders_status ON cycle_work_orders(status);

-- Cycle Receipts (1:1 with CWO)
CREATE TABLE IF NOT EXISTS cycle_receipts (
	id TEXT PRIMARY KEY,
	cwo_id TEXT NOT NULL UNIQUE,
	shipment_id TEXT NOT NULL,
	delivered_outcome TEXT NOT NULL,
	evidence TEXT,
	verification_notes TEXT,
	status TEXT NOT NULL CHECK(status IN ('draft', 'submitted', 'verified')) DEFAULT 'draft',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (cwo_id) REFERENCES cycle_work_orders(id) ON DELETE CASCADE,
	FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_cycle_receipts_cwo ON cycle_receipts(cwo_id);
CREATE INDEX IF NOT EXISTS idx_cycle_receipts_shipment ON cycle_receipts(shipment_id);
CREATE INDEX IF NOT EXISTS idx_cycle_receipts_status ON cycle_receipts(status);

-- Receipts (1:1 with Shipment)
CREATE TABLE IF NOT EXISTS receipts (
	id TEXT PRIMARY KEY,
	shipment_id TEXT NOT NULL UNIQUE,
	delivered_outcome TEXT NOT NULL,
	evidence TEXT,
	verification_notes TEXT,
	status TEXT NOT NULL CHECK(status IN ('draft', 'submitted', 'verified')) DEFAULT 'draft',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_receipts_shipment ON receipts(shipment_id);
CREATE INDEX IF NOT EXISTS idx_receipts_status ON receipts(status);
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
		// Fresh install - check if we have old schema tables (migrations needed)
		var oldTableCount int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('operations', 'expeditions', 'missions')").Scan(&oldTableCount)
		if err != nil {
			return err
		}

		if oldTableCount > 0 {
			// Old schema exists - run migrations to upgrade
			return RunMigrations()
		} else {
			// Completely fresh install - create modern schema directly
			// Also create schema_version at max version to prevent migrations from running
			_, err = db.Exec(SchemaSQL)
			if err != nil {
				return err
			}
			// Mark all migrations as applied for fresh installs
			_, err = db.Exec(`
				CREATE TABLE IF NOT EXISTS schema_version (
					version INTEGER PRIMARY KEY,
					applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
				)
			`)
			if err != nil {
				return err
			}
			// Insert all migration versions as applied
			for i := 1; i <= 41; i++ {
				_, err = db.Exec("INSERT INTO schema_version (version) VALUES (?)", i)
				if err != nil {
					return err
				}
			}
			return nil
		}
	}

	// schema_version table exists - run any pending migrations
	return RunMigrations()
}

// GetSchemaSQL returns the authoritative schema SQL for use by tests.
// Tests should use this instead of hardcoding their own schema to prevent drift.
func GetSchemaSQL() string {
	return SchemaSQL
}
