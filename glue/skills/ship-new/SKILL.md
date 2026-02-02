---
name: ship-new
description: Create a new shipment for implementation work. Use when user says /ship-new or wants to create a shipment from exploration or directly.
---

# Ship New Skill

Create a new shipment for implementation work.

## Usage

```
/ship-new "Shipment Title"
/ship-new --from CON-xxx    (create from conclave exploration)
/ship-new                   (will prompt for title)
```

## Flow

### Step 1: Get Title and Source

If `--from CON-xxx` provided:
- Get conclave details: `orc conclave show CON-xxx`
- Use conclave title as default shipment title
- Link shipment to conclave via `--conclave` flag

If title argument provided:
- Use it as shipment title

If neither:
- Ask: "What should this shipment implement?"

### Step 2: Detect Context

```bash
orc status
```

From output, identify:
- Current commission (from workbench context)
- If no commission detected, ask user which commission

### Step 3: Create Shipment

```bash
orc shipment create "<Title>" \
  --commission <COMM-xxx> \
  --description "Implementation shipment for <title>"
```

If from conclave, add `--conclave CON-xxx`.

Capture the created `SHIP-xxx` ID from output.

### Step 4: Focus Shipment

```bash
orc focus <SHIP-xxx>
```

This auto-transitions status: draft → exploring

### Step 5: Confirm Ready

Output:
```
Shipment created and focused:
  SHIP-xxx: <Title>
  Status: exploring

Next steps:
  - Add tasks: orc task create "Task title" --shipment SHIP-xxx
  - Plan tasks: /ship-plan
  - Or discuss requirements - I'll help capture tasks
```

## Example Session

```
User: /ship-new "User Authentication"

Agent: [runs orc status, detects COMM-001]
       [runs orc shipment create "User Authentication" --commission COMM-001]
       [runs orc focus SHIP-xxx]

Agent: Shipment created and focused:
         SHIP-xxx: User Authentication
         Status: exploring

       Next steps:
         - Add tasks: orc task create "Task title" --shipment SHIP-xxx
         - Plan tasks: /ship-plan
```
