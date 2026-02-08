# Ledger Entities (Database Schema)

**Source**: ORC SQLite database schema (`~/.orc/orc.db`)
**Purpose**: Structured work coordination and tracking
**Status**: Schema definition reference

---

## Commission

**Database Table**: `commissions`

**Definition**: Body of work being tracked. The top-level organizational unit in ORC.

**Fields**:
- `id` (TEXT PRIMARY KEY) - Format: COMM-XXX
- `title` (TEXT NOT NULL) - Human-readable commission name
- `description` (TEXT) - Detailed commission description
- `status` (TEXT) - Values: 'active', 'paused', 'complete', 'archived'
- `created_at`, `updated_at`, `completed_at` (DATETIME)

**Example**:
- COMM-001: "ORC 2.0 Implementation"

**Relationships**:
- Has many Work Orders
- Has many Workbenches

---

## Work Order

**Database Table**: `work_orders`

**Definition**: Task in the commission backlog. Work orders are churned through by IMPs working in workbenches.

**Fields**:
- `id` (TEXT PRIMARY KEY) - Format: WO-XXX
- `commission_id` (TEXT NOT NULL) - Foreign key to commissions
- `title` (TEXT NOT NULL) - Task name
- `description` (TEXT) - Detailed description
- `type` (TEXT) - Values: 'research', 'implementation', 'fix', 'documentation', 'maintenance'
- `status` (TEXT) - Values: 'backlog', 'next', 'in_progress', 'complete'
- `priority` (TEXT) - Values: 'low', 'medium', 'high'
- `assigned_imp` (TEXT) - IMP identifier (optional, e.g., "IMP-ZSH")
- `context_ref` (TEXT) - Reference to additional context
- `created_at`, `updated_at`, `completed_at` (DATETIME)

**Examples**:
- WO-008: "Workbench Creation and Worktree Integration"
- WO-004: "Clarify Forest Factory Architecture Model"

**Relationships**:
- Belongs to Commission

**Work Types**:
- `research` - Exploratory work, design
- `implementation` - Building features, writing code
- `fix` - Bug fixes
- `documentation` - Docs, knowledge capture
- `maintenance` - Chores, cleanup, refactoring

---

## Workbench

**Database Table**: `workbenches`

**Definition**: Physical worktree on disk. Isolated development environment where IMPs work.

**Key Concept**: Workbenches belong to commissions, not individual work orders. Multiple work orders are completed using a commission's workbenches.

**Fields**:
- `id` (TEXT PRIMARY KEY) - Format: WORKBENCH-XXX
- `commission_id` (TEXT NOT NULL) - Foreign key to commissions
- `name` (TEXT NOT NULL) - Descriptive name
- `path` (TEXT NOT NULL UNIQUE) - Filesystem path (e.g., ~/src/worktrees/ml-feature)
- `repos` (TEXT) - JSON array of repositories in workbench
- `status` (TEXT) - Values: 'active', 'idle', 'archived'
- `created_at`, `updated_at` (DATETIME)

**Example**:
```
Commission: "Auth Refactor"
├── workbench-backend (path: ~/src/worktrees/auth-backend, repos: ["main-app"])
├── workbench-frontend (path: ~/src/worktrees/auth-frontend, repos: ["main-app"])
└── workbench-api (path: ~/src/worktrees/auth-api, repos: ["api-service"])
```

**Relationships**:
- Belongs to Commission

**Special Case**: Working on ORC itself happens directly in ~/src/orc (no workbench needed)

---

## Handoff

**Database Table**: `handoffs`

**Definition**: Session context snapshot for continuity between Claude sessions.

**Fields**:
- `id` (TEXT PRIMARY KEY) - Format: HO-XXX
- `created_at` (DATETIME)
- `handoff_note` (TEXT NOT NULL) - Narrative note from previous Claude
- `active_commission_id` (TEXT) - Foreign key to commissions
- `active_work_orders` (TEXT) - JSON array of in-progress WO IDs
- `active_workbench_id` (TEXT) - Foreign key to workbenches (optional)
- `todos_snapshot` (TEXT) - JSON snapshot of todo list

**Purpose**: Enables orc handoff → orc prime workflow

**Relationships**:
- References current Commission
- References active Work Orders (multiple possible)
- Optionally references active Workbench

---

## Simplified Architecture

The ORC ledger uses a flat, simple structure:

```
Commission (body of work)
├── Work Orders [many] (backlog of tasks, typed)
└── Workbenches [many] (worktrees from multiple repos)
```

**Workflow**:
1. Create Commission for body of work
2. Create Work Orders (tasks) in commission
3. Create Workbenches (worktrees) for commission
4. IMPs work in workbenches to complete work orders
5. Commission completes when work orders done

**Key Points**:
- Flat hierarchy (no nested operations)
- Work orders typed (research vs implementation etc)
- Workbenches are commission-level workspaces
- No forced 1:1 relationships
- Flexible and simple

---

**Last Updated**: 2026-01-13
**Status**: Reference documentation
