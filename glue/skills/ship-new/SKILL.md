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
```

### Step 6: Enter Ideation Mode

After confirming the shipment is ready, automatically chain to orc-ideate:

```
Now entering ideation mode...
```

Then invoke the orc-ideate skill (output its guidance banner and begin ideation session).

This makes ideation the default entry point for new shipments.

## Example Session

```
User: /ship-new "User Authentication"

Agent: [runs orc status, detects COMM-001]
       [runs orc shipment create "User Authentication" --commission COMM-001]
       [runs orc focus SHIP-xxx]

Agent: Shipment created and focused:
         SHIP-xxx: User Authentication
         Status: exploring

       Now entering ideation mode...

       ## IDEATION MODE

       Human: share ideas freely.
       Agent: ask questions, explore implications, capture notes as we go.

       Note types: idea, question, finding, decision, concern, spec
       When ready: /ship-synthesize to tidy messy notes, or /ship-plan to go straight to tasks.

User: [begins sharing ideas about authentication...]
```
