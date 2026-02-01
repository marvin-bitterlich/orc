---
name: conclave
description: Quick ideation setup. Creates a conclave and exploration tome, focuses it, ready for ideation. Use when user says /conclave or wants to spin up a design discussion space.
---

# Conclave Skill

Quick setup for ideation and design exploration.

## Usage

```
/conclave "Topic Name"
/conclave              (will prompt for topic)
```

## Flow

### Step 1: Get Topic

If argument provided:
- Use it as the conclave title

If no argument:
- Ask: "What topic should this conclave explore?"

### Step 2: Detect Context

```bash
orc status
```

From output, identify:
- Current commission (from workbench or gatehouse context)
- If no commission detected, ask user which commission

### Step 3: Create Conclave

```bash
orc conclave create "<Topic>" \
  --commission <COMM-xxx> \
  --description "Design exploration for <topic>"
```

Capture the created `CON-xxx` ID from output.

### Step 4: Create Exploration Tome

```bash
orc tome create "<Topic> Exploration" \
  --conclave <CON-xxx> \
  --commission <COMM-xxx> \
  --description "Capturing ideas, questions, and decisions about <topic>"
```

Capture the created `TOME-xxx` ID from output.

### Step 5: Focus Conclave

```bash
orc focus <CON-xxx>
```

### Step 6: Confirm Ready

Output:
```
Conclave ready for ideation:
  CON-xxx: <Topic>
  TOME-xxx: <Topic> Exploration

Capture ideas:
  orc note create "Note title" --tome TOME-xxx --type idea

Or just discuss - I'll help capture notes as we go.

When ready to ship: /exorcism
```

## Role Support

Works for both Goblin and IMP:

**IMP context** (workbench):
- Detects commission from workbench assignment or asks

**Goblin context** (gatehouse):
- Uses workshop context or asks

## Example Session

```
User: /conclave "API Rate Limiting"

Agent: [runs orc status, detects COMM-001]
       [runs orc conclave create "API Rate Limiting" --commission COMM-001]
       [runs orc tome create "API Rate Limiting Exploration" --conclave CON-025]
       [runs orc focus CON-025]

Agent: Conclave ready for ideation:
         CON-025: API Rate Limiting
         TOME-052: API Rate Limiting Exploration

       Capture ideas:
         orc note create "Note title" --tome TOME-052 --type idea

       Or just discuss - I'll help capture notes as we go.

       When ready to ship: /exorcism
```

## Ideation Mode

After setup, the agent should:
- Help capture ideas as notes during discussion
- Use appropriate note types (idea, question, concern, decision)
- Periodically summarize captured notes
- Suggest `/exorcism` when discussion converges
