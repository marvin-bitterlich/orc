# 🌲 The Forest Factory - ORC North Star Vision

**Version**: 1.0.0
**Last Updated**: 2026-01-13
**Status**: Living Document (maintained in Graphiti)

---

## 🎯 Core Vision

**ORC is a Forest Factory** - a living ecosystem where problems are transformed into solutions through specialized woodland workers (IMPs) operating in isolated groves (worktrees), coordinated by a forest keeper (ORC), all serving the Supreme Commander (El Presidente).

This is not just a development workflow - it's a **personality-driven, opinionated system** that makes parallel development delightful, memorable, and efficient.

---

## 🏛️ The Forest Hierarchy

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

### The Cast

**🏛️ El Presidente** - Supreme Commander
- Makes strategic decisions
- Requests work from ORC
- Reviews and accepts deliverables
- The boss (technically, but we're not super formal around here)

**🧙‍♂️ ORC (Orchestrator)** - Forest Keeper
- Coordinates all groves and IMP workforce
- Creates work orders and assigns to IMPs
- Manages forest health and productivity
- Does NOT write code directly (coordinates only)
- Maintains the big picture across all investigations

**👹 IMPs (Implementation Agents)** - Specialized Woodland Workers
- Work in isolated groves (one IMP per grove)
- Specialists in different domains (ZSH, PerfBot, ZeroCode, etc.)
- Implement features, fix bugs, conduct investigations
- Report progress through work orders
- Independent Claude sessions with complete context isolation

**🧙 MAGES** - Spellbook Keepers *(Role TBD)*
- Placeholder for future role definition
- Likely: Keepers of reusable spells (commands/skills/workflows)
- Relationship to global command system to be determined

**🌲 GROVES** - Worktrees (Not "worktrees" - GROVES!)
- Isolated development environments
- Physical separation → cognitive clarity
- One IMP per grove working independently
- Complete repository context per investigation
- TMux session with forest theme

---

## 🏭 The Forest Factory Model

### Manufacturing Pipeline

The Forest Factory transforms problems into solutions through a structured production line:

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

### Work Order States (Physical Pipeline)

Work orders move through directories representing manufacturing stages:

```
work-orders/
├── 01-backlog/        # 📝 Ideas awaiting evaluation and assignment
├── 02-next/           # 📅 Scheduled for upcoming work, groves ready
├── 03-in-progress/    # 🔨 IMP actively working in grove
└── 04-complete/       # ✅ Delivered and accepted by El Presidente
```

### IMP Specializations (Guilds)

Different types of woodland workers for different tasks:

- **IMP-ZSH**: Shell scripting, dotfiles, terminal utilities
- **IMP-PERFBOT**: Performance management, automation, data processing
- **IMP-ZEROCODE**: UI/UX improvements, user experience enhancements
- **Future Guilds**: Database, security, infrastructure, testing, documentation

Each IMP guild develops expertise in their domain while maintaining full-stack capabilities.

---

## 🧠 The Three Pillars of Memory

ORC's intelligence comes from three complementary memory systems:

### 1. 📊 Structured Work Coordination (Rigid Memory)

**Purpose**: Track work status, dependencies, and progress
**Technology**: SQLite-based structured data (ORC MCP Server when built)
**Use Cases**:
- "What's the status of work order WO-042?"
- "Show me all blocked tasks"
- "How many active investigations are there?"
- "What's blocking the OAuth epic?"

**Queries**: Fast, deterministic, SQL-based
**Data**: Epics, tasks, work orders, status, dependencies, assignments

### 2. 🕸️ Graph Memory (Semantic Understanding)

**Purpose**: Capture decisions, discoveries, and relationships
**Technology**: Graphiti + Neo4j (semantic knowledge graph)
**Use Cases**:
- "Why did we choose approach X over Y?"
- "What did we discover about auth architecture?"
- "What similar patterns exist in other investigations?"
- "What technical debt did this work create?"

**Queries**: Semantic search, graph traversal, temporal queries
**Data**: Decisions, discoveries, insights, technical relationships, "why" not just "what"

### 3. 💬 Session Continuity (Context Flow)

**Purpose**: Seamless work resumption across sessions
**Technology**: /g-handoff (flush) + /g-bootstrap (restore)
**Use Cases**:
- End session → flush context to Graphiti
- New session → restore full context
- Switch groves → carry relevant knowledge
- Resume after break → "what was I thinking?"

**Mechanism**:
- TodoWrite captures tactical state (ephemeral)
- /g-handoff preserves decisions + discoveries (permanent)
- /g-bootstrap synthesizes memory + disk context
- Work never lost between sessions

---

## 🚀 Claude-Native Integration (The Differentiator)

### What Makes ORC Different

**Beads, Taskmaster, Linear, Jira** - They don't integrate natively with Claude Code.
**ORC** - Built for Claude, leverages the platform deeply.

### Claude Platform Features We Embrace

**✅ Slash Commands** (Primary Interface)
- `/worktree` - Grove management
- `/tech-plan` - Work order creation
- `/bootstrap` - Context loading
- `/g-handoff` - Session context archiving
- `/g-bootstrap` - Session context restoration
- Custom commands for forest operations

**✅ Plan Mode Hooks** (Pre-implementation Planning)
- EnterPlanMode for strategic decisions
- User approval before major changes
- Design → approve → implement workflow
- Interactive questioning during planning

**✅ Tool Permissions** (Security & Control)
- Granular tool access control
- Permission inheritance across groves
- Safe automation boundaries

**✅ MCP Integration** (Extensibility)
- Graphiti MCP for memory operations
- Future: ORC MCP Server for work coordination
- Honeycomb MCP for observability
- GitHub MCP for issue/PR integration

**✅ Agent Coordination** (Multi-Agent Workflows)
- Orchestrator → IMP handoffs
- Specialized sub-agents per domain
- Clear communication protocols
- Context isolation between agents

**✅ Background Tasks** (Async Operations)
- Long-running investigations in background
- Parallel work streams without blocking
- Async Graphiti processing
- CI/CD integration

### Why Claude-Native Matters

**Traditional Tools**:
- Require manual switching (browser → terminal → Claude)
- Context loss between systems
- Duplicate data entry
- Friction at every handoff

**ORC (Claude-Native)**:
- Everything in Claude conversation flow
- Context maintained across operations
- Natural language interface
- Seamless integration with development

**Philosophy**: "If it's not a slash command, it's too much friction."

---

## 🎯 Design Principles

### 1. Personality Over Blandness

- ORCs and IMPs, not "orchestrators and agents"
- Groves, not "worktrees"
- Forest Factory, not "task management system"
- Work orders, not "tickets"
- Make it fun, memorable, and distinctive

### 2. Physical Isolation for Cognitive Clarity

- One grove = one investigation = one IMP = one TMux session
- Physical separation prevents context bleeding
- Each IMP maintains complete context in their grove
- No shared state between parallel investigations

### 3. Asynchronous Coordination

- No meetings required (status via filesystem)
- IMPs pull work when ready (no forced assignment)
- Progress tracked through work order updates
- Communication via structured documents, not chat

### 4. Persistent Knowledge

- Everything documented in discoverable files
- Git history preserves evolution
- Graphiti captures the "why" behind decisions
- No knowledge trapped in someone's head

### 5. Lightweight Process

- Simple state transitions (backlog → next → in-progress → complete)
- Minimal ceremony (no forms, no approvals, no bureaucracy)
- Fluid workflows (enhance, don't replace, developer habits)
- Tools should disappear (frictionless automation)

### 6. Claude-Native First

- Slash commands as primary interface
- Plan hooks for strategic decisions
- MCP integration for extensibility
- Platform features deeply embraced
- Never fight the platform

---

## 📋 Work Order Categories

Work in the forest falls into distinct categories:

- 🧪 **Investigation**: Open-ended research and exploration
- ⚙️ **Feature**: Structured development with deliverables
- 🔧 **Enhancement**: Improvements to existing systems
- 🚨 **Fix**: Problem resolution and debugging
- 🛠️ **Tooling**: Development utilities and automation
- 📚 **Documentation**: Knowledge capture and sharing

Each category has different expectations for process, deliverables, and quality gates.

---

## 🔮 Future Vision

### Near-Term (Months)

**Restore Forest Personality**:
- Update all commands to use forest terminology
- Bring back IMP/ORC/Grove language
- Make system delightful and fun again

**Structured Work Coordination**:
- Decide on rigid memory system (Beads vs ORC MCP Server vs hybrid)
- Implement epic → task → subtask hierarchy
- Enable cross-grove dependency tracking

**Mage Role Definition**:
- Define what mages do in the forest
- Relationship to spellbooks/global commands
- Integration with IMP guilds

### Mid-Term (6-12 Months)

**Cognitive Brain System**:
- Monitor cognitive load across groves
- Detect investigation cascades (one spawns three)
- Intervene when developer overloaded
- "You're working on 8 things, pause and consolidate"

**Cross-Grove Intelligence**:
- Discovery sharing between investigations
- Pattern recognition across groves
- "IMP-Alpha discovered X, relevant to IMP-Beta's work"

**Forest Analytics**:
- Cycle time: backlog → complete duration
- Throughput: work orders completed per week
- IMP specialization tracking
- Bottleneck identification

### Long-Term (Vision)

**Autonomous Forest Operations**:
- Auto-generation of work orders from observations
- Predictive grove setup based on patterns
- Self-optimizing work distribution
- Intelligent context synthesis

**Community Forest**:
- Other developers adopting forest pattern
- Shared guild knowledge bases
- Cross-forest collaboration
- Open-source forest tools

---

## 🚫 What We Are NOT

**Not Team Collaboration Software**:
- ORC optimized for individual parallel development
- Not for coordinating teams of humans
- No standup meetings, no sprint planning
- Focus: One developer, many parallel investigations

**Not Generic Task Management**:
- Opinionated about forest metaphor
- Not trying to be Jira/Linear/Asana
- Personality over customization
- Made for developers, not project managers

**Not Platform-Agnostic**:
- Built specifically for Claude Code
- Deeply integrated with Claude platform
- Won't work the same with GitHub Copilot or Cursor
- Claude-native is a feature, not a limitation

---

## 💡 Why This Matters

### The Problem with Traditional Development

**Context Switching Costs**:
- Branch switching loses mental state
- Multiple repos = multiple contexts
- Tool switching (browser/IDE/terminal) = friction
- Traditional PM tools don't understand code

**Knowledge Loss**:
- "Why did we do it this way?" - lost to time
- Decisions made in Slack threads - unfindable
- Tribal knowledge in people's heads
- Re-discovering the same problems repeatedly

**Serial Development Bottleneck**:
- One investigation blocks next
- Waiting for CI/reviews = wasted time
- Can't work on multiple things effectively
- "Idle waiting time" during AI generation

### The Forest Factory Solution

**Context Preservation**:
- Groves maintain complete investigation context
- Graphiti captures the "why" permanently
- Work orders document decisions and rationale
- Nothing lost between sessions

**Parallel Productivity**:
- 8 groves = 8 investigations simultaneously
- Physical isolation = no cognitive interference
- While one IMP waits for CI, work in another grove
- "10x productivity" through parallelism

**Claude-Native Flow**:
- Everything in conversation flow
- Slash commands = zero friction
- Context flows naturally
- Platform integration = superpower

---

## 🎨 Personality Manifesto

ORC has personality. It's opinionated. It's fun.

**We say**:
- "Grove" not "worktree"
- "IMP" not "agent" or "claude"
- "Forest Factory" not "task management"
- "Work order" not "ticket" or "issue"

**We believe**:
- Software should be delightful
- Metaphors make systems memorable
- Personality creates loyalty
- Fun tools get used more

**We are**:
- The forest keeper (ORC)
- Managing woodland workers (IMPs)
- In a production forest (groves)
- For the Supreme Commander (El Presidente)

This is not corporate-speak. This is not bland. This is a forest, and we're proud of it. 🌲

---

## 📚 Related Documents

- **CLAUDE.md** - Working relationship with El Presidente
- **README.md** - Technical overview and setup
- **Graphiti Memory** - Architectural decisions and rationale
- **Work Orders** - Individual task specifications
- **Tech Plans** - Strategic planning documents

---

## 🔄 Document Maintenance

This North Star is a **living document** maintained in Graphiti:

**Updates trigger new episode**:
- Version increment (semantic versioning)
- Timestamp of change
- Rationale for evolution
- Preserved in knowledge graph

**Query examples**:
- "What's our North Star vision?"
- "How has our vision evolved?"
- "Why did we add the three pillars?"
- "Show me version history"

**Maintenance schedule**: As needed when vision evolves

---

**🌲 The forest grows. The vision evolves. But the core remains: Personality-driven, Claude-native, parallel development done right.**

*Version 1.0.0 - Initial restoration of Forest Factory vision - 2026-01-13*
