---
name: ship-complete
description: Complete a shipment after all tasks are done. Use when user says /ship-complete or wants to mark a shipment as finished.
---

# Ship Complete Skill

Mark a shipment as complete after verifying all work is done.

## Usage

```
/ship-complete              (complete focused shipment)
/ship-complete SHIP-xxx     (complete specific shipment)
/ship-complete --force      (complete even with incomplete tasks)
```

## Flow

### Step 1: Get Shipment

If argument provided:
- Use specified SHIP-xxx

If no argument:
- Get focused shipment from `orc focus --show`
- If no focus, ask which shipment to complete

### Step 2: Verify Readiness

```bash
orc shipment show <SHIP-xxx>
orc task list --shipment <SHIP-xxx>
```

Check:
- All tasks are complete
- Shipment is not pinned
- No blocking issues

### Step 3: Handle Issues

If tasks incomplete and no --force:
```
Cannot complete SHIP-xxx: 2 tasks incomplete
  - TASK-yyy: In Progress
  - TASK-zzz: Ready

Options:
  1. Complete remaining tasks first
  2. Use --force to complete anyway (tasks will be abandoned)
```

If shipment is pinned:
```
Cannot complete SHIP-xxx: shipment is pinned

Unpin first:
  orc shipment unpin SHIP-xxx
```

### Step 4: Complete Shipment

```bash
orc shipment complete <SHIP-xxx>
```

Or with force:
```bash
orc shipment complete <SHIP-xxx> --force
```

### Step 5: Clear Focus

```bash
orc focus --clear
```

### Step 6: Summary

Output:
```
Shipment completed:
  SHIP-xxx: <Title>
  Tasks: X/X done
  Duration: Y days

Next steps:
  orc summary              - View remaining work
  /ship-queue claim        - Claim next from queue
  /ship-new "Title"        - Start new shipment
```

## Example Session

```
User: /ship-complete

Agent: [gets focused shipment SHIP-250]
       [runs orc shipment show SHIP-250]
       [runs orc task list --shipment SHIP-250]

Agent: Completing SHIP-250: Core Model Simplification

       Tasks: 7/7 complete
       All checks passed.

       [runs orc shipment complete SHIP-250]
       [runs orc focus --clear]

Agent: Shipment completed:
         SHIP-250: Core Model Simplification
         Tasks: 7/7 done

       Next steps:
         orc summary - View remaining work
         /ship-queue claim - Claim next from queue
```

## Auto-Complete

Note: Shipments can also auto-complete when the last task is marked complete,
if all tasks are done. This skill is for manual completion or verification.
