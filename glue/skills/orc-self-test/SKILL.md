---
name: orc-self-test
description: Integration self-testing for ORC. Tests the plan/apply infrastructure pattern by creating test entities, applying infrastructure, verifying state, and cleaning up. Use when you want to verify ORC is working correctly end-to-end.
---

# ORC Self-Test

Run integration tests for the plan/apply infrastructure pattern.

## Prerequisites

- `~/src/orc-canary` must exist as a git repository (test repo for worktrees)
- ORC must be installed and working (`orc --version`)

## Flow

### 1. Setup Test Factory

```bash
orc factory create "[TEST] Self-Test Factory"
```

Capture the factory ID (e.g., `FACT-xxx`).

### 2. Create Test Workshop

```bash
orc workshop create "[TEST] Self-Test Workshop" --factory FACT-xxx
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

### 5. Apply Infrastructure

```bash
orc infra apply WORK-xxx --yes
```

Verify output shows:
- Gatehouse created
- Workbenches created (or "nothing to do" if already exists)

### 6. Verify Filesystem State

```bash
ls -la ~/.orc/ws/WORK-xxx-*/
ls -la ~/.orc/ws/WORK-xxx-*/.orc/config.json
```

Verify:
- Gatehouse directory exists
- Config file exists with proper content

### 7. Cleanup

Delete test entities in reverse order:

```bash
orc workbench delete BENCH-xxx --force
orc workshop delete WORK-xxx --force
orc factory delete FACT-xxx --force
```

Remove test directories:

```bash
rm -rf ~/.orc/ws/WORK-xxx-*
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
[PASS] Infra apply creates filesystem
[PASS] Gatehouse directory exists
[PASS] Config file exists
[PASS] Cleanup successful

All tests passed!
```

## On Failure

If any step fails:
1. Report which step failed and the error
2. Attempt cleanup of any created resources
3. Suggest running `orc doctor` for diagnostics
