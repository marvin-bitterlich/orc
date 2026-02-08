# ORC Architecture

**Updated:** 2026-02-03

---

## C4 Model Overview

This document uses C4 model terminology for architectural description:

| Level | Scope | Documented Here |
|-------|-------|-----------------|
| C1: System Context | External systems, actors | Yes |
| C2: Container | Deployable/distinct units | Yes |
| C3: Component | Modules within containers | Yes |
| C4: Code | Files, functions, classes | No (too granular) |

---

## C1: System Context

ORC (Orchestrator) is a commission coordination system for managing complex, multi-repository software development work. It coordinates between human operators and Claude AI agents.

**External Actors:**
- **El Presidente** - Human operator who drives commissions
- **Claude Code** - AI agent runtime that executes skills
- **Git** - Version control (worktrees for isolation)
- **TMux** - Terminal multiplexer for agent sessions

---

## C2: Containers

| Container | Location | Description |
|-----------|----------|-------------|
| CLI | cmd/orc/, internal/ | ORC command-line tool binary |
| Database | internal/db/, schema/ | SQLite ledger with Atlas migrations |
| Skills | glue/skills/ | Claude Code skill definitions |
| Hooks | glue/hooks/ | Git and Claude Code hooks |
| Config | .orc/ | Runtime configuration per workspace |
| Documentation | *.md, docs/ | Project and workflow documentation |

---

## C3: Components

### CLI (cmd/orc/, internal/)

| Component | Location | Description |
|-----------|----------|-------------|
| Entry | cmd/orc/ | CLI entry and command registration |
| App | internal/app/ | Application services and use cases |
| CLI | internal/cli/ | Cobra command implementations |
| Core | internal/core/ | Domain entities and logic |
| Models | internal/models/ | Data models |
| Adapters | internal/adapters/ | Interface adapters |
| Ports | internal/ports/ | Port interfaces |
| DB | internal/db/ | Database access layer |
| TMux | internal/tmux/ | TMux integration |

### Skills (glue/skills/)

**Shipment Workflow:**
| Skill | Description |
|-------|-------------|
| ship-new | Create new shipments |
| ship-synthesize | Knowledge compaction → summary note |
| ship-plan | C2/C3 engineering review → tasks |
| ship-queue | View shipyard queue |
| ship-complete | Complete shipments |
| ship-deploy | Deploy shipments |
| ship-verify | Verify shipments |

**IMP Workflow:**
| Skill | Description |
|-------|-------------|
| imp-start | Begin IMP work on shipment |
| imp-plan-create | Create C4 implementation plans |
| imp-plan-submit | Submit plans for review |
| imp-auto | Toggle auto mode |
| imp-rec | Create receipts |
| imp-escalate | Escalate to gatehouse |
| imp-nudge | Manual re-propulsion check |
| imp-poll | Check shipyard queue for work |

**Goblin Workflow:**
| Skill | Description |
|-------|-------------|
| goblin-escalation-receive | Handle incoming escalations |

**Setup & Admin:**
| Skill | Description |
|-------|-------------|
| orc-commission | Create new commissions |
| orc-workshop | Create new workshops |
| orc-workshop-archive | Archive workshops |
| orc-workshop-templates | Manage workshop templates |
| orc-workbench | Create new workbenches |
| orc-repo | Add repositories to config |

**Utilities:**
| Skill | Description |
|-------|-------------|
| orc-first-run | Interactive first-run walkthrough |
| orc-interview | Reusable interview primitive |
| orc-architecture | Maintain ARCHITECTURE.md |
| orc-debug | View debug logs |
| orc-help | Orientation to ORC skills |
| orc-ping | Verify ORC is working |
| orc-self-test | Integration self-testing |

**Exploration:**
| Skill | Description |
|-------|-------------|
| orc-ideate | Rapid idea capture |
| orc-journal | Capture observations and learnings |

### Database (internal/db/, schema/)

