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

## Clean Objective Execution

When clean objective is selected, execute these patterns via CLI:

### CONSOLIDATE-DUPLICATES

Merge redundant notes:
```bash
orc note merge <source-id> <target-id>
```

After merge:
- Source content is prepended to target
- Source is automatically closed with merge reference

### CLOSE-SUPERSEDED

Close notes that are now covered elsewhere:
```bash
orc note close <note-id> --reason superseded --by <better-note-id>
```

### DEFER-TO-LIBRARY

Park valid-but-not-now content:
```bash
# Create library note if needed
orc note create "Library: [topic]" --type learning --tome TOME-xxx

# Close deferred note with reference
orc note close <note-id> --reason deferred --by <library-note-id>
```

### EXTRACT-LAYER

Split notes mixing C4 levels:
1. Create new note for extracted layer
2. Update original to remove extracted content
3. Close original if fully extracted, or add cross-reference

```bash
# Create note for extracted layer
orc note create "[L1 Context]" --type vision --shipment SHIP-xxx

# Update original note
orc note update <original-id> --content "[remaining L3/L4 content]"
```

### BRIDGE-CONTEXT

Add references to orphan details:
```bash
# Update orphan note to reference context
orc note update <orphan-id> --content "Related to: NOTE-xxx\n\n[original content]"
```

### Execution Flow

For each proposed action from interview:
1. Show the action and affected notes
2. Execute via CLI command
3. Confirm success or report error
4. Move to next action

```
Executing action 1/4: MERGE NOTE-101 into NOTE-105
[running: orc note merge NOTE-101 NOTE-105]
✓ Merged NOTE-101 into NOTE-105

Executing action 2/4: CLOSE NOTE-102 --reason superseded --by NOTE-105
[running: orc note close NOTE-102 --reason superseded --by NOTE-105]
✓ Closed NOTE-102

All actions complete. Container tidied.
```

## Ship Objective Execution

When ship objective is selected, converge exploration into work:

### Step 1: Create Draft Shipment

Create a draft shipment as synthesis target:
```bash
orc shipment create "Shipment Title" --commission COMM-xxx --description "Synthesized from [container]"
```

### Step 2: SYNTHESIZE into Spec

Combine multiple findings/ideas into a unified spec:
```bash
# Create spec note attached to shipment
orc note create "Spec: [topic]" --type spec --shipment SHIP-xxx --content "[synthesized content]"

# Close source notes
orc note close <source-id> --reason synthesized --by <spec-note-id>
```

The agent:
1. Reads source notes
2. Synthesizes content (combines, deduplicates, structures)
3. Creates spec note with synthesized content
4. Closes source notes referencing the spec

### Step 3: PROMOTE-TO-DECISION

Extract implicit decisions into explicit decision notes:
```bash
# Create decision note
orc note create "Decision: [what was decided]" --type decision --shipment SHIP-xxx --content "[rationale and outcome]"

# Update source note to reference decision
orc note update <source-id> --content "[content] (See DECISION in NOTE-xxx)"
```

### Step 4: Guide Toward Roadmap

After spec exists, suggest roadmap creation:
```
Spec NOTE-xxx created for SHIP-xxx.

Next steps:
- Review spec for completeness
- Create roadmap: `orc note create "Roadmap: [topic]" --type roadmap --shipment SHIP-xxx`
- Add tasks: `orc task create "Task title" --shipment SHIP-xxx`

Would you like to:
[r] Create roadmap now
[t] Add tasks directly
[d] Done for now
```

### Ship Execution Flow

```
Creating draft shipment...
[running: orc shipment create "Skill Implementation" --commission COMM-001]
✓ Created SHIP-xxx

Synthesizing 5 notes into spec...
[reading: NOTE-101, NOTE-102, NOTE-103, NOTE-104, NOTE-105]
[creating spec with synthesized content]
✓ Created spec NOTE-xxx

Closing source notes...
✓ Closed NOTE-101 (synthesized)
✓ Closed NOTE-102 (synthesized)
...

Ship objective complete:
- Draft shipment: SHIP-xxx
- Spec: NOTE-xxx
- 5 notes synthesized
```

## Spot Check

Verification capability to confirm synthesis quality.

### Triggering Spot Check

**On request:**
```
spot check NOTE-302
```

**After batch close (agent offers):**
```
✓ Closed 5 notes (synthesized)

Would you like to spot check any? [y]es / [n]o
> y
Which note? NOTE-___
```

### Spot Check Process

1. Read the closed note content
2. Read the target note (from close_reason --by reference)
3. Identify where source content landed
4. Identify what was intentionally left behind

### Output Format

```
## Spot Check: NOTE-302

### Source Content
[original content from NOTE-302]

### Where It Landed
| Source excerpt | Target location |
|----------------|-----------------|
| "No auto modes" | NOTE-311 Non-Goals section |
| "Three actors" | NOTE-311 Roles table |
| "5 question limit" | NOTE-311 Interview Flow |

### Left Behind
- Internal tracking notes (supersedes lineage) - intentionally omitted
- Draft wording variations - captured in final form

### Verdict
✓ Content preserved in target
```

### Handling Discrepancies

If content appears missing:
```
⚠️ Possible gap detected

Source excerpt: "[excerpt]"
Not found in target NOTE-311

Options:
[a] Add to target now
[i] Ignore (intentionally omitted)
[r] Reopen source note
```

## Commands Reference

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
