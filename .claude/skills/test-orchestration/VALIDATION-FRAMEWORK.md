# Validation Framework

Comprehensive validation system for orchestration testing with 25 checkpoints across 6 phases.

## Overview

The validation framework ensures that every aspect of the orchestration test is verified:
- Infrastructure setup (mission, workspace, markers)
- TMux session structure
- Deputy ORC health and context detection
- Work order assignment and visibility
- Implementation progress and file changes
- Feature functionality and quality

**Total Checkpoints**: 25 across 6 phases
**Success Criteria**: All 25 checkpoints must pass for overall success

## Phase 1: Environment Setup (4 checkpoints)

### Purpose
Validate that the test mission workspace is correctly provisioned with all required markers and metadata.

### Checkpoints

#### 1.1: Mission ID Format
- **Check**: Mission created with correct ID format
- **Expected**: `MISSION-TEST-ORC-{timestamp}` where timestamp is epoch seconds
- **Method**: Regex match `^MISSION-TEST-ORC-[0-9]+$`
- **Failure Impact**: Critical - cannot proceed without valid mission

#### 1.2: Workspace Directory
- **Check**: Mission workspace directory exists
- **Expected**: Directory at `~/src/missions/{MISSION_ID}`
- **Method**: `os.Stat()` or `[ -d ~/src/missions/{MISSION_ID} ]`
- **Failure Impact**: Critical - cannot write markers

#### 1.3: Mission Marker Valid
- **Check**: `.orc-mission` marker file contains valid JSON
- **Expected**: JSON file with `mission_id`, `workspace_path`, `created_at`
- **Method**: `jq -e . {path}/.orc-mission` or `json.Unmarshal()`
- **Failure Impact**: Critical - deputy context detection will fail

#### 1.4: Workspace Metadata Valid
- **Check**: `.orc/metadata.json` file contains valid JSON
- **Expected**: JSON file with `active_mission_id`, `last_updated`
- **Method**: `jq -e . {path}/.orc/metadata.json` or `json.Unmarshal()`
- **Failure Impact**: Important - affects context detection priority

### Validation Script
```bash
# Run from mission workspace
[ -d ~/src/missions/$MISSION_ID ] || exit 1
jq -e '.mission_id' ~/src/missions/$MISSION_ID/.orc-mission || exit 1
jq -e '.active_mission_id' ~/src/missions/$MISSION_ID/.orc/metadata.json || exit 1
echo "Phase 1: 4/4 checkpoints passed"
```

### Phase Output
`turns/00-setup.md` with:
- Mission ID
- Workspace path
- Created files list
- Checkbox for each checkpoint (✓ or ✗)
- Overall status: PASS or FAIL

## Phase 2: TMux Session Deployment (5 checkpoints)

### Purpose
Validate that the TMux environment is correctly set up with deputy and IMP windows.

### Checkpoints

#### 2.1: Grove Database Record
- **Check**: Grove exists in database
- **Expected**: `orc grove list` shows the test grove
- **Method**: `orc grove list | grep test-canary-{timestamp}`
- **Failure Impact**: Critical - grove not tracked

#### 2.2: Worktree Directory
- **Check**: Worktree directory exists
- **Expected**: Directory at `~/src/worktrees/test-canary-{timestamp}`
- **Method**: `[ -d ~/src/worktrees/test-canary-{timestamp} ]`
- **Failure Impact**: Critical - no workspace for IMPs

#### 2.3: TMux Session Exists
- **Check**: TMux session created
- **Expected**: Session named `orc-{MISSION_ID}`
- **Method**: `tmux has-session -t orc-{MISSION_ID}`
- **Failure Impact**: Critical - cannot coordinate agents

#### 2.4: Deputy Window
- **Check**: Deputy window exists
- **Expected**: Window named "deputy" in session
- **Method**: `tmux list-windows -t orc-{MISSION_ID} | grep deputy`
- **Failure Impact**: Critical - no coordination pane

#### 2.5: IMP Window Layout
- **Check**: IMP window has correct 3-pane layout
- **Expected**: Window named "imp" with 3 panes (vim | claude | shell)
- **Method**: `scripts/verify-tmux-layout.sh orc-{MISSION_ID}`
- **Failure Impact**: Critical - IMP environment incorrect

### Validation Script
```bash
# Use helper script
./scripts/verify-tmux-layout.sh orc-$MISSION_ID
# Expected output: status=OK, layout=valid, imp_pane_count=3
```

### Phase Output
`turns/01-grove-deployed.md` with:
- Grove ID and path
- TMux session name
- Window count and names
- Pane layout details
- Validation results
- Status: PASS or FAIL

