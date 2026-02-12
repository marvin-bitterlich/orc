---
name: orc-self-test
description: Integration self-testing for ORC. Tests the gotmux-based tmux workflow by creating test entities, running apply, and verifying infrastructure. Use when you want to verify ORC is working correctly end-to-end.
---

# ORC Self-Test

Run integration tests for the gotmux-based tmux workflow and workbench infrastructure.

## Safety: Default TMux Server

**CRITICAL**: The `orc tmux apply` command operates on the default tmux server. The test
creates a uniquely named session (`orc-self-test`) to avoid collisions. Always use exact
session matching (`-t "=$SESSION_NAME"`) for all verification commands.

**Rule**: Always qualify tmux targets as `session:window.pane`. Always use exact matching
with `=` prefix. Always run cleanup even on failure.

## Prerequisites

- `./orc` must be available (use `make dev` to build)
- tmux must be installed
- Test repository `orc-canary` must exist (check `orc repo list`)

## Variables

Throughout the test, track these variables. Print them after creation for debugging:

```
TEST_REPO_ID=<from orc repo list>
FACTORY_ID=<created>
WORKSHOP_ID=<created>
WORKBENCH_ID=<created>
WORKBENCH_NAME=<created>
WORKBENCH_PATH=<created>
SESSION_NAME=<workshop name>
```

## Flow

### 1. Detect Test Repository

```bash
TEST_REPO_ID=$(./orc repo list | grep orc-canary | awk '{print $1}')
if [ -z "$TEST_REPO_ID" ]; then
  echo "[FAIL] orc-canary repo not registered. Run: orc repo create orc-canary --path ~/src/orc-canary --default-branch master"
  exit 1
fi
echo "[PASS] Test repo found: $TEST_REPO_ID"
```

### 2. Create Test Entities

```bash
FACTORY_ID=$(./orc factory create "orc-self-test" 2>&1 | grep -o 'FACT-[0-9]*')
echo "Created factory: $FACTORY_ID"

WORKSHOP_ID=$(./orc workshop create --factory "$FACTORY_ID" --name "orc-self-test" 2>&1 | grep -o 'WORK-[0-9]*')
echo "Created workshop: $WORKSHOP_ID"

./orc workbench create --workshop "$WORKSHOP_ID" --repo-id "$TEST_REPO_ID"
# Parse output for BENCH ID, name, path
```

### 3. Verify Filesystem Created (Immediate)

**Key test**: Workbench creation is atomic (DB + worktree + config in one operation).

```bash
# Check worktree exists
[ -d "$WORKBENCH_PATH" ] || { echo "[FAIL] Worktree not created"; exit 1; }
echo "[PASS] Worktree exists: $WORKBENCH_PATH"

# Check config file exists and has correct ID
[ -f "$WORKBENCH_PATH/.orc/config.json" ] || { echo "[FAIL] Config not created"; exit 1; }
echo "[PASS] Config file exists"

grep -q "$WORKBENCH_ID" "$WORKBENCH_PATH/.orc/config.json" || { echo "[FAIL] Config has wrong ID"; exit 1; }
echo "[PASS] Config contains correct workbench ID"
```

### 4. Apply TMux Session

Use `orc tmux apply --yes` to create the session. This replaces the old `orc tmux start`.

```bash
# Apply session via ORC (creates session + windows + enrichment in one command)
./orc tmux apply "$WORKSHOP_ID" --yes

SESSION_NAME="orc-self-test"

# Verify session exists
if ! tmux has-session -t "=$SESSION_NAME" 2>/dev/null; then
  echo "[FAIL] TMux session not created: $SESSION_NAME"
  # Run cleanup
  exit 1
fi
echo "[PASS] TMux session created: $SESSION_NAME"
```

**Important**: Use `-t "=$SESSION_NAME"` (with `=` prefix) for exact session matching.
This prevents tmux from matching partial names against other sessions.

### 5. Verify Pane Structure

