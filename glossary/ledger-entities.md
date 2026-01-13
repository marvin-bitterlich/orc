# Ledger Entities (Database Schema)

**Source**: ORC SQLite database schema (`~/.orc/orc.db`)
**Purpose**: Structured work coordination and tracking
**Status**: Implemented

---

## Mission

**Database Table**: `missions`

**Definition**: Strategic goal representing months of work. The highest-level organizational unit in ORC.

**Fields**:
- `id` (TEXT PRIMARY KEY) - Format: MISSION-XXX
- `title` (TEXT NOT NULL) - Human-readable mission name
- `description` (TEXT) - Detailed mission description
- `status` (TEXT) - Values: 'active', 'paused', 'complete', 'archived'
- `created_at`, `updated_at`, `completed_at` (DATETIME)

**Example**:
- MISSION-001: "ORC 2.0 Implementation"

**Relationships**:
- Has many Operations

**Questions/Contentious**:
- How many missions should be active at once?
- Can missions have dependencies on each other?

---

## Operation

**Database Table**: `operations`

**Definition**: Phase of work within a mission. Represents weeks of coordinated effort.

**Fields**:
- `id` (TEXT PRIMARY KEY) - Format: OP-XXX
- `mission_id` (TEXT NOT NULL) - Foreign key to missions
- `title` (TEXT NOT NULL) - Operation name
- `description` (TEXT) - Detailed description
- `status` (TEXT) - Values: 'backlog', 'active', 'complete', 'cancelled'
- `created_at`, `updated_at`, `completed_at` (DATETIME)

**Examples**:
- OP-002: "Prototype New Functionality"
- OP-003: "Research & Design"

**Relationships**:
- Belongs to Mission
- Has many Work Orders

**Questions/Contentious**:
- What's the difference between 'active' operation and 'active' work orders?
- Can you have multiple active operations in one mission?

---

## Work Order

**Database Table**: `work_orders`

**Definition**: Concrete task that needs doing. Lives in backlog until claimed.

**Fields**:
- `id` (TEXT PRIMARY KEY) - Format: WO-XXX
- `operation_id` (TEXT NOT NULL) - Foreign key to operations
- `title` (TEXT NOT NULL) - Task name
- `description` (TEXT) - Detailed description
- `status` (TEXT) - Values: 'backlog', 'next', 'in_progress', 'complete'
- `assigned_imp` (TEXT) - IMP identifier (e.g., "IMP-ZSH")
- `context_ref` (TEXT) - Reference to additional context
- `created_at`, `updated_at`, `claimed_at` (DATETIME)

**Examples**:
- WO-008: "Grove Creation and Worktree Integration"
- WO-004: "Clarify Forest Factory Architecture Model"

**Relationships**:
- Belongs to Operation
- Can have one Expedition (optional)

**Questions/Contentious**:
- **CRITICAL**: What's the relationship between Work Order and Expedition?
- Is a Work Order "claimed" when an Expedition starts?
- Can you have a Work Order without ever creating an Expedition?
- What does `assigned_imp` mean here vs expedition's `assigned_imp`?

---

## Expedition

**Database Table**: `expeditions`

**Definition**: ??? NEEDS CLARIFICATION ???

**Fields**:
- `id` (TEXT PRIMARY KEY) - Format: EXP-XXX (assumed)
- `name` (TEXT NOT NULL) - Expedition name
- `work_order_id` (TEXT) - **OPTIONAL** foreign key to work_orders
- `assigned_imp` (TEXT) - IMP identifier
- `status` (TEXT) - Values: 'planning', 'active', 'paused', 'complete'
- `created_at`, `updated_at` (DATETIME)

**Relationships**:
- Optionally belongs to Work Order
- Has one Grove (usually)
- Has many Plans

**Questions/Contentious**:
- **CRITICAL**: What IS an expedition conceptually?
- Is it "active execution of a work order"?
- Why is `work_order_id` optional? When would you have expedition without WO?
- Is it the same as "investigation" in old terminology?
- When do you create an expedition vs just working on a WO?
- Should expeditions ALWAYS have a work order, or are some explorations?