## Phase 3: Deputy ORC Health (4 checkpoints)

### Purpose
Validate that deputy ORC is operational and correctly detecting mission context.

### Checkpoints

#### 3.1: Deputy Health Check
- **Check**: Deputy health check passes
- **Expected**: `scripts/check-deputy-health.sh` returns status=OK
- **Method**: Run health check script
- **Failure Impact**: Critical - deputy not functional

#### 3.2: Status Command Shows Mission
- **Check**: `orc status` displays test mission ID
- **Expected**: Output contains mission ID
- **Method**: `cd {mission_dir} && orc status | grep {MISSION_ID}`
- **Failure Impact**: Critical - context detection broken

#### 3.3: Summary Shows Deputy Context
- **Check**: `orc summary` displays deputy context header
- **Expected**: Output contains "Deputy View" or "Deputy Context"
- **Method**: `cd {mission_dir} && orc summary | grep -i deputy`
- **Failure Impact**: Important - indicates context mode

#### 3.4: Work Order Creation Works
- **Check**: Can create work orders in deputy context
- **Expected**: Work order created successfully, auto-scoped to mission
- **Method**: `cd {mission_dir} && orc work-order create "Test WO"`
- **Failure Impact**: Critical - cannot assign work

### Validation Script
```bash
# Use helper script
./scripts/check-deputy-health.sh $MISSION_ID
# Expected output: status=OK, context=deputy, mission={MISSION_ID}
```

### Phase Output
`turns/02-deputy-verified.md` with:
- Health check results
- Context detection confirmation
- Sample command outputs
- Validation results
- Status: PASS or FAIL

## Phase 4: Work Assignment (3 checkpoints)

### Purpose
Validate that work orders are correctly created and visible in deputy summary.

### Checkpoints

#### 4.1: Parent Work Order Created
- **Check**: Parent work order exists
- **Expected**: Work order with title "Implement POST /echo endpoint"
- **Method**: `orc work-order list | grep "POST /echo"`
- **Failure Impact**: Critical - no work to track

#### 4.2: Child Work Orders Created
- **Check**: All 4 child work orders exist
- **Expected**: 4 work orders with specific titles
- **Method**: `orc work-order list | wc -l` shows at least 5 (parent + 4 children)
- **Failure Impact**: Critical - incomplete work breakdown

#### 4.3: Summary Shows Work Orders
- **Check**: All work orders visible in summary
- **Expected**: `orc summary` shows all work orders scoped to test mission
- **Method**: `cd {mission_dir} && orc summary | grep -c "WO-"` >= 5
- **Failure Impact**: Important - visibility issue

### Validation Script
```bash
cd ~/src/missions/$MISSION_ID
WO_COUNT=$(orc work-order list | grep -c "WO-")
[ "$WO_COUNT" -ge 5 ] || exit 1
echo "Phase 4: 3/3 checkpoints passed"
```

### Phase Output
`turns/03-work-assigned.md` with:
- Parent work order ID and title
- Child work order IDs and titles
- Summary output
- Validation results
- Status: PASS or FAIL

## Phase 5: Implementation Monitoring (4 checkpoints)

### Purpose
Validate that IMPs are making progress on the implementation.

### Checkpoints

#### 5.1: Files Modified
- **Check**: Required files modified in grove
- **Expected**: main.go, main_test.go, README.md exist
- **Method**: `[ -f {grove_path}/main.go ]` etc.
- **Failure Impact**: Critical - no implementation

#### 5.2: Git Shows Changes
- **Check**: Git detects uncommitted changes
- **Expected**: `git status` shows modified files
- **Method**: `cd {grove_path} && git status --short | wc -l` > 0
- **Failure Impact**: Important - indicates progress

#### 5.3: Work Orders Updated
- **Check**: At least some work orders marked complete
- **Expected**: Work order list shows status changes
- **Method**: `orc work-order list | grep -c "\[completed\]"` > 0
- **Failure Impact**: Important - progress tracking

#### 5.4: No Visible Errors
- **Check**: No errors in IMP pane
- **Expected**: Pane capture doesn't show "error" or "fail"
- **Method**: `scripts/monitor-imp-progress.sh` shows imp_status != "error_detected"
- **Failure Impact**: Important - indicates problems

### Validation Script
```bash
# Use helper script
./scripts/monitor-imp-progress.sh orc-$MISSION_ID $MISSION_ID
# Check: progress_percent > 0
```

### Phase Output
`turns/04-progress-N.md` (multiple files) with:
- Timestamp
- Files changed
- Work order status updates
- IMP activity observations
- Progress percentage
- Current state: in_progress, completed, or stuck