```bash
# Use exact session targeting to avoid cross-session interference
# List panes with session:window qualification
PANES=$(tmux list-panes -t "=$SESSION_NAME" -F "#{pane_index}:#{pane_current_path}")

echo "Panes found:"
echo "$PANES"

PANE_COUNT=$(echo "$PANES" | wc -l | tr -d ' ')
if [ "$PANE_COUNT" -ne 3 ]; then
  echo "[FAIL] Expected 3 panes, found $PANE_COUNT"
  exit 1
fi
echo "[PASS] Found 3 panes"

# Verify all panes are in workbench directory
while IFS=: read -r idx path; do
  if [ "$path" != "$WORKBENCH_PATH" ]; then
    echo "[FAIL] Pane $idx in wrong directory: $path (expected: $WORKBENCH_PATH)"
    exit 1
  fi
done <<< "$PANES"
echo "[PASS] All panes in correct directory"
```

### 6. Verify Pane Options (@pane_role, @bench_id, @workshop_id)

Pane identity uses tmux pane options (NOT shell env vars -- those can't be
read by tmux format strings). Verify using `#{@pane_role}`, `#{@bench_id}`,
and `#{@workshop_id}` format strings.

```bash
# Get pane indices (respects pane-base-index setting)
PANE_INDICES=$(tmux list-panes -t "=$SESSION_NAME" -F "#{pane_index}")

# Expected roles in order: vim, goblin, shell
# Use case statement instead of bash array (more reliable in all contexts)
IDX=0

while read -r pane_idx; do
  TARGET="=$SESSION_NAME.$pane_idx"

  # Check @pane_role
  ROLE=$(tmux display-message -t "$TARGET" -p '#{@pane_role}')
  case $IDX in
    0) EXPECTED="vim" ;;
    1) EXPECTED="goblin" ;;
    2) EXPECTED="shell" ;;
  esac

  if [ "$ROLE" != "$EXPECTED" ]; then
    echo "[FAIL] Pane $pane_idx: expected @pane_role=$EXPECTED, got '$ROLE'"
    exit 1
  fi
  echo "[PASS] Pane $pane_idx: @pane_role=$ROLE"

  # Check @bench_id
  BENCH=$(tmux display-message -t "$TARGET" -p '#{@bench_id}')
  if [ -z "$BENCH" ]; then
    echo "[FAIL] Pane $pane_idx: @bench_id not set"
    exit 1
  fi
  if [ "$BENCH" != "$WORKBENCH_ID" ]; then
    echo "[FAIL] Pane $pane_idx: expected @bench_id=$WORKBENCH_ID, got '$BENCH'"
    exit 1
  fi
  echo "[PASS] Pane $pane_idx: @bench_id=$BENCH"

  # Check @workshop_id
  WS=$(tmux display-message -t "$TARGET" -p '#{@workshop_id}')
  if [ -z "$WS" ]; then
    echo "[FAIL] Pane $pane_idx: @workshop_id not set"
    exit 1
  fi
  if [ "$WS" != "$WORKSHOP_ID" ]; then
    echo "[FAIL] Pane $pane_idx: expected @workshop_id=$WORKSHOP_ID, got '$WS'"
    exit 1
  fi
  echo "[PASS] Pane $pane_idx: @workshop_id=$WS"

  IDX=$((IDX + 1))
done <<< "$PANE_INDICES"

echo "[PASS] All pane options set correctly (@pane_role, @bench_id, @workshop_id)"
```

### 7. Verify Vim Width (50% main-pane-width)

```bash
# Get the vim pane width and total window width
WINDOW_WIDTH=$(tmux display-message -t "=$SESSION_NAME" -p '#{window_width}')
VIM_PANE_WIDTH=$(tmux list-panes -t "=$SESSION_NAME" -F '#{pane_width}' | head -1)

# Check vim pane is approximately 50% of window width (within 2 columns tolerance)
HALF_WIDTH=$((WINDOW_WIDTH / 2))
DIFF=$((VIM_PANE_WIDTH - HALF_WIDTH))
if [ "$DIFF" -lt 0 ]; then DIFF=$((-DIFF)); fi

if [ "$DIFF" -le 2 ]; then
  echo "[PASS] Vim pane width ~50% ($VIM_PANE_WIDTH of $WINDOW_WIDTH)"
else
  echo "[FAIL] Vim pane width not ~50%: $VIM_PANE_WIDTH of $WINDOW_WIDTH (diff: $DIFF)"
  exit 1
fi
```