| Component | Description |
|-----------|-------------|
| schema.sql | Database schema definition |
| migrations/ | Atlas migration files |

---

## Skill Workflow

```
/ship-synthesize → Summary note (knowledge compaction)
       ↓
/ship-plan → Tasks (C2/C3 engineering review)
       ↓
/imp-plan-create → Plans (C4 file-level detail)
       ↓
Implementation → Code
```

**Zoom Level Ownership:**
- C2/C3 (containers, components): ship-plan, orc-architecture
- C4 (files, functions): imp-plan-create

---

## Core Hierarchy

```
Commission (coordination scope)
├── Shipments (execution containers with lifecycle)
│   ├── Notes (ideas, questions, decisions, specs)
│   └── Tasks (atomic units of work)
└── Workbenches (git worktrees with IMP agents)
```

**Design Principles:**
- **Simplicity Over Hierarchy** - Flat structure where possible
- **Commission-Centric Organization** - Commission is the coordination boundary
- **Database as Source of Truth** - Single authoritative data source (SQLite)
- **Skill-Driven Workflows** - Claude skills encode process knowledge

---

## Core Features

### 1. Commission Management
- **Commission**: Top-level coordination scope (e.g., "Sidekiq Deprecation", "Auth Refactor")
- Owns multiple Workbenches (workspaces) and Work Orders (tasks)
- Each commission can have a dedicated TMux session with IMP agents in workbenches
- Support for both ORC-development commissions and application-code commissions

### 2. Work Order System
**Flat hierarchy with Epic/Parent grouping:**
- Work orders belong directly to commissions (no forced intermediate layers)
- Optional parent/child relationships for logical grouping
- Status field with 7 states: ready, design, implement, deploy, blocked, paused, complete
- Type categorization: research, implementation, fix, documentation, maintenance
- Pinnable work orders for visibility
- Emoji-rich CLI output for quick status scanning

**Key Commands:**
```bash
orc work-order create "Task description" [--parent WO-XXX]
orc work-order claim WO-XXX
orc work-order complete WO-XXX
orc work-order pin WO-XXX
orc work-order update WO-XXX --title "New title"
orc summary                    # Hierarchical view with pinned items
```

### 3. Workbench Management (Git Worktree Integration)
**Workbench**: An isolated git worktree for a commission, registered in the database

- Workbenches belong to commissions, not individual work orders
- Multiple workbenches per commission (e.g., backend, frontend, api repos)
- Work orders are assigned to workbenches (via assigned_workbench_id)
- IMP (Implementation) = Workbench (conceptual equivalence, no separate entity)

**Key Commands:**
```bash
orc workbench create [name] --repos main-app --commission COMM-001
orc workbench list [--commission COMM-XXX]
orc workbench show WORKBENCH-XXX
orc workbench rename WORKBENCH-XXX new-name
```

**Workbench Features:**
- Creates git worktree automatically
- Writes config.json to .orc/ subdirectory (reference only)
- Writes .orc-commission marker for context detection
- Opens in TMux with 3-pane IMP layout: vim | claude | shell

### 4. Two-Tier Agent Architecture (ORC + IMPs)
**Concept**: Simple two-tier coordination with one orchestrator and multiple implementation agents

**Agent Types:**
- **ORC (Orchestrator)**: Master coordinator that manages commissions, creates workbenches, assigns epics
- **IMP (Implementation Agent)**: Works within a workbench on assigned epics/tasks

**Architecture Principles:**
- Single ORC orchestrator
- IMPs operate within workbench boundaries
- ORC coordinates via mail system and epic assignments
- IMPs can create workbenches/epics but not commissions

**Mail System:**
- Direct IMP ↔ ORC communication
- Work orders/tasks used for coordination

### 5. Context Preservation & Handoffs
**orc prime & handoff Integration:**

