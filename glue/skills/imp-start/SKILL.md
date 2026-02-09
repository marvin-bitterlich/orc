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

6. **Enable auto mode and spawn watchdog (if --auto flag)**
   ```bash
   # Enable auto mode for shipment
   orc shipment auto SHIP-xxx

   # Start patrol (creates monitoring session for watchdog)
   orc patrol start BENCH-xxx

   # Apply infrastructure to spawn watchdog pane
   orc infra apply WORK-xxx --yes
   ```

   Wait for watchdog pane to appear before continuing.

7. **Output**
   - Without --auto: "Task TASK-xxx claimed. Run /imp-plan-create to create an implementation plan."
   - With --auto: "Task TASK-xxx claimed. Auto mode enabled, watchdog spawned. Run /imp-plan-create to create an implementation plan."

## Error Handling

- No focused shipment → "No shipment focused. Run `orc focus SHIP-xxx` first."
- No ready tasks → "No ready tasks in shipment. Check `orc task list --shipment SHIP-xxx`."
- Already in_progress task → "Task TASK-xxx already in progress. Run /imp-plan-create."
- Patrol start fails → Report error, continue without watchdog
- Infra apply fails → Report error, end patrol, continue without watchdog

## Notes

- Use --auto for autonomous execution with watchdog monitoring
- Watchdog monitors IMP progress and nudges when idle
- Use /imp-auto to toggle mode mid-flight (spawns/removes watchdog)
