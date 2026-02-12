---
name: orc-workshop-archive
description: Guide through archiving a workshop safely. Use when user says /orc-workshop-archive or wants to archive a workshop.
---

# Workshop Archive Skill

Interactive skill that walks users through safely archiving a workshop and its workbenches.

## Usage

```
/orc-workshop-archive
/orc-workshop-archive WORK-xxx
```

## Philosophy

This skill is an expert **user** of the ORC CLI, not an expert builder. If commands fail or syntax is unclear, consult `--help`:

```bash
orc workshop archive --help
orc workbench archive --help
orc tmux apply --help
orc infra cleanup --help
```

## Flow

### Step 1: Select Workshop

If workshop ID provided as argument, use it. Otherwise:

```bash
orc workshop list
```

Display the list and ask:

> "Which workshop would you like to archive?"

### Step 2: List Workbenches

```bash
orc workbench list --workshop WORK-xxx
```

Show the workbenches and their paths. If no workbenches:

> "Workshop has no workbenches. Proceeding to archive workshop."

Skip to Step 4.

### Step 3: Check Each Workbench

For each workbench, perform safety checks before archiving:

**3a. Check if path exists**

```bash
orc workbench show BENCH-xxx
```

Get the path. If path doesn't exist on filesystem, skip git checks:

> "Workbench BENCH-xxx path doesn't exist. Skipping git checks."

Archive directly: `orc workbench archive BENCH-xxx`

**3b. Check for uncommitted changes**

```bash
git -C <path> status --short
```

If output is non-empty:

> "⚠️ Workbench BENCH-xxx has uncommitted changes:
> [show changes]
>
> Would you like to:
> 1. Commit changes (I'll help you)
> 2. Proceed anyway (changes will be lost when directory removed)
> 3. Abort archiving"

If user chooses to commit, guide through commit process. If user declines, warn but continue.

**3c. Check for unpushed commits**

```bash
git -C <path> log @{u}.. --oneline 2>/dev/null
```

If output is non-empty (and command succeeded):

> "⚠️ Workbench BENCH-xxx has unpushed commits:
> [show commits]
>
> Would you like to:
> 1. Push commits
> 2. Proceed anyway (commits exist locally but not on remote)
> 3. Abort archiving"

If user chooses to push, run `git -C <path> push`. If user declines, warn but continue.

**Note:** If `@{u}` fails (no upstream), skip this check.

**3d. Archive workbench**

```bash
orc workbench archive BENCH-xxx
```

Repeat for each workbench.

### Step 4: Archive Workshop

```bash
orc workshop archive WORK-xxx
```

### Step 5: Cleanup Offer

Offer infrastructure cleanup:

> "Workshop archived. Would you like to clean up infrastructure?
>
> 1. Kill TMux session (`orc tmux apply WORK-xxx` reconciles tmux state)
> 2. Remove directories (`orc infra cleanup` removes orphan paths)
> 3. Both
> 4. Skip cleanup (leave infrastructure in place)"

Based on choice:

**Kill TMux session:**
```bash
orc tmux apply WORK-xxx
```

**Remove directories:**
```bash
orc infra cleanup
```

### Step 6: Confirm Complete

Output summary:

```
Workshop WORK-xxx archived:
  Workbenches archived: N
  TMux session: [cleaned / intact]
  Directories: [removed / intact]

The workshop and workbenches remain in the database but are marked archived.
```

## Example Session

```
User: /orc-workshop-archive

Agent: Which workshop would you like to archive?
       [runs orc workshop list]

       WORK-001: orc development
       WORK-002: dlq cleanup
       WORK-003: auth refactor

User: WORK-002

Agent: Checking workbenches for WORK-002...
       [runs orc workbench list --workshop WORK-002]

       BENCH-015: intercom-015 (~/wb/intercom-015)
       BENCH-016: orc-016 (~/wb/orc-016)

       Checking BENCH-015...
       [runs git -C ~/wb/intercom-015 status --short]

       ⚠️ Workbench BENCH-015 has uncommitted changes:
         M src/app.rb
         ?? tmp/debug.log

       Would you like to:
       1. Commit changes
       2. Proceed anyway
       3. Abort

User: 2

Agent: Proceeding with uncommitted changes.
       [runs git -C ~/wb/intercom-015 log @{u}.. --oneline]

       No unpushed commits.
       [runs orc workbench archive BENCH-015]
       ✓ BENCH-015 archived

       Checking BENCH-016...
       [runs git -C ~/wb/orc-016 status --short]

       Clean working directory.
       [runs git -C ~/wb/orc-016 log @{u}.. --oneline]

       No unpushed commits.
       [runs orc workbench archive BENCH-016]
       ✓ BENCH-016 archived

       [runs orc workshop archive WORK-002]
       ✓ Workshop WORK-002 archived

       Would you like to clean up infrastructure?
       1. Kill TMux session
       2. Remove directories
       3. Both
       4. Skip

User: 3

Agent: [runs orc tmux apply WORK-002]
       ✓ TMux windows removed

       [runs orc infra cleanup]
       ✓ Orphan directories removed

       Workshop WORK-002 archived:
         Workbenches archived: 2
         TMux session: cleaned
         Directories: removed
```

## Error Handling

| Error | Remediation |
|-------|-------------|
| Workshop not found | Show available workshops |
| Workbench path doesn't exist | Skip git checks, archive anyway |
| No upstream branch | Skip unpushed check |
| Git command fails | Show error, offer to proceed anyway |
| Archive command fails | Show error, ask user how to proceed |

## CLI Reference

| Command | Purpose |
|---------|---------|
| `orc workshop list` | List workshops |
| `orc workshop archive WORK-xxx` | Archive workshop |
| `orc workbench list --workshop WORK-xxx` | List workbenches in workshop |
| `orc workbench show BENCH-xxx` | Get workbench details (including path) |
| `orc workbench archive BENCH-xxx` | Archive workbench |
| `orc tmux apply WORK-xxx` | Reconcile tmux session (removes orphan windows) |
| `orc infra cleanup` | Remove orphan directories |
