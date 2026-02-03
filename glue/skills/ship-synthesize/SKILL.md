---
name: ship-synthesize
description: Knowledge compaction for shipments. Transforms messy exploration notes into a single summary note. Use when a shipment has accumulated notes that need consolidation before planning.
---

# Ship Synthesize

Transform chaotic exploration notes into a single compacted summary note. Read this one note and know everything relevant from the shipment's exploration phase.

## Usage

```
/ship-synthesize              (synthesize focused shipment)
/ship-synthesize SHIP-xxx     (synthesize specific shipment)
```

## When to Use

- Shipment has accumulated multiple notes during exploration
- Notes contain scattered ideas, open questions, or competing approaches
- Before running /ship-plan (synthesis provides clean input for planning)
- When you want to close out exploration and crystallize knowledge

## Flow

### Step 1: Identify Target

If argument provided:
- Use specified SHIP-xxx

If no argument:
```bash
orc focus --show
```
- Use focused shipment
- If no focus: "No shipment focused. Run `orc focus SHIP-xxx` first."

### Step 2: Survey Notes

Collect data:
```bash
orc shipment show SHIP-xxx
orc note list --shipment SHIP-xxx
```

From the collected data, compute:
- Total note count
- Notes grouped by type (idea, question, concern, decision, finding, spec, frq, etc.)
- Notes grouped by status (open vs closed)
- Open questions (type=question, status=open)
- Unaddressed concerns (type=concern, status=open)
- Potential duplicates (notes with very similar titles or content)
- Stale notes (old notes with no recent activity)

### Step 3: Present Survey

Output:
```
## Survey: SHIP-xxx (Shipment Title)

Notes: X total (Y open, Z closed)

| Type     | Open | Closed |
|----------|------|--------|
| idea     | X    | Y      |
| question | X    | Y      |
| concern  | X    | Y      |
| finding  | X    | Y      |
| decision | X    | Y      |
| spec     | X    | Y      |

Observations:
- [any notable patterns, duplicates, or issues]
```

### Step 4: Identify Themes

Group notes into 3-5 conceptual themes:
- "Scattered ideas about [topic]" - multiple related ideas not yet synthesized
- "Unanswered questions about [topic]" - open questions on same subject
- "Competing approaches to [topic]" - conflicting ideas or proposals
- "Unresolved concerns about [topic]" - open concerns needing attention

Present themes:
```
Themes identified:

1. Scattered ideas about focus scope (3 notes)
   NOTE-506, NOTE-507, NOTE-505

2. Unanswered questions about actor permissions (2 notes)
   NOTE-503, NOTE-504

3. Competing approaches to summary display (2 notes)
   NOTE-508, NOTE-509

Select theme to explore (1-3), or [a]ll:
```

### Step 5: Interview (per theme)

For each selected theme, run interview using orc-interview format:
- Max 5 questions per theme
- Surface decision points
- Apply resolution patterns based on answers

See "Interview Questions by Pattern" section below.

### Step 6: Apply Patterns

Based on interview answers, apply appropriate patterns:

| Pattern | When | Action |
|---------|------|--------|
| SYNTHESIZE | Multiple notes → one conclusion | Combine content into summary section |
| RESOLVE-QUESTION | Open question answered | Record answer in summary |
| RECORD-DECISION | Implicit decision made explicit | Document decision with rationale |
| CONSOLIDATE | Same concept, different words | Merge into single summary section |
| DEFER | Valid but not now | Note as "parked" in summary |
| DISCARD | No longer relevant | Note as "discarded" with reason |

### Step 7: Generate Summary Note

Create summary note with this structure:

```bash
orc note create "Summary: [Shipment Title]" --type spec --shipment SHIP-xxx --content "[content]"
```

Summary note template:
```markdown
# Summary: [Shipment Title]

Synthesized from [N] notes on [date].

## Context
[What problem/opportunity started this exploration]

## Key Decisions
1. **[Decision topic]**: [What was decided]
   - Rationale: [Why]
   - Source: NOTE-xxx, NOTE-yyy

2. **[Decision topic]**: [What was decided]
   - Rationale: [Why]

## Resolved Questions
- **[Question]**: [Answer]
- **[Question]**: [Answer]

## Current Understanding
[Synthesized knowledge - the "what we know now" section]

## Parked
- [Topic]: [Why parked, what would unpark it]

## Discarded
- [Topic]: [Why no longer relevant]

## Source Notes
Synthesized from: NOTE-xxx, NOTE-yyy, NOTE-zzz (now closed)
```

