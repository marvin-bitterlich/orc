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
	entity_type TEXT NOT NULL CHECK(entity_type IN ('task', 'plan', 'note', 'shipment', 'conclave', 'tome')),
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
	focused_conclave_id TEXT,
	active_commission_id TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (factory_id) REFERENCES factories(id),
	FOREIGN KEY (active_commission_id) REFERENCES commissions(id)
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
	focused_id TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (workshop_id) REFERENCES workshops(id),
	FOREIGN KEY (repo_id) REFERENCES repos(id)
);

-- Shipyards (one per factory, auto-created)
CREATE TABLE IF NOT EXISTS shipyards (
	id TEXT PRIMARY KEY,
	factory_id TEXT NOT NULL UNIQUE,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (factory_id) REFERENCES factories(id)
);

-- Commissions (Tracks of work - what you're working on)
-- Workshop → Commissions is 1:many (a workshop can have multiple commissions)
CREATE TABLE IF NOT EXISTS commissions (
	id TEXT PRIMARY KEY,
	factory_id TEXT,
	workshop_id TEXT,
	title TEXT NOT NULL,
	description TEXT,
	status TEXT NOT NULL CHECK(status IN ('initial', 'active', 'paused', 'complete', 'archived', 'deleted')) DEFAULT 'initial',
	pinned INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	started_at DATETIME,
	completed_at DATETIME,
	updated_at DATETIME,
	FOREIGN KEY (factory_id) REFERENCES factories(id),
	FOREIGN KEY (workshop_id) REFERENCES workshops(id)
);

-- Messages (Agent mail system - actor-to-actor)
CREATE TABLE IF NOT EXISTS messages (
	id TEXT PRIMARY KEY,
	sender TEXT NOT NULL,
	recipient TEXT NOT NULL,
	subject TEXT NOT NULL,
	body TEXT NOT NULL,
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	read INTEGER DEFAULT 0
);

-- Shipments (Work containers)
-- Lifecycle: draft → queued → assigned → active → complete
-- conclave_id = source/origin conclave, shipyard_id = when in queue, assigned_workbench_id = when assigned
CREATE TABLE IF NOT EXISTS shipments (
	id TEXT PRIMARY KEY,
	commission_id TEXT NOT NULL,
	title TEXT NOT NULL,
	description TEXT,
	status TEXT NOT NULL CHECK(status IN ('draft', 'queued', 'assigned', 'active', 'complete')) DEFAULT 'draft',
	assigned_workbench_id TEXT,
	repo_id TEXT,
	branch TEXT,
	pinned INTEGER DEFAULT 0,
	conclave_id TEXT,
	shipyard_id TEXT,
	autorun INTEGER DEFAULT 0,
	priority INTEGER,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	completed_at DATETIME,
	FOREIGN KEY (commission_id) REFERENCES commissions(id),
	FOREIGN KEY (assigned_workbench_id) REFERENCES workbenches(id),
	FOREIGN KEY (repo_id) REFERENCES repos(id),
	FOREIGN KEY (conclave_id) REFERENCES conclaves(id),
	FOREIGN KEY (shipyard_id) REFERENCES shipyards(id)
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
	container_id TEXT,
	container_type TEXT CHECK(container_type IN ('conclave')),
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
	depends_on TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	claimed_at DATETIME,
	completed_at DATETIME,
	FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE,
	FOREIGN KEY (commission_id) REFERENCES commissions(id),
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

-- Plans (Implementation plans - 1:many with Task)
CREATE TABLE IF NOT EXISTS plans (
	id TEXT PRIMARY KEY,
	commission_id TEXT NOT NULL,
	task_id TEXT NOT NULL,
	title TEXT NOT NULL,
	description TEXT,
	content TEXT,
	status TEXT NOT NULL CHECK(status IN ('draft', 'pending_review', 'approved', 'escalated', 'superseded')) DEFAULT 'draft',
	pinned INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	approved_at DATETIME,
	conclave_id TEXT,
	promoted_from_id TEXT,
	promoted_from_type TEXT,
	supersedes_plan_id TEXT,
	FOREIGN KEY (commission_id) REFERENCES commissions(id),
	FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
	FOREIGN KEY (conclave_id) REFERENCES conclaves(id) ON DELETE SET NULL,
	FOREIGN KEY (supersedes_plan_id) REFERENCES plans(id) ON DELETE SET NULL
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
	close_reason TEXT,
	closed_by_note_id TEXT,
	FOREIGN KEY (commission_id) REFERENCES commissions(id),
	FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE SET NULL,
	FOREIGN KEY (conclave_id) REFERENCES conclaves(id) ON DELETE SET NULL,
	FOREIGN KEY (tome_id) REFERENCES tomes(id) ON DELETE SET NULL,
	FOREIGN KEY (closed_by_note_id) REFERENCES notes(id) ON DELETE SET NULL
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
CREATE INDEX IF NOT EXISTS idx_workshops_commission ON workshops(active_commission_id);
CREATE INDEX IF NOT EXISTS idx_workbenches_workshop ON workbenches(workshop_id);
CREATE INDEX IF NOT EXISTS idx_workbenches_status ON workbenches(status);
CREATE INDEX IF NOT EXISTS idx_workbenches_repo ON workbenches(repo_id);
CREATE INDEX IF NOT EXISTS idx_commissions_factory ON commissions(factory_id);
CREATE INDEX IF NOT EXISTS idx_commissions_workshop ON commissions(workshop_id);
CREATE INDEX IF NOT EXISTS idx_commissions_status ON commissions(status);
CREATE INDEX IF NOT EXISTS idx_messages_recipient ON messages(recipient, read);
CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender);
CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_shipments_commission ON shipments(commission_id);
CREATE INDEX IF NOT EXISTS idx_shipments_status ON shipments(status);
CREATE INDEX IF NOT EXISTS idx_shipments_workbench ON shipments(assigned_workbench_id);
CREATE INDEX IF NOT EXISTS idx_tomes_commission ON tomes(commission_id);
CREATE INDEX IF NOT EXISTS idx_tomes_conclave ON tomes(conclave_id);
CREATE INDEX IF NOT EXISTS idx_tasks_shipment ON tasks(shipment_id);
CREATE INDEX IF NOT EXISTS idx_tasks_commission ON tasks(commission_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_workbench ON tasks(assigned_workbench_id);
CREATE INDEX IF NOT EXISTS idx_tasks_tome ON tasks(tome_id);
CREATE INDEX IF NOT EXISTS idx_tasks_conclave ON tasks(conclave_id);
CREATE INDEX IF NOT EXISTS idx_operations_commission ON operations(commission_id);
CREATE INDEX IF NOT EXISTS idx_handoffs_created ON handoffs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_prs_shipment ON prs(shipment_id);
CREATE INDEX IF NOT EXISTS idx_prs_repo ON prs(repo_id);
CREATE INDEX IF NOT EXISTS idx_prs_commission ON prs(commission_id);
CREATE INDEX IF NOT EXISTS idx_prs_status ON prs(status);
CREATE INDEX IF NOT EXISTS idx_plans_commission ON plans(commission_id);
CREATE INDEX IF NOT EXISTS idx_plans_task ON plans(task_id);
CREATE INDEX IF NOT EXISTS idx_plans_status ON plans(status);
CREATE INDEX IF NOT EXISTS idx_plans_supersedes ON plans(supersedes_plan_id);
CREATE INDEX IF NOT EXISTS idx_conclaves_commission ON conclaves(commission_id);
CREATE INDEX IF NOT EXISTS idx_conclaves_status ON conclaves(status);
CREATE INDEX IF NOT EXISTS idx_notes_commission ON notes(commission_id);
CREATE INDEX IF NOT EXISTS idx_notes_shipment ON notes(shipment_id);
CREATE INDEX IF NOT EXISTS idx_shipyards_factory ON shipyards(factory_id);
CREATE INDEX IF NOT EXISTS idx_tomes_container ON tomes(container_id);
CREATE INDEX IF NOT EXISTS idx_shipments_conclave ON shipments(conclave_id);
CREATE INDEX IF NOT EXISTS idx_shipments_shipyard ON shipments(shipyard_id);

-- Receipts (1:1 with Task)
CREATE TABLE IF NOT EXISTS receipts (
	id TEXT PRIMARY KEY,
	task_id TEXT NOT NULL UNIQUE,
	delivered_outcome TEXT NOT NULL,
	evidence TEXT,
	verification_notes TEXT,
	status TEXT NOT NULL CHECK(status IN ('draft', 'submitted', 'verified')) DEFAULT 'draft',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_receipts_task ON receipts(task_id);
CREATE INDEX IF NOT EXISTS idx_receipts_status ON receipts(status);

-- Gatehouses (1:1 with Workshop - Goblin seat)
CREATE TABLE IF NOT EXISTS gatehouses (
	id TEXT PRIMARY KEY,
	workshop_id TEXT NOT NULL UNIQUE,
	status TEXT NOT NULL DEFAULT 'active',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (workshop_id) REFERENCES workshops(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_gatehouses_workshop ON gatehouses(workshop_id);
CREATE INDEX IF NOT EXISTS idx_gatehouses_status ON gatehouses(status);

-- Kennels (IMP monitoring seats - linked to Workbenches)
CREATE TABLE IF NOT EXISTS kennels (
	id TEXT PRIMARY KEY,
	workbench_id TEXT NOT NULL,
	status TEXT NOT NULL CHECK(status IN ('vacant', 'occupied', 'away')) DEFAULT 'vacant',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (workbench_id) REFERENCES workbenches(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_kennels_workbench ON kennels(workbench_id);
CREATE INDEX IF NOT EXISTS idx_kennels_status ON kennels(status);

-- Dogbeds (Goblin docking points - linked to Gatehouses)
CREATE TABLE IF NOT EXISTS dogbeds (
	id TEXT PRIMARY KEY,
	gatehouse_id TEXT NOT NULL,
	status TEXT NOT NULL CHECK(status IN ('vacant', 'occupied')) DEFAULT 'vacant',
	docked_kennel_id TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (gatehouse_id) REFERENCES gatehouses(id) ON DELETE CASCADE,
	FOREIGN KEY (docked_kennel_id) REFERENCES kennels(id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS idx_dogbeds_gatehouse ON dogbeds(gatehouse_id);
CREATE INDEX IF NOT EXISTS idx_dogbeds_status ON dogbeds(status);
CREATE INDEX IF NOT EXISTS idx_dogbeds_kennel ON dogbeds(docked_kennel_id);

-- Patrols (monitoring sessions - owned by watchdog actor)
CREATE TABLE IF NOT EXISTS patrols (
	id TEXT PRIMARY KEY,
	kennel_id TEXT NOT NULL,
	target TEXT NOT NULL,
	status TEXT NOT NULL CHECK(status IN ('active', 'completed', 'escalated')) DEFAULT 'active',
	config TEXT,
	started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	ended_at DATETIME,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (kennel_id) REFERENCES kennels(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_patrols_kennel ON patrols(kennel_id);
CREATE INDEX IF NOT EXISTS idx_patrols_status ON patrols(status);

-- Stucks (consecutive failure rollups)
CREATE TABLE IF NOT EXISTS stucks (
	id TEXT PRIMARY KEY,
	patrol_id TEXT NOT NULL,
	first_check_id TEXT,
	check_count INTEGER DEFAULT 1,
	status TEXT NOT NULL CHECK(status IN ('open', 'resolved', 'escalated')) DEFAULT 'open',
	resolved_at DATETIME,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (patrol_id) REFERENCES patrols(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_stucks_patrol ON stucks(patrol_id);
CREATE INDEX IF NOT EXISTS idx_stucks_status ON stucks(status);

-- Checks (individual observations)
CREATE TABLE IF NOT EXISTS checks (
	id TEXT PRIMARY KEY,
	patrol_id TEXT NOT NULL,
	stuck_id TEXT,
	pane_content TEXT,
	outcome TEXT NOT NULL CHECK(outcome IN ('working', 'idle', 'menu', 'typed', 'error')),
	captured_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (patrol_id) REFERENCES patrols(id) ON DELETE CASCADE,
	FOREIGN KEY (stuck_id) REFERENCES stucks(id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS idx_checks_patrol ON checks(patrol_id);
CREATE INDEX IF NOT EXISTS idx_checks_stuck ON checks(stuck_id);
CREATE INDEX IF NOT EXISTS idx_checks_outcome ON checks(outcome);

-- Approvals (1:1 with Plan)
CREATE TABLE IF NOT EXISTS approvals (
	id TEXT PRIMARY KEY,
	plan_id TEXT NOT NULL UNIQUE,
	task_id TEXT NOT NULL,
	mechanism TEXT NOT NULL,
	reviewer_input TEXT,
	reviewer_output TEXT,
	outcome TEXT NOT NULL CHECK (outcome IN ('approved', 'escalated')),
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE,
	FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_approvals_plan ON approvals(plan_id);
CREATE INDEX IF NOT EXISTS idx_approvals_task ON approvals(task_id);
CREATE INDEX IF NOT EXISTS idx_approvals_outcome ON approvals(outcome);

-- Escalations (references approval, plan, task)
CREATE TABLE IF NOT EXISTS escalations (
	id TEXT PRIMARY KEY,
	approval_id TEXT,
	plan_id TEXT NOT NULL,
	task_id TEXT NOT NULL,
	reason TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'resolved', 'dismissed')),
	routing_rule TEXT NOT NULL DEFAULT 'workshop_gatehouse',
	origin_actor_id TEXT NOT NULL,
	target_actor_id TEXT,
	resolution TEXT,
	resolved_by TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	resolved_at DATETIME,
	FOREIGN KEY (approval_id) REFERENCES approvals(id) ON DELETE SET NULL,
	FOREIGN KEY (plan_id) REFERENCES plans(id),
	FOREIGN KEY (task_id) REFERENCES tasks(id)
);
CREATE INDEX IF NOT EXISTS idx_escalations_approval ON escalations(approval_id);
CREATE INDEX IF NOT EXISTS idx_escalations_plan ON escalations(plan_id);
CREATE INDEX IF NOT EXISTS idx_escalations_task ON escalations(task_id);
CREATE INDEX IF NOT EXISTS idx_escalations_status ON escalations(status);
CREATE INDEX IF NOT EXISTS idx_escalations_target ON escalations(target_actor_id);

-- Manifests (1:1 with Shipment)
CREATE TABLE IF NOT EXISTS manifests (
	id TEXT PRIMARY KEY,
	shipment_id TEXT NOT NULL UNIQUE,
	created_by TEXT NOT NULL,
	attestation TEXT,
	tasks TEXT,
	ordering_notes TEXT,
	status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'launched')),
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_manifests_shipment ON manifests(shipment_id);
CREATE INDEX IF NOT EXISTS idx_manifests_status ON manifests(status);
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
			for i := 1; i <= 47; i++ {
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
