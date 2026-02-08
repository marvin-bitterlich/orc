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
- Has many Workshops
- Has many Shipments

---

## Workshop

**Database Table**: `workshops`

**Definition**: Collection of workbenches for coordinated work.

**Fields**:
- `id` (TEXT PRIMARY KEY) - Format: WORK-XXX
- `name` (TEXT NOT NULL) - Workshop name
- `description` (TEXT) - Workshop description
- `factory_id` (TEXT NOT NULL) - Foreign key to factories
- `gatehouse_path` (TEXT) - Path to gatehouse directory
- `active_commission_id` (TEXT) - Currently active commission
- `status` (TEXT) - Values: 'active', 'archived'
- `created_at`, `updated_at` (DATETIME)

**Relationships**:
- Belongs to Factory
- Has one Gatehouse (1:1)
- Has many Workbenches

---

## Workbench

**Database Table**: `workbenches`

**Definition**: Physical worktree on disk. Isolated development environment where IMPs work.

**Key Concept**: Workbenches belong to workshops. Tasks are assigned to workbenches for implementation.

**Fields**:
- `id` (TEXT PRIMARY KEY) - Format: BENCH-XXX
- `workshop_id` (TEXT NOT NULL) - Foreign key to workshops
- `name` (TEXT NOT NULL) - Descriptive name
- `path` (TEXT NOT NULL UNIQUE) - Filesystem path (e.g., ~/src/worktrees/ml-feature)
- `repos` (TEXT) - JSON array of repositories in workbench
- `focused_id` (TEXT) - Currently focused shipment
- `status` (TEXT) - Values: 'active', 'idle', 'archived'
- `created_at`, `updated_at` (DATETIME)

**Example**:
```
Workshop: "ORC Development" (WORK-014)
├── BENCH-044 (path: ~/wb/orc-044)
└── BENCH-051 (path: ~/wb/orc-45)
```

**Relationships**:
- Belongs to Workshop
- Has assigned Tasks

---

## Shipment

**Database Table**: `shipments`

**Definition**: Unit of work with exploration → implementation lifecycle.

**Fields**:
- `id` (TEXT PRIMARY KEY) - Format: SHIP-XXX
- `commission_id` (TEXT NOT NULL) - Foreign key to commissions
- `title` (TEXT NOT NULL) - Shipment name
- `description` (TEXT) - Detailed description
- `branch` (TEXT) - Git branch name
- `status` (TEXT) - Values: 'draft', 'exploring', 'specced', 'tasked', 'in_progress', 'complete'
- `created_at`, `updated_at` (DATETIME)

**Relationships**:
- Belongs to Commission
- Has many Tasks
- Has many Notes

**Lifecycle**:
```
draft → exploring → specced → tasked → in_progress → complete
```

---

## Task

**Database Table**: `tasks`

**Definition**: Specific implementation work within a shipment.

**Fields**:
- `id` (TEXT PRIMARY KEY) - Format: TASK-XXX
- `shipment_id` (TEXT NOT NULL) - Foreign key to shipments
- `title` (TEXT NOT NULL) - Task name
- `description` (TEXT) - Detailed description
- `assigned_workbench_id` (TEXT) - Foreign key to workbenches
- `status` (TEXT) - Values: 'ready', 'in_progress', 'blocked', 'paused', 'complete'
- `created_at`, `updated_at`, `completed_at` (DATETIME)

**Examples**:
- TASK-823: "Create docs/ directory and move architecture"
- TASK-826: "Merge AGENTS.md into CLAUDE.md"

**Relationships**:
- Belongs to Shipment
- Optionally assigned to Workbench

---

## Handoff

**Database Table**: `handoffs`

**Definition**: Session context snapshot for continuity between Claude sessions.

**Fields**:
- `id` (TEXT PRIMARY KEY) - Format: HO-XXX
- `created_at` (DATETIME)
- `handoff_note` (TEXT NOT NULL) - Narrative note from previous Claude
- `active_commission_id` (TEXT) - Foreign key to commissions
- `active_shipment_id` (TEXT) - Foreign key to shipments
- `active_workbench_id` (TEXT) - Foreign key to workbenches (optional)

**Purpose**: Enables orc handoff → orc prime workflow

**Relationships**:
- References current Commission
- References active Shipment
- Optionally references active Workbench

---

## Simplified Architecture

The ORC ledger uses a hierarchical structure:

```
Commission (body of work)
├── Workshops [many] (coordination spaces)
│   ├── Gatehouse (Goblin coordination point)
│   └── Workbenches [many] (IMP workspaces)
└── Shipments [many] (units of work)
    └── Tasks [many] (implementation work)
```

**Workflow**:
1. Create Commission for body of work
2. Create Workshop with Workbenches
3. Create Shipment for exploration/implementation
4. Create Tasks from ship-plan
5. IMPs claim and complete tasks in workbenches
6. Shipment completes when all tasks done

**Key Points**:
- Commission → Shipment for work tracking
- Workshop → Workbench for physical workspaces
- Tasks assigned to workbenches for implementation
- Plan/Apply pattern for infrastructure

---

**Last Updated**: 2026-02-08
**Status**: Reference documentation