### Step 8: Spot Check (Optional)

Offer verification:
```
Summary NOTE-xxx created.

Spot check any source note to verify content was captured?
[y]es / [n]o
```

If yes:
1. User specifies which note to check
2. Read source note content
3. Show where each piece landed in summary
4. Show what was intentionally left behind
5. Offer to amend if gaps found

Spot check output:
```
## Spot Check: NOTE-302

### Source Content
[key points from NOTE-302]

### Where It Landed
| Source excerpt | Summary location |
|----------------|------------------|
| "Focus scope" | Key Decisions #1 |
| "Actor rules" | Current Understanding |

### Left Behind
- Draft wording variations - captured in final form
- Internal notes - intentionally omitted

### Verdict
✓ Content preserved in summary
```

### Step 9: Close Source Notes

After summary approved:
```bash
orc note close NOTE-xxx --reason synthesized --by NOTE-summary
orc note close NOTE-yyy --reason synthesized --by NOTE-summary
...
```

Output:
```
Closing source notes...
✓ NOTE-xxx closed (synthesized)
✓ NOTE-yyy closed (synthesized)
✓ NOTE-zzz closed (synthesized)

Synthesis complete:
  Summary: NOTE-xxx
  Source notes closed: N

Next: /ship-plan to create tasks from synthesized knowledge
```

## Interview Questions by Pattern

### SYNTHESIZE
```
[Question X/Y]

Notes [list] all discuss [topic] from different angles. They converge on
[observation]. Combining them into a single "Key Decision" or "Current
Understanding" section would capture the essence without the scattered pieces.

1. Approve - synthesize into summary
2. Keep separate - they're distinct enough
3. Skip
4. Discuss
```

### RESOLVE-QUESTION
```
[Question X/Y]

NOTE-xxx asks "[question]". Based on our discussion and other notes, the
answer appears to be [answer]. Recording this in "Resolved Questions"
closes the loop.

1. Approve - that's the answer
2. Different answer - [specify]
3. Skip - still open
4. Discuss
```

### RECORD-DECISION
```
[Question X/Y]

Looking at NOTE-xxx and NOTE-yyy, there's an implicit decision: [decision].
It's not written down explicitly but the notes assume it. Making it explicit
in "Key Decisions" prevents future confusion.

1. Approve - record the decision
2. That's not actually decided
3. Skip
4. Discuss
```

### DEFER
```
[Question X/Y]

NOTE-xxx raises [topic]. It's valid but not actionable right now - we'd need
[condition] before it becomes relevant. Parking it keeps it visible without
cluttering active work.

1. Approve - park it
2. Actually it's actionable now
3. Skip
4. Discuss
```

### DISCARD
```
[Question X/Y]

NOTE-xxx discussed [topic], but since then [change] happened. The note is
no longer relevant. Discarding it (with reason) keeps the summary clean.

1. Approve - discard it
2. Still relevant because [reason]
3. Skip
4. Discuss
```

## CLI Commands Reference

```bash
# Survey
orc shipment show SHIP-xxx
orc note list --shipment SHIP-xxx

# Create summary
orc note create "Summary: Title" --type spec --shipment SHIP-xxx --content "..."

# Close source notes
orc note close NOTE-xxx --reason synthesized --by NOTE-yyy

# Reason vocabulary: superseded, synthesized, resolved, deferred, duplicate, stale, discarded
```

## Example Session

```
> /ship-synthesize

[runs orc focus --show, finds SHIP-276]
[runs orc note list --shipment SHIP-276]

## Survey: SHIP-276 (Skill Cognitive Redesign)

Notes: 4 total (4 open, 0 closed)

| Type     | Open | Closed |
|----------|------|--------|
| finding  | 2    | 0      |
| decision | 2    | 0      |

Observations:
- Two findings led to two decisions
- No open questions or concerns
- Notes are coherent, not conflicting

Themes identified:

1. Analysis of lost capabilities (2 notes)
   NOTE-510, NOTE-511

2. Skill specifications (2 notes)
   NOTE-512, NOTE-513

Select theme to explore (1-2), or [a]ll:

> a

[Interview runs for each theme...]
[Summary note generated...]
[User approves...]
[Source notes closed...]

Synthesis complete:
  Summary: NOTE-514
  Source notes closed: 4

Next: /ship-plan to create tasks from synthesized knowledge
```