## Phase 6: Feature Validation (5 checkpoints)

### Purpose
Validate that the implemented feature works correctly and meets requirements.

### Checkpoints

#### 6.1: Code Compiles
- **Check**: Go build succeeds
- **Expected**: `go build` exits with code 0
- **Method**: `cd {grove_path} && go build`
- **Failure Impact**: Critical - broken code

#### 6.2: Tests Pass
- **Check**: All unit tests pass
- **Expected**: `go test` exits with code 0
- **Method**: `cd {grove_path} && go test ./...`
- **Failure Impact**: Critical - quality gate

#### 6.3: Manual Test Works
- **Check**: Curl test returns correct response
- **Expected**: Response contains "echo" and "timestamp" fields
- **Method**: `curl -X POST http://localhost:8080/echo -d '{"message":"test"}' | jq -e '.echo, .timestamp'`
- **Failure Impact**: Critical - feature broken

#### 6.4: README Updated
- **Check**: README contains /echo documentation
- **Expected**: `grep -i "/echo" README.md` returns match
- **Method**: `cd {grove_path} && grep "/echo" README.md`
- **Failure Impact**: Important - documentation requirement

#### 6.5: Feature Meets Requirements
- **Check**: All requirements from TEST-WORKLOAD.md met
- **Expected**: Comprehensive check of all acceptance criteria
- **Method**: `scripts/validate-feature.sh {grove_path}` returns status=OK
- **Failure Impact**: Critical - incomplete feature

### Validation Script
```bash
# Use helper script
./scripts/validate-feature.sh ~/src/worktrees/test-canary-$TIMESTAMP
# Expected output: status=OK, success_count >= 4
```

### Phase Output
`turns/05-validation.md` with:
- Build results (exit code, output)
- Test results (exit code, output)
- Manual test results (curl response)
- README verification
- Validation results (X/5 checkpoints passed)
- Status: PASS or FAIL

## Success Criteria

### Overall Test Success

The test **PASSES** if:
- All 25 checkpoints pass (100% success rate)
- Feature fully implemented and working
- All tests pass
- Documentation complete

The test **FAILS** if:
- Any Phase 1-4 checkpoint fails (critical infrastructure/setup issues)
- Less than 4/5 Phase 6 checkpoints pass (feature quality issues)
- Overall success rate < 92% (23/25 checkpoints)

### Partial Success

A test with 23-24 checkpoints passing may be considered "PARTIAL SUCCESS" if:
- All critical checkpoints pass (Phase 1-4, Phase 6 checks 6.1-6.3)
- Only optional checkpoints fail (Phase 6.4-6.5)

## Reporting

### Human-Readable Report

`turns/06-final-report.md` contains:

```markdown
# Orchestration Test: Final Report

## Executive Summary
- Test Run: test-20260114-123456
- Start: 2026-01-14 12:34:56
- End: 2026-01-14 12:45:23
- Duration: 10m 27s
- **Result: ✓ PASS**

## Phase Results
1. Setup: ✓ PASS (4/4 checks)
2. TMux Deploy: ✓ PASS (5/5 checks)
3. Deputy Health: ✓ PASS (4/4 checks)
4. Work Assigned: ✓ PASS (3/3 checks)
5. Implementation: ✓ PASS (4/4 checks)
6. Validation: ✓ PASS (5/5 checks)

**Overall: 25/25 checks passed (100%)**

[... detailed results ...]
```

### Machine-Readable Results

`turns/results.json` contains:

```json
{
  "test_run_id": "test-20260114-123456",
  "start_time": "2026-01-14T12:34:56Z",
  "end_time": "2026-01-14T12:45:23Z",
  "duration_seconds": 627,
  "overall_result": "PASS",
  "phases": {
    "setup": {
      "checkpoints_passed": 4,
      "checkpoints_total": 4,
      "status": "PASS",
      "checkpoints": [
        {"id": "1.1", "name": "Mission ID format", "status": "PASS"},
        {"id": "1.2", "name": "Workspace directory", "status": "PASS"},
        {"id": "1.3", "name": "Mission marker valid", "status": "PASS"},
        {"id": "1.4", "name": "Workspace metadata valid", "status": "PASS"}
      ]
    },
    "tmux_deploy": {
      "checkpoints_passed": 5,
      "checkpoints_total": 5,
      "status": "PASS"
    },
    "deputy_health": {
      "checkpoints_passed": 4,
      "checkpoints_total": 4,
      "status": "PASS"
    },
    "work_assigned": {
      "checkpoints_passed": 3,
      "checkpoints_total": 3,
      "status": "PASS"
    },
    "implementation": {
      "checkpoints_passed": 4,
      "checkpoints_total": 4,
      "status": "PASS"
    },
    "validation": {
      "checkpoints_passed": 5,
      "checkpoints_total": 5,
      "status": "PASS"
    }
  },
  "success_rate": 1.0,
  "feature_validation": {
    "build": "pass",
    "tests": "pass",
    "manual_test": "pass",
    "readme": "pass",
    "requirements": "pass"
  }
}
```

