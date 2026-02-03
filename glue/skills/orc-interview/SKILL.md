---
name: orc-interview
description: Reusable interview primitive for surfacing decisions. Use standalone (/orc-interview) for ad-hoc decision-making, or nested from other skills (ship-synthesize, ship-plan) for structured interviews.
---

# ORC Interview

A lightweight interview mechanism for surfacing and resolving decisions. Works standalone or as a primitive called by other skills.

## Usage

```
/orc-interview                    (start ad-hoc interview)
/orc-interview "Topic to decide"  (interview about specific topic)
```

When called from another skill (ship-synthesize, ship-plan, etc.), the interview runs inline without taking over the conversation.

## Interview Format

Each question follows this natural language structure:

```
[Question X/Y]

<Context and stakes - what's being decided and why it matters>

<Trade-offs between options>

<Recommendation with rationale>

1. Approve - accept the recommendation
2. Plan B - choose the alternative
3. Skip - defer this decision
4. Discuss - explore further before deciding
```

### Format Guidelines

- **Natural language**: Write conversationally, not as rigid forms
- **Context first**: Explain the situation and what's at stake
- **Trade-offs visible**: Show what each choice enables or costs
- **Clear recommendation**: State what you'd suggest and why
- **Numbered options**: Always 1-4 (approve, alternative, skip, discuss)
- **Progress indicator**: Always show [Question X/Y]

### Example Question

```
[Question 2/5]

These two notes propose different approaches to focus scope - NOTE-506 says
any actor can focus anything, while NOTE-507 limits it to commission-level
entities. The difference matters because unrestricted focus adds flexibility
but could clutter the summary view with noise.

I'd recommend NOTE-507's approach (commission-level only) - it keeps focus
meaningful as a navigation concept rather than just a bookmark.

1. Approve - go with commission-level focus
2. Universal focus instead
3. Skip
4. Discuss
```

## Standalone Mode

When invoked directly via `/orc-interview`:

### Step 1: Identify Topic

If argument provided:
- Use it as the interview topic

If no argument:
- Ask: "What decision or topic should we work through?"

### Step 2: Gather Context

Ask clarifying questions to understand:
- What's being decided
- What options exist
- What constraints apply
- Who's affected

### Step 3: Run Interview

Conduct structured interview:
- Max 5 questions (default)
- Each question surfaces one decision point
- Record answers as decisions are made

### Step 4: Summarize Decisions

Output:
```
Interview complete. Decisions made:

1. [Topic]: [Decision]
2. [Topic]: [Decision]
...

These decisions can inform next steps or be recorded as notes.
```

## Nested Mode

When called from another skill (ship-synthesize, ship-plan, orc-architecture):

### Behavior

- **Lightweight**: No greeting, no "interview complete" ceremony
- **Inline**: Questions appear in the parent skill's flow
- **Returns decisions**: Parent skill receives answers and continues
- **Respects parent's max**: Parent can specify max questions

### Integration Pattern

Parent skill prepares:
- List of themes/topics to interview about
- Max questions (optional, default 5)
- Context already gathered

Interview runs:
- Questions presented one at a time
- User responds with 1, 2, 3, or 4
- Decisions accumulated

Parent skill receives:
- List of decisions made
- Any topics skipped
- Any topics needing more discussion

### Example Integration

```
# In ship-synthesize, after identifying themes:

Running interview for 3 themes...

[Question 1/5]

Notes 503-507 have scattered ideas about focus scope. NOTE-507 proposes
limiting to commission-level entities, while NOTE-506 suggests universal
focus. This affects how flexible vs. meaningful the focus command feels.

I'd recommend commission-level only - keeps focus as a navigation waypoint
rather than a general bookmark.

1. Approve commission-level focus
2. Go with universal focus
3. Skip - decide later
4. Discuss further

> 1

[Question 2/5]
...
```

## Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| max_questions | 5 | Maximum questions in one interview session |
| topic | (none) | Optional topic to focus the interview |

## Response Handling

| Response | Meaning | Action |
|----------|---------|--------|
| 1 | Approve | Accept recommendation, record decision |
| 2 | Plan B | Accept alternative, record decision |
| 3 | Skip | Defer this decision, move to next question |
| 4 | Discuss | Pause for clarification, then re-present or revise |

When user chooses "4. Discuss":
- Invite them to share concerns or questions
- Incorporate their input
- Re-present the question with adjusted framing, or
- Move on if they indicate readiness

## Best Practices

### Writing Good Questions

1. **Lead with context**: Don't assume the reader remembers everything
2. **Make stakes clear**: Why does this decision matter?
3. **Show trade-offs honestly**: Don't hide downsides of your recommendation
4. **Be opinionated**: A recommendation is more helpful than pure neutrality
5. **Keep it concise**: One screen, readable at a glance

### Respecting User Time

- Don't pad with unnecessary questions
- If something is obvious, don't ask - just note it
- 5 questions max keeps interviews focused
- Skip ceremony when nested

### Handling Uncertainty

If you're unsure what to recommend:
- Say so: "I'm not confident here, but leaning toward X because..."
- Offer "4. Discuss" as a genuine option
- Be ready to learn from the user's perspective
