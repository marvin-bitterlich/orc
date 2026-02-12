---
name: self-test
description: Orchestrate all ORC integration checks using Claude Teams. Runs orc doctor, then spawns parallel teammates for orc-tmux-check, orc-hello-check, docs-doctor, and optional make-bootstrap-check.
---

# Self-Test Runner

Orchestrate all ORC integration checks. This is a **team-lead skill** that runs preflight checks directly, then spawns a Claude Team for parallel execution of all integration checks.

## Usage

```
/self-test
```

## Prerequisites

- ORC installed and working (`orc --version`)
- `./orc` local binary built (`make dev`)

## Flow

### Phase 1: Preflight (team lead runs directly)

#### Step 1: Run orc doctor

```bash
./orc doctor
```

If `orc doctor` reports any unhealthy checks, **stop immediately** and report the failures to the user. Do not proceed to Phase 2 until the environment is healthy.

#### Step 2: Detect optional tooling

Check for tart and sshpass availability:

```bash
command -v tart && echo "tart: available" || echo "tart: not found"
command -v sshpass && echo "sshpass: available" || echo "sshpass: not found"
```

#### Step 3: Present preflight summary

Display a summary table to the user:

```
Self-Test Preflight
--------------------
ORC Doctor:    PASS
tart:          [available / not found]
sshpass:       [available / not found]

Available checks:
  1. orc-tmux-check      (plan/apply lifecycle test)
  2. orc-hello-check     (first-run flow test)
  3. Docs Doctor         (documentation validation)
  4. make-bootstrap-check [only if tart + sshpass available]
```

#### Step 4: Ask user which checks to run

Use `AskUserQuestion` to ask:

```
Which checks should I run? (comma-separated numbers, or "all")
Example: 1,2,3 or all
```

If the user selects check 4 but tart/sshpass are missing, warn them and exclude it.

---

### Phase 2: Team Execution

After user confirms, create a Claude Team and spawn one teammate per selected check. All teammates run in parallel.

**Important:** Each teammate should use `./orc` (local binary) for all ORC commands.

---

#### Teammate 1 -- orc-tmux-check

Spawn a teammate with this prompt:

```
You are running the ORC infra integration check. Use ./orc (local binary) for all commands.

Follow these steps exactly:

1. CREATE TEST FACTORY
   ./orc factory create "[TEST] Self-Test Infra $(date +%s)"
   Capture the factory ID (FACT-xxx).

2. CREATE TEST WORKSHOP
   ./orc workshop create --factory FACT-xxx --name "[TEST] Self-Test Workshop"
   Capture workshop ID (WORK-xxx).

3. CREATE TEST WORKBENCH
   ./orc workbench create --workshop WORK-xxx --repo-id REPO-001
   Capture workbench ID (BENCH-xxx).

4. CHECK INFRASTRUCTURE PLAN
   ./orc infra plan WORK-xxx
   Verify plan shows CREATE actions for workbench, tmux session, and tmux windows.

5. APPLY INFRASTRUCTURE
   ./orc infra apply WORK-xxx --yes
   Verify output shows workbenches and tmux session created.

6. VERIFY FILESYSTEM STATE
   ls -la ~/wb/BENCH-xxx-*/
   Confirm workbench directory exists.

7. VERIFY TMUX STATE
   tmux list-sessions | grep "Self-Test"
   tmux list-windows -t "[TEST] Self-Test Workshop"
   Confirm session and window exist.

8. ARCHIVE AND CLEANUP
   ./orc workbench archive BENCH-xxx
   ./orc workshop archive WORK-xxx
   ./orc infra apply WORK-xxx --yes
   Verify worktree directory removed and tmux window removed.

9. DELETE DB ENTITIES
   ./orc factory delete FACT-xxx --force

10. VERIFY CLEANUP
    ls ~/wb/BENCH-xxx-* 2>/dev/null && echo "ERROR: Orphan directory found" || echo "OK: No orphans"
    tmux list-sessions 2>/dev/null | grep "Self-Test" && echo "ERROR: Session still exists" || echo "OK: Session cleaned"

Report results as:
  [PASS] or [FAIL] for each step with details.
  If any step fails, attempt cleanup (archive then force delete) before reporting.

Final summary format:
  Infra Check: PASS or FAIL
  Details: [list of per-step results]
```

---

#### Teammate 2 -- orc-hello-check

Spawn a teammate with this prompt:

