---
name: orc-self-test
description: Integration self-testing for ORC. Tests the plan/apply infrastructure pattern by creating test entities, applying infrastructure, verifying state, and cleaning up. Use when you want to verify ORC is working correctly end-to-end.
---

# ORC Self-Test

Run integration tests for the plan/apply infrastructure pattern, including tmux session management.

## Prerequisites

- `~/src/orc-canary` must exist as a git repository (test repo for worktrees)
- ORC must be installed and working (`orc --version`)
- tmux must be installed

## Flow

### 1. Setup Test Factory

```bash
orc factory create --name "[TEST] Self-Test Factory"
```

Capture the factory ID (e.g., `FACT-xxx`).

### 2. Create Test Workshop

```bash
orc workshop create --factory FACT-xxx --name "[TEST] Self-Test Workshop"
```

Capture the workshop ID (e.g., `WORK-xxx`).

### 3. Create Test Workbench

```bash
orc workbench create test-bench --workshop WORK-xxx
```

Capture the workbench ID (e.g., `BENCH-xxx`).

### 4. Check Infrastructure Plan

```bash
orc infra plan WORK-xxx
```

Verify the plan shows:
- Gatehouse: `CREATE`
- Workbench: `CREATE` (or `MISSING` if no repo linked)
- TMux Session: `CREATE` (session doesn't exist yet)
- TMux Windows: `CREATE` for each workbench

### 5. Apply Infrastructure

```bash
orc infra apply WORK-xxx --yes
```

Verify output shows:
- Gatehouse created
- Workbenches created
- TMux session created (if shown)

### 6. Verify Filesystem State

```bash
ls -la ~/.orc/ws/WORK-xxx-*/
ls -la ~/.orc/ws/WORK-xxx-*/.orc/config.json
```

Verify:
- Gatehouse directory exists
- Config file exists with proper content

### 7. Verify TMux State

```bash
tmux list-sessions | grep -q "Self-Test"
tmux list-windows -t "[TEST] Self-Test Workshop"
```

Verify:
- Session exists
- Window for workbench exists

### 8. Archive and Cleanup (New Pattern)

Archive workbenches first, then workshop:

```bash
# Archive workbench (soft-delete)
orc workbench archive BENCH-xxx

# Archive workshop (requires all workbenches archived)
orc workshop archive WORK-xxx

# Apply infrastructure to remove orphans (workbenches, tmux windows)
orc infra apply WORK-xxx --yes
```

Verify:
- Worktree directory removed
- TMux window removed

### 9. Delete DB Entities

Delete test entities in reverse order:

```bash
orc workbench delete BENCH-xxx --force
orc workshop delete WORK-xxx --force
orc factory delete FACT-xxx --force
```

### 10. Verify Cleanup

```bash
# Verify no orphan directories remain
ls ~/.orc/ws/WORK-xxx-* 2>/dev/null && echo "ERROR: Orphan directory found" || echo "OK: No orphans"

# Verify tmux session cleaned up
tmux list-sessions 2>/dev/null | grep -q "Self-Test" && echo "ERROR: Session still exists" || echo "OK: Session cleaned"
```

## Success Criteria

Report test results:

```
ORC Self-Test Results
---------------------
[PASS] Factory created
[PASS] Workshop created
[PASS] Workbench created (DB only)
[PASS] Infra plan shows correct state
[PASS] Infra plan shows TMux CREATE status
[PASS] Infra apply creates filesystem
[PASS] Gatehouse directory exists
[PASS] Config file exists
[PASS] TMux session exists
[PASS] TMux window exists
[PASS] Workbench archived
[PASS] Workshop archived
[PASS] Infra apply removes orphans
[PASS] TMux window removed
[PASS] Entity cleanup successful

All tests passed!
```

## On Failure

If any step fails:
1. Report which step failed and the error
2. Attempt cleanup of any created resources (try archive first, then force delete)
3. Suggest running `orc doctor` for diagnostics
