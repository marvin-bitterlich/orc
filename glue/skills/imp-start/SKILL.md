---
name: imp-start
description: Begin autonomous work on a shipment. Run preflight checks, checkout branch, claim first task.
---

# IMP Start

Begin autonomous work on the focused shipment.

## Usage

```
/imp-start        (manual mode - can stop freely)
/imp-start --auto (auto mode - Stop hook blocks until complete)
```

## Flow

1. **Get focused shipment**
   ```bash
   orc status
   ```
   Identify the focused shipment (SHIP-xxx).

2. **Preflight checks**
   - Git clean? (`git status --short`)
   - Is there an in_progress task already? (skip to /imp-plan-create if so)
   - Are there ready tasks in the shipment?

3. **Handle dirty git state**
   If git is dirty, offer to stash:
   ```bash
   git stash push -m "imp-start auto-stash"
   ```

4. **Checkout shipment branch**
   Get the shipment branch and checkout:
   ```bash
   orc shipment show SHIP-xxx
   git checkout <branch-name>
   ```

5. **Claim first ready task**
   ```bash
   orc task claim --shipment SHIP-xxx
   ```

6. **Enable auto mode (if --auto flag)**
   ```bash
   orc shipment auto SHIP-xxx
   ```

7. **Output**
   - Without --auto: "Task TASK-xxx claimed. Run /imp-plan-create to create an implementation plan."
   - With --auto: "Task TASK-xxx claimed. Auto mode enabled. Run /imp-plan-create to create an implementation plan."

## Error Handling

- No focused shipment → "No shipment focused. Run `orc focus SHIP-xxx` first."
- No ready tasks → "No ready tasks in shipment. Check `orc task list --shipment SHIP-xxx`."
- Already in_progress task → "Task TASK-xxx already in progress. Run /imp-plan-create."

## Notes

- Use --auto for autonomous execution (Stop hook blocks until shipment complete)
- Use /imp-auto to toggle mode mid-flight
