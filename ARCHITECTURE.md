# ORC 2.0 - System Overview

**Generated:** 2026-01-14
**Source:** Oracle synthesis from knowledge graph (Graphiti)

---

## Executive Summary

ORC (Orchestrator) is a mission coordination system for managing complex, multi-repository software development work. It combines traditional task management with AI-powered context preservation, enabling seamless handoffs between human operators and Claude AI agents across multiple workspaces.

**Key Innovation:** ORC maintains two complementary knowledge systems - SQLite for structured operational data and Graphiti for semantic memory and design decisions - allowing both deterministic queries and intelligent pattern synthesis.

---

## Core Architecture: "Forest Factory Model"

ORC 2.0 uses a simplified, flat hierarchy optimized for real-world workflows:

```
Mission (coordination scope)
├── Work Orders (tasks, flat structure with optional parent/child grouping)
└── Groves (git worktrees - isolated workspaces with IMP agents)
```

**Design Principles:**
- **Simplicity Over Hierarchy** - Removed Operations and Expeditions layers
- **Mission-Centric Organization** - Mission is the coordination boundary
- **Database as Source of Truth** - Single authoritative data source (SQLite)
- **Physical Reality Wins** - Model matches actual workflow, not idealized process

---

## Core Features

### 1. Mission Management
- **Mission**: Top-level coordination scope (e.g., "Sidekiq Deprecation", "Auth Refactor")
- Owns multiple Groves (workspaces) and Work Orders (tasks)
- Each mission can have a dedicated TMux session with IMP agents in groves
- Support for both ORC-development missions and application-code missions

### 2. Work Order System
**Flat hierarchy with Epic/Parent grouping:**
- Work orders belong directly to missions (no forced intermediate layers)
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

### 3. Grove Management (Git Worktree Integration)
**Grove**: An isolated git worktree for a mission, registered in the database

- Groves belong to missions, not individual work orders
- Multiple groves per mission (e.g., backend, frontend, api repos)
- Work orders are assigned to groves (via assigned_grove_id)
- IMP (Implementation) = Grove (conceptual equivalence, no separate entity)

**Key Commands:**
```bash
orc grove create [name] --repos main-app --mission MISSION-001
orc grove list [--mission MISSION-XXX]
orc grove show GROVE-XXX
orc grove rename GROVE-XXX new-name
orc grove open GROVE-XXX      # Opens in new TMux window with IMP layout
```

**Grove Features:**
- Creates git worktree automatically
- Writes metadata.json to .orc/ subdirectory (reference only)
- Writes .orc-mission marker for context detection
- Opens in TMux with 3-pane IMP layout: vim | claude | shell

### 4. Two-Tier Agent Architecture (ORC + IMPs)
**Concept**: Simple two-tier coordination with one orchestrator and multiple implementation agents

**Agent Types:**
- **ORC (Orchestrator)**: Master coordinator that manages missions, creates groves, assigns epics
- **IMP (Implementation Agent)**: Works within a grove on assigned epics/tasks

**Architecture Principles:**
- Single ORC orchestrator
- IMPs operate within grove boundaries
- ORC coordinates via mail system and epic assignments
- IMPs can create groves/epics but not missions

**Mail System:**
- Direct IMP ↔ ORC communication
- Work orders/tasks used for coordination

### 5. Context Preservation & Handoffs
**g-bootstrap & g-handoff Integration:**

Session boundaries are preserved through:
- **Handoffs**: Narrative summaries stored in SQLite + Graphiti
- **g-bootstrap**: Restores full context at session start (git state, Graphiti memories, ledger)
- **g-handoff**: Captures session work, decisions, discoveries at session end

**Key Features:**
- Cross-session memory preservation
- New Claude sessions start with full historical context
- Zero "cold start" problem
- Episodes stored in Graphiti with semantic searchability

### 6. TMux Integration
**One TMux session per mission:**
```
TMux Session: "Mission Name" (orc-MISSION-XXX)
├── Pane 0: ORC (coordination)
├── Pane 1: IMP in grove-1 (vim)
├── Pane 2: IMP in grove-1 (claude)
└── Pane 3: IMP in grove-1 (shell)
```

**Features:**
- `orc grove open GROVE-XXX` creates new window with IMP layout
- All panes CD into grove directory
- Easy context switching between coordination and implementation

**Agent Starting Pattern:**
ORC uses **direct prompt injection** when starting Claude agents in TMux:

```bash
claude "Run `orc prime`"
```

This pattern replaces SessionStart hooks (which are broken in Claude Code v2.1.7). When agents start:
1. TMux sends the command with prompt: `claude "Run \`orc prime\`"`
2. Claude receives the prompt and executes `orc prime`
3. `orc prime` detects the agent's location (grove/mission/global) and provides appropriate context
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
- **Primary Database**: SQLite (structured operational data)
- **Semantic Memory**: Graphiti (knowledge graph, design decisions, patterns)
- **Version Control**: Git (with worktree integration)
- **Session Management**: TMux
- **AI Integration**: Claude API (via Claude Code CLI)

