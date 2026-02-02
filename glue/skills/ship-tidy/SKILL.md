---
name: ship-tidy
description: Tidy up a shipment by reviewing tasks, updating descriptions, and organizing work. Use when user says /ship-tidy or wants to clean up shipment state.
---

# Ship Tidy Skill

Review and organize a shipment's tasks and state.

## Usage

```
/ship-tidy              (tidy focused shipment)
/ship-tidy SHIP-xxx     (tidy specific shipment)
```

## Flow

### Step 1: Get Shipment

If argument provided:
- Use specified SHIP-xxx

If no argument:
- Get focused shipment from `orc focus --show`
- If no focus, ask which shipment to tidy

### Step 2: Review Current State

```bash
orc shipment show <SHIP-xxx>
orc task list --shipment <SHIP-xxx>
```

Display:
- Shipment title, description, status
- All tasks with their status
- Identify any issues (stale tasks, unclear descriptions)

### Step 3: Suggest Improvements

Review and suggest:
- Tasks that could be combined or split
- Unclear task titles that need rewording
- Missing tasks based on shipment scope
- Tasks that might be out of scope
- Status updates needed

### Step 4: Apply Changes (with confirmation)

For each suggested change, ask user to confirm before:
- Updating task titles/descriptions: `orc task update TASK-xxx --title "..." --description "..."`
- Deleting unnecessary tasks: `orc task delete TASK-xxx`
- Creating missing tasks: `orc task create "..." --shipment SHIP-xxx`
- Reordering (if applicable)

### Step 5: Summary

Output:
```
Shipment tidied:
  SHIP-xxx: <Title>
  Tasks: X total (Y ready, Z in_progress, W complete)

Changes made:
  - Updated: TASK-xxx title
  - Created: TASK-yyy
  - Deleted: TASK-zzz
```

## Example Session

```
User: /ship-tidy

Agent: [gets focused shipment SHIP-250]
       [runs orc shipment show SHIP-250]
       [runs orc task list --shipment SHIP-250]

Agent: Reviewing SHIP-250: Core Model Simplification

       Current tasks:
       - TASK-627 [complete]: Add shipment status field
       - TASK-628 [complete]: Implement transitions
       - TASK-629 [complete]: Build migration command
       - TASK-630 [in_progress]: Update summary output

       Suggestions:
       1. TASK-630 description could be more specific

       Apply changes? [y/n]
```