```
You are running the ORC hello integration check. Use ./orc (local binary) for all commands.
This is an AUTOMATED version of the hello exercise -- no user interaction required.

Follow these steps exactly:

1. CREATE TEST FACTORY
   TEST_SUFFIX=$(date +%s)
   ./orc factory create "Hello Check $TEST_SUFFIX"
   Capture factory ID (FACT-xxx).

2. CREATE TEST COMMISSION
   ./orc commission create "Hello Test $TEST_SUFFIX"
   Capture commission ID (COMM-xxx).

3. CREATE TEST WORKSHOP
   ./orc workshop create --factory FACT-xxx --name "Hello Workshop $TEST_SUFFIX"
   Capture workshop ID (WORK-xxx).

4. CREATE TEST WORKBENCH
   ./orc workbench create --workshop WORK-xxx --repo-id REPO-001
   Capture workbench ID (BENCH-xxx).

5. APPLY INFRASTRUCTURE
   ./orc infra apply WORK-xxx --yes
   Verify infrastructure is created.

6. RUN SUMMARY
   ./orc summary
   Verify the summary output shows the test entities and system is interactive.
   This is the success criterion: reaching this point means the hello flow works.

7. CLEANUP
   # Archive in correct order
   ./orc workbench archive BENCH-xxx
   ./orc workshop archive WORK-xxx
   ./orc infra apply WORK-xxx --yes
   ./orc commission archive COMM-xxx
   # Delete remaining DB entities
   ./orc factory delete FACT-xxx --force

Report results as:
  [PASS] or [FAIL] for each step with details.
  If any step fails, attempt cleanup (archive + infra apply) before reporting.

Final summary format:
  Hello Check: PASS or FAIL
  Details: [list of per-step results]
```

---

#### Teammate 3 -- docs-doctor

Spawn a teammate with this prompt:

```
You are running the ORC docs-doctor check.

Follow the docs-doctor skill exactly as defined in .claude/skills/docs-doctor/SKILL.md.
Read that file first for the full instructions.

The docs-doctor skill uses a fan-out pattern with haiku subagents for parallel validation.
You should follow that pattern: spawn haiku subagents for each check category
(structural, lane, CLI, schema, getting started coherence, ER diagram),
collect findings, and synthesize a report.

Use ./orc (local binary) for any ORC commands (e.g., validating CLI commands).

Final summary format:
  Docs Doctor: PASS or FAIL
  Details: [per-category results: structural, lanes, CLI, schema, getting-started, ER diagram]
```

---

#### Teammate 4 -- make-bootstrap-check (optional, only if tart available)

Only spawn this teammate if tart and sshpass were detected in preflight.

Spawn a teammate with this prompt:

```
You are running the ORC VM bootstrap test.

Run the bootstrap test in a fresh macOS VM:

  make bootstrap-test

This executes scripts/bootstrap-test.sh which:
1. Creates a fresh macOS VM using tart
2. Installs Go via Homebrew
3. Copies ORC repo into VM
4. Runs make bootstrap
5. Verifies orc is in PATH and works
6. Verifies bootstrap artifacts (FACT-001, REPO-001)
7. Verifies CLI functionality
8. Cleans up VM on success

Report the full output and:

Final summary format:
  VM Bootstrap Test: PASS or FAIL
  Details: [output summary]
```

---

### Phase 3: Result Collection

After all teammates complete, collect their results and synthesize a single summary report.

#### Summary Report Format

```
ORC Self-Test Results
=====================

Preflight:
  ORC Doctor:    PASS
  tart:          [available / not found]
  sshpass:       [available / not found]

Check Results:
  +----------------------+--------+---------------------------+
  | Check                | Result | Notes                     |
  +----------------------+--------+---------------------------+
  | orc-tmux-check       | PASS   |                           |
  | orc-hello-check      | PASS   |                           |
  | Docs Doctor          | PASS   | 1 warning (auto-fixable)  |
  | make-bootstrap-check | SKIP   | tart not available        |
  +----------------------+--------+---------------------------+

Overall: PASS (3/3 selected checks passed)
```

Use `SKIP` for checks the user did not select or that were unavailable.

If any check failed, include the failure details below the summary table.

#### Team Shutdown

After collecting all results and presenting the summary, shut down all teammates.

## On Failure

If any check fails:
1. Report which check failed with full details from the teammate
2. The teammate should have already attempted cleanup
3. Suggest running `orc doctor` for diagnostics
4. Suggest running the individual check for more detail