### 8. Test Guest Pane Relocation via Apply

```bash
# Create a guest pane (no @pane_role) in the test session
tmux split-window -t "=$SESSION_NAME" -h "sleep 30"

PANE_COUNT_BEFORE=$(tmux list-panes -t "=$SESSION_NAME" | wc -l | tr -d ' ')
if [ "$PANE_COUNT_BEFORE" -ne 4 ]; then
  echo "[FAIL] Expected 4 panes after guest split, found $PANE_COUNT_BEFORE"
  exit 1
fi
echo "[PASS] Guest pane created (4 panes total)"

# Run apply again -- it should relocate guest panes
./orc tmux apply "$WORKSHOP_ID" --yes

# Verify guest pane relocated (back to 3 in main window)
PANE_COUNT_AFTER=$(tmux list-panes -t "=$SESSION_NAME:$WORKBENCH_NAME" | wc -l | tr -d ' ')
if [ "$PANE_COUNT_AFTER" -ne 3 ]; then
  echo "[FAIL] Expected 3 panes after apply, found $PANE_COUNT_AFTER"
  exit 1
fi
echo "[PASS] Guest pane relocated (3 panes remain in workbench window)"

# Verify -imps window exists
IMPS_WINDOW="${WORKBENCH_NAME}-imps"
if ! tmux list-windows -t "=$SESSION_NAME" -F "#{window_name}" | grep -q "^${IMPS_WINDOW}$"; then
  echo "[FAIL] IMPs window not created"
  exit 1
fi
echo "[PASS] IMPs window exists: $IMPS_WINDOW"
```

### 9. Test Dead Pane Pruning

```bash
# Kill the sleep process in the -imps window to create a dead pane
tmux send-keys -t "=$SESSION_NAME:$IMPS_WINDOW" C-c
sleep 1

# Run apply again -- it should detect and prune the dead pane (or kill the -imps window)
./orc tmux apply "$WORKSHOP_ID" --yes

# The -imps window should be gone (it only had the one dead pane)
if tmux list-windows -t "=$SESSION_NAME" -F "#{window_name}" | grep -q "^${IMPS_WINDOW}$"; then
  echo "[FAIL] -imps window still exists after pruning"
  exit 1
fi
echo "[PASS] Dead pane pruned, empty -imps window killed"
```

### 10. Test Idempotency

```bash
# Run apply again -- should be a no-op (except enrichment and layout reconciliation)
./orc tmux apply "$WORKSHOP_ID" --yes

# Verify session still exists and has correct pane count
if ! tmux has-session -t "=$SESSION_NAME" 2>/dev/null; then
  echo "[FAIL] Session disappeared after idempotent apply"
  exit 1
fi

PANE_COUNT=$(tmux list-panes -t "=$SESSION_NAME:$WORKBENCH_NAME" | wc -l | tr -d ' ')
if [ "$PANE_COUNT" -ne 3 ]; then
  echo "[FAIL] Expected 3 panes after idempotent apply, found $PANE_COUNT"
  exit 1
fi

# Re-verify pane roles survived
ROLES=$(tmux list-panes -t "=$SESSION_NAME:$WORKBENCH_NAME" -F '#{@pane_role}')
ROLE_COUNT=$(echo "$ROLES" | grep -c -E '^(vim|goblin|shell)$')
if [ "$ROLE_COUNT" -ne 3 ]; then
  echo "[FAIL] Pane roles not preserved after idempotent apply"
  exit 1
fi
echo "[PASS] Idempotent apply preserved session, panes, and roles"
```

### 11. Verify Auto-Enrichment

`orc tmux apply` applies enrichment automatically. Verify the `@orc_enriched`
window option was set on the workbench window.

```bash
# Verify @orc_enriched window option was set by auto-enrichment
ENRICHED=$(tmux show-options -t "=$SESSION_NAME:$WORKBENCH_NAME" -wqv @orc_enriched 2>/dev/null)
if [ "$ENRICHED" = "1" ]; then
  echo "[PASS] Auto-enrichment applied (@orc_enriched=1)"
else
  echo "[FAIL] Auto-enrichment not applied (expected @orc_enriched=1, got '$ENRICHED')"
  exit 1
fi
```

