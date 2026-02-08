---
name: orc-ideate
description: Enter ideation mode for collaborative exploration of ideas. Use when user says /ideate or wants to brainstorm on a focused shipment.
---

# ORC Ideate

Collaborative ideation session for exploring ideas with the human.

## Usage

    /ideate

## Prerequisites

Must have a shipment focused:

    orc focus --show

If no shipment focused, error: "No shipment focused. Create one with /ship-new or focus an existing one with orc focus SHIP-xxx."

## On Invoke: Guidance Banner

Output this banner to establish shared understanding:

    ## IDEATION MODE

    Human: share ideas freely.
    Agent: ask questions, explore implications, capture notes as we go.

    Note types: idea, question, finding, decision, concern, spec
    When ready: /ship-synthesize to tidy messy notes, or /ship-plan to go straight to tasks.

Then wait for the human to share their first idea.

## Agent Behavior: Socratic-lite

During ideation, follow these principles:

1. **Follow the human's lead** - They drive the direction, you support
2. **Ask clarifying questions** - "What problem does this solve?" "Who would use this?"
3. **Explore implications** - "If we did X, then Y would need to change..."
4. **Don't drive or push** - Offer perspectives, don't steer
5. **Recognize deep-dive moments** - When themes emerge that need structured exploration, suggest: "This feels like it needs deeper exploration. Want to run /orc-interview on this?"

## Continuous Capture

Throughout the session, capture notes as ideas crystallize:

    orc note create "<title>" --shipment <SHIP-xxx> --type <type>

**Choose the appropriate type based on content:**

| Content | Type |
|---------|------|
| A possibility to explore | idea |
| Something uncertain | question |
| An observation or fact discovered | finding |
| A conclusion reached | decision |
| A risk or worry | concern |
| A requirement or constraint | spec |

**When to capture:**
- When an idea is articulated clearly enough to stand alone
- When a question is raised that should be tracked
- When a decision is made (even implicitly)
- When a concern is voiced

**Don't over-capture** - not every sentence needs a note. Capture the meaningful moments.

## Natural Ending

There is no "done" command or ceremony. The session ends naturally when:
- Human says they're done or moves on to another topic
- Human invokes another skill
- Human starts a different kind of work

**On ending**, you may mention next steps but don't push:
- "When you're ready to organize these ideas: /ship-synthesize"
- "If you want to go straight to tasks: /ship-plan"

## Example Session

    > /ideate

    ## IDEATION MODE

    Human: share ideas freely.
    Agent: ask questions, explore implications, capture notes as we go.

    Note types: idea, question, finding, decision, concern, spec
    When ready: /ship-synthesize to tidy messy notes, or /ship-plan to go straight to tasks.

    > I'm thinking we need a way for IMPs to reset their context

    Interesting. What's driving that need - is it about memory limits,
    or more about getting a fresh perspective on a problem?

    > More about clean slate after finishing work. Context from old work pollutes

    That makes sense. So after completing a task or shipment, the accumulated
    context becomes noise rather than signal.

    [captures: orc note create "IMPs need clean context reset after completing work" --shipment SHIP-xxx --type idea]

    Idea captured: NOTE-xxx

    How would the IMP trigger this reset - command, automatic, or skill?

    > skill probably. /imp-respawn or something

    [captures: orc note create "Skill-based respawn: /imp-respawn" --shipment SHIP-xxx --type idea]

    Idea captured: NOTE-xxx

    What should happen on respawn - just restart Claude, or also reset
    focus and other state?

    [conversation continues naturally...]
