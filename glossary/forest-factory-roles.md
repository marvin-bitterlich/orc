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
- Coordinates all groves and IMP workforce
- Creates work orders and assigns to IMPs
- Manages forest health and productivity
- **Does NOT write code directly** (coordinates only)
- Maintains the big picture across all investigations

**In Practice**:
- The orchestrator Claude instance
- Works in ~/src/orc directory
- Creates and manages ledger entries
- Sets up groves for investigations
- Provides handoff context to IMPs

**Safety Boundaries**:
If asked for direct code changes or debugging:
> "El Presidente, I'm the Orchestrator - I coordinate investigations but don't work directly on code. Switch to the `[investigation-name]` TMux window to work with the investigation-claude on that technical task."

---

### 👹 IMP - Implementation Agent (Specialized Woodland Worker)

**Role**: Code implementation and technical work in groves

**Characteristics**:
- Work in isolated groves (one IMP per grove)
- Specialists in different domains (ZSH, PerfBot, ZeroCode, etc.)
- Implement features, fix bugs, conduct investigations
- Report progress through work orders
- Independent Claude sessions with complete context isolation

**IMP Guilds** (Specializations):
- **IMP-ZSH**: Shell scripting, dotfiles, terminal utilities
- **IMP-PERFBOT**: Performance management, automation, data processing
- **IMP-ZEROCODE**: UI/UX improvements, user experience enhancements
- **Future Guilds**: Database, security, infrastructure, testing, documentation

**In Practice**:
- Investigation Claude instances
- Work in ~/src/worktrees/[grove-name] directories
- Each has dedicated TMux session
- Focus on implementation, not coordination

**Questions/Contentious**:
- **CRITICAL**: Is IMP an entity we track in the database?
- Is IMP just a string identifier ("IMP-ZSH")?
- Do we need an `imps` table to track capabilities/specializations?
- Or is IMP just a role that gets assigned to expeditions?

---

### 🧙 MAGE - Spellbook Keeper

**Role**: ??? TBD - Placeholder for future definition ???

**Possibilities**:
- Keepers of reusable spells (commands/skills/workflows)
- Relationship to global command system
- Could manage /skills or /global-commands
- Might handle cross-grove knowledge transfer

**Status**: Role not yet defined

**Questions/Contentious**:
- Do we need this role?
- Should global commands be "spells" managed by mages?
- Is this premature complexity?

---

### 🌲 GROVE - Worktree (Isolated Development Environment)

**Role**: Physical workspace for investigations

**Characteristics**:
- Isolated development environments
- Physical separation → cognitive clarity
- One IMP per grove working independently
- Complete repository context per investigation
- TMux session with forest theme

**Terminology**:
- We say "Grove" not "worktree"
- Groves are named descriptively (e.g., `ml-auth-refactor`)
- Located in ~/src/worktrees/

**In Practice**:
- Git worktree on disk
- Has dedicated TMux session
- Linked to expedition in database
- Contains CLAUDE.md for IMP instructions

**Questions/Contentious**:
- Should groves be auto-created when expedition starts?
- Can a grove exist without an expedition?
- When do groves get archived?

---

## Conceptual Hierarchy

From NORTH_STAR.md:

```
🏛️ EL PRESIDENTE (Supreme Commander)
    │
    └── 🧙‍♂️ ORC (Orchestrator - Forest Keeper)
            │
            ├── 👹 IMPs (Implementation Agents - Woodland Workers)
            │   ├── IMP working in Grove A
            │   ├── IMP working in Grove B
            │   └── IMP working in Grove C
            │
            ├── 🧙 MAGES (Spellbook Keepers - TBD)
            │   └── [Placeholder - role to be defined]
            │
            └── 🌲 GROVES (Worktrees - Isolated Development Environments)
                ├── Grove: ml-feature-alpha
                ├── Grove: ml-bugfix-beta
                └── Grove: ml-investigation-gamma
```

---

## Work Order States (Manufacturing Pipeline)

From NORTH_STAR.md - Physical directory structure:

```
work-orders/
├── 01-backlog/        # 📝 Ideas awaiting evaluation and assignment
├── 02-next/           # 📅 Scheduled for upcoming work, groves ready
├── 03-in-progress/    # 🔨 IMP actively working in grove
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
      Grove Setup →
        IMP Assignment →
          Implementation →
            Quality Check →
              Delivery to El Presidente
```

**Questions/Contentious**:
- Where does "Expedition" fit in this pipeline?
- Is "Grove Setup → IMP Assignment → Implementation" = "Expedition"?
- Or is Expedition separate from this flow?

---

## Terminology Standards

**We say** (from NORTH_STAR Personality Manifesto):
- "Grove" not "worktree"
- "IMP" not "agent" or "claude"
- "Forest Factory" not "task management"
- "Work order" not "ticket" or "issue"

**Philosophy**: "Personality over blandness"

---

## Summary of CRITICAL Questions

1. **IMP as Entity**: Should IMPs be tracked in database, or just string identifiers?
2. **Grove Lifecycle**: When are groves created relative to expeditions/work orders?
3. **Mage Role**: Do we need this? What would it do?
4. **Work Order Directories**: Should we use filesystem directories or database fields for state?
5. **Pipeline vs Reality**: How does the conceptual pipeline map to actual ledger entities?

---

**Last Updated**: 2026-01-13
**Status**: Initial dump from NORTH_STAR.md - needs discussion