---

## Grove

**Database Table**: `groves`

**Definition**: Physical worktree on disk. Isolated development environment.

**Fields**:
- `id` (TEXT PRIMARY KEY) - Format: GROVE-XXX (assumed)
- `path` (TEXT NOT NULL UNIQUE) - Filesystem path (e.g., ~/src/worktrees/ml-feature)
- `repos` (TEXT) - List of repositories in grove
- `expedition_id` (TEXT) - Foreign key to expeditions
- `status` (TEXT) - Values: 'active', 'idle', 'archived'
- `created_at`, `updated_at` (DATETIME)

**Relationships**:
- Belongs to Expedition (usually)

**Questions/Contentious**:
- Is expedition_id optional or required?
- Can you have a grove without an expedition?
- What's the difference between 'idle' and 'archived'?
- Should groves be auto-archived when expedition completes?

---

## Plan

**Database Table**: `plans`

**Definition**: Tech plan created during an expedition.

**Fields**:
- `id` (TEXT PRIMARY KEY) - Format: PLAN-XXX (assumed)
- `expedition_id` (TEXT NOT NULL) - Foreign key to expeditions
- `title` (TEXT NOT NULL) - Plan title
- `status` (TEXT) - Values: 'draft', 'approved', 'rejected', 'implemented'
- `graphiti_episode_uuid` (TEXT) - Link to Graphiti episode
- `created_at`, `approved_at` (DATETIME)

**Relationships**:
- Belongs to Expedition
- Links to Graphiti episode (optional)

**Questions/Contentious**:
- Is this the same as "IMP" from old terminology?
- Should this be renamed?
- Do we need to track plans in the ledger, or just in filesystem?
- What's the relationship between tech-plans/ directory and this table?

---

## Handoff

**Database Table**: `handoffs`

**Definition**: Session context snapshot for continuity between Claude sessions.

**Fields**:
- `id` (TEXT PRIMARY KEY) - Format: HO-XXX
- `created_at` (DATETIME)
- `handoff_note` (TEXT NOT NULL) - Narrative note from previous Claude
- `active_mission_id`, `active_operation_id`, `active_work_order_id`, `active_expedition_id` (TEXT)
- `todos_snapshot` (TEXT) - JSON snapshot of todo list
- `graphiti_episode_uuid` (TEXT) - Link to Graphiti episode

**Purpose**: Enables /g-handoff → /g-bootstrap workflow

**Relationships**:
- References current Mission, Operation, Work Order, Expedition
- Links to Graphiti episode

**Questions/Contentious**:
- Should we have multiple active contexts, or just one at a time?
- How do we handle switching between multiple expeditions?

---

## Dependency

**Database Table**: `dependencies`

**Definition**: Tracks blocking relationships between expeditions.

**Fields**:
- `id` (INTEGER PRIMARY KEY AUTOINCREMENT)
- `source_expedition_id` (TEXT NOT NULL) - The expedition that blocks
- `blocks_expedition_id` (TEXT NOT NULL) - The expedition being blocked
- `reason` (TEXT) - Why the dependency exists
- `created_at` (DATETIME)

**Relationships**:
- Between two Expeditions

**Questions/Contentious**:
- Should dependencies be at Work Order level instead?
- How do we visualize and manage these?
- Do we need dependency tracking yet, or is it premature?

---

## Summary of CRITICAL Questions

1. **Expedition vs Work Order**: What's the actual difference? Is Expedition just "WO in execution"?
2. **IMP Entity**: Should IMPs be their own table, or just string identifiers?
3. **Optional Relationships**: Why is expedition.work_order_id optional?
4. **Plan vs IMP**: Is "Plan" table misnamed? Should it be "IMP"?
5. **Grove Lifecycle**: When are groves created/archived relative to expeditions?

---

**Last Updated**: 2026-01-13
**Status**: Initial dump - needs discussion with El Presidente
