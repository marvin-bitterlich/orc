---
name: orc-workbench
description: Guide through creating a workbench. Use when user says /orc-workbench or wants to create a new workbench for implementation work.
---

# Workbench Creation Skill

Guide users through creating a workbench for implementation work.

## Usage

```
/orc-workbench
```

Workbench name is auto-generated as `{repo}-{number}` based on the linked repo.

## Flow

### Step 1: Detect Context

```bash
orc status
orc workshop list
```

Identify:
- Current workshop (if in workshop context)
- Available workshops
- If no workshop detected, ask which workshop

### Step 2: Select Repo (Required)

```bash
orc repo list
```

Ask user which repo to link. The workbench name will be auto-generated from this.

### Step 3: Create Workbench Record

```bash
orc workbench create --workshop WORK-xxx --repo-id REPO-xxx
```

The name is auto-generated as `{repo}-{number}` (e.g., `intercom-015`).

### Step 4: Start TMux Session

```bash
orc tmux apply WORK-xxx --yes
```

This creates/reconciles the tmux session for the workshop, adding the new workbench window with the standard pane layout.

### Step 5: Confirm Ready

Output:
```
Workbench created:
  BENCH-015: intercom-015
  Path: ~/wb/intercom-015
  Workshop: WORK-xxx

To start working:
  cd ~/wb/intercom-015
  # Or attach to TMux: orc tmux connect WORK-xxx
```

## Example Session

```
User: /orc-workbench

Agent: [runs orc status, detects WORK-003]
       [runs orc repo list, user selects REPO-001 (intercom)]
       [runs orc workbench create --workshop WORK-003 --repo-id REPO-001]
       [runs orc tmux apply WORK-003 --yes]

Agent: Workbench created:
         BENCH-015: intercom-015
         Path: ~/wb/intercom-015
         Workshop: WORK-003

       To start working:
         cd ~/wb/intercom-015
```