### 12. Cleanup

**IMPORTANT**: Run cleanup even if tests fail. Always use exact session matching.

```bash
# Kill the tmux session (exact match)
tmux kill-session -t "=$SESSION_NAME" 2>/dev/null

if tmux has-session -t "=$SESSION_NAME" 2>/dev/null; then
  echo "[FAIL] Session still exists after kill"
else
  echo "[PASS] TMux session killed"
fi

# Archive workbench and clean up worktree
./orc workbench archive "$WORKBENCH_ID"
rm -rf "$WORKBENCH_PATH"

# Prune git worktrees
(cd ~/src/orc-canary && git worktree prune 2>/dev/null)

if [ -d "$WORKBENCH_PATH" ]; then
  echo "[FAIL] Worktree still exists: $WORKBENCH_PATH"
else
  echo "[PASS] Worktree cleaned up"
fi
echo "[PASS] Test entities archived"
```

## Success Criteria

```
ORC Self-Test Results
---------------------
[PASS] Test repo found
[PASS] Factory created
[PASS] Workshop created
[PASS] Workbench created (immediate: DB + worktree + config)
[PASS] Worktree exists
[PASS] Config file exists
[PASS] Config correct
[PASS] TMux session created (via apply)
[PASS] Found 3 panes
[PASS] All panes in correct directory
[PASS] @pane_role options set (vim, goblin, shell)
[PASS] @bench_id options set on all panes
[PASS] @workshop_id options set on all panes
[PASS] Vim pane width ~50%
[PASS] Guest pane created
[PASS] Guest pane relocated (via apply)
[PASS] IMPs window created
[PASS] Dead pane pruned, empty -imps window killed
[PASS] Idempotent apply preserved session, panes, and roles
[PASS] Auto-enrichment applied (@orc_enriched=1)
[PASS] TMux session killed
[PASS] Worktree cleaned up
[PASS] Test entities archived

All tests passed!
```

## On Failure

If ANY step fails, you MUST still run cleanup:

1. Report which step failed and the error
2. **Always** attempt cleanup:
   ```bash
   tmux kill-session -t "=orc-self-test" 2>/dev/null
   ./orc workbench archive $WORKBENCH_ID 2>/dev/null
   rm -rf $WORKBENCH_PATH 2>/dev/null
   (cd ~/src/orc-canary && git worktree prune 2>/dev/null)
   ```
3. Suggest running `./orc doctor` for diagnostics

## Safety Rules for Agents

1. **Exact session matching**: Always use `-t "=$SESSION_NAME"` (with `=`) to prevent partial matching
2. **Never bare `tmux` commands**: Always qualify with session:window.pane targets
3. **Never `orc tmux enrich`**: It mutates global tmux bindings on the user's live server
4. **Never `cd` in main shell**: Use subshells `(cd ... && cmd)` to avoid changing agent working directory
5. **Always cleanup**: Even on failure. Kill session, archive workbench, remove worktree
6. **Use `./orc`**: The local binary in the workbench, not global `orc` or `orc-dev`

## Key Architecture Notes

**Removed (vs old architecture):**
- Gatehouse entity (absorbed into workbench)
- `orc infra plan/apply` for tmux (replaced by gotmux)
- Deferred filesystem creation (now immediate)
- Smug YAML config generation (replaced by gotmux programmatic API)
- Separate `orc tmux start` and `orc tmux refresh` commands (unified into `apply`)

**Current patterns:**
- Workbench creation is atomic (DB + worktree + config)
- Gotmux creates tmux sessions programmatically
- `orc tmux apply WORK-xxx --yes` is the single command for session lifecycle
- `apply` uses plan/execute pattern: compares desired (DB) vs actual (tmux) state
- `respawn-pane -k` sets pane root process at creation time (no SendKeys)
- `@pane_role` tmux pane option is authoritative for pane identity
- `@bench_id` and `@workshop_id` tmux pane options provide workbench context
- Shell env vars (`PANE_ROLE`, `BENCH_ID`, `WORKSHOP_ID`) are NOT used -- all identity is via tmux pane options
- `apply` auto-applies enrichment (no separate `orc tmux enrich` step needed)
- Guest pane relocation and dead pane pruning handled automatically by `apply`
