---
name: imp-poll
description: Check shipyard queue and claim work. IMP uses this to find available shipments when idle.
---

# IMP Poll

Check the shipyard queue for available work and optionally claim a shipment.

## When to Use

- IMP has no current shipment assigned
- IMP completed a shipment and needs new work
- IMP wants to see what's queued

## Flow

### Step 1: Check Current State

```bash
orc status
```

Verify:
- Is there already a shipment assigned to this workbench? If yes, suggest `/imp-start` instead.
- Get the commission context for filtering the queue.

### Step 2: Display Queue

```bash
orc shipyard list
```

Show the queue with this format:
```
Checking shipyard queue...

 #  SHIP        TITLE                           PRI   TASKS
 1  SHIP-240    Critical hotfix                 1     3/3
 2  SHIP-237    Plan/Apply Refactor             -     0/11

Options:
[c] Claim #1 (top of queue)
[n] Claim specific shipment
[r] Refresh queue
[q] Quit
```

### Step 3: Handle Selection

**[c] Claim top shipment:**
```bash
orc shipyard claim
```

Output: "Claimed SHIP-xxx. Run `/imp-start` to begin work."

**[n] Claim specific:**
Prompt: "Enter shipment number (1-N) or ID (SHIP-xxx):"

If number: Map to shipment ID from displayed list
If ID: Use directly

```bash
orc shipment assign SHIP-xxx BENCH-yyy
```

Where BENCH-yyy is the current workbench from context.

**[r] Refresh:**
Re-run `orc shipyard list` and display updated queue.

**[q] Quit:**
Exit without claiming.

## After Claiming

After successful claim:
```
✓ Claimed SHIP-xxx: [title]
  Assigned to: BENCH-yyy
  Tasks: X ready

Run /imp-start to begin work on the first task.
```

## Error Handling

- Empty queue → "Shipyard queue is empty. No work available."
- Already has shipment → "Workbench already has SHIP-xxx assigned. Run /imp-start or /imp-nudge."
- Claim fails → "Failed to claim shipment: [error]"

## Example Session

```
> /imp-poll

Checking shipyard queue...

YARD-001 (COMM-001) - 2 shipments queued

 #  SHIP      TITLE                    PRI  TASKS   QUEUED
--  ----      -----                    ---  -----   ------
 1  SHIP-240  Critical hotfix          1    0/3     2h ago
 2  SHIP-237  Plan/Apply Refactor      -    0/11    1d ago

[c] Claim #1 (SHIP-240)
[n] Claim specific
[r] Refresh
[q] Quit

> c

✓ Claimed SHIP-240: Critical hotfix
  Assigned to: BENCH-014
  Tasks: 3 ready

Run /imp-start to begin work on the first task.
```
