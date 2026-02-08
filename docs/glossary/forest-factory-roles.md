# Forest Factory Roles & Concepts

**Purpose**: Personality-driven metaphors and role definitions
**Status**: Living document

---

## The Cast

### El Presidente - Supreme Commander

**Role**: Strategic decision maker and boss

**Responsibilities**:
- Makes strategic decisions
- Requests work from agents
- Reviews and accepts deliverables
- Ultimate authority (though we're not super formal around here)

**In Practice**:
- The human developer
- Initiates commissions and shipments
- Approves plans and escalations
- Commands the forest

---

### IMP - Implementation Agent (Specialized Woodland Worker)

**Role**: Code implementation and technical work in workbenches

**Characteristics**:
- Work in isolated workbenches (one IMP per workbench)
- Implement features, fix bugs, conduct research
- Report progress through tasks and receipts
- Independent Claude sessions with complete context isolation

**Place**: Workbench (BENCH-XXX)

**In Practice**:
- Implementation Claude instances
- Work in ~/wb/[workbench-name] directories
- Each has dedicated TMux window
- Focus on implementation, not coordination

**Workflow**:
1. `orc prime` - Get context
2. `/imp-start` - Claim task
3. `/imp-plan-create` - Research and plan
4. Implement changes
5. `/imp-rec` - Complete and chain to next

---

### Goblin - Workshop Gatekeeper

**Role**: Review and escalation handling at workshop level

**Characteristics**:
- Reviews IMP plans before approval
- Handles escalations requiring judgment
- Coordinates across workbenches in a workshop
- Does NOT write code directly

**Place**: Gatehouse (GATE-XXX)

**In Practice**:
- Coordination Claude instance
- Works in ~/.orc/ws/[workshop-slug]/ directory
- Reviews via `/goblin-escalation-receive`
- Autonomous for clear-cut decisions
- Waits for El Presidente on judgment calls

**Autonomy Boundaries**:
- Reject plans that violate CLAUDE.md checklists
- Approve plans that follow documented patterns
- Wait for human input on architectural decisions

---

### Watchdog - IMP Monitor

**Role**: Monitor IMP progress and health

**Characteristics**:
- Tracks IMP activity and progress
- Reports anomalies
- One watchdog per workbench

**Place**: Watchdog (WATCH-XXX)

---

### Workbench - Worktree (Isolated Development Environment)

**Role**: Physical workspace for implementation work

**Characteristics**:
- Isolated development environments
- Physical separation → cognitive clarity
- One IMP per workbench working independently
- Complete repository context per workbench
- TMux window with 3-pane layout: vim | claude | shell

**Terminology**:
- We say "Workbench" not "worktree"
- Workbenches are named descriptively (e.g., `orc-044`)
- Located in ~/wb/ or ~/src/worktrees/

**In Practice**:
- Git worktree on disk
- Has dedicated TMux window in workshop session
- Linked to workshop in database
- Contains .orc/config.json with place_id

---

## Place-Based Actor Model

ORC uses a place-based actor model where identity is tied to "where you are":

```
Place Hierarchy:

Workshop (WORK-XXX)
├── Gatehouse (GATE-XXX) ← Goblin sits here
└── Workbenches (BENCH-XXX) ← IMPs sit here
    └── Watchdog (WATCH-XXX) ← Monitors IMP
```

**Identity from Place**:
- `BENCH-` prefix → IMP role
- `GATE-` prefix → Goblin role
- `WATCH-` prefix → Watchdog role

**Config Format**:
```json
{
  "version": "1.0",
  "place_id": "BENCH-044"
}
```

---

## Shipment Workflow

The manufacturing pipeline for exploration → implementation:

```
exploring (messy notes)
  → /ship-synthesize → Summary note (knowledge compaction)
  → /ship-plan → Tasks (C2/C3 engineering review)
  → /imp-plan-create → Plans (C4 file-level detail)
  → Implementation → Code
  → /imp-rec → Receipt and task completion
```

---

## Terminology Standards

**We say**:
- "Workbench" not "worktree"
- "IMP" not "agent"
- "Shipment" not "ticket" or "issue"
- "Task" not "work order"
- "Goblin" not "orchestrator"

**Philosophy**: "Personality over blandness"

---

**Last Updated**: 2026-02-08
**Status**: Reference documentation