## CI/CD Integration

### Exit Codes

- **0**: All tests passed (25/25 checkpoints)
- **1**: Test failed (< 23/25 checkpoints)

### Usage in Pipeline

```bash
#!/bin/bash
# Run orchestration test
claude --skill test-orchestration

# Check exit code
if [ $? -eq 0 ]; then
  echo "✓ Orchestration test passed"

  # Extract success rate from results
  SUCCESS_RATE=$(jq -r '.success_rate' skills/test-orchestration/turns/results.json)
  echo "Success rate: $(echo "$SUCCESS_RATE * 100" | bc)%"

  exit 0
else
  echo "✗ Orchestration test failed"

  # Show which phases failed
  cat skills/test-orchestration/turns/06-final-report.md

  exit 1
fi
```

### Metrics Collection

Extract metrics from results.json:
- Test duration: `.duration_seconds`
- Success rate: `.success_rate`
- Phase results: `.phases[].status`
- Feature quality: `.feature_validation`

## Error Handling

### Checkpoint Failure Strategy

When a checkpoint fails:

1. **Phase 1-2 failures**: ABORT immediately
   - Environment/infrastructure not ready
   - Cannot proceed without basic setup

2. **Phase 3-4 failures**: ABORT after phase completion
   - Deputy/work assignment broken
   - No point monitoring broken system

3. **Phase 5 failures**: CONTINUE to Phase 6
   - Maybe partial implementation completed
   - Validate what was done

4. **Phase 6 failures**: CONTINUE to Phase 7
   - Generate report showing what failed
   - Useful for debugging

### Recovery Actions

For common failures:
- **TMux session exists**: Kill old session, retry
- **Grove exists**: Delete old grove, recreate
- **Mission exists**: Use unique timestamp in ID
- **Port in use**: Kill old server process
- **Timeout**: Proceed to validation anyway

## Performance Targets

### Expected Timings

- **Phase 1**: < 30 seconds
- **Phase 2**: 1-2 minutes
- **Phase 3**: < 30 seconds
- **Phase 4**: < 30 seconds
- **Phase 5**: 10-20 minutes (implementation)
- **Phase 6**: 2-3 minutes
- **Total**: 15-25 minutes typical

### Timeout Budget

- **Per-phase timeout**: 5 minutes (except Phase 5)
- **Phase 5 timeout**: 30 minutes
- **Overall timeout**: 40 minutes max

## Validation Tools Reference

### Helper Scripts

All in `scripts/` directory:

1. **verify-tmux-layout.sh**
   - Checks: Session exists, window count, pane layout
   - Output: key=value format
   - Used in: Phase 2

2. **check-deputy-health.sh**
   - Checks: Context detection, orc commands work
   - Output: key=value format
   - Used in: Phase 3

3. **monitor-imp-progress.sh**
   - Checks: Work order status, file changes, IMP activity
   - Output: key=value format
   - Used in: Phase 5

4. **validate-feature.sh**
   - Checks: Build, tests, manual test, README
   - Output: key=value format
   - Used in: Phase 6

5. **cleanup-test-env.sh**
   - Actions: Kill TMux, remove worktree, delete mission
   - Output: key=value format
   - Used in: Phase 7

### ORC CLI Commands

Added for testing support:

1. **orc mission delete [id]**
   - Deletes mission from database
   - Supports --force flag

2. **orc grove delete [id]**
   - Deletes grove from database
   - Supports --remove-worktree flag

3. **orc debug session-info**
   - Shows current context detection
   - Displays all markers and metadata

4. **orc debug validate-context [dir]**
   - Validates directory has proper ORC context
   - Checks markers, metadata, JSON validity

## Maintenance

### Adding New Checkpoints

1. Add to appropriate phase in this document
2. Update skill.md Phase instructions
3. Update config.json checkpoint counts
4. Update helper script if needed
5. Update report template

### Changing Success Criteria

1. Update this document's "Success Criteria" section
2. Update skill.md exit code logic
3. Update CI/CD documentation
4. Communicate changes to team

---

**Status**: ✓ Complete
**Last Updated**: 2026-01-14
**Maintained by**: ORC Development Team