Session boundaries are preserved through:
- **Handoffs**: Narrative summaries stored in SQLite database
- **orc prime**: Injects context at session start (commission, epic, tasks, recent handoff)
- **orc handoff create**: Captures session work, decisions, discoveries at session end

**Key Features:**
- Cross-session memory preservation via handoff narratives
- New Claude sessions start with full historical context
- Zero "cold start" problem
- Handoffs searchable via CLI (`orc handoff list`, `orc handoff show`)

### 6. TMux Integration
**One TMux session per commission:**
```
TMux Session: "Commission Name" (orc-COMM-XXX)
├── Pane 0: ORC (coordination)
├── Pane 1: IMP in workbench-1 (vim)
├── Pane 2: IMP in workbench-1 (claude)
└── Pane 3: IMP in workbench-1 (shell)
```

**Features:**
- Workbench directories contain `.orc-config.json` for context detection
- All panes CD into workbench directory
- Easy context switching between coordination and implementation

**Agent Starting Pattern:**
ORC uses **direct prompt injection** when starting Claude agents in TMux:

```bash
claude "Run `orc prime`"
```

This pattern replaces SessionStart hooks (which are broken in Claude Code v2.1.7). When agents start:
1. TMux sends the command with prompt: `claude "Run \`orc prime\`"`
2. Claude receives the prompt and executes `orc prime`
3. `orc prime` detects the agent's location (workbench/commission/global) and provides appropriate context
4. Agent begins work with full context immediately

**Benefits:**
- Reliable context injection (not dependent on broken hooks)
- Immediate agent activation
- Clear, explicit agent instructions
- Easier debugging (command visible in TMux history)
- Works consistently across all agent types (IMPs, ORC)

---

## Technology Stack

### Primary Technologies
- **Language**: Go (CLI binary)
- **Database**: SQLite (single source of truth for all operational data)
- **Version Control**: Git (with worktree integration)
- **Session Management**: TMux
- **AI Integration**: Claude API (via Claude Code CLI)

### Database Schema (SQLite)

**Core Tables:**
- `commissions` - Top-level coordination scopes
- `work_orders` - Tasks with status, type, parent_id, assigned_workbench_id, pinned flag
- `workbenches` - Git worktrees registered to commissions
- `handoffs` - Session handoff narratives for context preservation

