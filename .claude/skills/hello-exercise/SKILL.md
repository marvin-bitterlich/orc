---
name: hello-exercise
description: Manual integration test for the orc hello first-run flow. Tests /orc-first-run skill with an isolated test factory on the dev machine.
---

# Hello Exercise

Manual integration test for the orc hello first-run flow using an isolated test factory.

## Usage

```
/hello-exercise
```

Run this to verify the complete first-run experience works correctly without polluting your production data.

## Prerequisites

- ORC installed and working (`orc --version`)
- Claude Code available (`claude --version`)
- REPO-001 exists (created by `make bootstrap`)

## Flow

### 1. Create Test Factory

Generate a unique test factory:

```bash
# Generate unique suffix
TEST_SUFFIX=$(date +%s)
orc factory create "Hello Exercise $TEST_SUFFIX"
```

Capture the factory ID (e.g., `FACT-xxx`). Display to user:

```
Created test factory: FACT-xxx "Hello Exercise $TEST_SUFFIX"
This isolates your test from production data.
```

### 2. Launch Hello with Test Factory

Launch orc hello with the test factory:

```bash
orc hello --factory FACT-xxx
```

Explain to user:

```
Launching orc hello with test factory FACT-xxx...

This will start Claude Code with the /orc-first-run skill.
Complete the first-run walkthrough as a new user would.

When the skill finishes (you see the "Happy building!" message),
return here and I'll verify the results.
```

Wait for user to confirm they've completed the first-run flow.

### 3. Verify Results

After user confirms completion, verify entities were created:

```bash
# Check for commission in the test factory context
orc commission list

# Check for workshop with test factory
orc workshop list --factory FACT-xxx

# Check for workbench in the workshop
orc workbench list
```

For each check, report:
- `[PASS]` if entity exists
- `[FAIL]` if entity missing

Also verify:

```bash
# Verify REPO-001 is still registered
orc repo show REPO-001
```

### 4. Cleanup

Clean up test entities using the archive + infra apply pattern:

```bash
# Get IDs from the verification step, then:

# Archive workbenches first
orc workbench archive BENCH-xxx

# Archive workshop
orc workshop archive WORK-xxx

# Apply tmux to clean up session (reconciles tmux state)
orc tmux apply WORK-xxx --yes

# Archive the test commission
orc commission archive COMM-xxx
```

**Note:** ORC uses soft-delete (archive) rather than hard-delete for most entities. Archived entities are inert and hidden from normal listings. The factory cannot be deleted while it has workshops, but archiving the workshop makes it inactive. This is by design - it preserves audit history.

### 5. Report Results

Display test summary:

```
Hello Exercise Results
--------------------------
[PASS] Test factory created (FACT-xxx)
[PASS] Hello launched with --factory flag
[PASS] Commission created during first-run
[PASS] Workshop created with correct factory
[PASS] Workbench created
[PASS] REPO-001 still exists
[PASS] Cleanup successful (entities archived, infrastructure removed)

All tests passed! The first-run flow is working correctly.
```

## On Failure

If any step fails:
1. Report which step failed and the error
2. Attempt cleanup by archiving any created resources (archive is always safe)
3. Run `orc tmux apply WORK-xxx --yes` to clean up tmux session
4. Suggest running `orc doctor` for diagnostics

## Notes

- This tests the actual /orc-first-run skill with user interaction
- Unlike bootstrap-test.sh, this runs on your dev machine (not in VM)
- Uses isolated factory to avoid polluting production data
- The user must complete the first-run walkthrough manually
