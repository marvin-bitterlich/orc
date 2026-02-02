---
name: ship-plan
description: Plan and create tasks for a shipment based on requirements. Use when user says /ship-plan or wants to break down shipment work into tasks.
---

# Ship Plan Skill

Plan tasks for a shipment by analyzing requirements and creating actionable tasks.

## Usage

```
/ship-plan              (plan focused shipment)
/ship-plan SHIP-xxx     (plan specific shipment)
```

## Flow

### Step 1: Get Shipment

If argument provided:
- Use specified SHIP-xxx

If no argument:
- Get focused shipment from `orc focus --show`
- If no focus, ask which shipment to plan

### Step 2: Gather Context

```bash
orc shipment show <SHIP-xxx>
orc task list --shipment <SHIP-xxx>
```

Also check:
- If shipment has a source conclave, review its notes
- If shipment has a spec note, review it

### Step 3: Analyze Requirements

Based on shipment description and any linked exploration:
- Identify distinct pieces of work
- Determine dependencies between tasks
- Estimate scope of each task
- Ensure tasks are self-contained (IMP can complete without questions)

### Step 4: Propose Tasks

Present proposed tasks:
```
Proposed tasks for SHIP-xxx:

1. [Task title]
   Description: ...

2. [Task title]
   Description: ...

Create these tasks? [y/n/edit]
```

### Step 5: Create Tasks

For each approved task:
```bash
orc task create "<Title>" \
  --shipment <SHIP-xxx> \
  --description "<Description>"
```

### Step 6: Update Shipment Status

If tasks were created and shipment was in exploring status:
```bash
# Status auto-transitions: exploring → tasked when first task created
```

### Step 7: Summary

Output:
```
Shipment planned:
  SHIP-xxx: <Title>
  Status: tasked
  Tasks created: X

Ready for work:
  orc task claim TASK-xxx
  /imp-start
```

## Guidelines

- Tasks should be completable in one session
- Each task should be self-contained
- Include verification steps in description
- Don't over-decompose (3-10 tasks typical)

## Example Session

```
User: /ship-plan

Agent: [gets focused shipment SHIP-300]
       [runs orc shipment show SHIP-300]
       [checks for linked notes/specs]

Agent: Planning SHIP-300: Add User Preferences

       Based on the shipment description, I propose:

       1. Add preferences table to schema
          - Create schema.sql entry
          - Run atlas apply

       2. Create preferences repository
          - Interface in ports/secondary
          - SQLite implementation

       3. Add preferences service
          - Interface in ports/primary
          - Implementation in app/

       4. Add CLI commands
          - orc preferences get/set

       Create these 4 tasks? [y/n/edit]
```
