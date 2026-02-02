---
name: ship-queue
description: View and manage the shipyard queue. Use when user says /ship-queue or wants to see pending shipments and priorities.
---

# Ship Queue Skill

View and manage shipments in the shipyard queue.

## Usage

```
/ship-queue                     (view queue)
/ship-queue claim               (claim next shipment)
/ship-queue priority SHIP-xxx N (set priority)
```

## Flow

### Step 1: Determine Action

If "claim" argument:
- Go to Claim flow

If "priority SHIP-xxx N" arguments:
- Go to Priority flow

Otherwise:
- Go to View flow

---

## View Flow

### Step 2: Get Queue

```bash
orc shipyard queue
```

Display queue with:
- Position
- Shipment ID and title
- Task counts (done/total)
- Priority if set
- Assignment status

### Step 3: Show Options

Output:
```
Shipyard Queue:
  1. SHIP-xxx: Title (0/5 tasks) [P1]
  2. SHIP-yyy: Title (2/4 tasks)
  3. SHIP-zzz: Title (0/3 tasks)

Actions:
  /ship-queue claim              - Claim next available
  /ship-queue priority SHIP-xxx 1 - Set priority
  orc shipment show SHIP-xxx     - View details
```

---

## Claim Flow

### Step 2: Check Current Assignment

```bash
orc status
```

If already working on a shipment:
- Warn user about current assignment
- Ask to confirm switching

### Step 3: Claim from Queue

```bash
orc shipyard claim
```

This:
- Assigns top shipment to current workbench
- Focuses the shipment
- Transitions status if needed

### Step 4: Confirm

Output:
```
Claimed from queue:
  SHIP-xxx: <Title>
  Tasks: X total (Y ready)

Start work:
  orc task list --shipment SHIP-xxx
  orc task claim TASK-yyy
```

---

## Priority Flow

### Step 2: Set Priority

```bash
orc shipment priority <SHIP-xxx> --priority <N>
```

Priority 1 = highest, NULL = default FIFO

### Step 3: Confirm

Output:
```
Priority set:
  SHIP-xxx now at priority N

Updated queue position: #1
```

## Example Session

```
User: /ship-queue

Agent: [runs orc shipyard queue]

Agent: Shipyard Queue (3 shipments):

       #  ID        Title                    Tasks   Priority
       1  SHIP-248  Test Shipyard Badge      0/0     [P1]
       2  SHIP-247  Exorcism/Shipment Split  0/2     -
       3  SHIP-229  Bugfixes Batch           3/10    -

       Actions:
         /ship-queue claim - Claim #1
         /ship-queue priority SHIP-247 1 - Bump priority
```
