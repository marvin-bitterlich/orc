---
name: ship-freshness
description: Rebase branch against master and validate tasks/notes are still relevant. Use when returning to a shipment after time away, or before starting implementation on a stale branch.
---

# Ship Freshness

Rebase the current branch against latest master and validate that tasks/notes are still relevant to the updated codebase.

## Usage

```
/ship-freshness              (freshen focused shipment)
/ship-freshness SHIP-xxx     (freshen specific shipment)
```

## When to Use

- Returning to a shipment after time away
- Before starting implementation on a stale branch
- After significant changes merged to master by others
- Reality check before /imp-start

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

### Step 3: Get Branch Info

```bash
orc shipment show SHIP-xxx
git branch --show-current
```

Identify the shipment branch and current branch.

### Step 4: Fetch and Rebase

```bash
git fetch origin master
git rebase origin/master
```

**If conflicts:**
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

### Step 5: Show What Changed

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

```
Ship Freshness Complete:

  Branch: rebased ✓
  Master commits: X new
  Tasks: Y validated (Z need review)
  Notes: A validated (B need review)

Ready to proceed:
  /imp-start     - Begin implementation
  /ship-plan     - Re-plan if needed
```

## Error Handling

| Error | Action |
|-------|--------|
| No focused shipment | Ask for shipment ID |
| Rebase conflicts | Report conflicts, stop |
| No branch assigned | "Shipment has no branch. Assign with orc shipment update" |
| Too many notes | Suggest /ship-synthesize first |

## Notes

- This is a reality check, not a full re-synthesis
- Does not modify tasks/notes automatically - reports findings
- User decides whether to update based on findings
- Run before /imp-start on long-running shipments