**Removed Tables (simplified in 2.0):**
- ~~`operations`~~ - Removed (too rigid, use work_order.type instead)
- ~~`expeditions`~~ - Removed (1:1:1 mapping didn't match workflow)
- ~~`dependencies`~~ - Not implemented yet (can add later if needed)

**Key Fields:**
- `work_orders.status`: ready | design | implement | deploy | blocked | paused | complete
- `work_orders.type`: research | implementation | fix | documentation | maintenance
- `work_orders.parent_id`: Optional epic/parent reference
- `work_orders.assigned_workbench_id`: Workbench assignment
- `work_orders.pinned`: Boolean for visibility

---

## Database System: SQLite

### SQLite (Single Source of Truth)
**Purpose:** Authoritative source for all structured operational data

**Stores:**
- Commissions, epics, rabbit holes, tasks
- Workbenches (git worktrees)
- Handoffs (session narratives)
- Tags and task-tag associations
- Current state (status, assignments, timestamps)

**Characteristics:**
- Fast, local, transactional
- Schema-enforced data integrity
- Deterministic queries (e.g., "show all ready tasks")
- Files like config.json are DERIVED from this, never read for decisions

### Handoff System

**Data Flow:**
```
Session Work → SQLite (operational state)
            ↓
         Handoff (at session end)
            ↓
    Narrative stored in SQLite
            ↓
    Next session reads via orc prime
```

**Bootstrap Process:**
```
New Session starts
    ↓
orc prime queries:
    - SQLite: Current commission, epic, tasks
    - SQLite: Recent handoff narrative
    - Git: Current branch, uncommitted changes
    ↓
Full context restored → Work begins
```

**Handoff Process:**
```
Session ending
    ↓
orc handoff create:
    - Handoff record in SQLite (narrative, task_ids, timestamp)
    - Config updated with current handoff ID
    ↓
Next session can bootstrap from this
```

**Key Insight:** SQLite tells you WHERE YOU ARE. Handoff narratives tell you HOW YOU GOT HERE.

---

## Roadmap & Future Features

### Active Development (In Progress)

**Semantic Epic System** ⚡ PRIORITY
- 9-epic type system with type-specific rules
- Tag-based task auto-routing
- Working modes for different contexts
- **Status**: Active development (EPIC-178)

**WO-021: Core Architecture Design**
- Cross-workbench coordination mechanisms (WO-016)
- Workbench lifecycle management (WO-015)
- Work order state transition model (WO-014)
- Inter-agent mail system refinement

**WO-057: Factory Production Line Issues** (Master ORC Tooling)
- Continuous improvements to CLI tools
- Bug fixes and UX refinements
- IMP escalations from COMM-002

### Completed Major Features (2.0)

✅ **Forest Factory Architecture Simplification**
- Removed Operations and Expeditions layers
- Flat Commission → Work Order structure
- Type-based categorization

✅ **Epic/Parent Work Order Hierarchy**
- Single-parent grouping
- Visual hierarchy in summary output
- Nested display with indentation

✅ **Status Field Consolidation**
- Merged status + phase into single field
- 7 semantic states (ready → complete)
- Emoji indicators for quick scanning

✅ **IMP Agent Deployment (COMM-002)**
- First real-world commission validated
- TMux automation working
- ORC ↔ IMP communication working
- **Validation**: Everything worked first try

✅ **Workbench Management**
- Create, list, rename commands
- Git worktree integration
- Commission context detection

✅ **Pinned Work Orders**
- Pin/unpin commands
- Inline emoji display (no separate section)
- Priority visibility

✅ **Work Order Update**
- Update title and/or description
- Set/change parent relationships

### Near-Term Roadmap

**Phase 1: Semantic Epic System**
- Formalize 9-epic type system
- Tag-based task organization
- Working modes for different contexts
- Epic focus and context injection

**Phase 2: Cross-Commission Coordination**
- Refine mail system
- IMP → ORC escalation workflow
- ORC → IMP directive system
- Async communication patterns

**Phase 3: Tooling & Integration**
- vim-orc plugin (fugitive-style)
- El Presidente's desk (document staging)
- ORC Claude skills plugin
- Room concept for contextual spaces

**Phase 4: Workflow Automation**
- Task state transitions (auto/manual rules)
- Workbench lifecycle automation
- TMux session management
- Handoff automation improvements

### Long-Term Vision

**Factory-Themed Terminology**
- Replace generic terms (Commission, Epic, Operation)
- Restore forest/factory metaphor consistently
- Personality + functionality

**Advanced Context Management**
- Proactive context suggestions
- Cross-session pattern detection
- Handoff quality scoring
- Context inheritance patterns

**Multi-Agent Orchestration**
- ORC coordinates multiple IMPs across workbenches
- IMP-to-IMP communication via ORC
- Commission-based coordination structure
- Resource allocation and prioritization

---

## Success Metrics

**First-Try Execution:**
- COMM-002 deployment: ✅ Worked on first attempt
- ORC + IMP coordination: ✅ Validated in real operation
- TMux automation: ✅ Clean integration
- Proto-mail system: ✅ Bidirectional communication working

**Handoff System:**
- Context preservation: ✅ Zero cold-start issues
- Cross-session continuity: ✅ Narratives span multiple days
- Handoff CLI: ✅ create, list, show commands working

**Operational Metrics:**
- Sessions with zero cold-start issues: 100%
- Architecture decisions preserved: 100%
- IMP tooling escalations resolved: 5/5

---

## Key Design Decisions

### Validated Patterns

**1. Simplicity Over Hierarchy**
- Flat structures preferred
- Remove unnecessary layers
- Consolidate overlapping concepts

**2. Keep It On The Rails**
- Constrained hierarchy (single parent) preferred
- Reject flexible but messy many-to-many (tags, labels)
- "Definitely needs hierarchy... not messy like tags"

**3. Database as Source of Truth**
- Single authoritative source (SQLite)
- Files are reference only, never read for decisions
- Prevent state drift

**4. Validated By Usage**
- Pragmatic validation through implementation
- First-try success validates entire integration
- Real-world operation trumps theoretical correctness

---

## Getting Started

### Installation
```bash
cd ~/src/orc
go build -o orc cmd/orc/main.go
# Binary available at: ./orc
```

### Prerequisites

**Claude Code Workspace Trust**

ORC requires Claude Code to trust specific directories where it creates workspaces.

Add to `~/.claude/settings.json`:
```json
{
  "permissions": {
    "additionalDirectories": [
      "~/src/worktrees",
      "~/src/factories"
    ]
  }
}
```

**Validation**: Run `orc doctor` to verify configuration.

**See**: INSTALL.md for detailed setup instructions.

### Initialize ORC
```bash
orc init    # Creates ~/.orc/ directory and initializes database
```

### Create Your First Commission
```bash
orc commission create "My Project" --description "Project description"
orc work-order create "Setup repository" --commission COMM-001
orc workbench create project-main --repos my-repo --commission COMM-001
```

### Work in a Workbench
```bash
cd ~/src/worktrees/project-main
orc status                    # Shows commission context
orc work-order claim WO-001
# ... do work ...
orc work-order complete WO-001
```

### Session Boundaries
```bash
# At session start
orc prime                     # Restores full context

# At session end
orc handoff create --note "Session summary..."
```

---

## Architecture Highlights

**What Makes ORC Unique:**

1. **SQLite Source of Truth** - Single authoritative database for all state
2. **Zero Cold-Start** - Full context preservation via handoff narratives
3. **Multi-Agent Coordination** - ORC/IMP architecture with mail system
4. **Git Worktree Native** - First-class support for isolated workspaces
5. **Semantic Epic System** - 9-epic types with working modes
6. **TMux Integration** - One session per commission, programmatic layout
7. **Flat But Structured** - Simple hierarchy with optional grouping
8. **Validation-Driven** - Proven by first-try production deployment

---

## Technical References

**Codebase Structure:**
- `internal/cli/` - Command implementations (commission, work-order, workbench, etc.)
- `internal/models/` - Database models and queries
- `internal/context/` - Commission context detection
- `internal/tmux/` - TMux session management
- `internal/db/` - SQLite database setup and migrations

**Key Files:**
- `internal/cli/summary.go` - Hierarchical work order display
- `internal/cli/workbench.go` - Workbench management and TMux integration
- `internal/cli/work_order.go` - Work order CRUD operations
- `internal/models/workorder.go` - Work order model and queries

**External Dependencies:**
- Claude Code CLI (for AI integration)
- Git (worktree support)
- TMux (session management)

---

## Glossary

**Commission**: Top-level coordination scope, owns workbenches and work orders
**Work Order**: Task within a commission, flat hierarchy with optional parent
**Workbench**: Git worktree registered to a commission, physical workspace
**IMP**: Implementation (conceptual layer over workbench, not separate entity)
**ORC**: Main orchestrator that coordinates IMPs across commissions
**Proto-Mail**: Work-order-based bidirectional messaging
**Handoff**: Session boundary artifact (narrative + work state)
**orc prime**: Context injection at session start
**orc handoff**: Context capture at session end
**Epic**: Work order with children (parent in hierarchy)

---

## Contact & Contributions

**Repository**: ~/src/orc
**Documentation**: CLAUDE.md (project context), glossary/ (definitions)
**Status**: Active development, production-validated (COMM-002)