### Database Schema (SQLite)

**Core Tables:**
- `missions` - Top-level coordination scopes
- `work_orders` - Tasks with status, type, parent_id, assigned_grove_id, pinned flag
- `groves` - Git worktrees registered to missions
- `handoffs` - Session handoff narratives with Graphiti episode UUIDs

**Removed Tables (simplified in 2.0):**
- ~~`operations`~~ - Removed (too rigid, use work_order.type instead)
- ~~`expeditions`~~ - Removed (1:1:1 mapping didn't match workflow)
- ~~`dependencies`~~ - Not implemented yet (can add later if needed)

**Key Fields:**
- `work_orders.status`: ready | design | implement | deploy | blocked | paused | complete
- `work_orders.type`: research | implementation | fix | documentation | maintenance
- `work_orders.parent_id`: Optional epic/parent reference
- `work_orders.assigned_grove_id`: Grove assignment
- `work_orders.pinned`: Boolean for visibility

---

## Dual Database System: SQLite + Graphiti

### SQLite (Operational Database)
**Purpose:** Single source of truth for structured operational data

**Stores:**
- Missions, work orders, groves, handoffs
- Current state (status, assignments, timestamps)
- Deterministic queries (e.g., "show all ready work orders")

**Characteristics:**
- Fast, local, transactional
- Schema-enforced data integrity
- Authoritative for current state
- Files like metadata.json are DERIVED from this, never read for decisions

### Graphiti (Semantic Memory / Knowledge Graph)
**Purpose:** AI-accessible semantic memory and pattern synthesis

**Stores:**
- Design decisions with rationale and alternatives considered
- Session handoffs (narrative summaries)
- Technical patterns and preferences
- Entity relationships (WHO decided WHAT, WHY)
- Discovery insights and open questions

**Characteristics:**
- Graph-based (nodes + facts + episodes)
- Semantic search capabilities
- Cross-session context preservation
- Enables Oracle pattern queries

### How They Interact

**Data Flow:**
```
Session Work → SQLite (operational state)
            ↓
         Handoff (at session end)
            ↓
    Graphiti Episode (narrative + decisions)
            ↓
    Knowledge Graph (nodes, facts, relationships)
            ↓
    Oracle Queries (pattern synthesis)
```

**Complementary Roles:**
- **SQLite answers**: "What work orders are ready?" (deterministic)
- **Graphiti answers**: "What design patterns does El Presidente prefer?" (synthesized)

**Bootstrap Process:**
```
New Session starts
    ↓
g-bootstrap queries:
    - SQLite: Current missions, work orders, groves
    - Graphiti: Recent handoffs, design decisions, open questions
    - Git: Current branch, uncommitted changes
    ↓
Full context restored → Work begins
```

**Handoff Process:**
```
Session ending
    ↓
g-handoff creates:
    - Handoff record in SQLite (work_order_ids, timestamp)
    - Episode in Graphiti (narrative, decisions, discoveries)
    - Knowledge graph updates (nodes, facts)
    ↓
Next session can bootstrap from this
```

**Key Insight:** SQLite tells you WHERE YOU ARE, Graphiti tells you HOW YOU GOT HERE and WHY.

---

## Roadmap & Future Features

### Active Development (In Progress)

**WO-030: Oracle Integration** ⚡ PRIORITY
- Design pattern lookup interface
- Query Graphiti for "what would El Presidente want?"
- Return decisions with rationale + references
- Approval workflow for new patterns
- **Status**: Early experimentation showing promising results
- **Recent Success**: Pattern synthesis working ("dynamite" - El Presidente)

**WO-021: Core Architecture Design**
- Cross-grove coordination mechanisms (WO-016)
- Grove lifecycle management (WO-015)
- Work order state transition model (WO-014)
- Inter-agent mail system refinement

**WO-057: Factory Production Line Issues** (Master ORC Tooling)
- Continuous improvements to CLI tools
- Bug fixes and UX refinements
- IMP escalations from MISSION-002

### Completed Major Features (2.0)

✅ **Forest Factory Architecture Simplification**
- Removed Operations and Expeditions layers
- Flat Mission → Work Order structure
- Type-based categorization

✅ **Epic/Parent Work Order Hierarchy**
- Single-parent grouping
- Visual hierarchy in summary output
- Nested display with indentation

✅ **Status Field Consolidation**
- Merged status + phase into single field
- 7 semantic states (ready → complete)
- Emoji indicators for quick scanning

✅ **IMP Agent Deployment (MISSION-002)**
- First real-world mission validated
- TMux automation working
- ORC ↔ IMP communication working
- **Validation**: Everything worked first try

✅ **Grove Management**
- Create, list, rename, open commands
- Git worktree integration
- TMux window spawning (orc grove open)
- Mission context detection

✅ **Pinned Work Orders**
- Pin/unpin commands
- Inline emoji display (no separate section)
- Priority visibility

✅ **Work Order Update**
- Update title and/or description
- Set/change parent relationships

### Near-Term Roadmap

**Phase 1: Oracle Productionization**
- Build full Oracle query interface
- Seed Graphiti with existing patterns
- Integrate with CLI commands
- Add approval workflow

**Phase 2: Cross-Mission Coordination**
- Refine mail system
- IMP → ORC escalation workflow
- ORC → IMP directive system
- Async communication patterns

**Phase 3: Discovery Sharing**
- Cross-grove discovery propagation
- Mission-scoped knowledge sharing
- Graphiti group_id segregation
- Context inheritance patterns

**Phase 4: Workflow Automation**
- Work order state transitions (auto/manual rules)
- Grove lifecycle automation
- TMux session management
- Handoff automation improvements

### Long-Term Vision

**Factory-Themed Terminology**
- Replace generic terms (Mission, Epic, Operation)
- Restore forest/factory metaphor consistently
- Personality + functionality

**Advanced Oracle**
- Proactive pattern suggestions
- Sentiment tracking (capture "cool", "dynamite" reactions)
- Pattern evolution over time
- Contradiction detection

**Multi-Agent Orchestration**
- ORC coordinates multiple IMPs across groves
- IMP-to-IMP communication via ORC
- Mission-based coordination structure
- Resource allocation and prioritization

---

## Success Metrics

**First-Try Execution:**
- MISSION-002 deployment: ✅ Worked on first attempt
- ORC + IMP coordination: ✅ Validated in real operation
- TMux automation: ✅ Clean integration
- Proto-mail system: ✅ Bidirectional communication working

**Oracle Validation:**
- Pattern synthesis: ✅ "Dynamite" results from 15+ facts
- Cross-session queries: ✅ Knowledge spanning multiple days
- Design preference extraction: ✅ Accurate patterns identified

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
      "~/src/missions"
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

### Create Your First Mission
```bash
orc mission create "My Project" --description "Project description"
orc work-order create "Setup repository" --mission MISSION-001
orc grove create project-main --repos my-repo --mission MISSION-001
```

### Work in a Grove
```bash
cd ~/src/worktrees/project-main
orc status                    # Shows mission context
orc work-order claim WO-001
# ... do work ...
orc work-order complete WO-001
```

### Session Boundaries
```bash
# At session start
/g-bootstrap                  # Restores full context

# At session end
/g-handoff                    # Captures session to Graphiti
```

---

## Architecture Highlights

**What Makes ORC Unique:**

1. **Dual Knowledge Systems** - Deterministic (SQLite) + Semantic (Graphiti)
2. **Zero Cold-Start** - Full context preservation across sessions
3. **Multi-Agent Coordination** - ORC/IMP architecture with mail system
4. **Git Worktree Native** - First-class support for isolated workspaces
5. **Oracle Pattern System** - AI queries historical design preferences
6. **TMux Integration** - One session per mission, programmatic layout
7. **Flat But Structured** - Simple hierarchy with optional grouping
8. **Validation-Driven** - Proven by first-try production deployment

---

## Technical References

**Codebase Structure:**
- `internal/cli/` - Command implementations (mission, work-order, grove, etc.)
- `internal/models/` - Database models and queries
- `internal/context/` - Mission context detection
- `internal/tmux/` - TMux session management
- `internal/db/` - SQLite database setup and migrations

**Key Files:**
- `internal/cli/summary.go` - Hierarchical work order display
- `internal/cli/grove.go` - Grove management and TMux integration
- `internal/cli/work_order.go` - Work order CRUD operations
- `internal/models/workorder.go` - Work order model and queries

**External Dependencies:**
- Graphiti (MCP server for semantic memory)
- Claude Code CLI (g-bootstrap, g-handoff commands)
- Git (worktree support)
- TMux (session management)

---

## Glossary

**Mission**: Top-level coordination scope, owns groves and work orders
**Work Order**: Task within a mission, flat hierarchy with optional parent
**Grove**: Git worktree registered to a mission, physical workspace
**IMP**: Implementation (conceptual layer over grove, not separate entity)
**ORC**: Main orchestrator that coordinates IMPs across missions
**Proto-Mail**: Work-order-based bidirectional messaging (WO-061 ↔ WO-065)
**Oracle**: Design pattern lookup interface querying Graphiti
**Handoff**: Session boundary artifact (narrative + work state)
**g-bootstrap**: Context restoration at session start
**g-handoff**: Context capture at session end
**Epic**: Work order with children (parent in hierarchy)

---

## Contact & Contributions

**Repository**: ~/src/orc
**Documentation**: CLAUDE.md (project context), glossary/ (definitions)
**Knowledge Base**: Graphiti (group_id: "orc")
**Status**: Active development, production-validated (MISSION-002)

---

*This document was synthesized from 50+ knowledge graph facts, 20+ design episodes, and operational data from ORC's dual database system. Generated using Oracle pattern synthesis capabilities.*
