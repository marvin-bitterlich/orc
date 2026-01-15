# Handoff Command

Capture current IMP session context in ORC ledger before ending session.

## Role

You are a **Session Context Archiver** for IMPs that captures agent discoveries, decisions, and work state into ORC's ledger, enabling seamless session continuity for grove-based work.

## Usage

```
/handoff
```

**Purpose**: Preserve IMP session context in ORC ledger:
- **What was accomplished** this session
- **Current state** of epic work
- **Decisions made** and rationale
- **Open questions** and blockers
- **Recommended next steps** for resume

## Core Rules

**IMPORTANT RESTRICTIONS:**
- **TodoWrite tool is NOT ALLOWED** - This tool is banned in ORC workflow
- **TODO markdown files are NOT ALLOWED** - No TODO.md or similar files
- **Use ORC ledger only** - All task tracking must be in `orc task` commands

## Process

<step number="1" name="gather_session_state">
**Gather Session Context:**

Detect current grove context:
- Run `orc prime` to understand IMP identity and assignment
- Identify which grove (GROVE-ID) we're operating in
- Identify which epic(s) are assigned to this grove

Analyze conversation history for:
- **Key decisions**: "We chose X because Y"
- **Technical discoveries**: "Found that Z uses pattern A"
- **Architectural insights**: "System works by B"
- **Open questions**: "Need to investigate C"
- **Blockers**: "Waiting on D"
- **Tasks completed**: Check `orc task list` for recently completed tasks

Identify investigated artifacts:
- File paths read or edited
- Modules/components explored
- APIs or services touched

Determine next steps:
- What should be resumed first?
- What's ready to work on next?
- What needs follow-up?
</step>

<step number="2" name="create_ledger_handoff">
**Create Ledger Handoff:**

**Write Narrative Note for Next Claude:**

Craft a Claude-to-Claude handoff note in markdown format:
- Write in second person ("You were working on...")
- Focus on narrative flow, not structured data
- Include what was accomplished, current state, what's next
- Add important context and gotchas
- Keep it warm but professional

**Create Ledger Handoff:**

```bash
orc handoff create \
  --note "$(cat <<'EOF'
Hey next Claude! Here's where we are in this grove:

## What We Accomplished
[List key completions from this session]

## Current State
[Describe active work on assigned epic(s), decisions made, discoveries]

## What's Next
[Clear next steps with priority - focus on ready tasks in epic]

## Important Context
[Gotchas, blockers, open questions]
EOF
)" \
  --mission [MISSION-ID from grove config] \
  --grove [GROVE-ID from grove config]
```

**Result:** Ledger handoff created instantly, next IMP boot will include this context.

**Benefits:**
- Next Claude (IMP) gets context in <1 second via `orc prime`
- Structured relationships via database
- Grove-scoped handoffs for focused continuity
</step>

<step number="3" name="prompt_clear">
**Prompt User to Clear Session:**

After handoff is created, tell El Presidente:

```
✓ Ledger handoff created: HO-XXX
  Created: [timestamp]
  Mission: [MISSION-ID]
  Grove: [GROVE-ID]

✓ Context preserved for grove [GROVE-NAME].

When you reconnect to this grove, the IMP will automatically receive this handoff
via `orc prime` during boot.
```

**What Happens When IMP Reconnects:**
1. TMux pane runs `orc connect`
2. Claude launches and runs `orc prime`
3. Prime detects grove context
4. Prime includes latest handoff for this grove
5. New IMP session starts with full context
</step>

## Implementation Logic

**Decision Extraction:**
- Look for phrases: "decided to", "chose", "going with", "selected"
- Capture context: what was decided and why
- Avoid implementation details, focus on strategic choices

**Discovery Extraction:**
- Look for phrases: "discovered", "found that", "realized", "identified"
- Capture insights about system behavior, architecture, patterns
- NOT raw code - conceptual understanding only

**Task Progress:**
- Query `orc task list --status completed` to see what was finished
- Include recently completed task IDs in handoff
- Note any tasks moved to implement status

## Expected Behavior

When El Presidente runs `/handoff`:

1. **"🔍 Gathering session state..."** - Detect grove, analyze conversation
2. **"💾 Creating ledger handoff..."** - Create handoff in ORC ledger with grove link
3. **"✅ Handoff complete: HO-XXX"** - Display handoff ID and timestamp

**Example Output:**
```
🔍 Gathering session state...
   - Grove: migration (GROVE-002)
   - Epic: EPIC-117 (Worker Migrations & Teardowns)
   - 2 tasks completed this session
   - 2 key decisions captured
   - 1 technical discovery recorded
   - 1 open question noted

✓ Ledger handoff created: HO-019
  Created: 2026-01-15 10:45
  Mission: MISSION-002
  Grove: GROVE-002

✓ Context preserved for grove migration.

When you reconnect to this grove, the IMP will automatically receive this handoff
via `orc prime` during boot.
```

## Integration Notes

**Works With:**
- `orc prime`: Injects handoff context during IMP boot
- `orc connect`: Launches IMP with prime directive
- `orc task`: Track task completion in ledger (NO TodoWrite)

**Storage:**
- All data in ORC SQLite ledger
- Persistent across sessions
- Grove-scoped for focused context

**Session Continuity Flow:**
1. End of IMP session: Run `/handoff` to capture context
2. TMux pane can be killed or respawned
3. TMux respawn runs `orc connect`
4. Claude boots and runs `orc prime`
5. Prime includes handoff for this grove
6. New IMP starts with orientation context
