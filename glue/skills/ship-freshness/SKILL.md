---
name: ship-freshness
description: Rebase branch against master (or sync from upstream for fork repos) and validate tasks/notes are still relevant. Use when returning to a shipment after time away, or before starting implementation on a stale branch.
---

# Ship Freshness

Rebase the current branch against latest master (or sync from upstream for fork repos) and validate that tasks/notes are still relevant to the updated codebase.

## Usage

```
/ship-freshness              (freshen focused shipment)
/ship-freshness SHIP-xxx     (freshen specific shipment)
```

## When to Use

- Returning to a shipment after time away
- Before starting implementation on a stale branch
- After significant changes merged to master by others
- After an upstream PR is merged (fork repos)
- Reality check before starting implementation work

## Flow

### Step 1: Get Shipment

If argument provided:
- Use specified SHIP-xxx

If no argument:
```bash
orc focus --show
```
- Use focused shipment
- If no focus: "No shipment focused. Run `orc focus SHIP-xxx` first."

### Step 2: Check Note Count

```bash
orc note list --shipment SHIP-xxx
```

Count open notes. If 5 or more:
```
⚠️ Found X open notes in SHIP-xxx.

Consider running /ship-synthesize first to compact notes before freshening.
Proceed anyway? [y/n]
```

### Step 3: Get Branch Info and Check Fork Mode

```bash
orc shipment show SHIP-xxx
git branch --show-current
```

Identify the shipment branch and current branch.

Then check if the repo is in fork mode:

```bash
orc repo show <repo-id>
```

If `upstream_url` is present, the repo is in fork mode. Store:
- **upstream_url**: The canonical repo URL
- **upstream_branch**: The target branch on upstream (defaults to `default_branch`)

### Step 4: Fetch and Rebase

**If upstream URL is set (fork mode):**

Sync the local default branch from upstream, then rebase the worktree branch:

```bash
git fetch upstream
git checkout <default_branch>
git merge upstream/<upstream_branch>
git push origin <default_branch>
git checkout <worktree_branch>
git rebase <default_branch>
```

**If no upstream URL (direct mode):**

```bash
git fetch origin master
git rebase origin/master
```

**If conflicts (either mode):**
```
❌ Rebase conflicts detected:

  CONFLICT: path/to/file.go

Resolve conflicts manually, then run:
  git rebase --continue

Or abort with:
  git rebase --abort
```
Stop here - do not proceed until conflicts resolved.

**If clean:**
```
✓ Branch rebased cleanly against master
```

**If fork mode and merge conflicts during upstream sync:**
```
❌ Merge conflicts syncing from upstream:

  CONFLICT: path/to/file.go

Resolve conflicts in <default_branch>, then:
  git add <resolved-files>
  git commit
  git push origin <default_branch>
  git checkout <worktree_branch>
  git rebase <default_branch>
```
Stop here - do not proceed until conflicts resolved.

### Step 5: Show What Changed

**If fork mode:**

```bash
git log --oneline <default_branch>@{1}..<default_branch>
```

Display recent upstream commits synced to the default branch:
```
Changes from upstream since last sync:

  abc1234 feat: add new CLI command
  def5678 fix: resolve edge case in guards
  ghi9012 docs: update architecture

X commits synced from upstream.
```

**If direct mode:**

```bash
git log --oneline origin/master@{1}..origin/master
```

Display recent commits to master since last fetch:
```
Changes in master since last sync:

  abc1234 feat: add new CLI command
  def5678 fix: resolve edge case in guards
  ghi9012 docs: update architecture

X commits merged to master.
```

### Step 6: Validate Tasks

```bash
orc task list --shipment SHIP-xxx
```

For each task, consider:
- Does the task still make sense given master changes?
- Are there new conflicts or dependencies?
- Has the work already been done in master?

Report findings:
```
Task Validation:

  TASK-xxx: Create new component
    Status: Still relevant ✓

  TASK-yyy: Fix bug in handler
    ⚠️ May be affected by commit def5678
    Review: internal/handler.go was modified
```

### Step 7: Validate Notes

```bash
orc note list --shipment SHIP-xxx --status open
```

For each open note:
- Is the information still accurate?
- Have assumptions changed?

Report findings:
```
Note Validation:

  NOTE-xxx: Architecture decision
    Status: Still accurate ✓

  NOTE-yyy: Performance concern
    ⚠️ May need update - related code changed
```

### Step 8: Summary

**Direct mode:**
```
Ship Freshness Complete:

  Branch: rebased ✓
  Master commits: X new
  Tasks: Y validated (Z need review)
  Notes: A validated (B need review)

Ready to proceed:
  /ship-plan     - Re-plan if needed
  orc task list  - View tasks to work on
```

**Fork mode:**
```
Ship Freshness Complete (fork):

  Upstream sync: ✓
  Branch: rebased against <default_branch> ✓
  Fork origin: pushed ✓
  Upstream commits: X synced
  Tasks: Y validated (Z need review)
  Notes: A validated (B need review)

Ready to proceed:
  /ship-plan     - Re-plan if needed
  orc task list  - View tasks to work on
```

## Error Handling

| Error | Action |
|-------|--------|
| No focused shipment | Ask for shipment ID |
| Rebase conflicts | Report conflicts, stop |
| No branch assigned | "Shipment has no branch. Assign with orc shipment update" |
| Too many notes | Suggest /ship-synthesize first |
| `upstream` remote not configured | Suggest `git remote add upstream <upstream_url>` |
| Upstream fetch fails | Check URL and network connectivity |
| Merge conflicts during upstream sync | Report conflicts, stop before rebase |

## Notes

- This is a reality check, not a full re-synthesis
- Does not modify tasks/notes automatically - reports findings
- User decides whether to update based on findings
- Run before starting implementation on long-running shipments
