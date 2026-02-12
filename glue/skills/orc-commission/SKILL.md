---
name: orc-commission
description: Guide through creating a new commission. Use when user says /orc-commission or wants to create a commission for tracking work.
---

# Commission Creation Skill

Guide users through creating a new commission in ORC.

## Usage

```
/orc-commission <title>
/orc-commission              (will prompt for title)
```

## Flow

### Step 1: Gather Commission Title

If title not provided, ask:
- "What is the commission title?" (brief description of the work)

### Step 2: Gather Description (Optional)

Ask for an optional description:
- "What is the commission about?" (more detail, or press Enter to skip)

### Step 3: Detect Context

```bash
orc status
```

Identify the current factory context. If no factory detected, commissions are created in the default factory.

### Step 4: Create Commission

```bash
orc commission create "<title>" --description "<description>"
```

Capture the created COMM-xxx ID.

### Step 5: Verify Creation

```bash
orc commission show COMM-xxx
```

### Step 6: Confirm Ready

Output:
```
Commission created:
  COMM-xxx: <title>
  Factory: <factory-name>

Next steps:
  orc focus COMM-xxx               # Focus on this commission
  orc shipment create "<title>"    # Create a shipment for work
  orc workshop create "<name>"     # Create a workshop if needed
```

## Example Session

```
User: /orc-commission Authentication Refactor

Agent: What is the commission about?
       (Optional description, or press Enter to skip)

User: Modernize the authentication system to use OAuth2

Agent: [runs orc commission create "Authentication Refactor" --description "Modernize the authentication system to use OAuth2"]

Agent: Commission created:
         COMM-xxx: Authentication Refactor
         Factory: default

       Next steps:
         orc focus COMM-xxx               # Focus on this commission
         orc shipment create "<title>"    # Create a shipment for work
```

## IMP Usage

IMPs can use this skill to create commissions when:
- Exploring work that doesn't fit an existing commission
- Starting a new area of investigation
- Organizing related shipments under a common theme

After creation, the IMP should:
1. Focus on the new commission
2. Create a shipment for the specific work
3. Continue with normal IMP workflow

## Error Handling

- If `orc commission create` fails, show the error message
- If no factory context, commission is created in default factory
