-- ORC Database Schema
-- This file defines the SQLite schema for the ORC orchestration system.
-- Use Atlas for migrations: see CLAUDE.md for workflow.

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
	entity_type TEXT NOT NULL CHECK(entity_type IN ('task', 'plan', 'note', 'shipment', 'tome')),
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
	active_commission_id TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (factory_id) REFERENCES factories(id),
	FOREIGN KEY (active_commission_id) REFERENCES commissions(id)
);

-- Workbenches (Git worktrees within a workshop)
-- Path is computed dynamically as ~/wb/{name}, not stored
CREATE TABLE IF NOT EXISTS workbenches (
	id TEXT PRIMARY KEY,
	workshop_id TEXT NOT NULL,
	name TEXT NOT NULL UNIQUE,
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

-- Shipments (Work containers)
-- Lifecycle: draft → ready → in-progress → closed
CREATE TABLE IF NOT EXISTS shipments (
	id TEXT PRIMARY KEY,
	commission_id TEXT NOT NULL,
	title TEXT NOT NULL,
	description TEXT,
	status TEXT NOT NULL CHECK(status IN ('draft', 'ready', 'in-progress', 'closed')) DEFAULT 'draft',
	closed_reason TEXT,
	assigned_workbench_id TEXT,
	repo_id TEXT,
	branch TEXT,
	pinned INTEGER DEFAULT 0,
	priority INTEGER,
	spec_note_id TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	completed_at DATETIME,
	FOREIGN KEY (commission_id) REFERENCES commissions(id),
	FOREIGN KEY (assigned_workbench_id) REFERENCES workbenches(id),
	FOREIGN KEY (repo_id) REFERENCES repos(id),
	FOREIGN KEY (spec_note_id) REFERENCES notes(id) ON DELETE SET NULL
);

-- Tomes (Knowledge containers)
CREATE TABLE IF NOT EXISTS tomes (
	id TEXT PRIMARY KEY,
	commission_id TEXT NOT NULL,
	title TEXT NOT NULL,
	description TEXT,
	status TEXT NOT NULL CHECK(status IN ('open', 'closed')) DEFAULT 'open',
	assigned_workbench_id TEXT,
	pinned INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	closed_at DATETIME,
	FOREIGN KEY (commission_id) REFERENCES commissions(id),
	FOREIGN KEY (assigned_workbench_id) REFERENCES workbenches(id)
);

-- Tasks (Atomic units of work)
CREATE TABLE IF NOT EXISTS tasks (
	id TEXT PRIMARY KEY,
	shipment_id TEXT,
	commission_id TEXT NOT NULL,
	tome_id TEXT,
	title TEXT NOT NULL,
	description TEXT,
	type TEXT CHECK(type IN ('research', 'implementation', 'fix', 'documentation', 'maintenance')),
	status TEXT NOT NULL CHECK(status IN ('open', 'in-progress', 'blocked', 'closed')) DEFAULT 'open',
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
	promoted_from_id TEXT,
	promoted_from_type TEXT,
	supersedes_plan_id TEXT,
	FOREIGN KEY (commission_id) REFERENCES commissions(id),
	FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
	FOREIGN KEY (supersedes_plan_id) REFERENCES plans(id) ON DELETE SET NULL
);

-- Notes (Observations and learnings)
CREATE TABLE IF NOT EXISTS notes (
	id TEXT PRIMARY KEY,
	commission_id TEXT NOT NULL,
	shipment_id TEXT,
	tome_id TEXT,
	title TEXT NOT NULL,
	content TEXT,
	type TEXT,
	status TEXT NOT NULL CHECK(status IN ('open', 'in_flight', 'resolved', 'closed')) DEFAULT 'open',
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
CREATE INDEX IF NOT EXISTS idx_shipments_commission ON shipments(commission_id);
CREATE INDEX IF NOT EXISTS idx_shipments_status ON shipments(status);
CREATE INDEX IF NOT EXISTS idx_shipments_workbench ON shipments(assigned_workbench_id);
CREATE INDEX IF NOT EXISTS idx_tomes_commission ON tomes(commission_id);
CREATE INDEX IF NOT EXISTS idx_tasks_shipment ON tasks(shipment_id);
CREATE INDEX IF NOT EXISTS idx_tasks_commission ON tasks(commission_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_workbench ON tasks(assigned_workbench_id);
CREATE INDEX IF NOT EXISTS idx_tasks_tome ON tasks(tome_id);
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
CREATE INDEX IF NOT EXISTS idx_notes_commission ON notes(commission_id);
CREATE INDEX IF NOT EXISTS idx_notes_shipment ON notes(shipment_id);
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
	routing_rule TEXT NOT NULL DEFAULT 'workshop',
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

-- Workshop Logs (audit trail for workshop changes)
CREATE TABLE IF NOT EXISTS workshop_logs (
	id TEXT PRIMARY KEY,
	workshop_id TEXT NOT NULL,
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	actor_id TEXT,
	entity_type TEXT NOT NULL,
	entity_id TEXT NOT NULL,
	action TEXT NOT NULL CHECK(action IN ('create', 'update', 'delete')),
	field_name TEXT,
	old_value TEXT,
	new_value TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (workshop_id) REFERENCES workshops(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_workshop_logs_workshop ON workshop_logs(workshop_id);
CREATE INDEX IF NOT EXISTS idx_workshop_logs_timestamp ON workshop_logs(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_workshop_logs_actor ON workshop_logs(actor_id);
CREATE INDEX IF NOT EXISTS idx_workshop_logs_entity ON workshop_logs(entity_type, entity_id);

-- Hook Events (audit trail for Claude Code hook invocations)
CREATE TABLE IF NOT EXISTS hook_events (
	id TEXT PRIMARY KEY,
	workbench_id TEXT NOT NULL,
	hook_type TEXT NOT NULL CHECK(hook_type IN ('Stop', 'UserPromptSubmit')),
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	payload_json TEXT,
	cwd TEXT,
	session_id TEXT,
	shipment_id TEXT,
	shipment_status TEXT,
	task_count_incomplete INTEGER,
	decision TEXT NOT NULL CHECK(decision IN ('allow', 'block')),
	reason TEXT,
	duration_ms INTEGER,
	error TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (workbench_id) REFERENCES workbenches(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_hook_events_workbench ON hook_events(workbench_id);
CREATE INDEX IF NOT EXISTS idx_hook_events_timestamp ON hook_events(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_hook_events_type ON hook_events(hook_type);
