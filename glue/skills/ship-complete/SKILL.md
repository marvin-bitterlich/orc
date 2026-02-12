---
name: ship-complete
description: Close a shipment. Use when user says /ship-complete or wants to mark a shipment as finished.
---

# Ship Complete Skill

Close a shipment (terminal state).

## Usage

```
/ship-complete              (close focused shipment)
/ship-complete SHIP-xxx     (close specific shipment)
```

## Status Lifecycle

This is the final step in the shipment lifecycle:

```
draft → ready → in-progress → closed
```

Shipments should normally be in "in-progress" status before closing.

## Flow

### Step 1: Get Shipment

If argument provided:
- Use specified SHIP-xxx

If no argument:
- Get focused shipment from `orc focus --show`
- If no focus, ask which shipment to close

### Step 2: Verify Readiness

```bash
orc shipment show <SHIP-xxx>
```

Check:
- Shipment status is "in-progress" or "ready"
- Shipment is not pinned

### Step 3: Handle Issues

If status is "draft":
```
Cannot close SHIP-xxx: shipment is still in 'draft' status

Move it to ready first:
  orc shipment status SHIP-xxx --set ready
```

If shipment is pinned:
```
Cannot close SHIP-xxx: shipment is pinned

Unpin first:
  orc shipment unpin SHIP-xxx
```

### Step 4: Close Shipment

```bash
orc shipment complete <SHIP-xxx>
```

### Step 5: Clear Focus

```bash
orc focus --clear
```

### Step 6: Summary

Output:
```
Shipment closed:
  SHIP-xxx: <Title>
  Status: closed

Next steps:
  orc summary              - View remaining work
  /ship-new "Title"        - Start new shipment
```

## Example Session

```
User: /ship-complete

Agent: [gets focused shipment SHIP-250]
       [runs orc shipment show SHIP-250]

Agent: Closing SHIP-250: Core Model Simplification

       Status: in-progress
       Ready for closure.

       [runs orc shipment complete SHIP-250]
       [runs orc focus --clear]

Agent: Shipment closed:
         SHIP-250: Core Model Simplification
         Status: closed

       Next steps:
         orc summary - View remaining work
```
