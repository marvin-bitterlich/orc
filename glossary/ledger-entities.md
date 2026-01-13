# Ledger Entities (Database Schema)

**Source**: ORC SQLite database schema (`~/.orc/orc.db`)
**Purpose**: Structured work coordination and tracking
**Status**: Schema definition reference

---

## Mission

**Database Table**: `missions`

**Definition**: Body of work being tracked. The top-level organizational unit in ORC.

**Fields**:
- `id` (TEXT PRIMARY KEY) - Format: MISSION-XXX
- `title` (TEXT NOT NULL) - Human-readable mission name
- `description` (TEXT) - Detailed mission description
- `status` (TEXT) - Values: 'active', 'paused', 'complete', 'archived'
- `created_at`, `updated_at`, `completed_at` (DATETIME)

**Example**:
- MISSION-001: "ORC 2.0 Implementation"

**Relationships**:
- Has many Work Orders
- Has many Groves

---

## Work Order

**Database Table**: `work_orders`

**Definition**: Task in the mission backlog. Work orders are churned through by IMPs working in groves.

**Fields**:
- `id` (TEXT PRIMARY KEY) - Format: WO-XXX
- `mission_id` (TEXT NOT NULL) - Foreign key to missions
- `title` (TEXT NOT NULL) - Task name
- `description` (TEXT) - Detailed description
- `type` (TEXT) - Values: 'research', 'implementation', 'fix', 'documentation', 'maintenance'
- `status` (TEXT) - Values: 'backlog', 'next', 'in_progress', 'complete'
- `priority` (TEXT) - Values: 'low', 'medium', 'high'
- `assigned_imp` (TEXT) - IMP identifier (optional, e.g., "IMP-ZSH")
- `context_ref` (TEXT) - Reference to additional context
- `created_at`, `updated_at`, `completed_at` (DATETIME)

**Examples**:
- WO-008: "Grove Creation and Worktree Integration"
- WO-004: "Clarify Forest Factory Architecture Model"

**Relationships**:
- Belongs to Mission

**Work Types**:
- `research` - Exploratory work, design, investigation
- `implementation` - Building features, writing code
- `fix` - Bug fixes
- `documentation` - Docs, knowledge capture
- `maintenance` - Chores, cleanup, refactoring

---

## Grove

**Database Table**: `groves`

**Definition**: Physical worktree on disk. Isolated development environment where IMPs work.

**Key Concept**: Groves belong to missions, not individual work orders. Multiple work orders are completed using a mission's groves.

**Fields**:
- `id` (TEXT PRIMARY KEY) - Format: GROVE-XXX
- `mission_id` (TEXT NOT NULL) - Foreign key to missions
- `name` (TEXT NOT NULL) - Descriptive name
- `path` (TEXT NOT NULL UNIQUE) - Filesystem path (e.g., ~/src/worktrees/ml-feature)
- `repos` (TEXT) - JSON array of repositories in grove
- `status` (TEXT) - Values: 'active', 'idle', 'archived'
- `created_at`, `updated_at` (DATETIME)

**Example**:
```
Mission: "Auth Refactor"
├── grove-backend (path: ~/src/worktrees/auth-backend, repos: ["intercom"])
├── grove-frontend (path: ~/src/worktrees/auth-frontend, repos: ["intercom"])
└── grove-api (path: ~/src/worktrees/auth-api, repos: ["api-service"])
```

**Relationships**:
- Belongs to Mission

**Special Case**: Working on ORC itself happens directly in ~/src/orc (no grove needed)

---

## Handoff

**Database Table**: `handoffs`

**Definition**: Session context snapshot for continuity between Claude sessions.

**Fields**:
- `id` (TEXT PRIMARY KEY) - Format: HO-XXX
- `created_at` (DATETIME)
- `handoff_note` (TEXT NOT NULL) - Narrative note from previous Claude
- `active_mission_id` (TEXT) - Foreign key to missions
- `active_work_orders` (TEXT) - JSON array of in-progress WO IDs
- `active_grove_id` (TEXT) - Foreign key to groves (optional)
- `todos_snapshot` (TEXT) - JSON snapshot of todo list
- `graphiti_episode_uuid` (TEXT) - Link to Graphiti episode

**Purpose**: Enables /g-handoff → /g-bootstrap workflow

**Relationships**:
- References current Mission
- References active Work Orders (multiple possible)
- Optionally references active Grove
- Links to Graphiti episode

---

## Simplified Architecture

The ORC ledger uses a flat, simple structure:

```
Mission (body of work)
├── Work Orders [many] (backlog of tasks, typed)
└── Groves [many] (worktrees from multiple repos)
```

**Workflow**:
1. Create Mission for body of work
2. Create Work Orders (tasks) in mission
3. Create Groves (worktrees) for mission
4. IMPs work in groves to complete work orders
5. Mission completes when work orders done

**Key Points**:
- Flat hierarchy (no nested operations)
- Work orders typed (research vs implementation etc)
- Groves are mission-level workspaces
- No forced 1:1 relationships
- Flexible and simple

---

**Last Updated**: 2026-01-13
**Status**: Reference documentation
