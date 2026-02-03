---
name: imp-nudge
description: Manual re-propulsion. Check state and output appropriate next action.
---

# IMP Nudge

Manual re-propulsion. Check current state and output the appropriate next action.

## Flow

1. **Check current state**
   ```bash
   orc status
   orc task list --status implement
   orc plan list --task TASK-xxx
   orc escalation list --status pending
   ```

2. **Determine state and output next action**

   | State | Next Action |
   |-------|-------------|
   | In-progress task, no plan | "Run /imp-plan-create" |
   | In-progress task, plan exists | "Implement it, then run /imp-rec" |
   | Pending escalation | "Waiting on ESC-xxx. Resolve with El Presidente." |
   | Ready tasks available | Claim next: `orc task claim`, then /imp-plan-create |
   | No tasks (shipment complete) | "Shipment complete!" |
   | No tasks (blocked) | "Waiting on blocked tasks." |
   | No focused shipment | "Run `orc focus SHIP-xxx`" |

3. **Output**
   State the current state and the recommended next action clearly.

## Usage

`/imp-nudge` - Check state and get next action

## Notes

- Use when unsure what to do next
- Use to restart after interruption
- Use to verify you're on track
- The Stop hook uses similar logic to prevent premature stopping
