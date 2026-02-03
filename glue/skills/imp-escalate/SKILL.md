---
name: imp-escalate
description: Escalate to gatehouse when genuinely blocked.
---

# IMP Escalate

Escalate to the gatehouse when genuinely blocked and need human help.

## Usage

`/imp-escalate --reason "reason for escalation"`

## Flow

1. **Get current plan**
   ```bash
   orc plan list --task TASK-xxx
   ```

2. **Escalate plan**
   ```bash
   orc plan escalate PLAN-xxx --reason "reason for escalation"
   ```

3. **Output**
   "Escalated as ESC-xxx. Waiting for human resolution. Explain the problem to El Presidente when they arrive."

## When to Escalate

- Task requirements are ambiguous
- Architectural decision beyond IMP authority
- Blocked by external dependency
- Significant risk identified

Escalation is for genuine blockers - use sparingly.

## Notes

- Escalation fires a "flare" to the gatehouse
- The IMP should pause and wait for resolution
- This is a normal part of the workflow, not a failure
- Resolve with: `orc escalation resolve ESC-xxx --outcome approved`
