# ORC Architecture

**Updated:** 2026-02-11

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
| Skills (global) | glue/skills/ | Claude Code skill definitions (deployed via glue) |
| Skills (repo-local) | .claude/skills/ | Repo-specific skills (not deployed globally) |
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

### Global Skills (glue/skills/)

Skills deployed globally via the glue system.

**Shipment Workflow:**
| Skill | Description |
|-------|-------------|
| ship-new | Create new shipments |
| ship-synthesize | Knowledge compaction -> summary note |
| ship-plan | C2/C3 engineering review -> tasks |
| ship-complete | Complete shipments |
| ship-run | Bridge ship-plan -> Teams execution |
| ship-deploy | Deploy shipments |
| ship-freshness | Rebase and validate tasks/notes |

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
| orc-help | Orientation to ORC skills |

**Exploration:**
| Skill | Description |
|-------|-------------|
| orc-ideate | Rapid idea capture |
| orc-journal | Capture observations and learnings |

### Repo-Local Skills (.claude/skills/)

Skills specific to this repository (not deployed globally):

| Skill | Description |
|-------|-------------|
| hello-exercise | Manual test for first-run flow |
| bootstrap-test | Test make bootstrap in a fresh macOS VM using Tart |
| docs-doctor | Validate ORC documentation against code reality |
| orc-architecture | Maintain ARCHITECTURE.md |
| orc-self-test | Integration self-testing |
| self-test | Orchestrate all ORC integration checks via Claude Teams |

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
Teams assigns tasks to IMPs → Implementation
       ↓
/ship-deploy → Merge to main
       ↓
/ship-complete → Close shipment
```

**Zoom Level Ownership:**
- C2/C3 (containers, components): ship-plan, orc-architecture
- C4 (files, functions): IMP execution via Teams

---

## Core Hierarchy

```
Commission (coordination scope)
├── Shipments (execution containers with lifecycle)
│   ├── Notes (ideas, questions, decisions, specs)
│   └── Tasks (atomic units of work)
└── Workbenches (git worktrees for agents)
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
- Owns multiple Workshops with Workbenches and Shipments
- Each commission can have a dedicated TMux session with IMP agents in workbenches
- Support for both ORC-development commissions and application-code commissions

### 2. Shipment & Task System
**Shipment-based workflow:**
- Shipments represent units of work with a simple 4-status lifecycle: draft -> ready -> in-progress -> closed
- Tasks belong to shipments and represent specific implementation work
- Task lifecycle: open -> in-progress -> blocked -> closed
- All transitions are manual (Goblin decides)
- Type categorization: research, implementation, fix, documentation, maintenance
- Pinnable items for visibility

**Key Commands:**
```bash
orc shipment create "Title" --commission COMM-XXX
orc task create "Task description" --shipment SHIP-XXX
orc task complete TASK-XXX
orc summary                    # Hierarchical view with pinned items
```

### 3. Workbench Management (Git Worktree Integration)
**Workbench**: An isolated git worktree for a workshop, registered in the database

- Workbenches belong to workshops, assigned shipments for implementation
- Multiple workbenches per workshop (e.g., backend, frontend, api repos)
- Tasks are assigned to workbenches (via assigned_workbench_id)
- IMP (Implementation) = Workbench (conceptual equivalence, no separate entity)

**Key Commands:**
```bash
orc workbench create --workshop WORK-001 --repo-id REPO-001
orc workbench list [--workshop WORK-XXX]
orc workbench show BENCH-XXX
orc workbench rename BENCH-XXX new-name
```

**Workbench Features:**
- Creates git worktree automatically
- Writes config.json to .orc/ subdirectory (reference only)
- Writes .orc-commission marker for context detection
- Opens in TMux with 3-pane IMP layout: vim | claude | shell

### 4. Actor Model (Goblin, IMP)
**Concept**: Two-actor model with complementary roles

**Actor Types:**
- **Goblin (Coordinator)**: Human's long-running workbench pane. Creates/manages ORC tasks with the human. Memory and policy layer (what and why).
- **IMP (Worker)**: Disposable worker agent spawned by Claude Teams. Executes tasks using Teams primitives. Execution layer (how and who).

**Integration Model (Claude Teams):**
- ORC = memory and policy layer (what and why)
- Teams = execution layer (how and who)
- Goblin creates/manages ORC tasks with the human
- Teams workers (IMPs) execute using Teams primitives

**Communication:**
- Claude Teams messaging (replaces old mail/escalation system)
- Shared summary bus via `orc summary`

