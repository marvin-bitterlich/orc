# Forest Factory Roles & Concepts

**Source**: NORTH_STAR.md
**Purpose**: Personality-driven metaphors and role definitions
**Status**: Living document

---

## The Cast

### 🏛️ El Presidente - Supreme Commander

**Role**: Strategic decision maker and boss

**Responsibilities**:
- Makes strategic decisions
- Requests work from ORC
- Reviews and accepts deliverables
- Ultimate authority (though we're not super formal around here)

**In Practice**:
- The human developer
- Initiates missions and operations
- Approves work before implementation
- Commands the forest

---

### 🧙‍♂️ ORC - Orchestrator (Forest Keeper)

**Role**: Coordination layer between El Presidente and implementation work

**Responsibilities**:
- Coordinates all workbenches and IMP workforce
- Creates work orders and assigns to IMPs
- Manages forest health and productivity
- **Does NOT write code directly** (coordinates only)
- Maintains the big picture across all work

**In Practice**:
- The orchestrator Claude instance
- Works in ~/src/orc directory
- Creates and manages ledger entries
- Sets up workbenches for commissions
- Provides handoff context to IMPs

**Safety Boundaries**:
If asked for direct code changes or debugging:
> "El Presidente, I'm the Orchestrator - I coordinate commissions but don't work directly on code. Switch to the workbench's TMux window to work with the IMP on that technical task."

---

### 👹 IMP - Implementation Agent (Specialized Woodland Worker)

**Role**: Code implementation and technical work in workbenches

**Characteristics**:
- Work in isolated workbenches (one IMP per workbench)
- Specialists in different domains (ZSH, PerfBot, ZeroCode, etc.)
- Implement features, fix bugs, conduct research
- Report progress through work orders
- Independent Claude sessions with complete context isolation

**IMP Guilds** (Specializations):
- **IMP-ZSH**: Shell scripting, dotfiles, terminal utilities
- **IMP-PERFBOT**: Performance management, automation, data processing
- **IMP-ZEROCODE**: UI/UX improvements, user experience enhancements
- **Future Guilds**: Database, security, infrastructure, testing, documentation

**In Practice**:
- Implementation Claude instances
- Work in ~/src/worktrees/[workbench-name] directories
- Each has dedicated TMux session
- Focus on implementation, not coordination

**Questions/Contentious**:
- **CRITICAL**: Is IMP an entity we track in the database?
- Is IMP just a string identifier ("IMP-ZSH")?
- Do we need an `imps` table to track capabilities/specializations?
- Or is IMP just a role that gets assigned to workbenches?

---

### 🌲 WORKBENCH - Worktree (Isolated Development Environment)

**Role**: Physical workspace for implementation work

**Characteristics**:
- Isolated development environments
- Physical separation → cognitive clarity
- One IMP per workbench working independently
- Complete repository context per workbench
- TMux session with forest theme

**Terminology**:
- We say "Workbench" not "worktree"
- Workbenches are named descriptively (e.g., `ml-auth-refactor`)
- Located in ~/src/worktrees/

**In Practice**:
- Git worktree on disk
- Has dedicated TMux session
- Linked to commission in database
- Contains CLAUDE.md for IMP instructions

**Questions/Contentious**:
- Should workbenches be auto-created when commission starts?
- Can a workbench exist without a commission?
- When do workbenches get archived?

---

## Conceptual Hierarchy

From NORTH_STAR.md:

```
🏛️ EL PRESIDENTE (Supreme Commander)
    │
    └── 🧙‍♂️ ORC (Orchestrator - Forest Keeper)
            │
            ├── 👹 IMPs (Implementation Agents - Woodland Workers)
            │   ├── IMP working in Workbench A
            │   ├── IMP working in Workbench B
            │   └── IMP working in Workbench C
            │
            └── 🌲 WORKBENCHES (Worktrees - Isolated Development Environments)
                ├── Workbench: ml-feature-alpha
                ├── Workbench: ml-bugfix-beta
                └── Workbench: ml-research-gamma
```

---

## Work Order States (Manufacturing Pipeline)

From NORTH_STAR.md - Physical directory structure:

```
work-orders/
├── 01-backlog/        # 📝 Ideas awaiting evaluation and assignment
├── 02-next/           # 📅 Scheduled for upcoming work, workbenches ready
├── 03-in-progress/    # 🔨 IMP actively working in workbench
└── 04-complete/       # ✅ Delivered and accepted by El Presidente
```

**Questions/Contentious**:
- Current database has statuses: 'backlog', 'next', 'in_progress', 'complete'
- These don't match the directory structure (01-backlog, 02-next, etc.)
- Should we use filesystem directories OR database fields?
- Or keep them synchronized?

---

## Forest Factory Manufacturing Pipeline

From NORTH_STAR.md:

```
El Presidente Request →
  ORC Evaluation →
    Work Order Creation →
      Workbench Setup →
        IMP Assignment →
          Implementation →
            Quality Check →
              Delivery to El Presidente
```

**Questions/Contentious**:
- How do Workbenches fit in this pipeline?
- Is "Workbench Setup → IMP Assignment → Implementation" the core workflow?

---

## Terminology Standards

**We say** (from NORTH_STAR Personality Manifesto):
- "Workbench" not "worktree"
- "IMP" not "agent" or "claude"
- "Forest Factory" not "task management"
- "Work order" not "ticket" or "issue"

**Philosophy**: "Personality over blandness"

---

## Summary of CRITICAL Questions

1. **IMP as Entity**: Should IMPs be tracked in database, or just string identifiers?
2. **Workbench Lifecycle**: When are workbenches created relative to commissions/work orders?
3. **Work Order Directories**: Should we use filesystem directories or database fields for state?
4. **Pipeline vs Reality**: How does the conceptual pipeline map to actual ledger entities?

---

**Last Updated**: 2026-01-13
**Status**: Initial dump from NORTH_STAR.md - needs discussion
