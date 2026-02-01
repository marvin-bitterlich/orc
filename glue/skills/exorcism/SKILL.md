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

## Survey Flow

When `/exorcism` is invoked, follow these steps:

### Step 1: Identify Target

If argument provided (e.g., `/exorcism CON-018`):
- Use the provided container ID

If no argument:
```bash
orc status
```
- Use the focused container from output
- If no focus: "No target specified. Which container should I survey? (CON-xxx, SHIP-xxx, or TOME-xxx)"

### Step 2: Collect Data

For conclave:
```bash
orc conclave show CON-xxx
orc note list --conclave CON-xxx
```

For shipment:
```bash
orc shipment show SHIP-xxx
orc note list --shipment SHIP-xxx
```

For tome:
```bash
orc tome show TOME-xxx
orc note list --tome TOME-xxx
```

### Step 3: Analyze

From the collected data, compute:
- Total note count
- Notes grouped by type (idea, question, concern, etc.)
- Notes grouped by status (open vs closed)
- Open questions (type=question, status=open)
- Unaddressed concerns (type=concern, status=open)
- Potential duplicates (notes with very similar titles)
- Stale notes (status=open but old created_at, no recent updated_at)

### Step 4: Assess State

**Chaotic** if any of:
- 3+ open questions
- 2+ unaddressed concerns
- Multiple potential duplicates
- High ratio of ideas to decisions/specs

**Orderly** if:
- Questions mostly answered
- Concerns addressed
- Ideas synthesized into decisions/specs

### Step 5: Present Survey

Output format:
```
## Survey: [CONTAINER-ID] ([Container Title])

### Summary
- X tomes (if conclave), Y notes total
- Z open questions, W unaddressed concerns

### Notes by Type
| Type     | Open | Closed |
|----------|------|--------|
| idea     | X    | Y      |
| question | X    | Y      |
| concern  | X    | Y      |
| spec     | X    | Y      |
| decision | X    | Y      |

### State: [Chaotic/Orderly]
Signals: [list specific signals that led to assessment]

### Recommendation
**[Clean/Ship]** - [rationale based on state]

Select objective: [c]lean / [s]hip / [b]oth sequentially
```

### Step 6: Await Selection

Wait for user to choose c, s, or b.
- If clean: proceed to theme selection for clean patterns
- If ship: proceed to theme selection for ship patterns
- If both: clean first, then ship

## Theme Selection & Interview

After objective is selected:

### Theme Identification

From the survey data, identify 3-5 themes such as:
- "Scattered ideas about [topic]" - multiple related ideas not synthesized
- "Unanswered questions about [topic]" - open questions on same subject
- "Competing approaches to [topic]" - conflicting decisions/specs
- "Stale discussions on [topic]" - old notes with no resolution

Present themes:
```
Themes identified:
1. Scattered ideas about skill design (5 notes)
2. Unanswered questions about execution model (3 notes)
3. Competing approaches to ledger structure (2 specs)

Select a theme to explore (1-3), or [a]ll:
```

### Interview Flow

For each selected theme, conduct structured interview:

**Interview structure:**
- Max 5 questions per theme
- Each question surfaces a decision point
- Choices map to maintenance patterns

**Progress indicator:**
```
Theme: Scattered ideas about skill design
Question 2/5: [question text]
```

**Question format:**
```
Context: [brief context from notes]

Question: [clear decision question]

Choices:
[a] [Choice that maps to SYNTHESIZE]
[b] [Choice that maps to CLOSE-SUPERSEDED]
[c] [Choice that maps to DEFER-TO-LIBRARY]
[s] Skip this question
```

**Question types by objective:**

Clean objective questions:
- "These notes seem redundant. Merge them?" → CONSOLIDATE-DUPLICATES
- "This is now covered in [note]. Close as superseded?" → CLOSE-SUPERSEDED
- "This is valid but not actionable now. Park to library?" → DEFER-TO-LIBRARY
- "This mixes vision and code. Split by layer?" → EXTRACT-LAYER

Ship objective questions:
- "These ideas converge on [concept]. Synthesize into spec?" → SYNTHESIZE
- "This implicit decision should be explicit. Promote?" → PROMOTE-TO-DECISION
- "Ready to create draft shipment for [scope]?" → (creates shipment)

### Action Proposal

After interview, summarize proposed actions:
```
## Proposed Actions

Based on your answers:

1. MERGE NOTE-101 into NOTE-105 (synthesize)
2. CLOSE NOTE-102 --reason superseded --by NOTE-105
3. CLOSE NOTE-103 --reason deferred
4. CREATE shipment "Skill Implementation" with spec from NOTE-105

Execute? [y]es / [n]o / [r]eview details
```

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
