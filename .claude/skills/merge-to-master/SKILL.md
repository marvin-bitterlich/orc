---
name: merge-to-master
description: |
  Merge current worktree branch to master at ~/src/orc with full validation.
  Use when user says "/merge-to-master", "merge to master", "ship to master",
  or wants to integrate their worktree changes into the main ORC repository.
  Handles pre-flight checks, merge, hook installation, and post-merge cleanup.
---

# Merge to Master

Merge the current worktree branch into master at `~/src/orc` with full validation.

## Workflow

### 1. Pre-flight Checks

Before merging, verify:

```bash
# Must be clean
git status --porcelain  # Should be empty

# Must pass lint
make lint

# Detect current branch (the one to merge)
BRANCH=$(git rev-parse --abbrev-ref HEAD)
```

If any check fails, stop and report the issue.

### 2. Merge to Master

```bash
cd ~/src/orc
git checkout master
git pull origin master
git merge <BRANCH> --no-edit
```

Report merge result (fast-forward or merge commit).

### 3. Install Hooks

```bash
make init
```

Verify both hooks exist:
```bash
ls -la .git/hooks/{pre-commit,post-merge}
```

### 4. Post-merge Hook Validation

For fast-forward merges, the post-merge hook doesn't trigger automatically.
Run it manually to rebuild:

```bash
.git/hooks/post-merge
```

Confirm output shows:
- `make install` completed
- `make clean` completed

### 5. Push to Origin

```bash
git push origin master
```

### 6. Rebase Worktree Branch

Return to the worktree and rebase:

```bash
cd <original-worktree-path>
git rebase master
```

### 7. Close Shipment

If there's a focused shipment with all tasks complete, mark it complete:

```bash
orc status  # Check focused shipment
orc task list --shipment SHIP-XXX --status ready  # Verify no remaining tasks
orc shipment complete SHIP-XXX
```

### 8. Notify Goblin

Mail and nudge the goblin with a summary of completed work:

```bash
# Find the gatehouse for the current workshop
orc workshop show  # Note the gatehouse ID (GATE-XXX)

orc mail send "<summary of completed tasks and commit hash>" \
  --to GOBLIN-GATE-XXX \
  --subject "SHIP-XXX Complete" \
  --nudge
```

Include in the message:
- Shipment ID and title
- List of completed tasks with brief descriptions
- Commit hash on master

## Success Output

Report completion with:
- Branch merged
- Hooks verified
- Master pushed
- Worktree rebased
- Shipment closed (if applicable)
- Goblin notified and nudged

## Error Handling

| Error | Action |
|-------|--------|
| Dirty working tree | List uncommitted files, ask to commit or stash |
| Lint fails | Show failures, do not proceed |
| Merge conflicts | Report conflicts, do not auto-resolve |
| Push rejected | Check if remote has new commits, suggest pull --rebase |
