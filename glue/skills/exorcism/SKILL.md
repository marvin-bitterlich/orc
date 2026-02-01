---
name: exorcism
description: Ledger maintenance skill for cleaning entropy and converging exploration into work. Use when user says /exorcism or wants to tidy a conclave, consolidate notes, or synthesize exploration into a spec/shipment.
---

# Exorcism Skill

Ledger maintenance for cleaning up entropy, consolidating ideas, and maintaining semantic health.

## Objectives

| Objective | Purpose | Creates |
|-----------|---------|---------|
| **Clean** | Tidy existing container | Nothing new (maybe exorcism note) |
| **Ship** | Converge exploration into work | Draft shipment + spec |

### Selection

Explicit upfront or agent-proposed after survey:

```bash
/exorcism --clean CON-018
/exorcism --ship CON-018
```

Or let the agent propose after survey:
```
Agent: Surveyed CON-018. State: Chaotic.
       Recommend: Clean first, then ship?
       [c]lean / [s]hip / [b]oth sequentially
```

## Workflow

### 1. Discovery

- Analyze current focus (conclave, tome, or shipment)
- Produce high-level summary
- Identify themes (areas to dig into)
- Present to human with objective recommendation

### 2. Theme Selection

- Human picks a theme
- Agent enters interview mode

### 3. Interview (per theme)

- Max 5 questions
- Progress indicator: "2/3 questions remaining"
- Each question: context + why it matters + choices
- Choices map to patterns

## Patterns

| Pattern | When | Move | Objective |
|---------|------|------|-----------|
| SYNTHESIZE | Multiple notes → one conclusion | Combine into decision/spec | Ship |
| EXTRACT-LAYER | Note mixes C4 levels | Split by layer | Both |
| CLOSE-SUPERSEDED | Content now in better artifact | Close original | Both |
| CONSOLIDATE-DUPLICATES | Same concept, different words | Merge | Clean |
| PROMOTE-TO-DECISION | Implicit decision buried | Extract to decision note | Ship |
| BRIDGE-CONTEXT | Orphan L3/L4 detail | Add reference to L1/L2 | Clean |
| DEFER-TO-LIBRARY | Valid but not now | Park | Clean |
| SPLIT-SCOPE | Kitchen-sink container | Split into focused pieces | Clean |

## Commands

Pattern execution uses CLI operations:

```bash
orc note merge <source> <target>
orc note close <id> --reason <reason> [--by <note-id>]
```

**Reason vocabulary:** superseded, synthesized, resolved, deferred, duplicate, stale

## Exorcism Note

Each maintenance session can produce an `exorcism` note as a record:

```markdown
# Exorcism: CON-018 Consolidation

## Before
- 4 tomes, 17 notes scattered
- Multiple overlapping specs

## After
- All tomes closed
- Unified spec (NOTE-311)

## Key decisions
- Single command: /exorcism
- Two objectives: clean vs ship
```

## Reference

See NOTE-311 for full specification.