### 5. TMux Integration
**One TMux session per workshop:**
```
TMux Session: "Workshop Name" (orc-WORK-XXX)
├── Window 0: Goblin in BENCH-001 (claude | vim | shell)
├── Window 1: Workbench BENCH-002 (vim | claude | shell)
└── Window 2: Workbench BENCH-003 (vim | claude | shell)
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
- `shipments` - Work containers with lifecycle
- `tasks` - Atomic units of work within shipments
- `workbenches` - Git worktrees registered to workshops

**Key Fields:**
- `shipments.status`: draft | ready | in-progress | closed
- `tasks.status`: open | in-progress | blocked | closed
- `tasks.type`: research | implementation | fix | documentation | maintenance

### Entity Relationships (Core)

See **[docs/schema.md](schema.md)** for the complete ER diagram.

**Core Hierarchy:**
- **Factory** -> Workshop -> Workbench (infrastructure)
- **Commission** -> Shipment -> Task (work tracking)
See `internal/db/schema.sql` for the complete schema

---

## Database System: SQLite

### SQLite (Single Source of Truth)
**Purpose:** Authoritative source for all structured operational data

**Stores:**
- Commissions, shipments, tasks
- Workbenches (git worktrees)
- Tags and task-tag associations
- Current state (status, assignments, timestamps)

**Characteristics:**
- Fast, local, transactional
- Schema-enforced data integrity
- Deterministic queries (e.g., "show all ready tasks")
- Files like config.json are DERIVED from this, never read for decisions

---

## Current Status

ORC is in active production use with the following capabilities:

- **Commission & Workshop Management**: Full lifecycle support
- **Shipment Workflow**: draft -> ready -> in-progress -> closed (manual transitions)
- **Actor Model**: Goblin (coordinator) + IMP (worker via Claude Teams)
- **TMux Integration**: Multi-workbench sessions working
- **Skills System**: Claude Code skills for workflow automation

For current development work, see `orc summary` output.

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

**See**: [Getting Started](getting-started.md) for detailed setup instructions.

### Initialize ORC
```bash
orc init    # Creates ~/.orc/ directory and initializes database
```

### Create Your First Commission
```bash
orc commission create "My Project" --description "Project description"
orc workshop create my-workshop --commission COMM-001
orc workbench create --workshop WORK-001 --repo-id REPO-001
orc shipment create "Initial setup" --commission COMM-001
orc task create "Setup repository" --shipment SHIP-001
```

### Work in a Workbench
```bash
cd ~/src/worktrees/project-main
orc status                    # Shows commission context
orc task claim TASK-001
# ... do work ...
orc task complete TASK-001
```

---

## Architecture Highlights

**What Makes ORC Unique:**

1. **SQLite Source of Truth** - Single authoritative database for all state
2. **Two-Actor Model** - Goblin (coordinator) + IMP (worker via Claude Teams)
4. **Git Worktree Native** - First-class support for isolated workspaces
5. **Simple Lifecycles** - 4-status shipments, 3-status tasks, all manual transitions
6. **TMux Integration** - Smug-based session management with guest pane support
7. **Skills System** - Claude Code skills for workflow automation
8. **Immediate Infrastructure** - Workbench creation is atomic (DB + worktree + config in one shot)

---

## Technical References

**Codebase Structure:**
- `internal/cli/` - Command implementations (commission, task, shipment, workbench, etc.)
- `internal/core/` - Domain logic (guards, planners)
- `internal/app/` - Application services
- `internal/adapters/` - Infrastructure adapters (sqlite, tmux, filesystem)
- `internal/ports/` - Interface definitions
- `internal/db/` - SQLite database setup and schema

**Key Files:**
- `internal/cli/summary.go` - Hierarchical display
- `internal/cli/workbench.go` - Workbench management and TMux integration
- `internal/cli/task.go` - Task CRUD operations
- `internal/cli/shipment.go` - Shipment lifecycle

**External Dependencies:**
- Claude Code CLI (for AI integration)
- Git (worktree support)
- TMux (session management)

---

## Glossary

**Commission**: Top-level coordination scope, owns workshops and shipments
**Workshop**: Collection of workbenches for coordinated work
**Workbench (BENCH-XXX)**: Git worktree registered to a workshop, physical workspace
**Shipment (SHIP-XXX)**: Unit of work with 4-status lifecycle (draft, ready, in-progress, closed)
**Task (TASK-XXX)**: Specific implementation work within a shipment (open, in-progress, closed)
**Goblin**: Coordinator agent -- human's long-running workbench pane
**IMP**: Disposable worker agent spawned by Claude Teams
**orc prime**: Context injection at session start

---

## Contact & Contributions

**Repository**: ~/src/orc
**Documentation**: CLAUDE.md (project context), docs/schema.md (glossary and definitions)
**Status**: Active development, production use
